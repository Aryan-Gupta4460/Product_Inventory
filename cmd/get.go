package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	apperrors "product_inventory/internal/errors"

	"product_inventory/internal/store"

	"github.com/spf13/cobra"
)

// NewGetCmd returns a Cobra command for retrieving a product by ID.
func NewGetCmd(store store.ProductStore) *cobra.Command {

	cmd := &cobra.Command{
		Use:   "get <product-id>",
		Short: "Get product by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {

			id := args[0]

			product, err := store.Get(context.Background(), id)
			if err != nil {

				if errors.Is(err, apperrors.ErrProductNotFound) {
					fmt.Println("Product not found")
					return nil
				}

				return err
			}

			data, _ := json.MarshalIndent(product, "", "  ")
			fmt.Println(string(data))

			return nil
		},
	}

	return cmd
}
