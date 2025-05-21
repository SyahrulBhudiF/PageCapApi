package handler

import (
	"fmt"
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/application/usecase"
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/domain/dto"
	errorEntity "github.com/SyahrulBhudiF/Doc-Management.git/internal/domain/error"
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/shared/util"
	"github.com/SyahrulBhudiF/Doc-Management.git/pkg/response"
	"github.com/gin-gonic/gin"
	"net/http"
)

type PageCaptureHandler struct {
	pageCapture *usecase.PageCaptureUseCase
}

func NewPageCaptureHandler(pageCapture *usecase.PageCaptureUseCase) *PageCaptureHandler {
	return &PageCaptureHandler{
		pageCapture: pageCapture,
	}
}

// PageCapture godoc
// @Summary      Get Page Capture
// @Description  Get Page Capture
// @Tags         Page Capture
// @Accept       json
// @Produce      json
// @Param        key      path  string                  true  "Key for Page Capture"
// @Param        request  body  dto.PageCaptureRequest  true  "Page Capture Request"
// @Success 200 {file} binary "Successfully get Page Capture image"
// @Failure 400 {object} response.ErrorResponse "invalid request"
// @Failure 401 {object} response.ErrorResponse "unauthorized"
// @Failure 500 {object} response.ErrorResponse "internal server error"
// @Router       /page-capture/{key} [post]
func (h *PageCaptureHandler) PageCapture(c *gin.Context) {
	body, err := util.GetBody[dto.PageCaptureRequest](c, "body")
	if err != nil {
		response.BadRequest(c, "invalid request", err)
		return
	}

	key := c.Param("key")
	if key == "" {
		response.BadRequest(c, "missing required param parameter 'key'", nil)
		return
	}

	data, err := h.pageCapture.PageCapture(&body, key, c.Request.Context())
	if err != nil {
		if util.ErrorInList(err, errorEntity.ErrInvalidCredentials, errorEntity.ErrUserNotFound, errorEntity.ErrCloudinaryUpload) {
			response.Unauthorized(c, "unauthorized", err)
		} else {
			response.InternalServerError(c, err)
		}
		return
	}

	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, data.Filename))
	c.Data(http.StatusOK, "image/png", data.Content)
}
