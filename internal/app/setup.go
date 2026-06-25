package app

import (
	"context"
	"fmt"

	"github.com/glebarez/sqlite"
	"github.com/labstack/echo/v5"

	echomw "github.com/labstack/echo/v5/middleware"
	"github.com/osamikoyo/math-angel/internal/config"
	"github.com/osamikoyo/math-angel/internal/task/cache"
	taskhandler "github.com/osamikoyo/math-angel/internal/task/handler"
	"github.com/osamikoyo/math-angel/internal/task/importer"
	taskmodel "github.com/osamikoyo/math-angel/internal/task/model"
	taskrepo "github.com/osamikoyo/math-angel/internal/task/repository"
	taskservice "github.com/osamikoyo/math-angel/internal/task/service"
	userhandler "github.com/osamikoyo/math-angel/internal/user/handler"
	"github.com/osamikoyo/math-angel/internal/user/handler/middleware"
	usermodel "github.com/osamikoyo/math-angel/internal/user/model"
	userrepo "github.com/osamikoyo/math-angel/internal/user/repository"
	userservice "github.com/osamikoyo/math-angel/internal/user/service"
	"github.com/osamikoyo/math-angel/pkg/logger"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

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
func setupRepos(logger *logger.Logger, cfg *config.Config) (*taskrepo.Repository, *userrepo.Repository, error) {
	db, err := gorm.Open(sqlite.Open(cfg.DBpath))
	if err != nil {
		logger.Error("failed connect to db", zap.String("path", cfg.DBpath), zap.Error(err))
		return nil, nil, fmt.Errorf("failed connect to db: %w", err)
	}

	if err := db.AutoMigrate(&taskmodel.Task{}, &usermodel.User{}, &usermodel.Profile{}); err != nil {
		logger.Error("migration failed",
			zap.Error(err))

		return nil, nil, fmt.Errorf("failed migrate: %w", err)
	}

	// === FTS5 SETUP ===
	if err := setupFTS5(db, logger); err != nil {
		logger.Error("FTS5 setup failed", zap.Error(err))
		return nil, nil, fmt.Errorf("failed setup FTS5: %w", err)
	}

	logger.Info("database connected successfully")
	return taskrepo.NewRepository(db, logger), userrepo.NewRepository(db, logger), nil
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
func setupImporter(service *taskservice.TaskService, logger *logger.Logger, cfg *config.Config) (*importer.Importer, error) {
	importer, err := importer.NewImporter(service, cfg, logger)
	if err != nil {
		logger.Error("failed to setup importer", zap.Error(err))
		return nil, fmt.Errorf("failed setup importer: %w", err)
	}
	logger.Info("importer setup successfully")
	return importer, nil
}

// setupEcho configures the Echo framework with routes.
func setupEcho(taskservice *taskservice.TaskService, userservice *userservice.Service, logger *logger.Logger) *echo.Echo {
	e := echo.New()
	e.Use(echomw.RequestLogger())

	mw := middleware.NewMiddleware(userservice)
	taskhandler := taskhandler.NewHandler(taskservice, mw)
	userhandler := userhandler.NewHandler(mw, userservice)

	userhandler.RegisterRouters(e)
	taskhandler.RegisterRouters(e)

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
