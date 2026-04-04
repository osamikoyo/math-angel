package service

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/osamikoyo/math-angel/internal/errors"
	"github.com/osamikoyo/math-angel/internal/model"
)

// Repository defines the interface for task persistence operations.
type Repository interface {
	CreateTask(ctx context.Context, task *model.Task) error
	GetTasksByTypeAndLevel(ctx context.Context, taskType string, level string) ([]model.Task, error)
	GetTask(ctx context.Context, id uuid.UUID) (*model.Task, error)
	UpdateTask(ctx context.Context, id uuid.UUID, column string, value any) error
}

// Cash defines the interface for caching operations.
type Cash interface {
	SetTask(ctx context.Context, key string, task *model.Task) error
	SetTasks(ctx context.Context, key string, tasks []model.Task) error
	GetTasks(ctx context.Context, key string) ([]model.Task, error)
	GetTask(ctx context.Context, key string) (*model.Task, error)
}

// Service provides business logic for task management, including caching and repository interactions.
type Service struct {
	repo Repository // Repository for data persistence
	cash Cash       // Cache for fast data access

	timeout time.Duration // Timeout for operations
}

// NewService creates a new Service instance with the given repository, cache, and timeout.
func NewService(repo Repository, cash Cash, timeout time.Duration) *Service {
	return &Service{
		repo:    repo,
		cash:    cash,
		timeout: timeout,
	}
}

// CreateTask creates a new task, stores it in cache and repository.
func (s *Service) CreateTask(
	reqCtx context.Context,
	taskType string,
	problem string,
	solution string,
	boxed string,
	level string,
) error {
	ctx, cancel := context.WithTimeout(reqCtx, s.timeout)
	defer cancel()

	task := model.NewTask(
		taskType,
		problem,
		solution,
		boxed,
		level,
	)

	if err := s.cash.SetTask(ctx, getKeyForOne(task.ID.String()), task); err != nil {
		return err
	}

	if err := s.repo.CreateTask(ctx, task); err != nil {
		return err
	}

	return nil
}

// IncLike increments the like count for a task by ID.
func (s *Service) IncLike(reqCtx context.Context, id string) error {
	ctx, cancel := context.WithTimeout(reqCtx, s.timeout)
	defer cancel()

	uid, err := uuid.Parse(id)
	if err != nil {
		return errors.ErrBadUID
	}

	task, err := s.repo.GetTask(ctx, uid)
	if err != nil {
		return err
	}

	task.Likes++

	s.cash.SetTask(ctx, getKeyForOne(task.ID.String()), task)
	if err = s.repo.UpdateTask(ctx, uid, "likes", task.Likes); err != nil {
		return err
	}

	return nil
}

// DecLike decrements the like count for a task by ID.
func (s *Service) DecLike(reqCtx context.Context, id string) error {
	ctx, cancel := context.WithTimeout(reqCtx, s.timeout)
	defer cancel()

	uid, err := uuid.Parse(id)
	if err != nil {
		return errors.ErrBadUID
	}

	task, err := s.repo.GetTask(ctx, uid)
	if err != nil {
		return err
	}

	task.Likes--

	s.cash.SetTask(ctx, getKeyForOne(task.ID.String()), task)
	if err = s.repo.UpdateTask(ctx, uid, "likes", task.Likes); err != nil {
		return err
	}

	return nil
}

// IncDislike increments the dislike count for a task by ID.
func (s *Service) IncDislike(reqCtx context.Context, id string) error {
	ctx, cancel := context.WithTimeout(reqCtx, s.timeout)
	defer cancel()

	uid, err := uuid.Parse(id)
	if err != nil {
		return errors.ErrBadUID
	}

	task, err := s.repo.GetTask(ctx, uid)
	if err != nil {
		return err
	}

	task.Dislikes++

	s.cash.SetTask(ctx, getKeyForOne(task.ID.String()), task)
	if err = s.repo.UpdateTask(ctx, uid, "dislikes", task.Dislikes); err != nil {
		return err
	}

	return nil
}

// DecDislike decrements the dislike count for a task by ID.
func (s *Service) DecDislike(reqCtx context.Context, id string) error {
	ctx, cancel := context.WithTimeout(reqCtx, s.timeout)
	defer cancel()

	uid, err := uuid.Parse(id)
	if err != nil {
		return errors.ErrBadUID
	}

	task, err := s.repo.GetTask(ctx, uid)
	if err != nil {
		return err
	}

	task.Dislikes--

	s.cash.SetTask(ctx, getKeyForOne(task.ID.String()), task)
	if err = s.repo.UpdateTask(ctx, uid, "dislikes", task.Dislikes); err != nil {
		return err
	}

	return nil
}

