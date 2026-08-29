package cache

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/redis/go-redis/v9"
)

func TestTakeConsumesSingleUseValueAtomically(t *testing.T) {
	t.Parallel()

	redisServer := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})
	t.Cleanup(func() { _ = client.Close() })
	service := New(client, logger.NewWithJSON(&bytes.Buffer{}, slog.LevelError, "cache-test"))

	ctx := context.Background()
	const key = "oauth:state:single-use"
	want := struct {
		Subject string `json:"subject"`
	}{Subject: "user-123"}
	if err := service.Set(ctx, key, want, time.Minute); err != nil {
		t.Fatalf("set single-use value: %v", err)
	}

	start := make(chan struct{})
	errorsByAttempt := make([]error, 2)
	values := make([]string, 2)
	var wait sync.WaitGroup
	for index := range errorsByAttempt {
		wait.Add(1)
		go func() {
			defer wait.Done()
			<-start
			var got struct {
				Subject string `json:"subject"`
			}
			errorsByAttempt[index] = service.Take(ctx, key, &got)
			values[index] = got.Subject
		}()
	}
	close(start)
	wait.Wait()

	successes := 0
	notFound := 0
	for index, err := range errorsByAttempt {
		switch {
		case err == nil:
			successes++
			if values[index] != want.Subject {
				t.Fatalf("consumed subject = %q, want %q", values[index], want.Subject)
			}
		case errors.Is(err, ErrNotFound):
			notFound++
		default:
			t.Fatalf("take attempt %d: %v", index, err)
		}
	}
	if successes != 1 || notFound != 1 {
		t.Fatalf("take results: successes=%d not_found=%d", successes, notFound)
	}
}

func TestCacheErrorsDoNotLogRawKeys(t *testing.T) {
	t.Parallel()

	var output bytes.Buffer
	client := redis.NewClient(&redis.Options{
		Addr:         "127.0.0.1:1",
		DialTimeout:  10 * time.Millisecond,
		ReadTimeout:  10 * time.Millisecond,
		WriteTimeout: 10 * time.Millisecond,
		MaxRetries:   0,
	})
	t.Cleanup(func() { _ = client.Close() })
	service := New(client, logger.NewWithJSON(&output, slog.LevelError, "cache-test"))

	const rawKey = "auth:session:raw-bearer-token"
	if err := service.Set(context.Background(), rawKey, "value", time.Minute); err == nil {
		t.Fatal("expected cache connection error")
	}

	logged := output.String()
	if strings.Contains(logged, rawKey) || strings.Contains(logged, "raw-bearer-token") {
		t.Fatalf("cache log exposed raw key: %s", logged)
	}
	if !strings.Contains(logged, "key_fingerprint") {
		t.Fatalf("cache log omitted safe correlation fingerprint: %s", logged)
	}
}
