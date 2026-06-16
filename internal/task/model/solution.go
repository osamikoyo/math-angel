package model

import (
	"time"
)

type Solution struct {
	TaskID string `gorm:"primaryKey"`
	UserID uint
	CreatedAt time.Time
}

func NewSolution(taskID string, userID uint) *Solution {
	return &Solution{
		TaskID: taskID,
		UserID: userID,
		CreatedAt: time.Now(),
	}
}