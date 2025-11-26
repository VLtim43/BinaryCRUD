package compression

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// Magic bytes to identify LZW compressed files
var LZWMagic = []byte{'L', 'Z', 'W', 'W'}

// LZWCompressor handles LZW compression and decompression
type LZWCompressor struct{}

// NewLZWCompressor creates a new LZW compressor
func NewLZWCompressor() *LZWCompressor {
	return &LZWCompressor{}
}

// Compress compresses data using LZW algorithm
func (lzw *LZWCompressor) Compress(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("cannot compress empty data")
	}

	// Initialize dictionary with single bytes (0-255)
	dictionary := make(map[string]uint16)
	for i := 0; i < 256; i++ {
		dictionary[string([]byte{byte(i)})] = uint16(i)
	}

	var codes []uint16
	nextCode := uint16(256)
	current := ""

	for _, b := range data {
		combined := current + string([]byte{b})
		if _, exists := dictionary[combined]; exists {
			current = combined
		} else {
			// Output code for current
			codes = append(codes, dictionary[current])
			// Add new sequence to dictionary (if not full)
			if nextCode < 65535 {
				dictionary[combined] = nextCode
				nextCode++
			}
			current = string([]byte{b})
		}
	}

	// Output code for remaining
	if current != "" {
		codes = append(codes, dictionary[current])
	}

	// Build output
	// Format: [LZWW][originalSize(4)][codeCount(4)][codes as uint16...]
	var output bytes.Buffer

	// Magic bytes
	output.Write(LZWMagic)

	// Original size (4 bytes)
	originalSize := make([]byte, 4)
	binary.LittleEndian.PutUint32(originalSize, uint32(len(data)))
	output.Write(originalSize)

	// Code count (4 bytes)
	codeCount := make([]byte, 4)
	binary.LittleEndian.PutUint32(codeCount, uint32(len(codes)))
	output.Write(codeCount)

	// Codes (2 bytes each)
	for _, code := range codes {
		codeBytes := make([]byte, 2)
		binary.LittleEndian.PutUint16(codeBytes, code)
		output.Write(codeBytes)
	}

	return output.Bytes(), nil
}

// Decompress decompresses LZW-encoded data
func (lzw *LZWCompressor) Decompress(data []byte) ([]byte, error) {
	if len(data) < 12 { // Minimum: magic(4) + originalSize(4) + codeCount(4)
		return nil, fmt.Errorf("data too short to be valid LZW compressed data")
	}

	reader := bytes.NewReader(data)

	// Verify magic bytes
	magic := make([]byte, 4)
	if _, err := reader.Read(magic); err != nil {
		return nil, fmt.Errorf("failed to read magic bytes: %w", err)
	}
	if !bytes.Equal(magic, LZWMagic) {
		return nil, fmt.Errorf("invalid magic bytes: expected LZWW, got %s", string(magic))
	}

	// Read original size
	originalSizeBytes := make([]byte, 4)
	if _, err := reader.Read(originalSizeBytes); err != nil {
		return nil, fmt.Errorf("failed to read original size: %w", err)
	}
	originalSize := binary.LittleEndian.Uint32(originalSizeBytes)

	// Read code count
	codeCountBytes := make([]byte, 4)
	if _, err := reader.Read(codeCountBytes); err != nil {
		return nil, fmt.Errorf("failed to read code count: %w", err)
	}
	codeCount := binary.LittleEndian.Uint32(codeCountBytes)

	// Read codes
	codes := make([]uint16, codeCount)
	for i := uint32(0); i < codeCount; i++ {
		codeBytes := make([]byte, 2)
		if _, err := io.ReadFull(reader, codeBytes); err != nil {
			return nil, fmt.Errorf("failed to read code %d: %w", i, err)
		}
		codes[i] = binary.LittleEndian.Uint16(codeBytes)
	}

	// Initialize dictionary with single bytes (0-255)
	dictionary := make(map[uint16][]byte)
	for i := 0; i < 256; i++ {
		dictionary[uint16(i)] = []byte{byte(i)}
	}

	var output bytes.Buffer
	nextCode := uint16(256)

	if len(codes) == 0 {
		return output.Bytes(), nil
	}

	// First code - make a copy to avoid slice aliasing
	current := make([]byte, len(dictionary[codes[0]]))
	copy(current, dictionary[codes[0]])
	output.Write(current)

	for i := 1; i < len(codes); i++ {
		code := codes[i]
		var entry []byte

		if val, exists := dictionary[code]; exists {
			// Make a copy of the dictionary entry
			entry = make([]byte, len(val))
			copy(entry, val)
		} else if code == nextCode {
			// Special case: code not yet in dictionary
			// Create new slice with current + first byte of current
			entry = make([]byte, len(current)+1)
			copy(entry, current)
			entry[len(current)] = current[0]
		} else {
			return nil, fmt.Errorf("invalid code %d at position %d", code, i)
		}

		output.Write(entry)

		// Add new entry to dictionary
		if nextCode < 65535 {
			newEntry := make([]byte, len(current)+1)
			copy(newEntry, current)
			newEntry[len(current)] = entry[0]
			dictionary[nextCode] = newEntry
			nextCode++
		}

		current = entry
	}

	result := output.Bytes()
	if uint32(len(result)) != originalSize {
		return nil, fmt.Errorf("decompression size mismatch: expected %d, got %d", originalSize, len(result))
	}

	return result, nil
}

// CompressFile compresses a file and saves it to the output path
func (lzw *LZWCompressor) CompressFile(inputPath, outputPath string) error {
	data, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("failed to read input file: %w", err)
	}

	compressed, err := lzw.Compress(data)
	if err != nil {
		return fmt.Errorf("failed to compress: %w", err)
	}

	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	err = os.WriteFile(outputPath, compressed, 0644)
	if err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}

	return nil
}

// DecompressFile decompresses a file and saves it to the output path
func (lzw *LZWCompressor) DecompressFile(inputPath, outputPath string) error {
	data, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("failed to read input file: %w", err)
	}

	decompressed, err := lzw.Decompress(data)
	if err != nil {
		return fmt.Errorf("failed to decompress: %w", err)
	}

	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	err = os.WriteFile(outputPath, decompressed, 0644)
	if err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}

	return nil
}
