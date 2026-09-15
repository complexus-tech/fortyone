//go:build integration

package usersrepository

import (
	"context"
	"sync"
	"testing"
	"time"

	usersdomain "github.com/complexus-tech/projects-api/internal/modules/users/domain"
	usersql "github.com/complexus-tech/projects-api/internal/modules/users/repository/sqlc"
	"github.com/complexus-tech/projects-api/internal/testkit"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestSubscriberDeletionClaimLeaseRetryAndPurge(t *testing.T) {
	db := testkit.NewPostgres(t)
	repository := NewSubscriberDeletionRepository(db.Pool)
	now := time.Now().UTC()
	_, err := db.Pool.Exec(t.Context(), `INSERT INTO account_subscriber_deletions(id,email,next_attempt_at) VALUES($1,$2,$3)`, uuid.New(), "deleted@example.com", now)
	require.NoError(t, err)
	var mutex sync.Mutex
	var claims []usersdomain.SubscriberDeletion
	var failures []error
	var wait sync.WaitGroup
	for range 4 {
		wait.Go(func() {
			job, claimed, err := repository.Claim(t.Context(), now, now.Add(time.Minute))
			mutex.Lock()
			defer mutex.Unlock()
			if err != nil {
				failures = append(failures, err)
			}
			if claimed {
				claims = append(claims, job)
			}
		})
	}
	wait.Wait()
	require.Empty(t, failures)
	require.Len(t, claims, 1)
	old := claims[0]
	require.Error(t, repository.Complete(t.Context(), old.ID, uuid.New()))
	renewed, err := repository.Renew(t.Context(), old.ID, old.LeaseToken, now.Add(2*time.Minute), now.Add(3*time.Minute))
	require.NoError(t, err)
	require.False(t, renewed)
	current, claimed, err := repository.Claim(t.Context(), now.Add(2*time.Minute), now.Add(3*time.Minute))
	require.NoError(t, err)
	require.True(t, claimed)
	require.NotEqual(t, old.LeaseToken, current.LeaseToken)
	require.Error(t, repository.Complete(t.Context(), old.ID, old.LeaseToken))
	require.NoError(t, repository.Retry(t.Context(), current.ID, current.LeaseToken, now.Add(4*time.Minute), now.Add(2*time.Minute)))
	_, claimed, err = repository.Claim(t.Context(), now.Add(3*time.Minute), now.Add(5*time.Minute))
	require.NoError(t, err)
	require.False(t, claimed)
	current, claimed, err = repository.Claim(t.Context(), now.Add(4*time.Minute), now.Add(5*time.Minute))
	require.NoError(t, err)
	require.True(t, claimed)
	require.NoError(t, repository.Complete(t.Context(), current.ID, current.LeaseToken))
	var remaining int
	require.NoError(t, db.Pool.QueryRow(t.Context(), `SELECT count(*) FROM account_subscriber_deletions`).Scan(&remaining))
	require.Zero(t, remaining)
}

func TestSubscriberLifecycleFencesCoreDeletionWithoutRemoteTransaction(t *testing.T) {
	db := testkit.NewPostgres(t)
	repository := NewSubscriberDeletionRepository(db.Pool)
	ctx := t.Context()
	require.NoError(t, repository.WithinSubscriberLifecycle(ctx, " Member@Example.com ", func(locked context.Context) error {
		// A separate session uses the exact xact key used by account deletion.
		tx, err := db.Pool.Begin(ctx)
		require.NoError(t, err)
		defer tx.Rollback(ctx)
		var acquired bool
		require.NoError(t, tx.QueryRow(ctx, `SELECT pg_try_advisory_xact_lock(hashtextextended('subscriber:' || lower(btrim($1)),0))`, "member@example.com").Scan(&acquired))
		require.False(t, acquired)
		_, found, err := repository.ActiveAccount(locked, "member@example.com")
		require.NoError(t, err)
		require.False(t, found)
		return nil
	}))
	tx, err := db.Pool.Begin(ctx)
	require.NoError(t, err)
	defer tx.Rollback(ctx)
	lockCtx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	require.NoError(t, usersql.New(tx).LockAccountSubscriberLifecycle(lockCtx, usersql.LockAccountSubscriberLifecycleParams{Email: "member@example.com"}))
}
