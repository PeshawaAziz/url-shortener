package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisRateLimiter struct {
	client *redis.Client
}

func NewRedisRateLimiter(client *redis.Client) *RedisRateLimiter {
	return &RedisRateLimiter{client: client}
}

func (r *RedisRateLimiter) Allow(ctx context.Context, tenantID, slug string, limit int) (bool, error) {
	key := fmt.Sprintf("ratelimit:%s:%s", tenantID, slug)
	now := time.Now().Unix()
	windowStart := now - 3600

	script := `
        redis.call('ZREMRANGEBYSCORE', KEYS[1], 0, ARGV[1])
        local count = redis.call('ZCARD', KEYS[1])
        if count < tonumber(ARGV[2]) then
            redis.call('ZADD', KEYS[1], ARGV[3], ARGV[3])
            redis.call('EXPIRE', KEYS[1], 3600)
            return 1
        else
            return 0
        end
    `

	res, err := r.client.Eval(ctx, script, []string{key}, windowStart, limit, now).Result()
	if err != nil {
		return false, err
	}
	return res.(int64) == 1, nil
}
