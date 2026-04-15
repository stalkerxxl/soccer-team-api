package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/stalkerxxl/soccer-team-api/internal/api"
	"github.com/stalkerxxl/soccer-team-api/internal/api/middleware"
	"github.com/stalkerxxl/soccer-team-api/internal/auth"
	"github.com/stalkerxxl/soccer-team-api/internal/config"
	"github.com/stalkerxxl/soccer-team-api/internal/db"
	"github.com/stalkerxxl/soccer-team-api/internal/i18n"
	"github.com/stalkerxxl/soccer-team-api/internal/repository"
	"github.com/stalkerxxl/soccer-team-api/internal/service"
)

func main() {
	cfg := config.MustLoad()
	logger := newLogger(cfg.LogLevel)
	dbConnectCtx, dbConnectCancel := context.WithTimeout(context.Background(), cfg.DB.ConnectTimeout)
	defer dbConnectCancel()

	database, err := db.NewPostgres(dbConnectCtx, cfg.DB)
	if err != nil {
		logger.Error("database connection failed", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := database.Close(); err != nil {
			logger.Error("database close failed", "error", err)
		}
	}()

	logger.Info("database connected", "host", cfg.DB.Host, "port", cfg.DB.Port, "name", cfg.DB.Name)

	userRepo := repository.NewUserRepository(database)
	teamRepo := repository.NewTeamRepository(database)
	playerRepo := repository.NewPlayerRepository(database)
	marketListingRepo := repository.NewMarketListingRepository(database)

	txManager := db.NewTxManager(database)
	authTxManager, err := service.NewAuthTransactionManager(txManager)
	if err != nil {
		logger.Error("auth transaction manager init failed", "error", err)
		os.Exit(1)
	}

	marketTxManager, err := service.NewMarketTransactionManager(txManager)
	if err != nil {
		logger.Error("market transaction manager init failed", "error", err)
		os.Exit(1)
	}

	tokenProvider, err := auth.NewJWTTokenProvider(cfg.Auth.AccessSecret, cfg.Auth.AccessTTL)
	if err != nil {
		logger.Error("token provider init failed", "error", err)
		os.Exit(1)
	}

	localizer, err := i18n.NewLocalizer()
	if err != nil {
		logger.Error("localizer init failed", "error", err)
		os.Exit(1)
	}

	authService, err := service.NewAuthService(
		userRepo,
		tokenProvider,
		authTxManager,
		cfg.Auth.PasswordMinLen,
		cfg.Auth.PasswordMaxLen,
	)
	if err != nil {
		logger.Error("auth service init failed", "error", err)
		os.Exit(1)
	}

	teamService, err := service.NewTeamService(teamRepo)
	if err != nil {
		logger.Error("team service init failed", "error", err)
		os.Exit(1)
	}

	playerService, err := service.NewPlayerService(
		playerRepo,
		teamRepo,
	)
	if err != nil {
		logger.Error("player service init failed", "error", err)
		os.Exit(1)
	}

	marketService, err := service.NewMarketListingService(
		marketListingRepo,
		playerRepo,
		marketTxManager,
	)
	if err != nil {
		logger.Error("market listing service init failed", "error", err)
		os.Exit(1)
	}

	authHandler := api.NewAuthHandler(authService, localizer)
	meHandler := api.NewMeHandler(teamService, playerService, localizer)
	marketHandler := api.NewMarketHandler(teamService, marketService, localizer)
	authMiddleware := middleware.Auth(tokenProvider, localizer)

	server := &http.Server{
		Addr:         cfg.HTTP.Address(),
		Handler:      api.NewRouter(authHandler, meHandler, marketHandler, authMiddleware),
		ReadTimeout:  cfg.HTTP.ReadTimeout,
		WriteTimeout: cfg.HTTP.WriteTimeout,
		IdleTimeout:  cfg.HTTP.IdleTimeout,
	}

	logger.Info("starting api", "env", cfg.AppEnv, "addr", server.Addr)

	serverErrCh := make(chan error, 1)
	go func() {
		serverErrCh <- server.ListenAndServe()
	}()

	shutdownCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	select {
	case err := <-serverErrCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server stopped", "error", err)
			os.Exit(1)
		}

		logger.Info("server stopped")
		return
	case <-shutdownCtx.Done():
		logger.Info("shutdown signal received")
		stop()
	}

	ctx, cancel := context.WithTimeout(context.Background(), cfg.HTTP.ShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("graceful shutdown failed", "error", err)
		os.Exit(1)
	}

	if err := <-serverErrCh; err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Error("server stopped with unexpected error", "error", err)
		os.Exit(1)
	}

	logger.Info("server stopped gracefully")
}

func newLogger(level string) *slog.Logger {
	var slogLevel slog.Level

	switch level {
	case "debug":
		slogLevel = slog.LevelDebug
	case "warn":
		slogLevel = slog.LevelWarn
	case "error":
		slogLevel = slog.LevelError
	default:
		slogLevel = slog.LevelInfo
	}

	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slogLevel,
	}))
}
