package repository

import (
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/domain/dto"
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/domain/entity"
	_interface "github.com/SyahrulBhudiF/Doc-Management.git/internal/shared/repository/interface"
)

type PageCaptureRepository interface {
	_interface.IRepository[entity.PageCapture]
	GetUser(userID string) (*entity.User, error)
	GetPageCaptureByUserID(userID string, search string, orderBy string, sort string, page int, fullPage *bool, isMobile *bool) (*dto.PagesCaptureResponse, error)
}
