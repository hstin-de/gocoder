package generate

import (
	"bytes"
	"context"
	"encoding/binary"
	"hstin/gocoder/mapping"
	"hstin/gocoder/structures"
	"log"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"hstin/gocoder/config"
	"hstin/gocoder/utils"

	"github.com/paulmach/osm"
	"github.com/paulmach/osm/osmpbf"
)

var (
	refBBoxMap map[int64][4]float32
	adminTree  AdminTree
	nodeChan   = make(chan *osm.Node)
	insertChan = make(chan *InsertibleNode)
	wg         sync.WaitGroup
)

func CreateScanner() *osmpbf.Scanner {
	f, err := os.Open(config.Planet)
	if err != nil {
		panic(err)
	}

	scanner := osmpbf.New(context.Background(), f, runtime.GOMAXPROCS(-1))

	scanner.SkipRelations = true
	scanner.SkipWays = true

	scanner.FilterNode = func(node *osm.Node) bool {
		place := node.Tags.Find("place")
		if place == "" {
			return false
		}

		if !node.Tags.HasTag("name") {
			return false
		}

		if place == "country" || place == "state" || place == "region" ||
			place == "province" || place == "district" || place == "county" ||
			place == "subdistrict" || place == "continent" ||
			place == "archipelago" || place == "islet" || place == "square" ||
			place == "locality" || place == "polder" || place == "sea" ||
			place == "ocean" {
			return false
		}

		return true
	}

	return scanner
}

type InsertibleNode struct {
	AlternamteNames []string
	Node            structures.TmpNode
}

