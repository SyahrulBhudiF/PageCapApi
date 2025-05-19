package dto

import (
	"errors"
	"fmt"
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/domain/entity"
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/shared/util"
	"github.com/dlclark/regexp2"
	"github.com/google/uuid"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"
)

type UserResponse struct {
	UUID           uuid.UUID  `json:"uuid"`
	Name           string     `json:"name"`
	Email          string     `json:"email"`
	ProfilePicture string     `json:"profile_picture"`
	EmailVerified  *time.Time `json:"email_verified"`
	CreatedAt      *time.Time `json:"created_at"`
	UpdatedAt      *time.Time `json:"updated_at"`
}

func ToUserResponse(user entity.User) *UserResponse {
	return &UserResponse{
		UUID:           user.UUID,
		Name:           user.Name,
		Email:          user.Email,
		ProfilePicture: user.ProfilePicture,
		EmailVerified:  user.EmailVerified,
		CreatedAt:      &user.CreatedAt,
		UpdatedAt:      &user.UpdatedAt,
	}
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required" example:"Pass123!@#"`
	NewPassword string `json:"new_password" binding:"required" example:"1Pass123!@#"`
}

type UpdateUserProfileRequest struct {
	Name           string                `form:"name" binding:"required"`
	ProfilePicture *multipart.FileHeader `form:"profile_picture" binding:"required" swaggertype:"file"`
}

func (r UpdateUserProfileRequest) Validate() error {
	if r.ProfilePicture == nil {
		return errors.New("profile_picture is required")
	}

	ext := strings.ToLower(filepath.Ext(r.ProfilePicture.Filename))
	switch ext {
	case ".jpg", ".jpeg", ".png", ".webp", ".gif":
		return nil
	default:
		return fmt.Errorf("unsupported image format: %s", ext)
	}
}

func (c ChangePasswordRequest) Validate() error {
	if c.OldPassword == c.NewPassword {
		return errors.New("new password must be different from old password")
	}

	re := regexp2.MustCompile(util.PasswordPattern, 0)
	match, err := re.MatchString(c.NewPassword)
	if err != nil {
		return errors.New("failed to validate password")
	}
	if !match {
		return errors.New("password must have at least one lowercase letter, one uppercase letter, one digit, one special character, and be at least 8 characters long")
	}

	return nil
}
