package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/osamikoyo/math-angel/internal/errors"
	"github.com/osamikoyo/math-angel/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ====================== MOCKS ======================

type MockRepository struct{ mock.Mock }
type MockCache struct{ mock.Mock }

func (m *MockRepository) CreateTask(ctx context.Context, task *model.Task) error {
	args := m.Called(ctx, task)
	return args.Error(0)
}

func (m *MockRepository) GetTasksByTypeAndLevel(ctx context.Context, taskType, level string) ([]model.Task, error) {
	args := m.Called(ctx, taskType, level)
	return args.Get(0).([]model.Task), args.Error(1)
}

func (m *MockRepository) GetTask(ctx context.Context, id uuid.UUID) (*model.Task, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Task), args.Error(1)
}

func (m *MockRepository) UpdateTask(ctx context.Context, id uuid.UUID, column string, value any) error {
	args := m.Called(ctx, id, column, value)
	return args.Error(0)
}

func (m *MockRepository) SearchTasks(ctx context.Context, query string, limit, offset int) ([]model.TaskSearchResult, error) {
	args := m.Called(ctx, query, limit, offset)
	return args.Get(0).([]model.TaskSearchResult), args.Error(1)
}

// Cache mocks
func (m *MockCache) SetTask(ctx context.Context, key string, task *model.Task) error {
	args := m.Called(ctx, key, task)
	return args.Error(0)
}

func (m *MockCache) SetTasks(ctx context.Context, key string, tasks []model.Task) error {
	args := m.Called(ctx, key, tasks)
	return args.Error(0)
}

func (m *MockCache) GetTasks(ctx context.Context, key string) ([]model.Task, error) {
	args := m.Called(ctx, key)
	return args.Get(0).([]model.Task), args.Error(1)
}

func (m *MockCache) GetTask(ctx context.Context, key string) (*model.Task, error) {
	args := m.Called(ctx, key)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Task), args.Error(1)
}

// ====================== HELPERS ======================

func TestNewService(t *testing.T) {
	repo := new(MockRepository)
	cash := new(MockCache)
	svc := NewService(repo, cash, 5*time.Second)

	assert.NotNil(t, svc)
	assert.Equal(t, repo, svc.repo)
	assert.Equal(t, cash, svc.cash)
	assert.Equal(t, 5*time.Second, svc.timeout)
}

// ====================== CREATE TASK ======================

func TestCreateTask_Success(t *testing.T) {
	repo := new(MockRepository)
	cash := new(MockCache)
	svc := NewService(repo, cash, time.Second)

	taskType, problem, solution, boxed, level := "algebra", "2+2=?", "4", "4", "easy"

	cash.On("SetTask", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("*model.Task")).Return(nil)
	repo.On("CreateTask", mock.Anything, mock.AnythingOfType("*model.Task")).Return(nil)

	err := svc.CreateTask(context.Background(), taskType, problem, solution, boxed, level)
	assert.NoError(t, err)

	cash.AssertExpectations(t)
	repo.AssertExpectations(t)
}

// ====================== SEARCH ======================

func TestSearch_EmptyQuery(t *testing.T) {
	svc := NewService(new(MockRepository), new(MockCache), time.Second)

	_, err := svc.Search(context.Background(), "", 1, 20)
	assert.ErrorIs(t, err, errors.ErrEmptyQuery)
}

