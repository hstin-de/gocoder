package structures

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"sort"
)

type Point struct {
	ID          int64
	Coordinates [2]float32
}

func NewPoint(id int64, coords [2]float32) *Point {
	return &Point{ID: id, Coordinates: coords}
}

func (p *Point) Dimensions() int {
	return len(p.Coordinates)
}

func (p *Point) Dimension(i int) float32 {
	return p.Coordinates[i]
}

type KDTree struct {
	Root *KDNode
}

type KDNode struct {
	Point *Point
	// axis is the dimension this node splits on. We store it to avoid recalculating
	axis  int
	Left  *KDNode
	Right *KDNode
}

func New(points []*Point) *KDTree {
	return &KDTree{
		Root: buildKDTree(points, 0),
	}
}

func buildKDTree(points []*Point, axis int) *KDNode {
	n := len(points)
	if n == 0 {
		return nil
	}
	if n == 1 {
		return &KDNode{Point: points[0], axis: axis}
	}

	sort.Slice(points, func(i, j int) bool {
		return points[i].Dimension(axis) < points[j].Dimension(axis)
	})
	mid := n / 2
	root := points[mid]
	nextAxis := (axis + 1) % root.Dimensions()

	return &KDNode{
		Point: root,
		axis:  axis,
		Left:  buildKDTree(points[:mid], nextAxis),
		Right: buildKDTree(points[mid+1:], nextAxis),
	}
}

func (t *KDTree) Insert(p *Point) {
	if t.Root == nil {
		t.Root = &KDNode{Point: p, axis: 0}
		return
	}
	insertNode(t.Root, p)
}

func insertNode(n *KDNode, p *Point) {
	if p.Dimension(n.axis) < n.Point.Dimension(n.axis) {
		if n.Left == nil {
			n.Left = &KDNode{Point: p, axis: (n.axis + 1) % p.Dimensions()}
		} else {
			insertNode(n.Left, p)
		}
	} else {
		if n.Right == nil {
			n.Right = &KDNode{Point: p, axis: (n.axis + 1) % p.Dimensions()}
		} else {
			insertNode(n.Right, p)
		}
	}
}

func (t *KDTree) Remove(p *Point) *Point {
	if t.Root == nil || p == nil {
		return nil
	}
	removed, newSubtree := removeNode(t.Root, p)
	if removed == t.Root {
		t.Root = newSubtree
	}
	if removed == nil {
		return nil
	}
	return removed.Point
}

func removeNode(n *KDNode, p *Point) (*KDNode, *KDNode) {
	if pointsEqual(n.Point, p) {
		return removeRoot(n)
	}

	if p.Dimension(n.axis) < n.Point.Dimension(n.axis) {
		if n.Left == nil {
			return nil, nil
		}
		removed, newLeft := removeNode(n.Left, p)
		if removed != nil && removed == n.Left {
			n.Left = newLeft
		}
		return removed, nil
	} else {
		if n.Right == nil {
			return nil, nil
		}
		removed, newRight := removeNode(n.Right, p)
		if removed != nil && removed == n.Right {
			n.Right = newRight
		}
		return removed, nil
	}
}

func removeRoot(n *KDNode) (*KDNode, *KDNode) {
	if n.Left == nil && n.Right == nil {
		// Leaf
		return n, nil
	}
	if n.Right != nil {
		smallestParent, smallest := findSmallest(n.Right, n.axis, nil)
		if smallestParent == nil {
			n.Right = removeChildRoot(n.Right)
		} else {
			if smallestParent.Left == smallest {
				smallestParent.Left = removeChildRoot(smallest)
			} else {
				smallestParent.Right = removeChildRoot(smallest)
			}
		}
		smallest.Left = n.Left
		smallest.Right = n.Right
		return n, smallest
	}
	largestParent, largest := findLargest(n.Left, n.axis, nil)
	if largestParent == nil {
		n.Left = removeChildRoot(n.Left)
	} else {
		if largestParent.Left == largest {
			largestParent.Left = removeChildRoot(largest)
		} else {
			largestParent.Right = removeChildRoot(largest)
		}
	}
	largest.Left = n.Left
	largest.Right = n.Right
	return n, largest
}

func removeChildRoot(n *KDNode) *KDNode {
	if n.Left == nil && n.Right == nil {
		return nil
	}
	if n.Right != nil {
		return n.Right
	}
	return n.Left
}

