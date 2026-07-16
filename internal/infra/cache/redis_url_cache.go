package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/PeshawaAziz/url-shortener/internal/domain/url"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/singleflight"
)

type RedisURLCache struct {
	client *redis.Client
	sf     singleflight.Group
}

func NewRedisURLCache(client *redis.Client) *RedisURLCache {
	return &RedisURLCache{client: client}
}

func (c *RedisURLCache) key(tenantID uuid.UUID, slug string) string {
	return fmt.Sprintf("url:%s:%s", tenantID.String(), slug)
}

func (c *RedisURLCache) Get(ctx context.Context, tenantID uuid.UUID, slug string) (*url.URL, error) {
	val, err := c.client.Get(ctx, c.key(tenantID, slug)).Result()
	if errors.Is(err, redis.Nil) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if val == "NOT_FOUND" {
		return nil, url.ErrURLNotFound
	}

	var u url.URL
	if err := json.Unmarshal([]byte(val), &u); err != nil {
		return nil, err
	}
	return &u, nil
}

func (c *RedisURLCache) Set(ctx context.Context, u *url.URL) error {
	data, err := json.Marshal(u)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, c.key(u.TenantID, string(u.Slug)), data, 1*time.Hour).Err()
}

func (c *RedisURLCache) SetNegative(ctx context.Context, tenantID uuid.UUID, slug string) error {
	return c.client.Set(ctx, c.key(tenantID, slug), "NOT_FOUND", 60*time.Second).Err()
}
