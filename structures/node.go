package structures

import (
	"encoding/binary"
	"math"
)

// Node is a fixed-size (64-byte) structure that can be directly mapped into memory.
// Field layout chosen to minimize padding and ensure 64-byte total size.
//
// Layout (with offsets and sizes):

type Node struct {
	ID           int64   // 0-7
	NameOffset   uint64  // 8-15
	RegionOffset uint64  // 16-23
	Population   uint32  // 24-27
	Rank         uint16  // 28-29
	Timezone     uint16  // 30-31
	Country      uint8   // 32
	_            [3]byte // 33-35 (padding)

	Center      [2]float32 // 36-43 (8 bytes)
	BoundingBox [4]float32 // 44-59 (16 bytes)
	_           [4]byte    // 60-63 (padding to reach 64 bytes)
}

type Region struct {
	Region    string
	SubRegion string
}

type TmpNode struct {
	ID          int64
	Names       map[string]string
	Regions     map[string]Region
	Country     string
	Center      [2]float32
	BoundingBox [4]float32
	Rank        int
	Population  int64
	Timezone    string
}

const NodeSize = 64

func (n *Node) Serialize() []byte {
	var buf [NodeSize]byte

	binary.LittleEndian.PutUint64(buf[0:8], uint64(n.ID))
	binary.LittleEndian.PutUint64(buf[8:16], n.NameOffset)
	binary.LittleEndian.PutUint64(buf[16:24], n.RegionOffset)
	binary.LittleEndian.PutUint32(buf[24:28], n.Population)
	binary.LittleEndian.PutUint16(buf[28:30], n.Rank)
	binary.LittleEndian.PutUint16(buf[30:32], n.Timezone)
	buf[32] = n.Country
	// buf[33:36] is padding (already zero)

	binary.LittleEndian.PutUint32(buf[36:40], math.Float32bits(n.Center[0]))
	binary.LittleEndian.PutUint32(buf[40:44], math.Float32bits(n.Center[1]))

	off := 44
	for i := 0; i < 4; i++ {
		binary.LittleEndian.PutUint32(buf[off:off+4], math.Float32bits(n.BoundingBox[i]))
		off += 4
	}

	return buf[:]
}
