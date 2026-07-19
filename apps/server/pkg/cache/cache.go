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

var incrementWithTTLScript = redis.NewScript(`
local count = redis.call("INCR", KEYS[1])
local ttl = redis.call("PTTL", KEYS[1])
if ttl < 0 then
  redis.call("PEXPIRE", KEYS[1], ARGV[1])
end
return count
`)

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

// IncrementWithTTL atomically increments a counter and ensures it expires.
// It is intended for fixed-window counters such as request rate limits.
func (s *Service) IncrementWithTTL(ctx context.Context, key string, ttl time.Duration) (int64, error) {
	if ttl <= 0 {
		return 0, errors.New("cache counter ttl must be positive")
	}

	count, err := incrementWithTTLScript.Run(ctx, s.redis, []string{key}, ttl.Milliseconds()).Int64()
	if err != nil {
		s.log.Error(ctx, "failed to increment cache counter", "key", key, "error", err)
		return 0, err
	}

	return count, nil
}

// Delete removes a value from the cache
func (s *Service) Delete(ctx context.Context, key string) error {
	return s.redis.Del(ctx, key).Err()
}

// DeleteByPattern removes values from cache matching a pattern
// It uses SCAN to iterate over keys, which is safer for production than KEYS.
func (s *Service) DeleteByPattern(ctx context.Context, pattern string) error {
	var allKeys []string
	iter := s.redis.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		allKeys = append(allKeys, iter.Val())
	}
	if err := iter.Err(); err != nil {
		s.log.Error(ctx, "failed to scan keys for pattern", "pattern", pattern, "error", err)
		return err
	}

	if len(allKeys) > 0 {
		if err := s.redis.Del(ctx, allKeys...).Err(); err != nil {
			s.log.Error(ctx, "failed to delete keys by pattern", "pattern", pattern, "keys_count", len(allKeys), "error", err)
			return err
		}
	}

	return nil
}
