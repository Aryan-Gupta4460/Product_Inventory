package store

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"sync"

	"product_inventory/internal/domain"
	apperrors "product_inventory/internal/errors"
)

type MemoryStore struct {
	mu       sync.RWMutex
	products map[string]domain.Product
	logger   *slog.Logger
}

func NewMemoryStore(logger *slog.Logger) *MemoryStore {
	return &MemoryStore{
		products: make(map[string]domain.Product),
		logger:   logger,
	}
}

func (s *MemoryStore) Create(ctx context.Context, p domain.Product) error {
	if err := p.Validate(); err != nil {
		s.logger.WarnContext(ctx, "invalid product",
			"id", p.ID,
			"error", err,
		)
		return &apperrors.InvalidProductError{Reason: err.Error()}
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.products[p.ID]; exists {
		s.logger.WarnContext(ctx, "duplicate product",
			"id", p.ID,
		)
		return &apperrors.DuplicateProductError{ID: p.ID}
	}

	s.products[p.ID] = p

	s.logger.InfoContext(ctx, "product created",
		"id", p.ID,
	)

	return nil
}

func (s *MemoryStore) Get(ctx context.Context, id string) (domain.Product, error) {
	select {
	case <-ctx.Done():
		s.logger.WarnContext(ctx, "get cancelled",
			"id", id,
		)
		return domain.Product{}, ctx.Err()
	default:
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	product, exists := s.products[id]
	if !exists {
		s.logger.WarnContext(ctx, "product not found",
			"id", id,
		)
		return domain.Product{}, &apperrors.ProductNotFoundError{ID: id}
	}

	s.logger.DebugContext(ctx, "product retrieved",
		"id", id,
	)

	return product, nil
}

func (s *MemoryStore) Update(ctx context.Context, id string, updated domain.Product) error {
	select {
	case <-ctx.Done():
		s.logger.WarnContext(ctx, "update cancelled",
			"id", id,
		)
		return ctx.Err()
	default:
	}

	if err := updated.Validate(); err != nil {
		s.logger.WarnContext(ctx, "invalid product update",
			"id", id,
			"error", err,
		)
		return &apperrors.InvalidProductError{Reason: err.Error()}
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.products[id]; !exists {
		s.logger.WarnContext(ctx, "product not found for update",
			"id", id,
		)
		return &apperrors.ProductNotFoundError{ID: id}
	}

	updated.ID = id
	s.products[id] = updated

	s.logger.InfoContext(ctx, "product updated",
		"id", id,
	)

	return nil
}

func (s *MemoryStore) Delete(ctx context.Context, id string) error {
	select {
	case <-ctx.Done():
		s.logger.WarnContext(ctx, "delete cancelled",
			"id", id,
		)
		return ctx.Err()
	default:
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.products[id]; !exists {
		s.logger.WarnContext(ctx, "product not found for delete",
			"id", id,
		)
		return &apperrors.ProductNotFoundError{ID: id}
	}

	delete(s.products, id)

	s.logger.InfoContext(ctx, "product deleted",
		"id", id,
	)

	return nil
}

func (s *MemoryStore) List(ctx context.Context, filter ListFilter) ([]domain.Product, error) {
	select {
	case <-ctx.Done():
		s.logger.WarnContext(ctx, "list cancelled")
		return nil, ctx.Err()
	default:
	}

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

	s.logger.InfoContext(ctx, "products listed",
		"count", len(result),
		"category", filter.Category,
		"min_price", filter.MinPrice,
		"max_price", filter.MaxPrice,
	)

	return result, nil
}

func (s *MemoryStore) BulkImport(ctx context.Context, products []domain.Product) error {
	const workerCount = 10

	s.logger.InfoContext(ctx, "starting bulk import",
		"total_products", len(products),
		"workers", workerCount,
	)

	jobs := make(chan domain.Product)
	errCh := make(chan error, len(products))

	var wg sync.WaitGroup

	// Worker pool
	for i := 0; i < workerCount; i++ {
		wg.Add(1)

		go func(workerID int) {
			defer wg.Done()

			for p := range jobs {

				select {
				case <-ctx.Done():
					s.logger.WarnContext(ctx, "bulk import cancelled",
						"worker_id", workerID,
					)
					return
				default:
				}

				if err := s.Create(ctx, p); err != nil {
					s.logger.WarnContext(ctx, "failed to import product",
						"id", p.ID,
						"error", err,
					)
					errCh <- err
				}
			}
		}(i)
	}

	// Send jobs
	go func() {
		defer close(jobs)

		for _, p := range products {
			select {
			case <-ctx.Done():
				return
			case jobs <- p:
			}
		}
	}()

	// Close error channel after workers finish
	go func() {
		wg.Wait()
		close(errCh)
	}()

	// Aggregate errors
	var firstErr error
	errorCount := 0

	for err := range errCh {
		if err != nil {
			errorCount++
			if firstErr == nil {
				firstErr = err
			}
		}
	}

	if ctx.Err() != nil {
		s.logger.WarnContext(ctx, "bulk import terminated due to context cancellation")
		return ctx.Err()
	}

	if errorCount > 0 {
		s.logger.ErrorContext(ctx, "bulk import completed with errors",
			"error_count", errorCount,
		)
		return firstErr
	}

	s.logger.InfoContext(ctx, "bulk import completed successfully",
		"imported_count", len(products),
	)

	return nil
}
func (s *MemoryStore) BulkExport(ctx context.Context, filePath string, filter ListFilter) error {
	const workerCount = 10

	s.logger.InfoContext(ctx, "starting bulk export",
		"file", filePath,
		"workers", workerCount,
		"category", filter.Category,
		"min_price", filter.MinPrice,
		"max_price", filter.MaxPrice,
	)

	// Take snapshot
	s.mu.RLock()
	snapshot := make([]domain.Product, 0, len(s.products))
	for _, p := range s.products {
		snapshot = append(snapshot, p)
	}
	s.mu.RUnlock()

	jobs := make(chan domain.Product)
	results := make(chan domain.Product)
	var wg sync.WaitGroup

	// Worker pool for filtering
	for i := 0; i < workerCount; i++ {
		wg.Add(1)

		go func(workerID int) {
			defer wg.Done()

			for p := range jobs {

				select {
				case <-ctx.Done():
					s.logger.WarnContext(ctx, "bulk export cancelled",
						"worker_id", workerID,
					)
					return
				default:
				}

				if filter.Category != "" && p.Category != filter.Category {
					continue
				}
				if filter.MinPrice > 0 && p.Price < filter.MinPrice {
					continue
				}
				if filter.MaxPrice > 0 && p.Price > filter.MaxPrice {
					continue
				}

				results <- p
			}
		}(i)
	}

	// Send jobs
	go func() {
		defer close(jobs)

		for _, p := range snapshot {
			select {
			case <-ctx.Done():
				return
			case jobs <- p:
			}
		}
	}()

	// Close results after workers finish
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	var filtered []domain.Product
	for p := range results {
		filtered = append(filtered, p)
	}

	if ctx.Err() != nil {
		s.logger.WarnContext(ctx, "bulk export terminated due to context cancellation")
		return ctx.Err()
	}

	// Marshal JSON
	data, err := json.MarshalIndent(filtered, "", "  ")
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to marshal export data",
			"error", err,
		)
		return err
	}

	// Write file
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		s.logger.ErrorContext(ctx, "failed to write export file",
			"file", filePath,
			"error", err,
		)
		return err
	}

	s.logger.InfoContext(ctx, "bulk export completed successfully",
		"file", filePath,
		"exported_count", len(filtered),
	)

	return nil
}
