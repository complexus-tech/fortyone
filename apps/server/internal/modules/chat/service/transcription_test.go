package chatservice

import (
	"context"
	"errors"
	"testing"

	domain "github.com/complexus-tech/projects-api/internal/modules/chat/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type transcriptionRepoStub struct {
	state    domain.Transcription
	err      error
	requests int
	retry    bool
}

func (r *transcriptionRepoStub) GetTranscription(context.Context, domain.AttachmentScope, uuid.UUID) (domain.Transcription, error) {
	return r.state, r.err
}
func (r *transcriptionRepoStub) RequestTranscription(_ context.Context, _ domain.AttachmentScope, _ uuid.UUID, retry bool) (domain.Transcription, error) {
	r.requests++
	r.retry = retry
	return domain.Transcription{Status: "pending"}, nil
}

type transcriptionAccessStub struct {
	allowed bool
	calls   int
	err     error
}

func (a *transcriptionAccessStub) WorkspaceCanUseMaya(context.Context, uuid.UUID) (bool, error) {
	a.calls++
	return a.allowed, a.err
}

func TestTranscriptionRequestIsExplicitAndIdempotent(t *testing.T) {
	scope := domain.AttachmentScope{ActorScope: testScope(t), ConversationID: uuid.New()}
	for _, state := range []string{"pending", "processing", "ready", "failed"} {
		t.Run(state, func(t *testing.T) {
			repo := &transcriptionRepoStub{state: domain.Transcription{Status: state, Text: "Saved transcript"}}
			access := &transcriptionAccessStub{}
			service := NewTranscriptions(repo, access, false, nil, nil)
			result, err := service.Request(t.Context(), scope, uuid.New(), false)
			require.NoError(t, err)
			require.Equal(t, state, result.Status)
			require.Zero(t, repo.requests)
			require.Zero(t, access.calls)
		})
	}
	repo := &transcriptionRepoStub{state: domain.Transcription{Status: "not_requested"}}
	access := &transcriptionAccessStub{allowed: true}
	wakes := 0
	service := NewTranscriptions(repo, access, true, func(context.Context) error {
		wakes++
		return errors.New("redis unavailable")
	}, nil)
	result, err := service.Request(t.Context(), scope, uuid.New(), false)
	require.NoError(t, err)
	require.Equal(t, "pending", result.Status)
	require.Equal(t, 1, repo.requests)
	require.Equal(t, 1, wakes)
	repo.state.Status = "failed"
	_, err = service.Request(t.Context(), scope, uuid.New(), true)
	require.NoError(t, err)
	require.True(t, repo.retry)
	require.Equal(t, 2, repo.requests)
}

func TestTranscriptionChecksMembershipBeforeEntitlementAndLeavesStoredResultsReadable(t *testing.T) {
	scope := domain.AttachmentScope{ActorScope: testScope(t), ConversationID: uuid.New()}
	repo := &transcriptionRepoStub{state: domain.Transcription{Status: "not_requested"}, err: domain.ErrNotFound}
	access := &transcriptionAccessStub{allowed: true}
	service := NewTranscriptions(repo, access, true, nil, nil)
	_, err := service.Request(t.Context(), scope, uuid.New(), false)
	require.ErrorIs(t, err, domain.ErrNotFound)
	require.Zero(t, access.calls)
	require.Zero(t, repo.requests)
	repo.err = nil
	access.allowed = false
	_, err = service.Request(t.Context(), scope, uuid.New(), false)
	require.ErrorIs(t, err, domain.ErrTranscriptionAccess)
	require.Zero(t, repo.requests)
	repo.state = domain.Transcription{Status: "ready", Text: "A stored transcript"}
	result, err := service.Get(t.Context(), scope, uuid.New())
	require.NoError(t, err)
	require.Equal(t, "A stored transcript", result.Text)
	scope.Actor.WorkspaceID = uuid.New()
	_, err = service.Get(t.Context(), scope, uuid.New())
	require.ErrorIs(t, err, domain.ErrForbidden)
}

func TestTranscriptionUnavailableDoesNotQueueAnUnserviceableRequest(t *testing.T) {
	repo := &transcriptionRepoStub{state: domain.Transcription{Status: "not_requested"}}
	service := NewTranscriptions(repo, &transcriptionAccessStub{allowed: true}, false, nil, nil)
	_, err := service.Request(t.Context(), domain.AttachmentScope{ActorScope: testScope(t), ConversationID: uuid.New()}, uuid.New(), false)
	require.ErrorIs(t, err, domain.ErrTranscriptionUnavailable)
	require.Zero(t, repo.requests)
}
