package taskhandlers

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"testing"

	github "github.com/complexus-tech/projects-api/internal/modules/github/service"
	githubshared "github.com/complexus-tech/projects-api/internal/modules/github/shared"
	storydomain "github.com/complexus-tech/projects-api/internal/modules/stories/domain"
	"github.com/complexus-tech/projects-api/pkg/tasks"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/require"
)

type githubStorySyncReaderStub struct {
	story storydomain.Story
	err   error
}

func (reader *githubStorySyncReaderStub) GetStoryForMutation(
	context.Context,
	storydomain.MutationScope,
	uuid.UUID,
) (storydomain.Story, error) {
	return reader.story, reader.err
}

type githubStorySyncRepositorySpy struct {
	github.Repository
	syncLookups int
}

func (repo *githubStorySyncRepositorySpy) FindBidirectionalIssueSyncLinkByTeamID(
	context.Context,
	uuid.UUID,
	uuid.UUID,
) (githubshared.BidirectionalIssueSyncLink, error) {
	repo.syncLookups++
	// Stop before any network request, while proving the sync service was called.
	return githubshared.BidirectionalIssueSyncLink{}, sql.ErrNoRows
}

func TestHandleGitHubStorySyncSourceLifecycle(t *testing.T) {
	t.Parallel()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	privateKeyBase64 := base64.StdEncoding.EncodeToString(pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}))
	readErr := errors.New("database temporarily unavailable")

	for _, test := range []struct {
		name        string
		readErr     error
		wantErr     error
		syncLookups int
	}{
		{name: "deleted story", readErr: storydomain.ErrNotFound},
		{name: "wrapped deleted story", readErr: fmt.Errorf("load story: %w", storydomain.ErrNotFound)},
		{name: "transient read failure", readErr: readErr, wantErr: readErr},
		{name: "existing story", syncLookups: 1},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			log := testTaskLogger()
			repo := &githubStorySyncRepositorySpy{}
			service, err := github.New(log, repo, nil, nil, nil, github.Config{
				AppID:            1,
				PrivateKeyBase64: privateKeyBase64,
			})
			require.NoError(t, err)
			storyID, workspaceID := uuid.New(), uuid.New()
			handler := &handlers{
				log:           log,
				githubService: service,
				systemUserID:  uuid.New(),
				storySyncReader: &githubStorySyncReaderStub{
					story: storydomain.Story{ID: storyID, Team: uuid.New()},
					err:   test.readErr,
				},
			}
			payload, err := json.Marshal(tasks.GitHubStorySyncPayload{
				StoryID: storyID, WorkspaceID: workspaceID,
			})
			require.NoError(t, err)

			err = handler.HandleGitHubStorySync(t.Context(), asynq.NewTask(tasks.TypeGitHubStorySync, payload))
			if test.wantErr != nil {
				require.ErrorIs(t, err, test.wantErr)
				require.NotErrorIs(t, err, asynq.SkipRetry)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, test.syncLookups, repo.syncLookups)
		})
	}
}
