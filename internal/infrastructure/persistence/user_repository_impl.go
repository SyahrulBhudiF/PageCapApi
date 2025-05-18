package persistence

import (
	"context"
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/domain/contract/repository"
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/domain/entity"
	baseRepository "github.com/SyahrulBhudiF/Doc-Management.git/internal/shared/repository"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

var _ repository.UserRepository = (*UserRepositoryImpl)(nil)

type UserRepositoryImpl struct {
	*baseRepository.Repository[entity.User]
}

func NewUserRepository(db *gorm.DB) *UserRepositoryImpl {
	return &UserRepositoryImpl{
		Repository: &baseRepository.Repository[entity.User]{DB: db},
	}
}

func (u *UserRepositoryImpl) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
	var user entity.User
	err := u.DB.WithContext(ctx).Where("email = ?", email).Take(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (u *UserRepositoryImpl) UpdateEmailVerified(ctx context.Context, uuid uuid.UUID) error {
	var user entity.User
	now := time.Now()
	err := u.DB.WithContext(ctx).Model(&user).Where("uuid = ?", uuid).Update("email_verified", &now).Error
	if err != nil {
		return err
	}
	return nil
}
