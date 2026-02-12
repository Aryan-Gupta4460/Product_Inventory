package cmd

import (
	"context"
	"encoding/json"
	"os"

	"product_inventory/internal/store"

	"github.com/spf13/cobra"
)

// NewExportCmd returns a Cobra command for exporting products to a JSON file with optional filtering.
func NewExportCmd(ps store.ProductStore) *cobra.Command {

	var file string
	var category string

	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export products to JSON file",
		RunE: func(cmd *cobra.Command, args []string) error {

			filter := store.ListFilter{
				Category: category,
			}

			products, err := ps.List(context.Background(), filter)
			if err != nil {
				return err
			}

			data, err := json.MarshalIndent(products, "", "  ")
			if err != nil {
				return err
			}

			return os.WriteFile(file, data, 0644)
		},
	}

	cmd.Flags().StringVar(&file, "file", "", "Output file path")
	cmd.Flags().StringVar(&category, "category", "", "Filter by category")
	cmd.MarkFlagRequired("file")

	return cmd
}