// GetRandomTask retrieves a random task of the specified type and level, using cache if available.
func (s *Service) GetRandomTask(reqCtx context.Context, taskType string, level string) (*model.Task, error) {
	ctx, cancel := context.WithTimeout(reqCtx, s.timeout)
	defer cancel()

	cashedTasks, err := s.cash.GetTasks(ctx, getKeyForMany(taskType, level))
	if err == nil && cashedTasks != nil && len(cashedTasks) != 0 {
		if len(cashedTasks) > 1 {
			task := getRandomFromArr(cashedTasks)

			return &task, nil
		}
	}

	tasks, err := s.repo.GetTasksByTypeAndLevel(ctx, taskType, level)
	if err != nil {
		return nil, err
	}

	s.cash.SetTasks(ctx, getKeyForMany(taskType, level), tasks)

	task := getRandomFromArr(tasks)

	return &task, nil
}

// GetTask retrieves a task by ID, checking cache first, then repository.
func (s *Service) GetTask(reqCtx context.Context, id string) (*model.Task, error) {
	ctx, cancel := context.WithTimeout(reqCtx, s.timeout)
	defer cancel()

	var (
		task *model.Task
		err  error
	)

	task, err = s.cash.GetTask(ctx, getKeyForOne(id))
	if err == nil {
		return task, nil
	}

	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, errors.ErrBadUID
	}

	task, err = s.repo.GetTask(ctx, uid)
	if err != nil {
		return nil, err
	}

	s.cash.SetTask(ctx, getKeyForOne(id), task)

	return task, nil
}

// GetBests retrieves the top tasks by likes for the specified type and level, with pagination.
func (s *Service) GetBests(reqCtx context.Context, taskType string, level string, pageSize, pageIndex uint) ([]model.Task, error) {
	ctx, cancel := context.WithTimeout(reqCtx, s.timeout)
	defer cancel()

	var (
		tasks []model.Task
		err   error
	)

	tasks, err = s.cash.GetTasks(ctx, getKeyForBests(taskType, level))
	if err == nil {
		start := pageSize * (pageIndex - 1)
		if start >= uint(len(tasks)) {
			return []model.Task{}, nil
		}
		return tasks[start:min(start+pageSize, uint(len(tasks)))], nil
	}

	tasks, err = s.cash.GetTasks(ctx, getKeyForMany(taskType, level))
	if err != nil {
		tasks, err = s.repo.GetTasksByTypeAndLevel(ctx, taskType, level)
		if err != nil {
			return nil, err
		}
	}

	tasks = sortTasksByLikes(tasks)

	s.cash.SetTasks(ctx, getKeyForMany(taskType, level), tasks)

	start := pageSize * (pageIndex - 1)
	return tasks[start:min(start+pageSize, uint(len(tasks)))], nil
}

// getKeyForMany generates a cache key for multiple tasks by type and level.
func getKeyForMany(taskType string, level string) string {
	return fmt.Sprintf("%s:%s", taskType, level)
}

// getKeyForOne generates a cache key for a single task by ID.
func getKeyForOne(key string) string {
	return fmt.Sprintf("one:%s", key)
}

// getRandomFromArr selects a random element from a slice.
func getRandomFromArr[T any](arr []T) T {
	if len(arr) == 0 {
		panic("getRandomFromArr: empty slice")
	}

	randomIndex := rand.Intn(len(arr))

	return arr[randomIndex]
}

// getKeyForBests generates a cache key for sorted best tasks by type and level.
func getKeyForBests(taskType string, level string) string {
	return fmt.Sprintf("sorted:%s:%s", taskType, level)
}

// sortTasksByLikes sorts tasks in descending order by likes using quicksort.
func sortTasksByLikes(tasks []model.Task) []model.Task {
	if len(tasks) <= 1 {
		return tasks
	}

	pivot := tasks[len(tasks)/2].Likes
	left := []model.Task{}
	right := []model.Task{}
	middle := []model.Task{}

	for _, x := range tasks {
		if x.Likes < pivot {
			left = append(left, x)
		} else if x.Likes == pivot {
			middle = append(middle, x)
		} else {
			right = append(right, x)
		}
	}

	left = sortTasksByLikes(left)
	right = sortTasksByLikes(right)

	return append(append(left, middle...), right...)
}
