package test

import (
	"BinaryCRUD/backend/index"
	"BinaryCRUD/backend/utils"
	"fmt"
	"os"
	"testing"
)

func TestInitializeDAOIndexCreatesCorrectPath(t *testing.T) {
	filePath := "/tmp/test_dao_helper_path.bin"
	expectedIndexPath := "/tmp/test_dao_helper_path.idx"

	indexPath, _ := utils.InitializeDAOIndex(filePath)

	if indexPath != expectedIndexPath {
		t.Errorf("Expected index path %s, got %s", expectedIndexPath, indexPath)
	}
}

func TestInitializeDAOIndexCreatesNewTreeWhenNoIndex(t *testing.T) {
	testFile := fmt.Sprintf("/tmp/test_dao_helper_new_%d.bin", os.Getpid())
	indexPath := "/tmp/test_dao_helper_new_" + fmt.Sprintf("%d", os.Getpid()) + ".idx"
	defer os.Remove(testFile)
	defer os.Remove(indexPath)

	// Ensure no index exists
	os.Remove(indexPath)

	_, tree := utils.InitializeDAOIndex(testFile)

	if tree == nil {
		t.Fatal("Expected tree to be created, got nil")
	}

	// Verify it's an empty tree
	_, found := tree.Search(0)
	if found {
		t.Error("Expected empty tree, but found data")
	}
}

func TestInitializeDAOIndexLoadsExistingIndex(t *testing.T) {
	testFile := fmt.Sprintf("/tmp/test_dao_helper_load_%d.bin", os.Getpid())
	indexPath := "/tmp/test_dao_helper_load_" + fmt.Sprintf("%d", os.Getpid()) + ".idx"
	defer os.Remove(testFile)
	defer os.Remove(indexPath)

	// Create and save an index with some data
	tree := index.NewBTree(utils.DefaultBTreeOrder)
	tree.Insert(1, 100)
	tree.Insert(2, 200)
	err := tree.Save(indexPath)
	if err != nil {
		t.Fatalf("Failed to save test index: %v", err)
	}

	// Now initialize - should load the existing index
	_, loadedTree := utils.InitializeDAOIndex(testFile)

	// Verify the data was loaded
	_, found1 := loadedTree.Search(1)
	if !found1 {
		t.Error("Expected to find key 1 in loaded tree")
	}
	_, found2 := loadedTree.Search(2)
	if !found2 {
		t.Error("Expected to find key 2 in loaded tree")
	}

	offset1, found := loadedTree.Search(1)
	if found && offset1 != 100 {
		t.Errorf("Expected offset 100 for key 1, got %d", offset1)
	}
}

func TestInitializeDAOIndexHandlesNonBinExtension(t *testing.T) {
	filePath := "/tmp/test_dao_helper_noext"
	expectedIndexPath := "/tmp/test_dao_helper_noext.idx"

	indexPath, _ := utils.InitializeDAOIndex(filePath)

	if indexPath != expectedIndexPath {
		t.Errorf("Expected index path %s for non-.bin file, got %s", expectedIndexPath, indexPath)
	}
}

func TestInitializeDAOIndexTreeHasCorrectOrder(t *testing.T) {
	testFile := fmt.Sprintf("/tmp/test_dao_helper_order_%d.bin", os.Getpid())
	indexPath := "/tmp/test_dao_helper_order_" + fmt.Sprintf("%d", os.Getpid()) + ".idx"
	defer os.Remove(testFile)
	defer os.Remove(indexPath)

	// Ensure no index exists
	os.Remove(indexPath)

	_, tree := utils.InitializeDAOIndex(testFile)

	// Insert enough items to verify tree structure works
	// With order 4, we should be able to insert multiple items without issues
	for i := uint64(0); i < 10; i++ {
		tree.Insert(i, int64(i*100))
	}

	// Verify all items can be retrieved
	for i := uint64(0); i < 10; i++ {
		offset, found := tree.Search(i)
		if !found {
			t.Errorf("Expected to find key %d in tree", i)
		} else if offset != int64(i*100) {
			t.Errorf("Expected offset %d for key %d, got %d", i*100, i, offset)
		}
	}
}

func TestInitializeDAOIndexMultipleCalls(t *testing.T) {
	testFile := fmt.Sprintf("/tmp/test_dao_helper_multiple_%d.bin", os.Getpid())
	indexPath := "/tmp/test_dao_helper_multiple_" + fmt.Sprintf("%d", os.Getpid()) + ".idx"
	defer os.Remove(testFile)
	defer os.Remove(indexPath)

	// First call - creates new tree
	_, tree1 := utils.InitializeDAOIndex(testFile)
	tree1.Insert(1, 100)
	err := tree1.Save(indexPath)
	if err != nil {
		t.Fatalf("Failed to save tree: %v", err)
	}

	// Second call - should load existing tree
	_, tree2 := utils.InitializeDAOIndex(testFile)

	// Verify data from first call is present
	offset, found := tree2.Search(1)
	if !found {
		t.Error("Expected to find key 1 from previous initialization")
	} else if offset != 100 {
		t.Errorf("Expected offset 100, got %d", offset)
	}
}

func TestInitializeDAOIndexCorruptedIndex(t *testing.T) {
	testFile := fmt.Sprintf("/tmp/test_dao_helper_corrupt_%d.bin", os.Getpid())
	indexPath := "/tmp/test_dao_helper_corrupt_" + fmt.Sprintf("%d", os.Getpid()) + ".idx"
	defer os.Remove(testFile)
	defer os.Remove(indexPath)

	// Create a corrupted index file (invalid data)
	err := os.WriteFile(indexPath, []byte("corrupted data"), 0644)
	if err != nil {
		t.Fatalf("Failed to create corrupted index: %v", err)
	}

	// Should create new tree when loading fails
	_, tree := utils.InitializeDAOIndex(testFile)

	if tree == nil {
		t.Fatal("Expected tree to be created despite corrupted index")
	}

	// Should be an empty tree
	_, found := tree.Search(0)
	if found {
		t.Error("Expected empty tree after failed load, but found data")
	}
}

func TestInitializeDAOIndexPreservesIndexPath(t *testing.T) {
	testCases := []struct {
		filePath  string
		expected  string
	}{
		{"/tmp/test.bin", "/tmp/test.idx"},
		{"/tmp/test.bin.bin", "/tmp/test.bin.idx"},
		{"/tmp/test", "/tmp/test.idx"},
		{"/home/user/data/items.bin", "/home/user/data/items.idx"},
	}

	for _, tc := range testCases {
		indexPath, _ := utils.InitializeDAOIndex(tc.filePath)
		if indexPath != tc.expected {
			t.Errorf("For filePath %s, expected %s, got %s", tc.filePath, tc.expected, indexPath)
		}
	}
}
