package service

import (
	"fmt"
	"math/rand"

	"github.com/google/uuid"
	"github.com/osamikoyo/math-angel/internal/model"
)

type Repository interface {
	CreateTask(task *model.Task) error
	GetTasksByTypeAndLevel(taskType string, level uint) ([]model.Task, error)
	GetTask(id uuid.UUID) (*model.Task, error)
	UpdateTask(id uuid.UUID, update *model.Task) error
}

type Cash interface {
	SetTask(task *model.Task) error
	SetTasks(key string, tasks []model.Task) error
	GetTasks(key string) ([]model.Task, error)
	GetTask(id uuid.UUID) (*model.Task, error)
}

type Service struct {
	repo Repository
	cash Cash
}

func (s *Service) CreateTask(
	taskType string,
	desc string,
	decision string,
	rightAnswer string,
	level uint,
) error {
	task := model.NewTask(
		taskType,
		desc,
		decision,
		rightAnswer,
		level,
	)

	if err := s.cash.SetTask(task); err != nil {
		return err
	}

	if err := s.repo.CreateTask(task); err != nil {
		return err
	}

	return nil
}

func (s *Service) GetRandomTask(taskType string, level uint) (*model.Task, error) {
	cashedTasks, err := s.cash.GetTasks(getKey(taskType, level))
	if err == nil && cashedTasks != nil && len(cashedTasks) != 0 {
		task := getRandomFromArr(cashedTasks)

		return &task, nil
	}

	tasks, err := s.repo.GetTasksByTypeAndLevel(taskType, level)
	if err != nil {
		return nil, err
	}

	s.cash.SetTasks(getKey(taskType, level), tasks)

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