func findSmallest(n *KDNode, axis int, parent *KDNode) (*KDNode, *KDNode) {
	if n == nil {
		return nil, nil
	}
	if n.axis == axis {
		// keep going left if possible
		if n.Left == nil {
			return parent, n
		}
		return findSmallest(n.Left, axis, n)
	}
	// axis mismatch, we look both sides
	p1, s1 := findSmallest(n.Left, axis, n)
	p2, s2 := findSmallest(n.Right, axis, n)

	minNode := n
	minPar := parent
	if p1 != nil && s1 != nil && s1.Point.Dimension(axis) < minNode.Point.Dimension(axis) {
		minNode = s1
		minPar = p1
	}
	if p2 != nil && s2 != nil && s2.Point.Dimension(axis) < minNode.Point.Dimension(axis) {
		minNode = s2
		minPar = p2
	}
	if n.Point.Dimension(axis) < minNode.Point.Dimension(axis) {
		minNode = n
		minPar = parent
	}
	return minPar, minNode
}

func findLargest(n *KDNode, axis int, parent *KDNode) (*KDNode, *KDNode) {
	if n == nil {
		return nil, nil
	}
	if n.axis == axis {
		if n.Right == nil {
			return parent, n
		}
		return findLargest(n.Right, axis, n)
	}
	p1, s1 := findLargest(n.Left, axis, n)
	p2, s2 := findLargest(n.Right, axis, n)

	maxNode := n
	maxPar := parent
	if p1 != nil && s1 != nil && s1.Point.Dimension(axis) > maxNode.Point.Dimension(axis) {
		maxNode = s1
		maxPar = p1
	}
	if p2 != nil && s2 != nil && s2.Point.Dimension(axis) > maxNode.Point.Dimension(axis) {
		maxNode = s2
		maxPar = p2
	}
	if n.Point.Dimension(axis) > maxNode.Point.Dimension(axis) {
		maxNode = n
		maxPar = parent
	}
	return maxPar, maxNode
}

func pointsEqual(a, b *Point) bool {
	if a.ID != b.ID {
		return false
	}
	for i := 0; i < a.Dimensions(); i++ {
		if a.Dimension(i) != b.Dimension(i) {
			return false
		}
	}
	return true
}

func (t *KDTree) Balance() {
	allPoints := t.Points()
	t.Root = buildKDTree(allPoints, 0)
}

func (t *KDTree) Points() []*Point {
	if t.Root == nil {
		return nil
	}
	return collectPoints(t.Root)
}

func collectPoints(n *KDNode) []*Point {
	if n == nil {
		return nil
	}
	out := collectPoints(n.Left)
	out = append(out, n.Point)
	out = append(out, collectPoints(n.Right)...)
	return out
}

func (t *KDTree) String() string {
	return fmt.Sprintf("[%s]", printTreeKDNode(t.Root))
}

func printTreeKDNode(n *KDNode) string {
	if n == nil {
		return "-"
	}
	if n.Left == nil && n.Right == nil {
		return nodeString(n)
	}
	return fmt.Sprintf("(%s %s %s)", printTreeKDNode(n.Left), nodeString(n), printTreeKDNode(n.Right))
}

func nodeString(n *KDNode) string {
	return fmt.Sprintf("ID=%d,axis=%d,(%.2f,%.2f)",
		n.Point.ID, n.axis, n.Point.Coordinates[0], n.Point.Coordinates[1])
}

// -------------------------------------------------------------------
// KNN
// -------------------------------------------------------------------

// KNN returns the k-nearest neighbors of point p, sorted by distance ascending.
func (t *KDTree) KNN(p *Point, k int) []*Point {
	if t.Root == nil || p == nil || k <= 0 {
		return nil
	}
	neighbors := make([]*KDNode, 0, k)
	knnSearch(p, k, t.Root, &neighbors)
	out := make([]*Point, len(neighbors))
	for i, nn := range neighbors {
		out[i] = nn.Point
	}
	return out
}

// knnSearch does a DFS search in the KDTree, tracking up to k best nodes.
func knnSearch(p *Point, k int, node *KDNode, best *[]*KDNode) {
	if node == nil {
		return
	}

	// 1) "Visit" node
	insertNeighbor(p, node, k, best)

	// 2) Determine search path
	leftFirst := (p.Dimension(node.axis) < node.Point.Dimension(node.axis))
	first, second := node.Left, node.Right
	if !leftFirst {
		first, second = node.Right, node.Left
	}

	knnSearch(p, k, first, best)

	// 3) Check if we need to search the other side
	// planeDistance = difference in the splitting dimension
	axisDist := float32(math.Abs(float64(node.Point.Dimension(node.axis) - p.Dimension(node.axis))))
	if len(*best) < k || axisDist*axisDist < distSquared(p, (*best)[len(*best)-1].Point) {
		knnSearch(p, k, second, best)
	}
}

