package error

import (
	"errors"
	"fmt"
)

// Authentication errors
var (
	ErrInvalidCredentials     = errors.New("invalid credentials")
	ErrInvalidPhoneNumber     = errors.New("invalid phone number")
	ErrInvalidPassword        = errors.New("invalid password")
	ErrInvalidEmail           = errors.New("invalid email")
	ErrInvalidUserID          = errors.New("invalid user ID")
	ErrInvalidUser            = errors.New("invalid user")
	ErrEmailAlreadyExists     = errors.New("email already registered")
	ErrUserAlreadyHasPassword = errors.New("user already has a password")
	ErrUserAlreadyExists      = errors.New("user already exists")
	ErrUserNotFound           = errors.New("user not found")
	ErrAuthHeaderNotFound     = errors.New("authorization header not found")
	ErrOtpNotFound            = errors.New("otp not found")
	ErrInvalidOtp             = errors.New("invalid otp")
	ErrLimitExceeded          = errors.New("limit exceeded")
	ErrEmailNotVerified       = errors.New("email not verified")
	ErrPasswordNotSet         = errors.New("password not set")
)

// Request validation errors
var (
	ErrRequestBodyRequired = errors.New("invalid request: request body is required")
	ErrInvalidRequestBody  = errors.New("invalid request body")
	ErrInvalidRequest      = fmt.Errorf("invalid request")
)

// Token Error
var (
	ErrTokenNotFound           = errors.New("token not found")
	ErrInvalidToken            = errors.New("invalid token")
	ErrTokenExpired            = errors.New("token has expired")
	ErrTokenMismatch           = errors.New("token hash mismatch")
	ErrTokenAlreadyBlacklisted = errors.New("token already blacklisted")
)

var (
	ErrCloudinaryUpload = fmt.Errorf("cloudinary upload error")
	ErrImageTooLarge    = fmt.Errorf("image too large")
	ErrDataNotFound     = fmt.Errorf("data not found")
)
