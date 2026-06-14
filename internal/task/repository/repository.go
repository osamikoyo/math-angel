package repository

import (
	 
	"github.com/osamikoyo/math-angel/internal/task/cachedrepo"
	"github.com/osamikoyo/math-angel/pkg/logger"
	"gorm.io/gorm"
)

type Repository struct {
	logger *logger.Logger
	db     *gorm.DB
}

var _ cachedrepo.Repository = &Repository{}

func NewRepository(db *gorm.DB, logger *logger.Logger) *Repository {
	return &Repository{
		db:     db,
		logger: logger,
	}
}