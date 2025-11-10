package test

import (
	"os"
	"testing"

	"BinaryCRUD/backend/utils"
)

func TestCreateFile(t *testing.T) {
	testFile := "/tmp/test_create_file.bin"
	defer os.Remove(testFile)

	// Test creating a new file
	file, err := utils.CreateFile(testFile)
	if err != nil {
		t.Fatalf("failed to create file: %v", err)
	}
	defer file.Close()

	// Verify file exists
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Error("file was not created")
	}
}

func TestCreateFileRejectsExisting(t *testing.T) {
	testFile := "/tmp/test_create_existing.bin"
	defer os.Remove(testFile)

	// Create file first time
	file1, err := utils.CreateFile(testFile)
	if err != nil {
		t.Fatalf("failed to create file first time: %v", err)
	}
	file1.Close()

	// Try to create again - should fail
	file2, err := utils.CreateFile(testFile)
	if err == nil {
		file2.Close()
		t.Error("expected error when creating existing file, got none")
	}
}

func TestWriteToFile(t *testing.T) {
	testFile := "/tmp/test_write.bin"
	defer os.Remove(testFile)

	file, err := utils.CreateFile(testFile)
	if err != nil {
		t.Fatalf("failed to create file: %v", err)
	}
	defer file.Close()

	// Write content
	content := []byte("Hello")
	err = utils.WriteToFile(file, content)
	if err != nil {
		t.Errorf("failed to write to file: %v", err)
	}

	// Read back and verify
	file.Close()
	data, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	if string(data) != string(content) {
		t.Errorf("expected %s, got %s", string(content), string(data))
	}
}

func TestWriteHeaderToFile(t *testing.T) {
	testFile := "/tmp/test_write_header.bin"
	defer os.Remove(testFile)

	file, err := utils.CreateFile(testFile)
	if err != nil {
		t.Fatalf("failed to create file: %v", err)
	}
	defer file.Close()

	// Create header
	header, err := utils.WriteHeader(1, 2, 3)
	if err != nil {
		t.Fatalf("failed to create header: %v", err)
	}

	// Write header to file
	err = utils.WriteHeaderToFile(file, header)
	if err != nil {
		t.Errorf("failed to write header to file: %v", err)
	}

	// Read back and verify
	file.Close()
	data, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	expected := []byte{0x00, 0x00, 0x00, 0x01, 0x1f, 0x00, 0x00, 0x00, 0x02, 0x1f, 0x00, 0x00, 0x00, 0x03, 0x1e}
	if string(data) != string(expected) {
		t.Errorf("expected %v, got %v", expected, data)
	}
}

func TestWriteHeaderToFileRejectsNonEmpty(t *testing.T) {
	testFile := "/tmp/test_write_header_nonempty.bin"
	defer os.Remove(testFile)

	file, err := utils.CreateFile(testFile)
	if err != nil {
		t.Fatalf("failed to create file: %v", err)
	}
	defer file.Close()

	// Write some content first
	err = utils.WriteToFile(file, []byte("existing content"))
	if err != nil {
		t.Fatalf("failed to write initial content: %v", err)
	}

	// Try to write header - should fail
	header, err := utils.WriteHeader(1, 2, 3)
	if err != nil {
		t.Fatalf("failed to create header: %v", err)
	}

	err = utils.WriteHeaderToFile(file, header)
	if err == nil {
		t.Error("expected error when writing header to non-empty file, got none")
	}
}

