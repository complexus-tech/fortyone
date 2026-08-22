package figma

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type refreshRepository struct {
	Repository
	connection Connection
	link       StoryLink
	getCalls   int
	updates    []StoryLink
}

func (r *refreshRepository) GetStoryLink(context.Context, uuid.UUID, uuid.UUID) (StoryLink, error) {
	r.getCalls++
	return r.link, nil
}

func (r *refreshRepository) GetConnection(context.Context, uuid.UUID) (Connection, error) {
	return r.connection, nil
}

func (r *refreshRepository) UpdateStoryLink(_ context.Context, link StoryLink) error {
	r.updates = append(r.updates, link)
	return nil
}

func TestRefreshStoryLinkUsesRecentlySyncedArtifact(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.August, 22, 1, 0, 0, 0, time.UTC)
	repo := &refreshRepository{link: StoryLink{
		ID:           uuid.New(),
		WorkspaceID:  uuid.New(),
		LastSyncedAt: now.Add(-time.Minute),
		Artifact: Artifact{
			CanonicalURL: "https://www.figma.com/design/file-key/file-name",
			FileName:     "Cached file",
		},
	}}
	service := &Service{repo: repo, now: func() time.Time { return now }}

	link, err := service.RefreshStoryLink(context.Background(), repo.link.WorkspaceID, repo.link.ID)

	require.NoError(t, err)
	require.Equal(t, repo.link, link)
	require.Equal(t, 1, repo.getCalls)
	require.Empty(t, repo.updates)
}

func TestRefreshStoryLinkFallsBackToCachedArtifactWhenFigmaIsRateLimited(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.August, 22, 1, 0, 0, 0, time.UTC)
	const secret = "test-secret"
	tokenPayload, err := encryptToken(secret, Token{
		AccessToken: "access-token",
		ExpiresAt:   now.Add(time.Hour),
	})
	require.NoError(t, err)

	repo := &refreshRepository{
		connection: Connection{TokenPayload: tokenPayload},
		link: StoryLink{
			ID:           uuid.New(),
			WorkspaceID:  uuid.New(),
			LastSyncedAt: now.Add(-storyLinkRefreshInterval - time.Minute),
			Artifact: Artifact{
				CanonicalURL: "https://www.figma.com/design/file-key/file-name",
				FileName:     "Cached file",
			},
		},
	}
	requestCount := 0
	service := &Service{
		client: apiClient{http: &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
			requestCount++
			return &http.Response{
				StatusCode: http.StatusTooManyRequests,
				Header:     http.Header{"Retry-After": []string{"60"}},
				Body:       io.NopCloser(strings.NewReader(`{"message":"rate limited"}`)),
			}, nil
		})}},
		config: Config{SecretKey: secret},
		now:    func() time.Time { return now },
		repo:   repo,
	}

	link, err := service.RefreshStoryLink(context.Background(), repo.link.WorkspaceID, repo.link.ID)

	require.NoError(t, err)
	require.Equal(t, repo.link, link)
	require.Equal(t, 1, requestCount)
	require.Empty(t, repo.updates)
}
