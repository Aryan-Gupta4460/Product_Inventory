package cmd

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	apperrors "product_inventory/internal/errors"
	"product_inventory/internal/store"

	"github.com/spf13/cobra"
)

// NewDeleteCmd returns a Cobra command for deleting products with optional confirmation prompt.
func NewDeleteCmd(store store.ProductStore) *cobra.Command {

	var force bool

	cmd := &cobra.Command{
		Use:   "delete <product-id>",
		Short: "Delete product",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {

			id := args[0]

			if !force {
				fmt.Printf("Are you sure you want to delete %s? (y/n): ", id)
				reader := bufio.NewReader(os.Stdin)
				input, _ := reader.ReadString('\n')
				input = strings.TrimSpace(input)

				if input != "y" {
					fmt.Println("Deletion cancelled")
					return nil
				}
			}

			err := store.Delete(context.Background(), id)
			if err != nil {

				if errors.Is(err, apperrors.ErrProductNotFound) {
					fmt.Println("Product not found")
					return nil
				}

				return err
			}

			fmt.Println("Product deleted successfully")
			return nil
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "Force delete without confirmation")

	return cmd
}
