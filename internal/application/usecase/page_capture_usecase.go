package usecase

import (
	"bytes"
	"context"
	"fmt"
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/domain/contract/redis"
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/domain/contract/repository"
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/domain/dto"
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/domain/entity"
	errorEntity "github.com/SyahrulBhudiF/Doc-Management.git/internal/domain/error"
	rodService "github.com/SyahrulBhudiF/Doc-Management.git/internal/infrastructure/rod"
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/shared/util"
	"github.com/SyahrulBhudiF/Doc-Management.git/pkg/config"
	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"github.com/go-rod/rod"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"io"
)

type PageCaptureUseCase struct {
	repo            repository.PageCaptureRepository
	redis           redis.Service
	cloud           *cloudinary.Cloudinary
	cfg             *config.Config
	browserInstance *rod.Browser
}

func NewPageCaptureUseCase(repo repository.PageCaptureRepository, redis redis.Service, cfg *config.Config, cloud *cloudinary.Cloudinary, browser *rod.Browser) *PageCaptureUseCase {
	return &PageCaptureUseCase{
		repo:            repo,
		redis:           redis,
		cloud:           cloud,
		cfg:             cfg,
		browserInstance: browser,
	}
}

func (c *PageCaptureUseCase) PageCapture(body *dto.PageCaptureRequest, key string, ctx context.Context) (*dto.PageCaptureResponse, error) {
	redisKey := fmt.Sprintf("api_key:%s", key)
	cachedKey, err := c.redis.Get(redisKey)
	if err != nil {
		logrus.Error("failed to get redis key: ", err)
		return nil, err
	}

	if cachedKey == "" {
		logrus.Error("invalid api key")
		return nil, errorEntity.ErrInvalidCredentials
	}

	user, err := c.repo.GetUser(cachedKey)
	if err != nil {
		return nil, errorEntity.ErrUserNotFound
	}

	req := dto.ConvertToScreenshotOptions(body)
	data, err := rodService.CaptureScreenshot(ctx, c.browserInstance, *req)
	if err != nil {
		logrus.Error("failed to capture screenshot: ", err)
		return nil, err
	}

	reader := bytes.NewReader(data)
	buffer := util.GetBuffer()
	defer util.PutBuffer(buffer)
	if _, err := io.Copy(buffer, reader); err != nil {
		return nil, err
	}

	resp := &dto.PageCaptureResponse{
		Filename: "screenshot.png",
		Content:  buffer.Bytes(),
	}

	go func(dataCopy []byte, reqCopy dto.PageCaptureRequest, userUUID uuid.UUID) {
		params := util.UploadParamsPool.Get().(*uploader.UploadParams)
		defer util.UploadParamsPool.Put(params)

		publicID := fmt.Sprintf("capture/%s", uuid.NewString())
		overwrite := true
		params.PublicID = publicID
		params.Overwrite = &overwrite
		params.ResourceType = "image"

		uploadResult, err := c.cloud.Upload.Upload(context.Background(), bytes.NewReader(dataCopy), *params)
		if err != nil {
			logrus.Error("cloudinary upload failed: ", err)
			return
		}

		pageCapture := dto.ConvertRequestToEntity(reqCopy, userUUID)
		pageCapture.PublicId = publicID
		pageCapture.ImagePath = uploadResult.SecureURL

		if err := c.repo.Create(context.Background(), pageCapture); err != nil {
			logrus.WithFields(logrus.Fields{
				"user_id": userUUID,
				"error":   err.Error(),
			}).Error("database create failed")
			return
		}

		logrus.Info("Upload and DB insert completed successfully")
	}(buffer.Bytes(), *body, user.UUID)

	return resp, nil
}

func (c *PageCaptureUseCase) GetPageCapture(e *entity.User, search string, orderBy string, sort string, page int, fullPage *bool, isMobile *bool) (*dto.PagesCaptureResponse, error) {
	data, err := c.repo.GetPageCaptureByUserID(e.UUID.String(), search, orderBy, sort, page, fullPage, isMobile)
	if err != nil {
		return nil, err
	}

	if len(data.Data) == 0 {
		return nil, errorEntity.ErrDataNotFound
	}

	return data, nil
}
