package cachedrepo

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/osamikoyo/math-angel/internal/errors"
	"github.com/osamikoyo/math-angel/internal/task/model"
	"github.com/osamikoyo/math-angel/internal/task/service"
)

// Repository defines the interface for task persistence operations.
type Repository interface {
	CreateTask(ctx context.Context, task *model.Task) error
	GetTasksByTypeAndLevel(ctx context.Context, taskType string, level string) ([]model.Task, error)
	GetTask(ctx context.Context, id uuid.UUID) (*model.Task, error)
	UpdateTask(ctx context.Context, id uuid.UUID, column string, value any) error
	SearchTasks(ctx context.Context, query string, limit, offset int) ([]model.TaskSearchResult, error)
	CreateSolution(ctx context.Context, solution *model.Solution) error
	GetSolutionByUserAndTaskIDs(ctx context.Context, userID uint, taskID string) (*model.Solution, error)
}

// Cache defines the interface for caching operations.
type Cache interface {
	SetTask(ctx context.Context, key string, task *model.Task) error
	SetTasks(ctx context.Context, key string, tasks []model.Task) error
	GetTasks(ctx context.Context, key string) ([]model.Task, error)
	GetTask(ctx context.Context, key string) (*model.Task, error)
	SetSearchResults(ctx context.Context, key string, trs []model.TaskSearchResult) error
	GetSearchResults(ctx context.Context, key string) ([]model.TaskSearchResult, error)
	SetSolution(ctx context.Context, key string, solution *model.Solution) error
	GetSolution(ctx context.Context, key string) (*model.Solution, error)
}

// CachedRepository stores cache and repo logic
type CachedRepository struct {
	repo  Repository
	cache Cache
}

var _ service.CachedRepo = &CachedRepository{}

func NewCachedRepository(repo Repository, cache Cache) *CachedRepository {
	return &CachedRepository{
		repo:  repo,
		cache: cache,
	}
}

func (cr *CachedRepository) CreateTask(ctx context.Context, task *model.Task) error {
	if task == nil {
		return errors.ErrEmptyTask
	}

	if err := cr.repo.CreateTask(ctx, task); err != nil {
		return err
	}

	cr.cache.SetTask(ctx, getKeyForOne(task.UID), task)

	return nil
}

func (cr *CachedRepository) UpdateTask(ctx context.Context, uid uuid.UUID, column string, value any) error {
	if err := cr.repo.UpdateTask(ctx, uid, column, value); err != nil {
		return err
	}

	task, err := cr.repo.GetTask(ctx, uid)
	if err != nil {
		return err
	}

	cr.cache.SetTask(ctx, getKeyForOne(uid.String()), task)

	return nil
}

func (cr *CachedRepository) GetTask(ctx context.Context, id uuid.UUID) (*model.Task, error) {
	var (
		task *model.Task
		err  error
	)

	task, err = cr.cache.GetTask(ctx, getKeyForOne(id.String()))
	if err == nil {
		return task, nil
	}

	task, err = cr.repo.GetTask(ctx, id)
	if err != nil {
		return nil, err
	}

	cr.cache.SetTask(ctx, getKeyForOne(id.String()), task)

	return task, nil
}

func (cr *CachedRepository) GetTasks(ctx context.Context, taskType, level string) ([]model.Task, error) {
	var (
		tasks []model.Task
		err   error
	)

	tasks, err = cr.cache.GetTasks(ctx, getKeyForMany(taskType, level))
	if err == nil {
		return tasks, nil
	}

	tasks, err = cr.repo.GetTasksByTypeAndLevel(ctx, taskType, level)
	if err != nil {
		return nil, err
	}

	cr.cache.SetTasks(ctx, getKeyForMany(taskType, level), tasks)

	return tasks, nil
}

func (cr *CachedRepository) Search(ctx context.Context, query string, pageIndex, pageSize int) ([]model.TaskSearchResult, error) {
	results, err := cr.cache.GetSearchResults(ctx, getKeyForSearchQuery(query))
	if err == nil {
		start := pageSize * (pageIndex - 1)
		return results[start:min(start+pageSize, len(results))], nil
	}

	offset := (pageIndex - 1) * pageSize
	results, err = cr.repo.SearchTasks(ctx, query, pageSize, offset)
	if err != nil {
		return nil, err
	}

	return results, nil
}

func (cr *CachedRepository) CreateSolution(ctx context.Context, solution *model.Solution) error {
	if err := cr.repo.CreateSolution(ctx, solution); err != nil {
		return err
	}

	cr.cache.SetSolution(ctx, getKeyForSolution(solution.UserID, solution.TaskID), solution)

	return nil
}

func (cr *CachedRepository) GetSolution(ctx context.Context, userID uint, taskID string) (*model.Solution, error) {
	solution, err := cr.cache.GetSolution(ctx, getKeyForSolution(userID, taskID))
	if err == nil{
		return solution, nil
	}

	solution, err = cr.repo.GetSolutionByUserAndTaskIDs(ctx, userID, taskID)
	if err != nil{
		return nil, err
	}

	cr.cache.SetSolution(ctx, getKeyForSolution(userID, taskID), solution)

	return solution, nil
} 

// getKeyForMany generates a cache key for multiple tasks by type and level.
func getKeyForMany(taskType string, level string) string {
	return fmt.Sprintf("%s:%s", taskType, level)
}

// getKeyForOne generates a cache key for a single task by ID.
func getKeyForOne(key string) string {
	return fmt.Sprintf("one:%s", key)
}

func getKeyForSearchQuery(query string) string {
	return fmt.Sprintf("query:%s", query)
}

func getKeyForSolution(userID uint, taskID string) string {
	return fmt.Sprintf("solution:user_id:%d;task_id:%s", userID, taskID)
}
