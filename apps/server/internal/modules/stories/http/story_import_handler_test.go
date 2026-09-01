package storieshttp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"slices"
	"strings"
	"testing"

	storydomain "github.com/complexus-tech/projects-api/internal/modules/stories/domain"
	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/cache"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

type storyImportServiceStub struct {
	createdByKey map[string]uuid.UUID
	calls        []stories.CoreNewStory
	failTitle    string
	failErr      error
}

func (service *storyImportServiceStub) CreateExternal(
	_ context.Context,
	_ uuid.UUID,
	story stories.CoreNewStory,
	workspaceID uuid.UUID,
) (stories.CoreSingleStory, error) {
	service.calls = append(service.calls, story)
	if story.Title == service.failTitle {
		return stories.CoreSingleStory{}, service.failErr
	}
	if story.CreationKey == nil {
		return stories.CoreSingleStory{}, errors.New("missing creation key")
	}
	if service.createdByKey == nil {
		service.createdByKey = make(map[string]uuid.UUID)
	}
	if storyID, replayed := service.createdByKey[*story.CreationKey]; replayed {
		return stories.CoreSingleStory{ID: storyID, Workspace: workspaceID, CreatedNow: false}, nil
	}
	storyID := uuid.New()
	service.createdByKey[*story.CreationKey] = storyID
	return stories.CoreSingleStory{ID: storyID, Workspace: workspaceID, CreatedNow: true}, nil
}

type storyImportSessionResolverStub struct {
	userID uuid.UUID
}

func (resolver storyImportSessionResolverStub) Resolve(context.Context, *http.Request) (uuid.UUID, bool, error) {
	return resolver.userID, true, nil
}

type storyImportWorkspaceResolverStub struct {
	workspace mid.WorkspaceInfo
}

func (resolver storyImportWorkspaceResolverStub) ResolveCurrentWorkspace(
	context.Context,
	string,
	uuid.UUID,
) (mid.WorkspaceInfo, error) {
	return resolver.workspace, nil
}

func (storyImportWorkspaceResolverStub) RecordWorkspaceAccess(context.Context, uuid.UUID, uuid.UUID) error {
	return nil
}

type storyImportCacheStub struct {
	deletedPatterns []string
}

func (*storyImportCacheStub) Delete(context.Context, string) {}

func (cache *storyImportCacheStub) DeleteByPattern(_ context.Context, pattern string) {
	cache.deletedPatterns = append(cache.deletedPatterns, pattern)
}

func TestImportHandlerUsesAdminIndependentIdempotencyAndReportsReplay(t *testing.T) {
	t.Parallel()

	workspaceID, teamID := uuid.New(), uuid.New()
	service := &storyImportServiceStub{}
	handler := &Handlers{storyImporter: service}
	request := validStoryImportRequest(teamID, "JIRA-42", "Imported customer issue")
	refreshedExport := validStoryImportRequest(teamID, "JIRA-42", "Imported customer issue")
	refreshedExport.SourceDigest = strings.Repeat("b", sha256DigestHexLength)

	first := invokeStoryImportHandler(t, handler, workspaceID, uuid.New(), request)
	second := invokeStoryImportHandler(t, handler, workspaceID, uuid.New(), refreshedExport)

	if first.Counts.Created != 1 || first.Counts.Replayed != 0 || first.Counts.Failed != 0 || !first.Items[0].Created {
		t.Fatalf("first response = %#v", first)
	}
	if second.Counts.Created != 0 || second.Counts.Replayed != 1 || second.Counts.Failed != 0 || second.Items[0].Created {
		t.Fatalf("second response = %#v", second)
	}
	if first.Items[0].StoryID == nil || second.Items[0].StoryID == nil || *first.Items[0].StoryID != *second.Items[0].StoryID {
		t.Fatalf("story IDs differ: first=%v second=%v", first.Items[0].StoryID, second.Items[0].StoryID)
	}
	if len(service.calls) != 2 || service.calls[0].CreationKey == nil || service.calls[1].CreationKey == nil ||
		*service.calls[0].CreationKey != *service.calls[1].CreationKey {
		t.Fatalf("creation keys are not stable across admins: %#v", service.calls)
	}
	if service.calls[0].ExternalDelivery != storydomain.ExternalStoryDeliveryInternalOnly {
		t.Fatalf("external delivery = %v, want internal-only", service.calls[0].ExternalDelivery)
	}
	if service.calls[0].Reporter == nil || service.calls[1].Reporter == nil || *service.calls[0].Reporter == *service.calls[1].Reporter {
		t.Fatalf("expected distinct admin attribution without changing idempotency: %#v", service.calls)
	}
	if storyImportCreationKey(uuid.New(), teamID, request.Provider, request.SourceDigest, nil, request.Items[0].SourceKey) == *service.calls[0].CreationKey {
		t.Fatal("creation key is not workspace-scoped")
	}
}

