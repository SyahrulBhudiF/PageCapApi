package jwt

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type UserClaims struct {
	UUID     uuid.UUID `json:"uuid"`
	Username string    `json:"username"`
	jwt.RegisteredClaims
}

type Service interface {
	GenerateToken(userID uuid.UUID, username string) (accessToken string, refreshToken string, refreshTokenHash string, err error)
	ValidateToken(tokenString string, key string) (*UserClaims, error)
	HashToken(token string) (string, error)
	CompareTokenHash(token, hash string) error
}
