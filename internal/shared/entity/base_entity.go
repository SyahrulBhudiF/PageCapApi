package entity

import (
	"time"

	"github.com/google/uuid"
)

type Entity struct {
	UUID      uuid.UUID `gorm:"primaryKey;type:uuid;default:uuid_generate_v4();index" json:"uuid"`
	CreatedAt time.Time `gorm:"default:current_timestamp" json:"created_at"`
	UpdatedAt time.Time `gorm:"default:current_timestamp" json:"updated_at"`
}
