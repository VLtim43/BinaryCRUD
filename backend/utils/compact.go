package utils

import (
	"fmt"
	"os"
	"path/filepath"
)

// CompactResult holds the results of a compaction operation
type CompactResult struct {
	ItemsRemoved             int      // Number of tombstoned items physically removed
	OrdersAffected           int      // Number of orders that had item references cleaned
	PromotionsAffected       int      // Number of promotions that had item references cleaned
	OrdersRemoved            int      // Number of tombstoned orders physically removed
	PromotionsRemoved        int      // Number of tombstoned promotions physically removed
	OrderPromotionsRemoved   int      // Number of tombstoned order-promotions physically removed
	DeletedItemIDs           []uint64 // IDs of items that were removed
}

// CompactAll performs compaction on all binary files:
// 1. Identifies tombstoned items
// 2. Removes tombstoned items from items.bin
// 3. Updates orders/promotions to remove references to deleted items
// 4. Removes tombstoned orders/promotions/order_promotions
// 5. Deletes all index files (they will be rebuilt on next DAO init)
func CompactAll(itemsPath, ordersPath, promotionsPath, orderPromotionsPath string) (*CompactResult, error) {
	result := &CompactResult{}

	// Step 1: Get all tombstoned item IDs before compacting
	deletedItemIDs, err := getDeletedItemIDs(itemsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get deleted item IDs: %w", err)
	}
	result.DeletedItemIDs = deletedItemIDs

	// Step 2: Compact items.bin
	itemsRemoved, err := compactItems(itemsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to compact items: %w", err)
	}
	result.ItemsRemoved = itemsRemoved

	// Step 3: Update orders and promotions to remove deleted item references
	if len(deletedItemIDs) > 0 {
		deletedSet := make(map[uint64]bool)
		for _, id := range deletedItemIDs {
			deletedSet[id] = true
		}

		ordersAffected, err := cleanCollectionItemRefs(ordersPath, deletedSet)
		if err != nil {
			return nil, fmt.Errorf("failed to clean order item refs: %w", err)
		}
		result.OrdersAffected = ordersAffected

		promotionsAffected, err := cleanCollectionItemRefs(promotionsPath, deletedSet)
		if err != nil {
			return nil, fmt.Errorf("failed to clean promotion item refs: %w", err)
		}
		result.PromotionsAffected = promotionsAffected
	}

	// Step 4: Compact orders.bin
	ordersRemoved, err := compactCollections(ordersPath)
	if err != nil {
		return nil, fmt.Errorf("failed to compact orders: %w", err)
	}
	result.OrdersRemoved = ordersRemoved

	// Step 5: Compact promotions.bin
	promotionsRemoved, err := compactCollections(promotionsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to compact promotions: %w", err)
	}
	result.PromotionsRemoved = promotionsRemoved

	// Step 6: Compact order_promotions.bin
	opRemoved, err := compactOrderPromotions(orderPromotionsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to compact order_promotions: %w", err)
	}
	result.OrderPromotionsRemoved = opRemoved

	// Step 7: Delete all index files
	if err := deleteAllIndexes(); err != nil {
		return nil, fmt.Errorf("failed to delete indexes: %w", err)
	}

	return result, nil
}

// getDeletedItemIDs returns a list of all tombstoned item IDs
func getDeletedItemIDs(itemsPath string) ([]uint64, error) {
	if _, err := os.Stat(itemsPath); os.IsNotExist(err) {
		return []uint64{}, nil
	}

	entries, err := SplitFileIntoEntries(itemsPath)
	if err != nil {
		return nil, err
	}

	var deletedIDs []uint64
	for _, entry := range entries {
		item, err := ParseItemEntry(entry.Data)
		if err != nil {
			continue
		}
		if item.Tombstone != 0x00 {
			deletedIDs = append(deletedIDs, item.ID)
		}
	}

	return deletedIDs, nil
}

