package dao

// OrderDAO wraps CollectionDAO for orders
type OrderDAO struct {
	*CollectionDAO
}

// NewOrderDAO creates a DAO for orders.bin
func NewOrderDAO(filePath string) *OrderDAO {
	return &OrderDAO{
		CollectionDAO: &CollectionDAO{filePath: filePath},
	}
}
