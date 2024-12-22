package db

import (
	"fmt"

	"shareway/util"

	"github.com/redis/go-redis/v9"
)

func NewRedisClient(cfg util.Config) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:             fmt.Sprintf("%s:%d", cfg.RedisHost, cfg.RedisPort),
		Password:         cfg.RedisPassword,
		DB:               cfg.RedisDB,
		DisableIndentity: true,
		Protocol:         cfg.RedisProtocol,
	})
	return client
}
