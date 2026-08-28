package searchhttp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	search "github.com/complexus-tech/projects-api/internal/modules/search/service"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestParseSearchParamsUsesExplicitTypedValues(t *testing.T) {
	teamID := uuid.New()
	params, err := parseSearchParams(url.Values{
		"type":     {"stories"},
		"teamId":   {teamID.String()},
		"sortBy":   {"updated"},
		"page":     {"2"},
		"pageSize": {"40"},
	})

	require.NoError(t, err)
	require.Equal(t, search.SearchTypeStories, params.Type)
	require.Equal(t, search.SortByUpdated, params.SortBy)
	require.Equal(t, &teamID, params.TeamID)
	require.Equal(t, 2, params.Page)
	require.Equal(t, 40, params.PageSize)
}

func TestParseSearchParamsUsesDocumentedDefaults(t *testing.T) {
	t.Parallel()

	params, err := parseSearchParams(url.Values{})
	require.NoError(t, err)
	require.Equal(t, search.SearchTypeAll, params.Type)
	require.Equal(t, search.SortByRelevance, params.SortBy)
	require.Equal(t, 1, params.Page)
	require.Equal(t, defaultSearchPageSize, params.PageSize)
}

func TestParseSearchParamsRejectsMalformedInputs(t *testing.T) {
	t.Parallel()

	tests := map[string]url.Values{
		"unknown type":        {"type": {"files"}},
		"unknown sort":        {"sortBy": {"random"}},
		"bad team":            {"teamId": {"sensitive-team"}},
		"zero team":           {"teamId": {uuid.Nil.String()}},
		"repeated team":       {"teamId": {uuid.NewString(), uuid.NewString()}},
		"repeated query":      {"query": {"first-sensitive", "second-sensitive"}},
		"query too long":      {"query": {strings.Repeat("x", maximumSearchQueryRunes+1)}},
		"invalid query bytes": {"query": {string([]byte{0xff})}},
		"nul query":           {"query": {"sensitive\x00query"}},
		"bad page":            {"page": {"0"}},
		"page too large":      {"page": {"1001"}},
		"bad page size":       {"pageSize": {"many"}},
		"page size too large": {"pageSize": {"101"}},
		"repeated page size":  {"pageSize": {"20", "40"}},
		"overflowing page":    {"page": {strings.Repeat("9", 21)}},
	}

	for name, values := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			_, err := parseSearchParams(values)
			require.Error(t, err)
			for _, supplied := range values {
				for _, value := range supplied {
					if value != "" && strings.Contains(err.Error(), value) {
						t.Fatalf("error %q exposes query value", err)
					}
				}
			}
		})
	}
}

func TestParseSimilarStoriesQueryUsesTypedBoundsAndCap(t *testing.T) {
	t.Parallel()

	teamID := uuid.New()
	query, err := parseSimilarStoriesQuery(url.Values{
		"teamId": {teamID.String()},
		"title":  {"  Similar roadmap story  "},
		"limit":  {"500"},
	})
	require.NoError(t, err)
	require.Equal(t, &teamID, query.TeamID)
	require.Equal(t, "Similar roadmap story", query.Title)
	require.Equal(t, maximumSimilarStoriesLimit, query.Limit)

	defaults, err := parseSimilarStoriesQuery(url.Values{})
	require.NoError(t, err)
	require.Equal(t, defaultSimilarStoriesLimit, defaults.Limit)
}

func TestParseSimilarStoriesQueryRejectsUnsafeValuesWithoutEchoingThem(t *testing.T) {
	t.Parallel()

	for name, values := range map[string]url.Values{
		"repeated title":    {"title": {"first-sensitive", "second-sensitive"}},
		"oversized title":   {"title": {strings.Repeat("x", maximumSimilarityTitleRunes+1)}},
		"invalid team":      {"teamId": {"sensitive-team"}},
		"repeated limit":    {"limit": {"3", "5"}},
		"zero limit":        {"limit": {"0"}},
		"overflowing limit": {"limit": {strings.Repeat("9", 21)}},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			_, err := parseSimilarStoriesQuery(values)
			require.Error(t, err)
			for _, supplied := range values {
				for _, value := range supplied {
					if value != "" && strings.Contains(err.Error(), value) {
						t.Fatalf("error %q exposes query value", err)
					}
				}
			}
		})
	}
}

func TestSearchHandlerRejectsRepeatedQueryBeforeCallingService(t *testing.T) {
	t.Parallel()

	actorID, workspaceID := uuid.New(), uuid.New()
	request := httptest.NewRequest(http.MethodGet, "/workspaces/planning/search?query=first-sensitive&query=second-sensitive", nil)
	request.SetPathValue("workspaceSlug", "planning")
	request = request.WithContext(platformauth.SetUserID(request.Context(), actorID))
	recorder := httptest.NewRecorder()

	handler := New(nil)
	wrapped := mid.Workspace(nil, fixedSearchWorkspaceResolver{workspaceID: workspaceID})(handler.Search)
	if err := wrapped(request.Context(), recorder, request); err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("Search() status = %d, body=%s", recorder.Code, recorder.Body.String())
	}
	if strings.Contains(recorder.Body.String(), "first-sensitive") || strings.Contains(recorder.Body.String(), "second-sensitive") {
		t.Fatalf("response exposes rejected query value: %s", recorder.Body.String())
	}
}

type fixedSearchWorkspaceResolver struct {
	workspaceID uuid.UUID
}

func (resolver fixedSearchWorkspaceResolver) ResolveCurrentWorkspace(context.Context, string, uuid.UUID) (mid.WorkspaceInfo, error) {
	return mid.WorkspaceInfo{ID: resolver.workspaceID, Name: "Planning", Slug: "planning", UserRole: "member"}, nil
}

func (fixedSearchWorkspaceResolver) RecordWorkspaceAccess(context.Context, uuid.UUID, uuid.UUID) error {
	return nil
}
