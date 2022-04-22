package redis

import (
	"github.com/go-redis/redis/v8"
	"github.com/markphelps/flipt/config"
)

func NewRedisClient(cfg config.Config) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     cfg.Cache.Redis.Address,
		Password: cfg.Cache.Redis.Password,
		DB:       cfg.Cache.Redis.DB,
	})
}