func TestSearch_Success(t *testing.T) {
	repo := new(MockRepository)
	svc := NewService(repo, new(MockCache), time.Second)

	expected := []model.TaskSearchResult{}
	repo.On("SearchTasks", mock.Anything, "test", 10, 0).Return(expected, nil)

	result, err := svc.Search(context.Background(), "test", 1, 10)
	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

// ====================== LIKE / DISLIKE ======================

func TestIncLike_Success(t *testing.T) {
	repo := new(MockRepository)
	cash := new(MockCache)
	svc := NewService(repo, cash, time.Second)

	id := uuid.New()
	task := &model.Task{UID: id.String(), Likes: 5}

	repo.On("GetTask", mock.Anything, id).Return(task, nil)
	cash.On("SetTask", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Return(nil)
	repo.On("UpdateTask", mock.Anything, id, "likes", 6).Return(nil)

	err := svc.IncLike(context.Background(), id.String())
	assert.NoError(t, err)
	assert.Equal(t, 6, task.Likes)
}

func TestDecLike_Success(t *testing.T) {
	repo := new(MockRepository)
	cash := new(MockCache)
	svc := NewService(repo, cash, time.Second)

	id := uuid.New()
	task := &model.Task{UID: id.String(), Likes: 5}

	repo.On("GetTask", mock.Anything, id).Return(task, nil)
	cash.On("SetTask", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Return(nil)
	repo.On("UpdateTask", mock.Anything, id, "likes", 4).Return(nil)

	err := svc.DecLike(context.Background(), id.String())
	assert.NoError(t, err)
	assert.Equal(t, 4, task.Likes)
}

func TestIncDislike_Success(t *testing.T) {
	repo := new(MockRepository)
	cash := new(MockCache)
	svc := NewService(repo, cash, time.Second)

	id := uuid.New()
	task := &model.Task{UID: id.String(), Dislikes: 5}

	repo.On("GetTask", mock.Anything, id).Return(task, nil)
	cash.On("SetTask", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Return(nil)
	repo.On("UpdateTask", mock.Anything, id, "dislikes", 6).Return(nil)

	err := svc.IncDislike(context.Background(), id.String())
	assert.NoError(t, err)
	assert.Equal(t, 6, task.Dislikes)
}

func TestLike_BadUUID(t *testing.T) {
	svc := NewService(new(MockRepository), new(MockCache), time.Second)
	err := svc.IncLike(context.Background(), "invalid-uuid")
	assert.ErrorIs(t, err, errors.ErrBadUID)
}

// ====================== GET RANDOM TASK ======================

func TestGetRandomTask_CacheHit(t *testing.T) {
	cash := new(MockCache)
	svc := NewService(new(MockRepository), cash, time.Second)

	tasks := []model.Task{
		{UID: uuid.New().String(), Problem: "Task 1"},
		{UID: uuid.New().String(), Problem: "Task 2"},
	}

	cash.On("GetTasks", mock.Anything, "algebra:easy").Return(tasks, nil)

	task, err := svc.GetRandomTask(context.Background(), "algebra", "easy")
	assert.NoError(t, err)
	assert.Contains(t, []string{"Task 1", "Task 2"}, task.Problem)
}

func TestGetRandomTask_CacheMiss_RepoSuccess(t *testing.T) {
	repo := new(MockRepository)
	cash := new(MockCache)
	svc := NewService(repo, cash, time.Second)

	tasks := []model.Task{{UID: uuid.New().String()}}
	repo.On("GetTasksByTypeAndLevel", mock.Anything, "algebra", "easy").Return(tasks, nil)
	cash.On("GetTasks", mock.Anything, mock.Anything).Return(nil, assert.AnError)
	cash.On("SetTasks", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	_, err := svc.GetRandomTask(context.Background(), "algebra", "easy")
	assert.NoError(t, err)
}

// ====================== GET TASK ======================

func TestGetTask_CacheHit(t *testing.T) {
	cash := new(MockCache)
	svc := NewService(new(MockRepository), cash, time.Second)

	expected := &model.Task{UID: uuid.New().String()}
	cash.On("GetTask", mock.Anything, mock.AnythingOfType("string")).Return(expected, nil)

	task, err := svc.GetTask(context.Background(), expected.UID)
	assert.NoError(t, err)
	assert.Equal(t, expected, task)
}

func TestGetTask_CacheMiss_RepoSuccess(t *testing.T) {
	repo := new(MockRepository)
	cash := new(MockCache)
	svc := NewService(repo, cash, time.Second)

	id := uuid.New()
	task := &model.Task{UID: id.String()}

	cash.On("GetTask", mock.Anything, mock.Anything).Return(nil, assert.AnError)
	repo.On("GetTask", mock.Anything, id).Return(task, nil)
	cash.On("SetTask", mock.Anything, mock.Anything, task).Return(nil)

	result, err := svc.GetTask(context.Background(), id.String())
	assert.NoError(t, err)
	assert.Equal(t, task, result)
}

// ====================== GET BESTS ======================

func TestGetBests_CacheHit(t *testing.T) {
	cash := new(MockCache)
	svc := NewService(new(MockRepository), cash, time.Second)

	tasks := []model.Task{
		{Likes: 100}, {Likes: 50}, {Likes: 10},
	}

	cash.On("GetTasks", mock.Anything, "sorted:algebra:easy").Return(tasks, nil)

	result, err := svc.GetBests(context.Background(), "algebra", "easy", 10, 1)
	assert.NoError(t, err)
	assert.Equal(t, tasks[:2], result) // pageSize=10, но всего 3
}

func TestGetBests_CacheMiss_SortsAndCaches(t *testing.T) {
	repo := new(MockRepository)
	cash := new(MockCache)
	svc := NewService(repo, cash, time.Second)

	tasks := []model.Task{
		{Likes: 5}, {Likes: 15}, {Likes: 10},
	}

	cash.On("GetTasks", mock.Anything, "sorted:algebra:easy").Return(nil, assert.AnError)
	cash.On("GetTasks", mock.Anything, "algebra:easy").Return(nil, assert.AnError)
	repo.On("GetTasksByTypeAndLevel", mock.Anything, "algebra", "easy").Return(tasks, nil)
	cash.On("SetTasks", mock.Anything, "algebra:easy", mock.Anything).Return(nil)

	result, err := svc.GetBests(context.Background(), "algebra", "easy", 10, 1)
	assert.NoError(t, err)
	assert.Len(t, result, 3)
	assert.Equal(t, 15, result[0].Likes)
}

// ====================== RUN ALL ======================

func TestMain(m *testing.M) {
	m.Run()
}
