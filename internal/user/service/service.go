package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/osamikoyo/math-angel/internal/config"
	"github.com/osamikoyo/math-angel/internal/errors"
	"github.com/osamikoyo/math-angel/internal/user/crypt"
	"github.com/osamikoyo/math-angel/internal/user/model"
	"github.com/osamikoyo/math-angel/internal/user/token"

	taskmodel "github.com/osamikoyo/math-angel/internal/task/model"
)

type UserRepository interface {
	CreateUser(ctx context.Context, user *model.User) error
	GetUserByID(ctx context.Context, id uint) (*model.User, error)
	GetUserByUsername(ctx context.Context, username string) (*model.User, error)
	UpdateProfile(ctx context.Context, userID uint, column string, value any) error
}

type TaskCachedRepo interface {
	CreateTask(ctx context.Context, task *taskmodel.Task) error
	GetTask(ctx context.Context, id uuid.UUID) (*taskmodel.Task, error)
	GetTasks(ctx context.Context, taskType, level string) ([]taskmodel.Task, error)
	Search(ctx context.Context, query string, pageIndex, pageSize int) ([]taskmodel.TaskSearchResult, error)
	UpdateTask(ctx context.Context, uid uuid.UUID, column string, value any) error
}
type Service struct {
	userRepo UserRepository
	taskCachedRepo TaskCachedRepo

	config  *config.Config
	timeout time.Duration
}

func (s *Service) AuthenticateUser(reqCtx context.Context, username, password string) (bool, string, error) {
	if len(username) == 0 || len(password) == 0 {
		return false, "", errors.ErrEmptyUsernameOrPassword
	}

	ctx, cancel := context.WithTimeout(reqCtx, s.timeout)
	defer cancel()

	user, err := s.userRepo.GetUserByUsername(ctx, username)
	if err != nil {
		return false, "", err
	}

	if crypt.CheckPasswordHash(password, user.PasswordHash) {
		return true, fmt.Sprintf("%d", user.ID), nil
	}

	return false, "", nil
}

func (s *Service) RegisterUser(reqCtx context.Context, username, email, password string) error {
	ctx, cancel := context.WithTimeout(reqCtx, s.timeout)
	defer cancel()

	user, err := model.NewUser(email, username, password)
	if err != nil {
		return err
	}

	if err = s.userRepo.CreateUser(ctx, user); err != nil {
		return err
	}

	return nil
}

func (s *Service) LoginUser(reqCtx context.Context, username, password string) (string, error) {
	authed, userID, err := s.AuthenticateUser(reqCtx, username, password)
	if err != nil {
		return "", err
	}

	if !authed {
		return "", errors.ErrAuthFailed
	}

	tokenString, err := token.New(userID, s.config)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (s *Service) Refresh(reqCtx context.Context, tokenString string) (string, error) {
	userID, err := token.Validate(tokenString, s.config)
	if err != nil {
		return "", errors.ErrInvalidToken
	}

	newToken, err := token.New(userID, s.config)
	if err != nil {
		return "", err
	}

	return newToken, nil
}

func (s *Service) Validate(reqCtx context.Context, tokenString string) (string, error) {
	id, err := token.Validate(tokenString, s.config)
	if err != nil {
		return "", errors.ErrInvalidToken
	}

	return id, nil
}

func (s *Service) TaskSolved(reqCtx context.Context,userID uint, taskID uuid.UUID) error {
	ctx, cancel := context.WithTimeout(reqCtx, s.timeout)
	defer cancel()

	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil{
		return err
	}

	task, err := s.taskCachedRepo.GetTask(ctx, taskID)
	if err != nil{
		return err
	}

	column := ""
	value := uint(0)

	switch task.Type{
	case "easy":
		column = "easy_tasks_solved"
		value = user.Profile.EasyTasksSolved+1
	case "medium":
		column = "medium_tasks_solved"
		value = user.Profile.MediumTasksSolved+1
	case "hard":
		column = "hard_tasks_solved"
		value = user.Profile.HardTasksSolved+1
	}

	if err = s.userRepo.UpdateProfile(ctx, userID, column, value);err != nil{
		return err
	}

	return nil
}