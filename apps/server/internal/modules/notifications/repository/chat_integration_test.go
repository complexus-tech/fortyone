//go:build integration

package notificationsrepository

import (
	"context"
	"crypto/sha256"
	"errors"
	"testing"
	"time"

	notifications "github.com/complexus-tech/projects-api/internal/modules/notifications/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestChatNotificationsFollowLiveConversationAccess(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 2*time.Minute)
	defer cancel()
	f := newNotificationIntegrationFixture(t, ctx)
	var conversation uuid.UUID
	if err := f.postgres.Pool.QueryRow(ctx, "SELECT id FROM public.chat_conversations WHERE team_id=$1", f.teamA).Scan(&conversation); err != nil {
		t.Fatal(err)
	}
	message := uuid.New()
	digest := sha256.Sum256([]byte("test-chat-notification"))
	_, err := f.postgres.Pool.Exec(ctx, `INSERT INTO public.chat_conversation_messages(id,workspace_id,conversation_id,sequence,author_id,creation_id,creation_hash,text) VALUES($1,$2,$3,1,$4,$5,$6,'Test chat message')`, message, f.workspaceA, conversation, f.actorA, uuid.New(), digest[:])
	if err != nil {
		t.Fatal(err)
	}
	input := notifications.NewNotification{DedupeKey: uuid.NewString(), OccurredAt: time.Now(), RecipientID: f.recipientA, WorkspaceID: f.workspaceA, ActorID: f.actorA, EntityID: message, Type: notifications.NotificationTypeChatMessage, EntityType: notifications.EntityTypeChatMessage, Title: "Mentioned in chat", Message: notifications.NotificationMessage{Template: "You were mentioned in a conversation."}}
	created, inserted, err := f.repo.Create(ctx, input)
	if err != nil || !inserted {
		t.Fatalf("create chat notification: inserted=%v err=%v", inserted, err)
	}
	access := notifications.WorkspaceAccess{ActorID: f.recipientA, WorkspaceID: f.workspaceA}
	items, err := f.repo.List(ctx, notifications.ListQuery{Access: access, Limit: 25})
	if err != nil || len(items) != 0 {
		t.Fatalf("chat receipt leaked into work inbox: items=%d err=%v", len(items), err)
	}
	count, err := f.repo.CountUnread(ctx, access)
	require.NoError(t, err)
	require.Zero(t, count, "chat receipts must not contribute to the work badge")
	for _, kind := range []notifications.WorkspaceMutationKind{notifications.WorkspaceMutationReadAll, notifications.WorkspaceMutationDeleteAll} {
		affected, err := f.repo.MutateAll(ctx, notifications.WorkspaceMutation{Access: access, Kind: kind, At: time.Now().UTC()})
		require.NoError(t, err)
		require.Zero(t, affected, "work inbox bulk operations must leave chat delivery alone")
	}
	delivery, err := f.repo.GetPushDelivery(ctx, created.ID)
	require.NoError(t, err)
	require.NotNil(t, delivery, "hiding chat from the work inbox must preserve mobile push")
	replay, inserted, err := f.repo.Create(ctx, input)
	if err != nil || inserted || replay.ID != created.ID {
		t.Fatalf("chat replay duplicated: %v %v", inserted, err)
	}
	_, err = f.postgres.Pool.Exec(ctx, "DELETE FROM public.team_members WHERE team_id=$1 AND user_id=$2", f.teamA, f.recipientA)
	if err != nil {
		t.Fatal(err)
	}
	items, err = f.repo.List(ctx, notifications.ListQuery{Access: access, Limit: 25})
	if err != nil || len(items) != 0 {
		t.Fatalf("revoked notification still visible: items=%d err=%v", len(items), err)
	}
	count, err = f.repo.CountUnread(ctx, access)
	if err != nil || count != 0 {
		t.Fatalf("revoked unread count: %d %v", count, err)
	}
	delivery, err = f.repo.GetPushDelivery(ctx, created.ID)
	if err != nil || delivery != nil {
		t.Fatalf("revoked push delivery: %+v %v", delivery, err)
	}
	input.DedupeKey = uuid.NewString()
	revoked, _, err := f.repo.Create(ctx, input)
	if !errors.Is(err, notifications.ErrForbidden) || revoked.ID != uuid.Nil {
		t.Fatalf("created notification after revocation: %+v %v", revoked, err)
	}
}

