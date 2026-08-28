package invitations

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestGetInvitationEnforcesBearerLifecycle(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.August, 28, 12, 0, 0, 0, time.UTC)
	usedAt := now.Add(-time.Minute)
	repositoryFailure := errors.New("read invitation")

	tests := []struct {
		name       string
		invitation CoreWorkspaceInvitation
		repoErr    error
		wantErr    error
	}{
		{
			name:       "active invitation",
			invitation: CoreWorkspaceInvitation{ID: uuid.New(), ExpiresAt: now.Add(time.Hour)},
		},
		{
			name:       "expiration is exclusive",
			invitation: CoreWorkspaceInvitation{ID: uuid.New(), ExpiresAt: now},
			wantErr:    ErrInvitationExpired,
		},
		{
			name:       "used invitation",
			invitation: CoreWorkspaceInvitation{ID: uuid.New(), ExpiresAt: now.Add(time.Hour), UsedAt: &usedAt},
			wantErr:    ErrInvitationUsed,
		},
		{
			name:    "repository failure",
			repoErr: repositoryFailure,
			wantErr: repositoryFailure,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			tokens := newTestInvitationTokenManager(t)
			rawToken, storedToken, err := tokens.Issue()
			if err != nil {
				t.Fatalf("issue invitation token: %v", err)
			}
			repo := &invitationQueryRepository{
				invitation: test.invitation,
				getErr:     test.repoErr,
			}
			service := New(repo, tokens, nil, nil)
			service.now = func() time.Time { return now }

			invitation, err := service.GetInvitation(context.Background(), rawToken)
			if !errors.Is(err, test.wantErr) {
				t.Fatalf("GetInvitation() error = %v, want %v", err, test.wantErr)
			}
			if repo.getCalls != 1 {
				t.Fatalf("repository calls = %d, want 1", repo.getCalls)
			}
			if repo.lookup.KeyID != storedToken.KeyID || repo.lookup.Version != storedToken.Version {
				t.Fatalf("lookup metadata = %q/%d, want %q/%d", repo.lookup.KeyID, repo.lookup.Version, storedToken.KeyID, storedToken.Version)
			}
			if test.wantErr == nil && invitation.ID != test.invitation.ID {
				t.Fatalf("invitation ID = %s, want %s", invitation.ID, test.invitation.ID)
			}
		})
	}
}

func TestGetInvitationRejectsMalformedBearerBeforePersistence(t *testing.T) {
	t.Parallel()

	repo := &invitationQueryRepository{}
	service := New(repo, newTestInvitationTokenManager(t), nil, nil)

	for _, rawToken := range []string{"", "not-an-invitation-token", "wi1.unknown.invalid.invalid"} {
		_, err := service.GetInvitation(context.Background(), rawToken)
		if !errors.Is(err, ErrInvitationNotFound) {
			t.Fatalf("GetInvitation(%q) error = %v, want ErrInvitationNotFound", rawToken, err)
		}
	}
	if repo.getCalls != 0 {
		t.Fatalf("malformed bearers reached persistence %d times", repo.getCalls)
	}
}

func TestListUserInvitationsNormalizesEmailAndUsesUTCClock(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.August, 28, 14, 30, 0, 0, time.FixedZone("test", 2*60*60))
	want := []CoreWorkspaceInvitation{{ID: uuid.New()}}
	repo := &invitationQueryRepository{listed: want}
	service := New(repo, newTestInvitationTokenManager(t), nil, nil)
	service.now = func() time.Time { return now }

	got, err := service.ListUserInvitations(context.Background(), "  Person@Example.COM  ")
	if err != nil {
		t.Fatalf("ListUserInvitations() error = %v", err)
	}
	if repo.email != "person@example.com" {
		t.Fatalf("repository email = %q, want normalized email", repo.email)
	}
	if !repo.now.Equal(now.UTC()) || repo.now.Location() != time.UTC {
		t.Fatalf("repository time = %s (%s), want %s UTC", repo.now, repo.now.Location(), now.UTC())
	}
	if len(got) != 1 || got[0].ID != want[0].ID {
		t.Fatalf("listed invitations = %#v, want %#v", got, want)
	}
}

func TestListUserInvitationsRejectsInvalidEmailBeforePersistence(t *testing.T) {
	t.Parallel()

	repo := &invitationQueryRepository{}
	service := New(repo, newTestInvitationTokenManager(t), nil, nil)

	if _, err := service.ListUserInvitations(context.Background(), "not-an-email"); err == nil {
		t.Fatal("ListUserInvitations() error = nil, want validation error")
	}
	if repo.listByEmailCalls != 0 {
		t.Fatalf("invalid email reached persistence %d times", repo.listByEmailCalls)
	}
}

type invitationQueryRepository struct {
	Repository
	invitation       CoreWorkspaceInvitation
	getErr           error
	getCalls         int
	lookup           InvitationTokenLookup
	listed           []CoreWorkspaceInvitation
	listByEmailCalls int
	email            string
	now              time.Time
}

func (r *invitationQueryRepository) GetInvitation(_ context.Context, lookup InvitationTokenLookup) (CoreWorkspaceInvitation, error) {
	r.getCalls++
	r.lookup = lookup
	return r.invitation, r.getErr
}

func (r *invitationQueryRepository) ListInvitationsByEmail(_ context.Context, email string, now time.Time) ([]CoreWorkspaceInvitation, error) {
	r.listByEmailCalls++
	r.email = email
	r.now = now
	return r.listed, nil
}
