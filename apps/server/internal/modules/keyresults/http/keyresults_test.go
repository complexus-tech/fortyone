package keyresultshttp

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	keyresultsdomain "github.com/complexus-tech/projects-api/internal/modules/keyresults/domain"
	keyresults "github.com/complexus-tech/projects-api/internal/modules/keyresults/service"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/date"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

func TestToKeyResultPatchPreservesZeroClearAndEmptyCollectionIntent(t *testing.T) {
	t.Parallel()

	zero := 0.0
	leadID := uuid.New()
	contributors := []uuid.UUID{}
	startTime := time.Date(2026, time.August, 28, 0, 0, 0, 0, time.UTC)
	endTime := startTime.AddDate(0, 0, 1)
	startDate, endDate := date.Date(startTime), date.Date(endTime)
	patch := toKeyResultPatch(AppUpdateKeyResult{
		Name: "Ship API", StartValue: &zero, CurrentValue: &zero, TargetValue: &zero,
		Lead: &leadID, ClearLead: true, Contributors: &contributors,
		StartDate: &startDate, EndDate: &endDate,
	})

	if !patch.Name.Set || patch.Name.Value != "Ship API" ||
		!patch.StartValue.Set || patch.StartValue.Value != 0 ||
		!patch.CurrentValue.Set || patch.CurrentValue.Value != 0 ||
		!patch.TargetValue.Set || patch.TargetValue.Value != 0 {
		t.Fatalf("scalar patch = %#v", patch)
	}
	if !patch.Lead.Set || patch.Lead.Value != nil {
		t.Fatalf("lead patch = %#v, want explicit clear", patch.Lead)
	}
	if !patch.Contributors.Set || patch.Contributors.Value == nil || len(patch.Contributors.Value) != 0 {
		t.Fatalf("contributors patch = %#v, want explicit empty list", patch.Contributors)
	}
	if !patch.StartDate.Set || patch.StartDate.Value == nil || !patch.StartDate.Value.Equal(startTime) ||
		!patch.EndDate.Set || patch.EndDate.Value == nil || !patch.EndDate.Value.Equal(endTime) {
		t.Fatalf("date patch = %#v/%#v", patch.StartDate, patch.EndDate)
	}
}

func TestParseKeyResultFiltersUsesStrictUUIDsDatesAndSharedPagination(t *testing.T) {
	t.Parallel()

	workspaceID, userID, teamID, objectiveID, leadID, creatorID := uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New()
	query := url.Values{
		"teamIds":          {teamID.String(), teamID.String()},
		"objectiveIds":     {objectiveID.String() + "," + uuid.New().String()},
		"leadIds":          {leadID.String()},
		"createdBy":        {creatorID.String()},
		"measurementTypes": {"number"},
		"createdAfter":     {"2026-08-28T10:00:00+02:00"},
		"updatedBefore":    {"2026-09-01T18:00:00+02:00"},
		"page":             {"3"},
		"pageSize":         {"1000"},
	}
	request := &http.Request{URL: &url.URL{RawQuery: query.Encode()}}

	filters, err := parseKeyResultFilters(request, workspaceID, userID)
	if err != nil {
		t.Fatalf("parseKeyResultFilters() error = %v", err)
	}
	if filters.WorkspaceID != workspaceID || filters.CurrentUserID != userID || filters.Page != 3 || filters.PageSize != keyresults.MaximumPageSize {
		t.Fatalf("scope/pagination = %#v", filters)
	}
	if len(filters.TeamIDs) != 1 || filters.TeamIDs[0] != teamID ||
		len(filters.ObjectiveIDs) != 2 || len(filters.LeadIDs) != 1 || filters.LeadIDs[0] != leadID ||
		len(filters.CreatedBy) != 1 || filters.CreatedBy[0] != creatorID {
		t.Fatalf("UUID filters = %#v", filters)
	}
	wantCreatedAfter := time.Date(2026, time.August, 28, 8, 0, 0, 0, time.UTC)
	if filters.CreatedAfter == nil || !filters.CreatedAfter.Equal(wantCreatedAfter) {
		t.Fatalf("createdAfter = %v, want %v", filters.CreatedAfter, wantCreatedAfter)
	}
	wantUpdatedBefore := time.Date(2026, time.September, 1, 16, 0, 0, 0, time.UTC)
	if filters.UpdatedBefore == nil || !filters.UpdatedBefore.Equal(wantUpdatedBefore) {
		t.Fatalf("updatedBefore = %v, want %v", filters.UpdatedBefore, wantUpdatedBefore)
	}
}

