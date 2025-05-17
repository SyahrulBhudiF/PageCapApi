package main

import (
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/infrastructure/database"
	"github.com/SyahrulBhudiF/Doc-Management.git/pkg/config"
	"log"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	db, err := database.NewPostgresDB(cfg)
	if err != nil {
		log.Fatal(err)
	}

	if err := database.Migrate(db); err != nil {
		log.Fatal(err)
	}

	log.Println("Database migrations completed successfully")
}
