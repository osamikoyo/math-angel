package service

import (
	"context"
	"fmt"
	"log"
	"time"
	"unicode"

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
	GetUsers(ctx context.Context) ([]model.User, error)
}

type TaskCachedRepo interface {
	CreateTask(ctx context.Context, task *taskmodel.Task) error
	GetTask(ctx context.Context, id uuid.UUID) (*taskmodel.Task, error)
	GetTasks(ctx context.Context, taskType, level string) ([]taskmodel.Task, error)
	Search(ctx context.Context, query string, pageIndex, pageSize int) ([]taskmodel.TaskSearchResult, error)
	UpdateTask(ctx context.Context, uid uuid.UUID, column string, value any) error
	CreateSolution(ctx context.Context, solution *taskmodel.Solution) error
	GetSolution(ctx context.Context, userID uint, taskID string) (*taskmodel.Solution, error)
}

type UserCache interface {
	GetTop(ctx context.Context, key string) ([]model.User, error)
	SetTop(ctx context.Context, key string, top []model.User) error
}

type Service struct {
	userRepo       UserRepository
	userCache      UserCache
	taskCachedRepo TaskCachedRepo

	config  *config.Config
	timeout time.Duration
}

func NewService(userrepo UserRepository, taskCachedRepo TaskCachedRepo, config *config.Config, timeout time.Duration) *Service {
	return &Service{
		userRepo:       userrepo,
		taskCachedRepo: taskCachedRepo,
		config:         config,
		timeout:        timeout,
	}
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

func (s *Service) TaskSolved(reqCtx context.Context, userID uint, taskID uuid.UUID) error {
	ctx, cancel := context.WithTimeout(reqCtx, s.timeout)
	defer cancel()

	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}

	task, err := s.taskCachedRepo.GetTask(ctx, taskID)
	if err != nil {
		return err
	}

	task.Level = firstToLower(task.Level)

	column := ""
	value := uint(0)

	switch task.Level {
	case "easy":
		column = "easy_tasks_solved"
		value = user.Profile.EasyTasksSolved + 1
	case "medium":
		column = "medium_tasks_solved"
		value = user.Profile.MediumTasksSolved + 1
	case "hard":
		column = "hard_tasks_solved"
		value = user.Profile.HardTasksSolved + 1
	default:
		log.Print("invalid level: ", task.Type)

		return errors.ErrInvalidLevel
	}

	if err = s.userRepo.UpdateProfile(ctx, userID, column, value); err != nil {
		return err
	}

	solution := taskmodel.NewSolution(taskID.String(), userID)

	err = s.taskCachedRepo.CreateSolution(ctx, solution)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) GetProfile(reqCtx context.Context, userID uint) (*model.Profile, error) {
	ctx, cancel := context.WithTimeout(reqCtx, s.timeout)
	defer cancel()

	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return user.Profile, nil
}

type userWithPoints struct {
	user   *model.User
	points int
}

func getUserWithPoints(user *model.User) userWithPoints {
	points := 0

	points += int(user.Profile.EasyTasksSolved)
	points += int(user.Profile.MediumTasksSolved) * 2
	points += int(user.Profile.HardTasksSolved) * 3

	return userWithPoints{
		points: points,
		user:   user,
	}
}

func (s *Service) GetUserTop(reqCtx context.Context, pageIndex, pageSize uint) ([]model.User, error) {
	ctx, cancel := context.WithTimeout(reqCtx, s.timeout)
	defer cancel()

	top, err := s.userCache.GetTop(ctx, "bests_top")
	if err == nil {
		start := pageSize * (pageIndex - 1)

		return top[start:min(int(start+pageSize), len(top))], nil
	}

	top, err = s.userRepo.GetUsers(ctx)
	if err != nil {
		return nil, err
	}

	userWithPoints := make([]userWithPoints, len(top))
	for i, us := range top {
		userWithPoints[i] = getUserWithPoints(&us)
	}

	userWithPoints = sortUserWithPoins(userWithPoints)

	users := make([]model.User, len(top))
	for i, us := range userWithPoints {
		users[i] = *us.user
	}

	s.userCache.SetTop(ctx, "top", users)

	start := pageSize * (pageIndex - 1)

	return users[start:min(int(start+pageSize), len(users))], nil
}

func sortUserWithPoins(users []userWithPoints) []userWithPoints {
	if len(users) <= 1 {
		return users
	}

	pivot := users[len(users)/2].points
	left := []userWithPoints{}
	right := []userWithPoints{}
	middle := []userWithPoints{}

	for _, x := range users {
		if x.points < pivot {
			left = append(left, x)
		} else if x.points == pivot {
			middle = append(middle, x)
		} else {
			right = append(right, x)
		}
	}

	left = sortUserWithPoins(left)
	right = sortUserWithPoins(right)

	return append(append(left, middle...), right...)
}

func firstToLower(s string) string {
	if s == "" {
		return s
	}
	r := []rune(s)
	return string(append([]rune{unicode.ToLower(r[0])}, r[1:]...))
}
