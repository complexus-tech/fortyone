package apiv1http

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/complexus-tech/projects-api/internal/modules/developeraccess"
	developeroauthdomain "github.com/complexus-tech/projects-api/internal/modules/developeroauth/domain"
	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	workspaces "github.com/complexus-tech/projects-api/internal/modules/workspaces/service"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/complexus-tech/projects-api/internal/platform/idempotency"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace/noop"
)

const testAPIOAuthBearer = "header.api-audience.signature"

type machineCredentialMiss struct {
	calls int
}

func (stub *machineCredentialMiss) ResolveMachineCredential(context.Context, string) (platformauth.Actor, error) {
	stub.calls++
	return platformauth.Actor{}, errors.New("not a machine credential")
}

type apiAudienceOAuthVerifier struct {
	identity developeroauthdomain.AccessIdentity
	calls    int
}

func (stub *apiAudienceOAuthVerifier) VerifyAccessToken(_ context.Context, raw string) (developeroauthdomain.AccessIdentity, error) {
	stub.calls++
	if raw != testAPIOAuthBearer {
		return developeroauthdomain.AccessIdentity{}, errors.New("invalid API access token")
	}
	return stub.identity, nil
}

type oauthWorkspaceLookup struct {
	workspaceID uuid.UUID
	userID      uuid.UUID
	actor       platformauth.Actor
	actorErr    error
}

type oauthWorkspaceReader struct {
	memberships map[uuid.UUID]workspaces.CoreWorkspace
	calls       []oauthWorkspaceLookup
}

func (reader *oauthWorkspaceReader) Get(ctx context.Context, workspaceID, userID uuid.UUID) (workspaces.CoreWorkspace, error) {
	actor, actorErr := platformauth.GetActor(ctx)
	reader.calls = append(reader.calls, oauthWorkspaceLookup{
		workspaceID: workspaceID,
		userID:      userID,
		actor:       actor,
		actorErr:    actorErr,
	})
	workspace, ok := reader.memberships[workspaceID]
	if !ok {
		return workspaces.CoreWorkspace{}, workspaces.ErrNotFound
	}
	return workspace, nil
}

type oauthStoryServiceRecorder struct {
	created stories.CoreSingleStory
	input   stories.CoreNewStory
	actor   platformauth.Actor
	get     int
	list    int
	create  int
}

func (service *oauthStoryServiceRecorder) Get(context.Context, uuid.UUID, uuid.UUID) (stories.CoreSingleStory, error) {
	service.get++
	return stories.CoreSingleStory{}, stories.ErrNotFound
}

func (service *oauthStoryServiceRecorder) List(ctx context.Context, _ uuid.UUID, _ stories.CoreStoryFilters) ([]stories.CoreStoryList, error) {
	service.list++
	service.actor, _ = platformauth.GetActor(ctx)
	return []stories.CoreStoryList{}, nil
}

func (service *oauthStoryServiceRecorder) Create(ctx context.Context, input stories.CoreNewStory, _ uuid.UUID) (stories.CoreSingleStory, error) {
	service.create++
	service.input = input
	service.actor, _ = platformauth.GetActor(ctx)
	return service.created, nil
}

func (service *oauthStoryServiceRecorder) calls() int {
	return service.get + service.list + service.create
}

type oauthIdempotencyRecorder struct {
	beginResult idempotency.BeginResult
	scope       idempotency.Scope
	body        []byte
	actor       platformauth.Actor
	beginCalls  int
	completions int
}

func (recorder *oauthIdempotencyRecorder) Begin(ctx context.Context, scope idempotency.Scope, _ idempotency.Key, body []byte) (idempotency.BeginResult, error) {
	recorder.beginCalls++
	recorder.scope = scope
	recorder.body = append([]byte(nil), body...)
	recorder.actor, _ = platformauth.GetActor(ctx)
	return recorder.beginResult, nil
}

func (recorder *oauthIdempotencyRecorder) Complete(context.Context, idempotency.Lease, idempotency.Response) error {
	recorder.completions++
	return nil
}

