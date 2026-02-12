package cmd

import (
	"product_inventory/internal/store"

	"github.com/spf13/cobra"
)

// NewRootCmd returns the root Cobra command that contains all inventory CLI subcommands.
func NewRootCmd(store store.ProductStore) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "inventory-cli",
		Short:         "A product inventory management system",
		SilenceUsage:  true,
		SilenceErrors: true,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	cmd.AddCommand(NewCreateCmd(store))
	cmd.AddCommand(NewGetCmd(store))
	cmd.AddCommand(NewDeleteCmd(store))
	cmd.AddCommand(NewExportCmd(store))
	cmd.AddCommand(NewImportCmd(store))
	cmd.AddCommand(NewListCmd(store))
	cmd.AddCommand(NewUpdateCmd(store))

	return cmd
}
