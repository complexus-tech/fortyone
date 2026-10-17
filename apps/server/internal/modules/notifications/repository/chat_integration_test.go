//go:build integration

package notificationsrepository

import (
	"context"
	"crypto/sha256"
	"errors"
	notifications "github.com/complexus-tech/projects-api/internal/modules/notifications/domain"
	"github.com/google/uuid"
	"testing"
	"time"
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
	if err != nil || len(items) != 1 {
		t.Fatalf("visible chat notification: items=%d err=%v", len(items), err)
	}
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
	count, err := f.repo.CountUnread(ctx, access)
	if err != nil || count != 0 {
		t.Fatalf("revoked unread count: %d %v", count, err)
	}
	delivery, err := f.repo.GetPushDelivery(ctx, created.ID)
	if err != nil || delivery != nil {
		t.Fatalf("revoked push delivery: %+v %v", delivery, err)
	}
	input.DedupeKey = uuid.NewString()
	revoked, _, err := f.repo.Create(ctx, input)
	if !errors.Is(err, notifications.ErrForbidden) || revoked.ID != uuid.Nil {
		t.Fatalf("created notification after revocation: %+v %v", revoked, err)
	}
}
