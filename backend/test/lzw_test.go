package test

import (
	"BinaryCRUD/backend/compression"
	"bytes"
	"fmt"
	"os"
	"testing"
)

func TestLZWCompressDecompress(t *testing.T) {
	lzw := compression.NewLZWCompressor()

	original := []byte("Hello, World! This is a test of LZW compression.")

	compressed, err := lzw.Compress(original)
	if err != nil {
		t.Fatalf("Compression failed: %v", err)
	}

	decompressed, err := lzw.Decompress(compressed)
	if err != nil {
		t.Fatalf("Decompression failed: %v", err)
	}

	if !bytes.Equal(original, decompressed) {
		t.Errorf("Decompressed data doesn't match original.\nOriginal: %v\nDecompressed: %v", original, decompressed)
	}
}

func TestLZWCompressReducesSize(t *testing.T) {
	lzw := compression.NewLZWCompressor()

	// Repetitive data should compress well
	original := bytes.Repeat([]byte("AAABBBCCC"), 100)

	compressed, err := lzw.Compress(original)
	if err != nil {
		t.Fatalf("Compression failed: %v", err)
	}

	if len(compressed) >= len(original) {
		t.Errorf("Compressed size (%d) should be smaller than original (%d)", len(compressed), len(original))
	}

	t.Logf("Original: %d bytes, Compressed: %d bytes (%.2f%% reduction)",
		len(original), len(compressed), float64(len(original)-len(compressed))/float64(len(original))*100)
}

func TestLZWSingleByte(t *testing.T) {
	lzw := compression.NewLZWCompressor()

	original := []byte{0x42}

	compressed, err := lzw.Compress(original)
	if err != nil {
		t.Fatalf("Compression failed: %v", err)
	}

	decompressed, err := lzw.Decompress(compressed)
	if err != nil {
		t.Fatalf("Decompression failed: %v", err)
	}

	if !bytes.Equal(original, decompressed) {
		t.Errorf("Single byte decompression failed")
	}
}

func TestLZWRepeatedSingleByte(t *testing.T) {
	lzw := compression.NewLZWCompressor()

	// All same bytes - LZW handles this well
	original := bytes.Repeat([]byte{0x00}, 1000)

	compressed, err := lzw.Compress(original)
	if err != nil {
		t.Fatalf("Compression failed: %v", err)
	}

	decompressed, err := lzw.Decompress(compressed)
	if err != nil {
		t.Fatalf("Decompression failed: %v", err)
	}

	if !bytes.Equal(original, decompressed) {
		t.Errorf("Repeated single byte decompression failed")
	}

	t.Logf("Original: %d bytes, Compressed: %d bytes", len(original), len(compressed))
}

func TestLZWBinaryData(t *testing.T) {
	lzw := compression.NewLZWCompressor()

	// Simulate binary file structure like items.bin
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

	compressed, err := lzw.Compress(data)
	if err != nil {
		t.Fatalf("Compression failed: %v", err)
	}

	decompressed, err := lzw.Decompress(compressed)
	if err != nil {
		t.Fatalf("Decompression failed: %v", err)
	}

	if !bytes.Equal(data, decompressed) {
		t.Errorf("Binary data decompression failed")
	}

	t.Logf("Binary data: Original %d bytes, Compressed %d bytes", len(data), len(compressed))
}

func TestLZWAllByteValues(t *testing.T) {
	lzw := compression.NewLZWCompressor()

	// Test with all possible byte values
	original := make([]byte, 256)
	for i := 0; i < 256; i++ {
		original[i] = byte(i)
	}

	compressed, err := lzw.Compress(original)
	if err != nil {
		t.Fatalf("Compression failed: %v", err)
	}

	decompressed, err := lzw.Decompress(compressed)
	if err != nil {
		t.Fatalf("Decompression failed: %v", err)
	}

	if !bytes.Equal(original, decompressed) {
		t.Errorf("All byte values decompression failed")
	}
}

func TestLZWEmptyData(t *testing.T) {
	lzw := compression.NewLZWCompressor()

	_, err := lzw.Compress([]byte{})
	if err == nil {
		t.Error("Expected error for empty data")
	}
}

