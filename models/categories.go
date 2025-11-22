package models

// Category represents the category of a product in the catalog.
// It includes a human-readable unique code and a human-readable name.
type Category struct {
	ID       uint      `gorm:"primaryKey" json:"id"`
	Code     string    `gorm:"uniqueIndex;not null" json:"code"`
	Name     string    `gorm:"not null" json:"name"`
	Products []Product `gorm:"foreignKey:CategoryID" json:"products,omitempty"`
}

type CreateCategoryRequest struct {
	Code string `json:"code" binding:"required"`
	Name string `json:"name" binding:"required"`
}

func (c *Category) TableName() string {
	return "categories"
}
