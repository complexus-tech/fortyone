package adminhttp

import (
	"errors"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

func TestAdminPaginationParamsAreStrictAndBounded(t *testing.T) {
	t.Parallel()

	request := httptest.NewRequest("GET", "/admin/users?page=3&limit=250", nil)
	page, limit, err := paginationParams(request)
	if err != nil || page != 3 || limit != maxPageSize {
		t.Fatalf("paginationParams() = %d, %d, %v", page, limit, err)
	}

	for name, query := range map[string]string{
		"blank":          "page=",
		"malformed":      "page=sensitive-page-value",
		"zero":           "page=0",
		"repeated page":  "page=1&page=2",
		"repeated limit": "limit=10&limit=20",
		"overflow":       "page=999999999999999999999999999",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			request := httptest.NewRequest("GET", "/admin/users?"+query, nil)
			_, _, err := paginationParams(request)
			if err == nil || strings.Contains(err.Error(), "sensitive-page-value") {
				t.Fatalf("error = %v", err)
			}
		})
	}
}

func TestAdminOptionalUUIDQueryRejectsAmbiguousAndZeroValues(t *testing.T) {
	t.Parallel()

	want := uuid.New()
	request := httptest.NewRequest("GET", "/admin?workspaceId="+want.String(), nil)
	got, err := optionalUUIDQuery(request, "workspaceId")
	if err != nil || got == nil || *got != want {
		t.Fatalf("optionalUUIDQuery() = %v, %v", got, err)
	}

	for name, query := range map[string]string{
		"malformed": "workspaceId=sensitive-invalid-uuid",
		"zero":      "workspaceId=" + uuid.Nil.String(),
		"repeated":  "workspaceId=" + want.String() + "&workspaceId=" + uuid.NewString(),
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			request := httptest.NewRequest("GET", "/admin?"+query, nil)
			_, err := optionalUUIDQuery(request, "workspaceId")
			if err == nil || strings.Contains(err.Error(), "sensitive-invalid-uuid") {
				t.Fatalf("error = %v", err)
			}
		})
	}
}

func TestParseAuditLogQueryValidatesTextDatesAndRange(t *testing.T) {
	t.Parallel()

	workspaceID := uuid.New()
	request := httptest.NewRequest(
		"GET",
		"/admin/audit-logs?workspaceId="+workspaceID.String()+"&targetType=story&q=%20roadmap%20&action=updated&actor=maya&from=2026-08-01&to=2026-08-28T12:00:00Z",
		nil,
	)
	got, err := parseAuditLogQuery(request)
	if err != nil {
		t.Fatalf("parseAuditLogQuery() error = %v", err)
	}
	if got.WorkspaceID == nil || *got.WorkspaceID != workspaceID || got.TargetType != "story" || got.Query != "roadmap" || got.Action != "updated" || got.ActorQuery != "maya" {
		t.Fatalf("filters = %#v", got)
	}
	if got.From == nil || got.To == nil || got.From.Format(time.DateOnly) != "2026-08-01" || got.To.Format(time.RFC3339) != "2026-08-28T12:00:00Z" {
		t.Fatalf("date filters = %v, %v", got.From, got.To)
	}

	for name, query := range map[string]string{
		"repeated text":  "q=first&q=second",
		"nul text":       "q=sensitive%00value",
		"oversized text": "q=" + strings.Repeat("x", maximumAdminSearchRunes*4+1),
		"invalid from":   "from=sensitive-date-value",
		"reversed range": "from=2026-08-29&to=2026-08-28",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			request := httptest.NewRequest("GET", "/admin/audit-logs?"+query, nil)
			_, err := parseAuditLogQuery(request)
			if err == nil {
				t.Fatal("expected invalid audit filters to fail")
			}
			if strings.Contains(err.Error(), "sensitive-date-value") || strings.Contains(err.Error(), "sensitive\x00value") {
				t.Fatalf("error exposes supplied value: %v", err)
			}
		})
	}
}

func TestAdminQueryErrorsRetainTypedCauses(t *testing.T) {
	t.Parallel()

	request := httptest.NewRequest("GET", "/admin?q=one&q=two", nil)
	_, err := optionalTextQuery(request, "q", maximumAdminSearchRunes)
	if !errors.Is(err, web.ErrRepeatedQueryParameter) {
		t.Fatalf("error = %v, want repeated query cause", err)
	}
}
