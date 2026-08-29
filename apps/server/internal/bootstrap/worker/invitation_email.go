package workerbootstrap

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"math"
	"net/url"
	"strings"
	"time"

	invitations "github.com/complexus-tech/projects-api/internal/modules/invitations/service"
	"github.com/complexus-tech/projects-api/pkg/events"
	"github.com/complexus-tech/projects-api/pkg/mailer"
)

type invitationEmailSender struct {
	mailer     mailer.Service
	websiteURL string
	now        func() time.Time
}

func newInvitationEmailSender(service mailer.Service, websiteURL string) (*invitationEmailSender, error) {
	if service == nil {
		return nil, errors.New("invitation email mailer is required")
	}
	websiteURL = strings.TrimRight(strings.TrimSpace(websiteURL), "/")
	parsed, err := url.Parse(websiteURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return nil, errors.New("invitation email website URL must be an absolute URL")
	}
	return &invitationEmailSender{mailer: service, websiteURL: websiteURL, now: time.Now}, nil
}

func (s *invitationEmailSender) SendInvitationEmail(
	ctx context.Context,
	delivery invitations.InvitationEmailDelivery,
) error {
	if strings.TrimSpace(delivery.Token) == "" {
		return errors.New("invitation email token is required")
	}
	expiresIn := delivery.ExpiresAt.Sub(s.now())
	if expiresIn <= 0 {
		return errors.New("invitation has expired before email delivery")
	}

	unit, amount := "hours", int(math.Ceil(expiresIn.Hours()))
	if expiresIn >= 24*time.Hour {
		unit, amount = "days", int(math.Ceil(expiresIn.Hours()/24))
	}
	if amount == 1 {
		unit = strings.TrimSuffix(unit, "s")
	}

	joinURL := s.websiteURL + "/onboarding/join?token=" + url.QueryEscape(delivery.Token)
	messageID, err := invitationMessageID(delivery.IdempotencyKey)
	if err != nil {
		return err
	}
	if err := s.mailer.SendTemplated(ctx, mailer.TemplatedEmail{
		To:        []string{delivery.Email},
		Template:  "invites/invitation",
		Subject:   fmt.Sprintf("%s invited you to work with them in FortyOne", delivery.InviterName),
		MessageID: messageID,
		Data: map[string]any{
			"InviterName":     delivery.InviterName,
			"WorkspaceName":   delivery.WorkspaceName,
			"Role":            delivery.Role,
			"ExpiresIn":       fmt.Sprintf("%d %s", amount, unit),
			"VerificationURL": joinURL,
		},
	}); err != nil {
		return fmt.Errorf("deliver invitation email: %w", err)
	}
	return nil
}

func (s *invitationEmailSender) SendInvitationAccepted(
	ctx context.Context,
	payload events.InvitationAcceptedPayload,
	idempotencyKey string,
) error {
	messageID, err := invitationMessageID(idempotencyKey)
	if err != nil {
		return err
	}
	if err := s.mailer.SendTemplated(ctx, mailer.TemplatedEmail{
		To:        []string{payload.InviterEmail},
		Template:  "invites/acceptance",
		Subject:   fmt.Sprintf("Great news! %s has joined %s.", payload.InviteeName, payload.WorkspaceName),
		MessageID: messageID,
		Data: map[string]any{
			"InviterName":   payload.InviterName,
			"InviteeName":   payload.InviteeName,
			"WorkspaceName": payload.WorkspaceName,
			"WorkspaceURL":  fmt.Sprintf("https://%s.fortyone.app", payload.WorkspaceSlug),
			"Role":          payload.Role,
			"LoginURL":      s.websiteURL + "/login",
		},
	}); err != nil {
		return fmt.Errorf("deliver invitation acceptance email: %w", err)
	}
	return nil
}

func invitationMessageID(idempotencyKey string) (string, error) {
	idempotencyKey = strings.TrimSpace(idempotencyKey)
	if idempotencyKey == "" {
		return "", errors.New("invitation notification idempotency key is required")
	}
	digest := sha256.Sum256([]byte(idempotencyKey))
	return fmt.Sprintf("<invitation-%x@fortyone.app>", digest), nil
}
