package test

import (
	"BinaryCRUD/backend/compression"
	"bytes"
	"fmt"
	"os"
	"testing"
)

func TestHuffmanCompressDecompress(t *testing.T) {
	hc := compression.NewHuffmanCompressor()

	original := []byte("Hello, World! This is a test of Huffman compression.")

	compressed, err := hc.Compress(original)
	if err != nil {
		t.Fatalf("Compression failed: %v", err)
	}

	decompressed, err := hc.Decompress(compressed)
	if err != nil {
		t.Fatalf("Decompression failed: %v", err)
	}

	if !bytes.Equal(original, decompressed) {
		t.Errorf("Decompressed data doesn't match original.\nOriginal: %v\nDecompressed: %v", original, decompressed)
	}
}

func TestHuffmanCompressReducesSize(t *testing.T) {
	hc := compression.NewHuffmanCompressor()

	// Repetitive data should compress well
	original := bytes.Repeat([]byte("AAABBBCCC"), 100)

	compressed, err := hc.Compress(original)
	if err != nil {
		t.Fatalf("Compression failed: %v", err)
	}

	if len(compressed) >= len(original) {
		t.Errorf("Compressed size (%d) should be smaller than original (%d)", len(compressed), len(original))
	}

	t.Logf("Original: %d bytes, Compressed: %d bytes (%.2f%% reduction)",
		len(original), len(compressed), float64(len(original)-len(compressed))/float64(len(original))*100)
}

func TestHuffmanSingleByte(t *testing.T) {
	hc := compression.NewHuffmanCompressor()

	original := []byte{0x42}

	compressed, err := hc.Compress(original)
	if err != nil {
		t.Fatalf("Compression failed: %v", err)
	}

	decompressed, err := hc.Decompress(compressed)
	if err != nil {
		t.Fatalf("Decompression failed: %v", err)
	}

	if !bytes.Equal(original, decompressed) {
		t.Errorf("Single byte decompression failed")
	}
}

func TestHuffmanRepeatedSingleByte(t *testing.T) {
	hc := compression.NewHuffmanCompressor()

	// All same bytes - extreme case
	original := bytes.Repeat([]byte{0x00}, 1000)

	compressed, err := hc.Compress(original)
	if err != nil {
		t.Fatalf("Compression failed: %v", err)
	}

	decompressed, err := hc.Decompress(compressed)
	if err != nil {
		t.Fatalf("Decompression failed: %v", err)
	}

	if !bytes.Equal(original, decompressed) {
		t.Errorf("Repeated single byte decompression failed")
	}

	t.Logf("Original: %d bytes, Compressed: %d bytes", len(original), len(compressed))
}

func TestHuffmanBinaryData(t *testing.T) {
	hc := compression.NewHuffmanCompressor()

	// Simulate binary file structure like items.bin
	// [header][entry1][entry2]...
	var original bytes.Buffer

	// Header: entry count = 3
	original.Write([]byte{0x00, 0x03})

	// Entry 1: [recordLen][id][name "Burger"][price][tombstone]
	original.Write([]byte{0x00, 0x0B}) // record length = 11
	original.Write([]byte{0x00, 0x00}) // id = 0
	original.Write([]byte("Burger"))   // name
	original.Write([]byte{0x03, 0x83}) // price = 899
	original.Write([]byte{0x00})       // tombstone

	// Entry 2: [recordLen][id][name "Fries"][price][tombstone]
	original.Write([]byte{0x00, 0x0A}) // record length = 10
	original.Write([]byte{0x00, 0x01}) // id = 1
	original.Write([]byte("Fries"))    // name
	original.Write([]byte{0x01, 0x5D}) // price = 349
	original.Write([]byte{0x00})       // tombstone

	// Entry 3: [recordLen][id][name "Cola"][price][tombstone]
	original.Write([]byte{0x00, 0x09}) // record length = 9
	original.Write([]byte{0x00, 0x02}) // id = 2
	original.Write([]byte("Cola"))     // name
	original.Write([]byte{0x00, 0xC7}) // price = 199
	original.Write([]byte{0x00})       // tombstone

	data := original.Bytes()

	compressed, err := hc.Compress(data)
	if err != nil {
		t.Fatalf("Compression failed: %v", err)
	}

	decompressed, err := hc.Decompress(compressed)
	if err != nil {
		t.Fatalf("Decompression failed: %v", err)
	}

	if !bytes.Equal(data, decompressed) {
		t.Errorf("Binary data decompression failed")
	}

	t.Logf("Binary data: Original %d bytes, Compressed %d bytes", len(data), len(compressed))
}