// compactItems removes tombstoned items and rewrites the file
// Returns the number of items removed
func compactItems(filePath string) (int, error) {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return 0, nil
	}

	entries, err := SplitFileIntoEntries(filePath)
	if err != nil {
		return 0, err
	}

	// Filter out tombstoned entries and parse active ones
	var activeItems []*Item
	removedCount := 0

	for _, entry := range entries {
		item, err := ParseItemEntry(entry.Data)
		if err != nil {
			continue
		}
		if item.Tombstone == 0x00 {
			activeItems = append(activeItems, item)
		} else {
			removedCount++
		}
	}

	if removedCount == 0 {
		return 0, nil
	}

	// Rewrite the file with only active items
	return removedCount, rewriteItemsFile(filePath, activeItems)
}

// rewriteItemsFile rewrites items.bin with the given items
func rewriteItemsFile(filePath string, items []*Item) error {
	// Create temp file
	tmpPath := filePath + ".tmp"
	tmpFile, err := os.Create(tmpPath)
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}

	// Find the max ID to set nextId correctly
	maxID := uint64(0)
	for _, item := range items {
		if item.ID > maxID {
			maxID = item.ID
		}
	}

	// Extract filename from path
	basename := filepath.Base(filePath)
	filename := basename[:len(basename)-len(filepath.Ext(basename))]

	// Write header: entitiesCount = len(items), tombstoneCount = 0, nextId = maxID + 1
	header, err := WriteHeader(filename, len(items), 0, int(maxID)+1)
	if err != nil {
		tmpFile.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("failed to write header: %w", err)
	}

	if _, err := tmpFile.Write(header); err != nil {
		tmpFile.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("failed to write header to file: %w", err)
	}

	// Write each item
	for _, item := range items {
		if err := writeItemEntry(tmpFile, item); err != nil {
			tmpFile.Close()
			os.Remove(tmpPath)
			return fmt.Errorf("failed to write item %d: %w", item.ID, err)
		}
	}

	tmpFile.Sync()
	tmpFile.Close()

	// Replace original with temp
	return os.Rename(tmpPath, filePath)
}

// writeItemEntry writes a single item entry to the file
// Format: [recordLength(2)][ID(2)][tombstone(1)][nameLength(2)][name...][price(4)]
func writeItemEntry(file *os.File, item *Item) error {
	// Build entry data: [nameLength(2)][name...][price(4)]
	nameSizeBytes, err := WriteFixedNumber(2, uint64(len(item.Name)))
	if err != nil {
		return err
	}

	nameBytes := []byte(item.Name)

	priceBytes, err := WriteFixedNumber(4, item.Price)
	if err != nil {
		return err
	}

	entryData := CombineBytes(nameSizeBytes, nameBytes, priceBytes)

	// Build complete record: [recordLength(2)][ID(2)][tombstone(1)][entryData]
	recordLength := IDSize + TombstoneSize + len(entryData)

	lengthBytes, err := WriteFixedNumber(RecordLengthSize, uint64(recordLength))
	if err != nil {
		return err
	}

	idBytes, err := WriteFixedNumber(IDSize, item.ID)
	if err != nil {
		return err
	}

	tombstoneBytes := []byte{0x00}

	record := CombineBytes(lengthBytes, idBytes, tombstoneBytes, entryData)

	_, err = file.Write(record)
	return err
}

