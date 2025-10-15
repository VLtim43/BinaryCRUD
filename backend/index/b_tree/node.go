package b_tree

import "fmt"

// BPlusNode represents a node in the B+ tree
// Internal nodes: store keys and child pointers for navigation
// Leaf nodes: store key-value pairs (RecordID → FileOffset) and link to next leaf
type BPlusNode struct {
	IsLeaf   bool          // true if this is a leaf node
	Keys     []uint32      // sorted keys (RecordIDs)
	Values   []int64       // file offsets (only used in leaf nodes)
	Children []*BPlusNode  // child pointers (only used in internal nodes)
	Next     *BPlusNode    // pointer to next leaf node (only used in leaf nodes)
	Parent   *BPlusNode    // pointer to parent node (for easier insertion)
}

// NewLeafNode creates a new leaf node
func NewLeafNode() *BPlusNode {
	return &BPlusNode{
		IsLeaf:   true,
		Keys:     make([]uint32, 0),
		Values:   make([]int64, 0),
		Children: nil,
		Next:     nil,
		Parent:   nil,
	}
}

// NewInternalNode creates a new internal node
func NewInternalNode() *BPlusNode {
	return &BPlusNode{
		IsLeaf:   false,
		Keys:     make([]uint32, 0),
		Values:   nil,
		Children: make([]*BPlusNode, 0),
		Next:     nil,
		Parent:   nil,
	}
}

// InsertIntoLeaf inserts a key-value pair into a leaf node
// Maintains sorted order of keys
func (n *BPlusNode) InsertIntoLeaf(key uint32, value int64) {
	if !n.IsLeaf {
		panic("InsertIntoLeaf called on non-leaf node")
	}

	// Find insertion position (binary search could optimize this)
	pos := 0
	for pos < len(n.Keys) && n.Keys[pos] < key {
		pos++
	}

	// Check if key already exists (update value)
	if pos < len(n.Keys) && n.Keys[pos] == key {
		n.Values[pos] = value
		return
	}

	// Insert at position
	n.Keys = append(n.Keys[:pos], append([]uint32{key}, n.Keys[pos:]...)...)
	n.Values = append(n.Values[:pos], append([]int64{value}, n.Values[pos:]...)...)
}

// InsertIntoInternal inserts a key and child pointer into an internal node
func (n *BPlusNode) InsertIntoInternal(key uint32, child *BPlusNode) {
	if n.IsLeaf {
		panic("InsertIntoInternal called on leaf node")
	}

	// Find insertion position
	pos := 0
	for pos < len(n.Keys) && n.Keys[pos] < key {
		pos++
	}

	// Insert key
	n.Keys = append(n.Keys[:pos], append([]uint32{key}, n.Keys[pos:]...)...)
	// Insert child pointer (always one more child than keys)
	n.Children = append(n.Children[:pos+1], append([]*BPlusNode{child}, n.Children[pos+1:]...)...)

	// Set parent pointer
	child.Parent = n
}

// IsFull checks if node needs to be split (based on order)
func (n *BPlusNode) IsFull(order int) bool {
	return len(n.Keys) >= order
}

// SplitLeaf splits a full leaf node into two nodes
// Returns: new right node and the key to push up to parent
func (n *BPlusNode) SplitLeaf() (*BPlusNode, uint32) {
	if !n.IsLeaf {
		panic("SplitLeaf called on non-leaf node")
	}

	mid := len(n.Keys) / 2

	// Create new right node
	right := NewLeafNode()
	right.Keys = append([]uint32{}, n.Keys[mid:]...)
	right.Values = append([]int64{}, n.Values[mid:]...)
	right.Next = n.Next
	right.Parent = n.Parent

	// Update current (left) node
	n.Keys = n.Keys[:mid]
	n.Values = n.Values[:mid]
	n.Next = right

	// Key to push up is the first key of right node
	return right, right.Keys[0]
}

// SplitInternal splits a full internal node into two nodes
// Returns: new right node and the key to push up to parent
func (n *BPlusNode) SplitInternal() (*BPlusNode, uint32) {
	if n.IsLeaf {
		panic("SplitInternal called on leaf node")
	}

	mid := len(n.Keys) / 2
	pushUpKey := n.Keys[mid]

	// Create new right node
	right := NewInternalNode()
	right.Keys = append([]uint32{}, n.Keys[mid+1:]...)
	right.Children = append([]*BPlusNode{}, n.Children[mid+1:]...)
	right.Parent = n.Parent

	// Update parent pointers for moved children
	for _, child := range right.Children {
		child.Parent = right
	}

	// Update current (left) node
	n.Keys = n.Keys[:mid]
	n.Children = n.Children[:mid+1]

	return right, pushUpKey
}

// Search finds a value by key in a leaf node
// Returns: value (file offset), found (bool)
func (n *BPlusNode) Search(key uint32) (int64, bool) {
	if !n.IsLeaf {
		panic("Search called on non-leaf node")
	}

	for i, k := range n.Keys {
		if k == key {
			return n.Values[i], true
		}
	}
	return 0, false
}

// FindChild returns the child pointer to follow for a given key
func (n *BPlusNode) FindChild(key uint32) *BPlusNode {
	if n.IsLeaf {
		panic("FindChild called on leaf node")
	}

	// Find the appropriate child
	pos := 0
	for pos < len(n.Keys) && key >= n.Keys[pos] {
		pos++
	}

	return n.Children[pos]
}

// String returns a string representation of the node (for debugging)
func (n *BPlusNode) String() string {
	if n.IsLeaf {
		return fmt.Sprintf("Leaf[keys=%v, values=%v]", n.Keys, n.Values)
	}
	return fmt.Sprintf("Internal[keys=%v, children=%d]", n.Keys, len(n.Children))
}
