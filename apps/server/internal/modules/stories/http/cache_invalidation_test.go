package storieshttp

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/complexus-tech/projects-api/pkg/cache"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/redis/go-redis/v9"
)

func TestStoryCacheReportsInvalidationFailureWithoutLeakingKey(t *testing.T) {
	t.Parallel()

	client := redis.NewClient(&redis.Options{
		Addr:         "127.0.0.1:1",
		DialTimeout:  10 * time.Millisecond,
		ReadTimeout:  10 * time.Millisecond,
		WriteTimeout: 10 * time.Millisecond,
		MaxRetries:   0,
	})
	t.Cleanup(func() { _ = client.Close() })
	var output bytes.Buffer
	log := logger.NewWithJSON(&output, slog.LevelDebug, "story-cache-test")
	observed := newStoryCache(cache.New(client, log), log)

	const rawKey = "story:raw-sensitive-cache-key"
	observed.Delete(context.Background(), rawKey)

	if !strings.Contains(output.String(), "failed to invalidate story cache entry") {
		t.Fatalf("log = %q, want invalidation failure", output.String())
	}
	if strings.Contains(output.String(), rawKey) || strings.Contains(output.String(), "raw-sensitive") {
		t.Fatalf("cache invalidation log leaked raw key: %s", output.String())
	}
}
