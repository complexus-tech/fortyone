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

	outboundwebhooksdomain "github.com/complexus-tech/projects-api/internal/modules/outboundwebhooks/domain"
	outboundwebhooksservice "github.com/complexus-tech/projects-api/internal/modules/outboundwebhooks/service"
	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	teams "github.com/complexus-tech/projects-api/internal/modules/teams/service"
	workspaces "github.com/complexus-tech/projects-api/internal/modules/workspaces/service"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/complexus-tech/projects-api/internal/platform/idempotency"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace/noop"
)

type credentialResolverStub struct {
	actor platformauth.Actor
	err   error
	calls int
}

func (resolver *credentialResolverStub) ResolveDeveloperCredential(context.Context, string) (platformauth.Actor, error) {
	resolver.calls++
	return resolver.actor, resolver.err
}

type rateLimitStoreStub struct {
	count int64
	err   error
	keys  []string
}

func (store *rateLimitStoreStub) IncrementWithTTL(_ context.Context, key string, _ time.Duration) (int64, error) {
	store.keys = append(store.keys, key)
	return store.count, store.err
}

type workspaceReaderStub struct {
	workspace workspaces.CoreWorkspace
	err       error
	calls     int
}

func (reader *workspaceReaderStub) Get(context.Context, uuid.UUID, uuid.UUID) (workspaces.CoreWorkspace, error) {
	reader.calls++
	return reader.workspace, reader.err
}

type teamReaderStub struct{}

func (teamReaderStub) List(context.Context, uuid.UUID, uuid.UUID, ...teams.CoreListTeamsFilter) ([]teams.CoreTeam, error) {
	return []teams.CoreTeam{}, nil
}

func (teamReaderStub) GetByID(_ context.Context, teamID, workspaceID, _ uuid.UUID) (teams.CoreTeam, error) {
	return teams.CoreTeam{ID: teamID, Workspace: workspaceID}, nil
}

type storyReaderStub struct{}

func (storyReaderStub) Get(context.Context, uuid.UUID, uuid.UUID) (stories.CoreSingleStory, error) {
	return stories.CoreSingleStory{}, stories.ErrNotFound
}

func (storyReaderStub) List(context.Context, uuid.UUID, stories.CoreStoryFilters) ([]stories.CoreStoryList, error) {
	return []stories.CoreStoryList{}, nil
}

func (storyReaderStub) Create(context.Context, stories.CoreNewStory, uuid.UUID) (stories.CoreSingleStory, error) {
	return stories.CoreSingleStory{}, errors.New("unexpected story create")
}

type idempotencyManagerStub struct{}

func (idempotencyManagerStub) Begin(context.Context, idempotency.Scope, idempotency.Key, []byte) (idempotency.BeginResult, error) {
	return idempotency.BeginResult{State: idempotency.BeginStateNew, Lease: idempotency.Lease{
		ReceiptID: uuid.New(), Generation: 1, ExpiresAt: time.Now().Add(time.Minute),
	}}, nil
}

func (idempotencyManagerStub) Complete(context.Context, idempotency.Lease, idempotency.Response) error {
	return nil
}

type webhookManagerStub struct {
	calls       int
	created     outboundwebhooksdomain.CreatedEndpoint
	endpoint    outboundwebhooksdomain.Endpoint
	page        outboundwebhooksdomain.EndpointPage
	rotated     outboundwebhooksdomain.SigningSecret
	generation  int
	overlapEnd  time.Time
	createInput outboundwebhooksservice.CreateEndpointInput
	access      outboundwebhooksservice.Access
	operations  []string
}

