package invitations

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/complexus-tech/projects-api/pkg/events"
	"github.com/complexus-tech/projects-api/pkg/logger"
)

const (
	invitationOutboxInitialDelay = 30 * time.Second
	invitationOutboxMaximumDelay = 15 * time.Minute
)

// InvitationNotificationSender is deliberately separate from a general event
// publisher because the raw invitation bearer must exist only at the final
// email boundary. Implementations use the durable idempotency key for a stable
// provider/message identity.
type InvitationNotificationSender interface {
	SendInvitationEmail(context.Context, InvitationEmailDelivery) error
	SendInvitationAccepted(context.Context, events.InvitationAcceptedPayload, string) error
}

type OutboxDispatcher struct {
	log           *logger.Logger
	repo          InvitationOutboxRepository
	tokens        *InvitationTokenManager
	notifications InvitationNotificationSender
	now           func() time.Time
}

func NewOutboxDispatcher(
	log *logger.Logger,
	repo InvitationOutboxRepository,
	tokens *InvitationTokenManager,
	notifications InvitationNotificationSender,
) *OutboxDispatcher {
	return &OutboxDispatcher{log: log, repo: repo, tokens: tokens, notifications: notifications, now: time.Now}
}

type permanentOutboxError struct{ err error }

func (e permanentOutboxError) Error() string { return e.err.Error() }
func (e permanentOutboxError) Unwrap() error { return e.err }

func (d *OutboxDispatcher) DispatchReady(ctx context.Context) error {
	if d == nil || d.repo == nil || d.tokens == nil || d.notifications == nil || d.now == nil {
		return errors.New("invitation outbox dispatcher is not configured")
	}
	now := d.now().UTC()
	claimed, err := d.repo.ClaimInvitationOutboxEvents(
		ctx,
		invitationOutboxBatchSize,
		now,
		now.Add(-invitationOutboxStaleFor),
	)
	if err != nil {
		return err
	}

	var lifecycleErrors error
	for _, outbox := range claimed {
		dispatchErr := d.dispatch(ctx, outbox, now)
		if dispatchErr == nil {
			completeCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
			err := d.repo.CompleteInvitationOutboxEvent(completeCtx, outbox.ID, outbox.ClaimToken, d.now().UTC())
			cancel()
			if err != nil {
				lifecycleErrors = errors.Join(lifecycleErrors, fmt.Errorf("complete invitation outbox %s: %w", outbox.ID, err))
			}
			continue
		}

		terminal := outbox.AttemptCount >= invitationOutboxMaxTries
		var permanent permanentOutboxError
		if errors.As(dispatchErr, &permanent) {
			terminal = true
		}
		releasedAt := d.now().UTC()
		releaseCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		releaseErr := d.repo.RetryInvitationOutboxEvent(
			releaseCtx,
			outbox.ID,
			outbox.ClaimToken,
			dispatchErr.Error(),
			releasedAt.Add(invitationOutboxRetryDelay(outbox.AttemptCount)),
			releasedAt,
			terminal,
		)
		cancel()
		if releaseErr != nil {
			lifecycleErrors = errors.Join(lifecycleErrors, fmt.Errorf("release invitation outbox %s: %w", outbox.ID, releaseErr))
		} else if d.log != nil {
			d.log.Error(ctx, "failed to dispatch invitation outbox", "outbox_id", outbox.ID, "error", dispatchErr)
		}
	}
	return lifecycleErrors
}

func (d *OutboxDispatcher) dispatch(
	ctx context.Context,
	outbox CoreInvitationOutboxEvent,
	now time.Time,
) error {
	switch events.EventType(outbox.EventType) {
	case events.InvitationEmail:
		// Superseded, accepted, revoked, or expired invitations complete without
		// sending a stale credential.
		if outbox.InvitationUsedAt != nil || outbox.InvitationExpiresAt == nil || !outbox.InvitationExpiresAt.After(now) {
			return nil
		}
		if outbox.StoredToken == nil {
			return permanentOutboxError{err: errors.New("invitation email token metadata is unavailable")}
		}
		var storedPayload InvitationEmailOutboxPayload
		if err := decodeOutboxPayload(outbox.EventPayload, &storedPayload); err != nil {
			return permanentOutboxError{err: fmt.Errorf("decode invitation email outbox: %w", err)}
		}
		rawToken, err := d.tokens.Restore(*outbox.StoredToken)
		if err != nil {
			return permanentOutboxError{err: errors.New("restore invitation email token")}
		}
		if err := d.notifications.SendInvitationEmail(ctx, InvitationEmailDelivery{
			IdempotencyKey: outbox.IdempotencyKey,
			InviterName:    storedPayload.InviterName,
			Email:          storedPayload.Email,
			Token:          rawToken,
			Role:           storedPayload.Role,
			ExpiresAt:      storedPayload.ExpiresAt,
			WorkspaceID:    storedPayload.WorkspaceID,
			WorkspaceName:  storedPayload.WorkspaceName,
		}); err != nil {
			return fmt.Errorf("send invitation email: %w", err)
		}
		return nil
	case events.InvitationAccepted:
		var accepted events.InvitationAcceptedPayload
		if err := decodeOutboxPayload(outbox.EventPayload, &accepted); err != nil {
			return permanentOutboxError{err: fmt.Errorf("decode invitation acceptance outbox: %w", err)}
		}
		if err := d.notifications.SendInvitationAccepted(ctx, accepted, outbox.IdempotencyKey); err != nil {
			return fmt.Errorf("send invitation acceptance notification: %w", err)
		}
		return nil
	default:
		return permanentOutboxError{err: fmt.Errorf("unsupported invitation outbox event type %q", outbox.EventType)}
	}
}

func decodeOutboxPayload(payload []byte, destination any) error {
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		if err == nil {
			return errors.New("outbox payload contains more than one JSON value")
		}
		return err
	}
	return nil
}

func invitationOutboxRetryDelay(attempt int) time.Duration {
	if attempt <= 1 {
		return invitationOutboxInitialDelay
	}
	delay := invitationOutboxInitialDelay
	for current := 1; current < attempt && delay < invitationOutboxMaximumDelay; current++ {
		delay *= 2
		if delay >= invitationOutboxMaximumDelay {
			return invitationOutboxMaximumDelay
		}
	}
	return delay
}
