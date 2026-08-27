package storieshttp

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	users "github.com/complexus-tech/projects-api/internal/modules/users/service"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/google/uuid"
)

type failingStoryUserReader struct {
	err error
}

func (r failingStoryUserReader) GetUsersByIDs(context.Context, []uuid.UUID) ([]users.CoreUser, error) {
	return nil, r.err
}

func TestRespondCreatedStoryFallsBackWhenUserEnrichmentFails(t *testing.T) {
	storyID := uuid.New()
	workspaceID := uuid.New()
	reporterID := uuid.New()
	var logs bytes.Buffer
	handler := &Handlers{
		users: failingStoryUserReader{err: errors.New("sensitive upstream details")},
		log:   logger.NewWithText(&logs, slog.LevelError, "test"),
	}
	recorder := httptest.NewRecorder()

	err := handler.respondCreatedStory(context.Background(), recorder, stories.CoreSingleStory{
		ID:        storyID,
		Workspace: workspaceID,
		Reporter:  &reporterID,
		Title:     "Persisted story",
	})
	if err != nil {
		t.Fatalf("respondCreatedStory() error = %v", err)
	}
	if recorder.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusCreated)
	}

	var response struct {
		Data AppSingleStory `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.Data.ID != storyID || response.Data.Title != "Persisted story" {
		t.Fatalf("created story response = %#v", response.Data)
	}
	if response.Data.Reporter != nil {
		t.Fatalf("fallback response unexpectedly enriched reporter: %#v", response.Data.Reporter)
	}
	if strings.Contains(logs.String(), "sensitive upstream details") {
		t.Fatalf("response fallback logged sensitive enrichment details: %s", logs.String())
	}
	if !strings.Contains(logs.String(), "created story response user enrichment failed") {
		t.Fatalf("response fallback did not emit a safe diagnostic: %s", logs.String())
	}
}

type writeFailingResponseWriter struct {
	header   http.Header
	writeErr error
}

func (w *writeFailingResponseWriter) Header() http.Header {
	return w.header
}

func (*writeFailingResponseWriter) WriteHeader(int) {}

func (w *writeFailingResponseWriter) Write([]byte) (int, error) {
	return 0, w.writeErr
}

func TestRespondCreatedStoryReturnsResponseWriteFailure(t *testing.T) {
	wantErr := errors.New("connection closed")
	writer := &writeFailingResponseWriter{
		header:   make(http.Header),
		writeErr: wantErr,
	}

	err := (&Handlers{}).respondCreatedStory(context.Background(), writer, stories.CoreSingleStory{
		ID:        uuid.New(),
		Workspace: uuid.New(),
		Title:     "Persisted story",
	})
	if !errors.Is(err, wantErr) {
		t.Fatalf("respondCreatedStory() error = %v, want %v", err, wantErr)
	}
}
