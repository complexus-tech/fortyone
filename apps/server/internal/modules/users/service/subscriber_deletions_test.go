package users

import (
	"context"
	"errors"
	"testing"
	"time"

	usersdomain "github.com/complexus-tech/projects-api/internal/modules/users/domain"
	"github.com/complexus-tech/projects-api/pkg/brevo"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type subscriberRepositoryStub struct {
	jobs        []usersdomain.SubscriberDeletion
	events      []string
	current     bool
	completeErr error
	retryAt     time.Time
	account     usersdomain.AccountSubscriber
	found       bool
}

func (r *subscriberRepositoryStub) Claim(context.Context, time.Time, time.Time) (usersdomain.SubscriberDeletion, bool, error) {
	if len(r.jobs) == 0 {
		return usersdomain.SubscriberDeletion{}, false, nil
	}
	job := r.jobs[0]
	r.jobs = r.jobs[1:]
	return job, true, nil
}
func (r *subscriberRepositoryStub) Renew(context.Context, uuid.UUID, uuid.UUID, time.Time, time.Time) (bool, error) {
	r.events = append(r.events, "renew")
	return r.current, nil
}
func (r *subscriberRepositoryStub) Complete(ctx context.Context, _, _ uuid.UUID) error {
	r.events = append(r.events, "complete")
	if err := ctx.Err(); err != nil {
		return err
	}
	return r.completeErr
}
func (r *subscriberRepositoryStub) Retry(ctx context.Context, _, _ uuid.UUID, retryAt, _ time.Time) error {
	r.events = append(r.events, "retry")
	r.retryAt = retryAt
	return ctx.Err()
}
func (r *subscriberRepositoryStub) WithinSubscriberLifecycle(ctx context.Context, _ string, callback func(context.Context) error) error {
	r.events = append(r.events, "locked")
	defer func() { r.events = append(r.events, "unlocked") }()
	return callback(ctx)
}
func (r *subscriberRepositoryStub) ActiveAccount(context.Context, string) (usersdomain.AccountSubscriber, bool, error) {
	return r.account, r.found, nil
}

type subscriberProviderStub struct {
	repository *subscriberRepositoryStub
	deleteErr  error
	onDelete   func()
	updates    []brevo.CreateOrUpdateContactRequest
}

func (p *subscriberProviderStub) DeleteAccountContact(context.Context, string) error {
	p.repository.events = append(p.repository.events, "delete")
	if p.onDelete != nil {
		p.onDelete()
	}
	return p.deleteErr
}
func (p *subscriberProviderStub) CreateOrUpdateContact(_ context.Context, request brevo.CreateOrUpdateContactRequest) (*brevo.CreateOrUpdateContactResponse, error) {
	p.updates = append(p.updates, request)
	return &brevo.CreateOrUpdateContactResponse{}, nil
}

func subscriberCleanupFixture() (*SubscriberCleanup, *subscriberRepositoryStub, *subscriberProviderStub) {
	repo := &subscriberRepositoryStub{current: true, jobs: []usersdomain.SubscriberDeletion{{ID: uuid.New(), LeaseToken: uuid.New(), Email: "member@example.com", AttemptCount: 1}}}
	provider := &subscriberProviderStub{repository: repo}
	cleanup := NewSubscriberCleanup(repo, provider)
	cleanup.now = func() time.Time { return time.Date(2026, 9, 16, 10, 0, 0, 0, time.UTC) }
	return cleanup, repo, provider
}

func TestSubscriberCleanupPurgesOnlyAfterProviderConfirmation(t *testing.T) {
	cleanup, repo, _ := subscriberCleanupFixture()
	completed, err := cleanup.Dispatch(context.Background())
	require.NoError(t, err)
	require.Equal(t, 1, completed)
	require.Equal(t, []string{"locked", "renew", "delete", "unlocked", "complete"}, repo.events)
}

func TestSubscriberCleanupRetainsFailuresWithPrivateBoundedBackoff(t *testing.T) {
	cleanup, repo, provider := subscriberCleanupFixture()
	provider.deleteErr = errors.New("provider rejected member@example.com")
	completed, err := cleanup.Dispatch(context.Background())
	require.Zero(t, completed)
	require.Error(t, err)
	require.NotContains(t, err.Error(), "member@example.com")
	require.NotContains(t, repo.events, "complete")
	require.Contains(t, repo.events, "retry")
	require.Equal(t, cleanup.now().Add(time.Minute), repo.retryAt)
	require.Equal(t, time.Hour, subscriberRetryDelay(100))
}

func TestSubscriberCleanupExpiredClaimNeverTouchesProvider(t *testing.T) {
	cleanup, repo, _ := subscriberCleanupFixture()
	repo.current = false
	completed, err := cleanup.Dispatch(context.Background())
	require.Zero(t, completed)
	require.Error(t, err)
	require.NotContains(t, repo.events, "delete")
	require.NotContains(t, repo.events, "complete")
}

func TestSubscriberCleanupPersistenceFailureIsRetriable(t *testing.T) {
	cleanup, repo, _ := subscriberCleanupFixture()
	repo.completeErr = errors.New("database unavailable")
	completed, err := cleanup.Dispatch(context.Background())
	require.Zero(t, completed)
	require.ErrorIs(t, err, repo.completeErr)
}

func TestSubscriberCleanupFinalizesConfirmedDeleteAfterCancellation(t *testing.T) {
	cleanup, repo, provider := subscriberCleanupFixture()
	ctx, cancel := context.WithCancel(context.Background())
	provider.onDelete = cancel
	completed, err := cleanup.Dispatch(ctx)
	require.Equal(t, 1, completed)
	require.ErrorIs(t, err, context.Canceled)
	require.Contains(t, repo.events, "complete")
}

func TestSubscriberUpdateCannotReplayDeletedOrPendingAccount(t *testing.T) {
	cleanup, repo, provider := subscriberCleanupFixture()
	update := usersdomain.SubscriberUpdate{Email: "old@example.com"}
	require.NoError(t, cleanup.UpdateActiveAccount(context.Background(), update))
	require.Empty(t, provider.updates)
	repo.found = true
	repo.account = usersdomain.AccountSubscriber{Email: "current@example.com", FullName: "Current profile", CleanupPending: true}
	require.Error(t, cleanup.UpdateActiveAccount(context.Background(), update))
	require.Empty(t, provider.updates)
	repo.account.CleanupPending = false
	require.NoError(t, cleanup.UpdateActiveAccount(context.Background(), update))
	require.Equal(t, []brevo.CreateOrUpdateContactRequest{{Email: "current@example.com", Attributes: brevo.ContactAttributes{"NAME": "Current profile"}}}, provider.updates)
}

func TestSubscriberOnboardingCannotApplyOldIdentityToReusedEmail(t *testing.T) {
	cleanup, repo, provider := subscriberCleanupFixture()
	repo.found = true
	repo.account = usersdomain.AccountSubscriber{UserID: uuid.New(), Email: "member@example.com", FullName: "Current"}
	update := usersdomain.SubscriberUpdate{UserID: uuid.New(), Email: "member@example.com", ListIDs: []int64{6}}
	require.NoError(t, cleanup.UpdateActiveAccount(t.Context(), update))
	require.Empty(t, provider.updates)
	update.UserID = repo.account.UserID
	update.Attributes = map[string]string{"WORKSPACE_NAME": "Acme", "WORKSPACE_SLUG": "acme", "NAME": "Old name"}
	require.NoError(t, cleanup.UpdateActiveAccount(t.Context(), update))
	require.Len(t, provider.updates, 1)
	require.Equal(t, []int64{6}, provider.updates[0].ListIDs)
	require.Equal(t, "Current", provider.updates[0].Attributes["NAME"])
	require.Equal(t, "acme", provider.updates[0].Attributes["WORKSPACE_SLUG"])
}
