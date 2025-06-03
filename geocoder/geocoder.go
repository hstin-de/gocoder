package geocoder

import (
	"encoding/binary"
	"hstin/geocoder/config"
	"hstin/geocoder/structures"
	"hstin/geocoder/utils"
	"io"
	"log"
	"os"
	"runtime"
	"sort"
	"strings"
	"sync"
)

type Geocoder struct {
	DocumentMap structures.DocumentMap
	cache       map[string]structures.CacheEntry
	cacheLock   sync.RWMutex
	nSearch     *NodesSearch
	Trie        *structures.Trie
	Index       *structures.Index
	KDTree      *structures.KDTree
}

func (g *Geocoder) Close() error {
	return g.nSearch.Close()
}

func NewGeocoder(DatabaseFile string) (*Geocoder, error) {
	log.Println("Initializing Geocoder...")
	log.Println("Loading Database...")

	f, err := os.Open(DatabaseFile)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// Read header to get offsets and sizes (10 uint64 values)
	header := make([]byte, 8*10) // 10 uint64 values
	if _, err := f.Read(header); err != nil {
		return nil, err
	}

	nodesOffset := binary.LittleEndian.Uint64(header[0:])
	nodesSize := binary.LittleEndian.Uint64(header[8:])
	stringsOffset := binary.LittleEndian.Uint64(header[16:])
	stringsSize := binary.LittleEndian.Uint64(header[24:])
	dmOffset := binary.LittleEndian.Uint64(header[32:])
	dmSize := binary.LittleEndian.Uint64(header[40:])
	trieOffset := binary.LittleEndian.Uint64(header[48:])
	trieSize := binary.LittleEndian.Uint64(header[56:])
	indexOffset := binary.LittleEndian.Uint64(header[64:])
	indexSize := binary.LittleEndian.Uint64(header[72:])

	// Read KDTree size
	var kdTreeSize uint64
	if err := binary.Read(f, binary.LittleEndian, &kdTreeSize); err != nil {
		return nil, err
	}

	kdTreeOffset := indexOffset + indexSize

	// Languages start right after the header + kdTreeSize field
	languagesOffset := uint64(8*10 + 8)

	// Read languages section
	if _, err := f.Seek(int64(languagesOffset), io.SeekStart); err != nil {
		return nil, err
	}

	var languageCount uint64
	if err := binary.Read(f, binary.LittleEndian, &languageCount); err != nil {
		return nil, err
	}

	languages := make([]string, 0, languageCount)
	for i := uint64(0); i < languageCount; i++ {
		var langLength uint64
		if err := binary.Read(f, binary.LittleEndian, &langLength); err != nil {
			return nil, err
		}
		langData := make([]byte, langLength)
		if _, err := io.ReadFull(f, langData); err != nil {
			return nil, err
		}
		languages = append(languages, string(langData))
	}

	languageMap := make(map[string]int, len(languages))
	for i, lang := range languages {
		languageMap[lang] = i + 1
	}

	var (
		nSearch     NodesSearch
		documentMap structures.DocumentMap
		trie        structures.Trie
		index       structures.Index
		KDTree      structures.KDTree
	)

	if config.EnableForward {

		// 1. Load Document Map
		log.Printf("Loading document map (%d MB)...", dmSize/1024/1024)
		if err := documentMap.LoadFromFile(f, int64(dmOffset), dmSize); err != nil {
			return nil, err
		}
		runtime.GC()

		// 2. Load Trie
		log.Printf("Loading trie (%d MB)...", trieSize/1024/1024)
		if err := trie.LoadFromFile(f, int64(trieOffset), trieSize); err != nil {
			return nil, err
		}
		runtime.GC()

		// 3. Load Index
		log.Printf("Loading index (%d MB)...", indexSize/1024/1024)
		if err := index.LoadFromFile(f, int64(indexOffset), indexSize); err != nil {
			return nil, err
		}
		runtime.GC()

	}

	if config.EnableReverse {

		// 4. Load KDTree
		log.Printf("Loading KD tree (%d MB)...", kdTreeSize/1024/1024)
		if err := KDTree.LoadFromFile(f, int64(kdTreeOffset), kdTreeSize); err != nil {
			return nil, err
		}
		runtime.GC()

	}

	// 5. Load Nodes Search (this opens its own file handle)
	log.Printf("Loading nodes search...")
	nodeFile, err := os.Open(DatabaseFile)
	if err != nil {
		return nil, err
	}
	nSearch.LoadSingleFile(nodeFile, int64(nodesOffset), nodesSize, int64(stringsOffset), stringsSize, languageMap)

	// Final garbage collection
	runtime.GC()
	log.Printf("Geocoder initialization complete")

	return &Geocoder{
		DocumentMap: documentMap,
		cache:       make(map[string]structures.CacheEntry),
		nSearch:     &nSearch,
		Trie:        &trie,
		Index:       &index,
		KDTree:      &KDTree,
	}, nil
}

