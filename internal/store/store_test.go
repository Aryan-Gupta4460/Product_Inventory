package store_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"testing/quick"

	"log/slog"

	"product_inventory/internal/domain"
	apperrors "product_inventory/internal/errors"
	"product_inventory/internal/store"
)

func newLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{}))
}

func mk(id, name string, price float64, qty int, cat string) domain.Product {
	return domain.Product{ID: id, Name: name, Price: price, Quantity: qty, Category: cat}
}

func TestList_TableDriven(t *testing.T) {
	logger := newLogger()
	ms := store.NewMemoryStore(logger)
	ctx := context.Background()

	products := []domain.Product{
		mk("1", "A", 10, 1, "X"),
		mk("2", "B", 20, 2, "X"),
		mk("3", "C", 30, 3, "Y"),
	}
	for _, p := range products {
		if err := ms.Create(ctx, p); err != nil {
			t.Fatalf("seed create: %v", err)
		}
	}

	tests := []struct {
		name   string
		filter store.ListFilter
		want   int
	}{
		{"no-filter", store.ListFilter{}, 3},
		{"category X", store.ListFilter{Category: "X"}, 2},
		{"min price", store.ListFilter{MinPrice: 15}, 2},
		{"max price", store.ListFilter{MaxPrice: 15}, 1},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got, err := ms.List(ctx, tc.filter)
			if err != nil {
				t.Fatalf("List error: %v", err)
			}
			if len(got) != tc.want {
				t.Fatalf("got %d want %d", len(got), tc.want)
			}
		})
	}
}

func TestContextCancellation(t *testing.T) {
	logger := newLogger()
	ms := store.NewMemoryStore(logger)

	// cancelled context should cause Get/Update/Delete to return context error
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := ms.Get(ctx, "nope"); err == nil {
		t.Fatalf("expected error on cancelled context")
	}
}

func TestFileStore_Persistence(t *testing.T) {
	logger := newLogger()
	tmp := t.TempDir()
	path := filepath.Join(tmp, "data.json")

	fs := store.NewFileStore(path, logger)
	ctx := context.Background()

	p := mk("f1", "FileProd", 1.1, 1, "F")
	if err := fs.Create(ctx, p); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// New instance should load file
	fs2 := store.NewFileStore(path, logger)
	if _, err := fs2.Get(ctx, "f1"); err != nil {
		t.Fatalf("expected product after reload, got %v", err)
	}

	// Invalid create
	bad := mk("b", "--bad", -1, -1, "X")
	if err := fs.Create(ctx, bad); err == nil || !errors.Is(err, apperrors.ErrInvalidProduct) {
		t.Fatalf("expected invalid product error")
	}
}

func TestProductValidation_Quick(t *testing.T) {
	f := func(p domain.Product) bool {
		err := p.Validate()
		if p.Price < 0 || p.Quantity < 0 {
			return err != nil
		}
		return err == nil
	}
	_ = quick.Check(f, nil)
}

func TestMemoryStore_BulkImport_Duplicate(t *testing.T) {
	logger := newLogger()
	ms := store.NewMemoryStore(logger)
	ctx := context.Background()

	products := []domain.Product{
		mk("d1", "D", 1, 1, "X"),
		mk("d1", "Ddup", 2, 2, "X"),
	}

	if err := ms.BulkImport(ctx, products); err == nil {
		t.Fatalf("expected duplicate product error from BulkImport")
	}
}