func TestHuffmanAllByteValues(t *testing.T) {
	hc := compression.NewHuffmanCompressor()

	// Test with all possible byte values
	original := make([]byte, 256)
	for i := 0; i < 256; i++ {
		original[i] = byte(i)
	}

	compressed, err := hc.Compress(original)
	if err != nil {
		t.Fatalf("Compression failed: %v", err)
	}

	decompressed, err := hc.Decompress(compressed)
	if err != nil {
		t.Fatalf("Decompression failed: %v", err)
	}

	if !bytes.Equal(original, decompressed) {
		t.Errorf("All byte values decompression failed")
	}
}

func TestHuffmanEmptyData(t *testing.T) {
	hc := compression.NewHuffmanCompressor()

	_, err := hc.Compress([]byte{})
	if err == nil {
		t.Error("Expected error for empty data")
	}
}

func TestHuffmanInvalidMagic(t *testing.T) {
	hc := compression.NewHuffmanCompressor()

	invalidData := []byte("XXXX1234567890")

	_, err := hc.Decompress(invalidData)
	if err == nil {
		t.Error("Expected error for invalid magic bytes")
	}
}

func TestHuffmanFile(t *testing.T) {
	hc := compression.NewHuffmanCompressor()

	// Create temp files
	inputPath := fmt.Sprintf("/tmp/huffman_test_input_%d.bin", os.Getpid())
	compressedPath := fmt.Sprintf("/tmp/huffman_test_compressed_%d.huff", os.Getpid())
	outputPath := fmt.Sprintf("/tmp/huffman_test_output_%d.bin", os.Getpid())

	defer os.Remove(inputPath)
	defer os.Remove(compressedPath)
	defer os.Remove(outputPath)

	// Write test data
	original := bytes.Repeat([]byte("Test data for file compression. "), 50)
	err := os.WriteFile(inputPath, original, 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Compress file
	err = hc.CompressFile(inputPath, compressedPath)
	if err != nil {
		t.Fatalf("CompressFile failed: %v", err)
	}

	// Decompress file
	err = hc.DecompressFile(compressedPath, outputPath)
	if err != nil {
		t.Fatalf("DecompressFile failed: %v", err)
	}

	// Verify
	decompressed, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	if !bytes.Equal(original, decompressed) {
		t.Error("File compression/decompression round-trip failed")
	}

	// Check sizes
	compressedInfo, _ := os.Stat(compressedPath)
	t.Logf("File: Original %d bytes, Compressed %d bytes", len(original), compressedInfo.Size())
}

func TestHuffmanLargeFile(t *testing.T) {
	hc := compression.NewHuffmanCompressor()

	// Generate larger test data (100KB)
	original := make([]byte, 100*1024)
	for i := range original {
		// Mix of values with skewed distribution (like real binary files)
		switch i % 10 {
		case 0, 1, 2:
			original[i] = 0x00 // Common
		case 3, 4:
			original[i] = byte(i % 26) + 'a' // Letters
		default:
			original[i] = byte(i % 256) // Various
		}
	}

	compressed, err := hc.Compress(original)
	if err != nil {
		t.Fatalf("Compression failed: %v", err)
	}

	decompressed, err := hc.Decompress(compressed)
	if err != nil {
		t.Fatalf("Decompression failed: %v", err)
	}

	if !bytes.Equal(original, decompressed) {
		t.Error("Large file decompression failed")
	}

	ratio := float64(len(compressed)) / float64(len(original)) * 100
	t.Logf("Large file: Original %d bytes, Compressed %d bytes (%.2f%%)", len(original), len(compressed), ratio)
}

func TestHuffmanStats(t *testing.T) {
	hc := compression.NewHuffmanCompressor()

	stats := hc.GetCompressionStats(1000, 600)

	if stats["originalSize"] != 1000 {
		t.Errorf("Expected originalSize 1000, got %v", stats["originalSize"])
	}

	if stats["compressedSize"] != 600 {
		t.Errorf("Expected compressedSize 600, got %v", stats["compressedSize"])
	}

	// 600/1000 = 60%
	if stats["ratio"] != "60.00%" {
		t.Errorf("Expected ratio 60.00%%, got %v", stats["ratio"])
	}

	// Saved 40%
	if stats["spaceSaved"] != "40.00%" {
		t.Errorf("Expected spaceSaved 40.00%%, got %v", stats["spaceSaved"])
	}
}