func TestChatReadStateSuppressesNewReceiptAndPushReplay(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 2*time.Minute)
	defer cancel()
	f := newNotificationIntegrationFixture(t, ctx)
	var conversation uuid.UUID
	require.NoError(t, f.postgres.Pool.QueryRow(ctx, "SELECT id FROM public.chat_conversations WHERE team_id=$1", f.teamA).Scan(&conversation))
	root, reply := uuid.New(), uuid.New()
	digest := sha256.Sum256([]byte("chat-read-delivery"))
	_, err := f.postgres.Pool.Exec(ctx, `INSERT INTO public.chat_conversation_messages(id,workspace_id,conversation_id,sequence,author_id,creation_id,creation_hash,text,thread_root_id) VALUES($1,$2,$3,1,$4,$5,$6,'Root',NULL),($7,$2,$3,2,$4,$8,$6,'Reply',$1)`, root, f.workspaceA, conversation, f.actorA, uuid.New(), digest[:], reply, uuid.New())
	require.NoError(t, err)
	input := notifications.NewNotification{DedupeKey: uuid.NewString(), OccurredAt: time.Now().UTC(), RecipientID: f.recipientA, WorkspaceID: f.workspaceA, ActorID: f.actorA, EntityID: root, Type: notifications.NotificationTypeChatMessage, EntityType: notifications.EntityTypeChatMessage, Title: "New chat message", Message: notifications.NotificationMessage{Template: "New message"}}
	rootReceipt, _, err := f.repo.Create(ctx, input)
	require.NoError(t, err)
	_, err = f.postgres.Pool.Exec(ctx, `INSERT INTO public.chat_conversation_reads(conversation_id,user_id,last_read_sequence) VALUES($1,$2,2)`, conversation, f.recipientA)
	require.NoError(t, err)
	delivery, err := f.repo.GetPushDelivery(ctx, rootReceipt.ID)
	require.NoError(t, err)
	require.Nil(t, delivery, "read root messages must not be pushed by delayed jobs")
	_, _, err = f.repo.Create(ctx, input)
	require.ErrorIs(t, err, notifications.ErrForbidden, "receipt replay must recheck read state")
	input.DedupeKey = uuid.NewString()
	_, _, err = f.repo.Create(ctx, input)
	require.ErrorIs(t, err, notifications.ErrForbidden, "already read messages must not create new delivery")
	input.EntityID, input.DedupeKey = reply, uuid.NewString()
	replyReceipt, _, err := f.repo.Create(ctx, input)
	require.NoError(t, err, "reading the conversation must not clear independent unread replies")
	delivery, err = f.repo.GetPushDelivery(ctx, replyReceipt.ID)
	require.NoError(t, err)
	require.NotNil(t, delivery)
	_, err = f.postgres.Pool.Exec(ctx, `INSERT INTO public.chat_thread_reads(thread_root_id,user_id,last_read_sequence) VALUES($1,$2,2)`, root, f.recipientA)
	require.NoError(t, err)
	delivery, err = f.repo.GetPushDelivery(ctx, replyReceipt.ID)
	require.NoError(t, err)
	require.Nil(t, delivery, "read thread replies must not be pushed by delayed jobs")
}

func TestOrdinaryChatEmailSuppressionPersistsWithoutChangingTargetedDelivery(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 2*time.Minute)
	defer cancel()
	f := newNotificationIntegrationFixture(t, ctx)
	var conversation uuid.UUID
	require.NoError(t, f.postgres.Pool.QueryRow(ctx, "SELECT id FROM public.chat_conversations WHERE team_id=$1", f.teamA).Scan(&conversation))
	message := uuid.New()
	digest := sha256.Sum256([]byte("chat-email-policy"))
	_, err := f.postgres.Pool.Exec(ctx, `INSERT INTO public.chat_conversation_messages(id,workspace_id,conversation_id,sequence,author_id,creation_id,creation_hash,text) VALUES($1,$2,$3,1,$4,$5,$6,'Team update')`, message, f.workspaceA, conversation, f.actorA, uuid.New(), digest[:])
	require.NoError(t, err)
	input := notifications.NewNotification{DedupeKey: uuid.NewString(), OccurredAt: time.Now().UTC(), RecipientID: f.recipientA, WorkspaceID: f.workspaceA, ActorID: f.actorA, EntityID: message, Type: notifications.NotificationTypeChatMessage, EntityType: notifications.EntityTypeChatMessage, Title: "New chat message", Message: notifications.NotificationMessage{Template: "New message", EmailSuppressed: true}}
	created, _, err := f.repo.Create(ctx, input)
	require.NoError(t, err)
	require.True(t, created.Message.EmailSuppressed)
	scope := notifications.DeliveryScope{RecipientID: f.recipientA, WorkspaceID: f.workspaceA}
	email, err := f.repo.GetEmailDelivery(ctx, notifications.EmailNotificationQuery{Scope: scope, NotificationID: created.ID})
	require.NoError(t, err)
	require.NotNil(t, email)
	require.False(t, email.EmailEnabled, "explicit per-message policy must override enabled account email preferences")
	pending, err := f.repo.ListEmailDigest(ctx, scope)
	require.NoError(t, err)
	require.Nil(t, pending, "another notification's digest wake-up must not pick up ordinary team messages")
	push, err := f.repo.GetPushDelivery(ctx, created.ID)
	require.NoError(t, err)
	require.NotNil(t, push)
	input.DedupeKey, input.Title, input.Message.EmailSuppressed = uuid.NewString(), "Mentioned in chat", false
	targeted, _, err := f.repo.Create(ctx, input)
	require.NoError(t, err)
	email, err = f.repo.GetEmailDelivery(ctx, notifications.EmailNotificationQuery{Scope: scope, NotificationID: targeted.ID})
	require.NoError(t, err)
	require.NotNil(t, email)
	require.True(t, email.EmailEnabled, "existing targeted delivery retains current account preferences")
	pending, err = f.repo.ListEmailDigest(ctx, scope)
	require.NoError(t, err)
	require.NotNil(t, pending)
}