func (manager *webhookManagerStub) CreateEndpoint(_ context.Context, access outboundwebhooksservice.Access, input outboundwebhooksservice.CreateEndpointInput) (outboundwebhooksdomain.CreatedEndpoint, error) {
	manager.calls++
	manager.operations = append(manager.operations, "create")
	manager.access = access
	manager.createInput = input
	if manager.created.Endpoint.ID != uuid.Nil {
		return manager.created, nil
	}
	return outboundwebhooksdomain.CreatedEndpoint{}, errors.New("unexpected webhook call")
}
func (manager *webhookManagerStub) GetEndpoint(context.Context, outboundwebhooksservice.Access, uuid.UUID) (outboundwebhooksdomain.Endpoint, error) {
	manager.calls++
	manager.operations = append(manager.operations, "get")
	if manager.endpoint.ID == uuid.Nil {
		return outboundwebhooksdomain.Endpoint{}, outboundwebhooksdomain.ErrEndpointNotFound
	}
	return manager.endpoint, nil
}
func (manager *webhookManagerStub) ListEndpoints(context.Context, outboundwebhooksservice.Access, *outboundwebhooksdomain.EndpointCursor, int) (outboundwebhooksdomain.EndpointPage, error) {
	manager.calls++
	manager.operations = append(manager.operations, "list")
	return manager.page, nil
}
func (manager *webhookManagerStub) ReplaceSubscriptions(context.Context, outboundwebhooksservice.Access, uuid.UUID, []outboundwebhooksdomain.EventType, string) error {
	manager.calls++
	manager.operations = append(manager.operations, "subscriptions")
	return nil
}
func (manager *webhookManagerStub) DisableEndpoint(context.Context, outboundwebhooksservice.Access, uuid.UUID, string, string) error {
	manager.calls++
	manager.operations = append(manager.operations, "disable")
	return nil
}
func (manager *webhookManagerStub) RotateEndpointSecret(context.Context, outboundwebhooksservice.Access, uuid.UUID, string) (outboundwebhooksdomain.SigningSecret, int, time.Time, error) {
	manager.calls++
	manager.operations = append(manager.operations, "rotate")
	return manager.rotated, manager.generation, manager.overlapEnd, nil
}

func TestPublicAPIFailsClosedAndKeepsVersionedErrorEnvelope(t *testing.T) {
	t.Parallel()

	workspaceID := uuid.New()
	resolver := &credentialResolverStub{err: errors.New("must not be called without a bearer")}
	workspaces := &workspaceReaderStub{}
	app := publicAPITestApp(t, resolver, &rateLimitStoreStub{count: 1}, workspaces, &webhookManagerStub{})
	request := httptest.NewRequest(http.MethodGet, "/api/v1/workspaces/"+workspaceID.String()+"?access_token=ignored", nil)
	request.AddCookie(&http.Cookie{Name: "token", Value: "ignored"})
	recorder := httptest.NewRecorder()

	app.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusUnauthorized, recorder.Code)
	require.Zero(t, resolver.calls)
	require.Zero(t, workspaces.calls)
	require.JSONEq(t, `{"error":{"code":"machine_authentication_required","message":"A valid machine bearer credential is required.","requestId":"`+recorder.Header().Get("X-Request-ID")+`"}}`, recorder.Body.String())
}

func TestPublicAPIRateLimitAndWorkspaceBoundariesFailBeforeServices(t *testing.T) {
	t.Parallel()

	credentialWorkspaceID := uuid.New()
	otherWorkspaceID := uuid.New()
	actor := testMachineActor(t, platformauth.PrincipalPersonalToken, credentialWorkspaceID, platformauth.ScopeWorkspacesRead)
	for _, test := range []struct {
		name       string
		store      *rateLimitStoreStub
		path       string
		wantStatus int
		wantCode   string
	}{
		{name: "rate limit unavailable", store: &rateLimitStoreStub{err: errors.New("redis unavailable")}, path: credentialWorkspaceID.String(), wantStatus: http.StatusServiceUnavailable, wantCode: "service_unavailable"},
		{name: "credential workspace mismatch", store: &rateLimitStoreStub{count: 1}, path: otherWorkspaceID.String(), wantStatus: http.StatusForbidden, wantCode: "access_denied"},
	} {
		t.Run(test.name, func(t *testing.T) {
			workspaceReader := &workspaceReaderStub{}
			app := publicAPITestApp(t, &credentialResolverStub{actor: actor}, test.store, workspaceReader, &webhookManagerStub{})
			request := httptest.NewRequest(http.MethodGet, "/api/v1/workspaces/"+test.path, nil)
			request.Header.Set("Authorization", "Bearer valid-machine-token")
			recorder := httptest.NewRecorder()

			app.ServeHTTP(recorder, request)

			require.Equal(t, test.wantStatus, recorder.Code)
			require.Contains(t, recorder.Body.String(), `"code":"`+test.wantCode+`"`)
			require.Zero(t, workspaceReader.calls)
		})
	}
}

