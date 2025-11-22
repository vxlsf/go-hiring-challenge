package models

import (
	"github.com/shopspring/decimal"
)

// Product represents a product in the catalog.
// It includes a unique code and a price.
type Product struct {
	ID         uint            `gorm:"primaryKey" json:"id"`
	Code       string          `gorm:"uniqueIndex;not null" json:"code"`
	Price      decimal.Decimal `gorm:"type:decimal(10,2);not null" json:"price"`
	CategoryID uint            `gorm:"not null" json:"category_id"`
	Category   Category        `gorm:"foreignKey:CategoryID" json:"category"`
	Variants   []Variant       `gorm:"foreignKey:ProductID" json:"variants"`
}

func (p *Product) TableName() string {
	return "products"
}
