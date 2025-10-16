package dao

// Binary structure of order.bin:
//
// HEADER:
//   [Count:4 bytes][0x1E]
//
// ORDER RECORD:
//   [Tombstone:1 byte][0x1F][ItemCount:4 bytes][0x1F][Items...][0x1F][Timestamp:8 bytes][0x1E]
//
// EACH ITEM IN ITEMS:
//   [ItemID:4 bytes][0x1F][Quantity:4 bytes][0x1F]
//
// Notes:
//   - All multi-byte integers are little-endian
//   - Tombstone: 0=active, 1=deleted
//   - RecordID is assigned from Count before incrementing
//   - 0x1F = Unit Separator, 0x1E = Record Separator
//   - Items array is a sequence of ItemID/Quantity pairs, each followed by 0x1F

import (
	"BinaryCRUD/backend/serialization"
	"fmt"
)

type OrderDAO struct {
	filename string
}

func NewOrderDAO(filename string) *OrderDAO {
	return &OrderDAO{
		filename: filename,
	}
}

func (dao *OrderDAO) Write(items []serialization.OrderItem) error {
	// Append entry and get result with recordID and offset
	_, err := serialization.AppendOrder(dao.filename, items)
	if err != nil {
		return err
	}

	return nil
}

func (dao *OrderDAO) Read() ([]serialization.Order, error) {
	return serialization.ReadAllOrders(dao.filename)
}

func (dao *OrderDAO) Print() (string, error) {
	return serialization.PrintOrderBinaryFile(dao.filename)
}

// GetByID retrieves an order by its record ID using sequential search
// Note: Orders are not indexed, so this performs a sequential scan
func (dao *OrderDAO) GetByID(recordID uint32) (*serialization.Order, error) {
	orders, err := serialization.ReadAllOrders(dao.filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read orders: %w", err)
	}

	// Sequential search through orders
	for _, order := range orders {
		if order.RecordID == recordID {
			return &order, nil
		}
	}

	return nil, fmt.Errorf("order with ID %d not found", recordID)
}
