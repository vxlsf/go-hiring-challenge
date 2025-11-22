package models

import (
	"gorm.io/gorm"
)

type ProductsRepositoryInterface interface {
	GetProductsWithFilters(categoryCode string, priceLessThan *float64, offset, limit int) ([]Product, int64, error)
	GetProductByCode(code string) (*Product, error)
}

type ProductsRepository struct {
	db *gorm.DB
}

func NewProductsRepository(db *gorm.DB) *ProductsRepository {
	return &ProductsRepository{
		db: db,
	}
}

func (r *ProductsRepository) GetProductsWithFilters(categoryCode string, priceLessThan *float64, offset, limit int) ([]Product, int64, error) {
	var products []Product
	var total int64

	query := r.db.Model(&Product{}).Preload("Category").Preload("Variants")

	if categoryCode != "" {
		query = query.Joins("JOIN categories ON products.category_id = categories.id").
			Where("categories.code = ?", categoryCode)
	}

	if priceLessThan != nil {
		query = query.Where("price < ?", *priceLessThan)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Order("id ASC").Offset(offset).Limit(limit).Find(&products).Error; err != nil {
		return nil, 0, err
	}

	return products, total, nil
}

func (r *ProductsRepository) GetProductByCode(code string) (*Product, error) {
	var product Product
	if err := r.db.Preload("Category").Preload("Variants").
		Where("code = ?", code).First(&product).Error; err != nil {
		return nil, err
	}
	return &product, nil
}
