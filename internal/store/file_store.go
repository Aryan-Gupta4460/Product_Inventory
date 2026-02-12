package store

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"product_inventory/internal/domain"
	apperrors "product_inventory/internal/errors"
	"sync"
	"time"
)

type FileStore struct {
	mu       sync.RWMutex
	filePath string
	products map[string]domain.Product
	logger   *slog.Logger
}

func NewFileStore(path string, logger *slog.Logger) *FileStore {
	fs := &FileStore{
		filePath: path,
		products: make(map[string]domain.Product),
		logger:   logger,
	}

	// Load existing data if file exists
	fs.loadFromFile()

	return fs
}

func (s *FileStore) loadFromFile() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.logger.Debug("loading products from file", "path", s.filePath)

	data, err := os.ReadFile(s.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			s.logger.Info("data file does not exist, starting with empty store",
				"path", s.filePath,
			)
			return nil
		}

		s.logger.Error("failed to read data file",
			"path", s.filePath,
			"error", err,
		)
		return err
	}

	if len(data) == 0 {
		s.logger.Warn("data file is empty", "path", s.filePath)
		return nil
	}

	var products []domain.Product
	if err := json.Unmarshal(data, &products); err != nil {
		s.logger.Error("failed to unmarshal products",
			"path", s.filePath,
			"error", err,
		)
		return err
	}

	s.products = make(map[string]domain.Product)

	for _, p := range products {
		s.products[p.ID] = p
	}

	s.logger.Info("products loaded successfully",
		"count", len(products),
	)
	return nil
}

