package chatsessionsrepository

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	chatsessions "github.com/complexus-tech/projects-api/internal/modules/chatsessions/service"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

// TestChatPersistenceAndMutationApprovalPostgresContract exercises the
// owner-scoped upsert and durable approval state machine against PostgreSQL
// when an isolated migrated database is supplied. It is skipped in ordinary
// unit-test runs.
func TestChatPersistenceAndMutationApprovalPostgresContract(t *testing.T) {
	databaseURL := os.Getenv("CHATSESSIONS_TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("CHATSESSIONS_TEST_DATABASE_URL is not configured")
	}

	db, err := sqlx.Connect("postgres", databaseURL)
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	ctx := context.Background()
	owner := seedChatSessionActor(t, db)
	other := seedChatSessionActor(t, db)
	repository := New(
		logger.NewWithText(io.Discard, slog.LevelError, "chatsessions-test"),
		db,
	)
	sessionID := chatSessionTestID()

	t.Cleanup(func() {
		_, cleanupErr := db.ExecContext(ctx, `DELETE FROM workspaces WHERE workspace_id IN ($1, $2)`, owner.workspaceID, other.workspaceID)
		require.NoError(t, cleanupErr)
		_, cleanupErr = db.ExecContext(ctx, `DELETE FROM users WHERE user_id IN ($1, $2)`, owner.userID, other.userID)
		require.NoError(t, cleanupErr)
	})

	t.Run("session upsert is idempotent, concurrent, and owner scoped", func(t *testing.T) {
		session := chatsessions.CoreChatSession{
			ID:          sessionID,
			UserID:      owner.userID,
			WorkspaceID: owner.workspaceID,
			Title:       "Original title",
		}
		firstMessages := []any{map[string]any{"id": "message-1", "role": "user"}}
		created, err := repository.CreateSessionWithMessages(ctx, &session, firstMessages)
		require.NoError(t, err)
		require.Equal(t, session.Title, created.Title)

		session.Title = "A retry must not replace the established title"
		secondMessages := []any{map[string]any{"id": "message-2", "role": "assistant"}}
		persisted, err := repository.CreateSessionWithMessages(ctx, &session, secondMessages)
		require.NoError(t, err)
		require.Equal(t, "Original title", persisted.Title)
		require.Equal(t, firstMessages, mustGetChatMessages(t, repository, sessionID, owner))
		err = repository.SaveMessages(ctx, sessionID, owner.userID, owner.workspaceID, secondMessages)
		require.ErrorIs(t, err, chatsessions.ErrMessageWriteConflict)

		for _, foreignScope := range []chatSessionTestActor{
			{userID: other.userID, workspaceID: owner.workspaceID},
			{userID: owner.userID, workspaceID: other.workspaceID},
		} {
			wrongScopeSession := session
			wrongScopeSession.UserID = foreignScope.userID
			wrongScopeSession.WorkspaceID = foreignScope.workspaceID
			_, err = repository.CreateSessionWithMessages(ctx, &wrongScopeSession, []any{map[string]any{"id": "foreign"}})
			require.ErrorIs(t, err, chatsessions.ErrNotFound)
		}
		require.Equal(t, firstMessages, mustGetChatMessages(t, repository, sessionID, owner))

		const writers = 8
		errorsByWriter := make(chan error, writers)
		var waitGroup sync.WaitGroup
		for writer := 0; writer < writers; writer++ {
			waitGroup.Add(1)
			go func(writer int) {
				defer waitGroup.Done()
				concurrentSession := session
				_, writeErr := repository.CreateSessionWithMessages(
					ctx,
					&concurrentSession,
					[]any{map[string]any{"writer": writer}},
				)
				errorsByWriter <- writeErr
			}(writer)
		}
		waitGroup.Wait()
		close(errorsByWriter)
		for writeErr := range errorsByWriter {
			require.NoError(t, writeErr)
		}
		require.Equal(t, firstMessages, mustGetChatMessages(t, repository, sessionID, owner))
	})

	t.Run("message write generations reject stale finalization", func(t *testing.T) {
		writeSessionID := chatSessionTestID()
		writeSession := chatsessions.CoreChatSession{
			ID:          writeSessionID,
			UserID:      owner.userID,
			WorkspaceID: owner.workspaceID,
			Title:       "Generation contract",
		}
		userMessage := chatTestMessage("user-1", "user", map[string]any{
			"text": "Create the launch story",
			"type": "text",
		})
		firstReservation, err := repository.BeginMessageWrite(ctx, chatsessions.BeginMessageWriteParams{
			Session:   writeSession,
			Messages:  []any{userMessage},
			Operation: chatsessions.MessageWriteAppend,
		})
		require.NoError(t, err)
		secondReservation, err := repository.BeginMessageWrite(ctx, chatsessions.BeginMessageWriteParams{
			Session:   writeSession,
			Messages:  []any{userMessage},
			Operation: chatsessions.MessageWriteAppend,
		})
		require.NoError(t, err)
		require.Greater(t, secondReservation.Generation, firstReservation.Generation)
		require.NotEqual(t, firstReservation.Token, secondReservation.Token)

		finishedMessages := []any{
			userMessage,
			chatTestMessage("assistant-1", "assistant", map[string]any{
				"text": "The story is ready.",
				"type": "text",
			}),
		}
		staleResult, err := repository.FinalizeMessageWrite(ctx, chatsessions.FinalizeMessageWriteParams{
			SessionID:   writeSessionID,
			UserID:      owner.userID,
			WorkspaceID: owner.workspaceID,
			Messages:    finishedMessages,
			Generation:  firstReservation.Generation,
			Token:       firstReservation.Token,
		})
		require.NoError(t, err)
		require.False(t, staleResult.Applied)
		require.Equal(t, []any{userMessage}, mustGetChatMessages(t, repository, writeSessionID, owner))

		applied, err := repository.FinalizeMessageWrite(ctx, chatsessions.FinalizeMessageWriteParams{
			SessionID:   writeSessionID,
			UserID:      owner.userID,
			WorkspaceID: owner.workspaceID,
			Messages:    finishedMessages,
			Generation:  secondReservation.Generation,
			Token:       secondReservation.Token,
		})
		require.NoError(t, err)
		require.True(t, applied.Applied)
		replayed, err := repository.FinalizeMessageWrite(ctx, chatsessions.FinalizeMessageWriteParams{
			SessionID:   writeSessionID,
			UserID:      owner.userID,
			WorkspaceID: owner.workspaceID,
			Messages:    finishedMessages,
			Generation:  secondReservation.Generation,
			Token:       secondReservation.Token,
		})
		require.NoError(t, err)
		require.True(t, replayed.Applied)
		require.Equal(t, finishedMessages, mustGetChatMessages(t, repository, writeSessionID, owner))
	})

	t.Run("generation zero legacy approvals can enter the reservation protocol", func(t *testing.T) {
		writeSessionID := chatSessionTestID()
		writeSession := chatsessions.CoreChatSession{
			ID:          writeSessionID,
			UserID:      owner.userID,
			WorkspaceID: owner.workspaceID,
			Title:       "Migrated approval contract",
		}
		requestedPart := map[string]any{
			"approval":   map[string]any{"id": "approval-call-legacy"},
			"input":      map[string]any{"teamId": "team-1", "title": "Legacy pending"},
			"state":      "approval-requested",
			"toolCallId": "call-legacy",
			"type":       "tool-createStory",
		}
		respondedPart := map[string]any{
			"approval": map[string]any{
				"approved": true,
				"id":       "approval-call-legacy",
			},
			"input":      map[string]any{"teamId": "team-1", "title": "Legacy pending"},
			"state":      "approval-responded",
			"toolCallId": "call-legacy",
			"type":       "tool-createStory",
		}
		userMessage := chatTestMessage("legacy-user-1", "user", map[string]any{
			"text": "Create it",
			"type": "text",
		})
		requestedMessages := []any{
			userMessage,
			chatTestMessage("legacy-assistant-1", "assistant", requestedPart),
		}
		_, err := repository.CreateSessionWithMessages(ctx, &writeSession, requestedMessages)
		require.NoError(t, err)

		respondedMessages := []any{
			userMessage,
			chatTestMessage("legacy-assistant-1", "assistant", respondedPart),
		}
		reservation, err := repository.BeginMessageWrite(ctx, chatsessions.BeginMessageWriteParams{
			Session:   writeSession,
			Messages:  respondedMessages,
			Operation: chatsessions.MessageWriteApproval,
		})
		require.NoError(t, err)
		require.EqualValues(t, 1, reservation.Generation)
		require.Equal(t, respondedMessages, mustGetChatMessages(t, repository, writeSessionID, owner))
	})

	t.Run("uncertain mutation quarantines its fingerprint without blocking unrelated turns", func(t *testing.T) {
		writeSessionID := chatSessionTestID()
		writeSession := chatsessions.CoreChatSession{
			ID:          writeSessionID,
			UserID:      owner.userID,
			WorkspaceID: owner.workspaceID,
			Title:       "Uncertain mutation contract",
		}
		userMessage := chatTestMessage("uncertain-user-1", "user", map[string]any{
			"text": "Create the launch story",
			"type": "text",
		})
		reservation, err := repository.BeginMessageWrite(ctx, chatsessions.BeginMessageWriteParams{
			Session:   writeSession,
			Messages:  []any{userMessage},
			Operation: chatsessions.MessageWriteAppend,
		})
		require.NoError(t, err)
		assistantMessage := chatTestMessage("uncertain-assistant-1", "assistant", map[string]any{
			"text": "I could not confirm whether that change finished.",
			"type": "text",
		})
		currentMessages := []any{userMessage, assistantMessage}
		finalized, err := repository.FinalizeMessageWrite(ctx, chatsessions.FinalizeMessageWriteParams{
			SessionID:   writeSessionID,
			UserID:      owner.userID,
			WorkspaceID: owner.workspaceID,
			Messages:    currentMessages,
			Generation:  reservation.Generation,
			Token:       reservation.Token,
		})
		require.NoError(t, err)
		require.True(t, finalized.Applied)

		fingerprint := strings.Repeat("d", 64)
		approvalParams := mutationApprovalTestParams(writeSessionID, owner, "call-uncertain", fingerprint)
		claim, err := repository.ClaimMutationApproval(ctx, approvalParams)
		require.NoError(t, err)
		require.NotNil(t, claim.LeaseToken)
		approvalParams.LeaseToken = *claim.LeaseToken
		_, err = repository.StartMutationApproval(ctx, approvalParams)
		require.NoError(t, err)
		failed, err := repository.FailMutationApproval(ctx, approvalParams, chatsessions.MutationApprovalFailureCompletionUncertain)
		require.NoError(t, err)
		require.Equal(t, chatsessions.MutationApprovalExecutionFailed, failed.State)

		replayParams := approvalParams
		replayParams.ToolCallID = "call-uncertain-retry"
		replayParams.LeaseToken = uuid.Nil
		replay, err := repository.ClaimMutationApproval(ctx, replayParams)
		require.NoError(t, err)
		require.Equal(t, chatsessions.MutationApprovalExecutionFailed, replay.State)

		secondSessionID := chatSessionTestID()
		_, err = repository.CreateSessionWithMessages(ctx, &chatsessions.CoreChatSession{
			ID:          secondSessionID,
			UserID:      owner.userID,
			WorkspaceID: owner.workspaceID,
			Title:       "Cross-chat quarantine",
		}, []any{chatTestMessage("cross-chat-user", "user", map[string]any{
			"text": "Retry the same change in another chat",
			"type": "text",
		})})
		require.NoError(t, err)
		crossChatReplay := mutationApprovalTestParams(
			secondSessionID,
			owner,
			"call-uncertain-other-chat",
			fingerprint,
		)
		crossChatResult, err := repository.ClaimMutationApproval(ctx, crossChatReplay)
		require.NoError(t, err)
		require.Equal(t, chatsessions.MutationApprovalExecutionFailed, crossChatResult.State)

		unrelatedUserMessage := chatTestMessage("uncertain-user-2", "user", map[string]any{
			"text": "Show me the current workspace status instead.",
			"type": "text",
		})
		currentMessages = append(currentMessages, unrelatedUserMessage)
		unrelatedReservation, err := repository.BeginMessageWrite(ctx, chatsessions.BeginMessageWriteParams{
			Session:   writeSession,
			Messages:  currentMessages,
			Operation: chatsessions.MessageWriteAppend,
		})
		require.NoError(t, err)
		require.Greater(t, unrelatedReservation.Generation, reservation.Generation)
	})

	t.Run("expired unstarted fingerprint keeps a tombstone when a newer chat proceeds", func(t *testing.T) {
		firstSessionID := chatSessionTestID()
		secondSessionID := chatSessionTestID()
		for _, chatID := range []string{firstSessionID, secondSessionID} {
			_, err := repository.CreateSessionWithMessages(ctx, &chatsessions.CoreChatSession{
				ID:          chatID,
				UserID:      owner.userID,
				WorkspaceID: owner.workspaceID,
				Title:       "Ready lease transfer",
			}, []any{chatTestMessage("user-"+chatID, "user", map[string]any{
				"text": "Prepare the same safe change",
				"type": "text",
			})})
			require.NoError(t, err)
		}

		fingerprint := strings.Repeat("e", 64)
		firstParams := mutationApprovalTestParams(
			firstSessionID,
			owner,
			"call-ready-first-chat",
			fingerprint,
		)
		firstClaim, err := repository.ClaimMutationApproval(ctx, firstParams)
		require.NoError(t, err)
		require.NotNil(t, firstClaim.LeaseToken)
		require.NoError(t, expireMutationApprovalLease(ctx, db, firstParams))

		secondParams := mutationApprovalTestParams(
			secondSessionID,
			owner,
			"call-ready-second-chat",
			fingerprint,
		)
		secondClaim, err := repository.ClaimMutationApproval(ctx, secondParams)
		require.NoError(t, err)
		require.Equal(t, chatsessions.MutationApprovalExecutionClaimed, secondClaim.State)
		require.NotNil(t, secondClaim.LeaseToken)
		require.NotEqual(t, *firstClaim.LeaseToken, *secondClaim.LeaseToken)

		secondParams.LeaseToken = *secondClaim.LeaseToken
		started, err := repository.StartMutationApproval(ctx, secondParams)
		require.NoError(t, err)
		require.Equal(t, chatsessions.MutationApprovalExecutionStarted, started.State)
		completed, err := repository.CompleteMutationApproval(
			ctx,
			secondParams,
			json.RawMessage(`{"success":true,"storyId":"replacement-story"}`),
		)
		require.NoError(t, err)
		require.Equal(t, chatsessions.MutationApprovalExecutionCompleted, completed.State)

		staleReplay, err := repository.ClaimMutationApproval(ctx, firstParams)
		require.NoError(t, err)
		require.Equal(t, chatsessions.MutationApprovalExecutionCompleted, staleReplay.State)
		expectedExpiredOutput, err := json.Marshal(map[string]any{
			"error":   expiredApprovalOutputMessage,
			"success": false,
		})
		require.NoError(t, err)
		require.JSONEq(t, string(expectedExpiredOutput), string(staleReplay.Output))

		var rowCount int
		err = db.QueryRowx(`
			SELECT COUNT(*)
			FROM chat_mutation_approval_executions
			WHERE user_id = $1 AND workspace_id = $2 AND fingerprint = $3
		`, owner.userID, owner.workspaceID, fingerprint).Scan(&rowCount)
		require.NoError(t, err)
		require.Equal(t, 2, rowCount)
	})

	t.Run("exact persisted safe approval gets one atomic retry", func(t *testing.T) {
		retrySessionID := chatSessionTestID()
		retrySession := chatsessions.CoreChatSession{
			ID:          retrySessionID,
			UserID:      owner.userID,
			WorkspaceID: owner.workspaceID,
			Title:       "Safe retry contract",
		}
		input := map[string]any{"teamId": "team-1", "title": "One"}
		userMessage := chatTestMessage("retry-user-1", "user", map[string]any{"text": "Create it", "type": "text"})
		requestedPart := map[string]any{
			"approval":   map[string]any{"id": "approval-retry-1"},
			"input":      input,
			"state":      "approval-requested",
			"toolCallId": "call-retry-1",
			"type":       "tool-createStory",
		}
		respondedPart := map[string]any{
			"approval":   map[string]any{"approved": true, "id": "approval-retry-1"},
			"input":      input,
			"state":      "approval-responded",
			"toolCallId": "call-retry-1",
			"type":       "tool-createStory",
		}
		requestedMessages := []any{userMessage, chatTestMessage("retry-assistant-1", "assistant", requestedPart)}
		respondedMessages := []any{userMessage, chatTestMessage("retry-assistant-1", "assistant", respondedPart)}
		_, err := repository.CreateSessionWithMessages(ctx, &retrySession, requestedMessages)
		require.NoError(t, err)

		firstReservation, err := repository.BeginMessageWrite(ctx, chatsessions.BeginMessageWriteParams{
			Session: retrySession, Messages: respondedMessages, Operation: chatsessions.MessageWriteApproval,
		})
		require.NoError(t, err)
		fingerprint := "82b130ea1cd11c40fb87781033df92b460b42613047ff15eead2a2342f30c07a"
		approvalParams := mutationApprovalTestParams(retrySessionID, owner, "call-retry-1", fingerprint)
		firstClaim, err := repository.ClaimMutationApproval(ctx, approvalParams)
		require.NoError(t, err)
		require.NotNil(t, firstClaim.LeaseToken)
		approvalParams.LeaseToken = *firstClaim.LeaseToken
		_, err = repository.StartMutationApproval(ctx, approvalParams)
		require.NoError(t, err)
		_, err = repository.FailMutationApproval(ctx, approvalParams, chatsessions.MutationApprovalFailureCompletionUncertain)
		require.NoError(t, err)

		uncertainPart := map[string]any{
			"approval": map[string]any{"approved": true, "id": "approval-retry-1"},
			"input":    input,
			"output": map[string]any{
				"error":   chatsessions.MutationApprovalUncertainOutputMessage,
				"success": false,
			},
			"state":      "output-available",
			"toolCallId": "call-retry-1",
			"type":       "tool-createStory",
		}
		uncertainMessages := []any{userMessage, chatTestMessage("retry-assistant-1", "assistant", uncertainPart)}
		finalized, err := repository.FinalizeMessageWrite(ctx, chatsessions.FinalizeMessageWriteParams{
			SessionID: retrySessionID, UserID: owner.userID, WorkspaceID: owner.workspaceID,
			Messages: uncertainMessages, Generation: firstReservation.Generation, Token: firstReservation.Token,
		})
		require.NoError(t, err)
		require.True(t, finalized.Applied)

		retryReservation, err := repository.BeginMessageWrite(ctx, chatsessions.BeginMessageWriteParams{
			Session: retrySession, Messages: respondedMessages, Operation: chatsessions.MessageWriteApproval,
		})
		require.NoError(t, err)
		assertMutationApprovalRetryReady(t, db, approvalParams)
		reopened := mustGetChatMessages(t, repository, retrySessionID, owner)
		reopenedPart := reopened[1].(map[string]any)["parts"].([]any)[0].(map[string]any)
		require.Equal(t, "approval-responded", reopenedPart["state"])
		require.NotContains(t, reopenedPart, "output")

		otherSessionID := chatSessionTestID()
		_, err = repository.CreateSessionWithMessages(ctx, &chatsessions.CoreChatSession{
			ID: otherSessionID, UserID: owner.userID, WorkspaceID: owner.workspaceID, Title: "Safe retry quarantine",
		}, []any{chatTestMessage("retry-other-user", "user", map[string]any{"text": "Try elsewhere", "type": "text"})})
		require.NoError(t, err)
		crossOrigin := mutationApprovalTestParams(otherSessionID, owner, "call-retry-other", fingerprint)
		crossOriginResult, err := repository.ClaimMutationApproval(ctx, crossOrigin)
		require.NoError(t, err)
		require.Equal(t, chatsessions.MutationApprovalExecutionFailed, crossOriginResult.State)
		require.Equal(t, mutationApprovalWrongOriginFailure, crossOriginResult.FailureCode)

		secondClaim, err := repository.ClaimMutationApproval(ctx, mutationApprovalTestParams(retrySessionID, owner, "call-retry-1", fingerprint))
		require.NoError(t, err)
		require.NotNil(t, secondClaim.LeaseToken)
		approvalParams.LeaseToken = *secondClaim.LeaseToken
		_, err = repository.StartMutationApproval(ctx, approvalParams)
		require.NoError(t, err)
		_, err = repository.FailMutationApproval(ctx, approvalParams, chatsessions.MutationApprovalFailureCompletionUncertain)
		require.NoError(t, err)
		secondFinalization, err := repository.FinalizeMessageWrite(ctx, chatsessions.FinalizeMessageWriteParams{
			SessionID: retrySessionID, UserID: owner.userID, WorkspaceID: owner.workspaceID,
			Messages: uncertainMessages, Generation: retryReservation.Generation, Token: retryReservation.Token,
		})
		require.NoError(t, err)
		require.True(t, secondFinalization.Applied)

		thirdReservation, err := repository.BeginMessageWrite(ctx, chatsessions.BeginMessageWriteParams{
			Session: retrySession, Messages: respondedMessages, Operation: chatsessions.MessageWriteApproval,
		})
		require.NoError(t, err)
		require.NotNil(t, thirdReservation.Messages, "the terminal uncertainty should be returned as a canonical repair")
		assertMutationApprovalFailedAfterRetry(t, db, approvalParams)
		thirdClaim, err := repository.ClaimMutationApproval(ctx, mutationApprovalTestParams(retrySessionID, owner, "call-retry-1", fingerprint))
		require.NoError(t, err)
		require.Equal(t, chatsessions.MutationApprovalExecutionFailed, thirdClaim.State)
	})

	t.Run("soft deleted session cannot start a prepared safe retry", func(t *testing.T) {
		deletedSessionID := chatSessionTestID()
		_, err := repository.CreateSessionWithMessages(ctx, &chatsessions.CoreChatSession{
			ID: deletedSessionID, UserID: owner.userID, WorkspaceID: owner.workspaceID, Title: "Deleted retry",
		}, []any{chatTestMessage("deleted-retry-user", "user", map[string]any{"text": "Retry", "type": "text"})})
		require.NoError(t, err)
		leaseToken := uuid.New()
		fingerprint := strings.Repeat("f", 64)
		_, err = db.ExecContext(ctx, `
			INSERT INTO chat_mutation_approval_executions (
				session_id, user_id, workspace_id, tool_call_id, fingerprint,
				status, lease_token, lease_expires_at, failed_at, failure_code,
				attempt_count, reconciliation_count, last_reconciliation_resolution,
				last_reconciliation_evidence, last_reconciled_at
			) VALUES (
				$1, $2, $3, $4, $5, 'retry_ready', $6,
				CURRENT_TIMESTAMP + INTERVAL '30 seconds', CURRENT_TIMESTAMP,
				'completion_persistence_uncertain', 2, 1, 'safe_retry_prepared',
				CAST('{"kind":"server_verified_idempotent_retry","toolName":"createStory"}' AS jsonb),
				CURRENT_TIMESTAMP
			)
		`, deletedSessionID, owner.userID, owner.workspaceID, "call-deleted-retry", fingerprint, leaseToken)
		require.NoError(t, err)
		require.NoError(t, repository.DeleteSession(ctx, deletedSessionID, owner.userID, owner.workspaceID))
		params := mutationApprovalTestParams(deletedSessionID, owner, "call-deleted-retry", fingerprint)
		params.LeaseToken = leaseToken
		_, err = repository.StartMutationApproval(ctx, params)
		require.ErrorIs(t, err, chatsessions.ErrNotFound)
	})

	t.Run("execution states, lease recovery, and reconciliation are durable", func(t *testing.T) {
		completedFingerprint := strings.Repeat("a", 64)
		completedParams := mutationApprovalTestParams(sessionID, owner, "call-completed", completedFingerprint)
		claimed, err := repository.ClaimMutationApproval(ctx, completedParams)
		require.NoError(t, err)
		require.Equal(t, chatsessions.MutationApprovalExecutionClaimed, claimed.State)
		require.NotNil(t, claimed.LeaseToken)
		completedParams.LeaseToken = *claimed.LeaseToken

		pending, err := repository.ClaimMutationApproval(ctx, completedParams)
		require.NoError(t, err)
		require.Equal(t, chatsessions.MutationApprovalExecutionReady, pending.State)
		require.NotNil(t, pending.LeaseExpiresAt)

		wrongLease := completedParams
		wrongLease.LeaseToken = uuid.New()
		_, err = repository.StartMutationApproval(ctx, wrongLease)
		require.ErrorIs(t, err, chatsessions.ErrMutationApprovalLease)

		started, err := repository.StartMutationApproval(ctx, completedParams)
		require.NoError(t, err)
		require.Equal(t, chatsessions.MutationApprovalExecutionStarted, started.State)
		executing, err := repository.ClaimMutationApproval(ctx, completedParams)
		require.NoError(t, err)
		require.Equal(t, chatsessions.MutationApprovalExecutionExecuting, executing.State)

		completionOutput := json.RawMessage(`{"success":true,"storyId":"story-1"}`)
		completed, err := repository.CompleteMutationApproval(ctx, completedParams, completionOutput)
		require.NoError(t, err)
		require.Equal(t, chatsessions.MutationApprovalExecutionCompleted, completed.State)
		require.JSONEq(t, string(completionOutput), string(completed.Output))
		replayed, err := repository.ClaimMutationApproval(ctx, completedParams)
		require.NoError(t, err)
		require.Equal(t, chatsessions.MutationApprovalExecutionCompleted, replayed.State)
		require.JSONEq(t, string(completionOutput), string(replayed.Output))

		expiredFingerprint := strings.Repeat("b", 64)
		expiredParams := mutationApprovalTestParams(sessionID, owner, "call-expired", expiredFingerprint)
		firstLease, err := repository.ClaimMutationApproval(ctx, expiredParams)
		require.NoError(t, err)
		require.NotNil(t, firstLease.LeaseToken)
		require.NoError(t, expireMutationApprovalLease(ctx, db, expiredParams))
		reclaimed, err := repository.ClaimMutationApproval(ctx, expiredParams)
		require.NoError(t, err)
		require.Equal(t, chatsessions.MutationApprovalExecutionClaimed, reclaimed.State)
		require.NotNil(t, reclaimed.LeaseToken)
		require.NotEqual(t, *firstLease.LeaseToken, *reclaimed.LeaseToken)

		expiredParams.LeaseToken = *reclaimed.LeaseToken
		_, err = repository.StartMutationApproval(ctx, expiredParams)
		require.NoError(t, err)
		require.NoError(t, expireMutationApprovalLease(ctx, db, expiredParams))
		uncertain, err := repository.ClaimMutationApproval(ctx, expiredParams)
		require.NoError(t, err)
		require.Equal(t, chatsessions.MutationApprovalExecutionFailed, uncertain.State)
		require.Equal(t, mutationApprovalLeaseExpiredFailure, uncertain.FailureCode)

		verifiedOutput := json.RawMessage(`{"success":true,"storyId":"story-reconciled"}`)
		reconciled, err := repository.ReconcileMutationApproval(ctx, expiredParams, chatsessions.MutationApprovalReconciliation{
			Resolution: chatsessions.MutationApprovalReconciliationVerifiedCompleted,
			Evidence: chatsessions.MutationApprovalReconciliationEvidence{
				Kind:      chatsessions.MutationApprovalReconciliationEvidenceIdempotencyLookup,
				Reference: "story-create:call-expired",
				Summary:   "The idempotency receipt proves that the story was created.",
			},
			Output: verifiedOutput,
		})
		require.NoError(t, err)
		require.Equal(t, chatsessions.MutationApprovalExecutionCompleted, reconciled.State)
		require.JSONEq(t, string(verifiedOutput), string(reconciled.Output))
		assertMutationApprovalReconciliationAudit(t, db, expiredParams, "verified_completed")

		notAppliedFingerprint := strings.Repeat("c", 64)
		notAppliedParams := mutationApprovalTestParams(sessionID, owner, "call-not-applied", notAppliedFingerprint)
		notAppliedLease, err := repository.ClaimMutationApproval(ctx, notAppliedParams)
		require.NoError(t, err)
		require.NotNil(t, notAppliedLease.LeaseToken)
		notAppliedParams.LeaseToken = *notAppliedLease.LeaseToken
		_, err = repository.StartMutationApproval(ctx, notAppliedParams)
		require.NoError(t, err)
		_, err = repository.FailMutationApproval(ctx, notAppliedParams, chatsessions.MutationApprovalFailureCompletionUncertain)
		require.NoError(t, err)

		reset, err := repository.ReconcileMutationApproval(ctx, notAppliedParams, chatsessions.MutationApprovalReconciliation{
			Resolution: chatsessions.MutationApprovalReconciliationVerifiedNotApplied,
			Evidence: chatsessions.MutationApprovalReconciliationEvidence{
				Kind:      chatsessions.MutationApprovalReconciliationEvidenceWorkspaceProbe,
				Reference: "story:absent:call-not-applied",
				Summary:   "The owner-scoped workspace probe proves that the story is absent.",
			},
		})
		require.NoError(t, err)
		require.Equal(t, chatsessions.MutationApprovalExecutionReady, reset.State)
		assertMutationApprovalReset(t, db, notAppliedParams)

		freshClaim, err := repository.ClaimMutationApproval(ctx, notAppliedParams)
		require.NoError(t, err)
		require.Equal(t, chatsessions.MutationApprovalExecutionClaimed, freshClaim.State)
		require.NotNil(t, freshClaim.LeaseToken)
		require.NotEqual(t, notAppliedParams.LeaseToken, *freshClaim.LeaseToken)

		for _, foreignScope := range []chatSessionTestActor{
			{userID: other.userID, workspaceID: owner.workspaceID},
			{userID: owner.userID, workspaceID: other.workspaceID},
		} {
			foreignParams := completedParams
			foreignParams.UserID = foreignScope.userID
			foreignParams.WorkspaceID = foreignScope.workspaceID
			_, err = repository.ClaimMutationApproval(ctx, foreignParams)
			require.ErrorIs(t, err, chatsessions.ErrNotFound)
		}
	})
}

