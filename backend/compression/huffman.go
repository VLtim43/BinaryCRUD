package compression

import (
	"bytes"
	"container/heap"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// Magic bytes to identify Huffman compressed files
var HuffmanMagic = []byte{'H', 'U', 'F', 'F'}

// HuffmanNode represents a node in the Huffman tree
type HuffmanNode struct {
	Byte   byte
	Freq   int
	Left   *HuffmanNode
	Right  *HuffmanNode
	IsLeaf bool
}

// Priority queue implementation for building Huffman tree
type nodeHeap []*HuffmanNode

func (h nodeHeap) Len() int           { return len(h) }
func (h nodeHeap) Less(i, j int) bool { return h[i].Freq < h[j].Freq }
func (h nodeHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *nodeHeap) Push(x any) {
	*h = append(*h, x.(*HuffmanNode))
}

func (h *nodeHeap) Pop() any {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

// HuffmanCompressor handles Huffman compression and decompression
type HuffmanCompressor struct {
	root     *HuffmanNode
	codeMap  map[byte]string // byte -> bit string
	freq     [256]int
}

// NewHuffmanCompressor creates a new Huffman compressor
func NewHuffmanCompressor() *HuffmanCompressor {
	return &HuffmanCompressor{
		codeMap: make(map[byte]string),
	}
}

// buildFrequencyTable counts the frequency of each byte in the data
func (hc *HuffmanCompressor) buildFrequencyTable(data []byte) {
	// Reset frequencies
	for i := range hc.freq {
		hc.freq[i] = 0
	}

	// Count each byte
	for _, b := range data {
		hc.freq[b]++
	}
}

// buildTree constructs the Huffman tree from frequency table
func (hc *HuffmanCompressor) buildTree() {
	// Create a min-heap of nodes
	h := &nodeHeap{}
	heap.Init(h)

	// Add all bytes with non-zero frequency
	for i := 0; i < 256; i++ {
		if hc.freq[i] > 0 {
			heap.Push(h, &HuffmanNode{
				Byte:   byte(i),
				Freq:   hc.freq[i],
				IsLeaf: true,
			})
		}
	}

	// Handle edge case: empty or single-byte data
	if h.Len() == 0 {
		hc.root = nil
		return
	}

	if h.Len() == 1 {
		// Single unique byte - create a simple tree
		node := heap.Pop(h).(*HuffmanNode)
		hc.root = &HuffmanNode{
			Freq:   node.Freq,
			Left:   node,
			IsLeaf: false,
		}
		return
	}

	// Build tree by combining two smallest nodes repeatedly
	for h.Len() > 1 {
		// Pop two smallest
		left := heap.Pop(h).(*HuffmanNode)
		right := heap.Pop(h).(*HuffmanNode)

		// Create parent node
		parent := &HuffmanNode{
			Freq:   left.Freq + right.Freq,
			Left:   left,
			Right:  right,
			IsLeaf: false,
		}

		heap.Push(h, parent)
	}

	hc.root = heap.Pop(h).(*HuffmanNode)
}

// buildCodeMap generates the bit codes for each byte
func (hc *HuffmanCompressor) buildCodeMap() {
	hc.codeMap = make(map[byte]string)
	if hc.root == nil {
		return
	}
	hc.buildCodesRecursive(hc.root, "")
}

func (hc *HuffmanCompressor) buildCodesRecursive(node *HuffmanNode, code string) {
	if node == nil {
		return
	}

	if node.IsLeaf {
		if code == "" {
			code = "0" // Single node case
		}
		hc.codeMap[node.Byte] = code
		return
	}

	hc.buildCodesRecursive(node.Left, code+"0")
	hc.buildCodesRecursive(node.Right, code+"1")
}

// Compress compresses the input data using Huffman coding
func (hc *HuffmanCompressor) Compress(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("cannot compress empty data")
	}

	// Step 1: Build frequency table
	hc.buildFrequencyTable(data)

	// Step 2: Build Huffman tree
	hc.buildTree()

	// Step 3: Build code map
	hc.buildCodeMap()

	// Step 4: Encode data to bit string
	var bitString bytes.Buffer
	for _, b := range data {
		code, ok := hc.codeMap[b]
		if !ok {
			return nil, fmt.Errorf("no code found for byte 0x%02x", b)
		}
		bitString.WriteString(code)
	}

	// Step 5: Convert bit string to bytes
	bits := bitString.String()
	paddingBits := (8 - (len(bits) % 8)) % 8

	// Pad with zeros
	for i := 0; i < paddingBits; i++ {
		bits += "0"
	}

	// Convert to bytes
	compressedData := make([]byte, len(bits)/8)
	for i := 0; i < len(bits); i += 8 {
		var b byte
		for j := 0; j < 8; j++ {
			if bits[i+j] == '1' {
				b |= 1 << (7 - j)
			}
		}
		compressedData[i/8] = b
	}

	// Step 6: Serialize the tree
	treeData := hc.serializeTree()

	// Step 7: Build final output
	// Format: [HUFF][originalSize(4)][treeSize(2)][tree][paddingBits(1)][compressedData]
	var output bytes.Buffer

	// Magic bytes
	output.Write(HuffmanMagic)

	// Original size (4 bytes)
	originalSize := make([]byte, 4)
	binary.LittleEndian.PutUint32(originalSize, uint32(len(data)))
	output.Write(originalSize)

	// Tree size (2 bytes)
	treeSize := make([]byte, 2)
	binary.LittleEndian.PutUint16(treeSize, uint16(len(treeData)))
	output.Write(treeSize)

	// Tree data
	output.Write(treeData)

	// Padding bits (1 byte)
	output.WriteByte(byte(paddingBits))

	// Compressed data
	output.Write(compressedData)

	return output.Bytes(), nil
}

// serializeTree serializes the Huffman tree for storage
// Format: For each node, if leaf: [1][byte], if internal: [0][left][right]
func (hc *HuffmanCompressor) serializeTree() []byte {
	var buf bytes.Buffer
	hc.serializeNode(&buf, hc.root)
	return buf.Bytes()
}

func (hc *HuffmanCompressor) serializeNode(buf *bytes.Buffer, node *HuffmanNode) {
	if node == nil {
		return
	}

	if node.IsLeaf {
		buf.WriteByte(1) // Leaf marker
		buf.WriteByte(node.Byte)
	} else {
		buf.WriteByte(0) // Internal node marker
		hc.serializeNode(buf, node.Left)
		hc.serializeNode(buf, node.Right)
	}
}

// Decompress decompresses Huffman-encoded data
func (hc *HuffmanCompressor) Decompress(data []byte) ([]byte, error) {
	if len(data) < 11 { // Minimum: magic(4) + size(4) + treeSize(2) + padding(1)
		return nil, fmt.Errorf("data too short to be valid Huffman compressed data")
	}

	reader := bytes.NewReader(data)

	// Verify magic bytes
	magic := make([]byte, 4)
	if _, err := reader.Read(magic); err != nil {
		return nil, fmt.Errorf("failed to read magic bytes: %w", err)
	}
	if !bytes.Equal(magic, HuffmanMagic) {
		return nil, fmt.Errorf("invalid magic bytes: expected HUFF, got %s", string(magic))
	}

	// Read original size
	originalSizeBytes := make([]byte, 4)
	if _, err := reader.Read(originalSizeBytes); err != nil {
		return nil, fmt.Errorf("failed to read original size: %w", err)
	}
	originalSize := binary.LittleEndian.Uint32(originalSizeBytes)

	// Read tree size
	treeSizeBytes := make([]byte, 2)
	if _, err := reader.Read(treeSizeBytes); err != nil {
		return nil, fmt.Errorf("failed to read tree size: %w", err)
	}
	treeSize := binary.LittleEndian.Uint16(treeSizeBytes)

	// Read tree data
	treeData := make([]byte, treeSize)
	if _, err := reader.Read(treeData); err != nil {
		return nil, fmt.Errorf("failed to read tree data: %w", err)
	}

	// Deserialize tree
	treeReader := bytes.NewReader(treeData)
	hc.root = hc.deserializeNode(treeReader)

	// Read padding bits
	paddingBits, err := reader.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("failed to read padding bits: %w", err)
	}

	// Read compressed data
	compressedData, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read compressed data: %w", err)
	}

	// Convert bytes to bit string
	var bitString bytes.Buffer
	for _, b := range compressedData {
		for i := 7; i >= 0; i-- {
			if b&(1<<i) != 0 {
				bitString.WriteByte('1')
			} else {
				bitString.WriteByte('0')
			}
		}
	}

	// Remove padding bits
	bits := bitString.String()
	if paddingBits > 0 && len(bits) >= int(paddingBits) {
		bits = bits[:len(bits)-int(paddingBits)]
	}

	// Decode using tree
	output := make([]byte, 0, originalSize)
	node := hc.root

	for _, bit := range bits {
		if node == nil {
			return nil, fmt.Errorf("invalid compressed data: null node during traversal")
		}

		if bit == '0' {
			node = node.Left
		} else {
			node = node.Right
		}

		if node == nil {
			return nil, fmt.Errorf("invalid compressed data: traversal led to null")
		}

		if node.IsLeaf {
			output = append(output, node.Byte)
			node = hc.root

			if uint32(len(output)) >= originalSize {
				break
			}
		}
	}

	if uint32(len(output)) != originalSize {
		return nil, fmt.Errorf("decompression size mismatch: expected %d, got %d", originalSize, len(output))
	}

	return output, nil
}

