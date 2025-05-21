package persistence

import (
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/domain/contract/repository"
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/domain/dto"
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

func (p *PageCaptureImpl) GetPageCaptureByUserID(
	userID string,
	search string,
	orderBy string,
	sort string,
	page int,
	fullPage *bool,
	isMobile *bool,
) (*dto.PagesCaptureResponse, error) {

	var captures []entity.PageCapture
	var total int64

	limit := 10
	offset := (page - 1) * limit

	if sort != "asc" && sort != "desc" {
		sort = "desc"
	}

	if orderBy == "" {
		orderBy = "created_at"
	}

	query := p.DB.Model(&entity.PageCapture{}).Where("user_id = ?", userID)

	if search != "" {
		query = query.Where(`
			url ILIKE ? OR 
			image_path ILIKE ? OR 
			public_id ILIKE ?
		`, "%"+search+"%", "%"+search+"%", "%"+search+"%")
	}

	if fullPage != nil {
		query = query.Where("full_page = ?", *fullPage)
	}

	if isMobile != nil {
		query = query.Where("is_mobile = ?", *isMobile)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	err := query.
		Order(orderBy + " " + sort).
		Limit(limit).
		Offset(offset).
		Find(&captures).Error

	if err != nil {
		return nil, err
	}

	totalPages := int((total + int64(limit) - 1) / int64(limit))

	result := &dto.PagesCaptureResponse{
		Data:       captures,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}

	return result, nil
}
