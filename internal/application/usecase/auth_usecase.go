package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/domain/contract/jwt"
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/domain/contract/redis"
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/domain/contract/repository"
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/domain/dto"
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/domain/entity"
	errorEntity "github.com/SyahrulBhudiF/Doc-Management.git/internal/domain/error"
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/infrastructure/mail"
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/shared/util"
	"github.com/SyahrulBhudiF/Doc-Management.git/pkg/config"
	"github.com/google/uuid"
	"github.com/markbates/goth"
	"github.com/sirupsen/logrus"
	"google.golang.org/api/idtoken"
	"sync"
	tm "time"
)

type AuthUseCase struct {
	repo  repository.UserRepository
	jwt   jwt.Service
	redis redis.Service
	mail  *mail.Service
	cfg   *config.Config
}

func NewAuthUseCase(repo repository.UserRepository, jwt jwt.Service, redis redis.Service, mail *mail.Service, cfg *config.Config) *AuthUseCase {
	return &AuthUseCase{
		repo:  repo,
		jwt:   jwt,
		redis: redis,
		mail:  mail,
		cfg:   cfg,
	}
}

func (a *AuthUseCase) Register(req *dto.RegisterRequest, ctx context.Context) (*entity.User, error) {
	existingUser, _ := a.repo.FindByEmail(ctx, req.Email)

	if existingUser != nil {
		logrus.Error("User already exists")
		return nil, errorEntity.ErrEmailAlreadyExists
	}

	hashedPassword := util.HashPassword(req.Password, a.cfg.Server.Salt)

	newUser, err := entity.NewUser(
		req.Email,
		hashedPassword,
		req.Name,
		"",
	)

	if err != nil {
		logrus.Error("Failed to create new user")
		return nil, err
	}

	err = a.repo.Create(ctx, newUser)
	if err != nil {
		logrus.Error("Failed to create new user in database")
		return nil, err
	}

	logrus.Info("User created successfully")
	return newUser, nil
}

func (a *AuthUseCase) Login(req *dto.LoginRequest, ctx context.Context) (*dto.LoginResponse, error) {
	existingUser, err := a.repo.FindByEmail(ctx, req.Email)
	if err != nil {
		logrus.Error("User not found")
		return nil, errorEntity.ErrUserNotFound
	}

	if existingUser.EmailVerified == nil {
		logrus.Error("Email not verified")
		return nil, errorEntity.ErrEmailNotVerified
	}

	if !util.ComparePassword(existingUser.Password, req.Password, a.cfg.Server.Salt) {
		logrus.Error("Invalid password")
		return nil, errorEntity.ErrInvalidPassword
	}

	acc, refresh, refreshHash, err := a.jwt.GenerateToken(existingUser.UUID, existingUser.Email)
	if err != nil {
		logrus.Error("Failed to generate token")
		return nil, err
	}

	accessExpire, _ := tm.ParseDuration(a.cfg.Jwt.AccessTokenExpire)
	refreshExpire, _ := tm.ParseDuration(a.cfg.Jwt.RefreshTokenExpire)

	jsonUser, err := json.Marshal(existingUser)
	if err != nil {
		logrus.Error("Failed to marshal user data")
		return nil, err
	}

	err = a.redis.Set(fmt.Sprintf("user:%s", existingUser.UUID.String()), jsonUser, accessExpire)
	if err != nil {
		logrus.Error("Failed to set user in Redis:", err)
		return nil, err
	}

	err = a.redis.Set(fmt.Sprintf("user_refresh:%s", existingUser.UUID.String()), refreshHash, refreshExpire)
	if err != nil {
		logrus.Error("Failed to set user_refresh in Redis:", err)
		return nil, err
	}

	logrus.Info("User logged in successfully")
	return &dto.LoginResponse{
		AccessToken:  acc,
		RefreshToken: refresh,
	}, nil
}

