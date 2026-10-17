//go:build integration

package workspacesecurityrepository_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	workspaces "github.com/complexus-tech/projects-api/internal/modules/workspaces/repository"
	domain "github.com/complexus-tech/projects-api/internal/modules/workspacesecurity/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func sessionVersion(t *testing.T, f fixture, id uuid.UUID) string {
	t.Helper()
	var version string
	require.NoError(t, f.pool.QueryRow(t.Context(), `SELECT CAST(xmin AS text) FROM workspace_browser_sessions WHERE workspace_id=$1 AND session_id=$2`, f.scope.WorkspaceID, id).Scan(&version))
	return version
}

func TestFreshSessionChecksAvoidWritesButKeepAuthoritativeRevocation(t *testing.T) {
	f := seed(t)
	scope := domain.Scope{WorkspaceID: f.scope.WorkspaceID, ActorID: f.member}
	session := identity()
	require.NoError(t, f.repo.CheckSession(t.Context(), scope, session))
	initial := sessionVersion(t, f, session.ID)
	for range 10 {
		require.NoError(t, f.repo.CheckSession(t.Context(), scope, session))
	}
	require.Equal(t, initial, sessionVersion(t, f, session.ID), "fresh auth hits must not update the registry tuple")

	other := scope
	other.ActorID = f.guest
	require.ErrorIs(t, f.repo.CheckSession(t.Context(), other, session), domain.ErrForbidden, "fresh path must verify session owner")
	changedIdentity := session
	changedIdentity.AuthenticatedAt = session.AuthenticatedAt.Add(time.Second)
	require.ErrorIs(t, f.repo.CheckSession(t.Context(), scope, changedIdentity), domain.ErrForbidden, "fresh path must verify immutable issuance time")

	exec(t, f.pool, `UPDATE workspace_browser_sessions SET last_seen_at=CURRENT_TIMESTAMP-INTERVAL '2 minutes' WHERE workspace_id=$1 AND session_id=$2`, scope.WorkspaceID, session.ID)
	aged := sessionVersion(t, f, session.ID)
	require.NoError(t, f.repo.CheckSession(t.Context(), scope, session))
	require.NotEqual(t, aged, sessionVersion(t, f, session.ID), "aged telemetry must refresh")

	beforeExpiry := sessionVersion(t, f, session.ID)
	session.ExpiresAt = session.ExpiresAt.Add(time.Hour)
	require.NoError(t, f.repo.CheckSession(t.Context(), scope, session))
	require.NotEqual(t, beforeExpiry, sessionVersion(t, f, session.ID), "changed expiry must be persisted")
	require.NoError(t, f.repo.RevokeSession(t.Context(), f.scope, session.ID, "test revocation"))
	require.ErrorIs(t, f.repo.CheckSession(t.Context(), scope, session), domain.ErrForbidden)

	next := identity()
	require.NoError(t, f.repo.CheckSession(t.Context(), scope, next))
	require.NoError(t, f.repo.RevokeMember(t.Context(), f.scope, f.member, "test member revocation"))
	require.ErrorIs(t, f.repo.CheckSession(t.Context(), scope, next), domain.ErrForbidden)
}

func TestConcurrentSessionChecksAndActivityTouchesDoNotSerializeFreshWorkspace(t *testing.T) {
	f := seed(t)
	ctx, cancel := context.WithTimeout(t.Context(), 20*time.Second)
	defer cancel()
	workspaceRepository := workspaces.New(f.pool)
	people := []uuid.UUID{f.scope.ActorID, f.member, f.guest}
	sessions := make([]domain.SessionIdentity, len(people))
	for i, person := range people {
		sessions[i] = identity()
		require.NoError(t, f.repo.CheckSession(ctx, domain.Scope{WorkspaceID: f.scope.WorkspaceID, ActorID: person}, sessions[i]))
		exec(t, f.pool, `UPDATE users SET last_login_at=CURRENT_TIMESTAMP-INTERVAL '2 minutes' WHERE user_id=$1`, person)
	}
	exec(t, f.pool, `UPDATE workspaces SET last_accessed_at=(CURRENT_TIMESTAMP AT TIME ZONE 'UTC')-INTERVAL '2 minutes' WHERE workspace_id=$1`, f.scope.WorkspaceID)
	var workers sync.WaitGroup
	failures := make(chan error, 24)
	start := make(chan struct{})
	for worker := range 24 {
		workers.Add(1)
		go func() {
			defer workers.Done()
			<-start
			index := worker % len(people)
			scope := domain.Scope{WorkspaceID: f.scope.WorkspaceID, ActorID: people[index]}
			for range 10 {
				if err := f.repo.CheckSession(ctx, scope, sessions[index]); err != nil {
					failures <- err
					return
				}
				if err := workspaceRepository.RecordAccess(ctx, scope.WorkspaceID, scope.ActorID); err != nil {
					failures <- err
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

	// Revocation competes with fresh reads through the same workspace lock.
	// Once it returns, no subsequent fresh-path read may accept that session.
	member := domain.Scope{WorkspaceID: f.scope.WorkspaceID, ActorID: f.member}
	for range 8 {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for range 10 {
				err := f.repo.CheckSession(ctx, member, sessions[1])
				if err != nil && !errors.Is(err, domain.ErrForbidden) {
					t.Errorf("concurrent policy read: %v", err)
					return
				}
			}
		}()
	}
	require.NoError(t, f.repo.RevokeSession(ctx, f.scope, sessions[1].ID, "concurrent test revocation"))
	workers.Wait()
	for range 10 {
		require.ErrorIs(t, f.repo.CheckSession(ctx, member, sessions[1]), domain.ErrForbidden)
	}
}