type WeightedNode struct {
	node Node
	rank int
}

func (g *Geocoder) Cache(query string, nodes []Node) {
	cacheIDs := make([]int64, 0)

	for _, node := range nodes {
		cacheIDs = append(cacheIDs, node.DocumentID)
	}
	g.cacheLock.Lock()
	g.cache[query] = structures.CacheEntry{
		Results: cacheIDs,
		Found:   len(nodes),
	}
	g.cacheLock.Unlock()
}

func (g *Geocoder) Search(
	query string,
	maxResults int,
	useCache bool,
	lang string,
) (map[string]interface{}, bool) {
	if query == "" {
		return map[string]interface{}{
			"found":   0,
			"results": []Node{},
		}, false
	}

	normalizedQuery := strings.ToLower(strings.TrimSpace(query))

	returnMap := make(map[int64]Node)

	if useCache && !config.DisableCache {
		g.cacheLock.RLock()
		cached, ok := g.cache[normalizedQuery]
		g.cacheLock.RUnlock()

		if ok {

			returnDocs := make([]Node, 0, cached.Found)
			for _, docID := range cached.Results {
				node := g.nSearch.GetNode(docID, lang)
				returnDocs = append(returnDocs, node)
			}

			if maxResults > 0 && len(returnDocs) > maxResults {
				returnDocs = returnDocs[:maxResults]
			}

			return map[string]interface{}{
				"found":   cached.Found,
				"results": returnDocs,
			}, true
		}
	}

	// 1) TRIE SEARCH
	trieResults := g.Trie.Search(normalizedQuery)

	for _, docID := range trieResults {
		node := g.nSearch.GetNode(docID, lang)
		node.Rank += 500
		returnMap[node.ID] = node
	}

	// 2) FUZZY SEARCH
	if len(returnMap) < 10 && len(normalizedQuery) > 2 {
		maxDistance := 1
		if len(normalizedQuery) > 4 {
			maxDistance = 2
		}

		indexResults := g.Index.Search(normalizedQuery, maxDistance)

		for _, docID := range indexResults {
			node := g.nSearch.GetNode(docID, lang)
			node.Rank -= 100
			if _, ok := returnMap[node.ID]; !ok {
				returnMap[node.ID] = node
			}
		}
	}

	// Collect and sort the nodes
	returnDocs := sortNodes(utils.MapToSlice(returnMap))
	foundElements := len(returnDocs)

	// Cache with normalized query if cache isnt disabled
	if !config.DisableCache {
		g.Cache(normalizedQuery, returnDocs)
	}

	// If maxResults > 0, limit the returned slice
	if maxResults > 0 && len(returnDocs) > maxResults {
		returnDocs = returnDocs[:maxResults]
	}

	return map[string]interface{}{
		"found":   foundElements,
		"results": returnDocs,
	}, false
}

func (g *Geocoder) Reverse(lat float64, lng float64, lang string) map[string]interface{} {
	results := g.KDTree.KNN(
		structures.NewPoint(0, [2]float32{float32(lat), float32(lng)}),
		1,
	)
	if len(results) == 0 {
		return map[string]interface{}{
			"results": []Node{},
		}
	}

	return map[string]interface{}{
		"results": []Node{g.nSearch.GetNode(results[0].ID, lang)},
	}
}

func (g *Geocoder) GetNode(docID int64, lang string) Node {
	return g.nSearch.GetNode(int64(g.DocumentMap[docID]), lang)
}

func sortNodes(nodes []Node) []Node {
	// Sort by Rank descending
	sort.Slice(nodes, func(i, j int) bool {
		if nodes[i].Rank != nodes[j].Rank {
			// Primary sort: higher rank first
			return nodes[i].Rank > nodes[j].Rank
		}
		if nodes[i].Name != nodes[j].Name {
			// Secondary sort: if ranks are equal, sort by name alphabetically
			return nodes[i].Name < nodes[j].Name
		}
		// Tertiary sort: if rank and name are the same, sort by ID
		return nodes[i].ID < nodes[j].ID
	})

	return nodes
}
