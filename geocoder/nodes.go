package geocoder

import (
	"encoding/binary"
	"fmt"
	"hstin/gocoder/mapping"
	"hstin/gocoder/structures"
	"hstin/gocoder/utils"
	"os"
	"syscall"
	"unsafe"
)

type Node struct {
	ID          int64      `json:"id"`
	DocumentID  int64      `json:"-"`
	Name        string     `json:"name"`
	Country     string     `json:"country"`
	Region      string     `json:"region"`
	SubRegion   string     `json:"subregion"`
	Coordinates [2]float32 `json:"coordinates"`
	BoundingBox [4]float32 `json:"boundingBox"`
	Population  uint32     `json:"population"`
	Timezone    string     `json:"timezone"`
	Rank        int        `json:"-"`
}

type NodesSearch struct {
	Nodes       []structures.Node
	Strings     StringSearcher
	nodesData   []byte
	stringsData []byte
	file        *os.File
	LanguageMap map[string]int
}

type StringSearcher struct {
	stringData []byte
}

func (s *StringSearcher) Get(offset uint64) []string {
	data := s.stringData[offset:]

	numberOfStrings := binary.LittleEndian.Uint32(data)
	result := make([]string, 0, numberOfStrings)

	pos := 4

	for i := uint32(0); i < numberOfStrings; i++ {
		length := binary.LittleEndian.Uint16(data[pos:])
		pos += 2
		result = append(result, string(data[pos:pos+int(length)]))
		pos += int(length)
	}

	return result
}

func (g *NodesSearch) Close() error {
	if g.nodesData != nil {
		if err := syscall.Munmap(g.nodesData); err != nil {
			return err
		}
	}
	if g.stringsData != nil {
		if err := syscall.Munmap(g.stringsData); err != nil {
			return err
		}
	}
	return g.file.Close()
}

func (g *NodesSearch) LoadSingleFile(
	file *os.File,
	nodesOffset int64,
	nodesSize uint64,
	stringsOffset int64,
	stringsSize uint64,
	languageMap map[string]int,
) error {
	g.file = file
	g.LanguageMap = languageMap

	// Memory map nodes data
	pageSize := int64(syscall.Getpagesize())

	// Map nodes
	nodesPageOffset := (nodesOffset / pageSize) * pageSize
	nodesAdjustment := nodesOffset - nodesPageOffset
	nodesMmapSize := int(nodesSize + uint64(nodesAdjustment))

	nodesData, err := syscall.Mmap(
		int(file.Fd()),
		nodesPageOffset,
		nodesMmapSize,
		syscall.PROT_READ,
		syscall.MAP_SHARED,
	)
	if err != nil {
		return fmt.Errorf("failed to mmap nodes data: %w", err)
	}
	g.nodesData = nodesData

	// Map strings
	stringsPageOffset := (stringsOffset / pageSize) * pageSize
	stringsAdjustment := stringsOffset - stringsPageOffset
	stringsMmapSize := int(stringsSize + uint64(stringsAdjustment))

	stringsData, err := syscall.Mmap(
		int(file.Fd()),
		stringsPageOffset,
		stringsMmapSize,
		syscall.PROT_READ,
		syscall.MAP_SHARED,
	)
	if err != nil {
		syscall.Munmap(nodesData)
		return fmt.Errorf("failed to mmap strings data: %w", err)
	}
	g.stringsData = stringsData

	nodeCount := int(nodesSize / structures.NodeSize)
	g.Nodes = unsafe.Slice(
		(*structures.Node)(unsafe.Pointer(&nodesData[nodesAdjustment])),
		nodeCount,
	)

	g.Strings = StringSearcher{
		stringData: stringsData[stringsAdjustment : stringsAdjustment+int64(stringsSize)],
	}

	return nil
}

func (g *NodesSearch) GetNode(id int64, lang string) Node {
	node := g.Nodes[id]

	nameStrings := g.Strings.Get(node.NameOffset)
	regionStrings := g.Strings.Get(g.Nodes[id].RegionOffset)

	langKey, ok := g.LanguageMap[lang]
	if !ok {
		langKey = 0
	}

	startIndex := langKey * 2

	name := nameStrings[langKey]
	region := regionStrings[startIndex]
	subRegion := regionStrings[startIndex+1]

	var country string
	if int(node.Country) < len(mapping.CountryCodes) {
		country = mapping.CountryCodes[node.Country]
	} else {
		country = ""
	}

	var timezone string
	if int(node.Timezone) < len(utils.TimezoneNames) {
		timezone = utils.TimezoneNames[node.Timezone]
	} else {
		timezone = "Etc/UTC"
	}

	return Node{
		ID:          node.ID,
		DocumentID:  id,
		Name:        name,
		Country:     country,
		Region:      region,
		SubRegion:   subRegion,
		Coordinates: g.Nodes[id].Center,
		BoundingBox: g.Nodes[id].BoundingBox,
		Population:  g.Nodes[id].Population,
		Timezone:    timezone,
		Rank:        int(g.Nodes[id].Rank),
	}
}
