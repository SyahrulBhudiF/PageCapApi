package dto

import (
	"errors"
	"github.com/dlclark/regexp2"
)

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email" example:"john@example.com"`
	Name     string `json:"name" binding:"required" example:"John Doe"`
	Password string `json:"password" binding:"required,min=8" example:"Pass123!@#"`
	Confirm  string `json:"confirm" binding:"required,min=8" example:"Pass123!@#"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email" example:"john@example.com"`
	Password string `json:"password" binding:"required,min=8" example:"Pass123!@#"`
}

type LoginResponse struct {
	AccessToken  string `json:"access_token" binding:"required"`
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

var passwordPattern = `^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)(?=.*[@$!%*?&#])[A-Za-z\d@$!%*?&#]{8,}$`

func (r RegisterRequest) Validate() error {
	re := regexp2.MustCompile(passwordPattern, 0)
	match, err := re.MatchString(r.Password)
	if err != nil {
		return errors.New("failed to validate password")
	}
	if !match {
		return errors.New("password must have at least one lowercase letter, one uppercase letter, one digit, one special character, and be at least 8 characters long")
	}
	if r.Password != r.Confirm {
		return errors.New("confirm must match password")
	}
	return nil
}
