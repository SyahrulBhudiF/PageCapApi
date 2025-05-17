package redis

import (
	"context"
	"fmt"
	"github.com/SyahrulBhudiF/Doc-Management.git/pkg/config"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

func NewRedis(cfg *config.Config) (*redis.Client, error) {
	host := cfg.Redis.Host
	port := cfg.Redis.Port
	password := cfg.Redis.Password
	if host == "" || port == "" {
		return nil, fmt.Errorf("redis host or port is not set in the configuration")
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", host, port),
		Password: password,
	})

	logrus.Info("Connecting to Redis...")

	_, err := rdb.Ping(context.Background()).Result()

	if err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %v", err)
	}

	return rdb, nil
}