func (a *AuthUseCase) Logout(req *dto.LogoutRequest, user *entity.User, accessToken string) error {
	claims, err := a.jwt.ValidateToken(req.RefreshToken, a.cfg.Jwt.RefreshTokenSecret)
	if err != nil {
		logrus.Error("Invalid refresh token")
		return errorEntity.ErrInvalidToken
	}

	if claims.UUID != user.UUID {
		logrus.Error("Invalid user")
		return errorEntity.ErrInvalidUser
	}

	err = a.redis.Delete(fmt.Sprintf("user:%s", user.UUID))
	if err != nil {
		logrus.Error("Failed to delete access token from Redis")
		return err
	}

	err = a.redis.Delete(fmt.Sprintf("user_refresh:%s", user.UUID))
	if err != nil {
		logrus.Error("Failed to delete refresh token from Redis")
		return err
	}

	isBlacklisted, err := a.redis.Exists(fmt.Sprintf("blacklist:%s", req.RefreshToken))
	if err != nil {
		logrus.Error("Failed to check if token is blacklisted")
		return err
	}

	if isBlacklisted {
		logrus.Error("Token is already blacklisted")
		return errorEntity.ErrTokenAlreadyBlacklisted
	}

	expireDuration, _ := tm.ParseDuration(a.cfg.Jwt.RefreshTokenExpire)
	err = a.redis.Set(fmt.Sprintf("blacklist:%s", req.RefreshToken), "blacklisted", expireDuration)
	if err != nil {
		logrus.Error("Failed to set blacklist refresh token:", err)
		return err
	}

	expireDuration, _ = tm.ParseDuration(a.cfg.Jwt.AccessTokenExpire)
	err = a.redis.Set(fmt.Sprintf("blacklist:%s", accessToken), "blacklisted", expireDuration)
	if err != nil {
		logrus.Error("Failed to set blacklist access token:", err)
		return err
	}

	logrus.Info("User logged out successfully")

	return nil
}

func (a *AuthUseCase) RefreshToken(req *dto.RefreshTokenRequest, user *entity.User) (*dto.RefreshTokenResponse, error) {
	isBlacklisted, err := a.redis.Exists(fmt.Sprintf("blacklist:%s", req.RefreshToken))
	if err != nil {
		logrus.Error("Failed to check if token is blacklisted")
		return nil, err
	}

	if isBlacklisted {
		logrus.Error("Token is blacklisted")
		return nil, errorEntity.ErrTokenAlreadyBlacklisted
	}

	claims, err := a.jwt.ValidateToken(req.RefreshToken, a.cfg.Jwt.RefreshTokenSecret)
	if err != nil {
		logrus.Error("Invalid refresh token")
		return nil, errorEntity.ErrInvalidToken
	}

	if claims.UUID != user.UUID {
		logrus.Error("Invalid user")
		return nil, errorEntity.ErrInvalidUser
	}

	hashRefresh, err := a.redis.Get(fmt.Sprintf("user_refresh:%s", user.UUID))
	if err != nil {
		logrus.Error("Failed to get refresh token from Redis")
		return nil, err
	}

	err = a.jwt.CompareTokenHash(req.RefreshToken, hashRefresh)
	if err != nil {
		logrus.Error("Invalid refresh token")
		return nil, errorEntity.ErrInvalidToken
	}

	time, _ := tm.ParseDuration(a.cfg.Jwt.AccessTokenExpire)
	token, err := a.jwt.GenerateSingleToken(claims.UUID, claims.Email, time, a.cfg.Jwt.AccessTokenSecret)
	if err != nil {
		logrus.Error("Failed to generate new access token")
		return nil, err
	}

	return &dto.RefreshTokenResponse{AccessToken: token}, nil
}

func (a *AuthUseCase) SendOtp(body *dto.SendOtpRequest, ctx context.Context) error {
	existingUser, err := a.repo.FindByEmail(ctx, body.Email)
	if err != nil {
		logrus.Error("User not found")
		return errorEntity.ErrUserNotFound
	}

	limitKey := fmt.Sprintf("otp_limit:%s", existingUser.UUID)

	count, err := a.redis.Incr(limitKey)
	if err != nil {
		logrus.WithError(err).Error("Failed to increment OTP limit key")
		return fmt.Errorf("internal error")
	}

	if count == 1 {
		err = a.redis.Expire(limitKey, 5*tm.Minute)
		if err != nil {
			logrus.WithError(err).Error("Failed to set expiry on OTP limit key")
			return fmt.Errorf("internal error")
		}
	}

	if count > 5 {
		logrus.Warn("OTP request limit exceeded")
		return errorEntity.ErrLimitExceeded
	}

	otp := util.GenerateOTP()

	go func(email, otp string) {
		err := a.mail.SendMail(email, "OTP Verification", fmt.Sprintf("Your OTP is: %s, This will expired after 5 minutes", otp))
		if err != nil {
			logrus.WithError(err).Error("Failed to send OTP email")
		}
	}(existingUser.Email, otp)

	err = a.redis.Set(fmt.Sprintf("otp:%s", existingUser.UUID), otp, 5*tm.Minute)
	if err != nil {
		logrus.Error("Failed to set OTP in Redis")
		return err
	}

	logrus.Info("OTP sent successfully")
	return nil
}

