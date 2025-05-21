package module

import (
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/application/usecase"
	redisContract "github.com/SyahrulBhudiF/Doc-Management.git/internal/domain/contract/redis"
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/domain/contract/repository"
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/interface/http/handler"
	"github.com/SyahrulBhudiF/Doc-Management.git/pkg/config"
	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/go-rod/rod"
)

func InitPageCaptureModule(cfg *config.Config, repo repository.PageCaptureRepository, redis redisContract.Service, cloud *cloudinary.Cloudinary, browser *rod.Browser) *handler.PageCaptureHandler {
	pageCaptureUC := usecase.NewPageCaptureUseCase(repo, redis, cfg, cloud, browser)
	pageCaptureHandler := handler.NewPageCaptureHandler(pageCaptureUC)

	return pageCaptureHandler
}
