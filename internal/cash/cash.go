package cash

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/osamikoyo/math-angel/internal/model"
	"github.com/osamikoyo/math-angel/internal/service"
	"github.com/osamikoyo/math-angel/pkg/logger"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

var (
	ErrEmptyTask         = errors.New("empty task")
	ErrFailedMarshal     = errors.New("json marshal fail")
	ErrInternalCashError = errors.New("internal cash error")
	ErrFailedDecode      = errors.New("decode cash data fail")
)

type Cash struct {
	client     *redis.Client
	logger     *logger.Logger
	defaultExp time.Duration
}

var _ service.Cash = &Cash{}

func NewCash(client *redis.Client, logger *logger.Logger, defaultExp time.Duration) *Cash {
	return &Cash{
		client:     client,
		logger:     logger,
		defaultExp: defaultExp,
	}
}

func (c *Cash) SetTask(ctx context.Context, key string, task *model.Task) error {
	if task == nil {
		return ErrEmptyTask
	}

	c.logger.Info("set task",
		zap.String("key", key),
		zap.Any("task", task))

	data, err := json.Marshal(task)
	if err != nil {
		c.logger.Error("failed marshal task`",
			zap.Any("task", task))

		return ErrFailedMarshal
	}

	res := c.client.Set(ctx, key, data, c.defaultExp)
	if err := res.Err(); err != nil {
		c.logger.Error("failed set task",
			zap.String("key", key),
			zap.ByteString("data", data),
			zap.Error(err))

		return ErrInternalCashError
	}

	return nil
}

func (c *Cash) SetTasks(ctx context.Context, key string, tasks []model.Task) error {
	if tasks == nil {
		return ErrEmptyTask
	}

	c.logger.Error("set tasks",
		zap.String("key", key))

	data, err := json.Marshal(tasks)
	if err != nil {
		c.logger.Error("failed marshal tasks",
			zap.Error(err))

		return ErrFailedMarshal
	}

	res := c.client.Set(ctx, key, data, c.defaultExp)
	if err := res.Err(); err != nil {
		c.logger.Error("failed set tasks",
			zap.String("key", key),
			zap.ByteString("data", data),
			zap.Error(err))

		return ErrInternalCashError
	}

	return nil
}

func (c *Cash) GetTask(ctx context.Context, key string) (*model.Task, error) {
	c.logger.Info("fetch task",
		zap.String("key", key))

	data, err := c.client.Get(ctx, key).Result()
	if err != nil {
		c.logger.Warn("failed get task from cash",
			zap.String("key", key),
			zap.Error(err))

		return nil, ErrInternalCashError
	}

	var task model.Task

	if err := json.Unmarshal([]byte(data), &task); err != nil {
		c.logger.Error("failed unmarshal data from cash",
			zap.String("data", data),
			zap.Error(err))

		return nil, ErrFailedDecode
	}

	return &task, nil
}

func (c *Cash) GetTasks(ctx context.Context, key string) ([]model.Task, error) {
	c.logger.Info("fetch tasks",
		zap.String("key", key))

	data, err := c.client.Get(ctx, key).Result()
	if err != nil {
		c.logger.Warn("failed get tasks from cash",
			zap.String("key", key),
			zap.Error(err))

		return nil, ErrInternalCashError
	}

	var tasks []model.Task

	if err := json.Unmarshal([]byte(data), &tasks); err != nil {
		c.logger.Error("failed unmarshal data from cash",
			zap.String(data, data),
			zap.Error(err))

		return nil, ErrFailedDecode
	}

	return tasks, nil
}
