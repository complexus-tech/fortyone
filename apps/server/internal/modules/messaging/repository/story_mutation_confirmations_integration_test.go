package messagingrepository

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"sync/atomic"
	"testing"
	"time"

	messaging "github.com/complexus-tech/projects-api/internal/modules/messaging/service"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
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
	pool, err := pgxpool.New(context.Background(), databaseURL)
	require.NoError(t, err)
	require.NoError(t, pool.Ping(context.Background()))
	t.Cleanup(pool.Close)

	ctx := context.Background()
	repo := New(pool)
	workspaceID, userID, teamID := seedStoryMutationConfirmationActor(t, pool)
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

	t.Run("batch cancellation and expiry redact persisted proposals", func(t *testing.T) {
		proposal := []byte(`{"version":1,"items":[{"title":"Sensitive context","description":"From Slack","priority":"High"}]}`)
		registerBatch := func(confirmationID uuid.UUID, tokenHash []byte, expiresAt time.Time) messaging.StoryMutationConfirmationBinding {
			t.Helper()
			require.NoError(t, repo.RegisterStoryMutationConfirmation(ctx, messaging.StoryMutationConfirmationStateInput{
				ConfirmationID: confirmationID,
				WorkspaceID:    workspaceID,
				UserID:         userID,
				TeamID:         teamID,
				Operation:      messaging.StoryMutationCreateBatch,
				TokenHash:      tokenHash,
				Proposal:       proposal,
				ExpiresAt:      expiresAt,
			}))
			return bindingFor(confirmationID, tokenHash)
		}
		assertProposalRedacted := func(confirmationID uuid.UUID) {
			t.Helper()
			var persistedProposal []byte
			require.NoError(t, pool.QueryRow(ctx, `
				SELECT proposal
				FROM messaging_story_mutation_confirmations
				WHERE confirmation_id = $1
			`, confirmationID).Scan(&persistedProposal))
			require.Empty(t, persistedProposal)
		}

		cancelledID := uuid.New()
		cancelledBinding := registerBatch(cancelledID, bytes.Repeat([]byte{5}, sha256DigestSize), now.Add(time.Minute))
		_, err := repo.CancelStoryMutationConfirmation(ctx, cancelledBinding, now)
		require.NoError(t, err)
		assertProposalRedacted(cancelledID)

		expiredID := uuid.New()
		expiredBinding := registerBatch(expiredID, bytes.Repeat([]byte{6}, sha256DigestSize), now)
		_, _, err = repo.ApplyStoryMutationConfirmation(ctx, expiredBinding, now, func(context.Context) (messaging.StoryMutationResult, error) {
			t.Fatal("expired batch callback must not run")
			return messaging.StoryMutationResult{}, nil
		})
		require.ErrorIs(t, err, messaging.ErrExpiredConfirmation)
		assertProposalRedacted(expiredID)

		retryExpiredID := uuid.New()
		retryExpiredBinding := registerBatch(retryExpiredID, bytes.Repeat([]byte{7}, sha256DigestSize), now.Add(time.Minute))
		partialResult := messaging.StoryMutationResult{
			Status: "partial", Operation: messaging.StoryMutationCreateBatch, TeamID: teamID,
		}
		_, _, err = repo.ApplyStoryMutationConfirmation(ctx, retryExpiredBinding, now, func(context.Context) (messaging.StoryMutationResult, error) {
			return partialResult, errors.New("provider failure")
		})
		require.ErrorContains(t, err, "provider failure")
		_, _, err = repo.ApplyStoryMutationConfirmation(ctx, retryExpiredBinding, now.Add(time.Minute), func(context.Context) (messaging.StoryMutationResult, error) {
			t.Fatal("expired retry callback must not run")
			return messaging.StoryMutationResult{}, nil
		})
		require.ErrorIs(t, err, messaging.ErrExpiredConfirmation)
		assertProposalRedacted(retryExpiredID)
	})

	t.Run("opaque batch proposal is actor-bound and persisted with itemized result", func(t *testing.T) {
		confirmationID := uuid.New()
		tokenHash := bytes.Repeat([]byte{4}, sha256DigestSize)
		proposal := []byte(`{"version":1,"source_url":"https://acme.slack.com/archives/C1/p1","items":[{"title":"First","priority":"High"},{"title":"Second","priority":"Low"}]}`)
		require.NoError(t, repo.RegisterStoryMutationConfirmation(ctx, messaging.StoryMutationConfirmationStateInput{
			ConfirmationID: confirmationID,
			WorkspaceID:    workspaceID,
			UserID:         userID,
			TeamID:         teamID,
			Operation:      messaging.StoryMutationCreateBatch,
			TokenHash:      tokenHash,
			Proposal:       proposal,
			ExpiresAt:      now.Add(time.Minute),
		}))
		binding := bindingFor(confirmationID, tokenHash)
		loaded, err := repo.LoadStoryMutationConfirmation(ctx, binding)
		require.NoError(t, err)
		require.Equal(t, teamID, loaded.TeamID)
		require.Equal(t, messaging.StoryMutationCreateBatch, loaded.Operation)
		require.JSONEq(t, string(proposal), string(loaded.Proposal))

		wrongActor := binding
		wrongActor.UserID = uuid.New()
		_, err = repo.LoadStoryMutationConfirmation(ctx, wrongActor)
		require.ErrorIs(t, err, messaging.ErrInvalidConfirmation)

		partialResult := messaging.StoryMutationResult{
			Status:    "partial",
			Operation: messaging.StoryMutationCreateBatch,
			TeamID:    teamID,
			Items: []messaging.StoryMutationItemResult{{
				Index:     0,
				Status:    "applied",
				StoryID:   uuid.New(),
				Reference: "WEB-2",
				TeamID:    teamID,
				Title:     "First",
				Priority:  "High",
			}},
		}
		partial, duplicate, err := repo.ApplyStoryMutationConfirmation(ctx, binding, now, func(context.Context) (messaging.StoryMutationResult, error) {
			return partialResult, errors.New("transient provider failure")
		})
		require.ErrorContains(t, err, "transient provider failure")
		require.False(t, duplicate)
		require.Equal(t, partialResult.Items, partial.Items)
		var retainedProposal []byte
		var retainedResult []byte
		var lastError *string
		require.NoError(t, pool.QueryRow(ctx, `
			SELECT proposal, result, last_error
			FROM messaging_story_mutation_confirmations
			WHERE confirmation_id = $1
		`, confirmationID).Scan(&retainedProposal, &retainedResult, &lastError))
		require.JSONEq(t, string(proposal), string(retainedProposal))
		require.JSONEq(t, mustJSON(t, partialResult), string(retainedResult))
		require.NotNil(t, lastError)

		batchResult := messaging.StoryMutationResult{
			Status:    "applied",
			Operation: messaging.StoryMutationCreateBatch,
			TeamID:    teamID,
			Items: append(partialResult.Items, messaging.StoryMutationItemResult{
				Index: 1, Status: "applied", StoryID: uuid.New(), Reference: "WEB-3",
				TeamID: teamID, Title: "Second", Priority: "Low",
			}),
		}
		applied, duplicate, err := repo.ApplyStoryMutationConfirmation(ctx, binding, now, func(context.Context) (messaging.StoryMutationResult, error) {
			return batchResult, nil
		})
		require.NoError(t, err)
		require.False(t, duplicate)
		require.Equal(t, batchResult.Items, applied.Items)
		completed, err := repo.LoadStoryMutationConfirmation(ctx, binding)
		require.NoError(t, err)
		require.Empty(t, completed.Proposal)
		require.Empty(t, completed.LastError)
		require.NotNil(t, completed.Result)
		require.Equal(t, batchResult.Items, completed.Result.Items)
	})
}

