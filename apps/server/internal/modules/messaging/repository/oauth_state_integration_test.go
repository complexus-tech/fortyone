//go:build integration

package messagingrepository

import (
	"context"
	"crypto/rand"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/complexus-tech/projects-api/internal/testkit"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func TestProviderOAuthNoncePostgresIsPurposeBoundAndAtomic(t *testing.T) {
	testDatabase := testkit.NewPostgres(t)
	ctx := t.Context()
	workspaceID, userID := insertProviderOAuthNonceOwner(t, testDatabase, ctx)
	repository := New(testDatabase.Pool)
	now := time.Now().UTC()
	digest := make([]byte, 32)
	if _, err := rand.Read(digest); err != nil {
		t.Fatalf("generate test nonce digest: %v", err)
	}
	if err := repository.CreateNonce(ctx, NonceInput{
		Provider:    "slack",
		Purpose:     "oauth_install",
		NonceHash:   digest,
		WorkspaceID: workspaceID,
		UserID:      &userID,
		ExpiresAt:   now.Add(time.Minute),
	}); err != nil {
		t.Fatalf("CreateNonce() error = %v", err)
	}

	wrongWorkspace := uuid.New()
	for _, mismatch := range []NonceConsumeInput{
		{Provider: "github", Purpose: "oauth_install", NonceHash: digest, Now: now},
		{Provider: "slack", Purpose: "account_link", NonceHash: digest, Now: now},
		{Provider: "slack", Purpose: "oauth_install", NonceHash: digest, WorkspaceID: &wrongWorkspace, Now: now},
	} {
		if _, err := repository.ConsumeNonce(ctx, mismatch); !errors.Is(err, pgx.ErrNoRows) {
			t.Fatalf("mismatched ConsumeNonce() error = %v, want pgx.ErrNoRows", err)
		}
	}

	const consumers = 24
	var successes atomic.Int32
	var wait sync.WaitGroup
	errorsByAttempt := make(chan error, consumers)
	wait.Add(consumers)
	for range consumers {
		go func() {
			defer wait.Done()
			record, consumeErr := repository.ConsumeNonce(context.Background(), NonceConsumeInput{
				Provider:    "slack",
				Purpose:     "oauth_install",
				NonceHash:   digest,
				WorkspaceID: &workspaceID,
				UserID:      &userID,
				Now:         now,
			})
			if consumeErr == nil {
				if record.WorkspaceID != workspaceID || record.UserID == nil || *record.UserID != userID {
					errorsByAttempt <- errors.New("consumed provider nonce lost its identity binding")
					return
				}
				successes.Add(1)
				return
			}
			if !errors.Is(consumeErr, pgx.ErrNoRows) {
				errorsByAttempt <- consumeErr
			}
		}()
	}
	wait.Wait()
	close(errorsByAttempt)
	for consumeErr := range errorsByAttempt {
		t.Errorf("concurrent ConsumeNonce() error = %v", consumeErr)
	}
	if got := successes.Load(); got != 1 {
		t.Fatalf("successful consumers = %d, want 1", got)
	}
}

func insertProviderOAuthNonceOwner(t *testing.T, databaseState *testkit.Postgres, ctx context.Context) (uuid.UUID, uuid.UUID) {
	t.Helper()
	workspaceID, userID := uuid.New(), uuid.New()
	suffix := uuid.NewString()
	if _, err := databaseState.Pool.Exec(ctx, `
		INSERT INTO users (user_id, username, email, full_name)
		VALUES ($1, $2, $3, 'Provider OAuth nonce test')
	`, userID, "provider-oauth-nonce-"+suffix, "provider-oauth-nonce-"+suffix+"@example.test"); err != nil {
		t.Fatalf("insert provider OAuth nonce user: %v", err)
	}
	if _, err := databaseState.Pool.Exec(ctx, `
		INSERT INTO workspaces (workspace_id, name, slug, created_by)
		VALUES ($1, 'Provider OAuth nonce test', $2, $3)
	`, workspaceID, "provider-oauth-nonce-"+suffix, userID); err != nil {
		t.Fatalf("insert provider OAuth nonce workspace: %v", err)
	}
	return workspaceID, userID
}