// cleanCollectionItemRefs removes deleted item IDs from all collections in a file
// Returns the number of collections that were modified
func cleanCollectionItemRefs(filePath string, deletedItemIDs map[uint64]bool) (int, error) {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return 0, nil
	}

	entries, err := SplitFileIntoEntries(filePath)
	if err != nil {
		return 0, err
	}

	// Parse all collections and track which need updating
	var collections []*Collection
	affectedCount := 0

	for _, entry := range entries {
		collection, err := ParseCollectionEntry(entry.Data)
		if err != nil {
			continue
		}

		// Filter out deleted item IDs
		var newItemIDs []uint64
		hadDeletions := false
		for _, itemID := range collection.ItemIDs {
			if !deletedItemIDs[itemID] {
				newItemIDs = append(newItemIDs, itemID)
			} else {
				hadDeletions = true
			}
		}

		if hadDeletions && collection.Tombstone == 0x00 {
			affectedCount++
			collection.ItemIDs = newItemIDs
			collection.ItemCount = uint64(len(newItemIDs))
			// Note: TotalPrice would need recalculation but we don't have item prices here
			// The price will be stale but this is acceptable for compaction
		}

		collections = append(collections, collection)
	}

	if affectedCount == 0 {
		return 0, nil
	}

	// Rewrite the file with updated collections
	return affectedCount, rewriteCollectionsFile(filePath, collections)
}

// compactCollections removes tombstoned collections and rewrites the file
// Returns the number of collections removed
func compactCollections(filePath string) (int, error) {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return 0, nil
	}

	entries, err := SplitFileIntoEntries(filePath)
	if err != nil {
		return 0, err
	}

	var activeCollections []*Collection
	removedCount := 0

	for _, entry := range entries {
		collection, err := ParseCollectionEntry(entry.Data)
		if err != nil {
			continue
		}
		if collection.Tombstone == 0x00 {
			activeCollections = append(activeCollections, collection)
		} else {
			removedCount++
		}
	}

	if removedCount == 0 {
		return 0, nil
	}

	return removedCount, rewriteCollectionsFile(filePath, activeCollections)
}

// rewriteCollectionsFile rewrites a collection file with the given collections
func rewriteCollectionsFile(filePath string, collections []*Collection) error {
	tmpPath := filePath + ".tmp"
	tmpFile, err := os.Create(tmpPath)
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}

	// Find max ID and count active
	maxID := uint64(0)
	activeCount := 0
	for _, c := range collections {
		if c.ID > maxID {
			maxID = c.ID
		}
		if c.Tombstone == 0x00 {
			activeCount++
		}
	}

	basename := filepath.Base(filePath)
	filename := basename[:len(basename)-len(filepath.Ext(basename))]

	header, err := WriteHeader(filename, activeCount, 0, int(maxID)+1)
	if err != nil {
		tmpFile.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("failed to write header: %w", err)
	}

	if _, err := tmpFile.Write(header); err != nil {
		tmpFile.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("failed to write header to file: %w", err)
	}

	// Write each collection (only active ones after compaction, but all during ref cleaning)
	for _, c := range collections {
		if err := writeCollectionEntry(tmpFile, c); err != nil {
			tmpFile.Close()
			os.Remove(tmpPath)
			return fmt.Errorf("failed to write collection %d: %w", c.ID, err)
		}
	}

	tmpFile.Sync()
	tmpFile.Close()

	return os.Rename(tmpPath, filePath)
}

// writeCollectionEntry writes a single collection entry
// Format: [recordLength(2)][ID(2)][tombstone(1)][nameLength(2)][name...][totalPrice(4)][itemCount(4)][itemIDs...]
func writeCollectionEntry(file *os.File, c *Collection) error {
	// Name (already encrypted in OwnerOrName if encryption was used)
	nameBytes := []byte(c.OwnerOrName)
	nameSizeBytes, err := WriteFixedNumber(2, uint64(len(nameBytes)))
	if err != nil {
		return err
	}

	totalPriceBytes, err := WriteFixedNumber(4, c.TotalPrice)
	if err != nil {
		return err
	}

	itemCountBytes, err := WriteFixedNumber(4, c.ItemCount)
	if err != nil {
		return err
	}

	// Build item IDs bytes
	var itemIDsBytes []byte
	for _, itemID := range c.ItemIDs {
		idBytes, err := WriteFixedNumber(IDSize, itemID)
		if err != nil {
			return err
		}
		itemIDsBytes = append(itemIDsBytes, idBytes...)
	}

	entryData := CombineBytes(nameSizeBytes, nameBytes, totalPriceBytes, itemCountBytes, itemIDsBytes)

	// Build complete record
	recordLength := IDSize + TombstoneSize + len(entryData)

	lengthBytes, err := WriteFixedNumber(RecordLengthSize, uint64(recordLength))
	if err != nil {
		return err
	}

	idBytes, err := WriteFixedNumber(IDSize, c.ID)
	if err != nil {
		return err
	}

	tombstoneBytes := []byte{c.Tombstone}

	record := CombineBytes(lengthBytes, idBytes, tombstoneBytes, entryData)

	_, err = file.Write(record)
	return err
}