func GenerateDatabase() {

	if config.Planet == "" {
		log.Fatal("No planet file specified")
		return
	}

	if config.WhosOnFirst == "" {
		log.Fatal("No whosonfirst file specified")
		return
	}

	if config.WikimediaImportance == "" {
		log.Fatal("No wikimedia importance file specified")
		return
	}

	utils.LoadTimezones()
	utils.LoadImportanceMap()
	log.Println("[GENERATE] Using >", config.Planet, "< as input file.")
	log.Println("[GENERATE] Who's on First database:", config.WhosOnFirst)
	log.Println("[GENERATE] Wikimedia Importance:", config.WikimediaImportance)

	startTime := time.Now()

	refBBoxMap = LoadBoundingBoxes()
	adminTree = LoadAdminAreas()

	stringStore, err := structures.NewNames()
	if err != nil {
		log.Fatal(err)
	}

	scanner := CreateScanner()
	defer scanner.Close()

	var nodes = make([]structures.Node, 0)
	var documentMap = make(structures.DocumentMap)
	trie := structures.NewTrie()
	index := structures.NewIndex()

	kdPoints := make([]*structures.Point, 0)

	log.Println("[GENERATE] Indexing OSM nodes and generating data structures.")

	for i := 0; i < runtime.NumCPU(); i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for node := range nodeChan {
				tmpNode := structures.TmpNode{
					ID:       int64(node.ID),
					Names:    make(map[string]string),
					Center:   [2]float32{float32(node.Lat), float32(node.Lon)},
					Timezone: utils.GetTimezone(node.Lat, node.Lon),
				}

				tags := node.Tags.Map()

				// POPULATION
				tmpNode.Population = utils.ParseStringAsNumber(tags["population"])

				// RANK
				tmpNode.Rank = utils.CreateRank(tags, tmpNode.Population)

				// ADMIN AREAS
				adminArea := adminTree.GetCounty(node.Lat, node.Lon)
				tmpNode.Regions = adminArea.Regions
				tmpNode.Country = adminArea.Country

				// BOUNDING BOX
				if bbox, ok := refBBoxMap[int64(node.ID)]; ok {
					tmpNode.BoundingBox = bbox
				} else {
					tmpNode.BoundingBox = [4]float32{
						float32(node.Lat),
						float32(node.Lon),
						float32(node.Lat),
						float32(node.Lon),
					}
				}

				// NAMES
				name := tags["name"]
				tmpNode.Names["name"] = name

				for _, lang := range config.Languages {
					lang_name, ok := tags["name:"+lang]
					if ok {
						tmpNode.Names[lang] = lang_name
						continue
					}
					tmpNode.Names[lang] = name
				}

				// ALTERNATE NAMES
				var alternateNames []string = make([]string, 0)
				for key, value := range tags {
					if strings.HasPrefix(key, "name:") {
						alternateNames = append(alternateNames, value)
					}
				}

				insertChan <- &InsertibleNode{
					Node:            tmpNode,
					AlternamteNames: alternateNames,
				}
			}
		}()
	}

	var nodeWithoutCountry = 0
	var insertedIntoTrie = 0

	go func() {
		var documentID int64 = 0

		for insertibleNode := range insertChan {
			cityNode := insertibleNode.Node

			nameArray := []string{cityNode.Names["name"]}
			regionArray := []string{
				cityNode.Regions["name"].Region,
				cityNode.Regions["name"].SubRegion,
			}

			for _, lang := range config.Languages {
				nameArray = append(nameArray, cityNode.Names[lang])
				regionArray = append(regionArray, cityNode.Regions[lang].Region)
				regionArray = append(
					regionArray,
					cityNode.Regions[lang].SubRegion,
				)
			}

			nameOffset, err := stringStore.Store(nameArray)
			if err != nil {
				log.Fatal(err)
			}

			regionOffset, err := stringStore.Store(regionArray)
			if err != nil {
				log.Fatal(err)
			}

			tzIndex := 0
			for i, tz := range utils.TimezoneNames {
				if tz == cityNode.Timezone {
					tzIndex = i
					break
				}
			}

			nodes = append(nodes, structures.Node{
				ID:           cityNode.ID,
				NameOffset:   uint64(nameOffset),
				RegionOffset: uint64(regionOffset),
				Population:   uint32(cityNode.Population),
				Rank:         uint16(cityNode.Rank),
				Timezone:     uint16(tzIndex),
				Country:      uint8(mapping.GetCountryNumber(cityNode.Country)),
				Center:       cityNode.Center,
				BoundingBox:  cityNode.BoundingBox,
			})

			documentMap[cityNode.ID] = int32(documentID)

			trie.Insert(documentID, cityNode.Names["name"])
			index.AddDocument(documentID, cityNode.Names["name"])

			for _, altName := range insertibleNode.AlternamteNames {
				trie.Insert(documentID, altName)
				index.AddDocument(documentID, altName)
			}

			insertedIntoTrie += len(insertibleNode.AlternamteNames) + 1

			kdPoints = append(
				kdPoints,
				structures.NewPoint(documentID, cityNode.Center),
			)

			documentID++

			if cityNode.Country == "" {
				nodeWithoutCountry++
			}
		}
	}()

	for scanner.Scan() {
		switch o := scanner.Object().(type) {
		case *osm.Node:
			nodeChan <- o
		default:
			break
		}
	}

	close(nodeChan)
	wg.Wait()
	close(insertChan)

	log.Println("[GENERATE] Inserted into trie:", insertedIntoTrie)
	log.Println("[GENERATE] Nodes without country:", nodeWithoutCountry)
	log.Println(
		"[GENERATE] Building KD tree for fast nearest neighbor search.",
	)

	KDTree := structures.New(kdPoints)

	log.Println("[GENERATE] Saving database")

	err = SaveSingleDatabase(
		nodes,
		stringStore,
		documentMap,
		trie,
		index,
		KDTree,
		config.Output,
	)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("[GENERATE] Database saved.")
	log.Println("[GENERATE] Generation completed in", time.Since(startTime))
}