// insertNeighbor tries to insert 'node' into 'best' (up to k neighbors),
// keeping them sorted by ascending distance from p.
func insertNeighbor(p *Point, node *KDNode, k int, best *[]*KDNode) {
	d := distSquared(p, node.Point)
	if len(*best) < k {
		*best = append(*best, node)
		sort.Slice(*best, func(i, j int) bool {
			return distSquared(p, (*best)[i].Point) < distSquared(p, (*best)[j].Point)
		})
	} else {
		// compare with worst so far
		worstDist := distSquared(p, (*best)[k-1].Point)
		if d < worstDist {
			(*best)[k-1] = node
			// re-sort
			sort.Slice(*best, func(i, j int) bool {
				return distSquared(p, (*best)[i].Point) < distSquared(p, (*best)[j].Point)
			})
		}
	}
}

func distSquared(a, b *Point) float32 {
	dx := a.Dimension(0) - b.Dimension(0)
	dy := a.Dimension(1) - b.Dimension(1)
	return dx*dx + dy*dy
}

// -------------------------------------------------------------------
// Save / Load (Custom Binary Format)
// -------------------------------------------------------------------
/*
We'll do a simple **preorder** DFS serialization:

For each node:
1. Write a 1 byte flag: 0 = nil node, 1 = non-nil
2. If non-nil:
   2.1. Write n.Point.ID (int64)
   2.2. Write n.Point.Coordinates[2] (two float32s)
   2.3. Write n.axis (int32)
   2.4. Recursively write n.Left
   2.5. Recursively write n.Right
*/

func (t *KDTree) Save(w io.Writer) error {
	writer := bufio.NewWriter(w)
	if err := writeKDNode(t.Root, writer); err != nil {
		return err
	}
	return writer.Flush()
}

func writeKDNode(n *KDNode, w io.Writer) error {
	// 1. node flag
	if n == nil {
		// write 0 = nil
		if err := binary.Write(w, binary.LittleEndian, byte(0)); err != nil {
			return err
		}
		return nil
	}
	// non-nil
	if err := binary.Write(w, binary.LittleEndian, byte(1)); err != nil {
		return err
	}

	// 2.1. node.Point.ID (int64)
	if err := binary.Write(w, binary.LittleEndian, n.Point.ID); err != nil {
		return err
	}
	// 2.2. node.Point.Coordinates[2] (float32 each)
	if err := binary.Write(w, binary.LittleEndian, n.Point.Coordinates[0]); err != nil {
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, n.Point.Coordinates[1]); err != nil {
		return err
	}
	// 2.3. node.axis (int32)
	if err := binary.Write(w, binary.LittleEndian, int32(n.axis)); err != nil {
		return err
	}

	// 2.4. left subtree
	if err := writeKDNode(n.Left, w); err != nil {
		return err
	}
	// 2.5. right subtree
	if err := writeKDNode(n.Right, w); err != nil {
		return err
	}

	return nil
}

func (t *KDTree) Load(data []byte) error {
	bytesReader := bytes.NewReader(data)
	return t.LoadFromReader(bytesReader)
}

func (t *KDTree) LoadFromReader(r io.Reader) error {
	reader := bufio.NewReaderSize(r, 64*1024) // 64KB buffer

	root, err := readNode(reader)
	if err != nil {
		return err
	}

	t.Root = root
	return nil
}

func (t *KDTree) LoadFromFile(file io.ReaderAt, offset int64, size uint64) error {
	reader := io.NewSectionReader(file, offset, int64(size))
	return t.LoadFromReader(reader)
}

func readNode(r io.Reader) (*KDNode, error) {
	// 1. node flag
	var flag byte
	if err := binary.Read(r, binary.LittleEndian, &flag); err != nil {
		return nil, err
	}
	if flag == 0 {
		// nil
		return nil, nil
	}

	// non-nil: read all fields
	var id int64
	if err := binary.Read(r, binary.LittleEndian, &id); err != nil {
		return nil, err
	}
	var x, y float32
	if err := binary.Read(r, binary.LittleEndian, &x); err != nil {
		return nil, err
	}
	if err := binary.Read(r, binary.LittleEndian, &y); err != nil {
		return nil, err
	}

	var axis int32
	if err := binary.Read(r, binary.LittleEndian, &axis); err != nil {
		return nil, err
	}

	leftNode, err := readNode(r)
	if err != nil {
		return nil, err
	}
	rightNode, err := readNode(r)
	if err != nil {
		return nil, err
	}

	return &KDNode{
		Point: &Point{
			ID:          id,
			Coordinates: [2]float32{x, y},
		},
		axis:  int(axis),
		Left:  leftNode,
		Right: rightNode,
	}, nil
}
