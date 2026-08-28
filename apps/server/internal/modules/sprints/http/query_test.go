package sprintshttp

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	sprintdomain "github.com/complexus-tech/projects-api/internal/modules/sprints/domain"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

func TestParseSprintListQueryProducesTypedDomainQuery(t *testing.T) {
	t.Parallel()

	workspaceID, actorID := uuid.New(), uuid.New()
	objectiveID, teamID := uuid.New(), uuid.New()
	query, err := parseSprintListQuery(url.Values{
		"objectiveId": {objectiveID.String()},
		"teamId":      {teamID.String()},
		"search":      {"  quarterly plan  "},
		"page":        {"2"},
		"pageSize":    {"25"},
	}, workspaceID, actorID)
	if err != nil {
		t.Fatalf("parse sprint list query: %v", err)
	}
	if query.Query.WorkspaceID != workspaceID || query.Query.ActorID != actorID {
		t.Fatalf("query scope = %#v", query.Query)
	}
	if query.Query.Filter.ObjectiveID == nil || *query.Query.Filter.ObjectiveID != objectiveID ||
		query.Query.Filter.TeamID == nil || *query.Query.Filter.TeamID != teamID {
		t.Fatalf("identity filters = %#v", query.Query.Filter)
	}
	if query.Query.Filter.Search != "quarterly plan" {
		t.Fatalf("search = %q, want normalized query", query.Query.Filter.Search)
	}
	if query.Page == nil || query.Page.Page != 2 || query.Page.PageSize != 25 {
		t.Fatalf("page = %#v", query.Page)
	}
	if query.Query.Filter.Limit != 26 || query.Query.Filter.Offset != 25 {
		t.Fatalf("repository pagination = %#v", query.Query.Filter)
	}

	defaults, err := parseSprintListQuery(url.Values{}, workspaceID, actorID)
	if err != nil {
		t.Fatalf("parse default sprint list query: %v", err)
	}
	if defaults.Page != nil || defaults.Query.Filter.Limit != sprintdomain.DefaultListLimit || defaults.Query.Filter.Offset != 0 {
		t.Fatalf("default query = %#v", defaults)
	}
}

func TestParseSprintListQueryRejectsAmbiguousOrInvalidInput(t *testing.T) {
	t.Parallel()

	workspaceID, actorID := uuid.New(), uuid.New()
	for name, test := range map[string]struct {
		query url.Values
		cause error
	}{
		"repeated objective": {
			query: url.Values{"objectiveId": {uuid.NewString(), uuid.NewString()}},
			cause: web.ErrRepeatedQueryParameter,
		},
		"invalid team": {
			query: url.Values{"teamId": {"sensitive-team-value"}},
			cause: web.ErrInvalidQueryParameter,
		},
		"oversized team": {
			query: url.Values{"teamId": {strings.Repeat("x", web.DefaultMaxQueryParameterBytes+1)}},
			cause: web.ErrQueryParameterTooLong,
		},
		"repeated search": {
			query: url.Values{"search": {"first", "second"}},
			cause: web.ErrRepeatedQueryParameter,
		},
		"oversized search": {
			query: url.Values{"search": {strings.Repeat("x", maximumSprintSearchBytes+1)}},
			cause: web.ErrQueryParameterTooLong,
		},
		"too many search characters": {
			query: url.Values{"search": {strings.Repeat("é", sprintdomain.MaximumSearchLength+1)}},
			cause: sprintdomain.ErrInvalid,
		},
		"invalid search encoding": {
			query: url.Values{"search": {string([]byte{0xff})}},
			cause: web.ErrInvalidQueryParameter,
		},
		"invalid page": {
			query: url.Values{"page": {"page-secret-value"}},
			cause: web.ErrInvalidQueryParameter,
		},
		"repeated page": {
			query: url.Values{"page": {"1", "2"}},
			cause: web.ErrRepeatedQueryParameter,
		},
		"oversized integer": {
			query: url.Values{"page": {strings.Repeat("9", maximumSprintIntegerParameterBytes+1)}},
			cause: web.ErrQueryParameterTooLong,
		},
		"non-positive page": {
			query: url.Values{"page": {"0"}},
			cause: web.ErrInvalidQueryParameter,
		},
		"oversized page size": {
			query: url.Values{"pageSize": {"101"}},
			cause: web.ErrInvalidQueryParameter,
		},
		"offset outside repository range": {
			query: url.Values{"page": {"2147483649"}, "pageSize": {"1"}},
			cause: web.ErrInvalidQueryParameter,
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			_, err := parseSprintListQuery(test.query, workspaceID, actorID)
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

func TestListRejectsRepeatedSprintParametersBeforeCallingService(t *testing.T) {
	t.Parallel()

	actorID, workspaceID := uuid.New(), uuid.New()
	request := httptest.NewRequest(http.MethodGet, "/workspaces/planning/sprints?search=sensitive-search&search=second", nil)
	request.SetPathValue("workspaceSlug", "planning")
	request = request.WithContext(platformauth.SetUserID(request.Context(), actorID))
	recorder := httptest.NewRecorder()

	handler := New(nil, nil)
	withWorkspace := mid.Workspace(nil, fixedSprintWorkspaceResolver{workspaceID: workspaceID})(handler.List)
	if err := withWorkspace(request.Context(), recorder, request); err != nil {
		t.Fatalf("list sprints: %v", err)
	}
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
	if strings.Contains(recorder.Body.String(), "sensitive-search") {
		t.Fatalf("response exposes rejected query value: %s", recorder.Body.String())
	}
}

type fixedSprintWorkspaceResolver struct {
	workspaceID uuid.UUID
}

func (resolver fixedSprintWorkspaceResolver) ResolveCurrentWorkspace(
	_ context.Context,
	slug string,
	_ uuid.UUID,
) (mid.WorkspaceInfo, error) {
	return mid.WorkspaceInfo{ID: resolver.workspaceID, Slug: slug, UserRole: "member"}, nil
}

func (fixedSprintWorkspaceResolver) RecordWorkspaceAccess(context.Context, uuid.UUID, uuid.UUID) error {
	return nil
}
