package cmd

import (
	"context"
	"product_inventory/internal/domain"
	"product_inventory/internal/store"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

// NewCreateCmd returns a Cobra command for creating new products with auto-generated UUID.
func NewCreateCmd(store store.ProductStore) *cobra.Command {
	var name string
	var price float64
	var quantity int
	var category string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create product",
		RunE: func(cmd *cobra.Command, args []string) error {

			product := domain.Product{
				ID:       uuid.New().String(),
				Name:     name,
				Price:    price,
				Quantity: quantity,
				Category: category,
			}

			return store.Create(context.Background(), product)
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Product name")
	cmd.Flags().Float64Var(&price, "price", 0, "Product price")
	cmd.Flags().IntVar(&quantity, "quantity", 0, "Product quantity")
	cmd.Flags().StringVar(&category, "category", "", "Product category")

	return cmd
}
