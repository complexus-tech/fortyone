package integrationrequestshttp

import (
	"errors"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	integrationrequests "github.com/complexus-tech/projects-api/internal/modules/integrationrequests/service"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

func TestParseRequestListQueryMapsTypedFilters(t *testing.T) {
	t.Parallel()

	assigneeID := uuid.New()
	request := httptest.NewRequest(
		"GET",
		"/requests?status=pending&provider=github&priority=High&search=%20roadmap%20&assigneeId="+assigneeID.String()+"&createdAfter=2026-08-01&createdBefore=2026-08-28T12:00:00Z&page=3&pageSize=50",
		nil,
	)
	got, err := parseRequestListQuery(request)
	if err != nil {
		t.Fatalf("parseRequestListQuery() error = %v", err)
	}
	if got.Page != 3 || got.PageSize != 50 || got.Filter.Page != 3 || got.Filter.PageSize != 51 {
		t.Fatalf("pagination = %#v", got)
	}
	if got.Filter.Status != integrationrequests.StatusPending || got.Filter.Provider != integrationrequests.ProviderGitHub || got.Filter.Priority != "High" || got.Filter.Search != "roadmap" {
		t.Fatalf("filters = %#v", got.Filter)
	}
	if got.Filter.AssigneeID == nil || *got.Filter.AssigneeID != assigneeID {
		t.Fatalf("assignee = %v", got.Filter.AssigneeID)
	}
	if got.Filter.CreatedAfter == nil || got.Filter.CreatedBefore == nil || got.Filter.CreatedAfter.Format(time.DateOnly) != "2026-08-01" {
		t.Fatalf("date filters = %v, %v", got.Filter.CreatedAfter, got.Filter.CreatedBefore)
	}
}

func TestParseRequestListQueryRejectsAmbiguousMalformedAndOversizedValues(t *testing.T) {
	t.Parallel()

	for name, test := range map[string]struct {
		query string
		cause error
	}{
		"repeated enum": {
			query: "status=pending&status=accepted", cause: web.ErrRepeatedQueryParameter,
		},
		"unknown enum": {
			query: "provider=sensitive-provider", cause: web.ErrInvalidQueryParameter,
		},
		"invalid utf8": {
			query: "search=%FF", cause: web.ErrInvalidQueryParameter,
		},
		"nul search": {
			query: "search=sensitive%00search", cause: web.ErrInvalidQueryParameter,
		},
		"oversized search": {
			query: "search=" + strings.Repeat("x", maximumRequestSearchRunes*4+1), cause: web.ErrQueryParameterTooLong,
		},
		"zero assignee": {
			query: "assigneeId=" + uuid.Nil.String(), cause: web.ErrInvalidQueryParameter,
		},
		"invalid date": {
			query: "createdAfter=sensitive-date", cause: web.ErrInvalidQueryParameter,
		},
		"reversed dates": {
			query: "createdAfter=2026-08-29&createdBefore=2026-08-28", cause: nil,
		},
		"repeated page": {
			query: "page=1&page=2", cause: web.ErrRepeatedQueryParameter,
		},
		"overflow page": {
			query: "page=99999999999999999999999999", cause: web.ErrQueryParameterTooLong,
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			request := httptest.NewRequest("GET", "/requests?"+test.query, nil)
			_, err := parseRequestListQuery(request)
			if err == nil {
				t.Fatal("expected invalid request query")
			}
			if test.cause != nil && !errors.Is(err, test.cause) {
				t.Fatalf("error = %v, want cause %v", err, test.cause)
			}
			for _, sensitive := range []string{"sensitive-provider", "sensitive\x00search", "sensitive-date"} {
				if strings.Contains(err.Error(), sensitive) {
					t.Fatalf("error exposes supplied value: %v", err)
				}
			}
		})
	}
}
