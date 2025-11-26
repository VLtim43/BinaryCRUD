package utils

import (
	"path/filepath"
	"strings"
)

// BDATMagic is the magic bytes for binary data files
var BDATMagic = []byte{'B', 'D', 'A', 'T'}

const (
	// IDSize is the size of the ID field in bytes
	IDSize = 2

	// TombstoneSize is the size of the tombstone field in bytes
	TombstoneSize = 1

	// RecordLengthSize is the size of the record length prefix in bytes
	RecordLengthSize = 2

	// HeaderFieldSize is the size of each header field in bytes
	HeaderFieldSize = 4

	// MagicSize is the size of the magic bytes
	MagicSize = 4

	// FilenameLengthSize is the size of the filename length field
	FilenameLengthSize = 1

	// HeaderFixedSize is the fixed portion of the header (magic + counts)
	// Format: [magic(4)][filenameLen(1)][filename(N)][entitiesCount(4)][tombstoneCount(4)][nextId(4)]
	// The variable part is filename, fixed part = 4 + 1 + 4 + 4 + 4 = 17 bytes + filename
	HeaderFixedSize = MagicSize + FilenameLengthSize + (HeaderFieldSize * 3)

	// DefaultBTreeOrder is the default order for B+ tree indices
	DefaultBTreeOrder = 4

	// Data directory paths
	DataDir       = "data"
	BinDir        = "data/bin"
	IndexDir      = "data/indexes"
	CompressedDir = "data/compressed"
	SeedDir       = "data/seed"

	// Compression algorithms
	AlgorithmHuffman = "huffman"
	AlgorithmLZW     = "lzw"
	AlgorithmUnknown = "unknown"
)

// CalculateHeaderSize returns the total header size for a given filename
func CalculateHeaderSize(filename string) int {
	return HeaderFixedSize + len(filename)
}

// BinPath returns the full path for a file in the bin directory
func BinPath(filename string) string {
	return filepath.Join(BinDir, filename)
}

// IndexPath returns the full path for a file in the indexes directory
func IndexPath(filename string) string {
	return filepath.Join(IndexDir, filename)
}

// CompressedPath returns the full path for a file in the compressed directory
func CompressedPath(filename string) string {
	return filepath.Join(CompressedDir, filename)
}

// SeedPath returns the full path for a file in the seed directory
func SeedPath(filename string) string {
	return filepath.Join(SeedDir, filename)
}

// DetectCompressionAlgorithm determines the compression algorithm from a filename
func DetectCompressionAlgorithm(filename string) string {
	if strings.Contains(filename, ".huffman.") {
		return AlgorithmHuffman
	}
	if strings.Contains(filename, ".lzw.") {
		return AlgorithmLZW
	}
	return AlgorithmUnknown
}

// CompressedFilename generates the compressed filename for a given algorithm
func CompressedFilename(originalName, algorithm string) string {
	return originalName + "." + algorithm + ".compressed"
}

// DecompressedFilename extracts the original filename from a compressed filename
func DecompressedFilename(compressedName string) string {
	name := strings.TrimSuffix(compressedName, ".huffman.compressed")
	name = strings.TrimSuffix(name, ".lzw.compressed")
	return name
}
