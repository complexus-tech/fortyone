//go:build integration

package agentadapter

import (
	"encoding/json"
	"errors"
	"testing"

	agentrepository "github.com/complexus-tech/projects-api/internal/modules/agentinvocations/repository"
	agentruns "github.com/complexus-tech/projects-api/internal/modules/agentinvocations/service"
	chatdomain "github.com/complexus-tech/projects-api/internal/modules/chat/domain"
	messagingrepository "github.com/complexus-tech/projects-api/internal/modules/messaging/repository"
	messaging "github.com/complexus-tech/projects-api/internal/modules/messaging/service"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestChatMayaRespondsToThreadFollowupWithoutAnotherMention(t *testing.T) {
	f := newChatAgentFixture(t)
	worker := agentruns.New(agentrepository.New(f.pool), &RuntimeRouter{Chat: f.runtime})
	root, err := f.chats.CreateMessage(f.ctx, f.command(t))
	require.NoError(t, err)
	runs, err := f.runs.List(f.ctx, f.workspace, f.conversation)
	require.NoError(t, err)
	require.Len(t, runs, 1)
	require.NoError(t, worker.Process(f.ctx, runs[0].ID))
	require.Zero(t, f.assistant.classificationCalls, "explicit mentions bypass addressing classification")

	f.assistant.followupDecision = true
	command := chatdomain.CreateMessageCommand{Scope: f.scope, ConversationID: f.conversation, CreationID: uuid.New(), ThreadRootID: &root.ID, Text: "Could you break that down into the next steps?"}
	followup, err := f.chats.CreateMessage(f.ctx, command)
	require.NoError(t, err)
	replayed, err := f.chats.CreateMessage(f.ctx, command)
	require.NoError(t, err)
	require.Equal(t, followup.ID, replayed.ID)
	runs, err = f.runs.List(f.ctx, f.workspace, f.conversation)
	require.NoError(t, err)
	require.Len(t, runs, 2)
	var accepted agentruns.Run
	for _, run := range runs {
		if run.SourceID == followup.ID {
			accepted = run
		}
	}
	require.NotEqual(t, uuid.Nil, accepted.ID)
	require.NoError(t, worker.Process(f.ctx, accepted.ID))
	finished, err := f.runs.Get(f.ctx, f.workspace, accepted.ID)
	require.NoError(t, err)
	require.Equal(t, agentruns.StatusSucceeded, finished.Status)
	require.NotNil(t, finished.ReplyCommentID)
	reply, err := f.chats.GetMessage(f.ctx, f.scope, *finished.ReplyCommentID)
	require.NoError(t, err)
	require.Equal(t, &root.ID, reply.ThreadRootID)
	require.Equal(t, f.maya, reply.Author.ID)
	require.Equal(t, 2, f.assistant.calls)
	require.Equal(t, 1, f.assistant.classificationCalls)
	require.Equal(t, followup.Text, f.assistant.classification.Prompt)
	require.Len(t, f.assistant.classification.Conversation, 2, "classifier sees the question and Maya's answer")
	require.NoError(t, worker.Process(f.ctx, accepted.ID))
	require.Equal(t, 2, f.assistant.calls, "delivery recovery must not answer the same follow-up twice")
}

func TestChatMayaIgnoresOtherPeopleAndUnrelatedThreadMessages(t *testing.T) {
	f := newChatAgentFixture(t)
	worker := agentruns.New(agentrepository.New(f.pool), &RuntimeRouter{Chat: f.runtime})
	humanRoot, err := f.chats.CreateMessage(f.ctx, chatdomain.CreateMessageCommand{Scope: f.scope, ConversationID: f.conversation, CreationID: uuid.New(), Text: "Team, shall we review onboarding?"})
	require.NoError(t, err)
	_, err = f.chats.CreateMessage(f.ctx, chatdomain.CreateMessageCommand{Scope: f.scope, ConversationID: f.conversation, CreationID: uuid.New(), ThreadRootID: &humanRoot.ID, Text: "What should we do next?"})
	require.NoError(t, err)
	before, err := f.runs.List(f.ctx, f.workspace, f.conversation)
	require.NoError(t, err)
	require.Empty(t, before, "a human-only thread never queues Maya or calls the classifier")
	root, err := f.chats.CreateMessage(f.ctx, f.command(t))
	require.NoError(t, err)
	runs, err := f.runs.List(f.ctx, f.workspace, f.conversation)
	require.NoError(t, err)
	require.NoError(t, worker.Process(f.ctx, runs[0].ID))

	rich, err := json.Marshal(chatContentNode{Type: "doc", Content: []chatContentNode{{Type: "paragraph", Content: []chatContentNode{{Type: "mention", Attrs: map[string]any{"id": f.other.String(), "label": "Teammate"}}, {Type: "text", Text: " can you review this?"}}}}})
	require.NoError(t, err)
	_, err = f.chats.CreateMessage(f.ctx, chatdomain.CreateMessageCommand{Scope: f.scope, ConversationID: f.conversation, CreationID: uuid.New(), ThreadRootID: &root.ID, Text: "@Teammate can you review this?", RichContent: rich, MentionedUserIDs: []uuid.UUID{f.other}})
	require.NoError(t, err)
	runs, err = f.runs.List(f.ctx, f.workspace, f.conversation)
	require.NoError(t, err)
	require.Len(t, runs, 1, "direct requests to another person do not queue Maya")

	for _, scenario := range []struct {
		text string
		err  error
	}{
		{"Lunch at noon?", nil},
		{"Thanks!", nil},
		{"Could someone review this?", errors.New("classification unavailable")},
	} {
		f.assistant.followupDecision = false
		f.assistant.followupError = scenario.err
		message, err := f.chats.CreateMessage(f.ctx, chatdomain.CreateMessageCommand{Scope: f.scope, ConversationID: f.conversation, CreationID: uuid.New(), ThreadRootID: &root.ID, Text: scenario.text})
		require.NoError(t, err)
		runs, err = f.runs.List(f.ctx, f.workspace, f.conversation)
		require.NoError(t, err)
		for _, run := range runs {
			if run.SourceID != message.ID {
				continue
			}
			require.NoError(t, worker.Process(f.ctx, run.ID))
			finished, err := f.runs.Get(f.ctx, f.workspace, run.ID)
			require.NoError(t, err)
			require.Equal(t, agentruns.StatusSucceeded, finished.Status)
			require.True(t, finished.Proposal.Ignore)
			require.Nil(t, finished.Outcome)
			require.Nil(t, finished.ReplyCommentID)
		}
	}
	require.Equal(t, 1, f.assistant.calls, "only the original explicit request reaches the business assistant")
	var replies int
	require.NoError(t, f.pool.QueryRow(f.ctx, `SELECT count(*) FROM chat_conversation_messages WHERE conversation_id=$1 AND author_id=$2`, f.conversation, f.maya).Scan(&replies))
	require.Equal(t, 1, replies, "ignored messages must not publish replies or failure notices")
}

func TestChatMayaIgnoredDecisionSurvivesCompletionRecovery(t *testing.T) {
	f := newChatAgentFixture(t)
	repository := agentrepository.New(f.pool)
	worker := agentruns.New(repository, &RuntimeRouter{Chat: f.runtime})
	root, err := f.chats.CreateMessage(f.ctx, f.command(t))
	require.NoError(t, err)
	runs, err := f.runs.List(f.ctx, f.workspace, f.conversation)
	require.NoError(t, err)
	require.NoError(t, worker.Process(f.ctx, runs[0].ID))
	message, err := f.chats.CreateMessage(f.ctx, chatdomain.CreateMessageCommand{Scope: f.scope, ConversationID: f.conversation, CreationID: uuid.New(), ThreadRootID: &root.ID, Text: "Thanks!"})
	require.NoError(t, err)
	runs, err = f.runs.List(f.ctx, f.workspace, f.conversation)
	require.NoError(t, err)
	var accepted agentruns.Run
	for _, run := range runs {
		if run.SourceID == message.ID {
			accepted = run
		}
	}
	require.NotEqual(t, uuid.Nil, accepted.ID)
	claimed, err := repository.Claim(f.ctx, accepted.ID)
	require.NoError(t, err)
	claimed, err = repository.SaveProposal(f.ctx, claimed, agentruns.Proposal{Ignore: true})
	require.NoError(t, err)
	_, err = f.pool.Exec(f.ctx, `UPDATE agent_invocations SET retry_attempt_count=4 WHERE id=$1`, accepted.ID)
	require.NoError(t, err)
	claimed.RetryAttemptCount = 4
	require.True(t, claimed.HasDurableResponse())
	require.NoError(t, repository.Fail(f.ctx, claimed, "Temporary completion failure", true))
	pending, err := repository.Get(f.ctx, f.workspace, accepted.ID)
	require.NoError(t, err)
	require.Equal(t, agentruns.StatusQueued, pending.Status, "saved ignore decisions retain completion retries past preparation limit")
	_, err = f.pool.Exec(f.ctx, `UPDATE agent_invocations SET available_at=clock_timestamp() WHERE id=$1`, accepted.ID)
	require.NoError(t, err)
	require.NoError(t, worker.Process(f.ctx, accepted.ID))
	finished, err := repository.Get(f.ctx, f.workspace, accepted.ID)
	require.NoError(t, err)
	require.Equal(t, agentruns.StatusSucceeded, finished.Status)
	require.Nil(t, finished.ReplyCommentID)
	require.Nil(t, finished.FailureReplyID)
	require.Zero(t, f.assistant.classificationCalls, "recovery must not reclassify or answer an ignored message")
	require.Equal(t, 1, f.assistant.calls)
}

func TestChatMayaIgnoredAccountingRetriesSilentlyAndChargesOnce(t *testing.T) {
	f := newChatAgentFixture(t)
	repository := agentrepository.New(f.pool)
	worker := agentruns.New(repository, &RuntimeRouter{Chat: f.runtime})
	root, err := f.chats.CreateMessage(f.ctx, f.command(t))
	require.NoError(t, err)
	runs, err := f.runs.List(f.ctx, f.workspace, f.conversation)
	require.NoError(t, err)
	require.NoError(t, worker.Process(f.ctx, runs[0].ID))

	usage := messagingrepository.NewDailyUsageRepository(f.pool)
	f.runtime.Shared.Usage = usage
	f.assistant.followupUsage = messaging.Usage{InputTokens: 7, OutputTokens: 2, TotalTokens: 9}
	// Admission reads still work; only accounting writes fail in this isolated
	// database. Its transaction must roll back the idempotency receipt as well.
	_, err = f.pool.Exec(f.ctx, `ALTER TABLE messaging_assistant_daily_usage ADD CONSTRAINT test_usage_outage CHECK (total_tokens < 0)`)
	require.NoError(t, err)
	message, err := f.chats.CreateMessage(f.ctx, chatdomain.CreateMessageCommand{Scope: f.scope, ConversationID: f.conversation, CreationID: uuid.New(), ThreadRootID: &root.ID, Text: "Thanks!"})
	require.NoError(t, err)
	runs, err = f.runs.List(f.ctx, f.workspace, f.conversation)
	require.NoError(t, err)
	var accepted agentruns.Run
	for _, run := range runs {
		if run.SourceID == message.ID {
			accepted = run
		}
	}
	require.NotEqual(t, uuid.Nil, accepted.ID)
	for range 5 {
		require.Error(t, worker.Process(f.ctx, accepted.ID))
		pending, err := f.runs.Get(f.ctx, f.workspace, accepted.ID)
		require.NoError(t, err)
		require.Equal(t, agentruns.StatusQueued, pending.Status)
		require.True(t, pending.Proposal.Ignore)
		require.NotEmpty(t, pending.Proposal.IgnoreReceipt)
		require.Nil(t, pending.ReplyCommentID)
		require.Nil(t, pending.FailureReplyID)
		_, err = f.pool.Exec(f.ctx, `UPDATE agent_invocations SET available_at=clock_timestamp() WHERE id=$1`, accepted.ID)
		require.NoError(t, err)
	}
	runs, err = f.runs.List(f.ctx, f.workspace, f.conversation)
	require.NoError(t, err)
	require.Len(t, runs, 1, "silent bookkeeping must not appear as an active Maya response")
	require.Equal(t, 1, f.assistant.classificationCalls)
	require.Equal(t, 1, f.assistant.calls)
	pending, err := f.runs.Get(f.ctx, f.workspace, accepted.ID)
	require.NoError(t, err)
	_, err = f.runs.Cancel(f.ctx, agentruns.Control{WorkspaceID: f.workspace, RequesterID: f.user, RunID: pending.ID, ExpectedVersion: pending.Version})
	require.ErrorIs(t, err, agentruns.ErrConflict, "a settled no-response decision only has internal accounting left")
	_, err = f.pool.Exec(f.ctx, `DELETE FROM team_members WHERE team_id=$1 AND user_id=$2`, f.team, f.user)
	require.NoError(t, err)
	_, err = f.pool.Exec(f.ctx, `INSERT INTO workspace_settings(workspace_id,chat_enabled) VALUES($1,FALSE) ON CONFLICT(workspace_id) DO UPDATE SET chat_enabled=FALSE`, f.workspace)
	require.NoError(t, err)
	ids, err := repository.Pending(f.ctx, 100)
	require.NoError(t, err)
	require.Contains(t, ids, accepted.ID, "disabling chat cannot strand accounting for a saved no-response decision")

	_, err = f.pool.Exec(f.ctx, `ALTER TABLE messaging_assistant_daily_usage DROP CONSTRAINT test_usage_outage`)
	require.NoError(t, err)
	require.NoError(t, worker.Process(f.ctx, accepted.ID))
	finished, err := f.runs.Get(f.ctx, f.workspace, accepted.ID)
	require.NoError(t, err)
	require.Equal(t, agentruns.StatusSucceeded, finished.Status)
	require.Nil(t, finished.ReplyCommentID)
	require.Nil(t, finished.FailureReplyID)
	// A lost completion acknowledgement can retry accounting under a new lease.
	// The persisted original attempt key must still make it a no-op.
	finished.AttemptCount++
	require.NoError(t, f.runtime.FinalizeIgnored(f.ctx, finished, *finished.Proposal))
	snapshot, err := usage.Check(f.ctx, f.workspace, 0)
	require.NoError(t, err)
	require.EqualValues(t, 9, snapshot.TotalTokens)
	require.EqualValues(t, 1, snapshot.RequestCount)
	var replies int
	require.NoError(t, f.pool.QueryRow(f.ctx, `SELECT count(*) FROM chat_conversation_messages WHERE conversation_id=$1 AND author_id=$2`, f.conversation, f.maya).Scan(&replies))
	require.Equal(t, 1, replies)
}
