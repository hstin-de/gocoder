package structures

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
	"sync"
)

type Trie struct {
	Root  *TrieNode
	mutex sync.RWMutex
}

func NewTrie() *Trie {
	return &Trie{
		Root: newTrieNode(rune(0)), // sentinel char for the root
	}
}

func (t *Trie) Insert(docID int64, text string) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	node := t.Root
	text = strings.ToLower(text)
	for _, char := range text {
		node = node.insertChild(char)
	}
	node.IsEnd = true

	// Keep docs sorted and unique
	pos := sort.Search(len(node.Docs), func(i int) bool { return node.Docs[i] >= docID })
	if pos == len(node.Docs) {
		node.Docs = append(node.Docs, docID)
	} else if node.Docs[pos] != docID {
		node.Docs = append(node.Docs, 0)
		copy(node.Docs[pos+1:], node.Docs[pos:])
		node.Docs[pos] = docID
	}
}

// Search returns a slice of doc IDs that match a given prefix
func (t *Trie) Search(prefix string) []int64 {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	node := t.Root
	prefix = strings.ToLower(prefix)

	for _, char := range prefix {
		child := node.getChild(char)
		if child == nil {
			return nil
		}
		node = child
	}
	return collectDocs(node)
}

func collectDocs(node *TrieNode) []int64 {
	var result []int64
	if node.IsEnd {
		result = append(result, node.Docs...)
	}
	for _, pair := range node.Children {
		result = append(result, collectDocs(pair.Node)...)
	}
	return result
}

// -----------------------------------------------------
// Serialize / Deserialize the entire Trie
// -----------------------------------------------------

func (t *Trie) Serialize() ([]byte, error) {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	rootData := t.Root.toData()
	return json.Marshal(rootData)
}

func (t *Trie) Deserialize(data []byte) error {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	var rootData TrieNodeData
	if err := json.Unmarshal(data, &rootData); err != nil {
		return err
	}
	t.Root = fromData(rootData)
	return nil
}

func (t *Trie) Save(w io.Writer) error {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	if err := WriteNode(t.Root, w); err != nil {
		return fmt.Errorf("failed to write trie: %w", err)
	}
	return nil
}

func (t *Trie) Load(data []byte) error {
	br := bytes.NewReader(data)
	return t.LoadFromReader(br)
}

func (t *Trie) LoadFromReader(r io.Reader) error {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	bufferedReader := bufio.NewReaderSize(r, 256*1024) // 256 KB buffer

	rootNode, err := readNodeFast(bufferedReader)
	if err != nil {
		return fmt.Errorf("failed to read trie: %w", err)
	}

	t.Root = rootNode
	return nil
}

func (t *Trie) LoadFromFile(file io.ReaderAt, offset int64, size uint64) error {
	reader := io.NewSectionReader(file, offset, int64(size))
	return t.LoadFromReader(reader)
}

func WriteNode(node *TrieNode, w io.Writer) error {
	// 1. Write node.char (int32)
	if err := writeInt32(w, int32(node.Char)); err != nil {
		return err
	}

	// 2. Write node.isEnd (1 byte: 0 or 1)
	var endFlag byte
	if node.IsEnd {
		endFlag = 1
	}
	if err := writeByte(w, endFlag); err != nil {
		return err
	}

	// 3. Write len(docs) (uint64)
	if err := writeUint64(w, uint64(len(node.Docs))); err != nil {
		return err
	}

	// 4. Write each doc (int64)
	for _, docID := range node.Docs {
		if err := writeInt64(w, docID); err != nil {
			return err
		}
	}

	// 5. Write number of children (uint64)
	if err := writeUint64(w, uint64(len(node.Children))); err != nil {
		return err
	}

	// 6. Recursively write each child
	for _, c := range node.Children {
		if err := WriteNode(c.Node, w); err != nil {
			return err
		}
	}
	return nil
}

func readNodeFast(r io.Reader) (*TrieNode, error) {
	// 1. Read char (int32)
	charVal, err := readInt32(r)
	if err != nil {
		return nil, err
	}

	// 2. Read isEnd (1 byte)
	endFlag, err := readByte(r)
	if err != nil {
		return nil, err
	}
	isEnd := (endFlag == 1)

	// 3. Read len(docs) (uint64)
	docsLen, err := readUint64(r)
	if err != nil {
		return nil, err
	}

	// 4. Read each doc (int64)
	docs := make([]int64, docsLen)
	for i := uint64(0); i < docsLen; i++ {
		docID, err := readInt64(r)
		if err != nil {
			return nil, err
		}
		docs[i] = docID
	}

	// 5. Read number of children (uint64)
	childCount, err := readUint64(r)
	if err != nil {
		return nil, err
	}

	// 6. Recursively read each child
	children := make([]ChildPair, 0, childCount)
	for i := uint64(0); i < childCount; i++ {
		childNode, err := readNodeFast(r)
		if err != nil {
			return nil, err
		}
		children = append(children, ChildPair{
			Char: childNode.Char,
			Node: childNode,
		})
	}

	// Construct the node
	node := &TrieNode{
		Char:     rune(charVal),
		IsEnd:    isEnd,
		Docs:     docs,
		Children: children,
	}
	return node, nil
}

// ------------------ Non-reflective Reading Helpers ------------------

func readInt32(r io.Reader) (int32, error) {
	var buf [4]byte
	if _, err := io.ReadFull(r, buf[:]); err != nil {
		return 0, err
	}
	return int32(binary.LittleEndian.Uint32(buf[:])), nil
}

func readByte(r io.Reader) (byte, error) {
	var b [1]byte
	if _, err := io.ReadFull(r, b[:]); err != nil {
		return 0, err
	}
	return b[0], nil
}

func readUint64(r io.Reader) (uint64, error) {
	var buf [8]byte
	if _, err := io.ReadFull(r, buf[:]); err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint64(buf[:]), nil
}

func readInt64(r io.Reader) (int64, error) {
	uval, err := readUint64(r)
	return int64(uval), err
}

// ------------------ Non-reflective Writing Helpers ------------------

func writeInt32(w io.Writer, val int32) error {
	var buf [4]byte
	binary.LittleEndian.PutUint32(buf[:], uint32(val))
	_, err := w.Write(buf[:])
	return err
}

func writeByte(w io.Writer, val byte) error {
	_, err := w.Write([]byte{val})
	return err
}

func writeUint64(w io.Writer, val uint64) error {
	var buf [8]byte
	binary.LittleEndian.PutUint64(buf[:], val)
	_, err := w.Write(buf[:])
	return err
}

func writeInt64(w io.Writer, val int64) error {
	var buf [8]byte
	binary.LittleEndian.PutUint64(buf[:], uint64(val))
	_, err := w.Write(buf[:])
	return err
}