func TestLZWInvalidMagic(t *testing.T) {
	lzw := compression.NewLZWCompressor()

	invalidData := []byte("XXXX1234567890")

	_, err := lzw.Decompress(invalidData)
	if err == nil {
		t.Error("Expected error for invalid magic bytes")
	}
}

func TestLZWFile(t *testing.T) {
	lzw := compression.NewLZWCompressor()

	// Create temp files
	inputPath := fmt.Sprintf("/tmp/lzw_test_input_%d.bin", os.Getpid())
	compressedPath := fmt.Sprintf("/tmp/lzw_test_compressed_%d.lzw", os.Getpid())
	outputPath := fmt.Sprintf("/tmp/lzw_test_output_%d.bin", os.Getpid())

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
	err = lzw.CompressFile(inputPath, compressedPath)
	if err != nil {
		t.Fatalf("CompressFile failed: %v", err)
	}

	// Decompress file
	err = lzw.DecompressFile(compressedPath, outputPath)
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

func TestLZWLargeFile(t *testing.T) {
	lzw := compression.NewLZWCompressor()

	// Generate larger test data (100KB)
	original := make([]byte, 100*1024)
	for i := range original {
		// Mix of values with skewed distribution
		switch i % 10 {
		case 0, 1, 2:
			original[i] = 0x00 // Common
		case 3, 4:
			original[i] = byte(i % 26) + 'a' // Letters
		default:
			original[i] = byte(i % 256) // Various
		}
	}

	compressed, err := lzw.Compress(original)
	if err != nil {
		t.Fatalf("Compression failed: %v", err)
	}

	decompressed, err := lzw.Decompress(compressed)
	if err != nil {
		t.Fatalf("Decompression failed: %v", err)
	}

	if !bytes.Equal(original, decompressed) {
		t.Error("Large file decompression failed")
	}

	ratio := float64(len(compressed)) / float64(len(original)) * 100
	t.Logf("Large file: Original %d bytes, Compressed %d bytes (%.2f%%)", len(original), len(compressed), ratio)
}

func TestLZWRepeatingPatterns(t *testing.T) {
	lzw := compression.NewLZWCompressor()

	// LZW excels at repeating patterns
	original := bytes.Repeat([]byte("TOBEORNOTTOBEORTOBEORNOT"), 100)

	compressed, err := lzw.Compress(original)
	if err != nil {
		t.Fatalf("Compression failed: %v", err)
	}

	decompressed, err := lzw.Decompress(compressed)
	if err != nil {
		t.Fatalf("Decompression failed: %v", err)
	}

	if !bytes.Equal(original, decompressed) {
		t.Errorf("Repeating patterns decompression failed")
	}

	ratio := float64(len(compressed)) / float64(len(original)) * 100
	t.Logf("Repeating patterns: Original %d bytes, Compressed %d bytes (%.2f%%)", len(original), len(compressed), ratio)
}

func TestCompareHuffmanVsLZW(t *testing.T) {
	hc := compression.NewHuffmanCompressor()
	lzw := compression.NewLZWCompressor()

	testCases := []struct {
		name string
		data []byte
	}{
		{"Repetitive text", bytes.Repeat([]byte("Hello World! "), 100)},
		{"Single byte repeat", bytes.Repeat([]byte{0x00}, 1000)},
		{"Mixed binary", func() []byte {
			d := make([]byte, 1000)
			for i := range d {
				d[i] = byte(i % 256)
			}
			return d
		}()},
	}

	for _, tc := range testCases {
		huffmanCompressed, _ := hc.Compress(tc.data)
		lzwCompressed, _ := lzw.Compress(tc.data)

		t.Logf("%s: Original %d, Huffman %d (%.1f%%), LZW %d (%.1f%%)",
			tc.name,
			len(tc.data),
			len(huffmanCompressed), float64(len(huffmanCompressed))/float64(len(tc.data))*100,
			len(lzwCompressed), float64(len(lzwCompressed))/float64(len(tc.data))*100)
	}
}