func TestPublicAPIOAuthReadRechecksMembershipBeforeBindingWorkspace(t *testing.T) {
	t.Parallel()

	workspaceID, userID, grantID := uuid.New(), uuid.New(), uuid.New()
	identity := apiOAuthIdentity(userID, grantID, platformauth.ScopeWorkspacesRead)
	resolver, machine, oauth := newAPIOAuthResolver(t, identity)
	workspaceReader := &oauthWorkspaceReader{memberships: map[uuid.UUID]workspaces.CoreWorkspace{
		workspaceID: {
			ID: workspaceID, Slug: "product", Name: "Product", Color: "#123456", IsActive: true,
			UserRole: "member", CreatedAt: time.Date(2026, 8, 1, 10, 0, 0, 0, time.UTC),
			UpdatedAt: time.Date(2026, 8, 2, 10, 0, 0, 0, time.UTC),
		},
	}}
	store := &rateLimitStoreStub{count: 1}
	app := oauthPublicAPITestApp(t, resolver, store, workspaceReader, &oauthStoryServiceRecorder{}, &oauthIdempotencyRecorder{})
	request := httptest.NewRequest(http.MethodGet, "/api/v1/workspaces/"+workspaceID.String(), nil)
	request.Header.Set("Authorization", "Bearer "+testAPIOAuthBearer)
	recorder := httptest.NewRecorder()

	app.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusOK, recorder.Code, recorder.Body.String())
	require.Equal(t, 1, machine.calls)
	require.Equal(t, 1, oauth.calls)
	require.Len(t, workspaceReader.calls, 2, "membership must be checked before the resource read")

	beforeBinding := workspaceReader.calls[0]
	require.NoError(t, beforeBinding.actorErr)
	require.Equal(t, workspaceID, beforeBinding.workspaceID)
	require.Equal(t, userID, beforeBinding.userID)
	require.Equal(t, platformauth.PrincipalOAuthUser, beforeBinding.actor.Kind)
	require.Equal(t, userID, beforeBinding.actor.PrincipalID)
	require.Equal(t, grantID, beforeBinding.actor.CredentialID)
	require.Equal(t, uuid.Nil, beforeBinding.actor.WorkspaceID, "OAuth authentication must not preselect a workspace")

	authorizedRead := workspaceReader.calls[1]
	require.NoError(t, authorizedRead.actorErr)
	require.Equal(t, workspaceID, authorizedRead.actor.WorkspaceID, "binding may happen only after the membership lookup succeeds")
	require.Equal(t, userID, authorizedRead.actor.PrincipalID)
	require.Equal(t, grantID, authorizedRead.actor.CredentialID)
	require.Equal(t, []string{"rate-limit:api-v1:credential:" + grantID.String()}, store.keys)
	require.NotContains(t, store.keys[0], testAPIOAuthBearer)
}

func TestPublicAPIOAuthRejectsMissingOrCrossWorkspaceMembershipBeforeServices(t *testing.T) {
	t.Parallel()

	memberWorkspaceID := uuid.New()
	for _, test := range []struct {
		name        string
		workspaceID uuid.UUID
		memberships map[uuid.UUID]workspaces.CoreWorkspace
	}{
		{
			name:        "membership was removed",
			workspaceID: memberWorkspaceID,
			memberships: map[uuid.UUID]workspaces.CoreWorkspace{},
		},
		{
			name:        "requested another workspace",
			workspaceID: uuid.New(),
			memberships: map[uuid.UUID]workspaces.CoreWorkspace{
				memberWorkspaceID: {ID: memberWorkspaceID, UserRole: "member", IsActive: true},
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			userID, grantID := uuid.New(), uuid.New()
			resolver, _, _ := newAPIOAuthResolver(t, apiOAuthIdentity(userID, grantID, platformauth.ScopeStoriesRead))
			workspaceReader := &oauthWorkspaceReader{memberships: test.memberships}
			stories := &oauthStoryServiceRecorder{}
			receipts := &oauthIdempotencyRecorder{}
			webhooks := &webhookManagerStub{}
			app := oauthPublicAPITestAppWithWebhooks(
				t,
				resolver,
				&rateLimitStoreStub{count: 1},
				workspaceReader,
				stories,
				receipts,
				webhooks,
			)
			request := httptest.NewRequest(http.MethodGet, "/api/v1/workspaces/"+test.workspaceID.String()+"/stories", nil)
			request.Header.Set("Authorization", "Bearer "+testAPIOAuthBearer)
			recorder := httptest.NewRecorder()

			app.ServeHTTP(recorder, request)

			require.Equal(t, http.StatusForbidden, recorder.Code, recorder.Body.String())
			require.Contains(t, recorder.Body.String(), `"code":"access_denied"`)
			require.Len(t, workspaceReader.calls, 1)
			require.NoError(t, workspaceReader.calls[0].actorErr)
			require.Equal(t, uuid.Nil, workspaceReader.calls[0].actor.WorkspaceID)
			require.Equal(t, test.workspaceID, workspaceReader.calls[0].workspaceID)
			require.Equal(t, userID, workspaceReader.calls[0].userID)
			require.Zero(t, stories.calls(), "resource services must not run after membership denial")
			require.Zero(t, receipts.beginCalls)
			require.Zero(t, webhooks.calls)
		})
	}
}

