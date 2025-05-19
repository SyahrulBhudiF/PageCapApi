package usecase

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/domain/contract/redis"
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/domain/contract/repository"
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/domain/dto"
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/domain/entity"
	errorEntity "github.com/SyahrulBhudiF/Doc-Management.git/internal/domain/error"
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/shared/util"
	"github.com/SyahrulBhudiF/Doc-Management.git/pkg/config"
	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"io"
	"strings"
	tm "time"
)

type UserUseCase struct {
	repo  repository.UserRepository
	redis redis.Service
	cloud *cloudinary.Cloudinary
	cfg   *config.Config
}

func NewUserUseCase(repo repository.UserRepository, redis redis.Service, cfg *config.Config, cloud *cloudinary.Cloudinary) *UserUseCase {
	return &UserUseCase{
		repo:  repo,
		redis: redis,
		cloud: cloud,
		cfg:   cfg,
	}
}

func (c *UserUseCase) ChangePassword(d *dto.ChangePasswordRequest, e *entity.User, context context.Context) error {
	if e.Password == "" {
		logrus.WithError(errorEntity.ErrPasswordNotSet).Error("password is not set")
		return errorEntity.ErrPasswordNotSet
	}

	if response := util.ComparePassword(e.Password, d.OldPassword, c.cfg.Server.Salt); !response {
		logrus.Info("password not match")
		return errorEntity.ErrInvalidPassword
	}

	hashedPassword := util.HashPassword(d.NewPassword, c.cfg.Server.Salt)
	e.Password = hashedPassword
	if err := c.repo.Update(context, e); err != nil {
		logrus.WithError(err).Error("failed to update password")
		return err
	}

	if err := c.redis.Delete(fmt.Sprintf("user:%s", e.UUID)); err != nil {
		logrus.WithError(err).Error("failed to delete user cache")
		return err
	}

	accessExpire, _ := tm.ParseDuration(c.cfg.Jwt.AccessTokenExpire)
	jsonUser, _ := json.Marshal(e)
	if err := c.redis.Set(fmt.Sprintf("user:%s", e.UUID), jsonUser, accessExpire); err != nil {
		logrus.WithError(err).Error("failed to set user cache")
		return err
	}

	logrus.Info("password updated successfully")
	return nil
}

func (c *UserUseCase) UpdateUserProfile(d *dto.UpdateUserProfileRequest, e *entity.User, ctx context.Context) error {
	if strings.TrimSpace(d.Name) != "" {
		e.Name = strings.TrimSpace(d.Name)
		logrus.Info("name updated successfully")
	}

	if d.ProfilePicture != nil {
		if d.ProfilePicture.Size > 5*1024*1024 {
			return errorEntity.ErrImageTooLarge
		}

		file, err := d.ProfilePicture.Open()
		if err != nil {
			return fmt.Errorf("%w: %v", errorEntity.ErrCloudinaryUpload, err)
		}
		defer file.Close()

		buffer := util.GetBuffer()
		defer util.PutBuffer(buffer)
		if _, err := io.Copy(buffer, file); err != nil {
			return fmt.Errorf("%w: %v", errorEntity.ErrCloudinaryUpload, err)
		}

		params := util.UploadParamsPool.Get().(*uploader.UploadParams)
		defer util.UploadParamsPool.Put(params)

		publicID := fmt.Sprintf("profile_pictures/%s", uuid.NewString())
		overwrite := true
		params.PublicID = publicID
		params.Overwrite = &overwrite
		params.ResourceType = "image"
		params.Transformation = "w_500,h_500,c_limit,q_auto"

		uploadResult, err := c.cloud.Upload.Upload(ctx, bytes.NewReader(buffer.Bytes()), *params)
		if err != nil {
			logrus.Error("cloudinary upload failed")
			return fmt.Errorf("%w: %v", errorEntity.ErrCloudinaryUpload, err)
		}

		oldPublicID := e.PublicId
		e.ProfilePicture = uploadResult.SecureURL
		e.PublicId = publicID

		if oldPublicID != "" {
			go func(pubID string) {
				deleteCtx, cancel := context.WithTimeout(context.Background(), 30*tm.Second)
				defer cancel()

				_, err := c.cloud.Upload.Destroy(deleteCtx, uploader.DestroyParams{
					PublicID: pubID,
				})
				if err != nil {
					logrus.WithError(err).Error("failed to delete old profile picture")
				}
			}(oldPublicID)
		}

		logrus.Info("cloudinary upload successful")
	}

	if err := c.repo.Update(ctx, e); err != nil {
		logrus.WithFields(logrus.Fields{
			"user_id": e.UUID,
			"error":   err.Error(),
		}).Error("database update failed")
		return err
	}

	jsonUser, err := json.Marshal(e)
	if err != nil {
		logrus.WithError(err).Error("failed to marshal user data")
		return err
	}

	expire, _ := tm.ParseDuration(c.cfg.Jwt.AccessTokenExpire)
	if err := c.redis.Set(fmt.Sprintf("user:%s", e.UUID), jsonUser, expire); err != nil {
		logrus.WithError(err).Error("failed to set user cache")
		return err
	}

	return nil
}
