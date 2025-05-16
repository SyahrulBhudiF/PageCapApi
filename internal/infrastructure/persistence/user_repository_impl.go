package persistence

import (
	"context"
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/domain/contract/repository"
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/domain/entity"
	baseRepository "github.com/SyahrulBhudiF/Doc-Management.git/internal/shared/repository"
	"gorm.io/gorm"
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

func (u UserRepositoryImpl) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
	var user entity.User
	err := u.DB.WithContext(ctx).Where("email = ?", email).Take(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}
