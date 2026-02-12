package store

import (
	"log/slog"
	"product_inventory/internal/config"
)

func NewStore(cfg config.Config, logger *slog.Logger) ProductStore {
	switch cfg.StorageBackend {
	case "file":
		return NewFileStore(cfg.FilePath, logger)
	default:
		return NewMemoryStore(logger)
	}
}
