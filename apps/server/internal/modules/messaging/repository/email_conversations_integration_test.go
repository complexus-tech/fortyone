package messagingrepository

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	messaging "github.com/complexus-tech/projects-api/internal/modules/messaging/service"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

// TestEmailConversationPostgresContract exercises idempotent history, token
// rotation, summary CAS, consent CAS, and apply attempt fencing against an
// isolated migrated PostgreSQL database when one is supplied.
func TestEmailConversationPostgresContract(t *testing.T) {
	databaseURL := os.Getenv("MESSAGING_EMAIL_TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("MESSAGING_EMAIL_TEST_DATABASE_URL is not configured")
	}
	db, err := sqlx.Connect("postgres", databaseURL)
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	ctx := context.Background()
	repo := New(db)
	workspaceID, userID := seedEmailConversationActor(t, db)
	now := time.Now().UTC().Truncate(time.Microsecond)
	tokenHash := bytes.Repeat([]byte{1}, sha256DigestSize)
	thread, created, err := repo.CreateEmailThread(ctx, messaging.EmailThreadInput{
		Provider:              "brevo_email",
		WorkspaceID:           workspaceID,
		UserID:                userID,
		RecipientEmail:        "email-conversation@example.test",
		ExternalThreadID:      uuid.NewString(),
		RootInternetMessageID: "<maya-root@fortyone.app>",
		Context:               json.RawMessage(`{"objective":{"name":"Improve onboarding"}}`),
		ReplyTokenHash:        tokenHash,
		ReplyTokenExpiresAt:   now.Add(24 * time.Hour),
	})
	require.NoError(t, err)
	require.True(t, created)

	lookup, err := repo.FindEmailThreadByReplyToken(ctx, messaging.EmailReplyTokenLookup{
		Provider: "brevo_email", TokenHash: tokenHash, Now: now,
	})
	require.NoError(t, err)
	require.Equal(t, thread.ID, lookup.Thread.ID)
	_, err = db.ExecContext(ctx, `
		UPDATE workspace_members
		SET role = 'guest'
		WHERE workspace_id = $1 AND user_id = $2
	`, workspaceID, userID)
	require.NoError(t, err)
	_, err = repo.FindEmailThreadByReplyToken(ctx, messaging.EmailReplyTokenLookup{
		Provider: "brevo_email", TokenHash: tokenHash, Now: now,
	})
	require.NoError(t, err, "a current guest recipient keeps access to their email thread")

	rotatedHash := bytes.Repeat([]byte{2}, sha256DigestSize)
	_, aliasCreated, err := repo.CreateEmailReplyTokenAlias(ctx, messaging.EmailReplyTokenInput{
		ThreadID: thread.ID, WorkspaceID: workspaceID, UserID: userID,
		TokenHash: rotatedHash, ExpiresAt: now.Add(48 * time.Hour),
	})
	require.NoError(t, err)
	require.True(t, aliasCreated)
	_, err = repo.FindEmailThreadByReplyToken(ctx, messaging.EmailReplyTokenLookup{
		Provider: "brevo_email", TokenHash: tokenHash, Now: now,
	})
	require.NoError(t, err, "rotating an alias must not invalidate an older active address")
	_, err = db.ExecContext(ctx, `UPDATE users SET email = 'changed@example.test' WHERE user_id = $1`, userID)
	require.NoError(t, err)
	_, err = repo.FindEmailThreadByReplyToken(ctx, messaging.EmailReplyTokenLookup{
		Provider: "brevo_email", TokenHash: tokenHash, Now: now,
	})
	require.ErrorIs(t, err, messaging.ErrInvalidEmailReplyToken, "an old mailbox must lose authority after the account email changes")
	_, err = db.ExecContext(ctx, `UPDATE users SET email = 'email-conversation@example.test' WHERE user_id = $1`, userID)
	require.NoError(t, err)

	messageInput := messaging.EmailMessageInput{
		ThreadID: thread.ID, WorkspaceID: workspaceID, UserID: userID,
		IdempotencyKey: "brevo:event-1", Direction: messaging.EmailMessageDirectionInbound,
		Role: messaging.EmailMessageRoleUser, Kind: messaging.EmailMessageKindReply,
		ProviderMessageID: "event-1", InternetMessageID: "<reply-1@example.test>",
		InReplyToMessageID: "<maya-root@fortyone.app>", Subject: "Re: Objective update",
		Content: "Set the key result to 38%.", Context: json.RawMessage(`{}`),
		ProviderMetadata: json.RawMessage(`{"provider":"brevo"}`),
	}
	message, appended, err := repo.AppendEmailMessage(ctx, messageInput)
	require.NoError(t, err)
	require.True(t, appended)
	replayed, appended, err := repo.AppendEmailMessage(ctx, messageInput)
	require.NoError(t, err)
	require.False(t, appended)
	require.Equal(t, message.ID, replayed.ID)

	thread, err = repo.UpdateEmailThreadSummary(ctx, messaging.EmailThreadSummaryUpdate{
		ThreadID: thread.ID, WorkspaceID: workspaceID, UserID: userID,
		ExpectedSummaryThroughSequence: 0, Summary: "The user wants to update one key result.",
		ThroughSequence: message.Sequence,
	})
	require.NoError(t, err)
	_, err = repo.UpdateEmailThreadSummary(ctx, messaging.EmailThreadSummaryUpdate{
		ThreadID: thread.ID, WorkspaceID: workspaceID, UserID: userID,
		ExpectedSummaryThroughSequence: 0, Summary: "Stale summary.", ThroughSequence: message.Sequence,
	})
	require.ErrorIs(t, err, messaging.ErrEmailSummaryConflict)

	proposal, created, err := repo.RegisterEmailActionProposal(ctx, messaging.EmailActionProposalInput{
		ThreadID: thread.ID, WorkspaceID: workspaceID, UserID: userID,
		SourceMessageID: message.ID, IdempotencyKey: "proposal:event-1",
		ActionKind: "update_key_result", EntityType: "key_result", EntityID: uuid.New(),
		ExpectedEntityVersion: "2026-08-12T08:00:00Z",
		ProposedDiff:          json.RawMessage(`{"currentValue":{"from":32,"to":38}}`),
		ExpiresAt:             now.Add(time.Hour), Now: now,
	})
	require.NoError(t, err)
	require.True(t, created)
	pending, err := repo.ListPendingEmailActionProposals(ctx, messaging.EmailActionProposalListInput{
		ThreadID: thread.ID, WorkspaceID: workspaceID, UserID: userID, Now: now,
	})
	require.NoError(t, err)
	require.Len(t, pending, 1)
	require.Equal(t, proposal.ID, pending[0].ID)
	loadedProposal, err := repo.GetEmailActionProposal(ctx, messaging.EmailActionProposalKey{
		ProposalID: proposal.ID, ThreadID: thread.ID, WorkspaceID: workspaceID, UserID: userID,
	})
	require.NoError(t, err)
	require.JSONEq(t, string(proposal.ProposedDiff), string(loadedProposal.ProposedDiff))

	confirmed, duplicate, err := repo.DecideEmailActionProposal(ctx, messaging.EmailActionProposalDecision{
		ProposalID: proposal.ID, ThreadID: thread.ID, WorkspaceID: workspaceID, UserID: userID,
		ReplyTokenHash: rotatedHash, Decision: messaging.EmailActionProposalConfirmed, Now: now,
	})
	require.NoError(t, err)
	require.False(t, duplicate)
	require.Equal(t, messaging.EmailActionProposalConfirmed, confirmed.Status)
	_, _, err = repo.DecideEmailActionProposal(ctx, messaging.EmailActionProposalDecision{
		ProposalID: proposal.ID, ThreadID: thread.ID, WorkspaceID: workspaceID, UserID: userID,
		ReplyTokenHash: rotatedHash, Decision: messaging.EmailActionProposalCancelled, Now: now,
	})
	require.ErrorIs(t, err, messaging.ErrEmailProposalConflict)

	claimed, claimedNow, err := repo.ClaimEmailActionProposalApply(ctx, messaging.EmailActionProposalApplyClaim{
		ProposalID: proposal.ID, ThreadID: thread.ID, WorkspaceID: workspaceID, UserID: userID, Now: now,
	})
	require.NoError(t, err)
	require.True(t, claimedNow)
	require.Equal(t, 1, claimed.ApplyAttempt)
	_, _, err = repo.ClaimEmailActionProposalApply(ctx, messaging.EmailActionProposalApplyClaim{
		ProposalID: proposal.ID, ThreadID: thread.ID, WorkspaceID: workspaceID, UserID: userID,
		Now: now.Add(time.Second), RetryAfter: time.Minute,
	})
	require.ErrorIs(t, err, messaging.ErrEmailProposalApplyBusy)

	applied, duplicate, err := repo.CompleteEmailActionProposalApply(ctx, messaging.EmailActionProposalApplyCompletion{
		ProposalID: proposal.ID, ThreadID: thread.ID, WorkspaceID: workspaceID, UserID: userID,
		ApplyAttempt: claimed.ApplyAttempt, Status: messaging.EmailActionProposalApplied,
		Result: json.RawMessage(`{"currentValue":38}`), Now: now.Add(2 * time.Second),
	})
	require.NoError(t, err)
	require.False(t, duplicate)
	require.Equal(t, messaging.EmailActionProposalApplied, applied.Status)
	confirmedAgain, duplicate, err := repo.DecideEmailActionProposal(ctx, messaging.EmailActionProposalDecision{
		ProposalID: proposal.ID, ThreadID: thread.ID, WorkspaceID: workspaceID, UserID: userID,
		ReplyTokenHash: rotatedHash, Decision: messaging.EmailActionProposalConfirmed, Now: now.Add(3 * time.Second),
	})
	require.NoError(t, err)
	require.True(t, duplicate)
	require.Equal(t, messaging.EmailActionProposalApplied, confirmedAgain.Status)
	_, duplicate, err = repo.CompleteEmailActionProposalApply(ctx, messaging.EmailActionProposalApplyCompletion{
		ProposalID: proposal.ID, ThreadID: thread.ID, WorkspaceID: workspaceID, UserID: userID,
		ApplyAttempt: claimed.ApplyAttempt, Status: messaging.EmailActionProposalApplied,
		Result: json.RawMessage(`{"currentValue":38}`), Now: now.Add(4 * time.Second),
	})
	require.NoError(t, err)
	require.True(t, duplicate)

	expiringProposal, created, err := repo.RegisterEmailActionProposal(ctx, messaging.EmailActionProposalInput{
		ThreadID: thread.ID, WorkspaceID: workspaceID, UserID: userID,
		SourceMessageID: message.ID, IdempotencyKey: "proposal:event-2",
		ActionKind: "update_key_result", EntityType: "key_result", EntityID: uuid.New(),
		ExpectedEntityVersion: "2026-08-12T09:00:00Z",
		ProposedDiff:          json.RawMessage(`{"currentValue":{"from":38,"to":42}}`),
		ExpiresAt:             now.Add(5 * time.Minute), Now: now,
	})
	require.NoError(t, err)
	require.True(t, created)
	pending, err = repo.ListPendingEmailActionProposals(ctx, messaging.EmailActionProposalListInput{
		ThreadID: thread.ID, WorkspaceID: workspaceID, UserID: userID, Now: now.Add(6 * time.Minute),
	})
	require.NoError(t, err)
	require.Empty(t, pending)
	expiringProposal, err = repo.GetEmailActionProposal(ctx, messaging.EmailActionProposalKey{
		ProposalID: expiringProposal.ID, ThreadID: thread.ID, WorkspaceID: workspaceID, UserID: userID,
	})
	require.NoError(t, err)
	require.Equal(t, messaging.EmailActionProposalExpired, expiringProposal.Status)

	_, err = db.ExecContext(ctx, `
		DELETE FROM workspace_members
		WHERE workspace_id = $1 AND user_id = $2
	`, workspaceID, userID)
	require.NoError(t, err)
	_, err = repo.FindEmailThreadByReplyToken(ctx, messaging.EmailReplyTokenLookup{
		Provider: "brevo_email", TokenHash: rotatedHash, Now: now,
	})
	require.ErrorIs(t, err, messaging.ErrInvalidEmailReplyToken)
	_, _, err = repo.DecideEmailActionProposal(ctx, messaging.EmailActionProposalDecision{
		ProposalID: proposal.ID, ThreadID: thread.ID, WorkspaceID: workspaceID, UserID: userID,
		ReplyTokenHash: rotatedHash, Decision: messaging.EmailActionProposalConfirmed, Now: now,
	})
	require.ErrorIs(t, err, messaging.ErrInvalidEmailReplyToken)
}

func seedEmailConversationActor(t *testing.T, db *sqlx.DB) (uuid.UUID, uuid.UUID) {
	t.Helper()
	ctx := context.Background()
	userID := uuid.New()
	workspaceID := uuid.New()
	suffix := uuid.NewString()
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), "DELETE FROM workspaces WHERE workspace_id = $1", workspaceID)
		_, _ = db.ExecContext(context.Background(), "DELETE FROM users WHERE user_id = $1", userID)
	})
	_, err := db.ExecContext(ctx, `
		INSERT INTO users (user_id, username, email)
		VALUES ($1, $2, $3)
	`, userID, "email-conversation-"+suffix, "email-conversation-"+suffix+"@example.test")
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `
		INSERT INTO workspaces (workspace_id, name, slug, created_by)
		VALUES ($1, 'Email Conversation Test', $2, $3)
	`, workspaceID, "email-conversation-"+suffix, userID)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `
		INSERT INTO workspace_members (workspace_id, user_id, role)
		VALUES ($1, $2, 'member')
	`, workspaceID, userID)
	if err != nil && !strings.Contains(err.Error(), "duplicate") {
		require.NoError(t, err)
	}
	return workspaceID, userID
}
