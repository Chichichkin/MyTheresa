package models

type Category struct {
	ID       uint      `gorm:"primaryKey"`
	Code     string    `gorm:"unique"`
	Name     string    `gorm:"not null"`
	Products []Product `gorm:"foreignKey:CategoryID"`
}

func (p *Category) TableName() string {
	return "categories"
}
