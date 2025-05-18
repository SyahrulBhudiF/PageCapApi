package redis

import (
	"context"
	"fmt"
	_interface "github.com/SyahrulBhudiF/Doc-Management.git/internal/domain/contract/redis"
	"github.com/redis/go-redis/v9"
	"time"
)

var _ _interface.Service = (*Service)(nil)

type Service struct {
	redisClient *redis.Client
	prefix      string
	ctx         context.Context
}

func NewRedisService(redisClient *redis.Client, prefix string) *Service {
	return &Service{
		redisClient: redisClient,
		prefix:      fmt.Sprintf("%s:", prefix),
		ctx:         context.Background(),
	}
}

func (r *Service) Set(key string, value any, expiration time.Duration) error {
	err := r.redisClient.Set(r.ctx, key, value, expiration).Err()
	if err != nil {
		return err
	}
	return nil
}

func (r *Service) Get(key string) (string, error) {
	val, err := r.redisClient.Get(r.ctx, key).Result()
	if err != nil {
		return "", err
	}
	return val, nil
}

func (r *Service) Delete(key string) error {
	err := r.redisClient.Del(r.ctx, key).Err()
	if err != nil {
		return err
	}
	return nil
}

func (r *Service) Exists(key string) (bool, error) {
	val, err := r.redisClient.Exists(r.ctx, key).Result()
	if err != nil {
		return false, err
	}
	if val == 0 {
		return false, nil
	}
	return true, nil
}

func (r *Service) Incr(key string) (int64, error) {
	return r.redisClient.Incr(context.Background(), key).Result()
}

func (r *Service) Decr(key string) (int64, error) {
	return r.redisClient.Decr(context.Background(), key).Result()
}

func (r *Service) Expire(key string, expiration time.Duration) error {
	return r.redisClient.Expire(context.Background(), key, expiration).Err()
}
