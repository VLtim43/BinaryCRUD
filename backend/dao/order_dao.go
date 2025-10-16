package dao

import (
	"BinaryCRUD/backend/serialization"
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