func TestPublicAPIWorkspaceReadReturnsTypedContractAndRateHeaders(t *testing.T) {
	t.Parallel()

	workspaceID := uuid.New()
	actor := testMachineActor(t, platformauth.PrincipalPersonalToken, workspaceID, platformauth.ScopeWorkspacesRead)
	workspaceReader := &workspaceReaderStub{workspace: workspaces.CoreWorkspace{
		ID: workspaceID, Slug: "product", Name: "Product", Color: "#123456", IsActive: true,
		UserRole: "admin", CreatedAt: time.Date(2026, 8, 1, 10, 0, 0, 0, time.UTC), UpdatedAt: time.Date(2026, 8, 2, 10, 0, 0, 0, time.UTC),
	}}
	store := &rateLimitStoreStub{count: 1}
	app := publicAPITestApp(t, &credentialResolverStub{actor: actor}, store, workspaceReader, &webhookManagerStub{})
	request := httptest.NewRequest(http.MethodGet, "/api/v1/workspaces/"+workspaceID.String(), nil)
	request.Header.Set("Authorization", "Bearer valid-machine-token")
	recorder := httptest.NewRecorder()

	app.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusOK, recorder.Code, recorder.Body.String())
	require.Contains(t, recorder.Header().Get("RateLimit-Policy"), "api-v1-credential")
	require.Contains(t, recorder.Header().Get("RateLimit"), "r=599")
	require.Len(t, store.keys, 1)
	require.NotContains(t, store.keys[0], "valid-machine-token")
	require.JSONEq(t, `{"data":{"active":true,"color":"#123456","createdAt":"2026-08-01T10:00:00Z","id":"`+workspaceID.String()+`","name":"Product","role":"admin","slug":"product","updatedAt":"2026-08-02T10:00:00Z"}}`, recorder.Body.String())
}

func TestPublicAPIServiceAccountsCannotUseHumanMembershipReads(t *testing.T) {
	t.Parallel()

	workspaceID := uuid.New()
	actor := testMachineActor(t, platformauth.PrincipalServiceAccount, workspaceID, platformauth.ScopeWorkspacesRead)
	workspaceReader := &workspaceReaderStub{}
	app := publicAPITestApp(t, &credentialResolverStub{actor: actor}, &rateLimitStoreStub{count: 1}, workspaceReader, &webhookManagerStub{})
	request := httptest.NewRequest(http.MethodGet, "/api/v1/workspaces/"+workspaceID.String(), nil)
	request.Header.Set("Authorization", "Bearer valid-service-account-key")
	recorder := httptest.NewRecorder()

	app.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusForbidden, recorder.Code)
	require.Contains(t, recorder.Body.String(), `"code":"principal_not_supported"`)
	require.Zero(t, workspaceReader.calls)
}

func TestPublicAPIVersionedFallbacksNeverEmitPlainTextMuxErrors(t *testing.T) {
	t.Parallel()

	workspaceID := uuid.New()
	actor := testMachineActor(t, platformauth.PrincipalPersonalToken, workspaceID, platformauth.ScopeWorkspacesRead)
	app := publicAPITestApp(t, &credentialResolverStub{actor: actor}, &rateLimitStoreStub{count: 1}, &workspaceReaderStub{}, &webhookManagerStub{})
	for _, test := range []struct {
		name       string
		method     string
		path       string
		wantStatus int
		wantCode   string
		wantAllow  string
	}{
		{name: "known path wrong method", method: http.MethodDelete, path: "/api/v1/workspaces/" + workspaceID.String(), wantStatus: http.StatusMethodNotAllowed, wantCode: "method_not_allowed", wantAllow: "GET"},
		{name: "unknown versioned path", method: http.MethodGet, path: "/api/v1/not-a-route", wantStatus: http.StatusNotFound, wantCode: "resource_not_found"},
	} {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(test.method, test.path, nil)
			request.Header.Set("Authorization", "Bearer valid-machine-token")
			recorder := httptest.NewRecorder()

			app.ServeHTTP(recorder, request)

			require.Equal(t, test.wantStatus, recorder.Code, recorder.Body.String())
			require.Equal(t, "application/json", recorder.Header().Get("Content-Type"))
			require.Contains(t, recorder.Body.String(), `"code":"`+test.wantCode+`"`)
			require.Equal(t, test.wantAllow, recorder.Header().Get("Allow"))
		})
	}
}

