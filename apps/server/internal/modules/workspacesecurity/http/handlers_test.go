package workspacesecurityhttp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	domain "github.com/complexus-tech/projects-api/internal/modules/workspacesecurity/domain"
	"github.com/complexus-tech/projects-api/internal/platform/auth"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

type testService struct {
	Service
	events         []domain.AuditEvent
	calls, exports int
}

func (s *testService) Audit(context.Context, domain.Scope, domain.AuditFilter) ([]domain.AuditEvent, error) {
	s.calls++
	return s.events, nil
}
func (s *testService) RecordAuditExport(context.Context, domain.Scope, int) error {
	s.exports++
	return nil
}

type resolver struct{ workspaceID uuid.UUID }

func (r resolver) ResolveCurrentWorkspace(context.Context, string, uuid.UUID) (mid.WorkspaceInfo, error) {
	return mid.WorkspaceInfo{ID: r.workspaceID, Slug: "test", UserRole: "admin"}, nil
}
func (r resolver) RecordWorkspaceAccess(context.Context, uuid.UUID, uuid.UUID) error { return nil }
func request(t *testing.T, handler web.Handler, scope domain.Scope, path string) *httptest.ResponseRecorder {
	t.Helper()
	ctx := auth.SetUserID(t.Context(), scope.ActorID)
	req := httptest.NewRequest(http.MethodGet, path, nil)
	req.SetPathValue("workspaceSlug", "test")
	response := httptest.NewRecorder()
	if err := mid.Workspace(nil, resolver{scope.WorkspaceID})(handler)(ctx, response, req); err != nil {
		t.Fatal(err)
	}
	return response
}
func TestAuditCursorBoundToTenantActorAndFilters(t *testing.T) {
	now := time.Now().UTC()
	s := &testService{events: []domain.AuditEvent{{ID: uuid.New(), Source: "security", CreatedAt: now}, {ID: uuid.New(), Source: "security", CreatedAt: now.Add(-time.Second)}}}
	h, err := New(s, "test-secret")
	if err != nil {
		t.Fatal(err)
	}
	scope := domain.Scope{ActorID: uuid.New(), WorkspaceID: uuid.New()}
	first := request(t, h.Audit, scope, "/workspaces/test/security/audit?limit=1")
	if first.Code != http.StatusOK {
		t.Fatal(first.Body.String())
	}
	var body struct {
		Data struct {
			NextCursor string `json:"nextCursor"`
		} `json:"data"`
	}
	if err := json.Unmarshal(first.Body.Bytes(), &body); err != nil || body.Data.NextCursor == "" {
		t.Fatal("missing signed cursor", err)
	}
	base := "/workspaces/test/security/audit?limit=1&cursor=" + body.Data.NextCursor
	for _, scenario := range []struct {
		scope domain.Scope
		url   string
	}{{domain.Scope{ActorID: uuid.New(), WorkspaceID: scope.WorkspaceID}, base}, {domain.Scope{ActorID: scope.ActorID, WorkspaceID: uuid.New()}, base}, {scope, base + "&resourceType=story"}, {scope, base + "tampered"}} {
		before := s.calls
		response := request(t, h.Audit, scenario.scope, scenario.url)
		if response.Code != http.StatusBadRequest || s.calls != before {
			t.Fatal("cursor crossed authorization or filter boundary")
		}
	}
	if request(t, h.Audit, scope, base).Code != http.StatusOK {
		t.Fatal("valid cursor rejected")
	}
}
func TestAuditExportRejectsTruncationAndDoesNotAuditFailure(t *testing.T) {
	s := &testService{events: make([]domain.AuditEvent, 10001)}
	h, _ := New(s, "test-secret")
	scope := domain.Scope{ActorID: uuid.New(), WorkspaceID: uuid.New()}
	response := request(t, h.ExportAudit, scope, "/workspaces/test/security/audit/export")
	if response.Code != http.StatusBadRequest || s.exports != 0 {
		t.Fatal("oversized export was silently truncated or audited as success")
	}
	s.events = []domain.AuditEvent{{ID: uuid.New(), Source: "security", ActorType: "human_user", ResourceType: "workspace", Operation: "workspace.data_exported", Metadata: json.RawMessage(`{"count":42}`), CreatedAt: time.Now()}}
	response = request(t, h.ExportAudit, scope, "/workspaces/test/security/audit/export")
	if response.Code != http.StatusOK || s.exports != 1 || !strings.HasPrefix(response.Header().Get("Content-Type"), "text/csv") || !strings.Contains(response.Body.String(), "workspace.data_exported") {
		t.Fatal("CSV export not delivered and recorded")
	}
}
