package jwt

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"time"
)

type UserClaims struct {
	UUID  uuid.UUID `json:"uuid"`
	Email string    `json:"email"`
	jwt.RegisteredClaims
}

type Service interface {
	GenerateToken(userID uuid.UUID, email string) (accessToken string, refreshToken string, refreshTokenHash string, err error)
	ValidateToken(tokenString string, key string) (*UserClaims, error)
	HashToken(token string) (string, error)
	CompareTokenHash(token, hash string) error
	GenerateSingleToken(userID uuid.UUID, email string, expire time.Duration, secret string) (string, error)
}