func TestPublicAPIOAuthStoryWritePreservesUserAndIdempotencyIdentity(t *testing.T) {
	t.Parallel()

	workspaceID, teamID, storyID := uuid.New(), uuid.New(), uuid.New()
	userID, grantID := uuid.New(), uuid.New()
	identity := apiOAuthIdentity(userID, grantID, platformauth.ScopeStoriesWrite)
	resolver, _, _ := newAPIOAuthResolver(t, identity)
	workspaceReader := &oauthWorkspaceReader{memberships: map[uuid.UUID]workspaces.CoreWorkspace{
		workspaceID: {ID: workspaceID, UserRole: "member", IsActive: true},
	}}
	storyService := &oauthStoryServiceRecorder{created: stories.CoreSingleStory{
		ID: storyID, Workspace: workspaceID, Team: teamID, TeamCode: "ENG", SequenceID: 41,
		Title: "OAuth-created story", Priority: "No Priority", AutoSchedulingStatus: stories.AutoSchedulingStatusOff,
		CreatedAt: time.Date(2026, 8, 28, 12, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 8, 28, 12, 0, 0, 0, time.UTC),
	}}
	lease := idempotency.Lease{ReceiptID: uuid.New(), Generation: 1, ExpiresAt: time.Now().Add(time.Minute)}
	receipts := &oauthIdempotencyRecorder{beginResult: idempotency.BeginResult{
		State: idempotency.BeginStateNew,
		Lease: lease,
	}}
	store := &rateLimitStoreStub{count: 1}
	app := oauthPublicAPITestApp(t, resolver, store, workspaceReader, storyService, receipts)
	rawKey := "oauth-story-create-key-0001"
	rawBody := []byte(`{"title":"OAuth-created story","teamId":"` + teamID.String() + `"}`)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/workspaces/"+workspaceID.String()+"/stories", bytes.NewReader(rawBody))
	request.Header.Set("Authorization", "Bearer "+testAPIOAuthBearer)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Idempotency-Key", rawKey)
	recorder := httptest.NewRecorder()

	app.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusCreated, recorder.Code, recorder.Body.String())
	require.Len(t, workspaceReader.calls, 1)
	require.Equal(t, uuid.Nil, workspaceReader.calls[0].actor.WorkspaceID)
	require.Equal(t, 1, storyService.create)
	require.Equal(t, platformauth.PrincipalOAuthUser, storyService.actor.Kind)
	require.Equal(t, userID, storyService.actor.PrincipalID)
	require.Equal(t, grantID, storyService.actor.CredentialID)
	require.Equal(t, workspaceID, storyService.actor.WorkspaceID)
	userActorID, err := storyService.actor.UserID()
	require.NoError(t, err)
	require.Equal(t, userID, userActorID, "the story service must receive the consenting user as its actor")
	require.NotNil(t, storyService.input.CreationKey)
	require.True(t, strings.HasPrefix(*storyService.input.CreationKey, "api-v1:oauth_user:"+userID.String()+":"))
	require.NotContains(t, *storyService.input.CreationKey, rawKey)

	require.Equal(t, 1, receipts.beginCalls)
	require.Equal(t, 1, receipts.completions)
	require.Equal(t, rawBody, receipts.body)
	require.Equal(t, storyService.actor, receipts.actor)
	operation, err := idempotency.ParseOperation("stories.create")
	require.NoError(t, err)
	expectedScope, err := idempotency.NewScope(storyService.actor, idempotency.MethodPost, operation)
	require.NoError(t, err)
	require.Equal(t, expectedScope, receipts.scope, "receipt scope must retain OAuth user, workspace, method, and operation identity")
	require.Equal(t, []string{"rate-limit:api-v1:credential:" + grantID.String()}, store.keys)
}

