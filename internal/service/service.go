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
	SetTask(ctx context.Context, task *model.Task) error
	SetTasks(ctx context.Context, key string, tasks []model.Task) error
	GetTasks(ctx context.Context, key string) ([]model.Task, error)
	GetTask(ctx context.Context, id uuid.UUID) (*model.Task, error)
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

	if err := s.cash.SetTask(ctx, task); err != nil {
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

	s.cash.SetTask(ctx, task)
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

	s.cash.SetTask(ctx, task)
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

	s.cash.SetTask(ctx, task)
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

	s.cash.SetTask(ctx, task)
	if err = s.repo.UpdateTask(ctx, uid, "dislikes", task.Dislikes); err != nil {
		return err
	}

	return nil
}

func (s *Service) GetRandomTask(reqCtx context.Context, taskType string, level uint) (*model.Task, error) {
	ctx, cancel := context.WithTimeout(reqCtx, s.timeout)
	defer cancel()

	cashedTasks, err := s.cash.GetTasks(ctx, getKey(taskType, level))
	if err == nil && cashedTasks != nil && len(cashedTasks) != 0 {
		task := getRandomFromArr(cashedTasks)

		return &task, nil
	}

	tasks, err := s.repo.GetTasksByTypeAndLevel(ctx, taskType, level)
	if err != nil {
		return nil, err
	}

	s.cash.SetTasks(ctx, getKey(taskType, level), tasks)

	task := getRandomFromArr(tasks)

	return &task, nil
}

func getKey(taskType string, level uint) string {
	return fmt.Sprintf("%s:%d", taskType, level)
}

func getRandomFromArr[T any](arr []T) T {
	randomIndex := rand.Intn(len(arr) - 1)

	return arr[randomIndex]
}
