package index

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
)

// ExtensibleHash is a dynamic hash index that grows by doubling the directory
// Designed for composite keys (orderID, promotionID) -> file offset
type ExtensibleHash struct {
	globalDepth  int
	bucketSize   int
	directory    []*Bucket
}

// Bucket holds entries with the same hash prefix
type Bucket struct {
	localDepth int
	entries    []HashEntry
}

// HashEntry stores a composite key and its file offset
type HashEntry struct {
	OrderID     uint64
	PromotionID uint64
	Offset      int64
}

// NewExtensibleHash creates a new extensible hash index
func NewExtensibleHash(bucketSize int) *ExtensibleHash {
	if bucketSize < 2 {
		bucketSize = 4
	}

	// Start with global depth 1 (2 buckets)
	initialBucket := &Bucket{
		localDepth: 1,
		entries:    make([]HashEntry, 0, bucketSize),
	}

	return &ExtensibleHash{
		globalDepth: 1,
		bucketSize:  bucketSize,
		directory:   []*Bucket{initialBucket, initialBucket},
	}
}

// hash combines orderID and promotionID into a hash value
func (h *ExtensibleHash) hash(orderID, promotionID uint64) uint64 {
	// Simple hash combining both IDs
	// Using FNV-1a-like mixing
	hash := uint64(14695981039346656037) // FNV offset basis
	hash ^= orderID
	hash *= 1099511628211 // FNV prime
	hash ^= promotionID
	hash *= 1099511628211
	return hash
}

// getBucketIndex returns the directory index for a hash value
func (h *ExtensibleHash) getBucketIndex(hashValue uint64) int {
	// Use the last globalDepth bits
	mask := uint64((1 << h.globalDepth) - 1)
	return int(hashValue & mask)
}

// Insert adds a new entry to the hash index
func (h *ExtensibleHash) Insert(orderID, promotionID uint64, offset int64) error {
	hashValue := h.hash(orderID, promotionID)
	bucketIdx := h.getBucketIndex(hashValue)
	bucket := h.directory[bucketIdx]

	// Check for duplicate
	for _, entry := range bucket.entries {
		if entry.OrderID == orderID && entry.PromotionID == promotionID {
			return fmt.Errorf("duplicate key: orderID=%d, promotionID=%d", orderID, promotionID)
		}
	}

	// If bucket has space, insert directly
	if len(bucket.entries) < h.bucketSize {
		bucket.entries = append(bucket.entries, HashEntry{
			OrderID:     orderID,
			PromotionID: promotionID,
			Offset:      offset,
		})
		return nil
	}

	// Bucket is full, need to split
	return h.splitAndInsert(orderID, promotionID, offset, hashValue)
}

// splitAndInsert splits the bucket and inserts the new entry
func (h *ExtensibleHash) splitAndInsert(orderID, promotionID uint64, offset int64, hashValue uint64) error {
	bucketIdx := h.getBucketIndex(hashValue)
	bucket := h.directory[bucketIdx]

	// If local depth equals global depth, double the directory
	if bucket.localDepth == h.globalDepth {
		h.doubleDirectory()
		bucketIdx = h.getBucketIndex(hashValue) // Recalculate after doubling
	}

	// Split the bucket
	oldBucket := h.directory[bucketIdx]
	newLocalDepth := oldBucket.localDepth + 1

	// Create two new buckets
	bucket0 := &Bucket{
		localDepth: newLocalDepth,
		entries:    make([]HashEntry, 0, h.bucketSize),
	}
	bucket1 := &Bucket{
		localDepth: newLocalDepth,
		entries:    make([]HashEntry, 0, h.bucketSize),
	}

	// Redistribute existing entries
	splitBit := uint64(1 << (newLocalDepth - 1))
	for _, entry := range oldBucket.entries {
		entryHash := h.hash(entry.OrderID, entry.PromotionID)
		if entryHash&splitBit == 0 {
			bucket0.entries = append(bucket0.entries, entry)
		} else {
			bucket1.entries = append(bucket1.entries, entry)
		}
	}

	// Update directory pointers
	// Find all directory entries pointing to the old bucket and update them
	for i := range h.directory {
		if h.directory[i] == oldBucket {
			if uint64(i)&splitBit == 0 {
				h.directory[i] = bucket0
			} else {
				h.directory[i] = bucket1
			}
		}
	}

	// Now insert the new entry
	newHash := h.hash(orderID, promotionID)
	newBucketIdx := h.getBucketIndex(newHash)
	newBucket := h.directory[newBucketIdx]

	// Check if bucket still full (rare case with many collisions)
	if len(newBucket.entries) >= h.bucketSize {
		// Recursively split
		return h.splitAndInsert(orderID, promotionID, offset, newHash)
	}

	newBucket.entries = append(newBucket.entries, HashEntry{
		OrderID:     orderID,
		PromotionID: promotionID,
		Offset:      offset,
	})

	return nil
}