func TestPublicAPIOAuthApplicationStoryWriteUsesInstallationIdentityWithoutInstaller(t *testing.T) {
	t.Parallel()

	workspaceID, teamID, storyID := uuid.New(), uuid.New(), uuid.New()
	principalID, installationID := uuid.New(), uuid.New()
	identity := developeroauthdomain.AccessIdentity{
		PrincipalID: principalID, ApplicationID: uuid.New(), InstallationID: installationID,
		WorkspaceID: workspaceID, ActorKind: platformauth.PrincipalOAuthApplication,
		ClientID: "managed-application", Resource: "https://api.fortyone.app/api/v1",
		Scopes:    []string{string(platformauth.ScopeStoriesWrite)},
		ExpiresAt: time.Now().Add(time.Minute), OAuthCredential: uuid.New(),
	}
	resolver, _, _ := newAPIOAuthResolver(t, identity)
	workspaceReader := &oauthWorkspaceReader{memberships: map[uuid.UUID]workspaces.CoreWorkspace{}}
	storyService := &oauthStoryServiceRecorder{created: stories.CoreSingleStory{
		ID: storyID, Workspace: workspaceID, Team: teamID, TeamCode: "ENG", SequenceID: 42,
		Title: "Application-created story", Priority: "No Priority", AutoSchedulingStatus: stories.AutoSchedulingStatusOff,
		CreatedAt: time.Date(2026, 8, 28, 13, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 8, 28, 13, 0, 0, 0, time.UTC),
	}}
	receipts := &oauthIdempotencyRecorder{beginResult: idempotency.BeginResult{
		State: idempotency.BeginStateNew,
		Lease: idempotency.Lease{ReceiptID: uuid.New(), Generation: 1, ExpiresAt: time.Now().Add(time.Minute)},
	}}
	store := &rateLimitStoreStub{count: 1}
	app := oauthPublicAPITestApp(t, resolver, store, workspaceReader, storyService, receipts)
	rawKey := "oauth-application-story-key-0001"
	rawBody := []byte(`{"title":"Application-created story","teamId":"` + teamID.String() + `"}`)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/workspaces/"+workspaceID.String()+"/stories", bytes.NewReader(rawBody))
	request.Header.Set("Authorization", "Bearer "+testAPIOAuthBearer)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Idempotency-Key", rawKey)
	recorder := httptest.NewRecorder()

	app.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusCreated, recorder.Code, recorder.Body.String())
	require.Empty(t, workspaceReader.calls, "an installation principal must not be checked as the installer or a workspace member")
	require.Equal(t, 1, storyService.create)
	require.Equal(t, platformauth.PrincipalOAuthApplication, storyService.actor.Kind)
	require.Equal(t, principalID, storyService.actor.PrincipalID)
	require.Equal(t, installationID, storyService.actor.CredentialID)
	require.Equal(t, workspaceID, storyService.actor.WorkspaceID)
	_, err := storyService.actor.UserID()
	require.Error(t, err, "the application action must never be attributed to its installer")
	require.Nil(t, storyService.input.Reporter)
	require.NotNil(t, storyService.input.CreationKey)
	require.True(t, strings.HasPrefix(*storyService.input.CreationKey, "api-v1:oauth_application:"+installationID.String()+":"))
	require.NotContains(t, *storyService.input.CreationKey, principalID.String())
	require.Equal(t, installationID, receipts.actor.CredentialID)
	operation, err := idempotency.ParseOperation("stories.create")
	require.NoError(t, err)
	expectedActor := storyService.actor
	expectedActor.PrincipalID = installationID
	expectedScope, err := idempotency.NewScope(expectedActor, idempotency.MethodPost, operation)
	require.NoError(t, err)
	require.Equal(t, expectedScope, receipts.scope, "receipt scope must use the stable installation identity")
	require.Equal(t, []string{"rate-limit:api-v1:credential:" + installationID.String()}, store.keys)
}

