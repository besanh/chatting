package redis

import (
	"context"
	"time"

	log "github.com/besanh/logger/logging/slog"
	"github.com/redis/go-redis/v9"
)

type (
	IRedis interface {
		GetClient() *redis.Client
		Connect() error
	}

	RedisClient struct {
		Client *redis.Client
		Config RedisConfig
	}

	RedisConfig struct {
		Dsn string
	}
)

var Redis IRedis

func NewRedis(config RedisConfig) (IRedis, error) {
	redisClient := &RedisClient{
		Config: config,
	}

	if err := redisClient.Connect(); err != nil {
		log.Errorf("redis connect error: %v", err)
		return nil, err
	}
	return redisClient, nil
}

func (r *RedisClient) GetClient() *redis.Client {
	return r.Client
}

func (r *RedisClient) Connect() error {
	dsn, err := redis.ParseURL(r.Config.Dsn)
	if err != nil {
		return err
	}
	client := redis.NewClient(dsn)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		return err
	}
	str, err := client.Ping(ctx).Result()
	if err != nil {
		log.Errorf("redis ping error: %v", err)
		return err
	}
	log.Infof(str)
	r.Client = client
	return nil
}
