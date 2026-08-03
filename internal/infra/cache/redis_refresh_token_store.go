package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/PeshawaAziz/url-shortener/internal/domain/auth"
	"github.com/redis/go-redis/v9"
)

type RedisRefreshTokenStore struct {
	client *redis.Client
}

func NewRedisRefreshTokenStore(client *redis.Client) *RedisRefreshTokenStore {
	return &RedisRefreshTokenStore{client: client}
}

func (s *RedisRefreshTokenStore) key(hashedToken string) string {
	return fmt.Sprintf("refresh:%s", hashedToken)
}

func (s *RedisRefreshTokenStore) familyKey(familyID string) string {
	return fmt.Sprintf("family:%s", familyID)
}

func (s *RedisRefreshTokenStore) Save(ctx context.Context, hashedToken string, meta auth.RefreshTokenMetadata) error {
	data, err := json.Marshal(meta)
	if err != nil {
		return err
	}
	ttl := time.Until(meta.ExpiresAt)
	if ttl <= 0 {
		return fmt.Errorf("refresh token already expired")
	}
	pipe := s.client.TxPipeline()
	pipe.Set(ctx, s.key(hashedToken), data, ttl)
	pipe.SAdd(ctx, s.familyKey(meta.FamilyID), hashedToken)
	_, err = pipe.Exec(ctx)
	return err
}

func (s *RedisRefreshTokenStore) Get(ctx context.Context, hashedToken string) (*auth.RefreshTokenMetadata, error) {
	val, err := s.client.Get(ctx, s.key(hashedToken)).Result()
	if err == redis.Nil {
		return nil, auth.ErrTokenNotFound
	}
	if err != nil {
		return nil, err
	}
	var meta auth.RefreshTokenMetadata
	if err := json.Unmarshal([]byte(val), &meta); err != nil {
		return nil, err
	}
	return &meta, nil
}

func (s *RedisRefreshTokenStore) MarkUsed(ctx context.Context, hashedToken string) error {
	meta, err := s.Get(ctx, hashedToken)
	if err != nil {
		return err
	}
	meta.Used = true
	data, _ := json.Marshal(meta)
	ttl := time.Until(meta.ExpiresAt)
	if ttl <= 0 {
		return fmt.Errorf("token already expired")
	}
	return s.client.Set(ctx, s.key(hashedToken), data, ttl).Err()
}

func (s *RedisRefreshTokenStore) RevokeFamily(ctx context.Context, familyID string) error {
	hashedTokens, err := s.client.SMembers(ctx, s.familyKey(familyID)).Result()
	if err != nil {
		return err
	}
	keys := make([]string, 0, len(hashedTokens))
	for _, ht := range hashedTokens {
		keys = append(keys, s.key(ht))
	}
	keys = append(keys, s.familyKey(familyID))
	if len(keys) > 0 {
		return s.client.Del(ctx, keys...).Err()
	}
	return nil
}
