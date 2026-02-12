package store

import (
	"context"
	"product_inventory/internal/domain"
)

type ListFilter struct {
	Category string
	MinPrice float64
	MaxPrice float64
}

type ProductStore interface {
	Create(ctx context.Context, product domain.Product) error
	Get(ctx context.Context, id string) (domain.Product, error)
	Update(ctx context.Context, id string, product domain.Product) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, filter ListFilter) ([]domain.Product, error)
	BulkImport(ctx context.Context, products []domain.Product) error
	BulkExport(ctx context.Context, filePath string, filter ListFilter) error
}
