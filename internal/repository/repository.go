package repository

import (
	"context"
	"errors"

	selferrors "github.com/osamikoyo/math-angel/internal/errors"
	"github.com/osamikoyo/math-angel/internal/service"

	"github.com/google/uuid"
	"github.com/osamikoyo/math-angel/internal/model"
	"github.com/osamikoyo/math-angel/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Repository struct {
	logger *logger.Logger
	db     *gorm.DB
}

var _ service.Repository = &Repository{}

func NewRepository(db *gorm.DB, logger *logger.Logger) *Repository {
	return &Repository{
		db:     db,
		logger: logger,
	}
}

func (r *Repository) CreateTask(ctx context.Context, task *model.Task) error {
	if task == nil {
		return selferrors.ErrEmptyTask
	}

	r.logger.Info("create task",
		zap.Any("task", task))

	err := gorm.G[model.Task](r.db).Create(ctx, task)
	if err != nil {
		r.logger.Error("failed create task",
			zap.Any("task", task),
			zap.Error(err))

		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return selferrors.ErrAlreadyExist
		}

		return selferrors.ErrUnknown
	}

	r.logger.Info("task created successfully",
		zap.Any("task", task))

	return nil
}

func (r *Repository) UpdateTask(ctx context.Context, id uuid.UUID, column string, value any) error {
	r.logger.Info("updating task",
		zap.String("id", id.String()),
		zap.String("column", column),
		zap.Any("value", value))

	rows, err := gorm.G[model.Task](r.db).Update(ctx, column, value)
	if rows == 0 {
		r.logger.Error("not found task",
			zap.String("id", id.String()))

		return selferrors.ErrNotFound
	}

	if err != nil {
		r.logger.Error("failed update task",
			zap.String("column", column),
			zap.Any("value", value),
			zap.Error(err))

		return selferrors.ErrUnknown
	}

	r.logger.Info("task updated successfully",
		zap.String("id", id.String()),
		zap.String("column", column))

	return nil
}

func (r *Repository) GetTasksByTypeAndLevel(ctx context.Context, taskType string, level string) ([]model.Task, error) {
	r.logger.Info("fetching tasks",
		zap.String("type", taskType),
		zap.String("level", level))

	var (
		tasks []model.Task
		err   error
	)

	switch {
	case taskType == "all" && level == "all":
		tasks, err = gorm.G[model.Task](r.db).Find(ctx)
	case taskType == "all":
		tasks, err = gorm.G[model.Task](r.db).Where("level = ?", level).Find(ctx)
	case level == "all":
		tasks, err = gorm.G[model.Task](r.db).Where("type = ?", taskType).Find(ctx)
	default:
		tasks, err = gorm.G[model.Task](r.db).Where("type = ? AND level = ?", taskType, level).Find(ctx)
	}

	if err != nil {
		r.logger.Error("failed get tasks",
			zap.String("type", taskType),
			zap.String("level", level),
			zap.Error(err))

		return nil, selferrors.ErrUnknown
	}

	if len(tasks) == 0 {
		r.logger.Error("not found tasks",
			zap.String("type", taskType),
			zap.String("level", level))

		return nil, selferrors.ErrNotFound
	}

	r.logger.Info("tasks was successfully fetched")

	return tasks, nil
}

func (r *Repository) GetTask(ctx context.Context, id uuid.UUID) (*model.Task, error) {
	r.logger.Info("fetch task",
		zap.String("id", id.String()))

	task, err := gorm.G[model.Task](r.db).Where("id = ?", id).First(ctx)
	if err != nil {
		r.logger.Error("failed get task",
			zap.String("id", id.String()),
			zap.Error(err))

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, selferrors.ErrNotFound
		}

		return nil, selferrors.ErrUnknown
	}

	r.logger.Info("task fetched successfully",
		zap.Any("task", task))

	return &task, nil
}