func mustJSON(t *testing.T, value any) string {
	t.Helper()
	payload, err := json.Marshal(value)
	require.NoError(t, err)
	return string(payload)
}

func seedStoryMutationConfirmationActor(t *testing.T, pool *pgxpool.Pool) (uuid.UUID, uuid.UUID, uuid.UUID) {
	t.Helper()
	ctx := context.Background()
	userID := uuid.New()
	workspaceID := uuid.New()
	teamID := uuid.New()
	suffix := uuid.NewString()
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), "DELETE FROM workspaces WHERE workspace_id = $1", workspaceID)
		_, _ = pool.Exec(context.Background(), "DELETE FROM users WHERE user_id = $1", userID)
	})
	_, err := pool.Exec(ctx, `
		INSERT INTO users (user_id, username, email)
		VALUES ($1, $2, $3)
	`, userID, "confirmation-"+suffix, "confirmation-"+suffix+"@example.test")
	require.NoError(t, err)
	_, err = pool.Exec(ctx, `
		INSERT INTO workspaces (workspace_id, name, slug, created_by)
		VALUES ($1, 'Confirmation Test', $2, $3)
	`, workspaceID, "confirmation-"+suffix, userID)
	require.NoError(t, err)
	_, err = pool.Exec(ctx, `
		INSERT INTO teams (team_id, workspace_id, name, code, color)
		VALUES ($1, $2, 'Web', $3, '#111827')
	`, teamID, workspaceID, "WEB-"+suffix)
	require.NoError(t, err)
	return workspaceID, userID, teamID
}
