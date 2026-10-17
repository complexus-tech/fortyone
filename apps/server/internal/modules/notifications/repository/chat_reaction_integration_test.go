//go:build integration

package notificationsrepository

import (
	"context"
	"crypto/sha256"
	"testing"
	"time"

	notifications "github.com/complexus-tech/projects-api/internal/modules/notifications/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// Create the message, reaction, outbox event and resolved recipient as a single
// committed source event, including the timestamp shared by the real writer.
func reactionNotificationInput(t *testing.T, ctx context.Context, f notificationIntegrationFixture, thread bool) notifications.NewNotification {
	t.Helper()
	var conversation uuid.UUID
	require.NoError(t, f.postgres.Pool.QueryRow(ctx, "SELECT id FROM public.chat_conversations WHERE team_id=$1", f.teamA).Scan(&conversation))
	message, event := uuid.New(), uuid.New()
	digest := sha256.Sum256([]byte("reaction-notification"))
	tx, err := f.postgres.Pool.Begin(ctx)
	require.NoError(t, err)
	defer tx.Rollback(context.WithoutCancel(ctx))
	var sequence int64
	require.NoError(t, tx.QueryRow(ctx, `SELECT COALESCE(MAX(sequence),0)+1 FROM public.chat_conversation_messages WHERE conversation_id=$1`, conversation).Scan(&sequence))
	var rootID *uuid.UUID
	if thread {
		root := uuid.New()
		_, err = tx.Exec(ctx, `INSERT INTO public.chat_conversation_messages(id,workspace_id,conversation_id,sequence,author_id,creation_id,creation_hash,text) VALUES($1,$2,$3,$4,$5,$6,$7,'Thread root')`, root, f.workspaceA, conversation, sequence, f.actorA, uuid.New(), digest[:])
		require.NoError(t, err)
		rootID = &root
		sequence++
	}
	_, err = tx.Exec(ctx, `INSERT INTO public.chat_conversation_messages(id,workspace_id,conversation_id,sequence,author_id,creation_id,creation_hash,text,thread_root_id) VALUES($1,$2,$3,$4,$5,$6,$7,'My message',$8)`, message, f.workspaceA, conversation, sequence, f.recipientA, uuid.New(), digest[:], rootID)
	require.NoError(t, err)
	_, err = tx.Exec(ctx, `INSERT INTO public.chat_message_reactions(message_id,user_id,emoji) VALUES($1,$2,'👍')`, message, f.actorA)
	require.NoError(t, err)
	_, err = tx.Exec(ctx, `INSERT INTO public.chat_outbox(id,event_type,workspace_id,conversation_id,message_id,actor_id,thread_root_id,sequence,version,reaction_emoji) VALUES($1,'reaction.updated',$2,$3,$4,$5,$6,$7,1,'👍')`, event, f.workspaceA, conversation, message, f.actorA, rootID, sequence)
	require.NoError(t, err)
	_, err = tx.Exec(ctx, `INSERT INTO public.chat_outbox_recipients(event_id,user_id,reason) VALUES($1,$2,'reaction')`, event, f.recipientA)
	require.NoError(t, err)
	// Reading either stream must not suppress a later reaction to one's own message.
	_, err = tx.Exec(ctx, `INSERT INTO public.chat_conversation_reads(conversation_id,user_id,last_read_sequence) VALUES($1,$2,$3) ON CONFLICT(conversation_id,user_id) DO UPDATE SET last_read_sequence=EXCLUDED.last_read_sequence`, conversation, f.recipientA, sequence)
	require.NoError(t, err)
	if rootID != nil {
		_, err = tx.Exec(ctx, `INSERT INTO public.chat_thread_reads(thread_root_id,user_id,last_read_sequence) VALUES($1,$2,$3)`, rootID, f.recipientA, sequence)
		require.NoError(t, err)
	}
	require.NoError(t, tx.Commit(ctx))
	return notifications.NewNotification{
		DedupeKey: event.String(), OccurredAt: time.Now().UTC(), RecipientID: f.recipientA, WorkspaceID: f.workspaceA,
		ActorID: f.actorA, EntityID: message, Type: notifications.NotificationTypeChatMessage, EntityType: notifications.EntityTypeChatMessage,
		Title: "New reaction", Message: notifications.NotificationMessage{
			Template: "Someone reacted 👍 to your message", EmailSuppressed: true,
			ChatReaction: &notifications.ChatReaction{EventID: event, Emoji: "👍"},
		},
	}
}

func TestChatReactionReceiptAndPushIgnoreOriginalMessageReadState(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 2*time.Minute)
	defer cancel()
	f := newNotificationIntegrationFixture(t, ctx)
	for _, thread := range []bool{false, true} {
		input := reactionNotificationInput(t, ctx, f, thread)
		created, inserted, err := f.repo.Create(ctx, input)
		require.NoError(t, err)
		require.True(t, inserted)
		require.Equal(t, input.Message.ChatReaction, created.Message.ChatReaction)
		push, err := f.repo.GetPushDelivery(ctx, created.ID)
		require.NoError(t, err)
		require.NotNil(t, push, "reactions remain relevant after the source message was read")
		require.NotContains(t, string(push.Message), "chatReaction")
		replay, inserted, err := f.repo.Create(ctx, input)
		require.NoError(t, err)
		require.False(t, inserted)
		require.Equal(t, created.ID, replay.ID)
		input.Title = "Renamed team"
		input.Message.Template = "An updated message excerpt"
		replay, inserted, err = f.repo.Create(ctx, input)
		require.NoError(t, err)
		require.False(t, inserted)
		require.Equal(t, created.ID, replay.ID)
		email, err := f.repo.GetEmailDelivery(ctx, notifications.EmailNotificationQuery{Scope: notifications.DeliveryScope{RecipientID: f.recipientA, WorkspaceID: f.workspaceA}, NotificationID: created.ID})
		require.NoError(t, err)
		if email != nil {
			require.False(t, email.EmailEnabled)
		}
		_, err = f.postgres.Pool.Exec(ctx, `UPDATE public.notifications SET read_at=now() WHERE notification_id=$1`, created.ID)
		require.NoError(t, err)
		push, err = f.repo.GetPushDelivery(ctx, created.ID)
		require.NoError(t, err)
		require.Nil(t, push, "reading the notification itself still suppresses delayed push")
	}
}

