package category

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

func (r *GormRepo) ListAll(ctx context.Context) ([]models.Category, error) {
	var categories []models.Category
	err := r.db.WithContext(ctx).Find(&categories).Error
	if err != nil {
		return nil, err
	}
	return categories, nil
}

func (r *GormRepo) Create(ctx context.Context, newCategory models.Category) error {
	err := r.db.WithContext(ctx).Create(newCategory).Error
	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return errors.New("category code already exists")
		}
	}

	return err
}

func (r *GormRepo) GetByID(ctx context.Context, id int) (string, error) {
	var category models.Category
	err := r.db.WithContext(ctx).First(&category, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", nil
		}
		return "", err
	}
	return category.Code, nil
}

func (r *GormRepo) GetByCode(ctx context.Context, code string) (string, error) {
	var category models.Category
	err := r.db.WithContext(ctx).
		Where("code = ?", code).
		First(&category).
		Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", nil
		}
		return "", err
	}
	return category.Name, nil
}

func (r *GormRepo) GetProducts(ctx context.Context, code string) ([]models.Product, error) {
	var category models.Category
	err := r.db.WithContext(ctx).
		Preload("Products.Variants").
		Where("code = ?", code).
		First(&category).
		Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return category.Products, nil
}
