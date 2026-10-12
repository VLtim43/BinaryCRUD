package utils

import (
	"errors"
	"fmt"
	"math"
)

// Validation constants
const (
	// MaxNameLength is the maximum allowed length for names (customer, item, promotion)
	MaxNameLength = 255

	// MaxItemsPerCollection is the maximum number of items allowed in an order/promotion
	MaxItemsPerCollection = 1000

	// MaxRecordSize is the maximum allowed size for a single record (1MB)
	MaxRecordSize = 1 << 20

	// MaxPrice is the maximum price in cents (max uint32 = ~$42.9 million)
	MaxPrice = math.MaxUint32
)

// Validation errors
var (
	ErrNameEmpty      = errors.New("name cannot be empty")
	ErrNameTooLong    = fmt.Errorf("name exceeds maximum length of %d characters", MaxNameLength)
	ErrNoItems        = errors.New("must contain at least one item")
	ErrTooManyItems   = fmt.Errorf("exceeds maximum of %d items", MaxItemsPerCollection)
	ErrPriceOverflow  = errors.New("price calculation would overflow")
	ErrRecordTooLarge = fmt.Errorf("record size exceeds maximum of %d bytes", MaxRecordSize)
)

// ValidateName validates a name string (customer name, item name, promotion name)
func ValidateName(name string) error {
	if len(name) == 0 {
		return ErrNameEmpty
	}
	if len(name) > MaxNameLength {
		return ErrNameTooLong
	}
	return nil
}

// ValidateItemIDs validates a slice of item IDs for collections
func ValidateItemIDs(itemIDs []uint64) error {
	if len(itemIDs) == 0 {
		return ErrNoItems
	}
	if len(itemIDs) > MaxItemsPerCollection {
		return ErrTooManyItems
	}
	return nil
}

// ValidatePrice validates that a price is within acceptable bounds
func ValidatePrice(priceInCents uint64) error {
	if priceInCents > MaxPrice {
		return fmt.Errorf("price %d exceeds maximum of %d cents", priceInCents, MaxPrice)
	}
	return nil
}

// SafeAddUint64 adds two uint64 values with overflow checking
func SafeAddUint64(a, b uint64) (uint64, error) {
	if a > math.MaxUint64-b {
		return 0, ErrPriceOverflow
	}
	return a + b, nil
}

// ValidateRecordLength validates that a record length is within acceptable bounds
func ValidateRecordLength(length uint64) error {
	if length == 0 {
		return errors.New("record length cannot be zero")
	}
	if length > MaxRecordSize {
		return ErrRecordTooLarge
	}
	return nil
}

// ValidateOffset validates that an offset is within file bounds
func ValidateOffset(offset int64, fileSize int64) error {
	if offset < 0 {
		return fmt.Errorf("invalid negative offset: %d", offset)
	}
	if offset >= fileSize {
		return fmt.Errorf("offset %d is beyond file size %d", offset, fileSize)
	}
	return nil
}
