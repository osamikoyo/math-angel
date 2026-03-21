package service

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/osamikoyo/math-angel/internal/model"
)

var ErrBadUID = errors.New("bad uid")

type Repository interface {
	CreateTask(ctx context.Context, task *model.Task) error
	GetTasksByTypeAndLevel(ctx context.Context, taskType string, level uint) ([]model.Task, error)
	GetTask(ctx context.Context, id uuid.UUID) (*model.Task, error)
	UpdateTask(ctx context.Context, id uuid.UUID, column string, value any) error
}

type Cash interface {
	SetTask(ctx context.Context, key string, task *model.Task) error
	SetTasks(ctx context.Context, key string, tasks []model.Task) error
	GetTasks(ctx context.Context, key string) ([]model.Task, error)
	GetTask(ctx context.Context, key string) (*model.Task, error)
}

type Service struct {
	repo Repository
	cash Cash

	timeout time.Duration
}

func (s *Service) CreateTask(
	reqCtx context.Context,
	taskType string,
	desc string,
	decision string,
	rightAnswer string,
	level string,
) error {
	ctx, cancel := context.WithTimeout(reqCtx, s.timeout)
	defer cancel()

	task := model.NewTask(
		taskType,
		desc,
		decision,
		rightAnswer,
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

func (s *Service) IncLike(reqCtx context.Context, id string) error {
	ctx, cancel := context.WithTimeout(reqCtx, s.timeout)
	defer cancel()

	uid, err := uuid.Parse(id)
	if err != nil {
		return ErrBadUID
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

func (s *Service) DecLike(reqCtx context.Context, id string) error {
	ctx, cancel := context.WithTimeout(reqCtx, s.timeout)
	defer cancel()

	uid, err := uuid.Parse(id)
	if err != nil {
		return ErrBadUID
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

func (s *Service) IncDislike(reqCtx context.Context, id string) error {
	ctx, cancel := context.WithTimeout(reqCtx, s.timeout)
	defer cancel()

	uid, err := uuid.Parse(id)
	if err != nil {
		return ErrBadUID
	}

	task, err := s.repo.GetTask(ctx, uid)
	if err != nil {
		return err
	}

	task.Dislikes++

	s.cash.SetTask(ctx, getKeyForOne(task.ID.String()), task)
	if err = s.repo.UpdateTask(ctx, uid, "diclikes", task.Dislikes); err != nil {
		return err
	}

	return nil
}

func (s *Service) DecDislike(reqCtx context.Context, id string) error {
	ctx, cancel := context.WithTimeout(reqCtx, s.timeout)
	defer cancel()

	uid, err := uuid.Parse(id)
	if err != nil {
		return ErrBadUID
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

func (s *Service) GetRandomTask(reqCtx context.Context, taskType string, level uint) (*model.Task, error) {
	ctx, cancel := context.WithTimeout(reqCtx, s.timeout)
	defer cancel()

	cashedTasks, err := s.cash.GetTasks(ctx, getKeyForMany(taskType, level))
	if err == nil && cashedTasks != nil && len(cashedTasks) != 0 {
		task := getRandomFromArr(cashedTasks)

		return &task, nil
	}

	tasks, err := s.repo.GetTasksByTypeAndLevel(ctx, taskType, level)
	if err != nil {
		return nil, err
	}

	s.cash.SetTasks(ctx, getKeyForMany(taskType, level), tasks)

	task := getRandomFromArr(tasks)

	return &task, nil
}

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
		return nil, ErrBadUID
	}

	task, err = s.repo.GetTask(ctx, uid)
	if err != nil {
		return nil, err
	}

	s.cash.SetTask(ctx, getKeyForOne(id), task)

	return nil, err
}

func (s *Service) GetBests(reqCtx context.Context, taskType string, level uint, pageSize, pageIndex uint) ([]model.Task, error) {
	ctx, cancel := context.WithTimeout(reqCtx, s.timeout)
	defer cancel()

	var (
		tasks []model.Task
		err   error
	)

	tasks, err = s.cash.GetTasks(ctx, getKeyForBests(taskType, level))
	if err == nil {
		start := pageSize * (pageIndex - 1)
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

func getKeyForMany(taskType string, level uint) string {
	return fmt.Sprintf("%s:%d", taskType, level)
}

func getKeyForOne(key string) string {
	return fmt.Sprintf("one:%s", key)
}

func getRandomFromArr[T any](arr []T) T {
	randomIndex := rand.Intn(len(arr) - 1)

	return arr[randomIndex]
}

func getKeyForBests(taskType string, level uint) string {
	return fmt.Sprintf("sorted:%s:%d", taskType, level)
}

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
