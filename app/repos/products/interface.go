package products

import (
	"context"

	"github.com/mytheresa/go-hiring-challenge/models"
	"github.com/shopspring/decimal"
)

type Repository interface {
	ListAll(ctx context.Context) ([]models.Product, error)
	List(ctx context.Context, filters SearchFilters) ([]models.Product, error)
	GetByID(ctx context.Context, id string) (models.Product, error)
	GetByCode(ctx context.Context, code string) (models.Product, error)
	GetByCategory(ctx context.Context, category string) ([]models.Product, error)
}

type SearchFilters struct {
	Offset        int
	Limit         int
	Category      string
	PriceLessThan *decimal.Decimal
}
