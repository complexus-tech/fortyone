package figma

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/complexus-tech/projects-api/internal/platform/oauthstate"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type figmaOAuthStateRepositoryStub struct {
	Repository
	mu      sync.Mutex
	records map[string]OAuthState
}

func newFigmaOAuthStateRepositoryStub() *figmaOAuthStateRepositoryStub {
	return &figmaOAuthStateRepositoryStub{records: make(map[string]OAuthState)}
}

func (r *figmaOAuthStateRepositoryStub) SaveOAuthState(_ context.Context, state OAuthState) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.records[state.StateHash] = state
	return nil
}

func (r *figmaOAuthStateRepositoryStub) ConsumeOAuthState(_ context.Context, stateHash string, now time.Time) (OAuthState, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	state, ok := r.records[stateHash]
	if !ok || !now.Before(state.ExpiresAt) {
		return OAuthState{}, sql.ErrNoRows
	}
	delete(r.records, stateHash)
	return state, nil
}

func (r *figmaOAuthStateRepositoryStub) snapshot(stateHash string) (OAuthState, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	state, ok := r.records[stateHash]
	return state, ok
}

func TestFigmaOAuthStateIsOpaqueShortLivedAndIdentityBound(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.August, 28, 12, 0, 0, 0, time.UTC)
	repo := newFigmaOAuthStateRepositoryStub()
	service := &Service{repo: repo, now: func() time.Time { return now }}
	workspaceID, userID := uuid.New(), uuid.New()

	state, verifier, err := service.createOAuthState(context.Background(), workspaceID, userID, "acme")
	require.NoError(t, err)
	_, err = oauthstate.Parse(state)
	require.NoError(t, err)
	require.NotContains(t, state, workspaceID.String())
	require.NotContains(t, state, userID.String())
	require.NotEmpty(t, verifier)

	stored, ok := repo.snapshot(digest(state))
	require.True(t, ok)
	require.NotEqual(t, state, stored.StateHash)
	require.Equal(t, workspaceID, stored.WorkspaceID)
	require.Equal(t, userID, stored.UserID)
	require.Equal(t, "acme", stored.WorkspaceSlug)
	require.Equal(t, verifier, stored.CodeVerifier)
	require.Equal(t, now.Add(figmaOAuthStateTTL), stored.ExpiresAt)
}

func TestFigmaOAuthStateRejectsReplayExpiryAndInvalidBinding(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.August, 28, 12, 0, 0, 0, time.UTC)
	repo := newFigmaOAuthStateRepositoryStub()
	service := &Service{repo: repo, now: func() time.Time { return now }}
	ctx := context.Background()

	replayState, _, err := service.createOAuthState(ctx, uuid.New(), uuid.New(), "acme")
	require.NoError(t, err)
	_, err = service.consumeOAuthState(ctx, replayState)
	require.NoError(t, err)
	_, err = service.consumeOAuthState(ctx, replayState)
	require.ErrorIs(t, err, ErrFigmaOAuthStateInvalid)

	expiredState, _, err := service.createOAuthState(ctx, uuid.New(), uuid.New(), "acme")
	require.NoError(t, err)
	service.now = func() time.Time { return now.Add(figmaOAuthStateTTL) }
	_, err = service.consumeOAuthState(ctx, expiredState)
	require.ErrorIs(t, err, ErrFigmaOAuthStateInvalid)

	service.now = func() time.Time { return now }
	invalidBindingState, _, err := service.createOAuthState(ctx, uuid.New(), uuid.New(), "acme")
	require.NoError(t, err)
	repo.mu.Lock()
	corrupt := repo.records[digest(invalidBindingState)]
	corrupt.UserID = uuid.Nil
	repo.records[digest(invalidBindingState)] = corrupt
	repo.mu.Unlock()
	_, err = service.consumeOAuthState(ctx, invalidBindingState)
	require.ErrorIs(t, err, ErrFigmaOAuthStateBinding)
}

func TestFigmaOAuthStateHasExactlyOneConcurrentConsumer(t *testing.T) {
	t.Parallel()

	repo := newFigmaOAuthStateRepositoryStub()
	service := &Service{repo: repo, now: time.Now}
	state, _, err := service.createOAuthState(context.Background(), uuid.New(), uuid.New(), "acme")
	require.NoError(t, err)

	const consumers = 32
	var successes atomic.Int32
	var wait sync.WaitGroup
	wait.Add(consumers)
	for range consumers {
		go func() {
			defer wait.Done()
			if _, consumeErr := service.consumeOAuthState(context.Background(), state); consumeErr == nil {
				successes.Add(1)
			} else if !errors.Is(consumeErr, ErrFigmaOAuthStateInvalid) {
				t.Errorf("consumeOAuthState() error = %v", consumeErr)
			}
		}()
	}
	wait.Wait()
	require.EqualValues(t, 1, successes.Load())
}

func TestFigmaOAuthStateErrorsNeverEchoRawState(t *testing.T) {
	t.Parallel()

	service := &Service{repo: newFigmaOAuthStateRepositoryStub(), now: time.Now}
	rawState := strings.Repeat("sensitive-state", 4)
	_, err := service.consumeOAuthState(context.Background(), rawState)
	require.Error(t, err)
	require.NotContains(t, err.Error(), rawState)
}
