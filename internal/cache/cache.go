package cache

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/osamikoyo/math-angel/internal/cachedrepo"
	"github.com/osamikoyo/math-angel/internal/model"
	"github.com/osamikoyo/math-angel/pkg/logger"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

var (
	ErrEmptyTask          = errors.New("empty task")
	ErrFailedMarshal      = errors.New("json marshal fail")
	ErrInternalCacheError = errors.New("internal cash error")
	ErrFailedDecode       = errors.New("decode cash data fail")
)

type Cache struct {
	client     *redis.Client
	logger     *logger.Logger
	defaultExp time.Duration
}

var _ cachedrepo.Cache = &Cache{}

func NewCache(client *redis.Client, logger *logger.Logger, defaultExp time.Duration) *Cache {
	return &Cache{
		client:     client,
		logger:     logger,
		defaultExp: defaultExp,
	}
}

func (c *Cache) SetTask(ctx context.Context, key string, task *model.Task) error {
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

		return ErrInternalCacheError
	}

	return nil
}

func (c *Cache) SetTasks(ctx context.Context, key string, tasks []model.Task) error {
	if tasks == nil {
		return ErrEmptyTask
	}

	c.logger.Info("set tasks",
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

		return ErrInternalCacheError
	}

	return nil
}

func (c *Cache) GetTask(ctx context.Context, key string) (*model.Task, error) {
	c.logger.Info("fetch task",
		zap.String("key", key))

	data, err := c.client.Get(ctx, key).Result()
	if err != nil {
		c.logger.Warn("failed get task from cash",
			zap.String("key", key),
			zap.Error(err))

		return nil, ErrInternalCacheError
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

func (c *Cache) GetTasks(ctx context.Context, key string) ([]model.Task, error) {
	c.logger.Info("fetch tasks",
		zap.String("key", key))

	data, err := c.client.Get(ctx, key).Result()
	if err != nil {
		c.logger.Warn("failed get tasks from cash",
			zap.String("key", key),
			zap.Error(err))

		return nil, ErrInternalCacheError
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

func (c *Cache) SetSearchResults(ctx context.Context, key string, trs []model.TaskSearchResult) error {
	if len(trs) == 0 {
		return ErrEmptyTask
	}

	c.logger.Info("set search results",
		zap.String("key", key))

	data, err := json.Marshal(trs)
	if err != nil {
		c.logger.Error("failed marshal",
			zap.Error(err))

		return ErrFailedMarshal
	}

	err = c.client.Set(ctx, key, data, c.defaultExp).Err()
	if err != nil {
		c.logger.Error("set value failed",
			zap.String("key", key),
			zap.ByteString("value", data))

		return ErrInternalCacheError
	}

	return nil
}

func (c *Cache) GetSearchResults(ctx context.Context, key string) ([]model.TaskSearchResult, error) {
	c.logger.Info("fetch search results",
		zap.String("key", key))

	data, err := c.client.Get(ctx, key).Result()
	if err != nil {
		c.logger.Error("failed fetch",
			zap.String("key", key),
			zap.Error(err))

		return nil, ErrInternalCacheError
	}

	var tsr []model.TaskSearchResult

	if err = json.Unmarshal([]byte(data), &tsr); err != nil {
		c.logger.Error("failed unmarshal data",
			zap.String("data", data),
			zap.Error(err))

		return nil, ErrFailedDecode
	}

	return tsr, nil
}
