package repository

import (
	"context"
	"errors"

	selferrors "github.com/osamikoyo/math-angel/internal/errors"
	"github.com/osamikoyo/math-angel/internal/user/model"
	"github.com/osamikoyo/math-angel/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Repository struct {
	logger *logger.Logger
	db     *gorm.DB
}

func NewRepository(db *gorm.DB, logger *logger.Logger) *Repository {
	return &Repository{
		db:     db,
		logger: logger,
	}
}

// CreateUser create new user
func (r *Repository) CreateUser(ctx context.Context, user *model.User) error {
	if user == nil {
		return selferrors.ErrEmptyUser
	}

	r.logger.Info("create user",
		zap.String("username", user.Username),
		zap.String("email", user.Email))

	err := gorm.G[model.User](r.db).Create(ctx, user)
	if err != nil {
		r.logger.Error("failed create user",
			zap.String("username", user.Username),
			zap.Error(err))

		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return selferrors.ErrAlreadyExist
		}

		return selferrors.ErrUnknown
	}

	r.logger.Info("user created successfully",
		zap.Uint("id", user.ID),
		zap.String("username", user.Username))

	return nil
}

// GetUserByID fetch user from db by id
func (r *Repository) GetUserByID(ctx context.Context, id uint) (*model.User, error) {
	r.logger.Info("fetch user by id", zap.Uint("id", id))

	var user model.User
	err := r.db.WithContext(ctx).
		Preload("Profile").
		Where("id = ?", id).
		First(&user).Error

	if err != nil {
		r.logger.Error("failed get user by id",
			zap.Uint("id", id),
			zap.Error(err),
		)

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, selferrors.ErrNotFound
		}

		return nil, selferrors.ErrUnknown
	}

	r.logger.Info("user fetched successfully by id",
		zap.Any("user", user),
	)

	return &user, nil
}

// GetUserByUsername fetch user by username
func (r *Repository) GetUserByUsername(ctx context.Context, username string) (*model.User, error) {
	r.logger.Info("fetch user by username", zap.String("username", username))

	var user model.User
	user, err := gorm.G[model.User](r.db).Where("username = ?", username).First(ctx)
	if err != nil {
		r.logger.Error("failed get user by username",
			zap.String("username", username),
			zap.Error(err))

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, selferrors.ErrNotFound
		}

		return nil, selferrors.ErrUnknown
	}

	r.logger.Info("user fetched successfully by username",
		zap.Uint("id", user.ID),
		zap.String("username", user.Username))

	return &user, nil
}

func (r *Repository) UpdateProfile(ctx context.Context, userID uint, column string, value any) error {
	r.logger.Info("update profile",
		zap.Uint("user_id", userID),
		zap.String("column", column),
	)

	result := r.db.WithContext(ctx).
		Model(&model.Profile{}).
		Where("user_id = ?", userID).
		Update(column, value)

	if result.Error != nil {
		r.logger.Error("failed update profile",
			zap.Uint("user_id", userID),
			zap.String("column", column),
			zap.Error(result.Error),
		)
		return selferrors.ErrUnknown
	}

	if result.RowsAffected == 0 {
		r.logger.Warn("profile not found for update",
			zap.Uint("user_id", userID),
		)
		return selferrors.ErrNotFound
	}

	r.logger.Info("profile updated successfully",
		zap.Uint("user_id", userID),
		zap.String("column", column),
	)

	return nil
}

func (r *Repository) GetUsers(ctx context.Context) ([]model.User, error) {
	r.logger.Info("fetching users")

	users, err := gorm.G[model.User](r.db).
		Preload("Profile", nil).
		Find(ctx)
	if err != nil {
		r.logger.Info("failed fetch users",
			zap.Error(err))
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, selferrors.ErrNotFound
		}
		return nil, selferrors.ErrUnknown
	}

	r.logger.Info("users fetched successfully")
	return users, nil
}
