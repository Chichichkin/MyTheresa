package category

import (
	"context"

	"github.com/mytheresa/go-hiring-challenge/models"
)

type Repository interface {
	ListAll(ctx context.Context) ([]models.Category, error)
	Create(ctx context.Context, newCategory models.Category) error
	GetByID(ctx context.Context, id int) (string, error)
	GetByCode(ctx context.Context, code string) (string, error)
	GetProducts(ctx context.Context, code string) ([]models.Product, error)
}
