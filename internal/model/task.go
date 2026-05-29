package model

import (
	"github.com/google/uuid"
	"github.com/osamikoyo/math-angel/internal/errors"
)

type Task struct {
	ID       int    `gorm:"primaryKey;autoIncrement" json:"id"`
	UID      string `gorm:"type:text;uniqueIndex"`
	Type     string `gorm:"type:text" json:"type"`
	Problem  string `gorm:"type:text;uniqueIndex" json:"problem"`
	Solution string `gorm:"type:text" json:"solution"`
	Level    string `gorm:"type:text" json:"level"`
	Boxed    string `gorm:"type:text" json:"boxed"`
	Likes    uint   `json:"likes"`
	Dislikes uint   `json:"dislikes"`
}

func NewTask(taskType, problem, solution, boxed, level string) *Task {
	return &Task{
		UID:      uuid.New().String(),
		Type:     taskType,
		Solution: solution,
		Problem:  problem,
		Boxed:    boxed,
		Level:    level,
	}
}

func (t *Task) Validate() error {
	if len(t.UID) == 0 {
		return errors.ErrEmptyUID
	}

	err := uuid.Validate(t.UID)
	if err != nil {
		return errors.ErrInvalidUID
	}

	if len(t.Problem) == 0 {
		return errors.ErrEmptyProblem
	}

	if t.Level != "easy" || t.Level != "medium" || t.Level != "hard" {
		return errors.ErrInvalidLevel
	}
}
