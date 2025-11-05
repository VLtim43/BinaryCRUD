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
	content := "48656c6c6f" // "Hello" in hex
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

	if string(data) != content {
		t.Errorf("expected %s, got %s", content, string(data))
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

	expected := "000000011f000000021f00000003"
	if string(data) != expected {
		t.Errorf("expected %s, got %s", expected, string(data))
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
	err = utils.WriteToFile(file, "existing content")
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
