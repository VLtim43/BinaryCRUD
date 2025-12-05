package compression

import (
	"fmt"
)

// Compressor is the interface that all compression algorithms implement
type Compressor interface {
	Compress(data []byte) ([]byte, error)
	Decompress(data []byte) ([]byte, error)
	CompressFile(inputPath, outputPath string) error
	DecompressFile(inputPath, outputPath string) error
}

// Algorithm constants
const (
	AlgorithmHuffman = "huffman"
	AlgorithmLZW     = "lzw"
)

// NewCompressor creates a compressor for the given algorithm
func NewCompressor(algorithm string) (Compressor, error) {
	switch algorithm {
	case AlgorithmHuffman:
		return NewHuffmanCompressor(), nil
	case AlgorithmLZW:
		return NewLZWCompressor(), nil
	default:
		return nil, fmt.Errorf("unknown compression algorithm: %s", algorithm)
	}
}
