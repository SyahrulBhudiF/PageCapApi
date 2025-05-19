package config

import (
	"github.com/joho/godotenv"
	"os"
)

type Config struct {
	Database   DatabaseConfig
	Server     ServerConfig
	Jwt        JwtConfig
	Redis      RedisConfig
	Mail       MailConfig
	Oauth2     Oauth2Config
	Cloudinary CloudinaryConfig
}

type Oauth2Config struct {
	Google GoogleConfig
}

type GoogleConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

type CloudinaryConfig struct {
	CloudName string
	ApiKey    string
	ApiSecret string
}

type DatabaseConfig struct {
	DatabaseUrl string
	Host        string
	Port        string
	User        string
	Password    string
	DbName      string
	SSLMode     string
}

type ServerConfig struct {
	Port        string
	Environment string
	Salt        string
}

type JwtConfig struct {
	AccessTokenSecret  string
	AccessTokenExpire  string
	RefreshTokenSecret string
	RefreshTokenExpire string
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
}

type MailConfig struct {
	Host     string
	Port     string
	Username string
	Password string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	return &Config{
		Database: DatabaseConfig{
			DatabaseUrl: getEnv("DATABASE_URL", ""),
			Host:        getEnv("DB_HOST", ""),
			Port:        getEnv("DB_PORT", ""),
			User:        getEnv("DB_USER", ""),
			Password:    getEnv("DB_PASSWORD", ""),
			DbName:      getEnv("DB_NAME", ""),
			SSLMode:     getEnv("DB_SSL_MODE", ""),
		},
		Server: ServerConfig{
			Port:        getEnv("SERVER_PORT", ""),
			Environment: getEnv("ENVIRONMENT", ""),
			Salt:        getEnv("SALT", ""),
		},
		Jwt: JwtConfig{
			AccessTokenSecret:  getEnv("ACCESS_TOKEN_SECRET", ""),
			AccessTokenExpire:  getEnv("ACCESS_TOKEN_EXPIRE", ""),
			RefreshTokenSecret: getEnv("REFRESH_TOKEN_SECRET", ""),
			RefreshTokenExpire: getEnv("REFRESH_TOKEN_EXPIRE", ""),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", ""),
			Port:     getEnv("REDIS_PORT", ""),
			Password: getEnv("REDIS_PASSWORD", ""),
		},
		Mail: MailConfig{
			Host:     getEnv("MAIL_HOST", ""),
			Port:     getEnv("MAIL_PORT", ""),
			Username: getEnv("MAIL_USERNAME", ""),
			Password: getEnv("MAIL_PASSWORD", ""),
		},
		Oauth2: Oauth2Config{
			Google: GoogleConfig{
				ClientID:     getEnv("GOOGLE_CLIENT_ID", ""),
				ClientSecret: getEnv("GOOGLE_CLIENT_SECRET", ""),
				RedirectURL:  getEnv("GOOGLE_REDIRECT_URL", ""),
			},
		},
		Cloudinary: CloudinaryConfig{
			CloudName: getEnv("CLOUDINARY_CLOUD_NAME", ""),
			ApiKey:    getEnv("CLOUDINARY_API_KEY", ""),
			ApiSecret: getEnv("CLOUDINARY_API_SECRET", ""),
		},
	}, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