func TestPublicAPIOAuthApplicationCannotCrossItsInstalledWorkspace(t *testing.T) {
	t.Parallel()

	installedWorkspaceID, requestedWorkspaceID := uuid.New(), uuid.New()
	identity := developeroauthdomain.AccessIdentity{
		PrincipalID: uuid.New(), ApplicationID: uuid.New(), InstallationID: uuid.New(),
		WorkspaceID: installedWorkspaceID, ActorKind: platformauth.PrincipalOAuthApplication,
		ClientID: "managed-application", Resource: "https://api.fortyone.app/api/v1",
		Scopes:    []string{string(platformauth.ScopeStoriesWrite)},
		ExpiresAt: time.Now().Add(time.Minute), OAuthCredential: uuid.New(),
	}
	resolver, _, _ := newAPIOAuthResolver(t, identity)
	workspaceReader := &oauthWorkspaceReader{memberships: map[uuid.UUID]workspaces.CoreWorkspace{}}
	storyService := &oauthStoryServiceRecorder{}
	receipts := &oauthIdempotencyRecorder{}
	app := oauthPublicAPITestApp(t, resolver, &rateLimitStoreStub{count: 1}, workspaceReader, storyService, receipts)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/workspaces/"+requestedWorkspaceID.String()+"/stories", strings.NewReader(`{"title":"Wrong tenant","teamId":"`+uuid.NewString()+`"}`))
	request.Header.Set("Authorization", "Bearer "+testAPIOAuthBearer)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Idempotency-Key", "oauth-application-cross-tenant-0001")
	recorder := httptest.NewRecorder()

	app.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusForbidden, recorder.Code, recorder.Body.String())
	require.Empty(t, workspaceReader.calls)
	require.Zero(t, storyService.calls())
	require.Zero(t, receipts.beginCalls)
}

func TestPublicAPIOAuthApplicationCannotBorrowUserReadOrWebhookPolicies(t *testing.T) {
	t.Parallel()

	workspaceID, storyID := uuid.New(), uuid.New()
	identity := developeroauthdomain.AccessIdentity{
		PrincipalID: uuid.New(), ApplicationID: uuid.New(), InstallationID: uuid.New(),
		WorkspaceID: workspaceID, ActorKind: platformauth.PrincipalOAuthApplication,
		ClientID: "managed-application", Resource: "https://api.fortyone.app/api/v1",
		Scopes: []string{
			string(platformauth.ScopeWorkspacesRead), string(platformauth.ScopeStoriesRead),
			string(platformauth.ScopeCommentsRead), string(platformauth.ScopeWebhooksManage),
		},
		ExpiresAt: time.Now().Add(time.Minute), OAuthCredential: uuid.New(),
	}

	for _, test := range []struct {
		name string
		path string
	}{
		{name: "workspace membership read", path: "/api/v1/workspaces/" + workspaceID.String()},
		{name: "story read", path: "/api/v1/workspaces/" + workspaceID.String() + "/stories/" + storyID.String()},
		{name: "comment read", path: "/api/v1/workspaces/" + workspaceID.String() + "/stories/" + storyID.String() + "/comments"},
		{name: "webhook administration", path: "/api/v1/workspaces/" + workspaceID.String() + "/webhook-endpoints"},
	} {
		t.Run(test.name, func(t *testing.T) {
			resolver, _, _ := newAPIOAuthResolver(t, identity)
			workspaceReader := &oauthWorkspaceReader{memberships: map[uuid.UUID]workspaces.CoreWorkspace{}}
			storyService := &oauthStoryServiceRecorder{}
			webhooks := &webhookManagerStub{}
			app := oauthPublicAPITestAppWithWebhooks(
				t, resolver, &rateLimitStoreStub{count: 1}, workspaceReader,
				storyService, &oauthIdempotencyRecorder{}, webhooks,
			)
			request := httptest.NewRequest(http.MethodGet, test.path, nil)
			request.Header.Set("Authorization", "Bearer "+testAPIOAuthBearer)
			recorder := httptest.NewRecorder()

			app.ServeHTTP(recorder, request)

			require.Equal(t, http.StatusForbidden, recorder.Code, recorder.Body.String())
			require.Contains(t, recorder.Body.String(), `"code":"principal_not_supported"`)
			require.Empty(t, workspaceReader.calls, "application actors must never resolve through installer membership")
			require.Zero(t, storyService.calls())
			require.Zero(t, webhooks.calls)
		})
	}
}

