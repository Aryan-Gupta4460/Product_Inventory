package store

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"log/slog"
	"product_inventory/internal/domain"
)

func newLoggerStd() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{}))
}

func TestFileStore_LoadMalformedJSON(t *testing.T) {
	tmp := t.TempDir()
	p := filepath.Join(tmp, "bad.json")
	_ = os.WriteFile(p, []byte("notjson"), 0644)

	fs := &FileStore{filePath: p, products: make(map[string]domain.Product), logger: newLoggerStd()}
	if err := fs.loadFromFile(); err == nil {
		t.Fatalf("expected error when loading malformed JSON")
	}
}

func TestFileStore_BulkExport_Cancelled(t *testing.T) {
	tmp := t.TempDir()
	p := filepath.Join(tmp, "data.json")

	fs := &FileStore{filePath: p, products: map[string]domain.Product{"a": {ID: "a", Name: "A"}}, logger: newLoggerStd()}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if err := fs.BulkExport(ctx, filepath.Join(tmp, "out.json"), ListFilter{}); err == nil {
		t.Fatalf("expected context cancelled error")
	}
}

func TestFileStore_BulkImport_Overwrite(t *testing.T) {
	tmp := t.TempDir()
	p := filepath.Join(tmp, "data.json")

	fs := &FileStore{filePath: p, products: map[string]domain.Product{"x": {ID: "x", Name: "Old", Price: 1}}, logger: newLoggerStd()}

	products := []domain.Product{{ID: "x", Name: "New", Price: 2}}
	if err := fs.BulkImport(context.Background(), products); err != nil {
		t.Fatalf("bulk import failed: %v", err)
	}
	if got := fs.products["x"].Name; got != "New" {
		t.Fatalf("expected overwrite, got %s", got)
	}
}

func TestMemoryStore_BulkImport_ContextCancel(t *testing.T) {
	ms := NewMemoryStore(newLoggerStd())
	// create many products
	var products []domain.Product
	for i := 0; i < 100; i++ {
		products = append(products, domain.Product{ID: time.Now().String() + strconv.Itoa(i), Name: "P"})
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	// should return quickly (may return nil depending on implementation); assert no panic
	_ = ms.BulkImport(ctx, products)
}

func TestMemoryStore_Errors(t *testing.T) {
	ms := NewMemoryStore(newLoggerStd())
	ctx := context.Background()

	// Get not found
	if _, err := ms.Get(ctx, "nofile"); err == nil {
		t.Fatalf("expected not found error")
	}

	// Update not found
	if err := ms.Update(ctx, "nofile", domain.Product{ID: "nofile", Name: "N"}); err == nil {
		t.Fatalf("expected update not found error")
	}

	// Delete not found
	if err := ms.Delete(ctx, "nofile"); err == nil {
		t.Fatalf("expected delete not found error")
	}
}

func TestFileStore_ErrorsAndExport(t *testing.T) {
	tmp := t.TempDir()
	p := filepath.Join(tmp, "data.json")
	fs := NewFileStore(p, newLoggerStd())
	ctx := context.Background()

	// Update not found
	if err := fs.Update(ctx, "missing", domain.Product{ID: "missing", Name: "X"}); err == nil {
		t.Fatalf("expected update not found error")
	}

	// Delete not found
	if err := fs.Delete(ctx, "missing"); err == nil {
		t.Fatalf("expected delete not found error")
	}

	// Create and export
	_ = fs.Create(ctx, domain.Product{ID: "e1", Name: "E1", Price: 1})
	out := filepath.Join(tmp, "out.json")
	if err := fs.BulkExport(ctx, out, ListFilter{}); err != nil {
		t.Fatalf("export failed: %v", err)
	}
	if _, err := os.Stat(out); err != nil {
		t.Fatalf("expected out file exists: %v", err)
	}
}

func TestStore_ComprehensiveFlow(t *testing.T) {
	// Memory store flow
	ms := NewMemoryStore(newLoggerStd())
	ctx := context.Background()
	for i := 0; i < 20; i++ {
		id := "m" + strconv.Itoa(i)
		_ = ms.Create(ctx, domain.Product{ID: id, Name: "P" + strconv.Itoa(i), Price: float64(i), Quantity: i, Category: "C"})
	}
	if _, err := ms.List(ctx, ListFilter{Category: "C"}); err != nil {
		t.Fatalf("memory list failed: %v", err)
	}
	_ = ms.Update(ctx, "m1", domain.Product{Name: "Updated", Price: 9.9, Quantity: 1})
	_ = ms.Delete(ctx, "m2")

	// File store flow
	tmp := t.TempDir()
	p := filepath.Join(tmp, "fdata.json")
	fs := NewFileStore(p, newLoggerStd())
	_ = fs.Create(ctx, domain.Product{ID: "f1", Name: "F1", Price: 1})
	_ = fs.Update(ctx, "f1", domain.Product{ID: "f1", Name: "F1v2", Price: 2})
	if _, err := fs.List(ctx, ListFilter{}); err != nil {
		t.Fatalf("file list failed: %v", err)
	}
	_ = fs.Delete(ctx, "f1")

	// BulkImport many
	var bulk []domain.Product
	for i := 0; i < 50; i++ {
		bulk = append(bulk, domain.Product{ID: "b" + strconv.Itoa(i), Name: "B", Price: float64(i)})
	}
	_ = ms.BulkImport(ctx, bulk)
	_ = fs.BulkImport(ctx, bulk)

	// BulkExport to files
	_ = ms.BulkExport(ctx, filepath.Join(tmp, "mout.json"), ListFilter{})
	_ = fs.BulkExport(ctx, filepath.Join(tmp, "fout.json"), ListFilter{})
}
