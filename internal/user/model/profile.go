package model

type Profile struct {
	ID                uint `gorm:"primaryKey;autoIncrement"`
	HardTasksSolved   uint
	MediumTasksSolved uint
	EasyTasksSolved   uint
	TotalSolved       uint `gorm:"-"`

	UserID uint
}
