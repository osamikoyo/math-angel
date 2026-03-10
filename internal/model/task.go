package model

import "github.com/google/uuid"

type Task struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Type        string    `json:"type"`
	Descrition  string    `gorm:"unique" json:"description"`
	Decision    string    `json:"decision"`
	Level       string    `json:"level"`
	RightAnswer string    `json:"right_answer"`
	Likes       uint      `json:"likes"`
	Dislikes    uint      `json:"dislikes"`
}

func NewTask(taskType, desc, decision, rightAnswer string, level string) *Task {
	return &Task{
		ID:          uuid.New(),
		Type:        taskType,
		Decision:    decision,
		Descrition:  desc,
		RightAnswer: rightAnswer,
		Level:       level,
	}
}