// doubleDirectory doubles the size of the directory
func (h *ExtensibleHash) doubleDirectory() {
	h.globalDepth++
	newSize := 1 << h.globalDepth
	newDirectory := make([]*Bucket, newSize)

	// Copy existing pointers (each bucket now has two entries pointing to it)
	for i := range h.directory {
		newDirectory[i] = h.directory[i]
		newDirectory[i+len(h.directory)] = h.directory[i]
	}

	h.directory = newDirectory
}

// Search finds the offset for a composite key
func (h *ExtensibleHash) Search(orderID, promotionID uint64) (int64, bool) {
	hashValue := h.hash(orderID, promotionID)
	bucketIdx := h.getBucketIndex(hashValue)
	bucket := h.directory[bucketIdx]

	for _, entry := range bucket.entries {
		if entry.OrderID == orderID && entry.PromotionID == promotionID {
			return entry.Offset, true
		}
	}

	return 0, false
}

// Delete removes an entry from the hash index
func (h *ExtensibleHash) Delete(orderID, promotionID uint64) error {
	hashValue := h.hash(orderID, promotionID)
	bucketIdx := h.getBucketIndex(hashValue)
	bucket := h.directory[bucketIdx]

	for i, entry := range bucket.entries {
		if entry.OrderID == orderID && entry.PromotionID == promotionID {
			// Remove entry by swapping with last and truncating
			bucket.entries[i] = bucket.entries[len(bucket.entries)-1]
			bucket.entries = bucket.entries[:len(bucket.entries)-1]
			return nil
		}
	}

	return fmt.Errorf("key not found: orderID=%d, promotionID=%d", orderID, promotionID)
}

// GetByOrderID returns all entries with a given orderID
func (h *ExtensibleHash) GetByOrderID(orderID uint64) []HashEntry {
	result := make([]HashEntry, 0)
	seen := make(map[*Bucket]bool)

	for _, bucket := range h.directory {
		if seen[bucket] {
			continue
		}
		seen[bucket] = true

		for _, entry := range bucket.entries {
			if entry.OrderID == orderID {
				result = append(result, entry)
			}
		}
	}

	return result
}

// GetByPromotionID returns all entries with a given promotionID
func (h *ExtensibleHash) GetByPromotionID(promotionID uint64) []HashEntry {
	result := make([]HashEntry, 0)
	seen := make(map[*Bucket]bool)

	for _, bucket := range h.directory {
		if seen[bucket] {
			continue
		}
		seen[bucket] = true

		for _, entry := range bucket.entries {
			if entry.PromotionID == promotionID {
				result = append(result, entry)
			}
		}
	}

	return result
}

// GetAll returns all entries in the hash index
func (h *ExtensibleHash) GetAll() []HashEntry {
	result := make([]HashEntry, 0)
	seen := make(map[*Bucket]bool)

	for _, bucket := range h.directory {
		if seen[bucket] {
			continue
		}
		seen[bucket] = true

		result = append(result, bucket.entries...)
	}

	return result
}

// Size returns the total number of entries
func (h *ExtensibleHash) Size() int {
	count := 0
	seen := make(map[*Bucket]bool)

	for _, bucket := range h.directory {
		if seen[bucket] {
			continue
		}
		seen[bucket] = true
		count += len(bucket.entries)
	}

	return count
}

// GetGlobalDepth returns the current global depth
func (h *ExtensibleHash) GetGlobalDepth() int {
	return h.globalDepth
}

// GetDirectorySize returns the current directory size
func (h *ExtensibleHash) GetDirectorySize() int {
	return len(h.directory)
}

