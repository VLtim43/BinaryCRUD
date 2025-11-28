package compression

import (
	"BinaryCRUD/backend/utils"
	"fmt"
)

// Compressor defines the interface for compression algorithms
type Compressor interface {
	Compress(data []byte) ([]byte, error)
	Decompress(data []byte) ([]byte, error)
	CompressFile(inputPath, outputPath string) error
	DecompressFile(inputPath, outputPath string) error
}

// NewCompressor returns a Compressor implementation based on the algorithm name
func NewCompressor(algorithm string) (Compressor, error) {
	switch algorithm {
	case utils.AlgorithmHuffman:
		return NewHuffmanCompressor(), nil
	case utils.AlgorithmLZW:
		return NewLZWCompressor(), nil
	default:
		return nil, fmt.Errorf("unknown algorithm: %s", algorithm)
	}
}
