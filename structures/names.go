package structures

import (
	"bytes"
	"encoding/binary"
	"log"
	"sync"
)

// Names manages a memory-mapped (or file-backed) structure that stores arrays of strings.
// The structure on disk:
//
// [4 bytes: number_of_strings (uint32)]
// For each string:
//
//	[2 bytes: length_of_string (uint16)]
//	[N bytes: string_data (UTF-8)]
//
// This format allows easy iteration over stored strings and retrieval by offset.
type Names struct {
	buffer *bytes.Buffer
	offset int64
	mu     sync.Mutex
}

func NewNames() (*Names, error) {
	f := bytes.NewBuffer(nil)
	off := int64(0)
	return &Names{buffer: f, offset: off}, nil
}

func (n *Names) Bytes() []byte {
	return n.buffer.Bytes()
}

func (n *Names) Store(strings []string) (int64, error) {
	n.mu.Lock()
	defer n.mu.Unlock()

	startOffset := n.offset

	// Write number_of_strings (uint32)
	countBuf := make([]byte, 4)
	binary.LittleEndian.PutUint32(countBuf, uint32(len(strings)))
	if _, err := n.buffer.Write(countBuf); err != nil {
		return 0, err
	}
	n.offset += 4

	// Write each string
	for _, s := range strings {
		strBytes := []byte(s)
		length := len(strBytes)
		if length > 0xFFFF {
			return 0, log.Output(2, "String too long, exceeds 65535 bytes")
		}

		// Write length (uint16)
		lengthBuf := make([]byte, 2)
		binary.LittleEndian.PutUint16(lengthBuf, uint16(length))
		if _, err := n.buffer.Write(lengthBuf); err != nil {
			return 0, err
		}
		n.offset += 2

		// Write string data
		if _, err := n.buffer.Write(strBytes); err != nil {
			return 0, err
		}
		n.offset += int64(length)
	}

	return startOffset, nil
}

func (n *Names) Read(offset int64) ([]string, error) {
	n.mu.Lock()
	defer n.mu.Unlock()

	reader := bytes.NewReader(n.buffer.Bytes()[offset:])

	// Read number_of_strings (uint32)
	countBuf := make([]byte, 4)
	if _, err := reader.Read(countBuf); err != nil {
		return nil, err
	}
	numberOfStrings := binary.LittleEndian.Uint32(countBuf)

	result := make([]string, 0, numberOfStrings)
	for i := uint32(0); i < numberOfStrings; i++ {
		// Read length (uint16)
		lengthBuf := make([]byte, 2)
		if _, err := reader.Read(lengthBuf); err != nil {
			return nil, err
		}
		length := binary.LittleEndian.Uint16(lengthBuf)

		// Read string data
		strBytes := make([]byte, length)
		if _, err := reader.Read(strBytes); err != nil {
			return nil, err
		}

		result = append(result, string(strBytes))
	}

	return result, nil
}
