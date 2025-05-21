package database

import (
	"fmt"
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/domain/entity"
	"github.com/SyahrulBhudiF/Doc-Management.git/pkg/config"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"net/url"
	"strings"
	"time"
)

func NewPostgresDB(cfg *config.Config) (*gorm.DB, error) {
	if cfg.Database.DatabaseUrl == "" {
		return nil, fmt.Errorf("DATABASE_URL is not set in the configuration")
	}

	dbURL := cfg.Database.DatabaseUrl
	if strings.HasPrefix(dbURL, "postgresql://") {
		parsedURL, err := url.Parse(dbURL)
		if err != nil {
			return nil, fmt.Errorf("failed to parse database URL: %w", err)
		}

		userInfo := parsedURL.User
		if userInfo != nil {
			username := userInfo.Username()
			password, _ := userInfo.Password()

			parsedURL.User = url.UserPassword(
				url.QueryEscape(username),
				url.QueryEscape(password),
			)
			dbURL = parsedURL.String()
		}
	}

	db, err := gorm.Open(postgres.Open(dbURL), &gorm.Config{
		PrepareStmt: true,
		Logger:      logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)
	sqlDB.SetConnMaxIdleTime(5 * time.Minute)

	logrus.Info("Connected to PostgreSQL database")
	return db, nil
}

func Migrate(db *gorm.DB) error {
	if err := db.Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp"`).Error; err != nil {
		return fmt.Errorf("failed to create uuid-ossp extension: %w", err)
	}

	err := db.Migrator().DropTable(
		&entity.User{},
		&entity.PageCapture{},
	)
	if err != nil {
		return fmt.Errorf("failed to drop tables: %w", err)
	}

	if err := db.AutoMigrate(
		&entity.User{},
		&entity.PageCapture{},
	); err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	logrus.Info("Database migrations completed successfully")
	return nil
}
