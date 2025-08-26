package midleware

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/SyahrulBhudiF/Doc-Management.git/internal/domain/contract/jwt"
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/domain/contract/redis"
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/domain/contract/repository"
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/domain/entity"
	errorEntity "github.com/SyahrulBhudiF/Doc-Management.git/internal/domain/error"
	entity2 "github.com/SyahrulBhudiF/Doc-Management.git/internal/shared/entity"
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/shared/util"
	"github.com/SyahrulBhudiF/Doc-Management.git/pkg/config"
	"github.com/SyahrulBhudiF/Doc-Management.git/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
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
			logrus.Warn("no auth header")
			response.Unauthorized(c, "unauthorized", errorEntity.ErrAuthHeaderNotFound)
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, "Bearer ")
		if len(parts) != 2 {
			logrus.Warn("invalid auth header")
			response.Unauthorized(c, "unauthorized", errorEntity.ErrTokenNotFound)
			c.Abort()
			return
		}

		token := strings.TrimSpace(parts[1])

		if existingToken, err := m.redis.Get(fmt.Sprintf("blacklist:%s", token)); err == nil && existingToken != "" {
			logrus.Warn("token is blacklisted")
			response.Unauthorized(c, "unauthorized", errorEntity.ErrTokenAlreadyBlacklisted)
			c.Abort()
			return
		}

		claims, err := m.jwt.ValidateToken(token, m.cfg.Jwt.AccessTokenSecret)
		if err != nil {
			if util.ErrorInList(err, errorEntity.ErrTokenExpired, errorEntity.ErrInvalidToken) {
				logrus.Warn("invalid token: ", err)
				response.Unauthorized(c, "unauthorized", err)
				c.Abort()
				return
			} else {
				logrus.Error("failed to validate token: ", err)
				response.InternalServerError(c, err)
				c.Abort()
				return
			}
		}

		cachedUser, err := m.redis.Get(fmt.Sprintf("user:%s", claims.UUID))
		if err == nil && cachedUser != "" {
			var user entity.User
			if err := json.Unmarshal([]byte(cachedUser), &user); err == nil {
				c.Set("accessToken", token)
				c.Set("user", &user)
				c.Next()
				return
			}
		}

		user := entity.User{
			Entity: entity2.Entity{
				UUID: claims.UUID,
			},
		}

		if err := m.user.Find(c, &user); err != nil {
			logrus.Warn("failed to find user: ", err)
			response.Unauthorized(c, "unauthorized", errorEntity.ErrUserNotFound)
			c.Abort()
			return
		}

		jsonUser, _ := json.Marshal(user)
		if err := m.redis.Set(fmt.Sprintf("user:%s", claims.UUID), jsonUser, 0); err != nil {
			logrus.Error("failed to set user: ", err)
			response.InternalServerError(c, err)
			c.Abort()
			return
		}

		c.Set("accessToken", &token)
		c.Set("user", &user)

		c.Next()
	}
}
