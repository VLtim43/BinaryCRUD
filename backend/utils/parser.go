package utils

import (
	"fmt"
)

// Item represents a parsed item entry
type Item struct {
	ID        uint64
	Name      string
	Price     uint64
	Tombstone byte
}

// Collection represents a parsed collection (order/promotion) entry
type Collection struct {
	ID          uint64
	OwnerOrName string
	TotalPrice  uint64
	ItemCount   uint64
	ItemIDs     []uint64
	Tombstone   byte
}

// OrderPromotion represents a parsed order-promotion relationship entry
type OrderPromotion struct {
	OrderID     uint64
	PromotionID uint64
	Tombstone   byte
}

// ParseItemEntry parses a binary item entry
// Format: [ID(2)][tombstone(1)][nameLength(2)][name...][price(4)]
func ParseItemEntry(entryData []byte) (*Item, error) {
	parseOffset := 0

	// Read ID
	entryID, parseOffset, err := ReadFixedNumber(IDSize, entryData, parseOffset)
	if err != nil {
		return nil, fmt.Errorf("failed to read ID: %w", err)
	}

	// Read tombstone byte
	if parseOffset >= len(entryData) {
		return nil, fmt.Errorf("entry too short for tombstone")
	}
	tombstone := entryData[parseOffset]
	parseOffset += TombstoneSize

	// Read name size
	nameSize, parseOffset, err := ReadFixedNumber(2, entryData, parseOffset)
	if err != nil {
		return nil, fmt.Errorf("failed to read name size: %w", err)
	}

	// Read name
	name, parseOffset, err := ReadFixedString(int(nameSize), entryData, parseOffset)
	if err != nil {
		return nil, fmt.Errorf("failed to read name: %w", err)
	}

	// Read price
	price, _, err := ReadFixedNumber(4, entryData, parseOffset)
	if err != nil {
		return nil, fmt.Errorf("failed to read price: %w", err)
	}

	return &Item{
		ID:        entryID,
		Name:      name,
		Price:     price,
		Tombstone: tombstone,
	}, nil
}

// ParseCollectionEntry parses a binary collection (order/promotion) entry
// Format: [ID(2)][tombstone(1)][nameLength(2)][name...][totalPrice(4)][itemCount(4)][itemIDs...]
func ParseCollectionEntry(entryData []byte) (*Collection, error) {
	parseOffset := 0

	// Read ID
	entryID, parseOffset, err := ReadFixedNumber(IDSize, entryData, parseOffset)
	if err != nil {
		return nil, fmt.Errorf("failed to read ID: %w", err)
	}

	// Read tombstone byte
	if parseOffset >= len(entryData) {
		return nil, fmt.Errorf("entry too short for tombstone")
	}
	tombstone := entryData[parseOffset]
	parseOffset += TombstoneSize

	// Read name size
	nameSize, parseOffset, err := ReadFixedNumber(2, entryData, parseOffset)
	if err != nil {
		return nil, fmt.Errorf("failed to read name size: %w", err)
	}

	// Read name
	ownerOrName, parseOffset, err := ReadFixedString(int(nameSize), entryData, parseOffset)
	if err != nil {
		return nil, fmt.Errorf("failed to read name: %w", err)
	}

	// Read total price
	totalPrice, parseOffset, err := ReadFixedNumber(4, entryData, parseOffset)
	if err != nil {
		return nil, fmt.Errorf("failed to read total price: %w", err)
	}

	// Read item count
	itemCount, parseOffset, err := ReadFixedNumber(4, entryData, parseOffset)
	if err != nil {
		return nil, fmt.Errorf("failed to read item count: %w", err)
	}

	// Read item IDs (2 bytes each)
	itemIDs := make([]uint64, itemCount)
	for i := uint64(0); i < itemCount; i++ {
		itemID, newOffset, err := ReadFixedNumber(IDSize, entryData, parseOffset)
		if err != nil {
			return nil, fmt.Errorf("failed to read item ID %d: %w", i, err)
		}
		itemIDs[i] = itemID
		parseOffset = newOffset
	}

	return &Collection{
		ID:          entryID,
		OwnerOrName: ownerOrName,
		TotalPrice:  totalPrice,
		ItemCount:   itemCount,
		ItemIDs:     itemIDs,
		Tombstone:   tombstone,
	}, nil
}

// ParseOrderPromotionEntry parses a binary order-promotion relationship entry
// Format: [orderID(2)][promotionID(2)][tombstone(1)]
func ParseOrderPromotionEntry(entryData []byte) (*OrderPromotion, error) {
	if len(entryData) < IDSize*2+TombstoneSize {
		return nil, fmt.Errorf("entry too short: expected at least %d bytes, got %d", IDSize*2+TombstoneSize, len(entryData))
	}

	offset := 0

	// Read orderID
	orderID, newOffset, err := ReadFixedNumber(IDSize, entryData, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to read order ID: %w", err)
	}
	offset = newOffset

	// Read promotionID
	promotionID, newOffset, err := ReadFixedNumber(IDSize, entryData, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to read promotion ID: %w", err)
	}
	offset = newOffset

	// Read tombstone
	if offset >= len(entryData) {
		return nil, fmt.Errorf("entry too short for tombstone")
	}
	tombstone := entryData[offset]

	return &OrderPromotion{
		OrderID:     orderID,
		PromotionID: promotionID,
		Tombstone:   tombstone,
	}, nil
}