func apiOAuthIdentity(userID, grantID uuid.UUID, scopes ...platformauth.Scope) developeroauthdomain.AccessIdentity {
	rawScopes := make([]string, 0, len(scopes)+1)
	rawScopes = append(rawScopes, developeroauthdomain.ScopeOfflineAccess)
	for _, scope := range scopes {
		rawScopes = append(rawScopes, string(scope))
	}
	return developeroauthdomain.AccessIdentity{
		PrincipalID: userID, UserID: userID, ApplicationID: uuid.New(), GrantID: grantID,
		ActorKind: platformauth.PrincipalOAuthUser,
		ClientID:  "api-oauth-test-client",
		Resource:  "https://api.fortyone.app/api/v1",
		Scopes:    rawScopes,
		ExpiresAt: time.Now().Add(time.Minute), OAuthCredential: uuid.New(),
	}
}

func newAPIOAuthResolver(
	t *testing.T,
	identity developeroauthdomain.AccessIdentity,
) (*developeraccess.Resolver, *machineCredentialMiss, *apiAudienceOAuthVerifier) {
	t.Helper()
	machine := &machineCredentialMiss{}
	oauth := &apiAudienceOAuthVerifier{identity: identity}
	resolver, err := developeraccess.NewResolver(machine, oauth)
	require.NoError(t, err)
	return resolver, machine, oauth
}

func oauthPublicAPITestApp(
	t *testing.T,
	resolver *developeraccess.Resolver,
	store *rateLimitStoreStub,
	workspaces *oauthWorkspaceReader,
	stories *oauthStoryServiceRecorder,
	receipts *oauthIdempotencyRecorder,
) *web.App {
	t.Helper()
	return oauthPublicAPITestAppWithWebhooks(t, resolver, store, workspaces, stories, receipts, &webhookManagerStub{})
}

func oauthPublicAPITestAppWithWebhooks(
	t *testing.T,
	resolver *developeraccess.Resolver,
	store *rateLimitStoreStub,
	workspaces *oauthWorkspaceReader,
	stories *oauthStoryServiceRecorder,
	receipts *oauthIdempotencyRecorder,
	webhooks *webhookManagerStub,
) *web.App {
	t.Helper()
	log := logger.NewWithJSON(&bytes.Buffer{}, slog.LevelError, strings.ReplaceAll(t.Name(), "/", "-"))
	app := web.New(make(chan os.Signal, 1), noop.NewTracerProvider().Tracer("public-api-oauth-test"))
	app.SetLogger(log)
	Routes(Config{
		Log: log, SecretKey: strings.Repeat("test-public-api-secret", 2), Cache: store,
		DeveloperCredentials: resolver, Workspaces: workspaces, Teams: teamReaderStub{},
		Stories: stories, StoryComments: storyCommentReaderStub{}, Labels: labelReaderStub{},
		States: workflowStateReaderStub{}, Sprints: sprintReaderStub{}, Objectives: objectiveReaderStub{},
		KeyResults: keyResultReaderStub{}, Idempotency: receipts, Webhooks: webhooks,
	}, app)
	return app
}
