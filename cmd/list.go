package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"

	"product_inventory/internal/store"

	"github.com/spf13/cobra"
)

// NewListCmd returns a Cobra command for listing products with optional filtering and sorting.
func NewListCmd(ps store.ProductStore) *cobra.Command {

	var category string
	var minPrice float64
	var maxPrice float64
	var sortBy string
	var output string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List products",
		RunE: func(cmd *cobra.Command, args []string) error {

			filter := store.ListFilter{
				Category: category,
				MinPrice: minPrice,
				MaxPrice: maxPrice,
			}

			products, err := ps.List(context.Background(), filter)
			if err != nil {
				return err
			}

			// Sorting
			switch sortBy {
			case "name":
				sort.Slice(products, func(i, j int) bool {
					return products[i].Name < products[j].Name
				})
			case "price":
				sort.Slice(products, func(i, j int) bool {
					return products[i].Price < products[j].Price
				})
			case "quantity":
				sort.Slice(products, func(i, j int) bool {
					return products[i].Quantity < products[j].Quantity
				})
			}

			if output == "json" {
				data, _ := json.MarshalIndent(products, "", "  ")
				fmt.Println(string(data))
				return nil
			}

			for _, p := range products {
				fmt.Printf("ID: %s | Name: %s | Price: %.2f | Qty: %d | Category: %s\n",
					p.ID, p.Name, p.Price, p.Quantity, p.Category)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&category, "category", "", "Filter by category")
	cmd.Flags().Float64Var(&minPrice, "min-price", 0, "Minimum price")
	cmd.Flags().Float64Var(&maxPrice, "max-price", 0, "Maximum price")
	cmd.Flags().StringVar(&sortBy, "sort", "", "Sort by name|price|quantity")
	cmd.Flags().StringVar(&output, "output", "table", "Output format (table|json)")

	return cmd
}
