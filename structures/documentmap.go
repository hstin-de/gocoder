package structures

import (
	"bytes"
	"encoding/binary"
	"io"
)

type DocumentMap map[int64]int32

func (dm *DocumentMap) Save(w io.Writer) error {
	count := int64(len(*dm))
	if err := binary.Write(w, binary.LittleEndian, count); err != nil {
		return err
	}
	for key, value := range *dm {
		if err := binary.Write(w, binary.LittleEndian, key); err != nil {
			return err
		}
		if err := binary.Write(w, binary.LittleEndian, value); err != nil {
			return err
		}
	}
	return nil
}

func (dm *DocumentMap) Load(data []byte) error {
	br := bytes.NewReader(data)
	return dm.LoadFromReader(br)
}

func (dm *DocumentMap) LoadFromReader(r io.Reader) error {
	var count int64
	if err := binary.Read(r, binary.LittleEndian, &count); err != nil {
		return err
	}

	*dm = make(DocumentMap, count)

	for i := int64(0); i < count; i++ {
		var key int64
		var value int32
		if err := binary.Read(r, binary.LittleEndian, &key); err != nil {
			return err
		}
		if err := binary.Read(r, binary.LittleEndian, &value); err != nil {
			return err
		}
		(*dm)[key] = value
	}
	return nil
}

func (dm *DocumentMap) LoadFromFile(file io.ReaderAt, offset int64, size uint64) error {
	reader := io.NewSectionReader(file, offset, int64(size))
	return dm.LoadFromReader(reader)
}
