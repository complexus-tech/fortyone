package cache

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/redis/go-redis/v9"
)

// Common errors
var (
	ErrNotFound = errors.New("item not found in cache")
)

// Service defines the cache service interface
type Service struct {
	log   *logger.Logger
	redis *redis.Client
}

// New creates a new cache service instance
func New(redis *redis.Client, log *logger.Logger) *Service {
	return &Service{
		log:   log,
		redis: redis,
	}
}

// Set stores a value in the cache with the given TTL
func (s *Service) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	if err := s.redis.Set(ctx, key, data, ttl).Err(); err != nil {
		s.log.Error(ctx, "failed to set cache", "key", key, "error", err)
		return err
	}

	return nil
}

// Get retrieves a value from the cache
func (s *Service) Get(ctx context.Context, key string, dest any) error {
	data, err := s.redis.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return ErrNotFound
		}
		s.log.Error(ctx, "failed to get from cache", "key", key, "error", err)
		return err
	}

	if err := json.Unmarshal(data, dest); err != nil {
		s.log.Error(ctx, "failed to unmarshal cache data", "key", key, "error", err)
		return err
	}

	return nil
}

// Delete removes a value from the cache
func (s *Service) Delete(ctx context.Context, key string) error {
	return s.redis.Del(ctx, key).Err()
}

// DeleteByPattern removes values from cache matching a pattern
func (s *Service) DeleteByPattern(ctx context.Context, pattern string) error {
	keys, err := s.redis.Keys(ctx, pattern).Result()
	if err != nil {
		return err
	}

	if len(keys) > 0 {
		return s.redis.Del(ctx, keys...).Err()
	}

	return nil
}
