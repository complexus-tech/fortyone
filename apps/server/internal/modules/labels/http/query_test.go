package labelshttp

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

func TestParseLabelListQueryProducesTypedFiltersAndPagination(t *testing.T) {
	t.Parallel()

	teamID := uuid.New()
	query, err := parseLabelListQuery(url.Values{
		"teamId":   {teamID.String()},
		"search":   {"  planning  "},
		"page":     {"2"},
		"pageSize": {"25"},
	})
	if err != nil {
		t.Fatalf("parse label list query: %v", err)
	}
	if query.Filters.TeamID == nil || *query.Filters.TeamID != teamID || query.Filters.Search != "planning" {
		t.Fatalf("typed filters = %#v", query.Filters)
	}
	if query.Page == nil || query.Page.Page != 2 || query.Page.PageSize != 25 {
		t.Fatalf("page = %#v", query.Page)
	}
	if query.Filters.Limit == nil || *query.Filters.Limit != 26 || query.Filters.Offset != 25 {
		t.Fatalf("repository pagination = %#v", query.Filters)
	}

	direct, err := parseLabelListQuery(url.Values{"limit": {"40"}, "offset": {"80"}})
	if err != nil {
		t.Fatalf("parse direct label pagination: %v", err)
	}
	if direct.Page != nil || direct.Filters.Limit == nil || *direct.Filters.Limit != 40 || direct.Filters.Offset != 80 {
		t.Fatalf("direct pagination = %#v", direct)
	}
}

func TestParseLabelListQueryRejectsAmbiguousOrInvalidInput(t *testing.T) {
	t.Parallel()

	for name, test := range map[string]struct {
		query url.Values
		cause error
	}{
		"repeated team": {
			query: url.Values{"teamId": {uuid.NewString(), uuid.NewString()}},
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
			query: url.Values{"search": {strings.Repeat("x", maximumLabelSearchBytes+1)}},
			cause: web.ErrQueryParameterTooLong,
		},
		"too many search characters": {
			query: url.Values{"search": {strings.Repeat("é", maximumLabelSearchRunes+1)}},
			cause: web.ErrInvalidQueryParameter,
		},
		"invalid search encoding": {
			query: url.Values{"search": {string([]byte{0xff})}},
			cause: web.ErrInvalidQueryParameter,
		},
		"invalid page": {
			query: url.Values{"page": {"page-secret-value"}},
			cause: web.ErrInvalidQueryParameter,
		},
		"repeated page size": {
			query: url.Values{"pageSize": {"10", "20"}},
			cause: web.ErrRepeatedQueryParameter,
		},
		"oversized integer": {
			query: url.Values{"page": {strings.Repeat("9", maximumLabelIntegerParameterBytes+1)}},
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
		"ambiguous pagination": {
			query: url.Values{"page": {"2"}, "limit": {"20"}},
			cause: web.ErrInvalidQueryParameter,
		},
		"offset without limit": {
			query: url.Values{"offset": {"20"}},
			cause: web.ErrInvalidQueryParameter,
		},
		"offset outside repository range": {
			query: url.Values{"limit": {"20"}, "offset": {"2147483648"}},
			cause: web.ErrInvalidQueryParameter,
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			_, err := parseLabelListQuery(test.query)
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

func TestListRejectsRepeatedLabelParametersBeforeCallingService(t *testing.T) {
	t.Parallel()

	actorID, workspaceID := uuid.New(), uuid.New()
	request := httptest.NewRequest(http.MethodGet, "/workspaces/planning/labels?teamId=sensitive-team&teamId="+uuid.NewString(), nil)
	request.SetPathValue("workspaceSlug", "planning")
	request = request.WithContext(platformauth.SetUserID(request.Context(), actorID))
	recorder := httptest.NewRecorder()

	handler := New(nil, nil)
	withWorkspace := mid.Workspace(nil, fixedLabelWorkspaceResolver{workspaceID: workspaceID})(handler.List)
	if err := withWorkspace(request.Context(), recorder, request); err != nil {
		t.Fatalf("list labels: %v", err)
	}
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
	if strings.Contains(recorder.Body.String(), "sensitive-team") {
		t.Fatalf("response exposes rejected query value: %s", recorder.Body.String())
	}
}

type fixedLabelWorkspaceResolver struct {
	workspaceID uuid.UUID
}

func (resolver fixedLabelWorkspaceResolver) ResolveCurrentWorkspace(
	_ context.Context,
	slug string,
	_ uuid.UUID,
) (mid.WorkspaceInfo, error) {
	return mid.WorkspaceInfo{ID: resolver.workspaceID, Slug: slug, UserRole: "member"}, nil
}

func (fixedLabelWorkspaceResolver) RecordWorkspaceAccess(context.Context, uuid.UUID, uuid.UUID) error {
	return nil
}
