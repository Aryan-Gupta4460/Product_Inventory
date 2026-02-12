package cmd

import (
	"context"
	"errors"
	"fmt"
	apperrors "product_inventory/internal/errors"

	"product_inventory/internal/store"

	"github.com/spf13/cobra"
)

// NewUpdateCmd returns a Cobra command for updating product details with partial update support.
func NewUpdateCmd(store store.ProductStore) *cobra.Command {

	var name string
	var price float64
	var quantity int
	var category string

	cmd := &cobra.Command{
		Use:   "update <product-id>",
		Short: "Update product details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {

			id := args[0]

			ctx := context.Background()

			existing, err := store.Get(ctx, id)
			if err != nil {

				if errors.Is(err, apperrors.ErrProductNotFound) {
					fmt.Println("Product not found")
					return nil
				}

				return err
			}

			if cmd.Flags().Changed("name") {
				existing.Name = name
			}

			if cmd.Flags().Changed("price") {
				existing.Price = price
			}

			if cmd.Flags().Changed("quantity") {
				existing.Quantity = quantity
			}

			if cmd.Flags().Changed("category") {
				existing.Category = category
			}

			err = store.Update(ctx, id, existing)
			if err != nil {
				return err
			}

			fmt.Println("Product updated successfully")
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "New product name")
	cmd.Flags().Float64Var(&price, "price", 0, "New product price")
	cmd.Flags().IntVar(&quantity, "quantity", 0, "New product quantity")
	cmd.Flags().StringVar(&category, "category", "", "New product category")

	return cmd
}