func (a *AuthUseCase) VerifyEmail(req *dto.VerifyEmailRequest, ctx context.Context) error {
	existingUser, err := a.repo.FindByEmail(ctx, req.Email)
	if err != nil {
		logrus.Error("User not found")
		return errorEntity.ErrUserNotFound
	}

	otp, err := a.redis.Get(fmt.Sprintf("otp:%s", existingUser.UUID))
	if err != nil {
		logrus.Error("Failed to get OTP from Redis")
		return errorEntity.ErrOtpNotFound
	}

	if otp != req.Otp {
		logrus.Error("Invalid OTP")
		return errorEntity.ErrInvalidOtp
	}

	err = a.repo.UpdateEmailVerified(ctx, existingUser.UUID)
	if err != nil {
		logrus.Error("Failed to update email verified status")
		return err
	}

	if err := a.redis.Delete(fmt.Sprintf("otp:%s", existingUser.UUID)); err != nil {
		logrus.Error("Failed to delete otp key:", err)
		return err
	}

	if err := a.redis.Delete(fmt.Sprintf("otp_limit:%s", existingUser.UUID)); err != nil {
		logrus.Error("Failed to delete otp_limit key:", err)
		return err
	}

	logrus.Info("Email verified successfully")
	return nil
}

func (a *AuthUseCase) ForgotPassword(req *dto.ForgotPasswordRequest, ctx context.Context) error {
	existingUser, err := a.repo.FindByEmail(ctx, req.Email)
	if err != nil {
		logrus.Error("User not found")
		return errorEntity.ErrUserNotFound
	}

	otp, err := a.redis.Get(fmt.Sprintf("otp:%s", existingUser.UUID))
	if err != nil {
		logrus.Error("Failed to get OTP from Redis")
		return errorEntity.ErrOtpNotFound
	}

	if otp != req.Otp {
		logrus.Error("Invalid OTP")
		return errorEntity.ErrInvalidOtp
	}

	ctx, errHandler := util.NewErrorHandler(ctx)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		hashedPassword := util.HashPassword(req.Password, a.cfg.Server.Salt)
		existingUser.Password = hashedPassword
		existingUser.UpdatedAt = tm.Now()

		if err := a.repo.Update(ctx, existingUser); err != nil {
			logrus.Error("Failed to update user password")
			errHandler.SetError(err)
		}
		jsonUser, err := json.Marshal(existingUser)
		if err != nil {
			logrus.Error("Failed to marshal user data")
			errHandler.SetError(err)
			return
		}

		if err := a.redis.Set(fmt.Sprintf("user:%s", existingUser.UUID), jsonUser, 0); err != nil {
			logrus.Error("Failed to set user data in Redis:", err)
			errHandler.SetError(err)
		}
	}()

	go func() {
		defer wg.Done()
		if err := a.redis.Delete(fmt.Sprintf("otp:%s", existingUser.UUID)); err != nil {
			logrus.Error("Failed to delete otp key from Redis:", err)
			errHandler.SetError(err)
			return
		}
		if err := a.redis.Delete(fmt.Sprintf("otp_limit:%s", existingUser.UUID)); err != nil {
			logrus.Error("Failed to delete otp_limit key from Redis:", err)
			errHandler.SetError(err)
		}
	}()

	wg.Wait()

	if err := errHandler.Err(); err != nil {
		return err
	}

	logrus.Info("Password updated successfully")
	return nil
}

