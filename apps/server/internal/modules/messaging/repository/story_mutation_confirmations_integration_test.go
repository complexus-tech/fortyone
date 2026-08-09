package messagingrepository

import (
	"bytes"
	"context"
	"errors"
	"os"
	"sync/atomic"
	"testing"
	"time"

	messaging "github.com/complexus-tech/projects-api/internal/modules/messaging/service"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

// TestStoryMutationConfirmationPostgresContract exercises the row-lock/CAS
// semantics against PostgreSQL when an isolated migrated database is supplied.
// It is skipped in ordinary unit-test runs.
func TestStoryMutationConfirmationPostgresContract(t *testing.T) {
	databaseURL := os.Getenv("MESSAGING_CONFIRMATION_TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("MESSAGING_CONFIRMATION_TEST_DATABASE_URL is not configured")
	}
	db, err := sqlx.Connect("postgres", databaseURL)
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	ctx := context.Background()
	repo := New(db)
	workspaceID, userID, teamID := seedStoryMutationConfirmationActor(t, db)
	now := time.Now().UTC().Truncate(time.Microsecond)

	bindingFor := func(confirmationID uuid.UUID, tokenHash []byte) messaging.StoryMutationConfirmationBinding {
		return messaging.StoryMutationConfirmationBinding{
			ConfirmationID: confirmationID,
			WorkspaceID:    workspaceID,
			UserID:         userID,
			TokenHash:      tokenHash,
		}
	}
	register := func(confirmationID uuid.UUID, tokenHash []byte, expiresAt time.Time) {
		t.Helper()
		require.NoError(t, repo.RegisterStoryMutationConfirmation(ctx, messaging.StoryMutationConfirmationStateInput{
			ConfirmationID: confirmationID,
			WorkspaceID:    workspaceID,
			UserID:         userID,
			TeamID:         teamID,
			Operation:      messaging.StoryMutationCreate,
			TokenHash:      tokenHash,
			ExpiresAt:      expiresAt,
		}))
	}
	result := messaging.StoryMutationResult{
		Status:    "applied",
		Operation: messaging.StoryMutationCreate,
		StoryID:   uuid.New(),
		Reference: "WEB-1",
		TeamID:    teamID,
		Title:     "Durable confirmation",
		Priority:  "High",
	}

	t.Run("cancel then confirm cannot write", func(t *testing.T) {
		confirmationID := uuid.New()
		tokenHash := bytes.Repeat([]byte{1}, sha256DigestSize)
		register(confirmationID, tokenHash, now.Add(time.Minute))
		binding := bindingFor(confirmationID, tokenHash)

		cancelled, err := repo.CancelStoryMutationConfirmation(ctx, binding, now)
		require.NoError(t, err)
		require.Equal(t, storyMutationCancellationStatusCancelled, cancelled.Status)
		cancelled, err = repo.CancelStoryMutationConfirmation(ctx, binding, now)
		require.NoError(t, err)
		require.Equal(t, storyMutationCancellationStatusAlreadyCancelled, cancelled.Status)

		var applyCalls atomic.Int32
		_, _, err = repo.ApplyStoryMutationConfirmation(ctx, binding, now, func(context.Context) (messaging.StoryMutationResult, error) {
			applyCalls.Add(1)
			return result, nil
		})
		require.ErrorIs(t, err, messaging.ErrCancelledConfirmation)
		require.Zero(t, applyCalls.Load())
	})

	t.Run("confirm persists one result and cannot be cancelled", func(t *testing.T) {
		confirmationID := uuid.New()
		tokenHash := bytes.Repeat([]byte{2}, sha256DigestSize)
		register(confirmationID, tokenHash, now.Add(time.Minute))
		binding := bindingFor(confirmationID, tokenHash)

		var applyCalls atomic.Int32
		applied, duplicate, err := repo.ApplyStoryMutationConfirmation(ctx, binding, now, func(context.Context) (messaging.StoryMutationResult, error) {
			applyCalls.Add(1)
			return result, nil
		})
		require.NoError(t, err)
		require.False(t, duplicate)
		require.Equal(t, result.StoryID, applied.StoryID)

		_, duplicate, err = repo.ApplyStoryMutationConfirmation(ctx, binding, now, func(context.Context) (messaging.StoryMutationResult, error) {
			applyCalls.Add(1)
			return messaging.StoryMutationResult{}, errors.New("duplicate callback must not run")
		})
		require.NoError(t, err)
		require.True(t, duplicate)
		require.EqualValues(t, 1, applyCalls.Load())
		_, err = repo.CancelStoryMutationConfirmation(ctx, binding, now)
		require.ErrorIs(t, err, messaging.ErrAppliedConfirmation)
	})

	t.Run("expiry and actor binding are terminal", func(t *testing.T) {
		confirmationID := uuid.New()
		tokenHash := bytes.Repeat([]byte{3}, sha256DigestSize)
		register(confirmationID, tokenHash, now)
		binding := bindingFor(confirmationID, tokenHash)

		wrongActor := binding
		wrongActor.UserID = uuid.New()
		_, err := repo.CancelStoryMutationConfirmation(ctx, wrongActor, now)
		require.ErrorIs(t, err, messaging.ErrInvalidConfirmation)
		_, err = repo.CancelStoryMutationConfirmation(ctx, binding, now)
		require.ErrorIs(t, err, messaging.ErrExpiredConfirmation)
		_, _, err = repo.ApplyStoryMutationConfirmation(ctx, binding, now.Add(-time.Minute), func(context.Context) (messaging.StoryMutationResult, error) {
			t.Fatal("expired callback must not run")
			return messaging.StoryMutationResult{}, nil
		})
		require.ErrorIs(t, err, messaging.ErrExpiredConfirmation)
	})
}

func seedStoryMutationConfirmationActor(t *testing.T, db *sqlx.DB) (uuid.UUID, uuid.UUID, uuid.UUID) {
	t.Helper()
	ctx := context.Background()
	userID := uuid.New()
	workspaceID := uuid.New()
	teamID := uuid.New()
	suffix := uuid.NewString()
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), "DELETE FROM workspaces WHERE workspace_id = $1", workspaceID)
		_, _ = db.ExecContext(context.Background(), "DELETE FROM users WHERE user_id = $1", userID)
	})
	_, err := db.ExecContext(ctx, `
		INSERT INTO users (user_id, username, email)
		VALUES ($1, $2, $3)
	`, userID, "confirmation-"+suffix, "confirmation-"+suffix+"@example.test")
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `
		INSERT INTO workspaces (workspace_id, name, slug, created_by)
		VALUES ($1, 'Confirmation Test', $2, $3)
	`, workspaceID, "confirmation-"+suffix, userID)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `
		INSERT INTO teams (team_id, workspace_id, name, code, color)
		VALUES ($1, $2, 'Web', $3, '#111827')
	`, teamID, workspaceID, "WEB-"+suffix)
	require.NoError(t, err)
	return workspaceID, userID, teamID
}
