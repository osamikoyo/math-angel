package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/osamikoyo/math-angel/internal/model"
	"github.com/osamikoyo/math-angel/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var (
	ErrEmptyTask    = errors.New("empty task")
	ErrAlreadyExist = errors.New("task already exist")
	ErrUnknown      = errors.New("unknown error")
	ErrNotFound     = errors.New("not found")
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

func (r *Repository) UpdateTask(ctx context.Context, id uuid.UUID, column string, value any) error {
	rows, err := gorm.G[model.Task](r.db).Update(ctx, column, value)
	if rows == 0 {
		r.logger.Error("not found task",
			zap.String("id", id.String()))

		return ErrNotFound
	}

	if err != nil {
		r.logger.Error("failed update task",
			zap.String("column", column),
			zap.Any("value", value),
			zap.Error(err))

		return ErrUnknown
	}

	r.logger.Info("task updated successfully",
		zap.String("id", id.String()),
		zap.String("column", column))

	return nil
}

func (r *Repository) GetTasksByTypeAndLevel(ctx context.Context, taskType string, level uint) ([]model.Task, error) {
	tasks, err := gorm.G[model.Task](r.db).Where("type = ? AND level = ?", taskType, level).Find(ctx)
	if len(tasks) == 0 {
		r.logger.Error("not found tasks",
			zap.String("type", taskType),
			zap.Uint("level", level))

		return nil, ErrNotFound
	}
	if err != nil {
		r.logger.Error("failed get tasks",
			zap.String("type", taskType),
			zap.Uint("level", level),
			zap.Error(err))

		return nil, ErrUnknown
	}

	r.logger.Info("tasks was successfully fetched")

	return tasks, nil
}

func (r *Repository) GetTask(ctx context.Context, id uuid.UUID) (*model.Task, error) {
	task, err := gorm.G[model.Task](r.db).Where("id = ?", id).First(ctx)
	if err != nil {
		r.logger.Error("failed get task",
			zap.String("id", id.String()),
			zap.Error(err))

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}

		return nil, ErrUnknown
	}

	r.logger.Info("task fetched successfully",
		zap.Any("task", task))

	return &task, nil
}
