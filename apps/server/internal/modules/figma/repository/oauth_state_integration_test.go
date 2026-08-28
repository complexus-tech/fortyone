//go:build integration

package figmarepository

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	figmadomain "github.com/complexus-tech/projects-api/internal/modules/figma/domain"
	"github.com/complexus-tech/projects-api/internal/testkit"
	"github.com/google/uuid"
)

func TestFigmaOAuthStatePostgresIsAtomicAndExpirySafe(t *testing.T) {
	testDatabase := testkit.NewPostgres(t)
	ctx := t.Context()
	workspaceID, userID, workspaceSlug := insertFigmaOAuthStateOwner(t, testDatabase, ctx)
	repository := New(testDatabase.Pool)
	now := time.Now().UTC()
	state := figmadomain.OAuthState{
		StateHash:     "figma-state-" + uuid.NewString(),
		WorkspaceID:   workspaceID,
		UserID:        userID,
		WorkspaceSlug: workspaceSlug,
		CodeVerifier:  "pkce-verifier",
		ExpiresAt:     now.Add(time.Minute),
	}
	if err := repository.SaveOAuthState(ctx, state); err != nil {
		t.Fatalf("SaveOAuthState() error = %v", err)
	}

	const consumers = 24
	var successes atomic.Int32
	var wait sync.WaitGroup
	errorsByAttempt := make(chan error, consumers)
	wait.Add(consumers)
	for range consumers {
		go func() {
			defer wait.Done()
			consumed, consumeErr := repository.ConsumeOAuthState(context.Background(), state.StateHash, now)
			if consumeErr == nil {
				if consumed.WorkspaceID != workspaceID || consumed.UserID != userID || consumed.CodeVerifier != state.CodeVerifier {
					errorsByAttempt <- errors.New("consumed Figma state lost its identity or PKCE binding")
					return
				}
				successes.Add(1)
				return
			}
			if !errors.Is(consumeErr, figmadomain.ErrNotFound) {
				errorsByAttempt <- consumeErr
			}
		}()
	}
	wait.Wait()
	close(errorsByAttempt)
	for consumeErr := range errorsByAttempt {
		t.Errorf("concurrent ConsumeOAuthState() error = %v", consumeErr)
	}
	if got := successes.Load(); got != 1 {
		t.Fatalf("successful consumers = %d, want 1", got)
	}

	expired := state
	expired.StateHash = "figma-expired-" + uuid.NewString()
	expired.ExpiresAt = now.Add(-time.Second)
	if err := repository.SaveOAuthState(ctx, expired); err != nil {
		t.Fatalf("save expired OAuth state: %v", err)
	}
	if _, err := repository.ConsumeOAuthState(ctx, expired.StateHash, now); !errors.Is(err, figmadomain.ErrNotFound) {
		t.Fatalf("expired state error = %v, want domain not found", err)
	}
	if _, err := repository.ConsumeOAuthState(ctx, "wrong-state-hash", now); !errors.Is(err, figmadomain.ErrNotFound) {
		t.Fatalf("mismatched state error = %v, want domain not found", err)
	}
}

func insertFigmaOAuthStateOwner(
	t *testing.T,
	databaseState *testkit.Postgres,
	ctx context.Context,
) (uuid.UUID, uuid.UUID, string) {
	t.Helper()
	workspaceID, userID := uuid.New(), uuid.New()
	suffix := uuid.NewString()
	if _, err := databaseState.Pool.Exec(ctx, `
		INSERT INTO users (user_id, username, email, full_name)
		VALUES ($1, $2, $3, 'Figma OAuth state test')
	`, userID, "figma-oauth-state-"+suffix, "figma-oauth-state-"+suffix+"@example.test"); err != nil {
		t.Fatalf("insert Figma OAuth state user: %v", err)
	}
	workspaceSlug := "figma-oauth-state-" + suffix
	if _, err := databaseState.Pool.Exec(ctx, `
		INSERT INTO workspaces (workspace_id, name, slug, created_by)
		VALUES ($1, 'Figma OAuth state test', $2, $3)
	`, workspaceID, workspaceSlug, userID); err != nil {
		t.Fatalf("insert Figma OAuth state workspace: %v", err)
	}
	if _, err := databaseState.Pool.Exec(ctx, `
		INSERT INTO workspace_members (workspace_id, user_id, role)
		VALUES ($1, $2, 'admin')
	`, workspaceID, userID); err != nil {
		t.Fatalf("insert Figma OAuth state membership: %v", err)
	}
	return workspaceID, userID, workspaceSlug
}
