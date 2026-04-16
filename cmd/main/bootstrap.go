package main

import (
	"context"
	"fmt"
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

	appAddr := ":8080"
	metricsAddr := ":9090"

	errCh := make(chan error, 2)

	go func() {
		logger.Info("metrics server started", "addr", metricsAddr)
		errCh <- fmt.Errorf("metrics server: %w", http.ListenAndServe(metricsAddr, srv.MetricsHandler()))
	}()

	go func() {
		logger.Info("http server started", "addr", appAddr)
		errCh <- fmt.Errorf("http server: %w", http.ListenAndServe(appAddr, srv.Handler()))
	}()

	return <-errCh
}