func TestReadHeader(t *testing.T) {
	testFile := "/tmp/test_read_header.bin"
	defer os.Remove(testFile)

	// Create file with header
	file, err := utils.CreateFile(testFile)
	if err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	header, err := utils.WriteHeader(10, 5, 20)
	if err != nil {
		t.Fatalf("failed to create header: %v", err)
	}

	err = utils.WriteHeaderToFile(file, header)
	if err != nil {
		t.Fatalf("failed to write header: %v", err)
	}

	// Read header back
	entitiesCount, tombstoneCount, nextId, err := utils.ReadHeader(file)
	if err != nil {
		t.Errorf("failed to read header: %v", err)
	}

	if entitiesCount != 10 {
		t.Errorf("expected entitiesCount 10, got %d", entitiesCount)
	}
	if tombstoneCount != 5 {
		t.Errorf("expected tombstoneCount 5, got %d", tombstoneCount)
	}
	if nextId != 20 {
		t.Errorf("expected nextId 20, got %d", nextId)
	}

	file.Close()
}

func TestUpdateHeader(t *testing.T) {
	testFile := "/tmp/test_update_header.bin"
	defer os.Remove(testFile)

	// Create file with initial header
	file, err := utils.CreateFile(testFile)
	if err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	header, err := utils.WriteHeader(1, 2, 3)
	if err != nil {
		t.Fatalf("failed to create header: %v", err)
	}

	err = utils.WriteHeaderToFile(file, header)
	if err != nil {
		t.Fatalf("failed to write header: %v", err)
	}

	// Update header
	err = utils.UpdateHeader(file, 10, 20, 30)
	if err != nil {
		t.Errorf("failed to update header: %v", err)
	}

	// Read back and verify
	entitiesCount, tombstoneCount, nextId, err := utils.ReadHeader(file)
	if err != nil {
		t.Errorf("failed to read header: %v", err)
	}

	if entitiesCount != 10 {
		t.Errorf("expected entitiesCount 10, got %d", entitiesCount)
	}
	if tombstoneCount != 20 {
		t.Errorf("expected tombstoneCount 20, got %d", tombstoneCount)
	}
	if nextId != 30 {
		t.Errorf("expected nextId 30, got %d", nextId)
	}

	file.Close()
}

func TestAppendEntry(t *testing.T) {
	testFile := "/tmp/test_append_entry.bin"
	defer os.Remove(testFile)

	// Create file with empty header
	file, err := utils.CreateFile(testFile)
	if err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	header, err := utils.WriteHeader(0, 0, 1) // Start with nextId=1
	if err != nil {
		t.Fatalf("failed to create header: %v", err)
	}

	err = utils.WriteHeaderToFile(file, header)
	if err != nil {
		t.Fatalf("failed to write header: %v", err)
	}

	// Append first entry (without ID)
	entryData := []byte("hello")
	err = utils.AppendEntry(file, entryData)
	if err != nil {
		t.Errorf("failed to append entry: %v", err)
	}

	// Verify header was updated
	entitiesCount, tombstoneCount, nextId, err := utils.ReadHeader(file)
	if err != nil {
		t.Errorf("failed to read header: %v", err)
	}

	if entitiesCount != 1 {
		t.Errorf("expected entitiesCount 1, got %d", entitiesCount)
	}
	if tombstoneCount != 0 {
		t.Errorf("expected tombstoneCount 0, got %d", tombstoneCount)
	}
	if nextId != 2 {
		t.Errorf("expected nextId 2, got %d", nextId)
	}

	// Append second entry
	err = utils.AppendEntry(file, entryData)
	if err != nil {
		t.Errorf("failed to append second entry: %v", err)
	}

	// Verify header again
	entitiesCount, tombstoneCount, nextId, err = utils.ReadHeader(file)
	if err != nil {
		t.Errorf("failed to read header: %v", err)
	}

	if entitiesCount != 2 {
		t.Errorf("expected entitiesCount 2, got %d", entitiesCount)
	}
	if nextId != 3 {
		t.Errorf("expected nextId 3, got %d", nextId)
	}

	file.Close()

	// Read entire file and verify structure
	data, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	fileContent := string(data)

	// Should contain: header + entry1(with ID 0001) + separator + entry2(with ID 0002) + separator
	// ID is 2 bytes, so 0001 = "0001" in hex
	if len(fileContent) == 0 {
		t.Error("file is empty")
	}
}