func (a *AuthUseCase) GoogleLogin(g *goth.User, ctx context.Context) (*dto.LoginResponse, error) {
	now := tm.Now()

	user, err := a.repo.FindByEmail(ctx, g.Email)
	if err != nil {
		user = nil
	}

	if user == nil {
		user = &entity.User{
			Email:         g.Email,
			Name:          g.Name,
			EmailVerified: &now,
		}
		err := a.repo.Create(ctx, user)
		if err != nil {
			logrus.Error("Failed to create new user from Google login:", err)
			return nil, err
		}

		logrus.Info("New user created from Google login")
	} else if user.EmailVerified == nil {
		err := a.repo.UpdateEmailVerified(ctx, user.UUID)
		if err != nil {
			logrus.Error("Failed to update email verified status:", err)
			return nil, errorEntity.ErrEmailNotVerified
		}
		user.EmailVerified = &now
	}

	if user.UUID == uuid.Nil {
		logrus.Error("User UUID is empty")
		return nil, errorEntity.ErrUserNotFound
	}

	acc, refresh, refreshHash, err := a.jwt.GenerateToken(user.UUID, user.Email)
	if err != nil {
		logrus.Error("Failed to generate JWT token:", err)
		return nil, err
	}

	loginReq := &dto.LoginResponse{
		AccessToken:  acc,
		RefreshToken: refresh,
	}

	accessExpire, _ := tm.ParseDuration(a.cfg.Jwt.AccessTokenExpire)
	refreshExpire, _ := tm.ParseDuration(a.cfg.Jwt.RefreshTokenExpire)
	jsonUser, err := json.Marshal(user)
	if err != nil {
		logrus.Error("Failed to marshal user data")
		return nil, err
	}

	err = a.redis.Set(fmt.Sprintf("user:%s", user.UUID.String()), jsonUser, accessExpire)
	if err != nil {
		logrus.Error("Failed to set user data in Redis:", err)
		return nil, err
	}
	logrus.Info("Set user data in Redis")

	err = a.redis.Set(fmt.Sprintf("user_refresh:%s", user.UUID.String()), refreshHash, refreshExpire)
	if err != nil {
		logrus.Error("Failed to set refresh token in Redis:", err)
		return nil, err
	}
	logrus.Info("Set refresh token in Redis")

	return loginReq, nil
}

func (a *AuthUseCase) SetPassword(req *dto.SetPasswordRequest, e *entity.User, ctx context.Context) error {
	existingUser, err := a.repo.FindByEmail(ctx, e.Email)
	if err != nil {
		logrus.Error("User not found")
		return errorEntity.ErrUserNotFound
	}

	if existingUser.Password != "" {
		logrus.Error("User already has a password")
		return errorEntity.ErrUserAlreadyHasPassword
	}

	hashedPassword := util.HashPassword(req.Password, a.cfg.Server.Salt)
	existingUser.Password = hashedPassword
	existingUser.UpdatedAt = tm.Now()

	err = a.repo.Update(ctx, existingUser)
	if err != nil {
		logrus.Error("Failed to update user password")
		return err
	}

	jsonUser, _ := json.Marshal(existingUser)
	err = a.redis.Set(fmt.Sprintf("user:%s", existingUser.UUID), jsonUser, 0)
	if err != nil {
		logrus.Error("Failed to set user data in Redis:", err)
		return err
	}
	logrus.Info("Password updated successfully")
	return nil
}

func (a *AuthUseCase) GenerateApiKey(e *entity.User) (*dto.ApiKeyResponse, error) {
	key, err := util.GenerateAPIKey(40)
	if err != nil {
		logrus.Error("Failed to generate API key:", err)
		return nil, err
	}

	expire, _ := tm.ParseDuration(a.cfg.Server.ExpireKey)
	redisKey := fmt.Sprintf("api_key:%s", key)
	err = a.redis.Set(redisKey, e.UUID.String(), expire)
	if err != nil {
		logrus.Error("Failed to set API key in Redis:", err)
		return nil, err
	}

	return &dto.ApiKeyResponse{ApiKey: key}, nil
}

func (a *AuthUseCase) GoogleVerify(ctx context.Context, googleIdToken string) (*dto.LoginResponse, error) {
	googleClientID := a.cfg.Oauth2.Google.ClientID

	payload, err := idtoken.Validate(ctx, googleIdToken, googleClientID)
	if err != nil {
		logrus.WithError(err).Error("Failed to validate Google ID token")
		return nil, errorEntity.ErrInvalidToken
	}

	gothUser := goth.User{
		Email:     payload.Claims["email"].(string),
		Name:      payload.Claims["name"].(string),
		FirstName: payload.Claims["given_name"].(string),
		UserID:    payload.Subject,
		Provider:  "google",
	}

	return a.GoogleLogin(&gothUser, ctx)
}
