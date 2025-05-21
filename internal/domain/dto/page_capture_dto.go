package dto

import (
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/domain/entity"
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/infrastructure/rod"
	"github.com/google/uuid"
)

type PageCaptureRequest struct {
	Url          string `json:"url" validate:"required,url"`
	Width        int    `json:"width,omitempty"`
	Height       int    `json:"height,omitempty"`
	FullPage     bool   `json:"fullPage,omitempty"`
	DelaySeconds int    `json:"delaySeconds,omitempty"`
	IsMobile     bool   `json:"isMobile,omitempty"`
}

type PageCaptureResponse struct {
	Filename string `json:"filename"`
	Content  []byte `json:"content"`
}

type PagesCaptureResponse struct {
	Data       []entity.PageCapture `json:"data"`
	Total      int64                `json:"total"`
	Page       int                  `json:"page"`
	Limit      int                  `json:"limit"`
	TotalPages int                  `json:"total_pages"`
}

func ConvertToScreenshotOptions(req *PageCaptureRequest) *rod.ScreenshotOptions {
	return &rod.ScreenshotOptions{
		URL:          req.Url,
		Width:        req.Width,
		Height:       req.Height,
		FullPage:     req.FullPage,
		DelaySeconds: req.DelaySeconds,
		IsMobile:     req.IsMobile,
	}
}

func ConvertRequestToEntity(req PageCaptureRequest, userID uuid.UUID) *entity.PageCapture {
	intPtr := func(i int) *int {
		if i == 0 {
			return nil
		}
		return &i
	}

	return &entity.PageCapture{
		UserID:       userID,
		URL:          req.Url,
		Width:        intPtr(req.Width),
		Height:       intPtr(req.Height),
		FullPage:     req.FullPage,
		DelaySeconds: req.DelaySeconds,
		IsMobile:     req.IsMobile,
	}
}
