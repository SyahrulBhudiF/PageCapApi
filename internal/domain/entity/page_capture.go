package entity

import (
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/shared/entity"
	"github.com/google/uuid"
)

type PageCapture struct {
	entity.Entity
	UserID       uuid.UUID `json:"user_id" gorm:"not null"`
	URL          string    `json:"url"`
	ImagePath    string    `json:"image_path"`
	PublicId     string    `json:"public_id"`
	Width        *int      `json:"width,omitempty"`
	Height       *int      `json:"height,omitempty"`
	FullPage     bool      `json:"full_page"`
	DelaySeconds int       `json:"delay_seconds"`
	IsMobile     bool      `json:"is_mobile"`
}

func NewPageCapture(userID uuid.UUID, url string, imagePath string, publicId string, width *int, height *int, fullPage bool, delaySeconds int, IsMobile bool) (*PageCapture, error) {
	return &PageCapture{
		UserID:       userID,
		URL:          url,
		ImagePath:    imagePath,
		PublicId:     publicId,
		Width:        width,
		Height:       height,
		FullPage:     fullPage,
		DelaySeconds: delaySeconds,
		IsMobile:     IsMobile,
	}, nil
}

func (s *PageCapture) TableName() string {
	return "page_captures"
}