func SaveSingleDatabase(
	nodes []structures.Node,
	stringStore *structures.Names,
	documentMap structures.DocumentMap,
	trie *structures.Trie,
	index *structures.Index,
	kdTree *structures.KDTree,
	filename string,
) error {
	// Serialize all components
	var nodeBuffer bytes.Buffer
	for _, n := range nodes {
		nodeBuffer.Write(n.Serialize())
	}
	nodesBytes := nodeBuffer.Bytes()
	stringsBytes := stringStore.Bytes()

	// Serialize languages
	languagesBuf := new(bytes.Buffer)
	if err := binary.Write(
		languagesBuf,
		binary.LittleEndian,
		uint64(len(config.Languages)),
	); err != nil {
		return err
	}

	for _, lang := range config.Languages {
		langBytes := []byte(lang)
		if err := binary.Write(
			languagesBuf,
			binary.LittleEndian,
			uint64(len(langBytes)),
		); err != nil {
			return err
		}
		if _, err := languagesBuf.Write(langBytes); err != nil {
			return err
		}
	}
	serializedLanguages := languagesBuf.Bytes()

	// Serialize other components
	var DMBytesBuffer bytes.Buffer
	if err := documentMap.Save(&DMBytesBuffer); err != nil {
		return err
	}
	DMBytes := DMBytesBuffer.Bytes()

	var TrieBytesBuffer bytes.Buffer
	if err := trie.Save(&TrieBytesBuffer); err != nil {
		return err
	}
	TrieBytes := TrieBytesBuffer.Bytes()

	var IndexBytesBuffer bytes.Buffer
	if err := index.Save(&IndexBytesBuffer); err != nil {
		return err
	}
	IndexBytes := IndexBytesBuffer.Bytes()

	var KDTreeBytesBuffer bytes.Buffer
	if err := kdTree.Save(&KDTreeBytesBuffer); err != nil {
		return err
	}
	KDTreeBytes := KDTreeBytesBuffer.Bytes()

	// Calculate offsets - header is 10 uint64 values (80 bytes) + kdTreeSize (8 bytes)
	headerSize := 8 * 10
	kdTreeSizeFieldSize := 8
	languagesOffset := headerSize + kdTreeSizeFieldSize
	nodesOffset := languagesOffset + len(serializedLanguages)
	stringsOffset := nodesOffset + len(nodesBytes)
	dmOffset := stringsOffset + len(stringsBytes)
	trieOffset := dmOffset + len(DMBytes)
	indexOffset := trieOffset + len(TrieBytes)
	// kdTreeOffset is calculated in the loading code as indexOffset + indexSize

	// Create file and write header
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	// Write complete header with all offsets and sizes (10 uint64 values)
	header := make([]byte, headerSize)
	binary.LittleEndian.PutUint64(header[0:], uint64(nodesOffset))
	binary.LittleEndian.PutUint64(header[8:], uint64(len(nodesBytes)))
	binary.LittleEndian.PutUint64(header[16:], uint64(stringsOffset))
	binary.LittleEndian.PutUint64(header[24:], uint64(len(stringsBytes)))
	binary.LittleEndian.PutUint64(header[32:], uint64(dmOffset))
	binary.LittleEndian.PutUint64(header[40:], uint64(len(DMBytes)))
	binary.LittleEndian.PutUint64(header[48:], uint64(trieOffset))
	binary.LittleEndian.PutUint64(header[56:], uint64(len(TrieBytes)))
	binary.LittleEndian.PutUint64(header[64:], uint64(indexOffset))
	binary.LittleEndian.PutUint64(header[72:], uint64(len(IndexBytes)))

	if _, err := f.Write(header); err != nil {
		return err
	}

	// Write KDTree size separately (since we ran out of header space)
	if err := binary.Write(f, binary.LittleEndian, uint64(len(KDTreeBytes))); err != nil {
		return err
	}

	// Write all data sections
	if _, err := f.Write(serializedLanguages); err != nil {
		return err
	}
	if _, err := f.Write(nodesBytes); err != nil {
		return err
	}
	if _, err := f.Write(stringsBytes); err != nil {
		return err
	}
	if _, err := f.Write(DMBytes); err != nil {
		return err
	}
	if _, err := f.Write(TrieBytes); err != nil {
		return err
	}
	if _, err := f.Write(IndexBytes); err != nil {
		return err
	}
	if _, err := f.Write(KDTreeBytes); err != nil {
		return err
	}

	return nil
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
