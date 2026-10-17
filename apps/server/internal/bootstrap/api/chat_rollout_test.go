package api

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/complexus-tech/projects-api/internal/bootstrap/agentadapter"
	agenthttp "github.com/complexus-tech/projects-api/internal/modules/agentinvocations/http"
	agentruns "github.com/complexus-tech/projects-api/internal/modules/agentinvocations/service"
	chathttp "github.com/complexus-tech/projects-api/internal/modules/chat/http"
	"github.com/complexus-tech/projects-api/internal/platform/chatrollout"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/internal/platform/http/mux"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type rolloutSession struct{ user uuid.UUID }

func (s rolloutSession) Resolve(_ context.Context, r *http.Request) (uuid.UUID, bool, error) {
	return s.user, r.Header.Get("X-Test-Session") == "present", nil
}

type rolloutWorkspace struct {
	id      uuid.UUID
	enabled bool
}

func (s rolloutWorkspace) ResolveCurrentWorkspace(_ context.Context, slug string, _ uuid.UUID) (mid.WorkspaceInfo, error) {
	if slug != "pilot" {
		return mid.WorkspaceInfo{}, mid.ErrWorkspaceAccessDenied
	}
	return mid.WorkspaceInfo{ID: s.id, Slug: slug, UserRole: "member", ChatEnabled: s.enabled}, nil
}
func (rolloutWorkspace) RecordWorkspaceAccess(context.Context, uuid.UUID, uuid.UUID) error {
	return nil
}

func rolloutHTTPApp(t *testing.T, enabled bool, workspace uuid.UUID) *web.App {
	t.Helper()
	log := logger.NewWithText(io.Discard, slog.LevelError, "rollout-test")
	app := web.New(nil, nil)
	app.SetLogger(log)
	sessions := rolloutSession{uuid.New()}
	resolver := rolloutWorkspace{workspace, enabled}
	policy := chatrollout.New(chatrollout.ReaderFunc(func(context.Context, uuid.UUID) (bool, error) { return enabled, nil }))
	chathttp.Routes(chathttp.Config{Log: log, SecretKey: "test", BrowserSessions: sessions, WorkspaceResolver: resolver}, app)
	agenthttp.Routes(agenthttp.Config{Log: log, SecretKey: "test", BrowserSessions: sessions, WorkspaceResolver: resolver,
		Runs: agentruns.New(nil, nil), ChatAccess: agentadapter.ChatAccess{Rollout: policy}}, app)
	cfg := mux.Config{Log: log, SecretKey: "test"}
	registerChatAttachments(cfg, resolver, sessions, services{}, app)
	registerChatAttachmentThumbnails(cfg, resolver, sessions, services{}, app)
	registerChatTranscriptions(cfg, resolver, sessions, services{}, app)
	return app
}

func TestCapabilitiesAreAuthenticatedAndDoNotEnableOtherWorkspaces(t *testing.T) {
	workspace := uuid.New()
	for _, enabled := range []bool{false, true} {
		app := rolloutHTTPApp(t, enabled, workspace)
		response := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, "/workspaces/pilot/chat/capabilities", nil)
		app.ServeHTTP(response, request)
		require.Equal(t, http.StatusUnauthorized, response.Code)
		response = httptest.NewRecorder()
		request.Header.Set("X-Test-Session", "present")
		app.ServeHTTP(response, request)
		require.Equal(t, http.StatusOK, response.Code, response.Body.String())
		var body struct {
			Data struct {
				Enabled            bool `json:"enabled"`
				AttachmentsEnabled bool `json:"attachmentsEnabled"`
			} `json:"data"`
		}
		require.NoError(t, json.Unmarshal(response.Body.Bytes(), &body))
		require.Equal(t, enabled, body.Data.Enabled)
		require.Equal(t, enabled, body.Data.AttachmentsEnabled)
		require.Equal(t, "private, no-store", response.Header().Get("Cache-Control"))
		response = httptest.NewRecorder()
		request = httptest.NewRequest(http.MethodGet, "/workspaces/other/chat/capabilities", nil)
		request.Header.Set("X-Test-Session", "present")
		app.ServeHTTP(response, request)
		require.Equal(t, http.StatusNotFound, response.Code)
	}
}

func TestDisabledChatRoutesRejectBeforeStorageOrUploads(t *testing.T) {
	workspace := uuid.New()
	app := rolloutHTTPApp(t, false, workspace)
	id := uuid.NewString()
	base := "/workspaces/pilot/chat/"
	routes := [][2]string{
		{"GET", "conversations"}, {"POST", "conversations/direct"},
		{"GET", "conversations/" + id}, {"GET", "conversations/" + id + "/members"},
		{"GET", "conversations/" + id + "/messages"}, {"POST", "conversations/" + id + "/messages"},
		{"GET", "messages/" + id}, {"GET", "messages/" + id + "/replies"}, {"PATCH", "messages/" + id}, {"DELETE", "messages/" + id},
		{"PUT", "messages/" + id + "/reactions/eyes"}, {"DELETE", "messages/" + id + "/reactions/eyes"},
		{"POST", "conversations/" + id + "/read"}, {"POST", "conversations/" + id + "/typing"},
		{"GET", "conversations/" + id + "/search"}, {"GET", "conversations/" + id + "/resources"},
		{"POST", "conversations/" + id + "/attachments"}, {"GET", "conversations/" + id + "/attachments/" + id}, {"DELETE", "conversations/" + id + "/attachments/" + id},
		{"GET", "conversations/" + id + "/attachments/" + id + "/thumbnail"},
		{"GET", "conversations/" + id + "/attachments/" + id + "/transcription"}, {"POST", "conversations/" + id + "/attachments/" + id + "/transcription"},
	}
	for _, route := range routes {
		t.Run(route[0]+" "+route[1], func(t *testing.T) {
			response := httptest.NewRecorder()
			request := httptest.NewRequest(route[0], base+route[1], nil)
			request.Header.Set("X-Test-Session", "present")
			app.ServeHTTP(response, request)
			require.Equal(t, http.StatusForbidden, response.Code, response.Body.String())
			require.Contains(t, response.Body.String(), "Messages is not enabled")
		})
	}
}

func TestDisabledChatRunActionsCannotReachInvocationStorage(t *testing.T) {
	app := rolloutHTTPApp(t, false, uuid.New())
	base := "/workspaces/pilot/chat/conversations/" + uuid.NewString() + "/maya-runs"
	for _, action := range []string{"list", "approve", "reject", "retry", "cancel"} {
		method, path := http.MethodPost, base+"/"+uuid.NewString()+"/"+action
		if action == "list" {
			method, path = http.MethodGet, base
		}
		request := httptest.NewRequest(method, path, nil)
		request.Header.Set("X-Test-Session", "present")
		response := httptest.NewRecorder()
		app.ServeHTTP(response, request)
		require.Equal(t, http.StatusNotFound, response.Code, action+": "+response.Body.String())
	}
}
