package model

type TaskSearchResult struct {
	Task
	Snippet string  `gorm:"column:snippet"`
	Rank    float64 `gorm:"column:rank"`
}
