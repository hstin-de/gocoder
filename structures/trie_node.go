package structures

import (
	"sort"
)

type ChildPair struct {
	Char rune
	Node *TrieNode
}

type TrieNode struct {
	Char     rune
	Children []ChildPair
	IsEnd    bool
	Docs     []int64
}

func newTrieNode(r rune) *TrieNode {
	return &TrieNode{
		Char:     r,
		Children: make([]ChildPair, 0),
		Docs:     make([]int64, 0),
	}
}

func (n *TrieNode) getChild(r rune) *TrieNode {
	idx := sort.Search(len(n.Children), func(i int) bool {
		return n.Children[i].Char >= r
	})
	if idx < len(n.Children) && n.Children[idx].Char == r {
		return n.Children[idx].Node
	}
	return nil
}

func (n *TrieNode) insertChild(r rune) *TrieNode {
	idx := sort.Search(len(n.Children), func(i int) bool {
		return n.Children[i].Char >= r
	})
	// If child already exists, just return it
	if idx < len(n.Children) && n.Children[idx].Char == r {
		return n.Children[idx].Node
	}
	// Otherwise, create a new node and insert
	newNode := newTrieNode(r)
	n.Children = append(n.Children, ChildPair{}) // grow the slice by 1
	copy(n.Children[idx+1:], n.Children[idx:])
	n.Children[idx] = ChildPair{Char: r, Node: newNode}
	return newNode
}

// -----------------------------------------------------
// Serialization / Deserialization Helpers
// -----------------------------------------------------

type TrieNodeData struct {
	Char     rune           `json:"char"`
	IsEnd    bool           `json:"isEnd"`
	Docs     []int64        `json:"docs,omitempty"`
	Children []TrieNodeData `json:"children,omitempty"`
}

func (n *TrieNode) toData() TrieNodeData {
	childData := make([]TrieNodeData, len(n.Children))
	for i, pair := range n.Children {
		childData[i] = pair.Node.toData()
	}
	return TrieNodeData{
		Char:     n.Char,
		IsEnd:    n.IsEnd,
		Docs:     append([]int64(nil), n.Docs...), // copy to be safe
		Children: childData,
	}
}

func fromData(data TrieNodeData) *TrieNode {
	node := newTrieNode(data.Char)
	node.IsEnd = data.IsEnd
	node.Docs = data.Docs
	node.Children = make([]ChildPair, len(data.Children))
	for i, cData := range data.Children {
		childNode := fromData(cData)
		node.Children[i] = ChildPair{
			Char: cData.Char,
			Node: childNode,
		}
	}
	return node
}
