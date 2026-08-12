package emailthread

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"testing"
	"time"

	messaging "github.com/complexus-tech/projects-api/internal/modules/messaging/service"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type storeStub struct {
	threadInput messaging.EmailThreadInput
	aliasInput  messaging.EmailReplyTokenInput
	messages    []messaging.EmailMessageInput
	thread      messaging.EmailThreadRecord
}

func (s *storeStub) CreateEmailThread(_ context.Context, input messaging.EmailThreadInput) (messaging.EmailThreadRecord, bool, error) {
	s.threadInput = input
	return s.thread, true, nil
}

func (s *storeStub) CreateEmailReplyTokenAlias(_ context.Context, input messaging.EmailReplyTokenInput) (messaging.EmailReplyTokenRecord, bool, error) {
	s.aliasInput = input
	return messaging.EmailReplyTokenRecord{ThreadID: input.ThreadID}, true, nil
}

func (s *storeStub) AppendEmailMessage(_ context.Context, input messaging.EmailMessageInput) (messaging.EmailMessageRecord, bool, error) {
	s.messages = append(s.messages, input)
	messageID := input.InternetMessageID
	inReplyTo := input.InReplyToMessageID
	return messaging.EmailMessageRecord{
		ID:                 uuid.New(),
		ThreadID:           input.ThreadID,
		Subject:            input.Subject,
		Content:            input.Content,
		InternetMessageID:  &messageID,
		InReplyToMessageID: &inReplyTo,
	}, true, nil
}

func TestPrepareGuidanceCreatesOpaqueReplyAddressAndHistory(t *testing.T) {
	thread := messaging.EmailThreadRecord{ID: uuid.New(), WorkspaceID: uuid.New(), UserID: uuid.New()}
	store := &storeStub{thread: thread}
	service, err := New(store)
	require.NoError(t, err)
	service.now = func() time.Time { return time.Date(2026, 8, 12, 9, 0, 0, 0, time.UTC) }
	service.randomBytes = func(value []byte) (int, error) {
		for index := range value {
			value[index] = byte(index + 1)
		}
		return len(value), nil
	}

	prepared, err := service.PrepareGuidance(context.Background(), GuidanceInput{
		WorkspaceID:       thread.WorkspaceID,
		UserID:            thread.UserID,
		RecipientEmail:    " Joseph@Example.com ",
		ExternalThreadID:  "weekly:2026-W33",
		InternetMessageID: "<weekly@fortyone.app>",
		Subject:           "Your weekly strategy check-in",
		Content:           "Three objectives need attention.",
		Context:           json.RawMessage(`{"kind":"weekly"}`),
	})
	require.NoError(t, err)
	require.Regexp(t, `^maya\+[A-Za-z0-9_-]{43}@reply\.fortyone\.app$`, prepared.ReplyTo)
	require.Equal(t, "joseph@example.com", store.threadInput.RecipientEmail)
	require.Len(t, store.threadInput.ReplyTokenHash, sha256.Size)
	require.Equal(t, messaging.EmailMessageKindGuidance, store.messages[0].Kind)
	require.Equal(t, "Three objectives need attention.", store.messages[0].Content)
}

func TestPrepareReplyRotatesTokenAndPreservesRFCThread(t *testing.T) {
	thread := messaging.EmailThreadRecord{
		ID:                    uuid.New(),
		WorkspaceID:           uuid.New(),
		UserID:                uuid.New(),
		RootInternetMessageID: "<root@fortyone.app>",
	}
	store := &storeStub{thread: thread}
	service, err := New(store)
	require.NoError(t, err)

	prepared, err := service.PrepareReply(context.Background(), ReplyInput{
		Thread:            thread,
		InternetMessageID: "<maya-reply@fortyone.app>",
		InReplyTo:         "<user-reply@example.com>",
		Subject:           "Re: Your strategy check-in",
		Content:           "I can update the objective after confirmation.",
		Kind:              messaging.EmailMessageKindProposal,
		IdempotencyKey:    "event:proposal",
	})
	require.NoError(t, err)
	require.Equal(t, thread.ID, store.aliasInput.ThreadID)
	require.Equal(t, []string{"<root@fortyone.app>", "<user-reply@example.com>"}, prepared.References)

	email := ThreadedEmail(prepared, "joseph@example.com", "<p>Preview</p>", "Preview")
	require.Equal(t, "<user-reply@example.com>", email.InReplyTo)
	require.Equal(t, []string{"<root@fortyone.app>", "<user-reply@example.com>"}, email.References)
	require.Equal(t, "Preview", email.PlainTextBody)
}