func (s *FileStore) saveToFile() error {
	s.logger.Debug("saving products to file", "path", s.filePath)

	// Convert map to slice
	var products []domain.Product
	for _, p := range s.products {
		products = append(products, p)
	}

	data, err := json.MarshalIndent(products, "", "  ")
	if err != nil {
		s.logger.Error("failed to marshal products",
			"error", err,
		)
		return err
	}

	tempFile := s.filePath + ".tmp"
	if err := os.WriteFile(tempFile, data, 0644); err != nil {
		s.logger.Error("failed to write temp file",
			"path", tempFile,
			"error", err,
		)
		return err
	}
	if err := os.Rename(tempFile, s.filePath); err != nil {
		s.logger.Error("failed to replace data file",
			"path", s.filePath,
			"error", err,
		)
		return err
	}
	s.logger.Info("products saved successfully",
		"path", s.filePath,
		"count", len(products),
	)

	return nil
}
func (s *FileStore) Create(ctx context.Context, p domain.Product) error {

	if err := p.Validate(); err != nil {
		s.logger.Warn("invalid product",
			"id", p.ID,
			"error", err,
		)
		return &apperrors.InvalidProductError{Reason: err.Error()}
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.products[p.ID]; exists {
		s.logger.Warn("duplicate product",
			"id", p.ID,
		)
		return &apperrors.DuplicateProductError{ID: p.ID}
	}

	s.products[p.ID] = p

	if err := s.saveToFile(); err != nil {
		s.logger.Error("failed to persist product",
			"id", p.ID,
			"error", err,
		)
		return err
	}
	s.logger.Info("product created",
		"id", p.ID,
		"name", p.Name,
	)
	return nil
}

func (s *FileStore) Get(ctx context.Context, id string) (domain.Product, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	s.logger.Debug("fetching product", "id", id)

	p, exists := s.products[id]
	if !exists {
		err := &apperrors.ProductNotFoundError{ID: id}
		s.logger.Warn("product not found", "id", id)
		return domain.Product{}, err
	}

	s.logger.Debug("product fetched successfully", "id", id)
	return p, nil
}

func (s *FileStore) Update(ctx context.Context, id string, p domain.Product) error {
	s.logger.Info("updating product", "id", id)

	if err := p.Validate(); err != nil {
		s.logger.Warn("invalid product data", "id", id, "error", err)
		return &apperrors.InvalidProductError{Reason: err.Error()}
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.products[id]; !exists {
		err := &apperrors.ProductNotFoundError{ID: id}
		s.logger.Warn("update failed - product not found", "id", id)
		return err
	}

	p.ID = id
	s.products[id] = p

	if err := s.saveToFile(); err != nil {
		s.logger.Error("failed to save product after update", "id", id, "error", err)
		return err
	}

	s.logger.Info("product updated successfully", "id", id)
	return nil
}

func (s *FileStore) Delete(ctx context.Context, id string) error {
	s.logger.Debug("deleting product", "id", id)

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.products[id]; !exists {
		err := &apperrors.ProductNotFoundError{ID: id}
		s.logger.Warn("delete failed - product not found", "id", id)
		return err
	}

	delete(s.products, id)

	if err := s.saveToFile(); err != nil {
		s.logger.Error("failed to save file after delete", "id", id, "error", err)
		return err
	}

	s.logger.Info("product deleted successfully", "id", id)
	return nil
}
func (s *FileStore) List(ctx context.Context, filter ListFilter) ([]domain.Product, error) {

	s.logger.DebugContext(ctx, "listing products",
		"category", filter.Category,
		"min_price", filter.MinPrice,
		"max_price", filter.MaxPrice,
	)
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []domain.Product

	for _, p := range s.products {

		if filter.Category != "" && p.Category != filter.Category {
			continue
		}

		if filter.MinPrice > 0 && p.Price < filter.MinPrice {
			continue
		}

		if filter.MaxPrice > 0 && p.Price > filter.MaxPrice {
			continue
		}

		result = append(result, p)
	}

	s.logger.InfoContext(ctx, "products listed successfully",
		"count", len(result),
	)

	return result, nil
}

func (s *FileStore) BulkImport(ctx context.Context, products []domain.Product) error {

	s.logger.InfoContext(ctx, "starting bulk import",
		"count", len(products),
	)

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, p := range products {

		if err := p.Validate(); err != nil {
			s.logger.WarnContext(ctx, "invalid product in bulk import",
				"id", p.ID,
				"error", err,
			)
			return &apperrors.InvalidProductError{Reason: err.Error()}
		}

		if _, exists := s.products[p.ID]; exists {
			s.logger.WarnContext(ctx, "overwriting existing product during bulk import",
				"id", p.ID,
			)
		}

		s.products[p.ID] = p
	}

	if err := s.saveToFile(); err != nil {
		s.logger.ErrorContext(ctx, "failed to save file after bulk import",
			"error", err,
		)
		return err
	}

	s.logger.InfoContext(ctx, "bulk import completed successfully",
		"total_products", len(s.products),
	)

	return nil
}
func (s *FileStore) BulkExport(ctx context.Context, filePath string, filter ListFilter) error {
	start := time.Now()
	s.logger.InfoContext(ctx, "starting bulk export",
		"file", filePath,
	)

	select {
	case <-ctx.Done():
		s.logger.WarnContext(ctx, "bulk export cancelled",
			"error", ctx.Err(),
		)
		return ctx.Err()
	default:
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []domain.Product

	// Apply filtering
	for _, p := range s.products {

		if filter.Category != "" && p.Category != filter.Category {
			continue
		}
		if filter.MinPrice > 0 && p.Price < filter.MinPrice {
			continue
		}
		if filter.MaxPrice > 0 && p.Price > filter.MaxPrice {
			continue
		}

		result = append(result, p)
	}

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to marshal products",
			"error", err,
		)
		return err
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		s.logger.ErrorContext(ctx, "failed to write export file",
			"file", filePath,
			"error", err,
		)
		return err
	}

	s.logger.InfoContext(ctx, "bulk export completed",
		"file", filePath,
		"exported_count", len(result),
		"duration", time.Since(start).String(),
	)

	return nil
}
