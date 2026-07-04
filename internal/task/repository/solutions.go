package repository

import (
	"context"

	"errors"

	selferrors "github.com/osamikoyo/math-angel/internal/errors"
	"github.com/osamikoyo/math-angel/internal/task/model"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func (r *Repository) CreateSolution(ctx context.Context, solution *model.Solution) error {
	if solution == nil {
		r.logger.Error("empty solution")

		return selferrors.ErrEmptySolution
	}

	if err := gorm.G[model.Solution](r.db).Create(ctx, solution); err != nil {
		r.logger.Error("failed create solution",
			zap.Error(err))

		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil
		}

		return selferrors.ErrUnknown
	}

	r.logger.Info("solution created successfully")

	return nil
}

func (r *Repository) GetSolutionByUserAndTaskIDs(ctx context.Context, userID uint, taskID string) (*model.Solution, error) {
	r.logger.Info("fetching solution",
		zap.Uint("user_id", userID),
		zap.String("task_id", taskID))

	solution, err := gorm.G[*model.Solution](r.db).Where(
		"user_id = ? and task_id = ?", userID, taskID,
	).First(ctx)
	if err != nil {
		r.logger.Error("failed get solution",
			zap.Uint("user_id", userID),
			zap.String("task_id", taskID))

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, selferrors.ErrNotFound
		}

		return nil, selferrors.ErrUnknown
	}

	r.logger.Info("solution added successfully")

	return solution, nil
}