func TestChatReactionReceiptAndPushRecheckLiveSource(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 2*time.Minute)
	defer cancel()
	f := newNotificationIntegrationFixture(t, ctx)
	for _, test := range []struct {
		name   string
		mutate func(notifications.NewNotification)
	}{
		{"removed reaction", func(input notifications.NewNotification) {
			mustNotificationExec(t, ctx, f.postgres.Pool, `DELETE FROM public.chat_message_reactions WHERE message_id=$1`, input.EntityID)
		}},
		{"removed and readded reaction", func(input notifications.NewNotification) {
			mustNotificationExec(t, ctx, f.postgres.Pool, `UPDATE public.chat_message_reactions SET created_at=created_at+interval '1 second' WHERE message_id=$1`, input.EntityID)
		}},
		{"deleted message", func(input notifications.NewNotification) {
			mustNotificationExec(t, ctx, f.postgres.Pool, `UPDATE public.chat_conversation_messages SET deleted_at=now() WHERE id=$1`, input.EntityID)
		}},
		{"wrong recipient reason", func(input notifications.NewNotification) {
			mustNotificationExec(t, ctx, f.postgres.Pool, `UPDATE public.chat_outbox_recipients SET reason='mention' WHERE event_id=$1`, input.Message.ChatReaction.EventID)
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			input := reactionNotificationInput(t, ctx, f, false)
			created, _, err := f.repo.Create(ctx, input)
			require.NoError(t, err)
			test.mutate(input)
			_, _, err = f.repo.Create(ctx, input)
			require.ErrorIs(t, err, notifications.ErrForbidden)
			push, err := f.repo.GetPushDelivery(ctx, created.ID)
			require.NoError(t, err)
			require.Nil(t, push)
		})
	}
}

func TestChatReactionRejectsSelfAndForgedSources(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 2*time.Minute)
	defer cancel()
	f := newNotificationIntegrationFixture(t, ctx)
	for _, test := range []struct {
		name   string
		modify func(*notifications.NewNotification)
	}{
		{"unknown event", func(input *notifications.NewNotification) { input.Message.ChatReaction.EventID = uuid.New() }},
		{"wrong emoji", func(input *notifications.NewNotification) { input.Message.ChatReaction.Emoji = "🎉" }},
		{"wrong actor", func(input *notifications.NewNotification) { input.ActorID = f.revocableA }},
		{"wrong message", func(input *notifications.NewNotification) { input.EntityID = uuid.New() }},
		{"wrong workspace", func(input *notifications.NewNotification) { input.WorkspaceID = f.workspaceB }},
		{"wrong recipient", func(input *notifications.NewNotification) { input.RecipientID = f.revocableA }},
		{"self reaction", func(input *notifications.NewNotification) {
			mustNotificationExec(t, ctx, f.postgres.Pool, `UPDATE public.chat_message_reactions SET user_id=$2 WHERE message_id=$1`, input.EntityID, f.recipientA)
			mustNotificationExec(t, ctx, f.postgres.Pool, `UPDATE public.chat_outbox SET actor_id=$2 WHERE id=$1`, input.Message.ChatReaction.EventID, f.recipientA)
			input.ActorID = f.recipientA
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			input := reactionNotificationInput(t, ctx, f, false)
			test.modify(&input)
			_, _, err := f.repo.Create(ctx, input)
			require.ErrorIs(t, err, notifications.ErrForbidden)
		})
	}
}

