package caching

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/besanh/chatting/common/util"
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
		HMGet(key string, fields ...string) ([]any, error)
		HMSet(key string, values ...any) error
		HMDel(key string, fields ...string) error
		FLUSHALL() any
		Del(key []string) error
		HDel(key string, fields ...string) error
		GetKeysPattern(pattern string) ([]string, error)
		Close()
		Incr(ctx context.Context, key string) (int64, error)
		Decr(ctx context.Context, key string) (int64, error)
		SetRaw(ctx context.Context, key, value string) error
		HSetRaw(ctx context.Context, key, field, value string) error
		SCARD(ctx context.Context, key string) (int64, error)
		SPOP(ctx context.Context, key string) (string, error)
		SPOPN(ctx context.Context, key string, count int64) ([]string, error)
		SMembers(ctx context.Context, key string) ([]string, error)
		SADD(ctx context.Context, key string, value ...any) error
		SRANDMEMBER(ctx context.Context, key string, count int64) ([]string, error)
		SISMEMBER(ctx context.Context, key, value string) (bool, error)
		ZADD(ctx context.Context, key string, score float64, v string) error
		ZRangeByScore(ctx context.Context, key string, min, max float64, count int) ([]string, error)
		ZRem(ctx context.Context, key string, value ...any) error
		ZCount(ctx context.Context, key string, min, max float64) (int64, error)
		ZScore(ctx context.Context, key, member string) (float64, error)
		SADDRaw(ctx context.Context, key string, value ...string) error
		SREM(ctx context.Context, key string, value ...string) error
		RPush(ctx context.Context, key string, value any) error
		Expire(ctx context.Context, key string, time time.Duration) (bool, error)

		// Transaction
		TxPineLine() redis.Pipeliner
		TxSet(ctx context.Context, redisTx redis.Pipeliner, key string, value any, expire time.Duration)
		TxExec(ctx context.Context, redisTx redis.Pipeliner) error
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
		log.Error(err)
		return err
	}
	return nil
}

