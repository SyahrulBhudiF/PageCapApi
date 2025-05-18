package midleware

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/domain/contract/jwt"
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/domain/contract/redis"
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/domain/contract/repository"
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/domain/entity"
	errorEntity "github.com/SyahrulBhudiF/Doc-Management.git/internal/domain/error"
	entity2 "github.com/SyahrulBhudiF/Doc-Management.git/internal/shared/entity"
	"github.com/SyahrulBhudiF/Doc-Management.git/pkg/config"
	"github.com/SyahrulBhudiF/Doc-Management.git/pkg/response"
	"github.com/gin-gonic/gin"
	"strings"
)

type AuthMiddleware struct {
	user  repository.UserRepository
	redis redis.Service
	jwt   jwt.Service
	cfg   *config.Config
}

func NewAuthMiddleware(user repository.UserRepository, redis redis.Service, jwt jwt.Service, cfg *config.Config) *AuthMiddleware {
	return &AuthMiddleware{
		user:  user,
		redis: redis,
		jwt:   jwt,
		cfg:   cfg,
	}
}

func (m *AuthMiddleware) EnsureAuthenticated() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, "unauthorized", errorEntity.ErrAuthHeaderNotFound)
		}

		parts := strings.Split(authHeader, "Bearer ")
		if len(parts) != 2 {
			response.Unauthorized(c, "unauthorized", errorEntity.ErrTokenNotFound)
		}

		token := strings.TrimSpace(parts[1])
		claims, err := m.jwt.ValidateToken(token, m.cfg.Jwt.AccessTokenSecret)
		if err != nil {
			switch {
			case errors.Is(err, errorEntity.ErrTokenExpired):
				response.Unauthorized(c, "unauthorized", errorEntity.ErrTokenExpired)
			case errors.Is(err, errorEntity.ErrInvalidToken):
				response.Unauthorized(c, "unauthorized", errorEntity.ErrInvalidToken)
			default:
				response.InternalServerError(c, err)
			}
		}

		cachedUser, err := m.redis.Get(fmt.Sprintf("user:%s", claims.UUID))
		if err == nil && cachedUser != "" {
			var user entity.User
			if err := json.Unmarshal([]byte(cachedUser), &user); err == nil {
				c.Set("accessToken", token)
				c.Set("user", &user)
				c.Next()
			}
		}

		user := entity.User{
			Entity: entity2.Entity{
				UUID: claims.UUID,
			},
		}

		if err := m.user.Find(c, &user); err != nil {
			response.Unauthorized(c, "unauthorized", errorEntity.ErrUserNotFound)
		}

		jsonUser, _ := json.Marshal(user)
		if err := m.redis.Set(fmt.Sprintf("user:%s", claims.UUID), jsonUser, 0); err != nil {
			response.InternalServerError(c, err)
		}

		c.Set("accessToken", token)
		c.Set("user", &user)

		c.Next()
	}
}
