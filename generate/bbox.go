package generate

import (
	"context"
	"encoding/gob"
	"hstin/gocoder/config"
	"io"
	"log"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/paulmach/osm"
	"github.com/paulmach/osm/osmpbf"
)

func checkIfFileExistsAndIsValid(file string) (bool, map[int64][4]float32) {
	_, err := os.Stat(file)
	if os.IsNotExist(err) {
		log.Printf("[BBOX] File does not exist: %s", file)
		return false, nil
	}

	var refBBoxMap map[int64][4]float32
	f, err := os.Open(file)
	if err != nil {
		log.Printf("[BBOX] Error opening file %s: %v", file, err)
		return false, nil
	}
	defer f.Close()

	dec := gob.NewDecoder(f)
	err = dec.Decode(&refBBoxMap)
	if err != nil {
		log.Printf("[BBOX] Error decoding file %s: %v", file, err)
		return false, nil
	}

	log.Printf("[BBOX] File %s is valid and contains %d bounding boxes", file, len(refBBoxMap))
	return true, refBBoxMap
}

func LoadBoundingBoxes() map[int64][4]float32 {
	log.Printf("[BBOX] Loading bounding boxes")
	if valid, refBBoxMap := checkIfFileExistsAndIsValid(config.BoundingBoxes); valid {
		log.Printf("[BBOX] Existing bounding boxes found, skipping generation...")
		return refBBoxMap
	}

	log.Printf("[BBOX] Existing bounding boxes not found or invalid. Generating bounding boxes. This may take a while...")
	startTime := time.Now()

	f, err := os.Open(config.Planet)
	if err != nil {
		log.Fatalf("[BBOX] Failed to open planet file %s: %v", config.Planet, err)
	}
	defer f.Close()

	scanner := osmpbf.New(context.Background(), f, runtime.GOMAXPROCS(-1))
	defer func() {
		if err := scanner.Close(); err != nil {
			log.Printf("[BBOX] Error closing scanner: %v", err)
		}
	}()

	nodeIDs := make(map[int64]bool)
	wayIDs := make(map[int64]bool)
	refMembers := make(map[int64][]int64)

	scanner.SkipNodes = true
	scanner.SkipWays = true

	scanner.FilterRelation = func(relation *osm.Relation) bool {
		return relation.Tags.HasTag("boundary") && relation.Tags.Find("boundary") == "administrative" && relation.Tags.Find("admin_level") != "2"
	}

	log.Printf("[BBOX] Starting first pass to identify relevant administrative boundary relations")
	for scanner.Scan() {
		switch o := scanner.Object().(type) {
		case *osm.Relation:
			allRefs := make([]int64, 0)
			for _, member := range o.Members {
				allRefs = append(allRefs, member.Ref)
				if member.Type == osm.TypeWay {
					wayIDs[member.Ref] = true
				}
			}
			for _, ref := range allRefs {
				if _, ok := refMembers[ref]; !ok {
					refMembers[ref] = make([]int64, 0)
				}
				refs := make([]int64, 0)
				for _, member := range o.Members {
					if member.Type == osm.TypeWay {
						refs = append(refs, member.Ref)
					}
				}
				refMembers[ref] = append(refMembers[ref], refs...)
			}
		}
	}
	log.Printf("[BBOX] First pass completed. Loaded %d relations.", len(refMembers))

	if err := scanner.Err(); err != nil {
		log.Fatalf("[BBOX] Scanner error during first pass: %v", err)
	}

	if _, err := f.Seek(0, io.SeekStart); err != nil {
		log.Fatalf("[BBOX] Failed to seek to start of planet file: %v", err)
	}
	scanner = osmpbf.New(context.Background(), f, runtime.GOMAXPROCS(-1))
	scanner.SkipNodes = true
	scanner.SkipRelations = true

	scanner.FilterWay = func(way *osm.Way) bool {
		return wayIDs[int64(way.ID)]
	}

	ways := make(map[int64][]int64, 0)
	log.Printf("[BBOX] Starting second pass to collect ways and node IDs")
	for scanner.Scan() {
		switch o := scanner.Object().(type) {
		case *osm.Way:
			for _, node := range o.Nodes {
				nodeIDs[int64(node.ID)] = true
				ways[int64(o.ID)] = append(ways[int64(o.ID)], int64(node.ID))
			}
		}
	}
	log.Printf("[BBOX] Second pass completed. Loaded %d ways.", len(wayIDs))

	f.Seek(0, io.SeekStart)

	scanner = osmpbf.New(context.Background(), f, runtime.GOMAXPROCS(-1))
	scanner.SkipWays = true
	scanner.SkipRelations = true

	scanner.FilterNode = func(node *osm.Node) bool {
		return nodeIDs[int64(node.ID)]
	}

	nodes := make(map[int64][2]float32, 0)
	log.Printf("[BBOX] Starting third pass to collect node coordinates")
	for scanner.Scan() {
		switch o := scanner.Object().(type) {
		case *osm.Node:
			nodes[int64(o.ID)] = [2]float32{float32(o.Lat), float32(o.Lon)}
		}
	}
	log.Printf("[BBOX] Third pass completed. Loaded %d nodes.", len(nodes))

	refBBoxMap := make(map[int64][4]float32, len(refMembers))
	var mu sync.Mutex
	var wg sync.WaitGroup

	log.Printf("[BBOX] Starting bounding box calculations")
	for idx, ref := range refMembers {
		wg.Add(1)
		go func(idx int64, ref []int64) {
			defer wg.Done()

			latMin := float32(90)
			latMax := float32(-90)
			lonMin := float32(180)
			lonMax := float32(-180)

			for _, member := range ref {
				for _, nodeID := range ways[member] {
					if node, exists := nodes[nodeID]; exists {
						if node[0] < latMin {
							latMin = node[0]
						}
						if node[0] > latMax {
							latMax = node[0]
						}
						if node[1] < lonMin {
							lonMin = node[1]
						}
						if node[1] > lonMax {
							lonMax = node[1]
						}
					}
				}
			}

			mu.Lock()
			refBBoxMap[idx] = [4]float32{latMin, lonMin, latMax, lonMax}
			mu.Unlock()
		}(int64(idx), ref)
	}

	wg.Wait()
	log.Printf("[BBOX] Bounding box calculations completed")

	log.Printf("[BBOX] Saving bounding boxes to file")
	f, err = os.Create(config.BoundingBoxes)
	if err != nil {
		log.Fatalf("[BBOX] Failed to create bounding boxes file %s: %v", config.BoundingBoxes, err)
	}
	defer f.Close()

	enc := gob.NewEncoder(f)
	if err := enc.Encode(refBBoxMap); err != nil {
		log.Fatalf("[BBOX] Failed to encode bounding boxes: %v", err)
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("[BBOX] Scanner error during third pass: %v", err)
	}

	elapsed := time.Since(startTime)
	log.Printf("[BBOX] Bounding boxes generated in %s", elapsed)

	return refBBoxMap
}
