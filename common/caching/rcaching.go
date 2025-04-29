package caching

import (
	"context"
	"time"

	"github.com/besanh/chatbot_gpt/common/util"
	log "github.com/besanh/logger/logging/slog"
	"github.com/redis/go-redis/v9"
)

type (
	IRedisCache interface {
		Keys(pattern string) ([]string, error)
		Set(key string, value any, ttl time.Duration) error
		SetTTL(key string, value any, t time.Duration) (string, error)
		Get(key string) any
		IsExisted(key string) (bool, error)
		IsHExisted(list, key string) (bool, error)
		HGet(list, key string) (string, error)
		HGetAll(list string) (map[string]string, error)
		HSet(key string, values []any) error
		Del(key []string) error
		HDel(key string, fields ...string) error
		GetKeysPattern(pattern string) ([]string, error)
		SetRaw(ctx context.Context, key string, value string) error
		HSetRaw(ctx context.Context, key string, field string, value string) error
		Expire(ctx context.Context, key string, ttl time.Duration) error
		GetTTL(ctx context.Context, key string) (time.Duration, error)
	}
	RedisCache struct {
		client *redis.Client
	}
)

const (
	REDIS_KEEP_TTL = redis.KeepTTL
)

var RCache IRedisCache

func NewRedisCache(client *redis.Client) IRedisCache {
	return &RedisCache{
		client: client,
	}
}

func (r *RedisCache) Keys(pattern string) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	ret, err := r.client.Keys(ctx, pattern).Result()
	return ret, err
}

func (r *RedisCache) Set(key string, value any, ttl time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	val, _ := util.ParseAnyToString(value)
	if _, err := r.client.Set(ctx, key, val, ttl).Result(); err != nil {
		log.Errorf("", err)
		return err
	}
	return nil
}

func (r *RedisCache) Get(key string) any {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		log.Errorf("", err)
		return nil
	}
	return val
}

func (r *RedisCache) SetTTL(key string, value any, t time.Duration) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	ret, err := r.client.Set(ctx, key, value, t).Result()
	return ret, err
}

func (r *RedisCache) IsExisted(key string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	res, err := r.client.Exists(ctx, key).Result()
	if res == 0 || err != nil {
		return false, err
	}
	return true, nil
}

func (r *RedisCache) IsHExisted(list, key string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	res, err := r.client.HExists(ctx, list, key).Result()
	if !res || err != nil {
		return false, err
	}
	return true, nil
}

func (r *RedisCache) HGet(list, key string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	ret, err := r.client.HGet(ctx, list, key).Result()
	return ret, err
}

func (r *RedisCache) HGetAll(list string) (map[string]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	ret, err := r.client.HGetAll(ctx, list).Result()
	return ret, err
}

func (r *RedisCache) HSet(key string, values []any) error {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	_, err := r.client.HSet(ctx, key, values...).Result()
	return err
}

func (r *RedisCache) Del(key []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	err := r.client.Del(ctx, key...).Err()
	return err
}

func (r *RedisCache) HDel(key string, fields ...string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	err := r.client.HDel(ctx, key, fields...).Err()
	return err
}

func (r *RedisCache) GetKeysPattern(pattern string) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	ret, err := r.client.Keys(ctx, pattern).Result()
	return ret, err
}

func (c *RedisCache) SetRaw(ctx context.Context, key string, value string) error {
	_, err := c.client.Set(ctx, key, value, redis.KeepTTL).Result()
	return err
}

func (c *RedisCache) HSetRaw(ctx context.Context, key string, field string, value string) error {
	data := []any{field, value}
	_, err := c.client.HSet(ctx, key, data).Result()
	return err
}

func (c *RedisCache) Expire(ctx context.Context, key string, ttl time.Duration) error {
	err := c.client.Expire(ctx, key, ttl).Err()
	return err
}

func (c *RedisCache) GetTTL(ctx context.Context, key string) (time.Duration, error) {
	ttl, err := c.client.TTL(ctx, key).Result()
	return ttl, err
}
