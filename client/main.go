package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"

	"github.com/sysarmor/guard/client/cmd"
	"golang.org/x/sync/errgroup"
)

// Main is the entry point of the application
func Main(ctx context.Context) error {
	err := cmd.New().ExecuteContext(ctx)
	if err != nil {
		return fmt.Errorf("failed to execute command: %w", err)
	}
	return nil
}

func main() {
	ctx := context.Background()
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	eg, egCtx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		return Main(egCtx)
	})

	if err := eg.Wait(); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}
