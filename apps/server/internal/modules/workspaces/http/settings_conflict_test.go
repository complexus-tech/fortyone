package workspaceshttp

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	workspaces "github.com/complexus-tech/projects-api/internal/modules/workspaces/service"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type conflictSettingsStore struct {
	workspaces.Repository
	current   workspaces.CoreWorkspaceSettings
	submitted workspaces.CoreWorkspaceSettings
}

func (s *conflictSettingsStore) GetOrCreateWorkspaceSettings(context.Context, uuid.UUID) (workspaces.CoreWorkspaceSettings, error) {
	return s.current, nil
}
func (s *conflictSettingsStore) UpdateWorkspaceSettings(_ context.Context, _ uuid.UUID, settings workspaces.CoreWorkspaceSettings) (workspaces.CoreWorkspaceSettings, error) {
	s.submitted = settings
	return workspaces.CoreWorkspaceSettings{}, workspaces.ErrSettingsConflict
}

type settingsSession struct{ user uuid.UUID }

func (s settingsSession) Resolve(context.Context, *http.Request) (uuid.UUID, bool, error) {
	return s.user, true, nil
}

type settingsWorkspace struct{ id uuid.UUID }

func (s settingsWorkspace) ResolveCurrentWorkspace(context.Context, string, uuid.UUID) (mid.WorkspaceInfo, error) {
	return mid.WorkspaceInfo{ID: s.id, Slug: "test", UserRole: "admin"}, nil
}
func (settingsWorkspace) RecordWorkspaceAccess(context.Context, uuid.UUID, uuid.UUID) error {
	return nil
}

func TestConflictingPartialSettingsUpdateReturns409(t *testing.T) {
	id := uuid.New()
	store := &conflictSettingsStore{current: workspaces.CoreWorkspaceSettings{WorkspaceID: id, ChatEnabled: true, UpdatedAt: time.Date(2026, 10, 17, 14, 0, 0, 123000, time.UTC), WorkingDays: []int{1, 2, 3, 4, 5}, WorkingStartMinute: 540, WorkingEndMinute: 1020}}
	log := logger.NewWithText(io.Discard, slog.LevelError, "test")
	app := web.New(nil, nil)
	app.SetLogger(log)
	Routes(Config{Log: log, SecretKey: "test", BrowserSessions: settingsSession{uuid.New()}, WorkspaceResolver: settingsWorkspace{id}, Workspaces: workspaces.New(log, store, nil, workspaces.Dependencies{})}, app)
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPut, "/workspaces/test/settings", strings.NewReader(`{"objectiveEnabled":false}`))
	request.Header.Set("Content-Type", "application/json")
	app.ServeHTTP(response, request)
	require.Equal(t, http.StatusConflict, response.Code, response.Body.String())
	require.Contains(t, response.Body.String(), "refresh and try again")
	require.Equal(t, store.current.UpdatedAt, store.submitted.UpdatedAt)
	require.True(t, store.submitted.ChatEnabled, "the DB guard must reject this stale merged value")
}
