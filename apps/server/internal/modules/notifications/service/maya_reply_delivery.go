package notifications

import (
	"context"
	"fmt"

	domain "github.com/complexus-tech/projects-api/internal/modules/notifications/domain"
	"github.com/google/uuid"
)

type mayaReplyNotificationStore interface {
	DeferMayaReply(context.Context, domain.NewNotification, uuid.UUID) error
	DispatchMayaReplies(context.Context, int, func(context.Context, domain.NewNotification) error) (int, error)
}

// DeferMayaReply persists a decision independently of invocation completion.
// Only ordinary reply notifications are deferred; explicit mentions stay immediate.
func (s *Service) DeferMayaReply(ctx context.Context, input CoreNewNotification, commentID uuid.UUID) error {
	if err := input.Validate(); err != nil {
		return err
	}
	if input.Type != domain.NotificationTypeCommentReply || input.EntityType != domain.EntityTypeStory ||
		commentID == uuid.Nil || input.DedupeKey != "maya-reply:"+commentID.String()+":"+input.RecipientID.String() ||
		input.Message.Variables["commentId"].Value != commentID.String() || input.OccurredAt.IsZero() || input.InAppEnabled != nil {
		return domain.ErrInvalid
	}
	store, ok := s.repo.(mayaReplyNotificationStore)
	if !ok {
		return fmt.Errorf("Maya reply notification store unavailable")
	}
	return store.DeferMayaReply(ctx, input, commentID)
}

func (s *Service) DispatchMayaReplies(ctx context.Context, limit int) (int, error) {
	store, ok := s.repo.(mayaReplyNotificationStore)
	if !ok {
		return 0, fmt.Errorf("Maya reply notification store unavailable")
	}
	return store.DispatchMayaReplies(ctx, limit, s.deliverMayaReply)
}

func (s *Service) deliverMayaReply(ctx context.Context, input domain.NewNotification) error {
	notification, _, err := s.repo.Create(ctx, input)
	if err != nil || notification.ID == uuid.Nil {
		return err
	}
	// A prior attempt may have committed the inbox row before Redis failed.
	// Replaying its same ID recovers realtime delivery as well as push/email.
	if notification.InAppEnabled && (input.InAppEnabled == nil || *input.InAppEnabled) {
		if err := s.publishRealtime(ctx, notification.Public()); err != nil {
			return err
		}
		if err := s.enqueuePush(notification); err != nil {
			return err
		}
	}
	return s.enqueueDigest(notification)
}
