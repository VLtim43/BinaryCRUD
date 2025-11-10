package dao

// PromotionDAO wraps CollectionDAO for promotions
type PromotionDAO struct {
	*CollectionDAO
}

// NewPromotionDAO creates a DAO for promotions.bin
func NewPromotionDAO(filePath string) *PromotionDAO {
	return &PromotionDAO{
		CollectionDAO: &CollectionDAO{filePath: filePath},
	}
}
