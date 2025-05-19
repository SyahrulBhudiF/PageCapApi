package cloudinary

import (
	"github.com/SyahrulBhudiF/Doc-Management.git/pkg/config"
	"github.com/cloudinary/cloudinary-go/v2"
)

func NewCloudinary(cfg *config.Config) *cloudinary.Cloudinary {
	cld, err := cloudinary.NewFromParams(cfg.Cloudinary.CloudName, cfg.Cloudinary.ApiKey, cfg.Cloudinary.ApiSecret)
	if err != nil {
		panic("Failed to create Cloudinary instance: " + err.Error())
	}

	return cld
}