func TestImportHandlerUsesSourceNamespaceAcrossChangedFileDigests(t *testing.T) {
	t.Parallel()

	workspaceID, teamID := uuid.New(), uuid.New()
	service := &storyImportServiceStub{}
	handler := &Handlers{storyImporter: service}
	firstExport := validStoryImportRequest(teamID, "card-42", "Imported Trello card")
	firstExport.Provider = storyImportProviderFile
	firstExport.SourceNamespace = stringPointer("trello:board:marketing")
	refreshedExport := firstExport
	refreshedExport.SourceDigest = strings.Repeat("b", sha256DigestHexLength)

	first := invokeStoryImportHandler(t, handler, workspaceID, uuid.New(), firstExport)
	second := invokeStoryImportHandler(t, handler, workspaceID, uuid.New(), refreshedExport)

	if first.Counts.Created != 1 || second.Counts.Replayed != 1 ||
		first.Items[0].StoryID == nil || second.Items[0].StoryID == nil {
		t.Fatalf("first=%#v second=%#v", first, second)
	}
	if *first.Items[0].StoryID != *second.Items[0].StoryID {
		t.Fatal("same source namespace and key did not replay across changed file digests")
	}
	if len(service.calls) != 2 || service.calls[0].CreationKey == nil || service.calls[1].CreationKey == nil ||
		*service.calls[0].CreationKey != *service.calls[1].CreationKey {
		t.Fatalf("source namespace did not provide a stable creation key: %#v", service.calls)
	}
}

func TestImportHandlerScopesSourceKeysToSourceNamespace(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		provider  string
		sourceKey string
	}{
		{name: "generic file", provider: storyImportProviderFile, sourceKey: "card-42"},
		{name: "Jira CSV", provider: storyImportProviderJiraCSV, sourceKey: "JIRA-42"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			workspaceID, teamID := uuid.New(), uuid.New()
			service := &storyImportServiceStub{}
			handler := &Handlers{storyImporter: service}
			firstSource := validStoryImportRequest(teamID, test.sourceKey, "First imported story")
			firstSource.Provider = test.provider
			firstSource.SourceNamespace = stringPointer("source:first")
			secondSource := firstSource
			secondSource.SourceNamespace = stringPointer("source:second")

			first := invokeStoryImportHandler(t, handler, workspaceID, uuid.New(), firstSource)
			second := invokeStoryImportHandler(t, handler, workspaceID, uuid.New(), secondSource)

			if first.Counts.Created != 1 || second.Counts.Created != 1 ||
				first.Items[0].StoryID == nil || second.Items[0].StoryID == nil {
				t.Fatalf("first=%#v second=%#v", first, second)
			}
			if *first.Items[0].StoryID == *second.Items[0].StoryID {
				t.Fatal("different source namespaces were deduplicated")
			}
			if len(service.calls) != 2 || service.calls[0].CreationKey == nil || service.calls[1].CreationKey == nil ||
				*service.calls[0].CreationKey == *service.calls[1].CreationKey {
				t.Fatalf("creation keys collided across source namespaces: %#v", service.calls)
			}
		})
	}
}

