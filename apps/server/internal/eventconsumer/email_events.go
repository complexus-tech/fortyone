package eventconsumer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/url"
	"strings"
	"time"

	"github.com/complexus-tech/projects-api/pkg/events"
	"github.com/complexus-tech/projects-api/pkg/mailer"
)

func (c *Consumer) handleEmailVerification(ctx context.Context, event events.Event) error {
	var payload events.EmailVerificationPayload
	payloadBytes, err := json.Marshal(event.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}
	verificationURL, err := url.Parse(fmt.Sprintf("%s/verify/%s/%s", c.websiteURL, url.PathEscape(payload.Email), url.PathEscape(payload.Token)))
	if err != nil {
		return fmt.Errorf("build email verification URL: %w", err)
	}
	query := verificationURL.Query()
	if payload.IsMobile {
		query.Set("mobileApp", "true")
	}
	if payload.CallbackURL != "" {
		query.Set("callbackUrl", payload.CallbackURL)
	}
	verificationURL.RawQuery = query.Encode()

	c.log.Info(ctx, "consumer.handleEmailVerification", "email", payload.Email)

	data := map[string]any{
		"VerificationURL": verificationURL.String(),
		"ExpiresIn":       "10 minutes",
		"IsMobile":        payload.IsMobile,
		"OTP":             payload.Token,
	}

	templateName := "auth/verification"
	subject := "Your login link for FortyOne"
	if payload.IsMobile {
		templateName = "auth/verification_mobile"
		subject = "Your verification code for FortyOne"
	}

	if err := c.mailerService.SendTemplated(ctx, mailer.TemplatedEmail{
		To:       []string{payload.Email},
		Template: templateName,
		Subject:  subject,
		Data:     data,
	}); err != nil {
		c.log.Error(ctx, "failed to send verification email", "error", err, "email", payload.Email)
		return fmt.Errorf("failed to send verification email: %w", err)
	}

	c.log.Info(ctx, "successfully sent verification email", "email", payload.Email)
	return nil
}

func (c *Consumer) handleFeedbackContributorVerification(ctx context.Context, event events.Event) error {
	var payload events.FeedbackContributorVerificationPayload
	payloadBytes, err := json.Marshal(event.Payload)
	if err != nil {
		return fmt.Errorf("marshal feedback contributor verification payload: %w", err)
	}
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return fmt.Errorf("unmarshal feedback contributor verification payload: %w", err)
	}
	if strings.TrimSpace(payload.Email) == "" || strings.TrimSpace(payload.VerificationURL) == "" || strings.TrimSpace(payload.Code) == "" {
		return errors.New("feedback contributor verification payload is incomplete")
	}
	portalName := strings.TrimSpace(payload.PortalName)
	if portalName == "" {
		portalName = "this feedback portal"
	}
	if err := c.mailerService.SendTemplated(ctx, mailer.TemplatedEmail{
		To:       []string{payload.Email},
		Template: "feedback/verification",
		Subject:  "Verify your email for " + portalName,
		Data: map[string]any{
			"PortalName":      portalName,
			"VerificationURL": payload.VerificationURL,
			"Code":            payload.Code,
			"ExpiresIn":       "10 minutes",
		},
	}); err != nil {
		return fmt.Errorf("send feedback contributor verification email: %w", err)
	}
	return nil
}

func (c *Consumer) handleInvitationEmail(ctx context.Context, event events.Event) error {
	var payload events.InvitationEmailPayload
	payloadBytes, err := json.Marshal(event.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	c.log.Info(ctx, "consumer.handleInvitationEmail",
		"email", payload.Email,
		"workspace_id", payload.WorkspaceID)

	// Calculate expiration duration
	expiresIn := time.Until(payload.ExpiresAt)

	var expiresInStr string
	if expiresIn.Hours() >= 24 {
		// Use Ceil to round up to the next day
		days := int(math.Ceil(expiresIn.Hours() / 24))
		expiresInStr = fmt.Sprintf("%d days", days)
	} else {
		expiresInStr = fmt.Sprintf("%d hours", int(expiresIn.Hours()))
	}

	data := map[string]any{
		"InviterName":     payload.InviterName,
		"WorkspaceName":   payload.WorkspaceName,
		"Role":            payload.Role,
		"ExpiresIn":       expiresInStr,
		"VerificationURL": fmt.Sprintf("%s/onboarding/join?token=%s", c.websiteURL, payload.Token),
	}

	subject := fmt.Sprintf("%s invited you to work with them in FortyOne", payload.InviterName)
	if err := c.mailerService.SendTemplated(ctx, mailer.TemplatedEmail{
		To:       []string{payload.Email},
		Template: "invites/invitation",
		Subject:  subject,
		Data:     data,
	}); err != nil {
		c.log.Error(ctx, "failed to send invitation email", "error", err, "email", payload.Email)
		return fmt.Errorf("failed to send invitation email: %w", err)
	}

	c.log.Info(ctx, "successfully sent invitation email", "email", payload.Email)
	return nil
}

func (c *Consumer) handleInvitationAccepted(ctx context.Context, event events.Event) error {
	var payload events.InvitationAcceptedPayload
	payloadBytes, err := json.Marshal(event.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	c.log.Info(ctx, "consumer.handleInvitationAccepted",
		"inviter_email", payload.InviterEmail,
		"invitee_email", payload.InviteeEmail,
		"workspace_id", payload.WorkspaceID)

	data := map[string]any{
		"InviterName":   payload.InviterName,
		"InviteeName":   payload.InviteeName,
		"WorkspaceName": payload.WorkspaceName,
		"WorkspaceURL":  fmt.Sprintf("https://%s.fortyone.app", payload.WorkspaceSlug),
		"Role":          payload.Role,
		"LoginURL":      fmt.Sprintf("%s/login", c.websiteURL),
	}

	subject := fmt.Sprintf("Great news! %s has joined %s.", payload.InviteeName, payload.WorkspaceName)
	if err := c.mailerService.SendTemplated(ctx, mailer.TemplatedEmail{
		To:       []string{payload.InviterEmail},
		Template: "invites/acceptance",
		Subject:  subject,
		Data:     data,
	}); err != nil {
		c.log.Error(ctx, "failed to send invitation accepted email", "error", err, "email", payload.InviterEmail)
		return fmt.Errorf("failed to send invitation accepted email: %w", err)
	}

	c.log.Info(ctx, "successfully sent invitation accepted email", "email", payload.InviterEmail)
	return nil
}
