package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"

	"github.com/IhorXsh/Money-Transfer-Usecase/internal/domain"
	"github.com/IhorXsh/Money-Transfer-Usecase/internal/repository"
	"github.com/IhorXsh/Money-Transfer-Usecase/internal/server"
	"github.com/IhorXsh/Money-Transfer-Usecase/internal/telemetry"
	"github.com/IhorXsh/Money-Transfer-Usecase/internal/usecases/transfer"
)

func run() error {
	cfg, err := loadConfig(os.Args[1:])
	if err != nil {
		return err
	}

	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	shutdownTracing, err := telemetry.InitTracing(ctx, "money-transfer")
	if err != nil {
		return fmt.Errorf("init tracing: %w", err)
	}
	defer func() {
		if err := shutdownTracing(context.Background()); err != nil {
			logger.Error("tracing shutdown failed", "error", err)
		}
	}()

	accounts := map[domain.AccountID]*domain.Account{
		"a1": domain.NewAccount("a1", 100, domain.AccountStatusActive),
		"a2": domain.NewAccount("a2", 50, domain.AccountStatusActive),
		"a3": domain.NewAccount("a3", 10, domain.AccountStatusActive),
	}

	repo := repository.NewAccountRepo(accounts)
	uc := transfer.NewInteractor(repo)
	srv := server.New(logger, uc)

	errCh := make(chan error, 2)

	go func() {
		logger.Info("metrics server started", "addr", cfg.metricsAddr)
		errCh <- fmt.Errorf("metrics server: %w", http.ListenAndServe(cfg.metricsAddr, srv.MetricsHandler()))
	}()

	go func() {
		logger.Info("http server started", "addr", cfg.appAddr)
		errCh <- fmt.Errorf("http server: %w", http.ListenAndServe(cfg.appAddr, srv.Handler()))
	}()

	return <-errCh
}

type config struct {
	appAddr     string
	metricsAddr string
}

func loadConfig(args []string) (config, error) {
	fs := flag.NewFlagSet("money-transfer", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	appPort := fs.Int("app-port", 8080, "application HTTP port")
	metricsPort := fs.Int("metrics-port", 9090, "metrics HTTP port")
	if err := fs.Parse(args); err != nil {
		return config{}, fmt.Errorf("parse flags: %w", err)
	}
	if *appPort <= 0 || *appPort > 65535 {
		return config{}, fmt.Errorf("invalid app-port: %d", *appPort)
	}
	if *metricsPort <= 0 || *metricsPort > 65535 {
		return config{}, fmt.Errorf("invalid metrics-port: %d", *metricsPort)
	}

	return config{
		appAddr:     fmt.Sprintf(":%d", *appPort),
		metricsAddr: fmt.Sprintf(":%d", *metricsPort),
	}, nil
}