func TestGenericFileImportKeepsSourceDigestInIdempotencyScope(t *testing.T) {
	t.Parallel()

	workspaceID, teamID := uuid.New(), uuid.New()
	service := &storyImportServiceStub{}
	handler := &Handlers{storyImporter: service}
	firstFile := validStoryImportRequest(teamID, "row-42", "First file story")
	firstFile.Provider = storyImportProviderFile
	secondFile := validStoryImportRequest(teamID, "row-42", "Second file story")
	secondFile.Provider = storyImportProviderFile
	secondFile.SourceDigest = strings.Repeat("b", sha256DigestHexLength)

	first := invokeStoryImportHandler(t, handler, workspaceID, uuid.New(), firstFile)
	second := invokeStoryImportHandler(t, handler, workspaceID, uuid.New(), secondFile)

	if first.Counts.Created != 1 || second.Counts.Created != 1 || first.Items[0].StoryID == nil || second.Items[0].StoryID == nil {
		t.Fatalf("first=%#v second=%#v", first, second)
	}
	if *first.Items[0].StoryID == *second.Items[0].StoryID {
		t.Fatal("generic files with different source digests were deduplicated")
	}
	if len(service.calls) != 2 || service.calls[0].CreationKey == nil || service.calls[1].CreationKey == nil ||
		*service.calls[0].CreationKey == *service.calls[1].CreationKey {
		t.Fatalf("generic file creation keys did not retain digest scope: %#v", service.calls)
	}
}

func TestImportHandlerScopesBothProvidersToDestinationTeam(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		provider  string
		sourceKey string
	}{
		{name: "Jira CSV", provider: storyImportProviderJiraCSV, sourceKey: "TEAM-42"},
		{name: "generic file", provider: storyImportProviderFile, sourceKey: "row-42"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			workspaceID, firstTeamID, secondTeamID := uuid.New(), uuid.New(), uuid.New()
			service := &storyImportServiceStub{}
			handler := &Handlers{storyImporter: service}
			firstTeam := validStoryImportRequest(firstTeamID, test.sourceKey, "First team story")
			firstTeam.Provider = test.provider
			secondTeam := validStoryImportRequest(secondTeamID, test.sourceKey, "Second team story")
			secondTeam.Provider = test.provider

			first := invokeStoryImportHandler(t, handler, workspaceID, uuid.New(), firstTeam)
			replay := invokeStoryImportHandler(t, handler, workspaceID, uuid.New(), firstTeam)
			second := invokeStoryImportHandler(t, handler, workspaceID, uuid.New(), secondTeam)

			if first.Counts.Created != 1 || replay.Counts.Replayed != 1 || second.Counts.Created != 1 ||
				first.Items[0].StoryID == nil || replay.Items[0].StoryID == nil || second.Items[0].StoryID == nil {
				t.Fatalf("first=%#v replay=%#v second=%#v", first, replay, second)
			}
			if *first.Items[0].StoryID != *replay.Items[0].StoryID {
				t.Fatal("retry within the same destination team did not replay")
			}
			if *first.Items[0].StoryID == *second.Items[0].StoryID {
				t.Fatal("different destination teams were deduplicated")
			}
			if len(service.calls) != 3 || service.calls[0].CreationKey == nil || service.calls[1].CreationKey == nil ||
				service.calls[2].CreationKey == nil || *service.calls[0].CreationKey != *service.calls[1].CreationKey ||
				*service.calls[0].CreationKey == *service.calls[2].CreationKey {
				t.Fatalf("creation keys did not retain team scope: %#v", service.calls)
			}
		})
	}
}

