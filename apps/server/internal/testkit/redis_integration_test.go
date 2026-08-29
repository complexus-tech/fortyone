//go:build integration

package testkit

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestRedisIntegrationNamespaceIsolationAndCleanup(t *testing.T) {
	t.Parallel()

	first := NewRedis(t)
	second := NewRedis(t)
	if first.Namespace == second.Namespace {
		t.Fatal("independent Redis harnesses shared a namespace")
	}

	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	defer cancel()

	const writers = 32
	errors := make(chan error, writers*2)
	var waitGroup sync.WaitGroup
	for index := range writers {
		waitGroup.Add(2)
		go writeIntegrationRedisValue(ctx, &waitGroup, errors, first, "first", index)
		go writeIntegrationRedisValue(ctx, &waitGroup, errors, second, "second", index)
	}
	waitGroup.Wait()
	close(errors)
	for err := range errors {
		if err != nil {
			t.Fatal(err)
		}
	}

	for index := range writers {
		logicalKey := fmt.Sprintf("concurrent:%d", index)
		firstValue, err := first.Client.Get(ctx, first.Key(logicalKey)).Result()
		if err != nil || firstValue != "first" {
			t.Fatalf("read first namespace value %d: value %q, error %v", index, firstValue, err)
		}
		secondValue, err := second.Client.Get(ctx, second.Key(logicalKey)).Result()
		if err != nil || secondValue != "second" {
			t.Fatalf("read second namespace value %d: value %q, error %v", index, secondValue, err)
		}
	}

	if err := deleteRedisNamespace(ctx, first.Client, first.Namespace); err != nil {
		t.Fatalf("delete first Redis namespace: %v", err)
	}
	for index := range writers {
		logicalKey := fmt.Sprintf("concurrent:%d", index)
		assertRedisKeyExists(t, first.Client, first.Key(logicalKey), false)
		assertRedisKeyExists(t, second.Client, second.Key(logicalKey), true)
	}
}

func writeIntegrationRedisValue(
	ctx context.Context,
	waitGroup *sync.WaitGroup,
	errors chan<- error,
	testRedis *Redis,
	value string,
	index int,
) {
	defer waitGroup.Done()

	logicalKey := fmt.Sprintf("concurrent:%d", index)
	if err := testRedis.Client.Set(ctx, testRedis.Key(logicalKey), value, time.Minute).Err(); err != nil {
		errors <- fmt.Errorf("write %s Redis namespace value %d: %w", value, index, err)
	}
}
