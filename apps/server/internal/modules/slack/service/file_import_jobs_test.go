package slack

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	slackdomain "github.com/complexus-tech/projects-api/internal/modules/slack/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type fileImportStoreStub struct {
	registered   slackdomain.RegisterFileImport
	importID     uuid.UUID
	record       slackdomain.FileImport
	claimed      bool
	authorized   bool
	authorizeErr error
	cancelled    bool
	failed       bool
	recoverable  []uuid.UUID
}

func (s *fileImportStoreStub) RegisterSlackFileImport(_ context.Context, input slackdomain.RegisterFileImport) (uuid.UUID, error) {
	s.registered = input
	return s.importID, nil
}
func (s *fileImportStoreStub) ClaimSlackFileImport(_ context.Context, _ uuid.UUID) (slackdomain.FileImport, bool, error) {
	return s.record, s.claimed, nil
}
func (s *fileImportStoreStub) AuthorizeSlackFileImport(_ context.Context, _ uuid.UUID) (bool, error) {
	return s.authorized, s.authorizeErr
}
func (s *fileImportStoreStub) GetSlackWorkspaceByTeamID(context.Context, string) (slackdomain.Installation, error) {
	return slackdomain.Installation{}, nil
}
func (s *fileImportStoreStub) CompleteSlackFileImport(context.Context, uuid.UUID, int32, uuid.UUID) error {
	return nil
}
func (s *fileImportStoreStub) FailSlackFileImport(context.Context, uuid.UUID, int32) error {
	s.failed = true
	return nil
}
func (s *fileImportStoreStub) CancelSlackFileImport(context.Context, uuid.UUID, int32) error {
	s.cancelled = true
	return nil
}
func (s *fileImportStoreStub) ListRecoverableSlackFileImports(context.Context, int) ([]uuid.UUID, error) {
	return s.recoverable, nil
}

type fileImportTaskStub struct {
	ids []uuid.UUID
	err error
}

func (s *fileImportTaskStub) EnqueueSlackFileImport(_ context.Context, id uuid.UUID) error {
	s.ids = append(s.ids, id)
	return s.err
}

func TestSlackFileImportDispatcherPersistsBeforeQueueing(t *testing.T) {
	importID := uuid.New()
	store := &fileImportStoreStub{importID: importID}
	tasks := &fileImportTaskStub{}
	intent := SlackFileImportIntent{
		IdempotencyKey: "confirmed:file:F123", WorkspaceID: uuid.New(),
		ActorID: uuid.New(), InstallationID: uuid.New(), InstallGeneration: uuid.New(),
		StoryID: uuid.New(), SlackTeamID: "T123", SlackUserID: "U123", FileID: "F123",
		ChannelID: "C123", ThreadTS: "1.000", MessageTS: "1.001",
	}
	require.NoError(t, NewSlackFileImportDispatcher(store, tasks).QueueSlackFileImport(context.Background(), intent))
	require.Equal(t, intent.FileID, store.registered.FileID)
	require.Equal(t, intent.MessageTS, store.registered.MessageTS)
	require.Equal(t, intent.InstallGeneration, store.registered.InstallGeneration)
	require.Equal(t, []uuid.UUID{importID}, tasks.ids)
}

func TestSlackFileImportDispatcherPreservesIntentWhenEnqueueFails(t *testing.T) {
	store := &fileImportStoreStub{importID: uuid.New()}
	tasks := &fileImportTaskStub{err: errors.New("queue unavailable")}
	intent := SlackFileImportIntent{
		IdempotencyKey: "modal:file:F123", WorkspaceID: uuid.New(),
		ActorID: uuid.New(), InstallationID: uuid.New(), InstallGeneration: uuid.New(),
		StoryID: uuid.New(), SlackTeamID: "T123", SlackUserID: "U123", FileID: "F123",
	}
	err := NewSlackFileImportDispatcher(store, tasks).QueueSlackFileImport(context.Background(), intent)
	require.ErrorContains(t, err, "enqueue Slack file import")
	require.Equal(t, intent.FileID, store.registered.FileID)
	require.Equal(t, []uuid.UUID{store.importID}, tasks.ids)
}

func TestSlackFileImportProcessorCancelsWhenAccessChanges(t *testing.T) {
	importID := uuid.New()
	store := &fileImportStoreStub{
		claimed: true, authorized: false,
		record: slackdomain.FileImport{ID: importID, AttemptCount: 2},
	}
	processor := &SlackFileImportProcessor{store: store}
	require.NoError(t, processor.ProcessSlackFileImport(context.Background(), importID))
	require.True(t, store.cancelled)
	require.False(t, store.failed)
}

func TestSlackFileImportRecoveryEnqueuesPersistedRows(t *testing.T) {
	ids := []uuid.UUID{uuid.New(), uuid.New()}
	store := &fileImportStoreStub{recoverable: ids}
	tasks := &fileImportTaskStub{}
	processor := &SlackFileImportProcessor{store: store, tasks: tasks}
	count, err := processor.RecoverSlackFileImports(context.Background())
	require.NoError(t, err)
	require.Equal(t, len(ids), count)
	require.Equal(t, ids, tasks.ids)
}

func TestSlackFileImportTerminalFailureNotifiesSubmittingUser(t *testing.T) {
	called := false
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		called = true
		require.Equal(t, "/chat.postMessage", request.URL.Path)
		require.Equal(t, "Bearer xoxb-test", request.Header.Get("Authorization"))
		var payload struct {
			Channel string `json:"channel"`
			Text    string `json:"text"`
		}
		require.NoError(t, json.NewDecoder(request.Body).Decode(&payload))
		require.Equal(t, "U123", payload.Channel)
		require.Contains(t, payload.Text, "could not be attached")
		_, _ = writer.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	store := &fileImportStoreStub{}
	client := newSlackWebClient(server.Client())
	client.baseURL = server.URL
	processor := &SlackFileImportProcessor{store: store, web: client}
	record := slackdomain.FileImport{ID: uuid.New(), SlackUserID: "U123", AttemptCount: 8}
	require.Error(t, processor.fail(context.Background(), record, "xoxb-test"))
	require.True(t, store.failed)
	require.True(t, called)
}
