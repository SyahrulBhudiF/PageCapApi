package handler

import (
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/application/usecase"
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/domain/dto"
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/domain/entity"
	errorEntity "github.com/SyahrulBhudiF/Doc-Management.git/internal/domain/error"
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/shared/util"
	"github.com/SyahrulBhudiF/Doc-Management.git/pkg/response"
	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	user *usecase.UserUseCase
}

func NewUserHandler(user *usecase.UserUseCase) *UserHandler {
	return &UserHandler{
		user: user,
	}
}

// GetProfile godoc
// @Summary      Get Profile
// @Description  Get user profile
// @Tags         User
// @Accept       json
// @Produce      json
// @Success 200 {object} response.Response{data=dto.UserResponse} "Successfully get user profile"
// @Failure 400 {object} response.ErrorResponse "invalid request"
// @Failure 401 {object} response.ErrorResponse "unauthorized"
// @Failure 500 {object} response.ErrorResponse "internal server error"
// @Security BearerAuth
// @Router       /user/profile [get]
func (h *UserHandler) GetProfile(c *gin.Context) {
	user, err := util.GetBody[entity.User](c, "user")
	if err != nil {
		response.BadRequest(c, "invalid request", err)
		return
	}

	response.OK(c, "successfully get user profile", dto.ToUserResponse(user))
}

// ChangePassword godoc
// @Summary      Change Password
// @Description  Change user password
// @Tags         User
// @Accept       json
// @Produce      json
// @Param        request  body  dto.ChangePasswordRequest  true  "Change Password Request"
// @Success 200 {object} response.Response{data=dto.UserResponse} "Successfully change user password"
// @Failure 400 {object} response.ErrorResponse "invalid request"
// @Failure 401 {object} response.ErrorResponse "unauthorized"
// @Failure 500 {object} response.ErrorResponse "internal server error"
// @Security BearerAuth
// @Router       /user/change-password [patch]
func (h *UserHandler) ChangePassword(c *gin.Context) {
	user, err := util.GetBody[entity.User](c, "user")
	if err != nil {
		response.BadRequest(c, "invalid request", err)
		return
	}

	body, err := util.GetBody[dto.ChangePasswordRequest](c, "body")
	if err != nil {
		response.BadRequest(c, "invalid request", err)
		return
	}

	err = h.user.ChangePassword(&body, &user, c.Request.Context())
	if err != nil {
		if util.ErrorInList(err, errorEntity.ErrUserNotFound, errorEntity.ErrInvalidPassword, errorEntity.ErrPasswordNotSet) {
			response.Unauthorized(c, "unauthorized", err)
			return
		}
		response.InternalServerError(c, err)
		return
	}

	response.OK(c, "successfully change password", nil)
}
