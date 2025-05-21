package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/SyahrulBhudiF/Doc-Management.git/internal/application/usecase"
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/domain/dto"
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/domain/entity"
	errorEntity "github.com/SyahrulBhudiF/Doc-Management.git/internal/domain/error"
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/shared/util"
	"github.com/SyahrulBhudiF/Doc-Management.git/pkg/response"
	"github.com/gin-gonic/gin"
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

// GetPageCapture godoc
// @Summary      Get Page Capture
// @Description  Get Page Capture
// @Tags         Page Capture
// @Accept       json
// @Produce      json
// @Param        search    query     string  false  "Search keyword (matches url, image_path, or public_id)"
// @Param        orderBy   query     string  false  "Order by field (default: created_at)"
// @Param        sort      query     string  false  "Sort direction: asc or desc (default: desc)"
// @Param        page      query     int     false  "Page number (default: 1)"
// @Param        fullPage  query     bool    false  "Filter by full_page (true or false)"
// @Param        isMobile  query     bool    false  "Filter by is_mobile (true or false)"
// @Success 200 {object} response.Response{data=dto.PagesCaptureResponse} "Successfully get Page Capture"
// @Failure 400 {object} response.ErrorResponse "invalid request"
// @Failure 401 {object} response.ErrorResponse "unauthorized"
// @Failure 500 {object} response.ErrorResponse "internal server error"
// @Security BearerAuth
// @Router       /page-capture [get]
func (h *PageCaptureHandler) GetPageCapture(c *gin.Context) {
	user, err := util.GetBody[entity.User](c, "user")
	if err != nil {
		response.BadRequest(c, "invalid request", err)
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	search := c.DefaultQuery("search", "")
	orderBy := c.DefaultQuery("order_by", "created_at")
	sort := c.DefaultQuery("sort", "desc")

	var fullPage *bool
	if val, ok := c.GetQuery("full_page"); ok {
		b := val == "true"
		fullPage = &b
	}

	var isMobile *bool
	if val, ok := c.GetQuery("is_mobile"); ok {
		b := val == "true"
		isMobile = &b
	}

	pageCapture, err := h.pageCapture.GetPageCapture(&user, search, orderBy, sort, page, fullPage, isMobile)
	if err != nil {
		if util.ErrorInList(err, errorEntity.ErrUserNotFound) {
			response.Unauthorized(c, "unauthorized", err)
			return
		} else if util.ErrorInList(err, errorEntity.ErrDataNotFound) {
			response.NotFound(c, "data not found", err)
			return
		}
		response.InternalServerError(c, err)
		return
	}

	response.OK(c, "successfully get page capture", pageCapture)
}