func TestParseKeyResultFiltersRejectsAmbiguousMalformedOversizedAndReversedInput(t *testing.T) {
	t.Parallel()

	workspaceID, userID := uuid.New(), uuid.New()
	tooManyIDs := make([]string, maximumKeyResultFilterItems+1)
	for index := range tooManyIDs {
		tooManyIDs[index] = uuid.NewString()
	}
	for name, test := range map[string]struct {
		query url.Values
		cause error
	}{
		"repeated page": {
			query: url.Values{"page": {"1", "2"}}, cause: web.ErrRepeatedQueryParameter,
		},
		"malformed page": {
			query: url.Values{"page": {"not-a-page"}}, cause: web.ErrInvalidQueryParameter,
		},
		"oversized page": {
			query: url.Values{"page": {strings.Repeat("9", 21)}}, cause: web.ErrQueryParameterTooLong,
		},
		"offset out of range": {
			query: url.Values{"page": {"2147483649"}, "pageSize": {"1"}}, cause: web.ErrInvalidQueryParameter,
		},
		"repeated sort": {
			query: url.Values{"orderBy": {"name", "created_at"}}, cause: web.ErrRepeatedQueryParameter,
		},
		"blank sort": {
			query: url.Values{"orderBy": {""}}, cause: web.ErrInvalidQueryParameter,
		},
		"unsupported sort": {
			query: url.Values{"orderBy": {"private_column"}}, cause: ErrInvalidFilters,
		},
		"unsupported direction": {
			query: url.Values{"orderDirection": {"sideways"}}, cause: ErrInvalidFilters,
		},
		"too many UUIDs": {
			query: url.Values{"teamIds": {strings.Join(tooManyIDs, ",")}},
			cause: web.ErrInvalidQueryParameter,
		},
		"repeated date": {
			query: url.Values{"createdAfter": {"2026-08-01T00:00:00Z", "2026-08-02T00:00:00Z"}},
			cause: web.ErrRepeatedQueryParameter,
		},
		"blank date": {
			query: url.Values{"createdAfter": {""}}, cause: web.ErrInvalidQueryParameter,
		},
		"oversized date": {
			query: url.Values{"createdAfter": {strings.Repeat("x", maximumKeyResultDateBytes+1)}},
			cause: web.ErrQueryParameterTooLong,
		},
		"reversed created range": {
			query: url.Values{
				"createdAfter": {"2026-08-02T00:00:00Z"}, "createdBefore": {"2026-08-01T00:00:00Z"},
			},
			cause: ErrInvalidFilters,
		},
		"reversed end range": {
			query: url.Values{
				"endDateAfter": {"2026-08-02T00:00:00Z"}, "endDateBefore": {"2026-08-01T00:00:00Z"},
			},
			cause: ErrInvalidFilters,
		},
		"reversed updated range": {
			query: url.Values{
				"updatedAfter": {"2026-08-02T00:00:00Z"}, "updatedBefore": {"2026-08-01T00:00:00Z"},
			},
			cause: ErrInvalidFilters,
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			request := &http.Request{URL: &url.URL{RawQuery: test.query.Encode()}}
			_, err := parseKeyResultFilters(request, workspaceID, userID)
			if !errors.Is(err, ErrInvalidFilters) || !errors.Is(err, test.cause) {
				t.Fatalf("error = %v, want ErrInvalidFilters and %v", err, test.cause)
			}
		})
	}
}

func TestParseKeyResultFiltersRejectsInputsThatWouldBroadenAQuery(t *testing.T) {
	t.Parallel()

	workspaceID, userID := uuid.New(), uuid.New()
	for _, rawQuery := range []string{
		"teamIds=not-a-uuid",
		"objectiveIds=" + uuid.Nil.String(),
		"leadIds=" + uuid.New().String() + "%2C",
		"createdBy=not-a-user",
		"createdAfter=yesterday",
		"endDateBefore=2026-08-28",
		"updatedAfter=tomorrow",
	} {
		request := &http.Request{URL: &url.URL{RawQuery: rawQuery}}
		if _, err := parseKeyResultFilters(request, workspaceID, userID); !errors.Is(err, ErrInvalidFilters) {
			t.Fatalf("parseKeyResultFilters(%q) error = %v, want ErrInvalidFilters", rawQuery, err)
		}
	}
}

func TestKeyResultErrorStatus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		err  error
		want int
	}{
		{err: keyresults.ErrInvalid, want: http.StatusBadRequest},
		{err: keyresults.ErrInvalidReference, want: http.StatusBadRequest},
		{err: keyresults.ErrForbidden, want: http.StatusForbidden},
		{err: keyresults.ErrNotFound, want: http.StatusNotFound},
		{err: keyresults.ErrVersionConflict, want: http.StatusConflict},
		{err: errors.New("database unavailable"), want: http.StatusInternalServerError},
	}
	for _, test := range tests {
		if got := keyResultErrorStatus(test.err); got != test.want {
			t.Errorf("keyResultErrorStatus(%v) = %d, want %d", test.err, got, test.want)
		}
	}
}