// Save persists the hash index to a file
func (h *ExtensibleHash) Save(filePath string) error {
	// Ensure parent directory exists
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create index file: %w", err)
	}
	defer file.Close()

	// Write header: globalDepth (4 bytes) + bucketSize (4 bytes)
	header := make([]byte, 8)
	binary.LittleEndian.PutUint32(header[0:4], uint32(h.globalDepth))
	binary.LittleEndian.PutUint32(header[4:8], uint32(h.bucketSize))
	if _, err := file.Write(header); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	// Collect unique buckets and assign IDs
	bucketIDs := make(map[*Bucket]uint32)
	uniqueBuckets := make([]*Bucket, 0)
	for _, bucket := range h.directory {
		if _, exists := bucketIDs[bucket]; !exists {
			bucketIDs[bucket] = uint32(len(uniqueBuckets))
			uniqueBuckets = append(uniqueBuckets, bucket)
		}
	}

	// Write number of unique buckets (4 bytes)
	numBuckets := make([]byte, 4)
	binary.LittleEndian.PutUint32(numBuckets, uint32(len(uniqueBuckets)))
	if _, err := file.Write(numBuckets); err != nil {
		return fmt.Errorf("failed to write bucket count: %w", err)
	}

	// Write each unique bucket
	for _, bucket := range uniqueBuckets {
		// Local depth (4 bytes) + entry count (4 bytes)
		bucketHeader := make([]byte, 8)
		binary.LittleEndian.PutUint32(bucketHeader[0:4], uint32(bucket.localDepth))
		binary.LittleEndian.PutUint32(bucketHeader[4:8], uint32(len(bucket.entries)))
		if _, err := file.Write(bucketHeader); err != nil {
			return fmt.Errorf("failed to write bucket header: %w", err)
		}

		// Write entries: orderID (8) + promotionID (8) + offset (8) = 24 bytes each
		for _, entry := range bucket.entries {
			entryData := make([]byte, 24)
			binary.LittleEndian.PutUint64(entryData[0:8], entry.OrderID)
			binary.LittleEndian.PutUint64(entryData[8:16], entry.PromotionID)
			binary.LittleEndian.PutUint64(entryData[16:24], uint64(entry.Offset))
			if _, err := file.Write(entryData); err != nil {
				return fmt.Errorf("failed to write entry: %w", err)
			}
		}
	}

	// Write directory: bucket ID for each entry (4 bytes each)
	for _, bucket := range h.directory {
		dirEntry := make([]byte, 4)
		binary.LittleEndian.PutUint32(dirEntry, bucketIDs[bucket])
		if _, err := file.Write(dirEntry); err != nil {
			return fmt.Errorf("failed to write directory entry: %w", err)
		}
	}

	return nil
}

// LoadExtensibleHash loads a hash index from a file
func LoadExtensibleHash(filePath string) (*ExtensibleHash, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open index file: %w", err)
	}
	defer file.Close()

	// Read header
	header := make([]byte, 8)
	if _, err := file.Read(header); err != nil {
		return nil, fmt.Errorf("failed to read header: %w", err)
	}

	globalDepth := int(binary.LittleEndian.Uint32(header[0:4]))
	bucketSize := int(binary.LittleEndian.Uint32(header[4:8]))

	// Read number of unique buckets
	numBucketsData := make([]byte, 4)
	if _, err := file.Read(numBucketsData); err != nil {
		return nil, fmt.Errorf("failed to read bucket count: %w", err)
	}
	numBuckets := int(binary.LittleEndian.Uint32(numBucketsData))

	// Read unique buckets
	buckets := make([]*Bucket, numBuckets)
	for i := 0; i < numBuckets; i++ {
		bucketHeader := make([]byte, 8)
		if _, err := file.Read(bucketHeader); err != nil {
			return nil, fmt.Errorf("failed to read bucket header: %w", err)
		}

		localDepth := int(binary.LittleEndian.Uint32(bucketHeader[0:4]))
		entryCount := int(binary.LittleEndian.Uint32(bucketHeader[4:8]))

		entries := make([]HashEntry, entryCount)
		for j := 0; j < entryCount; j++ {
			entryData := make([]byte, 24)
			if _, err := file.Read(entryData); err != nil {
				return nil, fmt.Errorf("failed to read entry: %w", err)
			}

			entries[j] = HashEntry{
				OrderID:     binary.LittleEndian.Uint64(entryData[0:8]),
				PromotionID: binary.LittleEndian.Uint64(entryData[8:16]),
				Offset:      int64(binary.LittleEndian.Uint64(entryData[16:24])),
			}
		}

		buckets[i] = &Bucket{
			localDepth: localDepth,
			entries:    entries,
		}
	}

	// Read directory
	dirSize := 1 << globalDepth
	directory := make([]*Bucket, dirSize)
	for i := 0; i < dirSize; i++ {
		dirEntry := make([]byte, 4)
		if _, err := file.Read(dirEntry); err != nil {
			return nil, fmt.Errorf("failed to read directory entry: %w", err)
		}
		bucketID := binary.LittleEndian.Uint32(dirEntry)
		directory[i] = buckets[bucketID]
	}

	return &ExtensibleHash{
		globalDepth: globalDepth,
		bucketSize:  bucketSize,
		directory:   directory,
	}, nil
}
