package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/labstack/echo/v5"
	"github.com/osamikoyo/math-angel/internal/config"
	"github.com/osamikoyo/math-angel/internal/task/cache"
	"github.com/osamikoyo/math-angel/internal/task/cachedrepo"
	"github.com/osamikoyo/math-angel/internal/task/handler"
	"github.com/osamikoyo/math-angel/internal/task/importer"
	"github.com/osamikoyo/math-angel/internal/task/model"
	"github.com/osamikoyo/math-angel/internal/task/repository"
	"github.com/osamikoyo/math-angel/internal/task/service"
	"github.com/osamikoyo/math-angel/pkg/logger"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
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

	repo, err := setupRepo(logger, cfg)
	if err != nil {
		return nil, err
	}

	cache, err := setupCache(logger, cfg)
	if err != nil {
		return nil, err
	}

	cachedrepo := cachedrepo.NewCachedRepository(repo, cache)

	service := service.NewTaskService(cachedrepo, cfg.Timeout)

	var importer *importer.Importer
	if cfg.Importer.Enabled {
		importer, err = setupImporter(service, logger, cfg)
		if err != nil {
			return nil, err
		}
	}

	e := setupEcho(service, logger)

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

// setupCfgAndLogger loads the configuration and initializes the logger.
func setupCfgAndLogger(configPath string) (*config.Config, *logger.Logger, error) {
	logger.Init(logger.Config{
		AppName:   "math-angel",
		AddCaller: false,
		LogFile:   "logs/math-angel.log",
		LogLevel:  "debug",
	})
	l := logger.Get()

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		l.Error("failed load config", zap.String("path", configPath), zap.Error(err))
		return nil, nil, fmt.Errorf("failed load config: %w", err)
	}
	return cfg, l, nil
}

// setupRepo connects to the database and performs migrations.
func setupRepo(logger *logger.Logger, cfg *config.Config) (*repository.Repository, error) {
	db, err := gorm.Open(sqlite.Open(cfg.DBpath))
	if err != nil {
		logger.Error("failed connect to db", zap.String("path", cfg.DBpath), zap.Error(err))
		return nil, fmt.Errorf("failed connect to db: %w", err)
	}

	if err := db.AutoMigrate(&model.Task{}); err != nil {
		logger.Error("migration failed",
			zap.Error(err))

		return nil, fmt.Errorf("failed migrate: %w", err)
	}

	// === FTS5 SETUP ===
	if err := setupFTS5(db, logger); err != nil {
		logger.Error("FTS5 setup failed", zap.Error(err))
		return nil, fmt.Errorf("failed setup FTS5: %w", err)
	}

	logger.Info("database connected successfully")
	return repository.NewRepository(db, logger), nil
}

// setupCache connects to Redis for caching.
func setupCache(logger *logger.Logger, cfg *config.Config) (*cache.Cache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: "",
		DB:       0,
		Protocol: 2,
	})

	if err := client.Ping(context.Background()).Err(); err != nil {
		logger.Error("failed connect to redis", zap.String("addr", cfg.Redis.Addr), zap.Error(err))
		return nil, fmt.Errorf("failed connect to cache: %w", err)
	}

	logger.Info("redis connected successfully")
	return cache.NewCache(client, logger, cfg.Redis.ExpTime), nil
}

// setupImporter initializes the data importer.
func setupImporter(service *service.TaskService, logger *logger.Logger, cfg *config.Config) (*importer.Importer, error) {
	importer, err := importer.NewImporter(service, cfg, logger)
	if err != nil {
		logger.Error("failed to setup importer", zap.Error(err))
		return nil, fmt.Errorf("failed setup importer: %w", err)
	}
	logger.Info("importer setup successfully")
	return importer, nil
}

// setupEcho configures the Echo framework with routes.
func setupEcho(service *service.TaskService, logger *logger.Logger) *echo.Echo {
	e := echo.New()
	handler := handler.NewHandler(service)
	handler.RegisterRouters(e)

	logger.Info("echo configured successfully")
	return e
}

func setupFTS5(db *gorm.DB, logger *logger.Logger) error {
	ftsSQL := `
		CREATE VIRTUAL TABLE IF NOT EXISTS tasks_fts USING fts5(         
			problem,
			solution,
			tokenize='unicode61 remove_diacritics 0',
			content='tasks',
			content_rowid='id'
		);
	`
	if err := db.Exec(ftsSQL).Error; err != nil {
		return err
	}

	triggers := []string{
		// AFTER INSERT
		`CREATE TRIGGER IF NOT EXISTS tasks_ai AFTER INSERT ON tasks
		 BEGIN
			INSERT INTO tasks_fts(rowid, problem, solution)
			VALUES (new.id, new.problem, new.solution);
		 END;`,

		// AFTER UPDATE
		`CREATE TRIGGER IF NOT EXISTS tasks_au AFTER UPDATE ON tasks
		 BEGIN
			UPDATE tasks_fts SET 
				problem = new.problem,
				solution = new.solution
			WHERE rowid = new.id;
		 END;`,

		// AFTER DELETE
		`CREATE TRIGGER IF NOT EXISTS tasks_ad AFTER DELETE ON tasks
		 BEGIN
			DELETE FROM tasks_fts WHERE rowid = old.id;
		 END;`,
	}

	for _, trig := range triggers {
		if err := db.Exec(trig).Error; err != nil {
			logger.Warn("failed to create trigger", zap.Error(err))
		}
	}

	logger.Info("FTS5 virtual table and triggers created successfully")
	return nil
}
