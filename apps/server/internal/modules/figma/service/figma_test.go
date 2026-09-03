package figma

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	figmadomain "github.com/complexus-tech/projects-api/internal/modules/figma/domain"
	"github.com/complexus-tech/projects-api/internal/platform/credentialvault"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type linkRepositoryStub struct {
	Repository
	connectionErr error
	upserts       []StoryLink
}

func (r *linkRepositoryStub) GetConnection(context.Context, uuid.UUID) (Connection, error) {
	return Connection{}, r.connectionErr
}

func (r *linkRepositoryStub) UpsertStoryLink(_ context.Context, link StoryLink) (StoryLink, error) {
	r.upserts = append(r.upserts, link)
	link.ID = uuid.New()
	storyLinkID := uuid.New()
	link.StoryLinkID = &storyLinkID
	return link, nil
}

type figmaStoryServiceStub struct {
	StoryService
	story      Story
	activities []StoryActivity
}

func (s *figmaStoryServiceStub) Get(context.Context, uuid.UUID, uuid.UUID) (Story, error) {
	return s.story, nil
}

func (s *figmaStoryServiceStub) RecordActivity(_ context.Context, activity StoryActivity) error {
	s.activities = append(s.activities, activity)
	return nil
}

func TestLinkStoryPersistsDegradedLinkWithoutFigmaConnection(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.September, 3, 8, 0, 0, 0, time.UTC)
	workspaceID, actorID, storyID := uuid.New(), uuid.New(), uuid.New()
	repo := &linkRepositoryStub{connectionErr: figmadomain.ErrNotFound}
	stories := &figmaStoryServiceStub{story: Story{
		ID: storyID, SequenceID: 727, TeamCode: "PRD", Title: "Checkout flow",
	}}
	service := &Service{
		repo: repo, stories: stories, now: func() time.Time { return now },
	}
	rawURL := "https://www.figma.com/design/file-key/checkout?node-id=12-34"

	link, err := service.LinkStory(
		context.Background(), workspaceID, actorID, storyID, "acme", rawURL,
	)

	require.NoError(t, err)
	require.Len(t, repo.upserts, 1)
	require.Equal(t, workspaceID, link.WorkspaceID)
	require.Equal(t, storyID, link.StoryID)
	require.Equal(t, actorID, link.CreatedByUserID)
	require.Equal(t, "file-key", link.Artifact.FileKey)
	require.Equal(t, "Figma design", link.Artifact.FileName)
	require.Equal(t, "https://www.figma.com/design/file-key?node-id=12-34", link.Artifact.CanonicalURL)
	require.Equal(t, &now, link.UnavailableAt)
	require.NotNil(t, link.StoryLinkID)
	require.Len(t, stories.activities, 1)
}

func TestLinkStoryRejectsInvalidFigmaURLWithoutPersisting(t *testing.T) {
	t.Parallel()

	repo := &linkRepositoryStub{connectionErr: errors.New("not reached")}
	storyID := uuid.New()
	service := &Service{
		repo:    repo,
		stories: &figmaStoryServiceStub{story: Story{ID: storyID}},
		now:     time.Now,
	}

	_, err := service.LinkStory(
		context.Background(), uuid.New(), uuid.New(), storyID, "acme", "https://example.com/design/file",
	)

	require.ErrorIs(t, err, ErrInvalidFigmaURL)
	require.Empty(t, repo.upserts)
}

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
	workspaceID := uuid.New()
	connectionID := uuid.New()
	generation := uuid.New()
	vault := newFigmaTestCredentialVault(t)
	service := &Service{secrets: vault, now: func() time.Time { return now }}
	tokenPayload, err := service.sealToken(workspaceID, connectionID, generation, Token{
		AccessToken: "access-token",
		ExpiresAt:   now.Add(time.Hour),
	})
	require.NoError(t, err)

	repo := &refreshRepository{
		connection: Connection{
			ID:                     connectionID,
			WorkspaceID:            workspaceID,
			CredentialPayload:      tokenPayload,
			CredentialVersion:      int16(credentialvault.CurrentVersion),
			InstallationGeneration: generation,
		},
		link: StoryLink{
			ID:           uuid.New(),
			WorkspaceID:  workspaceID,
			LastSyncedAt: now.Add(-storyLinkRefreshInterval - time.Minute),
			Artifact: Artifact{
				CanonicalURL: "https://www.figma.com/design/file-key/file-name",
				FileName:     "Cached file",
			},
		},
	}
	requestCount := 0
	service = &Service{
		client: apiClient{http: &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
			requestCount++
			return &http.Response{
				StatusCode: http.StatusTooManyRequests,
				Header:     http.Header{"Retry-After": []string{"60"}},
				Body:       io.NopCloser(strings.NewReader(`{"message":"rate limited"}`)),
			}, nil
		})}},
		secrets: vault,
		now:     func() time.Time { return now },
		repo:    repo,
	}

	link, err := service.RefreshStoryLink(context.Background(), repo.link.WorkspaceID, repo.link.ID)

	require.NoError(t, err)
	require.Equal(t, repo.link, link)
	require.Equal(t, 1, requestCount)
	require.Empty(t, repo.updates)
}

func newFigmaTestCredentialVault(t testing.TB) *credentialvault.Vault {
	t.Helper()
	key := credentialvault.Key{
		Ref:      credentialvault.KeyRef{ID: "figma-test", Version: 1},
		Material: []byte("0123456789abcdef0123456789abcdef"),
	}
	vault, err := credentialvault.New(credentialvault.Config{
		Active: key.Ref,
		Keys:   []credentialvault.Key{key},
	})
	require.NoError(t, err)
	return vault
}
