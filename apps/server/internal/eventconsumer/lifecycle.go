package eventconsumer

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type consumerLifecycle struct {
	initialized atomic.Bool
}

const streamReadBlock = 2 * time.Second

// Initialize creates the required Redis consumer group. It is intentionally a
// separate startup step so the process cannot report ready when the stream
// contract is unavailable.
func (c *Consumer) Initialize(ctx context.Context) error {
	if c == nil || c.redis == nil {
		return errors.New("redis stream consumer requires a Redis client")
	}
	if ctx == nil {
		return errors.New("redis stream consumer context is required")
	}

	err := c.redis.XGroupCreateMkStream(ctx, eventStreamKey, eventConsumerGroup, "0").Err()
	if err != nil && !isRedisGroupError(err, "BUSYGROUP") {
		return fmt.Errorf("create Redis stream consumer group %q: %w", eventConsumerGroup, err)
	}
	c.lifecycle.initialized.Store(true)
	return nil
}

// Start initializes and runs the consumer using Redis Streams. New process
// bootstraps should call Initialize before starting any externally visible
// listener, then call Run under their lifecycle supervisor.
func (c *Consumer) Start(ctx context.Context) error {
	if err := c.Initialize(ctx); err != nil {
		return err
	}
	return c.Run(ctx)
}

// Run processes new and reclaimable stream messages until ctx is cancelled.
// Both loops are owned and joined here; neither outlives the consumer.
func (c *Consumer) Run(ctx context.Context) error {
	if c == nil || c.redis == nil {
		return errors.New("redis stream consumer requires a Redis client")
	}
	if ctx == nil {
		return errors.New("redis stream consumer context is required")
	}
	if !c.lifecycle.initialized.Load() {
		return errors.New("redis stream consumer must be initialized before it runs")
	}

	instanceID := uuid.New().String()
	c.log.Info(ctx, "starting redis stream consumer", "instance_id", instanceID)

	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	results := make(chan error, 2)
	go func() {
		results <- c.processNewMessageLoop(runCtx, instanceID)
	}()
	go func() {
		results <- c.claimPendingMessages(runCtx, instanceID)
	}()

	firstErr := <-results
	cancel()
	secondErr := <-results
	if ctx.Err() != nil {
		return nil
	}
	return errors.Join(firstErr, secondErr)
}

func (c *Consumer) processNewMessageLoop(ctx context.Context, instanceID string) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			if err := c.processNewMessages(ctx, instanceID); err != nil {
				if ctx.Err() != nil {
					return nil
				}
				if isRedisGroupError(err, "NOGROUP") {
					return fmt.Errorf("redis stream consumer group %q disappeared: %w", eventConsumerGroup, err)
				}
				c.log.Error(ctx, "error processing messages", "error", err)
				if !waitForConsumerRetry(ctx, time.Second) {
					return nil
				}
			}
		}
	}
}

func (c *Consumer) processNewMessages(ctx context.Context, instanceID string) error {
	streams, err := c.redis.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group:    eventConsumerGroup,
		Consumer: instanceID,
		Streams:  []string{eventStreamKey, ">"},
		Count:    streamReadCount,
		Block:    streamReadBlock,
	}).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil
		}
		return err
	}

	for _, stream := range streams {
		for _, message := range stream.Messages {
			if err := c.processStreamMessage(ctx, message, instanceID); err != nil {
				c.log.Error(ctx, "failed to process message", "message_id", message.ID, "error", err)
			}
		}
	}
	return nil
}

func (c *Consumer) claimPendingMessages(ctx context.Context, instanceID string) error {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			pending, err := c.redis.XPendingExt(ctx, &redis.XPendingExtArgs{
				Stream: eventStreamKey,
				Group:  eventConsumerGroup,
				Start:  "-",
				End:    "+",
				Count:  50,
			}).Result()
			if err != nil {
				if ctx.Err() != nil {
					return nil
				}
				if isRedisGroupError(err, "NOGROUP") {
					return fmt.Errorf("redis stream consumer group %q disappeared while reclaiming messages: %w", eventConsumerGroup, err)
				}
				c.log.Error(ctx, "failed to get pending messages", "error", err)
				continue
			}

			for _, pendingMessage := range pending {
				if pendingMessage.Idle <= pendingClaimTimeout {
					continue
				}
				claimed, err := c.redis.XClaim(ctx, &redis.XClaimArgs{
					Stream:   eventStreamKey,
					Group:    eventConsumerGroup,
					Consumer: instanceID,
					MinIdle:  pendingClaimTimeout,
					Messages: []string{pendingMessage.ID},
				}).Result()
				if err != nil {
					c.log.Error(ctx, "failed to claim message", "message_id", pendingMessage.ID, "error", err)
					continue
				}
				for _, message := range claimed {
					if err := c.processStreamMessage(ctx, message, instanceID); err != nil {
						c.log.Error(ctx, "failed to process claimed message", "message_id", message.ID, "error", err)
					}
				}
			}
		}
	}
}

func waitForConsumerRetry(ctx context.Context, delay time.Duration) bool {
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}

func isRedisGroupError(err error, code string) bool {
	if err == nil {
		return false
	}
	message := strings.TrimSpace(err.Error())
	return strings.HasPrefix(message, code+" ") || message == code || strings.Contains(message, " "+code+" ")
}