func TestChatReactionHonorsPushPreferencesAndCurrentMembership(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 2*time.Minute)
	defer cancel()
	for _, revokeActor := range []bool{false, true} {
		f := newNotificationIntegrationFixture(t, ctx)
		input := reactionNotificationInput(t, ctx, f, false)
		created, _, err := f.repo.Create(ctx, input)
		require.NoError(t, err)
		mustNotificationExec(t, ctx, f.postgres.Pool, `INSERT INTO public.notification_preferences(user_id,workspace_id,preferences) VALUES($1,$2,'{"chat_message":{"push":false}}')`, f.recipientA, f.workspaceA)
		push, err := f.repo.GetPushDelivery(ctx, created.ID)
		require.NoError(t, err)
		require.Nil(t, push)
		mustNotificationExec(t, ctx, f.postgres.Pool, `UPDATE public.notification_preferences SET preferences='{"chat_message":{"push":true,"in_app":false}}' WHERE user_id=$1 AND workspace_id=$2`, f.recipientA, f.workspaceA)
		fresh := reactionNotificationInput(t, ctx, f, false)
		suppressed, _, err := f.repo.Create(ctx, fresh)
		require.NoError(t, err)
		require.False(t, suppressed.InAppEnabled)
		push, err = f.repo.GetPushDelivery(ctx, suppressed.ID)
		require.NoError(t, err)
		require.Nil(t, push)
		user := f.recipientA
		if revokeActor {
			user = f.actorA
		}
		mustNotificationExec(t, ctx, f.postgres.Pool, `DELETE FROM public.team_members WHERE team_id=$1 AND user_id=$2`, f.teamA, user)
		_, _, err = f.repo.Create(ctx, input)
		require.ErrorIs(t, err, notifications.ErrForbidden)
		push, err = f.repo.GetPushDelivery(ctx, created.ID)
		require.NoError(t, err)
		require.Nil(t, push)
	}
}

func TestChatReactionReplayCannotReuseDifferentAdditionReceipt(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 2*time.Minute)
	defer cancel()
	f := newNotificationIntegrationFixture(t, ctx)
	original := reactionNotificationInput(t, ctx, f, false)
	created, _, err := f.repo.Create(ctx, original)
	require.NoError(t, err)
	newEvent := uuid.New()
	tx, err := f.postgres.Pool.Begin(ctx)
	require.NoError(t, err)
	defer tx.Rollback(context.WithoutCancel(ctx))
	_, err = tx.Exec(ctx, `DELETE FROM public.chat_message_reactions WHERE message_id=$1`, original.EntityID)
	require.NoError(t, err)
	_, err = tx.Exec(ctx, `INSERT INTO public.chat_message_reactions(message_id,user_id,emoji) VALUES($1,$2,'👍')`, original.EntityID, original.ActorID)
	require.NoError(t, err)
	_, err = tx.Exec(ctx, `INSERT INTO public.chat_outbox(id,event_type,workspace_id,conversation_id,message_id,actor_id,thread_root_id,sequence,version,reaction_emoji) SELECT $1,event_type,workspace_id,conversation_id,message_id,actor_id,thread_root_id,sequence,version,reaction_emoji FROM public.chat_outbox WHERE id=$2`, newEvent, original.Message.ChatReaction.EventID)
	require.NoError(t, err)
	_, err = tx.Exec(ctx, `INSERT INTO public.chat_outbox_recipients(event_id,user_id,reason) VALUES($1,$2,'reaction')`, newEvent, original.RecipientID)
	require.NoError(t, err)
	require.NoError(t, tx.Commit(ctx))
	_, _, err = f.repo.Create(ctx, original)
	require.ErrorIs(t, err, notifications.ErrForbidden, "new addition cannot revive the original delivery")
	push, err := f.repo.GetPushDelivery(ctx, created.ID)
	require.NoError(t, err)
	require.Nil(t, push)
	newAddition := original
	newAddition.Message.ChatReaction = &notifications.ChatReaction{EventID: newEvent, Emoji: "👍"}
	_, _, err = f.repo.Create(ctx, newAddition)
	require.ErrorIs(t, err, notifications.ErrConflict, "same message and actor do not allow reusing a different event's receipt")
	newAddition.DedupeKey = newEvent.String()
	latest, inserted, err := f.repo.Create(ctx, newAddition)
	require.NoError(t, err)
	require.True(t, inserted)
	require.NotEqual(t, created.ID, latest.ID)
	push, err = f.repo.GetPushDelivery(ctx, latest.ID)
	require.NoError(t, err)
	require.NotNil(t, push)
}
