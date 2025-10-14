package entity

import (
	"time"

	"github.com/SyahrulBhudiF/Doc-Management.git/internal/shared/entity"
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

func ToUserResponse(user *User) *User {
	return &User{
		Entity: entity.Entity{
			UUID:      user.UUID,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		},
		Email:          user.Email,
		Name:           user.Name,
		ProfilePicture: user.ProfilePicture,
		PublicId:       user.PublicId,
		EmailVerified:  user.EmailVerified,
	}
}

func (u *User) TableName() string {
	return "users"
}
