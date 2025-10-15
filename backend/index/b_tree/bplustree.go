package b_tree

import "fmt"

// BPlusTree represents a B+ tree index structure
// Order: maximum number of keys per node
// Root: pointer to the root node
type BPlusTree struct {
	Root  *BPlusNode
	Order int // max keys per node (typically 3-5 for demonstration, higher in production)
}

// NewBPlusTree creates a new B+ tree with the given order
func NewBPlusTree(order int) *BPlusTree {
	if order < 3 {
		panic("B+ tree order must be at least 3")
	}

	return &BPlusTree{
		Root:  NewLeafNode(),
		Order: order,
	}
}

// Insert adds a key-value pair to the tree
// Key: RecordID (uint32)
// Value: File offset (int64)
func (t *BPlusTree) Insert(key uint32, value int64) {
	fmt.Printf("[BTREE] Inserting key=%d, value=%d\n", key, value)

	// Find the leaf node where this key should go
	leaf := t.findLeaf(key)

	// Insert into leaf
	leaf.InsertIntoLeaf(key, value)

	// Check if leaf needs to split
	if leaf.IsFull(t.Order) {
		t.splitLeafAndPropagate(leaf)
	}
}

// Search finds a value by key
// Returns: file offset, found (bool)
func (t *BPlusTree) Search(key uint32) (int64, bool) {
	leaf := t.findLeaf(key)
	return leaf.Search(key)
}

// RangeSearch returns all key-value pairs in the range [startKey, endKey]
// Returns slice of (key, value) pairs
func (t *BPlusTree) RangeSearch(startKey, endKey uint32) [][2]interface{} {
	results := make([][2]interface{}, 0)

	// Find the leaf containing startKey
	leaf := t.findLeaf(startKey)

	// Traverse leaf nodes using Next pointers
	for leaf != nil {
		for i, k := range leaf.Keys {
			if k >= startKey && k <= endKey {
				results = append(results, [2]interface{}{k, leaf.Values[i]})
			}
			if k > endKey {
				return results
			}
		}
		leaf = leaf.Next
	}

	return results
}

// findLeaf navigates from root to the appropriate leaf node for a key
func (t *BPlusTree) findLeaf(key uint32) *BPlusNode {
	node := t.Root

	// Navigate down the tree
	for !node.IsLeaf {
		node = node.FindChild(key)
	}

	return node
}

// splitLeafAndPropagate handles leaf split and propagates changes up the tree
func (t *BPlusTree) splitLeafAndPropagate(leaf *BPlusNode) {
	fmt.Printf("[BTREE] Splitting leaf node: %v\n", leaf)

	// Split the leaf
	rightLeaf, pushUpKey := leaf.SplitLeaf()

	// If leaf has no parent, create new root
	if leaf.Parent == nil {
		fmt.Printf("[BTREE] Creating new root with key=%d\n", pushUpKey)
		newRoot := NewInternalNode()
		newRoot.Keys = []uint32{pushUpKey}
		newRoot.Children = []*BPlusNode{leaf, rightLeaf}
		leaf.Parent = newRoot
		rightLeaf.Parent = newRoot
		t.Root = newRoot
		return
	}

	// Insert into parent
	parent := leaf.Parent
	parent.InsertIntoInternal(pushUpKey, rightLeaf)

	// Check if parent needs to split
	if parent.IsFull(t.Order) {
		t.splitInternalAndPropagate(parent)
	}
}

// splitInternalAndPropagate handles internal node split and propagates up
func (t *BPlusTree) splitInternalAndPropagate(node *BPlusNode) {
	fmt.Printf("[BTREE] Splitting internal node: %v\n", node)

	// Split the internal node
	rightNode, pushUpKey := node.SplitInternal()

	// If node has no parent, create new root
	if node.Parent == nil {
		fmt.Printf("[BTREE] Creating new root with key=%d\n", pushUpKey)
		newRoot := NewInternalNode()
		newRoot.Keys = []uint32{pushUpKey}
		newRoot.Children = []*BPlusNode{node, rightNode}
		node.Parent = newRoot
		rightNode.Parent = newRoot
		t.Root = newRoot
		return
	}

	// Insert into parent
	parent := node.Parent
	parent.InsertIntoInternal(pushUpKey, rightNode)

	// Check if parent needs to split
	if parent.IsFull(t.Order) {
		t.splitInternalAndPropagate(parent)
	}
}

// PrintTree prints the tree structure (for debugging)
func (t *BPlusTree) PrintTree() {
	fmt.Println("\n=== B+ Tree Structure ===")
	t.printNode(t.Root, 0)
	fmt.Println("=========================\n")
}

// printNode recursively prints tree structure
func (t *BPlusTree) printNode(node *BPlusNode, level int) {
	indent := ""
	for i := 0; i < level; i++ {
		indent += "  "
	}

	if node.IsLeaf {
		fmt.Printf("%sLeaf: keys=%v, values=%v\n", indent, node.Keys, node.Values)
	} else {
		fmt.Printf("%sInternal: keys=%v\n", indent, node.Keys)
		for _, child := range node.Children {
			t.printNode(child, level+1)
		}
	}
}

// GetAllLeafEntries returns all entries by traversing leaf nodes (for serialization)
func (t *BPlusTree) GetAllLeafEntries() [][2]interface{} {
	entries := make([][2]interface{}, 0)

	// Find leftmost leaf
	node := t.Root
	for !node.IsLeaf {
		node = node.Children[0]
	}

	// Traverse all leaf nodes
	for node != nil {
		for i := range node.Keys {
			entries = append(entries, [2]interface{}{node.Keys[i], node.Values[i]})
		}
		node = node.Next
	}

	return entries
}

// Count returns the total number of entries in the tree
func (t *BPlusTree) Count() int {
	return len(t.GetAllLeafEntries())
}
