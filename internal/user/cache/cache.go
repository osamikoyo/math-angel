package cache

import (
	"context"
	"encoding/json"
	"time"

	"errors"
	selferrors "github.com/osamikoyo/math-angel/internal/errors"
	"github.com/osamikoyo/math-angel/internal/user/model"
	"github.com/osamikoyo/math-angel/pkg/logger"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

var (
	ErrEmptyUser          = errors.New("empty user")
	ErrFailedMarshal      = errors.New("json marshal fail")
	ErrInternalCacheError = errors.New("internal cache error")
	ErrFailedDecode       = errors.New("decode cache data fail")
)

type Cache struct {
	client     *redis.Client
	logger     *logger.Logger
	defaultExp time.Duration
}

func NewCache(client *redis.Client, logger *logger.Logger, defaultExp time.Duration) *Cache {
	return &Cache{
		client:     client,
		logger:     logger,
		defaultExp: defaultExp,
	}
}

// SetUser save user to cache
func (c *Cache) SetUser(ctx context.Context, key string, user *model.User) error {
	if user == nil {
		return ErrEmptyUser
	}

	c.logger.Info("set user to cache",
		zap.String("key", key),
		zap.Uint("id", user.ID),
		zap.String("username", user.Username))

	data, err := json.Marshal(user)
	if err != nil {
		c.logger.Error("failed marshal user",
			zap.Any("user", user),
			zap.Error(err))

		return ErrFailedMarshal
	}

	res := c.client.Set(ctx, key, data, c.defaultExp)
	if err := res.Err(); err != nil {
		c.logger.Error("failed set user to cache",
			zap.String("key", key),
			zap.Error(err))

		return ErrInternalCacheError
	}

	return nil
}

// GetUserByID fetch user by key
func (c *Cache) GetUserByID(ctx context.Context, key string) (*model.User, error) {
	c.logger.Info("fetch user from cache", zap.String("key", key))

	data, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			c.logger.Warn("user not found in cache", zap.String("key", key))
			return nil, selferrors.ErrNotFound
		}

		c.logger.Warn("failed get user from cache",
			zap.String("key", key),
			zap.Error(err))

		return nil, ErrInternalCacheError
	}

	var user model.User
	if err := json.Unmarshal([]byte(data), &user); err != nil {
		c.logger.Error("failed unmarshal user from cache",
			zap.String("key", key),
			zap.String("data", data),
			zap.Error(err))

		return nil, ErrFailedDecode
	}

	return &user, nil
}
