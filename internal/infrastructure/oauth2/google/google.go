package google

import (
	"github.com/SyahrulBhudiF/Doc-Management.git/pkg/config"
	"github.com/markbates/goth"
	"github.com/markbates/goth/providers/google"
)

func NewGoogle(cfg *config.Config) {
	goth.UseProviders(
		google.New(
			cfg.Oauth2.Google.ClientID,
			cfg.Oauth2.Google.ClientSecret,
			cfg.Oauth2.Google.RedirectURL,
			"openid", "email", "profile",
		),
	)
}