// deserializeNode reconstructs a node from serialized data
func (hc *HuffmanCompressor) deserializeNode(reader *bytes.Reader) *HuffmanNode {
	marker, err := reader.ReadByte()
	if err != nil {
		return nil
	}

	if marker == 1 {
		// Leaf node
		b, err := reader.ReadByte()
		if err != nil {
			return nil
		}
		return &HuffmanNode{
			Byte:   b,
			IsLeaf: true,
		}
	}

	// Internal node
	return &HuffmanNode{
		Left:   hc.deserializeNode(reader),
		Right:  hc.deserializeNode(reader),
		IsLeaf: false,
	}
}

// CompressFile compresses a file and saves it to the output path
func (hc *HuffmanCompressor) CompressFile(inputPath, outputPath string) error {
	// Read input file
	data, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("failed to read input file: %w", err)
	}

	// Compress
	compressed, err := hc.Compress(data)
	if err != nil {
		return fmt.Errorf("failed to compress: %w", err)
	}

	// Ensure output directory exists
	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Write output file
	err = os.WriteFile(outputPath, compressed, 0644)
	if err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}

	return nil
}

// DecompressFile decompresses a file and saves it to the output path
func (hc *HuffmanCompressor) DecompressFile(inputPath, outputPath string) error {
	// Read input file
	data, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("failed to read input file: %w", err)
	}

	// Decompress
	decompressed, err := hc.Decompress(data)
	if err != nil {
		return fmt.Errorf("failed to decompress: %w", err)
	}

	// Ensure output directory exists
	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Write output file
	err = os.WriteFile(outputPath, decompressed, 0644)
	if err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}

	return nil
}

// GetCompressionStats returns compression statistics
func (hc *HuffmanCompressor) GetCompressionStats(originalSize, compressedSize int) map[string]any {
	ratio := float64(compressedSize) / float64(originalSize) * 100
	saved := float64(originalSize-compressedSize) / float64(originalSize) * 100

	return map[string]any{
		"originalSize":   originalSize,
		"compressedSize": compressedSize,
		"ratio":          fmt.Sprintf("%.2f%%", ratio),
		"spaceSaved":     fmt.Sprintf("%.2f%%", saved),
	}
}