func TestPublicAPICreatesWebhookAndShowsSigningSecretOnce(t *testing.T) {
	t.Parallel()

	workspaceID := uuid.New()
	endpointID := uuid.New()
	actor := testMachineActor(t, platformauth.PrincipalPersonalToken, workspaceID, platformauth.ScopeWebhooksManage)
	workspaceReader := &workspaceReaderStub{workspace: workspaces.CoreWorkspace{
		ID: workspaceID, UserRole: "admin", IsActive: true,
	}}
	manager := &webhookManagerStub{created: outboundwebhooksdomain.CreatedEndpoint{
		Endpoint: outboundwebhooksdomain.Endpoint{
			ID: endpointID, WorkspaceID: workspaceID, Name: "deploys", URL: "https://example.com/hooks/fortyone",
			Status: outboundwebhooksdomain.EndpointActive, SecretGeneration: 1, SubscriptionGeneration: 1,
			Subscriptions: []outboundwebhooksdomain.EventType{outboundwebhooksdomain.EventStoryCreated},
			CreatedAt:     time.Date(2026, 8, 28, 10, 0, 0, 0, time.UTC), UpdatedAt: time.Date(2026, 8, 28, 10, 0, 0, 0, time.UTC),
		},
		Secret: outboundwebhooksdomain.NewSigningSecret("whsec_v1_show-once"),
	}}
	app := publicAPITestApp(t, &credentialResolverStub{actor: actor}, &rateLimitStoreStub{count: 1}, workspaceReader, manager)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/workspaces/"+workspaceID.String()+"/webhook-endpoints", strings.NewReader(`{"name":"deploys","url":"https://example.com/hooks/fortyone","subscriptions":["story.created"]}`))
	request.Header.Set("Authorization", "Bearer valid-machine-token")
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	app.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusCreated, recorder.Code, recorder.Body.String())
	require.Contains(t, recorder.Body.String(), `"signingSecret":"whsec_v1_show-once"`)
	require.NotContains(t, recorder.Body.String(), "valid-machine-token")
	require.Equal(t, actor.PrincipalID, manager.access.Actor.PrincipalID)
	require.Equal(t, "deploys", manager.createInput.Name)
	require.Equal(t, []outboundwebhooksdomain.EventType{outboundwebhooksdomain.EventStoryCreated}, manager.createInput.Subscriptions)
	require.Equal(t, recorder.Header().Get("X-Request-ID"), manager.createInput.RequestID)
}

