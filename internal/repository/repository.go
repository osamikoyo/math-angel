package repository

import (
	"context"
	"errors"

	"github.com/osamikoyo/math-angel/internal/model"
	"github.com/osamikoyo/math-angel/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var (
	ErrEmptyTask    = errors.New("empty task")
	ErrAlreadyExist = errors.New("task already exist")
	ErrUnknown      = errors.New("unknown error")
)

type Repository struct {
	logger *logger.Logger
	db     *gorm.DB
}

func NewRepository(db *gorm.DB, logger *logger.Logger) *Repository {
	return &Repository{
		db:     db,
		logger: logger,
	}
}

func (r *Repository) CreateTask(ctx context.Context, task *model.Task) error {
	if task == nil {
		return ErrEmptyTask
	}

	err := gorm.G[model.Task](r.db).Create(ctx, task)
	if err != nil {
		r.logger.Error("failed create task",
			zap.Any("task", task),
			zap.Error(err))

		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return ErrAlreadyExist
		}

		return ErrUnknown
	}

	r.logger.Info("task created successfully",
		zap.Any("task", task))

	return nil
}
