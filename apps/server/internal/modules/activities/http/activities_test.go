package activitieshttp

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/date"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

func TestParseActivityListQueryProducesTypedBoundedInput(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.August, 28, 16, 45, 0, 0, time.UTC)
	query, err := parseActivityListQuery(url.Values{
		"limit":     {"25"},
		"startDate": {"2026-08-01"},
		"endDate":   {"2026-08-12"},
	}, now)
	if err != nil {
		t.Fatalf("parse activity list query: %v", err)
	}
	if query.Limit != 25 {
		t.Fatalf("limit = %d, want 25", query.Limit)
	}
	if got := query.Filters.StartDate.Format(time.DateOnly); got != "2026-08-01" {
		t.Fatalf("start date = %s, want 2026-08-01", got)
	}
	wantEnd := date.EndOfDay(time.Date(2026, time.August, 12, 0, 0, 0, 0, time.UTC))
	if !query.Filters.EndDate.Equal(wantEnd) {
		t.Fatalf("end date = %s, want %s", query.Filters.EndDate, wantEnd)
	}

	defaults, err := parseActivityListQuery(url.Values{}, now)
	if err != nil {
		t.Fatalf("parse default activity list query: %v", err)
	}
	if defaults.Limit != defaultActivityLimit || defaults.Filters.StartDate.Format(time.DateOnly) != "2026-07-29" {
		t.Fatalf("default query = %#v", defaults)
	}
	wantDefaultEnd := date.EndOfDay(time.Date(2026, time.August, 28, 0, 0, 0, 0, time.UTC))
	if !defaults.Filters.EndDate.Equal(wantDefaultEnd) {
		t.Fatalf("default end date = %s, want %s", defaults.Filters.EndDate, wantDefaultEnd)
	}
}

func TestParseActivityListQueryRejectsAmbiguousOrInvalidInput(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.August, 28, 12, 0, 0, 0, time.UTC)
	for name, test := range map[string]struct {
		query url.Values
		cause error
	}{
		"repeated limit": {
			query: url.Values{"limit": {"10", "20"}},
			cause: web.ErrRepeatedQueryParameter,
		},
		"oversized limit": {
			query: url.Values{"limit": {strings.Repeat("9", maximumActivityLimitBytes+1)}},
			cause: web.ErrQueryParameterTooLong,
		},
		"invalid limit": {
			query: url.Values{"limit": {"secret"}},
			cause: ErrInvalidLimit,
		},
		"out of range limit": {
			query: url.Values{"limit": {"101"}},
			cause: ErrInvalidLimit,
		},
		"repeated date": {
			query: url.Values{"startDate": {"2026-08-01", "2026-08-02"}},
			cause: web.ErrRepeatedQueryParameter,
		},
		"invalid date": {
			query: url.Values{"startDate": {"date-secret-value"}},
			cause: ErrInvalidDate,
		},
		"reversed range": {
			query: url.Values{"startDate": {"2026-08-20"}, "endDate": {"2026-08-10"}},
			cause: ErrInvalidDate,
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			_, err := parseActivityListQuery(test.query, now)
			if !errors.Is(err, test.cause) {
				t.Fatalf("error = %v, want %v", err, test.cause)
			}
			for _, values := range test.query {
				for _, value := range values {
					if err != nil && strings.Contains(err.Error(), value) {
						t.Fatalf("error %q exposes query value", err)
					}
				}
			}
		})
	}
}

func TestGetActivitiesRejectsRepeatedParametersBeforeCallingService(t *testing.T) {
	t.Parallel()

	actorID, workspaceID := uuid.New(), uuid.New()
	request := httptest.NewRequest(http.MethodGet, "/workspaces/planning/activities?limit=sensitive-value&limit=20", nil)
	request.SetPathValue("workspaceSlug", "planning")
	request = request.WithContext(platformauth.SetUserID(request.Context(), actorID))
	recorder := httptest.NewRecorder()

	handler := New(nil, nil, nil)
	withWorkspace := mid.Workspace(nil, fixedActivityWorkspaceResolver{workspaceID: workspaceID})(handler.GetActivities)
	if err := withWorkspace(request.Context(), recorder, request); err != nil {
		t.Fatalf("get activities: %v", err)
	}
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
	if strings.Contains(recorder.Body.String(), "sensitive-value") {
		t.Fatalf("response exposes rejected query value: %s", recorder.Body.String())
	}
}

type fixedActivityWorkspaceResolver struct {
	workspaceID uuid.UUID
}

func (resolver fixedActivityWorkspaceResolver) ResolveCurrentWorkspace(
	_ context.Context,
	slug string,
	_ uuid.UUID,
) (mid.WorkspaceInfo, error) {
	return mid.WorkspaceInfo{ID: resolver.workspaceID, Slug: slug, UserRole: "member"}, nil
}

func (fixedActivityWorkspaceResolver) RecordWorkspaceAccess(
	context.Context,
	uuid.UUID,
	uuid.UUID,
) error {
	return nil
}
