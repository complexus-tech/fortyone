package chatadapter

import (
	"context"
	"errors"
	"fmt"
	"time"

	chatdomain "github.com/complexus-tech/projects-api/internal/modules/chat/domain"
	notifications "github.com/complexus-tech/projects-api/internal/modules/notifications/domain"
	"github.com/complexus-tech/projects-api/pkg/publisher"
	"github.com/google/uuid"
	"golang.org/x/sync/errgroup"
)

type EventStore interface {
	ClaimEvents(context.Context, int) ([]chatdomain.Event, error)
	CompleteEvent(context.Context, uuid.UUID, uuid.UUID) error
	ReleaseEvent(context.Context, uuid.UUID, uuid.UUID) error
	RenewEventLease(context.Context, uuid.UUID, uuid.UUID) error
	ReadRecipientIDs(context.Context, chatdomain.Event) ([]uuid.UUID, error)
	RecipientIDs(context.Context, uuid.UUID, uuid.UUID) ([]uuid.UUID, error)
	NotificationRecipients(context.Context, chatdomain.Event) ([]chatdomain.NotificationRecipient, error)
	NotificationContent(context.Context, chatdomain.Event) (chatdomain.NotificationContent, error)
}
type UpdatePublisher interface {
	PublishChatUpdated(context.Context, publisher.ChatUpdate) error
	PublishChatNotification(context.Context, publisher.ChatNotification) error
}
type NotificationCreator interface {
	Create(context.Context, notifications.NewNotification) (notifications.Notification, error)
}

type Dispatcher struct {
	Store         EventStore
	Updates       UpdatePublisher
	Notifications NotificationCreator
	leaseInterval time.Duration
}

// Dispatch claims committed events. Delivery is at least once: cache hints are
// repeatable and notification receipts deduplicate by event and recipient.
func (d Dispatcher) Dispatch(ctx context.Context, limit int) (int, error) {
	if d.Store == nil || d.Updates == nil || d.Notifications == nil {
		return 0, errors.New("chat delivery dependencies unavailable")
	}
	if limit < 1 || limit > 100 {
		limit = 50
	}
	completed := 0
	var failures []error
	// Claim only the event being processed. A large room must not consume the
	// lease time of another event waiting behind it in this worker.
	for processed := 0; processed < limit; processed++ {
		if err := ctx.Err(); err != nil {
			return completed, errors.Join(append(failures, err)...)
		}
		events, err := d.Store.ClaimEvents(ctx, 1)
		if err != nil {
			return completed, errors.Join(append(failures, fmt.Errorf("claim chat event: %w", err))...)
		}
		if len(events) == 0 {
			break
		}
		event := events[0]
		if err := d.deliverWithLease(ctx, event); err != nil {
			cleanup, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
			releaseErr := d.Store.ReleaseEvent(cleanup, event.ID, event.LeaseToken)
			cancel()
			failures = append(failures, fmt.Errorf("deliver chat event %s: %w", event.ID, errors.Join(err, releaseErr)))
			continue
		}
		if err := d.Store.CompleteEvent(ctx, event.ID, event.LeaseToken); err != nil {
			failures = append(failures, fmt.Errorf("complete chat event %s: %w", event.ID, err))
			continue
		}
		completed++
	}
	return completed, errors.Join(failures...)
}

