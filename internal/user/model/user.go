package model

import (
	"time"

	"github.com/osamikoyo/math-angel/internal/errors"
	"github.com/osamikoyo/math-angel/internal/user/crypt"
)

type User struct {
    ID           uint      `json:"id" gorm:"primaryKey"`
    Email        string    `json:"email" gorm:"uniqueIndex"`
    Username     string    `json:"username" gorm:"uniqueIndex"`
    PasswordHash string    `json:"-"`
    CreatedAt    time.Time `json:"created_at"`

    Profile *Profile `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
}


func NewUser(email, username, password string) (*User, error) {
	hashedpassword, err := crypt.HashPassword(password)
	if err != nil {
		return nil, errors.ErrFailedHash
	}

	return &User{
		Email:        email,
		Username:     username,
		PasswordHash: hashedpassword,
		CreatedAt:    time.Now(),
		Profile: &Profile{
			HardTasksSolved: 0,
			MediumTasksSolved: 0,
			EasyTasksSolved: 0,
		},
	}, nil
}
