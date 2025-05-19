package entity

import (
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/shared/entity"
	"time"
)

type User struct {
	entity.Entity
	Email          string     `json:"email" gorm:"unique;not null"`
	Password       string     `json:"password" gorm:"not null"`
	Name           string     `json:"name" gorm:"not null"`
	ProfilePicture string     `json:"profile_picture"`
	PublicId       string     `json:"public_id"`
	EmailVerified  *time.Time `json:"email_verified"`
}

func NewUser(email string, password string, name string, profilePicture string) (*User, error) {
	return &User{
		Email:          email,
		Password:       password,
		Name:           name,
		ProfilePicture: profilePicture,
	}, nil
}

func (u *User) TableName() string {
	return "users"
}