func TestStoryImportRequestRequiresCanonicalJiraIssueSourceKeys(t *testing.T) {
	t.Parallel()

	teamID := uuid.New()
	for _, sourceKey := range []string{
		"row-1",
		"J-1",
		"1JIRA-1",
		"JIRA",
		"JIRA-0",
		"JIRA-01",
		"JIRA_OLD-1",
	} {
		sourceKey := sourceKey
		t.Run(sourceKey, func(t *testing.T) {
			t.Parallel()
			request := validStoryImportRequest(teamID, sourceKey, "Imported story")
			if err := request.Validate(); err == nil {
				t.Fatalf("Validate() accepted non-Jira source key %q", sourceKey)
			}
		})
	}

	for _, sourceKey := range []string{"AB-1", "R2D2-900"} {
		request := validStoryImportRequest(teamID, sourceKey, "Imported story")
		if err := request.Validate(); err != nil {
			t.Fatalf("Validate() rejected Jira source key %q: %v", sourceKey, err)
		}
	}

	genericFile := validStoryImportRequest(teamID, "row-1", "Imported story")
	genericFile.Provider = storyImportProviderFile
	if err := genericFile.Validate(); err != nil {
		t.Fatalf("Validate() rejected generic file source key: %v", err)
	}
}

func TestStoryImportRequestRejectsInvalidSourceNamespaces(t *testing.T) {
	t.Parallel()

	invalidUTF8 := string([]byte{0xff})
	tests := []struct {
		name            string
		sourceNamespace string
	}{
		{name: "empty", sourceNamespace: ""},
		{name: "leading whitespace", sourceNamespace: " trello:board:marketing"},
		{name: "trailing whitespace", sourceNamespace: "trello:board:marketing "},
		{name: "invalid UTF-8", sourceNamespace: invalidUTF8},
		{name: "control character", sourceNamespace: "trello:board:\nmarketing"},
		{name: "over byte limit", sourceNamespace: strings.Repeat("n", maximumImportSourceNamespaceBytes+1)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			request := validStoryImportRequest(uuid.New(), "JIRA-1", "Imported story")
			request.SourceNamespace = &test.sourceNamespace
			if err := request.Validate(); err == nil {
				t.Fatalf("Validate() accepted source namespace %q", test.sourceNamespace)
			}
		})
	}

	request := validStoryImportRequest(uuid.New(), "JIRA-1", "Imported story")
	request.SourceNamespace = stringPointer(strings.Repeat("n", maximumImportSourceNamespaceBytes))
	if err := request.Validate(); err != nil {
		t.Fatalf("Validate() rejected maximum-length source namespace: %v", err)
	}
}

func TestImportHandlerReturnsBoundedPerItemFailuresWithoutEchoingStoryContent(t *testing.T) {
	t.Parallel()

	workspaceID, teamID := uuid.New(), uuid.New()
	service := &storyImportServiceStub{
		failTitle: "sensitive imported title",
		failErr:   errors.New("sensitive downstream failure details"),
	}
	handler := &Handlers{storyImporter: service}
	request := validStoryImportRequest(teamID, "JIRA-99", service.failTitle)
	body, response := invokeStoryImportHandlerWithBody(t, handler, workspaceID, uuid.New(), request)

	if response.Counts.Total != 1 || response.Counts.Created != 0 || response.Counts.Replayed != 0 || response.Counts.Failed != 1 {
		t.Fatalf("counts = %#v", response.Counts)
	}
	result := response.Items[0]
	if result.StoryID != nil || result.Created || result.Error == nil || result.Error.Code != "internal_error" {
		t.Fatalf("result = %#v", result)
	}
	if strings.Contains(body, service.failTitle) || strings.Contains(body, service.failErr.Error()) || strings.Contains(body, request.SourceDigest) {
		t.Fatalf("response echoed sensitive source content: %s", body)
	}
}

