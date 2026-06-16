package service

import (
	"context"
	"math/rand"
	"slices"
	"time"

	"github.com/google/uuid"
	"github.com/osamikoyo/math-angel/internal/errors"
	"github.com/osamikoyo/math-angel/internal/task/model"
)

type CachedRepo interface {
	CreateTask(ctx context.Context, task *model.Task) error
	GetTask(ctx context.Context, id uuid.UUID) (*model.Task, error)
	GetTasks(ctx context.Context, taskType, level string) ([]model.Task, error)
	Search(ctx context.Context, query string, pageIndex, pageSize int) ([]model.TaskSearchResult, error)
	UpdateTask(ctx context.Context, uid uuid.UUID, column string, value any) error
	CreateSolution(ctx context.Context, solution *model.Solution) error
	GetSolution(ctx context.Context, userID uint, taskID string) (*model.Solution, error)
}

// TaskService provides business logic for task management, including caching and repository interactions.
type TaskService struct {
	cachedrepo CachedRepo

	timeout time.Duration // Timeout for operations
}

// NewService creates a new Service instance with the given repository, cache, and timeout.
func NewTaskService(cachedrepo CachedRepo, timeout time.Duration) *TaskService {
	return &TaskService{
		cachedrepo: cachedrepo,
		timeout:    timeout,
	}
}

// CreateTask creates a new task, stores it in cache and repository.
func (s *TaskService) CreateTask(
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

	if err := task.Validate(); err != nil {
		return err
	}

	if err := s.cachedrepo.CreateTask(ctx, task); err != nil {
		return err
	}

	return nil
}

// Search finds tasks in db by query
func (s *TaskService) Search(reqCtx context.Context, query string, pageIndex int, pageSize int) ([]model.TaskSearchResult, error) {
	if query == "" {
		return nil, errors.ErrEmptyQuery
	}

	if pageIndex < 1 {
		pageIndex = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	ctx, cancel := context.WithTimeout(reqCtx, s.timeout)
	defer cancel()

	tasks, err := s.cachedrepo.Search(ctx, query, pageIndex, pageSize)
	if err != nil {
		return nil, err
	}

	return tasks, nil
}

// IncLike increments the like count for a task by ID.
func (s *TaskService) IncLike(reqCtx context.Context, id string) error {
	ctx, cancel := context.WithTimeout(reqCtx, s.timeout)
	defer cancel()

	uid, err := uuid.Parse(id)
	if err != nil {
		return errors.ErrBadUID
	}

	task, err := s.cachedrepo.GetTask(ctx, uid)
	if err != nil {
		return err
	}

	task.Likes++

	if err = s.cachedrepo.UpdateTask(ctx, uid, "likes", task.Likes); err != nil {
		return err
	}

	return nil
}

// DecLike decrements the like count for a task by ID.
func (s *TaskService) DecLike(reqCtx context.Context, id string) error {
	ctx, cancel := context.WithTimeout(reqCtx, s.timeout)
	defer cancel()

	uid, err := uuid.Parse(id)
	if err != nil {
		return errors.ErrBadUID
	}

	task, err := s.cachedrepo.GetTask(ctx, uid)
	if err != nil {
		return err
	}

	if task.Likes == 0 {
		return nil
	}

	task.Likes--

	if err = s.cachedrepo.UpdateTask(ctx, uid, "likes", task.Likes); err != nil {
		return err
	}

	return nil
}

// IncDislike increments the dislike count for a task by ID.
func (s *TaskService) IncDislike(reqCtx context.Context, id string) error {
	ctx, cancel := context.WithTimeout(reqCtx, s.timeout)
	defer cancel()

	uid, err := uuid.Parse(id)
	if err != nil {
		return errors.ErrBadUID
	}

	task, err := s.cachedrepo.GetTask(ctx, uid)
	if err != nil {
		return err
	}

	task.Dislikes++

	if err = s.cachedrepo.UpdateTask(ctx, uid, "dislikes", task.Dislikes); err != nil {
		return err
	}

	return nil
}

// DecDislike decrements the dislike count for a task by ID.
func (s *TaskService) DecDislike(reqCtx context.Context, id string) error {
	ctx, cancel := context.WithTimeout(reqCtx, s.timeout)
	defer cancel()

	uid, err := uuid.Parse(id)
	if err != nil {
		return errors.ErrBadUID
	}

	task, err := s.cachedrepo.GetTask(ctx, uid)
	if err != nil {
		return err
	}

	if task.Dislikes == 0 {
		return nil
	}

	task.Dislikes--

	if err = s.cachedrepo.UpdateTask(ctx, uid, "dislikes", task.Dislikes); err != nil {
		return err
	}

	return nil
}

// GetRandomTask retrieves a random task of the specified type and level, using cache if available.
func (s *TaskService) GetRandomTask(reqCtx context.Context, taskType string, level string) (*model.Task, error) {
	ctx, cancel := context.WithTimeout(reqCtx, s.timeout)
	defer cancel()

	tasks, err := s.cachedrepo.GetTasks(ctx, taskType, level)
	if err != nil {
		return nil, err
	}

	task := getRandomFromArr(tasks)

	return &task, nil
}

// GetTask retrieves a task by ID, checking cache first, then repository.
func (s *TaskService) GetTask(reqCtx context.Context, id string) (*model.Task, error) {
	ctx, cancel := context.WithTimeout(reqCtx, s.timeout)
	defer cancel()

	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, errors.ErrBadUID
	}

	task, err := s.cachedrepo.GetTask(ctx, uid)
	if err != nil {
		return nil, err
	}

	return task, nil
}

