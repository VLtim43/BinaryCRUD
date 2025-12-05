package index

import (
	"fmt"
	"sort"
)

// BTree is a simple B+ tree that maps ID (uint64) to file offset (int64)
type BTree struct {
	root  *BNode
	order int // Max keys per node
}

// BNode represents a node in the B+ tree
type BNode struct {
	isLeaf   bool
	keys     []uint64  // IDs
	offsets  []int64   // File offsets (only in leaf nodes)
	children []*BNode  // Child pointers (only in internal nodes)
	next     *BNode    // Next leaf pointer (only in leaf nodes)
}

// NewBTree creates a new B+ tree
func NewBTree(order int) *BTree {
	if order < 3 {
		order = 4
	}
	return &BTree{
		root:  newLeaf(),
		order: order,
	}
}

// newLeaf creates a new leaf node
func newLeaf() *BNode {
	return &BNode{
		isLeaf:  true,
		keys:    make([]uint64, 0),
		offsets: make([]int64, 0),
		next:    nil,
	}
}

// newInternal creates a new internal node
func newInternal() *BNode {
	return &BNode{
		isLeaf:   false,
		keys:     make([]uint64, 0),
		children: make([]*BNode, 0),
	}
}

// Insert adds a key-offset pair to the tree
func (t *BTree) Insert(id uint64, offset int64) error {
	// Split root if full
	if len(t.root.keys) >= t.order-1 {
		newRoot := newInternal()
		newRoot.children = append(newRoot.children, t.root)
		t.splitChild(newRoot, 0)
		t.root = newRoot
	}

	return t.insertNonFull(t.root, id, offset)
}

// insertNonFull inserts into a node that's not full
func (t *BTree) insertNonFull(node *BNode, id uint64, offset int64) error {
	if node.isLeaf {
		// Find insertion position
		pos := sort.Search(len(node.keys), func(i int) bool {
			return node.keys[i] >= id
		})

		// Check duplicate
		if pos < len(node.keys) && node.keys[pos] == id {
			return fmt.Errorf("duplicate ID: %d", id)
		}

		// Insert at position
		node.keys = append(node.keys, 0)
		node.offsets = append(node.offsets, 0)
		copy(node.keys[pos+1:], node.keys[pos:])
		copy(node.offsets[pos+1:], node.offsets[pos:])
		node.keys[pos] = id
		node.offsets[pos] = offset

		return nil
	}

	// Find child to insert into
	pos := sort.Search(len(node.keys), func(i int) bool {
		return node.keys[i] > id
	})

	child := node.children[pos]

	// Split child if full
	if len(child.keys) >= t.order-1 {
		t.splitChild(node, pos)
		if id > node.keys[pos] {
			pos++
		}
	}

	return t.insertNonFull(node.children[pos], id, offset)
}

// splitChild splits a full child node
func (t *BTree) splitChild(parent *BNode, idx int) {
	child := parent.children[idx]
	mid := len(child.keys) / 2

	// Safety check
	if mid >= len(child.keys) {
		return
	}

	// Create new right sibling
	var right *BNode
	var promoteKey uint64

	if child.isLeaf {
		right = newLeaf()
		right.keys = append(right.keys, child.keys[mid:]...)
		right.offsets = append(right.offsets, child.offsets[mid:]...)
		right.next = child.next
		child.next = right
		child.keys = child.keys[:mid]
		child.offsets = child.offsets[:mid]
		promoteKey = right.keys[0]
	} else {
		right = newInternal()
		promoteKey = child.keys[mid]

		// Check bounds before slicing
		if mid+1 < len(child.keys) {
			right.keys = append(right.keys, child.keys[mid+1:]...)
		}
		if mid+1 < len(child.children) {
			right.children = append(right.children, child.children[mid+1:]...)
		}

		child.keys = child.keys[:mid]
		if mid+1 < len(child.children) {
			child.children = child.children[:mid+1]
		}
	}

	// Insert promoted key into parent
	parent.keys = append(parent.keys, 0)
	copy(parent.keys[idx+1:], parent.keys[idx:])
	parent.keys[idx] = promoteKey

	// Insert right sibling into parent
	parent.children = append(parent.children, nil)
	copy(parent.children[idx+2:], parent.children[idx+1:])
	parent.children[idx+1] = right
}

// Search finds the offset for a given ID
func (t *BTree) Search(id uint64) (int64, bool) {
	return t.searchNode(t.root, id)
}

// searchNode searches recursively
func (t *BTree) searchNode(node *BNode, id uint64) (int64, bool) {
	if node.isLeaf {
		pos := sort.Search(len(node.keys), func(i int) bool {
			return node.keys[i] >= id
		})

		if pos < len(node.keys) && node.keys[pos] == id {
			return node.offsets[pos], true
		}
		return 0, false
	}

	// Internal node
	pos := sort.Search(len(node.keys), func(i int) bool {
		return node.keys[i] > id
	})

	return t.searchNode(node.children[pos], id)
}

// Delete removes an ID from the tree
func (t *BTree) Delete(id uint64) error {
	return t.deleteFromNode(t.root, id)
}

// deleteFromNode removes a key from a node
func (t *BTree) deleteFromNode(node *BNode, id uint64) error {
	if node.isLeaf {
		pos := sort.Search(len(node.keys), func(i int) bool {
			return node.keys[i] >= id
		})

		if pos >= len(node.keys) || node.keys[pos] != id {
			return fmt.Errorf("ID not found: %d", id)
		}

		// Remove key and offset
		node.keys = append(node.keys[:pos], node.keys[pos+1:]...)
		node.offsets = append(node.offsets[:pos], node.offsets[pos+1:]...)

		return nil
	}

	// Internal node - find child
	pos := sort.Search(len(node.keys), func(i int) bool {
		return node.keys[i] > id
	})

	return t.deleteFromNode(node.children[pos], id)
}

// GetAll returns all entries in sorted order
func (t *BTree) GetAll() map[uint64]int64 {
	result := make(map[uint64]int64)

	// Find leftmost leaf
	node := t.root
	for !node.isLeaf {
		node = node.children[0]
	}

	// Traverse leaves
	for node != nil {
		for i := range node.keys {
			result[node.keys[i]] = node.offsets[i]
		}
		node = node.next
	}

	return result
}

// Size returns the number of entries in the tree
func (t *BTree) Size() int {
	count := 0

	// Find leftmost leaf
	node := t.root
	for !node.isLeaf {
		node = node.children[0]
	}

	// Count entries in leaves
	for node != nil {
		count += len(node.keys)
		node = node.next
	}

	return count
}