func chatTestMessage(id, role string, parts ...any) map[string]any {
	return map[string]any{
		"id":    id,
		"parts": parts,
		"role":  role,
	}
}

type chatSessionTestActor struct {
	userID      uuid.UUID
	workspaceID uuid.UUID
}

func seedChatSessionActor(t *testing.T, db *sqlx.DB) chatSessionTestActor {
	t.Helper()
	ctx := context.Background()
	userID := uuid.New()
	workspaceID := uuid.New()
	suffix := strings.ReplaceAll(uuid.NewString(), "-", "")
	_, err := db.ExecContext(ctx, `
		INSERT INTO users (user_id, username, email)
		VALUES ($1, $2, $3)
	`, userID, "chat-test-"+suffix, "chat-test-"+suffix+"@example.invalid")
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `
		INSERT INTO workspaces (workspace_id, name, slug, created_by)
		VALUES ($1, $2, $3, $4)
	`, workspaceID, "Chat session test", "chat-test-"+suffix, userID)
	require.NoError(t, err)
	return chatSessionTestActor{userID: userID, workspaceID: workspaceID}
}

func chatSessionTestID() string {
	return strings.ReplaceAll(uuid.NewString(), "-", "")[:16]
}

func mustGetChatMessages(t *testing.T, repository *repo, sessionID string, actor chatSessionTestActor) []any {
	t.Helper()
	messages, err := repository.GetMessages(context.Background(), sessionID, actor.userID, actor.workspaceID)
	require.NoError(t, err)
	return messages
}