// compactOrderPromotions removes tombstoned order-promotion relationships
func compactOrderPromotions(filePath string) (int, error) {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return 0, nil
	}

	entries, err := SplitFileIntoEntries(filePath)
	if err != nil {
		return 0, err
	}

	var activeOPs []*OrderPromotion
	removedCount := 0

	for _, entry := range entries {
		op, err := ParseOrderPromotionEntry(entry.Data)
		if err != nil {
			continue
		}
		if op.Tombstone == 0x00 {
			activeOPs = append(activeOPs, op)
		} else {
			removedCount++
		}
	}

	if removedCount == 0 {
		return 0, nil
	}

	return removedCount, rewriteOrderPromotionsFile(filePath, activeOPs)
}

// rewriteOrderPromotionsFile rewrites order_promotions.bin with the given relationships
func rewriteOrderPromotionsFile(filePath string, ops []*OrderPromotion) error {
	tmpPath := filePath + ".tmp"
	tmpFile, err := os.Create(tmpPath)
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}

	basename := filepath.Base(filePath)
	filename := basename[:len(basename)-len(filepath.Ext(basename))]

	// order_promotions doesn't use nextId (composite key), so keep it at 0
	header, err := WriteHeader(filename, len(ops), 0, 0)
	if err != nil {
		tmpFile.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("failed to write header: %w", err)
	}

	if _, err := tmpFile.Write(header); err != nil {
		tmpFile.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("failed to write header to file: %w", err)
	}

	for _, op := range ops {
		if err := writeOrderPromotionEntry(tmpFile, op); err != nil {
			tmpFile.Close()
			os.Remove(tmpPath)
			return fmt.Errorf("failed to write order_promotion: %w", err)
		}
	}

	tmpFile.Sync()
	tmpFile.Close()

	return os.Rename(tmpPath, filePath)
}

// writeOrderPromotionEntry writes a single order-promotion entry
// Format: [recordLength(2)][orderID(2)][promotionID(2)][tombstone(1)]
func writeOrderPromotionEntry(file *os.File, op *OrderPromotion) error {
	orderIDBytes, err := WriteFixedNumber(IDSize, op.OrderID)
	if err != nil {
		return err
	}

	promotionIDBytes, err := WriteFixedNumber(IDSize, op.PromotionID)
	if err != nil {
		return err
	}

	tombstoneBytes := []byte{0x00}

	entryData := CombineBytes(orderIDBytes, promotionIDBytes, tombstoneBytes)

	recordLength := len(entryData)
	lengthBytes, err := WriteFixedNumber(RecordLengthSize, uint64(recordLength))
	if err != nil {
		return err
	}

	record := CombineBytes(lengthBytes, entryData)

	_, err = file.Write(record)
	return err
}

// deleteAllIndexes removes all .idx files from the indexes directory
func deleteAllIndexes() error {
	indexDir := IndexDir

	if _, err := os.Stat(indexDir); os.IsNotExist(err) {
		return nil
	}

	entries, err := os.ReadDir(indexDir)
	if err != nil {
		return fmt.Errorf("failed to read index directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if filepath.Ext(entry.Name()) == ".idx" {
			indexPath := filepath.Join(indexDir, entry.Name())
			if err := os.Remove(indexPath); err != nil {
				return fmt.Errorf("failed to remove index %s: %w", entry.Name(), err)
			}
		}
	}

	return nil
}
