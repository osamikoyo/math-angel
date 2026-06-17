package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/osamikoyo/math-angel/internal/config"
	"github.com/osamikoyo/math-angel/internal/task/cachedrepo"
	"github.com/osamikoyo/math-angel/internal/task/importer"
	taskservice "github.com/osamikoyo/math-angel/internal/task/service"
	userservice "github.com/osamikoyo/math-angel/internal/user/service"
	"github.com/osamikoyo/math-angel/pkg/logger"
	"go.uber.org/zap"
)

// App represents the main application containing the HTTP server, configuration, and dependencies.
type App struct {
	echo     *echo.Echo         // Echo framework for handling HTTP requests
	httpSrv  *http.Server       // HTTP server
	cfg      *config.Config     // Application configuration
	importer *importer.Importer // Data importer (if enabled)
	logger   *logger.Logger     // Logger for recording events
}

// SetupApp initializes the application by setting up configuration, logger, repository, cache, and other components.
func SetupApp(configPath string) (*App, error) {
	cfg, logger, err := setupCfgAndLogger(configPath)
	if err != nil {
		return nil, err
	}

	logger.Info("configuration and logger initialized successfully",
		zap.Any("cfg", cfg))

	taskrepo, userrepo, err := setupRepos(logger, cfg)
	if err != nil {
		return nil, err
	}

	cache, err := setupCache(logger, cfg)
	if err != nil {
		return nil, err
	}

	cachedrepo := cachedrepo.NewCachedRepository(taskrepo, cache)

	taskservice := taskservice.NewTaskService(cachedrepo, cfg.Timeout)

	userservice := userservice.NewService(userrepo, cachedrepo, cfg, cfg.Timeout)

	var importer *importer.Importer
	if cfg.Importer.Enabled {
		importer, err = setupImporter(taskservice, logger, cfg)
		if err != nil {
			return nil, err
		}
	}

	e := setupEcho(taskservice, userservice, logger)

	httpSrv := &http.Server{
		Addr:    cfg.Addr,
		Handler: e,
	}

	return &App{
		echo:     e,
		httpSrv:  httpSrv,
		cfg:      cfg,
		importer: importer,
		logger:   logger,
	}, nil
}

// Run starts the application, including the HTTP server and importer, and handles shutdown signals.
func (a *App) Run() error {
	ctx, stop := signal.NotifyContext(context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
		syscall.SIGINT,
	)
	defer stop()

	if a.cfg.Importer.Enabled && a.importer != nil {
		a.importer.Start(ctx)
	}

	errs := make(chan error, 1)

	go func() {
		a.logger.Info("starting server", zap.String("addr", a.cfg.Addr))

		if err := a.httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.logger.Error("server error", zap.Error(err))

			errs <- fmt.Errorf("server error: %w", err)
		}
	}()

	select {
	case <-ctx.Done():
		// Shutdown signal received, proceed to graceful shutdown
		a.logger.Info("received shutdown signal, gracefully stopping...")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := a.shutdown(shutdownCtx); err != nil {
			a.logger.Error("graceful shutdown failed", zap.Error(err))
			return err
		}

		a.logger.Info("server stopped gracefully")

		return nil
	case err := <-errs:
		a.logger.Error("server error", zap.Error(err))
		return err
	}
}

// shutdown gracefully shuts down the HTTP server.
func (a *App) shutdown(ctx context.Context) error {
	if err := a.httpSrv.Shutdown(ctx); err != nil {
		a.logger.Error("http server shutdown error", zap.Error(err))
	}

	return nil
}