// GetBests retrieves tasks by likes for the specified type and level, with pagination.
func (s *TaskService) GetTasks(reqCtx context.Context, taskType, level string, pageSize, pageIndex uint) ([]model.Task, error) {
	ctx, cancel := context.WithTimeout(reqCtx, s.timeout)
	defer cancel()

	tasks, err := s.cachedrepo.GetTasks(ctx, taskType, level)
	if err != nil {
		return nil, err
	}

	start := pageSize * (pageIndex - 1)
	return tasks[start:min(start+pageSize, uint(len(tasks)))], nil
}

// GetBests retrieves the top tasks by likes for the specified type and level, with pagination.
func (s *TaskService) GetBests(reqCtx context.Context, taskType string, level string, pageSize, pageIndex uint) ([]model.Task, error) {
	ctx, cancel := context.WithTimeout(reqCtx, s.timeout)
	defer cancel()

	tasks, err := s.cachedrepo.GetTasks(ctx, taskType, level)
	if err != nil {
		return nil, err
	}

	tasks = sortTasksByLikes(tasks)

	slices.Reverse(tasks)

	start := pageSize * (pageIndex - 1)
	return tasks[start:min(start+pageSize, uint(len(tasks)))], nil
}

func (s *TaskService) CreateSolution(reqCtx context.Context, taskID string, userID uint) error {
	ctx, cancel := context.WithTimeout(reqCtx, s.timeout)
	defer cancel()

	solution := model.NewSolution(taskID, userID)

	if err := s.cachedrepo.CreateSolution(ctx, solution);err != nil{
		return err
	}

	return nil
}

func (s *TaskService) TaskSolvedBy(reqCtx context.Context, taskID string, userID uint) (time.Time, error) {
	ctx, cancel := context.WithTimeout(reqCtx, s.timeout)
	defer cancel()

	solution, err := s.cachedrepo.GetSolution(ctx, userID, taskID)
	if err != nil{
		return time.Time{}, err
	}

	return solution.CreatedAt, nil
}

// getRandomFromArr selects a random element from a slice.
func getRandomFromArr[T any](arr []T) T {
	if len(arr) == 0 {
		panic("getRandomFromArr: empty slice")
	}

	randomIndex := rand.Intn(len(arr))

	return arr[randomIndex]
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
