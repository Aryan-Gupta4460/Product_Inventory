package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"product_inventory/internal/domain"
	"product_inventory/internal/store"

	"github.com/spf13/cobra"
)

// NewImportCmd returns a Cobra command for bulk importing products from a JSON file using a worker pool.
func NewImportCmd(store store.ProductStore) *cobra.Command {

	var file string

	cmd := &cobra.Command{
		Use:   "import",
		Short: "Import products from JSON file",
		RunE: func(cmd *cobra.Command, args []string) error {

			data, err := os.ReadFile(file)
			if err != nil {
				return err
			}

			var products []domain.Product
			if err := json.Unmarshal(data, &products); err != nil {
				return fmt.Errorf("invalid JSON format: %w", err)
			}

			err = store.BulkImport(context.Background(), products)
			if err != nil {
				return err
			}

			fmt.Printf("Successfully imported %d products\n", len(products))
			return nil
		},
	}

	cmd.Flags().StringVar(&file, "file", "", "JSON file path")
	cmd.MarkFlagRequired("file")

	return cmd
}
