package index

import (
	"encoding/binary"
	"fmt"
	"os"
	"sort"
)

// BPlusTree implements a B+ tree index for item IDs to file offsets
// Key: Item ID (uint32)
// Value: File offset in items.bin (int64)
type BPlusTree struct {
	root  *BPlusNode
	order int // Maximum number of children per node
}

// BPlusNode represents a node in the B+ tree
type BPlusNode struct {
	isLeaf   bool
	keys     []uint32      // Item IDs
	values   []int64       // File offsets (only in leaf nodes)
	children []*BPlusNode  // Child nodes (only in internal nodes)
	next     *BPlusNode    // Pointer to next leaf node (only in leaf nodes)
}

// Entry represents a key-value pair in the index
type Entry struct {
	Key    uint32 // Item ID
	Offset int64  // File offset in items.bin
}

// NewBPlusTree creates a new B+ tree with the specified order
func NewBPlusTree(order int) *BPlusTree {
	if order < 3 {
		order = 4 // Minimum order for B+ tree
	}
	return &BPlusTree{
		root:  newLeafNode(order),
		order: order,
	}
}

// newLeafNode creates a new leaf node
func newLeafNode(order int) *BPlusNode {
	return &BPlusNode{
		isLeaf: true,
		keys:   make([]uint32, 0, order-1),
		values: make([]int64, 0, order-1),
		next:   nil,
	}
}

// newInternalNode creates a new internal node
func newInternalNode(order int) *BPlusNode {
	return &BPlusNode{
		isLeaf:   false,
		keys:     make([]uint32, 0, order-1),
		children: make([]*BPlusNode, 0, order),
	}
}

// Insert adds a new key-value pair to the B+ tree
func (tree *BPlusTree) Insert(key uint32, offset int64) error {
	// If root is full, split it
	if len(tree.root.keys) >= tree.order-1 {
		oldRoot := tree.root
		tree.root = newInternalNode(tree.order)
		tree.root.children = append(tree.root.children, oldRoot)
		tree.splitChild(tree.root, 0)
	}

	return tree.insertNonFull(tree.root, key, offset)
}

// insertNonFull inserts into a node that is not full
func (tree *BPlusTree) insertNonFull(node *BPlusNode, key uint32, offset int64) error {
	if node.isLeaf {
		// Find position to insert
		pos := sort.Search(len(node.keys), func(i int) bool {
			return node.keys[i] >= key
		})

		// Check for duplicate key
		if pos < len(node.keys) && node.keys[pos] == key {
			return fmt.Errorf("duplicate key: %d", key)
		}

		// Insert key and value
		node.keys = append(node.keys, 0)
		node.values = append(node.values, 0)
		copy(node.keys[pos+1:], node.keys[pos:])
		copy(node.values[pos+1:], node.values[pos:])
		node.keys[pos] = key
		node.values[pos] = offset

		return nil
	}

	// Internal node - find child to insert into
	pos := sort.Search(len(node.keys), func(i int) bool {
		return node.keys[i] > key
	})

	child := node.children[pos]

	// Split child if full
	if len(child.keys) >= tree.order-1 {
		tree.splitChild(node, pos)
		if key > node.keys[pos] {
			pos++
		}
	}

	return tree.insertNonFull(node.children[pos], key, offset)
}

// splitChild splits a full child node
func (tree *BPlusTree) splitChild(parent *BPlusNode, childIndex int) {
	child := parent.children[childIndex]
	mid := len(child.keys) / 2

	// Create new node for the right half
	var newNode *BPlusNode
	if child.isLeaf {
		newNode = newLeafNode(tree.order)
		newNode.keys = append(newNode.keys, child.keys[mid:]...)
		newNode.values = append(newNode.values, child.values[mid:]...)
		newNode.next = child.next
		child.next = newNode
		child.keys = child.keys[:mid]
		child.values = child.values[:mid]
	} else {
		newNode = newInternalNode(tree.order)
		newNode.keys = append(newNode.keys, child.keys[mid+1:]...)
		newNode.children = append(newNode.children, child.children[mid+1:]...)
		child.keys = child.keys[:mid]
		child.children = child.children[:mid+1]
	}

	// Insert new key into parent
	parent.keys = append(parent.keys, 0)
	copy(parent.keys[childIndex+1:], parent.keys[childIndex:])
	parent.keys[childIndex] = child.keys[mid]

	// Insert new child into parent
	parent.children = append(parent.children, nil)
	copy(parent.children[childIndex+2:], parent.children[childIndex+1:])
	parent.children[childIndex+1] = newNode
}

