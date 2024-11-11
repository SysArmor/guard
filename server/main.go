package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"

	"github.com/sysarmor/guard/server/internal/controller"
	"github.com/sysarmor/guard/server/internal/repo/postgres"
	"github.com/sysarmor/guard/server/internal/service"
	_ "github.com/sysarmor/guard/server/pkg/log"
	"github.com/sysarmor/guard/server/route"
	"golang.org/x/sync/errgroup"
)

// Main is the entry of the server
func Main(ctx context.Context) (close func() error, err error) {
	var configPath string
	flag.StringVar(&configPath, "config", "config.yaml", "config file path")
	flag.Parse()

	cfg, err := loadConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("get config: %w", err)
	}

	repo, err := postgres.New(ctx, &cfg.Postgres)
	if err != nil {
		return nil, fmt.Errorf("new postgres: %w", err)
	}

	svc, err := service.New(cfg.Services, repo)
	if err != nil {
		return nil, fmt.Errorf("new service: %w", err)
	}

	r := route.New(controller.New(svc))
	close = func() error {
		return r.Close()
	}

	go func() {
		if err := r.Run(cfg.Addr); err != nil {
			slog.ErrorContext(ctx, "run", "error", err)
		}
	}()

	return close, nil
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	close, err := Main(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "main", "error", err)
		return // nolint
	}

	g, gCtx := errgroup.WithContext(ctx)
	g.Go(func() error {
		<-gCtx.Done()
		if close != nil {
			return close()
		}
		return nil
	})
	if err := g.Wait(); err != nil {
		slog.ErrorContext(ctx, "exit", "error", err)
	}
	slog.InfoContext(ctx, "exit")
}
