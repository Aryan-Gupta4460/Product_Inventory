package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"product_inventory/cmd"
	"product_inventory/internal/config"
	"product_inventory/internal/store"
	"product_inventory/pkg/logger"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		fmt.Println("failed to load config:", err)
		os.Exit(1)
	}

	log := logger.New()
	store := store.NewStore(cfg, log)

	rootCmd := cmd.NewRootCmd(store)

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		log.Error("command failed", "error", err)
		os.Exit(1)
	}
}