func TestListPaginatedHTTPBindsWorkspaceActorAndRejectsBroadeningFilters(t *testing.T) {
	t.Parallel()

	workspaceID, actorID, keyResultID := uuid.New(), uuid.New(), uuid.New()
	repository := &keyResultHTTPRepositoryStub{pageResult: keyresultsdomain.ListResponse{
		KeyResults: []keyresultsdomain.KeyResultWithObjective{{
			KeyResult:   keyresultsdomain.KeyResult{ID: keyResultID, Name: "Ship API"},
			WorkspaceID: workspaceID,
		}},
		TotalCount: 1, Page: 2, PageSize: 10,
	}}
	handler := New(keyresults.New(nil, repository), nil, nil, nil, nil)
	resolver := keyResultWorkspaceResolverStub{workspace: mid.WorkspaceInfo{
		ID: workspaceID, Name: "Workspace", Slug: "workspace", UserRole: "member",
	}}
	log := logger.NewWithText(io.Discard, slog.LevelError, "key-results-http-test")
	wrapped := mid.Workspace(log, resolver)(handler.ListPaginated)

	request := httptest.NewRequest(http.MethodGet, "/workspaces/workspace/key-results?page=2&pageSize=10", nil)
	request.SetPathValue("workspaceSlug", "workspace")
	recorder := httptest.NewRecorder()
	if err := wrapped(platformauth.SetUserID(context.Background(), actorID), recorder, request); err != nil {
		t.Fatalf("ListPaginated() error = %v", err)
	}
	if recorder.Code != http.StatusOK {
		t.Fatalf("ListPaginated() status = %d, body=%s", recorder.Code, recorder.Body.String())
	}
	if repository.pageCalls != 1 || repository.pageQuery.Access.WorkspaceID != workspaceID ||
		repository.pageQuery.Access.ActorID != actorID || !repository.pageQuery.Access.AllTeams ||
		repository.pageQuery.Filters.Page != 2 || repository.pageQuery.Filters.PageSize != 10 {
		t.Fatalf("repository page query = %#v", repository.pageQuery)
	}
	var response struct {
		Data AppKeyResultListResponse `json:"data"`
	}
	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.Data.TotalCount != 1 || len(response.Data.KeyResults) != 1 || response.Data.KeyResults[0].ID != keyResultID {
		t.Fatalf("response data = %#v", response.Data)
	}

	for _, rawQuery := range []string{
		"teamIds=not-a-uuid",
		"page=not-a-page",
		"orderBy=name&orderBy=created_at",
		"createdAfter=2026-09-01T00%3A00%3A00Z&createdBefore=2026-08-01T00%3A00%3A00Z",
	} {
		invalidRequest := httptest.NewRequest(http.MethodGet, "/workspaces/workspace/key-results?"+rawQuery, nil)
		invalidRequest.SetPathValue("workspaceSlug", "workspace")
		invalidRecorder := httptest.NewRecorder()
		if err := wrapped(platformauth.SetUserID(context.Background(), actorID), invalidRecorder, invalidRequest); err != nil {
			t.Fatalf("invalid ListPaginated(%q) error = %v", rawQuery, err)
		}
		if invalidRecorder.Code != http.StatusBadRequest || repository.pageCalls != 1 {
			t.Fatalf(
				"invalid filter %q status/calls = %d/%d, body=%s",
				rawQuery,
				invalidRecorder.Code,
				repository.pageCalls,
				invalidRecorder.Body.String(),
			)
		}
	}
}

type keyResultWorkspaceResolverStub struct {
	workspace mid.WorkspaceInfo
}

func (resolver keyResultWorkspaceResolverStub) ResolveCurrentWorkspace(context.Context, string, uuid.UUID) (mid.WorkspaceInfo, error) {
	return resolver.workspace, nil
}

func (keyResultWorkspaceResolverStub) RecordWorkspaceAccess(context.Context, uuid.UUID, uuid.UUID) error {
	return nil
}

type keyResultHTTPRepositoryStub struct {
	pageQuery  keyresultsdomain.PaginatedListQuery
	pageResult keyresultsdomain.ListResponse
	pageCalls  int
}

func (*keyResultHTTPRepositoryStub) CreateBatch(context.Context, keyresultsdomain.CreateCommand) ([]keyresultsdomain.KeyResult, error) {
	return nil, nil
}

func (*keyResultHTTPRepositoryStub) Update(context.Context, keyresultsdomain.UpdateCommand) (keyresultsdomain.MutationResult, error) {
	return keyresultsdomain.MutationResult{}, nil
}

func (*keyResultHTTPRepositoryStub) Delete(context.Context, keyresultsdomain.DeleteCommand) error {
	return nil
}

func (*keyResultHTTPRepositoryStub) Get(context.Context, keyresultsdomain.GetQuery) (keyresultsdomain.KeyResult, error) {
	return keyresultsdomain.KeyResult{}, nil
}

func (*keyResultHTTPRepositoryStub) List(context.Context, keyresultsdomain.ObjectiveListQuery) ([]keyresultsdomain.KeyResult, error) {
	return nil, nil
}

func (repository *keyResultHTTPRepositoryStub) ListPaginated(
	_ context.Context,
	query keyresultsdomain.PaginatedListQuery,
) (keyresultsdomain.ListResponse, error) {
	repository.pageCalls++
	repository.pageQuery = query
	return repository.pageResult, nil
}