// Search finds the file offset for a given item ID
func (tree *BPlusTree) Search(key uint32) (int64, bool) {
	return tree.searchNode(tree.root, key)
}

// searchNode recursively searches for a key in the tree
func (tree *BPlusTree) searchNode(node *BPlusNode, key uint32) (int64, bool) {
	if node.isLeaf {
		// Binary search in leaf node
		pos := sort.Search(len(node.keys), func(i int) bool {
			return node.keys[i] >= key
		})

		if pos < len(node.keys) && node.keys[pos] == key {
			return node.values[pos], true
		}
		return 0, false
	}

	// Internal node - find child to search
	pos := sort.Search(len(node.keys), func(i int) bool {
		return node.keys[i] > key
	})

	return tree.searchNode(node.children[pos], key)
}

// GetAllEntries returns all key-value pairs in sorted order
func (tree *BPlusTree) GetAllEntries() []Entry {
	entries := []Entry{}

	// Find leftmost leaf
	node := tree.root
	for !node.isLeaf {
		node = node.children[0]
	}

	// Traverse leaf nodes
	for node != nil {
		for i := range node.keys {
			entries = append(entries, Entry{
				Key:    node.keys[i],
				Offset: node.values[i],
			})
		}
		node = node.next
	}

	return entries
}

// SaveToFile persists the B+ tree to disk
func (tree *BPlusTree) SaveToFile(filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create index file: %w", err)
	}
	defer file.Close()

	// Write tree order
	if err := binary.Write(file, binary.LittleEndian, uint32(tree.order)); err != nil {
		return fmt.Errorf("failed to write tree order: %w", err)
	}

	// Get all entries in sorted order
	entries := tree.GetAllEntries()

	// Write number of entries
	if err := binary.Write(file, binary.LittleEndian, uint32(len(entries))); err != nil {
		return fmt.Errorf("failed to write entry count: %w", err)
	}

	// Write each entry
	for _, entry := range entries {
		if err := binary.Write(file, binary.LittleEndian, entry.Key); err != nil {
			return fmt.Errorf("failed to write key: %w", err)
		}
		if err := binary.Write(file, binary.LittleEndian, entry.Offset); err != nil {
			return fmt.Errorf("failed to write offset: %w", err)
		}
	}

	return nil
}

// LoadFromFile loads the B+ tree from disk
func LoadFromFile(filePath string) (*BPlusTree, error) {
	file, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist, return empty tree with default order
			return NewBPlusTree(4), nil
		}
		return nil, fmt.Errorf("failed to open index file: %w", err)
	}
	defer file.Close()

	// Read tree order
	var order uint32
	if err := binary.Read(file, binary.LittleEndian, &order); err != nil {
		return nil, fmt.Errorf("failed to read tree order: %w", err)
	}

	tree := NewBPlusTree(int(order))

	// Read number of entries
	var entryCount uint32
	if err := binary.Read(file, binary.LittleEndian, &entryCount); err != nil {
		return nil, fmt.Errorf("failed to read entry count: %w", err)
	}

	// Read and insert each entry
	for i := uint32(0); i < entryCount; i++ {
		var key uint32
		var offset int64

		if err := binary.Read(file, binary.LittleEndian, &key); err != nil {
			return nil, fmt.Errorf("failed to read key: %w", err)
		}
		if err := binary.Read(file, binary.LittleEndian, &offset); err != nil {
			return nil, fmt.Errorf("failed to read offset: %w", err)
		}

		if err := tree.Insert(key, offset); err != nil {
			return nil, fmt.Errorf("failed to insert entry: %w", err)
		}
	}

	return tree, nil
}

// Print returns a string representation of the tree structure (for debugging)
func (tree *BPlusTree) Print() string {
	return tree.printNode(tree.root, 0)
}

// printNode recursively prints the tree structure
func (tree *BPlusTree) printNode(node *BPlusNode, level int) string {
	if node == nil {
		return ""
	}

	result := ""
	indent := ""
	for i := 0; i < level; i++ {
		indent += "  "
	}

	if node.isLeaf {
		result += fmt.Sprintf("%sLeaf: keys=%v, offsets=%v\n", indent, node.keys, node.values)
	} else {
		result += fmt.Sprintf("%sInternal: keys=%v\n", indent, node.keys)
		for _, child := range node.children {
			result += tree.printNode(child, level+1)
		}
	}

	return result
}
