//go:build integration

package chatrepository

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	domain "github.com/complexus-tech/projects-api/internal/modules/chat/domain"
	service "github.com/complexus-tech/projects-api/internal/modules/chat/service"
	platformdatabase "github.com/complexus-tech/projects-api/internal/platform/database"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/require"
)

func TestChatRetriesRollbackReceiptOutboxAndPostCommitEffects(t *testing.T) {
	f := newFixture(t)
	var attempts, effects int
	chat := service.New(f.repo, service.WithMessageCreatedHook(func(ctx context.Context, _ domain.CreateMessageCommand, _ domain.Message) error {
		attempts++
		platformdatabase.AfterCommit(ctx, func() { effects++ })
		if attempts <= 4 {
			return &pgconn.PgError{Code: "40001"}
		}
		return nil
	}))
	command := domain.CreateMessageCommand{Scope: f.scope, ConversationID: f.room, CreationID: uuid.New(), Text: "Retry retains the same receipt"}
	message, err := chat.CreateMessage(f.ctx, command)
	require.NoError(t, err)
	require.Equal(t, 5, attempts)
	require.Equal(t, 1, effects)
	replay, err := chat.CreateMessage(f.ctx, command)
	require.NoError(t, err)
	require.True(t, replay.Replayed)
	require.Equal(t, message.ID, replay.ID)
	require.Equal(t, 1, effects)
	assertRetryMessageCounts(t, f, 1)
}

func TestChatConcurrentWritesKeepUniqueSequencesAndAtomicOutbox(t *testing.T) {
	f := newFixture(t)
	ctx, cancel := context.WithTimeout(f.ctx, 20*time.Second)
	defer cancel()
	var effects atomic.Int64
	chat := service.New(f.repo, service.WithMessageCreatedHook(func(ctx context.Context, _ domain.CreateMessageCommand, _ domain.Message) error {
		platformdatabase.AfterCommit(ctx, func() { effects.Add(1) })
		return nil
	}))
	const writers, perWriter = 8, 3
	var workers sync.WaitGroup
	start := make(chan struct{})
	failures := make(chan error, writers)
	for writer := range writers {
		workers.Add(1)
		go func() {
			defer workers.Done()
			<-start
			for index := range perWriter {
				command := domain.CreateMessageCommand{Scope: f.scope, ConversationID: f.room, CreationID: uuid.New(), Text: fmt.Sprintf("Concurrent writer %d message %d", writer, index)}
				message, err := chat.CreateMessage(ctx, command)
				if err != nil {
					failures <- err
					return
				}
				replayed, err := chat.CreateMessage(ctx, command)
				if err != nil {
					failures <- err
					return
				}
				if !replayed.Replayed || replayed.ID != message.ID {
					failures <- fmt.Errorf("creation replay changed identity")
					return
				}
			}
		}()
	}
	close(start)
	workers.Wait()
	close(failures)
	for err := range failures {
		require.NoError(t, err)
	}
	require.Equal(t, int64(writers*perWriter), effects.Load())
	assertRetryMessageCounts(t, f, writers*perWriter)
}

func assertRetryMessageCounts(t *testing.T, f fixture, want int) {
	t.Helper()
	var messages, sequences, events, eventMessages int
	require.NoError(t, f.pool.QueryRow(f.ctx, `SELECT count(*),count(DISTINCT sequence) FROM chat_conversation_messages WHERE conversation_id=$1`, f.room).Scan(&messages, &sequences))
	require.NoError(t, f.pool.QueryRow(f.ctx, `SELECT count(*),count(DISTINCT message_id) FROM chat_outbox WHERE conversation_id=$1 AND event_type='message.created'`, f.room).Scan(&events, &eventMessages))
	require.Equal(t, want, messages)
	require.Equal(t, want, sequences)
	require.Equal(t, want, events)
	require.Equal(t, want, eventMessages)
}
