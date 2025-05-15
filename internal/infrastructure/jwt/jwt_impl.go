package jwt

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	jwtContract "github.com/SyahrulBhudiF/Doc-Management.git/internal/domain/contract/jwt"
	errorEntity "github.com/SyahrulBhudiF/Doc-Management.git/internal/domain/error"
	"github.com/SyahrulBhudiF/Doc-Management.git/pkg/config"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"time"
)

type jwtService struct {
	accessSecret  string
	refreshSecret string
	accessExpire  time.Duration
	refreshExpire time.Duration
}

var _ jwtContract.Service = (*jwtService)(nil)

func NewJwtService(cfg *config.Config) jwtContract.Service {
	accessDuration, err := time.ParseDuration(cfg.Jwt.AccessTokenExpire)
	if err != nil {
		panic(fmt.Errorf("failed to parse access token duration: %w", err))
	}

	refreshDuration, err := time.ParseDuration(cfg.Jwt.RefreshTokenExpire)
	if err != nil {
		panic(fmt.Errorf("failed to parse refresh token duration: %w", err))
	}

	return &jwtService{
		accessSecret:  cfg.Jwt.AccessTokenSecret,
		refreshSecret: cfg.Jwt.RefreshTokenSecret,
		accessExpire:  accessDuration,
		refreshExpire: refreshDuration,
	}
}

func (j *jwtService) GenerateToken(userID uuid.UUID, username string) (accessToken string, refreshToken string, refreshTokenHash string, err error) {
	access, err := j.generateToken(userID, username, j.accessExpire, j.accessSecret)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to generate access token: %w", err)
	}

	refresh, err := j.generateToken(userID, username, j.refreshExpire, j.refreshSecret)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	refreshTokenHash, err = j.HashToken(refresh)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to hash refresh token: %w", err)
	}

	return access, refresh, refreshTokenHash, nil
}

func (j *jwtService) ValidateToken(tokenString string, key string) (*jwtContract.UserClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwtContract.UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return key, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, errorEntity.ErrTokenExpired
		}
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if claims, ok := token.Claims.(*jwtContract.UserClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errorEntity.ErrInvalidToken
}

func (j *jwtService) HashToken(token string) (string, error) {
	hasher := sha256.New()
	_, err := hasher.Write([]byte(token))
	if err != nil {
		return "", fmt.Errorf("failed to write token to hasher: %w", err)
	}

	hashBytes := hasher.Sum(nil)
	return hex.EncodeToString(hashBytes), nil
}

func (j *jwtService) CompareTokenHash(token, hash string) error {
	incomingHash, err := j.HashToken(token)
	if err != nil {
		return fmt.Errorf("failed to hash incoming token for comparison: %w", err)
	}

	storedHashBytes, err := hex.DecodeString(hash)
	if err != nil {
		return fmt.Errorf("failed to decode stored hex hash: %w", err)
	}
	incomingHashBytes, _ := hex.DecodeString(incomingHash)

	if subtle.ConstantTimeCompare(incomingHashBytes, storedHashBytes) == 1 {
		return nil
	}

	return errorEntity.ErrTokenMismatch
}

func (j *jwtService) generateToken(userID uuid.UUID, username string, duration time.Duration, key string) (string, error) {
	expirationTime := time.Now().Add(duration)
	claims := &jwtContract.UserClaims{
		UUID:     userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "doc-management",
			Subject:   fmt.Sprintf("%d", userID),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(key)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}