func TestImportHandlerInvalidatesListCachesOnceForSuccessfulBatch(t *testing.T) {
	t.Parallel()

	workspaceID, teamID := uuid.New(), uuid.New()
	service := &storyImportServiceStub{}
	cacheSpy := &storyImportCacheStub{}
	handler := &Handlers{storyImporter: service, cache: cacheSpy}
	request := validStoryImportRequest(teamID, "JIRA-1", "First imported story")
	request.Items = append(request.Items, AppStoryImportItem{
		SourceKey: "JIRA-2",
		Story:     AppNewStory{Title: "Second imported story", Priority: "No Priority", Team: teamID},
	})

	response := invokeStoryImportHandler(t, handler, workspaceID, uuid.New(), request)

	if response.Counts.Created != 2 || response.Counts.Failed != 0 {
		t.Fatalf("response counts = %#v", response.Counts)
	}
	wantPatterns := []string{
		fmt.Sprintf(cache.StoryListKey+"*", workspaceID.String()),
		fmt.Sprintf(cache.MyStoriesKey+"*", workspaceID.String()),
	}
	if !slices.Equal(cacheSpy.deletedPatterns, wantPatterns) {
		t.Fatalf("cache invalidations = %#v, want %#v", cacheSpy.deletedPatterns, wantPatterns)
	}
}

func TestImportHandlerRejectsMoreThanFiftyItems(t *testing.T) {
	t.Parallel()

	workspaceID, teamID := uuid.New(), uuid.New()
	service := &storyImportServiceStub{}
	handler := &Handlers{storyImporter: service}
	request := validStoryImportRequest(teamID, "JIRA-1", "Imported story")
	request.Items = make([]AppStoryImportItem, maximumStoryImportItems+1)
	for index := range request.Items {
		request.Items[index] = AppStoryImportItem{
			SourceKey: fmt.Sprintf("JIRA-%d", index+1),
			Story:     AppNewStory{Title: "Imported story", Priority: "No Priority", Team: teamID},
		}
	}

	recorder := recordStoryImportHandler(t, handler, workspaceID, uuid.New(), request)
	if recorder.Code != http.StatusBadRequest || len(service.calls) != 0 {
		t.Fatalf("status=%d calls=%d body=%s", recorder.Code, len(service.calls), recorder.Body.String())
	}
}

func TestImportHandlerInvokesCustomRequestValidation(t *testing.T) {
	t.Parallel()

	workspaceID, teamID := uuid.New(), uuid.New()
	service := &storyImportServiceStub{}
	handler := &Handlers{storyImporter: service}
	request := validStoryImportRequest(teamID, "JIRA-1", "Imported story")
	// This satisfies the tag-level length constraint. Only
	// AppStoryImportRequest.Validate rejects the non-hexadecimal digest.
	request.SourceDigest = strings.Repeat("z", sha256DigestHexLength)

	recorder := recordStoryImportHandler(t, handler, workspaceID, uuid.New(), request)

	if recorder.Code != http.StatusBadRequest || len(service.calls) != 0 {
		t.Fatalf("status=%d calls=%d body=%s", recorder.Code, len(service.calls), recorder.Body.String())
	}
}

func TestImportHandlerRejectsInvalidSourceNamespace(t *testing.T) {
	t.Parallel()

	workspaceID, teamID := uuid.New(), uuid.New()
	service := &storyImportServiceStub{}
	handler := &Handlers{storyImporter: service}
	request := validStoryImportRequest(teamID, "JIRA-1", "Imported story")
	request.SourceNamespace = stringPointer(" source:with-leading-space")

	recorder := recordStoryImportHandler(t, handler, workspaceID, uuid.New(), request)

	if recorder.Code != http.StatusBadRequest || len(service.calls) != 0 {
		t.Fatalf("status=%d calls=%d body=%s", recorder.Code, len(service.calls), recorder.Body.String())
	}
}

func TestImportHandlerRejectsGenericSourceKeyForJiraCSV(t *testing.T) {
	t.Parallel()

	workspaceID, teamID := uuid.New(), uuid.New()
	service := &storyImportServiceStub{}
	handler := &Handlers{storyImporter: service}
	request := validStoryImportRequest(teamID, "row-1", "Imported story")

	recorder := recordStoryImportHandler(t, handler, workspaceID, uuid.New(), request)

	if recorder.Code != http.StatusBadRequest || len(service.calls) != 0 {
		t.Fatalf("status=%d calls=%d body=%s", recorder.Code, len(service.calls), recorder.Body.String())
	}
}

