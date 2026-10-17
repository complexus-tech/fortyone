package notifications

import (
	"context"
	"fmt"
	domain "github.com/complexus-tech/projects-api/internal/modules/notifications/domain"
)

type documentNotificationDispatcher interface {
	DispatchDocumentNotifications(context.Context, int, func(context.Context, domain.Notification) error) (int, error)
}

// DispatchDocumentNotifications bridges committed document notifications into
// the same realtime, push and email machinery as ordinary notifications.
func (s *Service) DispatchDocumentNotifications(ctx context.Context, limit int) (int, error) {
	dispatcher, ok := s.repo.(documentNotificationDispatcher)
	if !ok {
		return 0, fmt.Errorf("document notification dispatcher unavailable")
	}
	return dispatcher.DispatchDocumentNotifications(ctx, limit, func(ctx context.Context, n domain.Notification) error {
		if n.InAppEnabled {
			if err := s.publishRealtime(ctx, n.Public()); err != nil {
				return err
			}
			if err := s.enqueuePush(n); err != nil {
				return err
			}
		}
		return s.enqueueDigest(n)
	})
}
