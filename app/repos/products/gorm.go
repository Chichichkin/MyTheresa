package products

import (
	"context"
	"errors"

	"github.com/mytheresa/go-hiring-challenge/models"
	"gorm.io/gorm"
)

type GormRepo struct {
	db *gorm.DB
}

func NewGormRepo(db *gorm.DB) *GormRepo {
	return &GormRepo{
		db: db,
	}
}

func (r *GormRepo) ListAll(ctx context.Context) ([]models.Product, error) {
	var products []models.Product
	err := r.db.WithContext(ctx).
		Preload("Variants").
		Find(&products).
		Error
	if err != nil {
		return nil, err
	}
	return products, nil
}

func (r *GormRepo) List(
	ctx context.Context,
	filters SearchFilters,
) ([]models.Product, error) {
	query := r.db.WithContext(ctx).
		Model(&models.Product{}).
		Preload("Variants").
		Preload("Category")

	if filters.Category != "" {
		query = query.
			Joins("JOIN categories ON categories.id = products.category_id").
			Where("categories.code = ?", filters.Category)
	}
	if filters.PriceLessThan != nil {
		query = query.Where("price < ?", *filters.PriceLessThan)
	}

	var products []models.Product
	err := query.Order("products.id ASC").
		Offset(filters.Offset).
		Limit(filters.Limit).
		Find(&products).
		Error
	if err != nil {
		return nil, err
	}
	return products, nil
}

func (r *GormRepo) GetByID(ctx context.Context, id string) (models.Product, error) {
	product := models.Product{}
	err := r.db.WithContext(ctx).
		Preload("Variants").
		First(&product, id).
		Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.Product{}, nil
		}
		return models.Product{}, err
	}
	return product, nil
}

func (r *GormRepo) GetByCode(ctx context.Context, code string) (models.Product, error) {
	product := models.Product{}
	err := r.db.WithContext(ctx).
		Preload("Variants").
		Where("code = ?", code).
		First(&product).
		Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.Product{}, nil
		}
		return models.Product{}, err
	}
	return product, nil
}

func (r *GormRepo) GetByCategory(
	ctx context.Context,
	category string,
) ([]models.Product, error) {
	var products []models.Product
	err := r.db.WithContext(ctx).
		Joins("JOIN categories ON categories.id = products.category_id").
		Where("categories.code = ?", category).
		Preload("Variants").
		Preload("Category").
		Find(&products).
		Error
	if err != nil {
		return nil, err
	}
	return products, nil
}
