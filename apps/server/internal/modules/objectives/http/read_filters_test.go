package objectiveshttp

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

func TestParseObjectiveListFilters(t *testing.T) {
	t.Parallel()

	teamID := uuid.New()
	filters, err := parseObjectiveListFilters(url.Values{
		"search": {"  quarterly plan  "},
		"teamId": {teamID.String()},
	})
	if err != nil {
		t.Fatalf("parse objective filters: %v", err)
	}
	if filters.Search != "quarterly plan" || filters.TeamID == nil || *filters.TeamID != teamID {
		t.Fatalf("filters = %#v", filters)
	}
}

func TestParseObjectiveListFiltersRejectsAmbiguousOrInvalidInput(t *testing.T) {
	t.Parallel()

	for name, test := range map[string]struct {
		query url.Values
		cause error
	}{
		"repeated search": {
			query: url.Values{"search": {"first", "second"}},
			cause: web.ErrRepeatedQueryParameter,
		},
		"oversized search": {
			query: url.Values{"search": {strings.Repeat("x", maximumObjectiveSearchBytes+1)}},
			cause: web.ErrQueryParameterTooLong,
		},
		"too many search characters": {
			query: url.Values{"search": {strings.Repeat("é", maximumObjectiveSearchRunes+1)}},
			cause: web.ErrInvalidQueryParameter,
		},
		"search contains NUL": {
			query: url.Values{"search": {"roadmap\x00private"}},
			cause: web.ErrInvalidQueryParameter,
		},
		"invalid team": {
			query: url.Values{"teamId": {"not-a-team"}},
			cause: web.ErrInvalidQueryParameter,
		},
		"blank team": {
			query: url.Values{"teamId": {""}},
			cause: web.ErrInvalidQueryParameter,
		},
		"zero team": {
			query: url.Values{"teamId": {uuid.Nil.String()}},
			cause: web.ErrInvalidQueryParameter,
		},
		"repeated team": {
			query: url.Values{"teamId": {uuid.NewString(), uuid.NewString()}},
			cause: web.ErrRepeatedQueryParameter,
		},
		"oversized team": {
			query: url.Values{"teamId": {strings.Repeat("x", maximumObjectiveTeamIDBytes+1)}},
			cause: web.ErrQueryParameterTooLong,
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if _, err := parseObjectiveListFilters(test.query); !errors.Is(err, test.cause) {
				t.Fatalf("error = %v, want %v", err, test.cause)
			}
		})
	}
}

func TestParseObjectiveListPaginationPreservesOptionalResponseShapeAndStrictBounds(t *testing.T) {
	t.Parallel()

	params, err := parseObjectiveListPagination(url.Values{})
	if err != nil || params != nil {
		t.Fatalf("absent pagination = %#v, %v, want nil", params, err)
	}

	params, err = parseObjectiveListPagination(url.Values{
		"page": {"3"}, "pageSize": {"1000"},
	})
	if err != nil {
		t.Fatalf("parse objective pagination: %v", err)
	}
	if params == nil || params.Page != 3 || params.PageSize != 100 || params.Offset() != 200 {
		t.Fatalf("pagination = %#v", params)
	}
}

func TestParseObjectivePaginationRejectsAmbiguousMalformedAndOverflowingInput(t *testing.T) {
	t.Parallel()

	for name, query := range map[string]url.Values{
		"repeated page":       {"page": {"1", "2"}},
		"blank page":          {"page": {""}},
		"malformed page":      {"page": {"not-a-page"}},
		"zero page size":      {"pageSize": {"0"}},
		"oversized page":      {"page": {strings.Repeat("9", 21)}},
		"offset out of range": {"page": {"2147483649"}, "pageSize": {"1"}},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if _, err := parseObjectiveListPagination(query); err == nil {
				t.Fatal("expected list pagination error")
			}
			if _, err := parseObjectiveActivityPagination(query); err == nil {
				t.Fatal("expected activity pagination error")
			}
		})
	}

	params, err := parseObjectiveActivityPagination(url.Values{})
	if err != nil || params.Page != 1 || params.PageSize != 20 || params.Offset() != 0 {
		t.Fatalf("activity defaults = %#v, %v", params, err)
	}
}

func TestListHTTPReturnsBadRequestForInvalidExplicitQueryBeforeCallingService(t *testing.T) {
	t.Parallel()

	workspaceID, actorID := uuid.New(), uuid.New()
	handler := New(nil, nil, nil, nil, nil, nil)
	log := logger.NewWithText(io.Discard, slog.LevelError, "objectives-http-test")
	wrapped := mid.Workspace(log, objectiveWorkspaceResolverStub{workspaceID: workspaceID})(handler.List)

	for _, rawQuery := range []string{
		"page=not-a-page",
		"page=1&page=2",
		"search=first&search=second",
		"teamId=",
	} {
		request := httptest.NewRequest(http.MethodGet, "/workspaces/workspace/objectives?"+rawQuery, nil)
		request.SetPathValue("workspaceSlug", "workspace")
		recorder := httptest.NewRecorder()
		if err := wrapped(platformauth.SetUserID(context.Background(), actorID), recorder, request); err != nil {
			t.Fatalf("List(%q) error = %v", rawQuery, err)
		}
		if recorder.Code != http.StatusBadRequest {
			t.Fatalf("List(%q) status = %d, body=%s", rawQuery, recorder.Code, recorder.Body.String())
		}
	}
}

type objectiveWorkspaceResolverStub struct {
	workspaceID uuid.UUID
}

func (resolver objectiveWorkspaceResolverStub) ResolveCurrentWorkspace(
	context.Context,
	string,
	uuid.UUID,
) (mid.WorkspaceInfo, error) {
	return mid.WorkspaceInfo{ID: resolver.workspaceID, Slug: "workspace", UserRole: "member"}, nil
}

func (objectiveWorkspaceResolverStub) RecordWorkspaceAccess(context.Context, uuid.UUID, uuid.UUID) error {
	return nil
}
