package repository

import (
	"context"
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/domain/entity"
	_interface "github.com/SyahrulBhudiF/Doc-Management.git/internal/shared/repository/interface"
)

type UserRepository interface {
	_interface.IRepository[entity.User]
	FindByEmail(ctx context.Context, email string) (*entity.User, error)
}
