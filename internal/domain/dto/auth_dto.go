package dto

import (
	"errors"
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/shared/util"
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

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type RefreshTokenResponse struct {
	AccessToken string `json:"access_token" binding:"required"`
}

type SendOtpRequest struct {
	Email string `json:"email" binding:"required,email" example:"john@example.com"`
}

type VerifyEmailRequest struct {
	Email string `json:"email" binding:"required" example:"john@example.com"`
	Otp   string `json:"otp" binding:"required"`
}

type GoogleVerifyRequest struct {
	Token string `json:"token" binding:"required"`
}

type ForgotPasswordRequest struct {
	Email    string `json:"email" binding:"required" example:"john@example.com"`
	Otp      string `json:"otp" binding:"required"`
	Password string `json:"password" binding:"required,min=8" example:"Password1!@#"`
}

type SetPasswordRequest struct {
	Password string `json:"password" binding:"required,min=8" example:"Password1!@#"`
	Confirm  string `json:"confirm" binding:"required,min=8" example:"Password1!@#"`
}

type ApiKeyResponse struct {
	ApiKey string `json:"api_key"`
}

func (r RegisterRequest) Validate() error {
	re := regexp2.MustCompile(util.PasswordPattern, 0)
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

func (r ForgotPasswordRequest) Validate() error {
	re := regexp2.MustCompile(util.PasswordPattern, 0)
	match, err := re.MatchString(r.Password)
	if err != nil {
		return errors.New("failed to validate password")
	}
	if !match {
		return errors.New("password must have at least one lowercase letter, one uppercase letter, one digit, one special character, and be at least 8 characters long")
	}
	return nil
}

func (r SetPasswordRequest) Validate() error {
	re := regexp2.MustCompile(util.PasswordPattern, 0)
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
