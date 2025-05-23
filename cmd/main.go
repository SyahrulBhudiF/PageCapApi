package main

import (
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/core"
	"github.com/sirupsen/logrus"
	_ "github.com/swaggo/files"
	_ "github.com/swaggo/gin-swagger"
)

// @title Page Capture API (PageCapApi)
// @version 1.0
// @description This is a documentation for PageCapApi
// @host localhost:8080
// @BasePath /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.
func main() {
	app, err := core.Bootstrap()
	if err != nil {
		logrus.Fatal("Failed to bootstrap app:", err)
	}

	if err := app.Router.Run(":8080"); err != nil {
		logrus.Fatal("Failed to run server:", err)
	}
}
