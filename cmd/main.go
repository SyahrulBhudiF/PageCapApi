package main

import (
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/core"
	"github.com/sirupsen/logrus"
)

func main() {
	app, err := core.Bootstrap()
	if err != nil {
		logrus.Fatal("Failed to bootstrap app:", err)
	}

	if err := app.Router.Run(":8080"); err != nil {
		logrus.Fatal("Failed to run server:", err)
	}
}