func (d Dispatcher) deliver(ctx context.Context, event chatdomain.Event) error {
	var recipients []uuid.UUID
	var err error
	if event.Type == "read.updated" {
		recipients, err = d.Store.ReadRecipientIDs(ctx, event)
	} else {
		recipients, err = d.Store.RecipientIDs(ctx, event.WorkspaceID, event.ConversationID)
	}
	if err != nil {
		return err
	}
	hints, hintCtx := errgroup.WithContext(ctx)
	hints.SetLimit(16)
	for _, recipient := range recipients {
		if hintCtx.Err() != nil {
			break
		}
		update := publisher.ChatUpdate{EventID: event.ID, WorkspaceID: event.WorkspaceID, UserID: recipient, ConversationID: event.ConversationID, Change: event.Type, Version: int64(event.Version)}
		if event.MessageID != uuid.Nil {
			messageID := event.MessageID
			update.MessageID = &messageID
		}
		hints.Go(func() error { return d.Updates.PublishChatUpdated(hintCtx, update) })
	}
	if err := hints.Wait(); err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if event.Type == "read.updated" {
		return nil
	}
	// The store resolves notification eligibility against current membership,
	// deletion state and the recipient's independent conversation/thread reads.
	targets, err := d.Store.NotificationRecipients(ctx, event)
	if err != nil {
		return err
	}
	if len(targets) == 0 {
		return nil
	}
	content, err := d.Store.NotificationContent(ctx, event)
	if errors.Is(err, chatdomain.ErrNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	title, body := notificationCopy(content)
	seen := make(map[uuid.UUID]struct{}, len(targets))
	for _, target := range targets {
		if target.UserID == uuid.Nil || target.UserID == event.ActorID {
			continue
		}
		if _, exists := seen[target.UserID]; exists {
			continue
		}
		seen[target.UserID] = struct{}{}
		switch target.Reason {
		case "mention", "reply", "dm", "message":
		case "reaction":
			if event.Type != "reaction.updated" || content.ReactionEmoji == "" {
				continue
			}
		default:
			continue
		}
		message := notifications.NotificationMessage{
			Template: "{message}", Variables: map[string]notifications.Variable{
				"message": {Value: body, Type: "plain_text"},
			},
			EmailSuppressed: target.Reason == "message" || target.Reason == "reaction",
		}
		if target.Reason == "reaction" {
			message.ChatReaction = &notifications.ChatReaction{EventID: event.ID, Emoji: content.ReactionEmoji}
		}
		notification, err := d.Notifications.Create(ctx, notifications.NewNotification{
			DedupeKey: fmt.Sprintf("chat:%s:%s", event.ID, target.UserID), OccurredAt: event.CreatedAt,
			WorkspaceID: event.WorkspaceID, RecipientID: target.UserID, ActorID: event.ActorID,
			Type: notifications.NotificationTypeChatMessage, EntityType: notifications.EntityTypeChatMessage, EntityID: event.MessageID,
			Title: title, Message: message,
		})
		if errors.Is(err, notifications.ErrForbidden) || errors.Is(err, notifications.ErrNotFound) {
			// A membership revocation or deletion can race the recipient lookup.
			// There is nothing left to deliver to this recipient.
			continue
		}
		if err != nil {
			return err
		}
		if notification.ID == uuid.Nil || !notification.InAppEnabled || notification.ReadAt != nil {
			continue
		}
		// A retry returns the original receipt even if the message was edited
		// meanwhile. Native, browser and mobile delivery must show the same copy.
		body := notification.Message.Template
		if message, ok := notification.Message.Variables["message"]; ok && body == "{message}" {
			body = message.Value
		}
		// Publish even when Create reuses an existing receipt: the earlier attempt
		// may have persisted successfully but failed before live publication.
		// Clients deduplicate this stable event/receipt pair.
		if err := d.Updates.PublishChatNotification(ctx, publisher.ChatNotification{
			NotificationID: notification.ID, EventID: event.ID,
			WorkspaceID: event.WorkspaceID, UserID: target.UserID, ActorID: event.ActorID,
			ConversationID: event.ConversationID, MessageID: event.MessageID, ThreadRootID: event.ThreadRootID,
			CreatedAt: notification.CreatedAt, OccurredAt: event.CreatedAt,
			Title: notification.Title, Body: body,
		}); err != nil {
			return err
		}
	}
	return nil
}

// Renew only the active event. If ownership is lost, cancel pending hints and
// notification writes; a replacement worker owns recovery from that point.
func (d Dispatcher) deliverWithLease(ctx context.Context, event chatdomain.Event) error {
	work, cancel := context.WithCancel(ctx)
	interval := d.leaseInterval
	if interval <= 0 {
		interval = 20 * time.Second
	}
	renewErrors := make(chan error, 1)
	done := make(chan struct{})
	go func() {
		defer close(done)
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-work.Done():
				return
			case <-ticker.C:
				renew, stopRenew := context.WithTimeout(work, 5*time.Second)
				err := d.Store.RenewEventLease(renew, event.ID, event.LeaseToken)
				stopRenew()
				if err != nil {
					if work.Err() != nil && errors.Is(err, context.Canceled) {
						return
					}
					renewErrors <- err
					cancel()
					return
				}
			}
		}
	}()
	deliveryErr := d.deliver(work, event)
	cancel()
	<-done
	select {
	case err := <-renewErrors:
		return errors.Join(deliveryErr, fmt.Errorf("renew chat event lease: %w", err))
	default:
		return deliveryErr
	}
}