func mutationApprovalTestParams(sessionID string, actor chatSessionTestActor, toolCallID, fingerprint string) chatsessions.MutationApprovalExecutionParams {
	return chatsessions.MutationApprovalExecutionParams{
		SessionID:   sessionID,
		UserID:      actor.userID,
		WorkspaceID: actor.workspaceID,
		ToolCallID:  toolCallID,
		Fingerprint: fingerprint,
	}
}

func expireMutationApprovalLease(ctx context.Context, db *sqlx.DB, params chatsessions.MutationApprovalExecutionParams) error {
	result, err := db.ExecContext(ctx, `
		UPDATE chat_mutation_approval_executions
		SET lease_expires_at = CURRENT_TIMESTAMP - INTERVAL '1 second'
		WHERE session_id = $1
			AND user_id = $2
			AND workspace_id = $3
			AND tool_call_id = $4
	`, params.SessionID, params.UserID, params.WorkspaceID, params.ToolCallID)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected != 1 {
		return errors.New("mutation approval lease was not found")
	}
	return nil
}

func assertMutationApprovalReconciliationAudit(t *testing.T, db *sqlx.DB, params chatsessions.MutationApprovalExecutionParams, expectedResolution string) {
	t.Helper()
	var resolution string
	var evidence json.RawMessage
	var count int
	err := db.QueryRowx(`
		SELECT last_reconciliation_resolution, last_reconciliation_evidence, reconciliation_count
		FROM chat_mutation_approval_executions
		WHERE session_id = $1
			AND user_id = $2
			AND workspace_id = $3
			AND tool_call_id = $4
	`, params.SessionID, params.UserID, params.WorkspaceID, params.ToolCallID).Scan(&resolution, &evidence, &count)
	require.NoError(t, err)
	require.Equal(t, expectedResolution, resolution)
	require.JSONEq(t, `{"kind":"idempotency_lookup","reference":"story-create:call-expired","summary":"The idempotency receipt proves that the story was created."}`, string(evidence))
	require.Equal(t, 1, count)
}