func (r *RedisCache) Get(key string) any {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		log.Error(err)
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

func (r *RedisCache) HMSet(key string, values ...any) error {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	ret, err := r.client.HMSet(ctx, key, values...).Result()
	if err != nil {
		return err
	}
	if !ret {
		err = errors.New("HashMap Set failed")
	}
	return err
}

func (r *RedisCache) HMDel(key string, fields ...string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	err := r.client.HDel(ctx, key, fields...).Err()
	return err
}

func (r *RedisCache) FLUSHALL() any {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	ret := r.client.FlushAll(ctx)
	return ret
}

func (r *RedisCache) HMGet(key string, fields ...string) ([]any, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	ret, err := r.client.HMGet(ctx, key, fields...).Result()
	return ret, err
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

func (r *RedisCache) Close() {
	r.client.Close()
}

func (c *RedisCache) Incr(ctx context.Context, key string) (int64, error) {
	cnt, err := c.client.Incr(ctx, key).Result()
	return cnt, err
}

func (c *RedisCache) SetRaw(ctx context.Context, key, value string) error {
	_, err := c.client.Set(ctx, key, value, redis.KeepTTL).Result()
	return err
}

func (c *RedisCache) HSetRaw(ctx context.Context, key, field, value string) error {
	data := []any{field, value}
	_, err := c.client.HSet(ctx, key, data).Result()
	return err
}

func (c *RedisCache) Decr(ctx context.Context, key string) (int64, error) {
	cnt, err := c.client.Decr(ctx, key).Result()
	if err == redis.Nil {
		return 0, errors.New("key not found")
	}
	return cnt, err
}

func (c *RedisCache) SADD(ctx context.Context, key string, value ...any) error {
	arr := make([]any, 0)
	for _, v := range value {
		str, err := valueToString(v)
		if err != nil {
			return err
		}
		arr = append(arr, str)
	}
	_, err := c.client.SAdd(ctx, key, arr).Result()
	return err
}

func (c *RedisCache) SADDRaw(ctx context.Context, key string, value ...string) error {
	_, err := c.client.SAdd(ctx, key, value).Result()
	return err
}

func (c *RedisCache) SMembers(ctx context.Context, key string) ([]string, error) {
	value, err := c.client.SMembers(ctx, key).Result()
	return value, err
}

func (c *RedisCache) SREM(ctx context.Context, key string, value ...string) error {
	_, err := c.client.SRem(ctx, key, value).Result()
	return err
}

func (c *RedisCache) SPOP(ctx context.Context, key string) (string, error) {
	value, err := c.client.SPop(ctx, key).Result()
	if err == redis.Nil {
		return "", nil
	}
	return value, err
}

func (c *RedisCache) SPOPN(ctx context.Context, key string, count int64) ([]string, error) {
	value, err := c.client.SPopN(ctx, key, count).Result()
	if err == redis.Nil {
		return nil, nil
	}
	return value, err
}

func (s *RedisCache) SCARD(ctx context.Context, key string) (int64, error) {
	value, err := s.client.SCard(ctx, key).Result()
	return value, err
}
func (s *RedisCache) SRANDMEMBER(ctx context.Context, key string, count int64) ([]string, error) {
	value, err := s.client.SRandMemberN(ctx, key, count).Result()
	return value, err
}

func (s *RedisCache) SISMEMBER(ctx context.Context, key, value string) (bool, error) {
	exists, err := s.client.SIsMember(ctx, key, value).Result()
	return exists, err
}

func (c *RedisCache) ZADD(ctx context.Context, key string, score float64, v string) error {
	members := redis.Z{
		Score:  score,
		Member: v,
	}
	_, err := c.client.ZAdd(ctx, key, members).Result()
	return err
}

func (c *RedisCache) ZRangeByScore(ctx context.Context, key string, min, max float64, count int) ([]string, error) {
	value, err := c.client.ZRangeByScore(ctx, key, &redis.ZRangeBy{
		Min:    fmt.Sprintf("%f", min),
		Max:    fmt.Sprintf("%f", max),
		Offset: 0,
		Count:  int64(count),
	}).Result()
	if err == redis.Nil {
		return nil, nil
	}
	return value, err
}

func (c *RedisCache) ZRem(ctx context.Context, key string, value ...any) error {
	_, err := c.client.ZRem(ctx, key, value...).Result()
	if err == redis.Nil {
		return nil
	}
	return err
}

func (c *RedisCache) ZCount(ctx context.Context, key string, min, max float64) (int64, error) {
	minStr := fmt.Sprintf("%f", min)
	if min == -1 {
		minStr = "-inf"
	}
	maxStr := fmt.Sprintf("%f", max)
	if max == -1 {
		maxStr = "+inf"
	}
	count, err := c.client.ZCount(ctx, key, minStr, maxStr).Result()
	return count, err
}

func (c *RedisCache) ZScore(ctx context.Context, key, member string) (float64, error) {
	score, err := c.client.ZScore(ctx, key, member).Result()
	if err == redis.Nil {
		return -1, nil
	}
	return score, err
}

func valueToString(value any) (string, error) {
	tmp, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	return string(tmp), nil
}

func (c *RedisCache) RPush(ctx context.Context, key string, value any) error {
	_, err := c.client.RPush(ctx, key, value).Result()
	if err != nil {
		return err
	}
	return err
}

func (c *RedisCache) TxPineLine() redis.Pipeliner {
	return c.client.TxPipeline()
}

func (c *RedisCache) TxSet(ctx context.Context, redisTx redis.Pipeliner, key string, value any, expire time.Duration) {
	redisTx.Set(ctx, key, value, expire)
}

func (c *RedisCache) TxExec(ctx context.Context, redisTx redis.Pipeliner) error {
	_, err := redisTx.Exec(ctx)
	return err
}

func (c *RedisCache) Expire(ctx context.Context, key string, time time.Duration) (bool, error) {
	exist, err := c.client.Expire(ctx, key, time).Result()
	return exist, err
}