func TestStoryImportRouteRequiresWorkspaceAdmin(t *testing.T) {
	t.Parallel()

	log := logger.NewWithText(io.Discard, slog.LevelError, "test")
	userID, workspaceID := uuid.New(), uuid.New()
	app := web.New(make(chan os.Signal, 1), nil)
	Routes(Config{
		Log:             log,
		BrowserSessions: storyImportSessionResolverStub{userID: userID},
		WorkspaceResolver: storyImportWorkspaceResolverStub{workspace: mid.WorkspaceInfo{
			ID: workspaceID, Slug: "acme", UserRole: string(mid.RoleMember),
		}},
	}, app)
	payload, err := json.Marshal(validStoryImportRequest(uuid.New(), "JIRA-1", "Imported story"))
	if err != nil {
		t.Fatalf("encode request: %v", err)
	}
	request := httptest.NewRequest(http.MethodPost, "/workspaces/acme/stories/import", strings.NewReader(string(payload)))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	app.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d; body=%s", recorder.Code, http.StatusForbidden, recorder.Body.String())
	}
}

func validStoryImportRequest(teamID uuid.UUID, sourceKey, title string) AppStoryImportRequest {
	return AppStoryImportRequest{
		Provider:     "jira_csv",
		SourceDigest: strings.Repeat("a", 64),
		Items: []AppStoryImportItem{{
			SourceKey: sourceKey,
			Story: AppNewStory{
				Title: title, Priority: "No Priority", Team: teamID,
			},
		}},
	}
}

func stringPointer(value string) *string {
	return &value
}

func invokeStoryImportHandler(
	t *testing.T,
	handler *Handlers,
	workspaceID, actorID uuid.UUID,
	request AppStoryImportRequest,
) AppStoryImportResponse {
	t.Helper()
	_, response := invokeStoryImportHandlerWithBody(t, handler, workspaceID, actorID, request)
	return response
}

func invokeStoryImportHandlerWithBody(
	t *testing.T,
	handler *Handlers,
	workspaceID, actorID uuid.UUID,
	request AppStoryImportRequest,
) (string, AppStoryImportResponse) {
	t.Helper()
	recorder := recordStoryImportHandler(t, handler, workspaceID, actorID, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
	if cacheControl := recorder.Header().Get("Cache-Control"); cacheControl != "private, no-store" {
		t.Fatalf("cache-control = %q", cacheControl)
	}
	var envelope struct {
		Data AppStoryImportResponse `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode response: %v; body=%s", err, recorder.Body.String())
	}
	return recorder.Body.String(), envelope.Data
}

func recordStoryImportHandler(
	t *testing.T,
	handler *Handlers,
	workspaceID, actorID uuid.UUID,
	request AppStoryImportRequest,
) *httptest.ResponseRecorder {
	t.Helper()
	payload, err := json.Marshal(request)
	if err != nil {
		t.Fatalf("encode request: %v", err)
	}
	httpRequest := httptest.NewRequest(http.MethodPost, "/workspaces/import/stories/import", strings.NewReader(string(payload)))
	httpRequest.Header.Set("Content-Type", "application/json")
	httpRequest.SetPathValue("workspaceSlug", "import")
	recorder := httptest.NewRecorder()
	resolver := storyImportWorkspaceResolverStub{workspace: mid.WorkspaceInfo{
		ID: workspaceID, Slug: "import", UserRole: string(mid.RoleAdmin),
	}}
	wrapped := mid.Workspace(nil, resolver)(handler.Import)
	if err := wrapped(platformauth.SetUserID(t.Context(), actorID), recorder, httpRequest); err != nil {
		t.Fatalf("invoke import handler: %v", err)
	}
	return recorder
}