func assertMutationApprovalReset(t *testing.T, db *sqlx.DB, params chatsessions.MutationApprovalExecutionParams) {
	t.Helper()
	var status string
	var startedAt *time.Time
	var leaseExpiresAt time.Time
	var resolution string
	err := db.QueryRowx(`
		SELECT status, started_at, lease_expires_at, last_reconciliation_resolution
		FROM chat_mutation_approval_executions
		WHERE session_id = $1
			AND user_id = $2
			AND workspace_id = $3
			AND tool_call_id = $4
	`, params.SessionID, params.UserID, params.WorkspaceID, params.ToolCallID).Scan(
		&status,
		&startedAt,
		&leaseExpiresAt,
		&resolution,
	)
	require.NoError(t, err)
	require.Equal(t, "ready", status)
	require.Nil(t, startedAt)
	require.LessOrEqual(t, leaseExpiresAt, time.Now())
	require.Equal(t, "verified_not_applied", resolution)
}

func assertMutationApprovalRetryReady(t *testing.T, db *sqlx.DB, params chatsessions.MutationApprovalExecutionParams) {
	t.Helper()
	var status string
	var resolution string
	var count int
	var leaseToken *uuid.UUID
	err := db.QueryRowx(`
		SELECT status, last_reconciliation_resolution, reconciliation_count, lease_token
		FROM chat_mutation_approval_executions
		WHERE session_id = $1 AND user_id = $2 AND workspace_id = $3 AND tool_call_id = $4
	`, params.SessionID, params.UserID, params.WorkspaceID, params.ToolCallID).Scan(&status, &resolution, &count, &leaseToken)
	require.NoError(t, err)
	require.Equal(t, "retry_ready", status)
	require.Equal(t, mutationApprovalSafeRetryResolution, resolution)
	require.Equal(t, 1, count)
	require.Nil(t, leaseToken)
}

func assertMutationApprovalFailedAfterRetry(t *testing.T, db *sqlx.DB, params chatsessions.MutationApprovalExecutionParams) {
	t.Helper()
	var status string
	var resolution string
	var count int
	err := db.QueryRowx(`
		SELECT status, last_reconciliation_resolution, reconciliation_count
		FROM chat_mutation_approval_executions
		WHERE session_id = $1 AND user_id = $2 AND workspace_id = $3 AND tool_call_id = $4
	`, params.SessionID, params.UserID, params.WorkspaceID, params.ToolCallID).Scan(&status, &resolution, &count)
	require.NoError(t, err)
	require.Equal(t, "failed_uncertain", status)
	require.Equal(t, mutationApprovalSafeRetryResolution, resolution)
	require.Equal(t, 1, count)
}
