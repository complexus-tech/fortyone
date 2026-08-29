package workerbootstrap

import (
	"context"
	"testing"
	"time"

	invitations "github.com/complexus-tech/projects-api/internal/modules/invitations/service"
	"github.com/complexus-tech/projects-api/pkg/events"
	"github.com/complexus-tech/projects-api/pkg/mailer"
	"github.com/stretchr/testify/require"
)

type invitationMailerSpy struct {
	templated []mailer.TemplatedEmail
	err       error
}

func (s *invitationMailerSpy) Send(context.Context, mailer.Email) error { return s.err }

func (s *invitationMailerSpy) SendTemplated(_ context.Context, email mailer.TemplatedEmail) error {
	s.templated = append(s.templated, email)
	return s.err
}

func TestInvitationEmailSenderBuildsOneTimeActionLink(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.August, 28, 9, 0, 0, 0, time.UTC)
	mailerSpy := &invitationMailerSpy{}
	sender, err := newInvitationEmailSender(mailerSpy, "https://fortyone.app/")
	require.NoError(t, err)
	sender.now = func() time.Time { return now }

	require.NoError(t, sender.SendInvitationEmail(context.Background(), invitations.InvitationEmailDelivery{
		IdempotencyKey: "invitation-email:test",
		InviterName:    "Ada",
		Email:          "invitee@example.com",
		Token:          "example",
		Role:           invitations.InvitationRoleMember,
		ExpiresAt:      now.Add(25 * time.Hour),
		WorkspaceName:  "Compiler Team",
	}))

	require.Len(t, mailerSpy.templated, 1)
	email := mailerSpy.templated[0]
	require.Equal(t, []string{"invitee@example.com"}, email.To)
	require.Equal(t, "invites/invitation", email.Template)
	require.NotEmpty(t, email.MessageID)
	data, ok := email.Data.(map[string]any)
	require.True(t, ok)
	require.Equal(t, "2 days", data["ExpiresIn"])
	require.Equal(t, "https://fortyone.app/onboarding/join?token=example", data["VerificationURL"])
}

func TestInvitationEmailSenderRejectsMissingOrExpiredBearer(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	sender, err := newInvitationEmailSender(&invitationMailerSpy{}, "https://fortyone.app")
	require.NoError(t, err)
	sender.now = func() time.Time { return now }

	err = sender.SendInvitationEmail(context.Background(), invitations.InvitationEmailDelivery{IdempotencyKey: "test", ExpiresAt: now.Add(time.Hour)})
	require.ErrorContains(t, err, "token is required")

	err = sender.SendInvitationEmail(context.Background(), invitations.InvitationEmailDelivery{IdempotencyKey: "test", Token: "token", ExpiresAt: now})
	require.ErrorContains(t, err, "expired")
}

func TestInvitationEmailSenderRequiresAbsoluteWebsiteURL(t *testing.T) {
	t.Parallel()

	_, err := newInvitationEmailSender(&invitationMailerSpy{}, "/relative")
	require.ErrorContains(t, err, "absolute URL")
}

func TestInvitationAcceptedSenderUsesStableMessageIdentity(t *testing.T) {
	t.Parallel()

	mailerSpy := &invitationMailerSpy{}
	sender, err := newInvitationEmailSender(mailerSpy, "https://fortyone.app")
	require.NoError(t, err)
	payload := events.InvitationAcceptedPayload{
		InviterEmail:  "admin@example.com",
		InviterName:   "Ada",
		InviteeName:   "Linus",
		WorkspaceName: "Compiler Team",
		WorkspaceSlug: "compiler-team",
		Role:          invitations.InvitationRoleMember,
	}

	require.NoError(t, sender.SendInvitationAccepted(context.Background(), payload, "invitation-accepted:stable"))
	require.NoError(t, sender.SendInvitationAccepted(context.Background(), payload, "invitation-accepted:stable"))
	require.Len(t, mailerSpy.templated, 2)
	require.Equal(t, mailerSpy.templated[0].MessageID, mailerSpy.templated[1].MessageID)
	require.Contains(t, mailerSpy.templated[0].MessageID, "@fortyone.app>")
}
