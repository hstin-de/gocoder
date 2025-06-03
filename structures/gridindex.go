package structures

import (
	"hstin/gocoder/geo"
	"runtime"
	"sync"

	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geojson"
)

type UniformGridIndex struct {
	cellSize float64
	grid     map[[2]int][]*geojson.Feature
	min      orb.Point
	max      orb.Point
}

func NewUniformGridIndex(features []*geojson.Feature, cellSize float64) *UniformGridIndex {
	idx := &UniformGridIndex{
		cellSize: cellSize,
		grid:     make(map[[2]int][]*geojson.Feature),
	}
	for _, f := range features {
		if p, ok := f.Geometry.(orb.Polygon); ok {
			b := p.Bound()
			if len(idx.min) == 0 || b.Min[0] < idx.min[0] {
				idx.min[0] = b.Min[0]
			}
			if len(idx.min) == 0 || b.Min[1] < idx.min[1] {
				idx.min[1] = b.Min[1]
			}
			if len(idx.max) == 0 || b.Max[0] > idx.max[0] {
				idx.max[0] = b.Max[0]
			}
			if len(idx.max) == 0 || b.Max[1] > idx.max[1] {
				idx.max[1] = b.Max[1]
			}
		}
	}
	cpu := runtime.NumCPU()
	chunk := (len(features) + cpu - 1) / cpu
	var wg sync.WaitGroup
	var mu sync.Mutex
	for i := 0; i < cpu; i++ {
		start := i * chunk
		end := start + chunk
		if end > len(features) {
			end = len(features)
		}
		if start >= len(features) {
			break
		}
		wg.Add(1)
		go func(ftrs []*geojson.Feature) {
			defer wg.Done()
			localGrid := make(map[[2]int][]*geojson.Feature)
			for _, f := range ftrs {
				if p, ok := f.Geometry.(orb.Polygon); ok {
					b := p.Bound()
					minX := int((b.Min[0] - idx.min[0]) / idx.cellSize)
					minY := int((b.Min[1] - idx.min[1]) / idx.cellSize)
					maxX := int((b.Max[0] - idx.min[0]) / idx.cellSize)
					maxY := int((b.Max[1] - idx.min[1]) / idx.cellSize)
					for x := minX; x <= maxX; x++ {
						for y := minY; y <= maxY; y++ {
							cell := [2]int{x, y}
							localGrid[cell] = append(localGrid[cell], f)
						}
					}
				}
			}
			mu.Lock()
			for cell, fs := range localGrid {
				idx.grid[cell] = append(idx.grid[cell], fs...)
			}
			mu.Unlock()
		}(features[start:end])
	}
	wg.Wait()
	return idx
}

func (idx *UniformGridIndex) Search(point orb.Point) *geojson.Feature {
	x := int((point[0] - idx.min[0]) / idx.cellSize)
	y := int((point[1] - idx.min[1]) / idx.cellSize)
	cell := [2]int{x, y}
	candidates, exists := idx.grid[cell]
	if !exists {
		return nil
	}
	for _, f := range candidates {
		if p, ok := f.Geometry.(orb.Polygon); ok {
			if geo.PointInPolygon(point, p) {
				return f
			}
		}
	}
	return nil
}

func LoadCountriesWithGrid(fc *geojson.FeatureCollection, cellSize float64) *UniformGridIndex {
	return NewUniformGridIndex(fc.Features, cellSize)
}
