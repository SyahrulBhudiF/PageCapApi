package persistence

import (
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/domain/contract/repository"
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/domain/entity"
	baseRepository "github.com/SyahrulBhudiF/Doc-Management.git/internal/shared/repository"
	"gorm.io/gorm"
)

type PageCaptureImpl struct {
	*baseRepository.Repository[entity.PageCapture]
}

var _ repository.PageCaptureRepository = (*PageCaptureImpl)(nil)

func NewPageCaptureRepository(db *gorm.DB) *PageCaptureImpl {
	return &PageCaptureImpl{
		Repository: &baseRepository.Repository[entity.PageCapture]{DB: db},
	}
}

func (p *PageCaptureImpl) GetUser(userID string) (*entity.User, error) {
	var user entity.User
	err := p.DB.Where("uuid = ?", userID).Take(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}