func TestPublicAPIExposesCompleteWebhookManagementContract(t *testing.T) {
	t.Parallel()

	workspaceID := uuid.New()
	endpointID := uuid.New()
	actor := testMachineActor(t, platformauth.PrincipalPersonalToken, workspaceID, platformauth.ScopeWebhooksManage)
	endpoint := outboundwebhooksdomain.Endpoint{
		ID: endpointID, WorkspaceID: workspaceID, Name: "deploys", URL: "https://example.com/hooks/fortyone",
		Status: outboundwebhooksdomain.EndpointActive, SecretGeneration: 1, SubscriptionGeneration: 1,
		Subscriptions: []outboundwebhooksdomain.EventType{outboundwebhooksdomain.EventStoryCreated},
		CreatedAt:     time.Date(2026, 8, 28, 10, 0, 0, 0, time.UTC), UpdatedAt: time.Date(2026, 8, 28, 10, 0, 0, 0, time.UTC),
	}
	for _, test := range []struct {
		name        string
		method      string
		suffix      string
		body        string
		operation   string
		wantStatus  int
		wantContent string
	}{
		{name: "list", method: http.MethodGet, operation: "list", wantStatus: http.StatusOK, wantContent: endpointID.String()},
		{name: "get", method: http.MethodGet, suffix: "/" + endpointID.String(), operation: "get", wantStatus: http.StatusOK, wantContent: endpointID.String()},
		{name: "subscriptions", method: http.MethodPut, suffix: "/" + endpointID.String() + "/subscriptions", body: `{"subscriptions":["story.updated"]}`, operation: "subscriptions", wantStatus: http.StatusNoContent},
		{name: "rotate", method: http.MethodPost, suffix: "/" + endpointID.String() + "/rotate-secret", operation: "rotate", wantStatus: http.StatusOK, wantContent: "whsec_v1_rotated"},
		{name: "disable", method: http.MethodPost, suffix: "/" + endpointID.String() + "/disable", body: `{"reason":"retired"}`, operation: "disable", wantStatus: http.StatusNoContent},
	} {
		t.Run(test.name, func(t *testing.T) {
			manager := &webhookManagerStub{
				endpoint: endpoint, page: outboundwebhooksdomain.EndpointPage{Items: []outboundwebhooksdomain.Endpoint{endpoint}},
				rotated: outboundwebhooksdomain.NewSigningSecret("whsec_v1_rotated"), generation: 2,
				overlapEnd: time.Date(2026, 8, 29, 10, 0, 0, 0, time.UTC),
			}
			workspaceReader := &workspaceReaderStub{workspace: workspaces.CoreWorkspace{ID: workspaceID, UserRole: "admin", IsActive: true}}
			app := publicAPITestApp(t, &credentialResolverStub{actor: actor}, &rateLimitStoreStub{count: 1}, workspaceReader, manager)
			request := httptest.NewRequest(test.method, "/api/v1/workspaces/"+workspaceID.String()+"/webhook-endpoints"+test.suffix, strings.NewReader(test.body))
			request.Header.Set("Authorization", "Bearer valid-machine-token")
			if test.body != "" {
				request.Header.Set("Content-Type", "application/json")
			}
			recorder := httptest.NewRecorder()

			app.ServeHTTP(recorder, request)

			require.Equal(t, test.wantStatus, recorder.Code, recorder.Body.String())
			if test.wantContent != "" {
				require.Contains(t, recorder.Body.String(), test.wantContent)
			}
			require.Equal(t, []string{test.operation}, manager.operations)
		})
	}
}

func publicAPITestApp(
	t *testing.T,
	resolver *credentialResolverStub,
	store *rateLimitStoreStub,
	workspaces *workspaceReaderStub,
	webhooks *webhookManagerStub,
) *web.App {
	t.Helper()
	log := logger.NewWithJSON(&bytes.Buffer{}, slog.LevelError, strings.ReplaceAll(t.Name(), "/", "-"))
	app := web.New(make(chan os.Signal, 1), noop.NewTracerProvider().Tracer("public-api-test"))
	app.SetLogger(log)
	Routes(Config{
		Log: log, SecretKey: strings.Repeat("test-public-api-secret", 2), Cache: store,
		DeveloperCredentials: resolver, Workspaces: workspaces, Teams: teamReaderStub{},
		Stories: storyReaderStub{}, StoryComments: storyCommentReaderStub{},
		Labels: labelReaderStub{}, States: workflowStateReaderStub{}, Sprints: sprintReaderStub{},
		Objectives: objectiveReaderStub{}, KeyResults: keyResultReaderStub{},
		Idempotency: idempotencyManagerStub{}, Webhooks: webhooks,
	}, app)
	return app
}

func testMachineActor(t *testing.T, kind platformauth.PrincipalKind, workspaceID uuid.UUID, scopes ...platformauth.Scope) platformauth.Actor {
	t.Helper()
	actor, err := platformauth.NewActor(uuid.New(), kind, uuid.New(), platformauth.MustScopeSet(scopes...), platformauth.UnrestrictedTeamAccess())
	require.NoError(t, err)
	actor, err = actor.WithWorkspace(workspaceID)
	require.NoError(t, err)
	return actor
}
