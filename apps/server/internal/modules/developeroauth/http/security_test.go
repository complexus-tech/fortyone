package developeroauthhttp

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	developeroauthdomain "github.com/complexus-tech/projects-api/internal/modules/developeroauth/domain"
	developeroauth "github.com/complexus-tech/projects-api/internal/modules/developeroauth/service"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/complexus-tech/projects-api/internal/platform/authorization"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type managementServiceSpy struct {
	createCalls             int
	listApplicationsCalls   int
	listSecretsCalls        int
	rotateCalls             int
	revokeSecretCalls       int
	installCalls            int
	listInstallationsCalls  int
	updateInstallationCalls int
	revokeInstallationCalls int
	lastAccess              developeroauthdomain.ManagementAccess
	lastApplicationID       uuid.UUID
	lastSecretID            uuid.UUID
	lastInstallationID      uuid.UUID
	lastCreateInput         developeroauth.CreateManagedApplicationInput
	lastRotateInput         developeroauth.RotateClientSecretInput
	lastInstallInput        developeroauth.InstallApplicationInput
	lastUpdateInput         developeroauth.UpdateApplicationInstallationInput
	createResult            developeroauthdomain.IssuedManagedApplication
	applicationsResult      []developeroauthdomain.ManagedApplication
	secretsResult           []developeroauthdomain.ClientSecret
	rotateResult            developeroauthdomain.IssuedClientSecret
	installationResult      developeroauthdomain.ApplicationInstallation
	installationsResult     []developeroauthdomain.ApplicationInstallation
	err                     error
}

func (spy *managementServiceSpy) CreateManagedApplication(
	_ context.Context,
	access developeroauthdomain.ManagementAccess,
	input developeroauth.CreateManagedApplicationInput,
) (developeroauthdomain.IssuedManagedApplication, error) {
	spy.createCalls++
	spy.lastAccess = access
	spy.lastCreateInput = input
	return spy.createResult, spy.err
}

func (spy *managementServiceSpy) ListManagedApplications(
	_ context.Context,
	access developeroauthdomain.ManagementAccess,
) ([]developeroauthdomain.ManagedApplication, error) {
	spy.listApplicationsCalls++
	spy.lastAccess = access
	return spy.applicationsResult, spy.err
}

func (spy *managementServiceSpy) ListClientSecrets(
	_ context.Context,
	access developeroauthdomain.ManagementAccess,
	applicationID uuid.UUID,
) ([]developeroauthdomain.ClientSecret, error) {
	spy.listSecretsCalls++
	spy.lastAccess = access
	spy.lastApplicationID = applicationID
	return spy.secretsResult, spy.err
}

func (spy *managementServiceSpy) RotateClientSecret(
	_ context.Context,
	access developeroauthdomain.ManagementAccess,
	applicationID uuid.UUID,
	input developeroauth.RotateClientSecretInput,
) (developeroauthdomain.IssuedClientSecret, error) {
	spy.rotateCalls++
	spy.lastAccess = access
	spy.lastApplicationID = applicationID
	spy.lastRotateInput = input
	return spy.rotateResult, spy.err
}

func (spy *managementServiceSpy) RevokeClientSecret(
	_ context.Context,
	access developeroauthdomain.ManagementAccess,
	applicationID uuid.UUID,
	secretID uuid.UUID,
	_ developeroauth.RevokeApplicationInput,
) error {
	spy.revokeSecretCalls++
	spy.lastAccess = access
	spy.lastApplicationID = applicationID
	spy.lastSecretID = secretID
	return spy.err
}

func (spy *managementServiceSpy) InstallApplication(
	_ context.Context,
	access developeroauthdomain.ManagementAccess,
	input developeroauth.InstallApplicationInput,
) (developeroauthdomain.ApplicationInstallation, error) {
	spy.installCalls++
	spy.lastAccess = access
	spy.lastInstallInput = input
	return spy.installationResult, spy.err
}

func (spy *managementServiceSpy) ListApplicationInstallations(
	_ context.Context,
	access developeroauthdomain.ManagementAccess,
) ([]developeroauthdomain.ApplicationInstallation, error) {
	spy.listInstallationsCalls++
	spy.lastAccess = access
	return spy.installationsResult, spy.err
}

func (spy *managementServiceSpy) UpdateApplicationInstallation(
	_ context.Context,
	access developeroauthdomain.ManagementAccess,
	installationID uuid.UUID,
	input developeroauth.UpdateApplicationInstallationInput,
) (developeroauthdomain.ApplicationInstallation, error) {
	spy.updateInstallationCalls++
	spy.lastAccess = access
	spy.lastInstallationID = installationID
	spy.lastUpdateInput = input
	return spy.installationResult, spy.err
}

func (spy *managementServiceSpy) RevokeApplicationInstallation(
	_ context.Context,
	access developeroauthdomain.ManagementAccess,
	installationID uuid.UUID,
	_ developeroauth.RevokeApplicationInput,
) error {
	spy.revokeInstallationCalls++
	spy.lastAccess = access
	spy.lastInstallationID = installationID
	return spy.err
}

type workspaceResolverStub struct {
	workspace mid.WorkspaceInfo
}

func (stub workspaceResolverStub) ResolveCurrentWorkspace(
	_ context.Context,
	slug string,
	_ uuid.UUID,
) (mid.WorkspaceInfo, error) {
	if slug != stub.workspace.Slug {
		return mid.WorkspaceInfo{}, mid.ErrWorkspaceAccessDenied
	}
	return stub.workspace, nil
}

func (workspaceResolverStub) RecordWorkspaceAccess(context.Context, uuid.UUID, uuid.UUID) error {
	return nil
}

func TestManagedApplicationJSONIsStrictAndBounded(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 8, 28, 12, 0, 0, 0, time.UTC)
	validFields := `"name":"Automation","redirectUris":[],"expiresAt":"` +
		now.Add(24*time.Hour).Format(time.RFC3339) + `","secretExpiresAt":"` +
		now.Add(time.Hour).Format(time.RFC3339) + `"`
	tests := []struct {
		name       string
		body       string
		wantStatus int
	}{
		{name: "unknown field", body: `{` + validFields + `,"unexpected":true}`, wantStatus: http.StatusBadRequest},
		{name: "trailing JSON", body: `{` + validFields + `} {}`, wantStatus: http.StatusBadRequest},
		{name: "malformed date", body: `{"name":"Automation","expiresAt":"tomorrow","secretExpiresAt":"` + now.Format(time.RFC3339) + `"}`, wantStatus: http.StatusBadRequest},
		{name: "missing date", body: `{"name":"Automation","secretExpiresAt":"` + now.Format(time.RFC3339) + `"}`, wantStatus: http.StatusBadRequest},
		{name: "bounded body", body: `{"name":"` + strings.Repeat("a", int(maximumManagementJSONBytes)+1) + `"}`, wantStatus: http.StatusRequestEntityTooLarge},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			workspace := mid.WorkspaceInfo{ID: uuid.New(), Slug: "acme", UserRole: string(mid.RoleAdmin)}
			ctx := managementContext(t, platformauth.NewHumanActor(uuid.New()), workspace)
			spy := &managementServiceSpy{}
			request := jsonRequest(http.MethodPost, "/workspaces/acme/oauth-applications", test.body)
			recorder := httptest.NewRecorder()

			require.NoError(t, New(spy).CreateManagedApplication(ctx, recorder, request))
			require.Equal(t, test.wantStatus, recorder.Code)
			require.Zero(t, spy.createCalls)
		})
	}
}

func TestClientSecretRotationRequiresAnExplicitBoundedOverlap(t *testing.T) {
	t.Parallel()

	expiresAt := time.Date(2026, 8, 29, 12, 0, 0, 0, time.UTC)
	tests := []struct {
		name           string
		overlapJSON    string
		wantStatus     int
		wantDuration   time.Duration
		wantServiceRun bool
	}{
		{name: "missing", overlapJSON: "", wantStatus: http.StatusBadRequest},
		{name: "null", overlapJSON: `,"overlapSeconds":null`, wantStatus: http.StatusBadRequest},
		{name: "below minimum", overlapJSON: `,"overlapSeconds":59`, wantStatus: http.StatusBadRequest},
		{name: "above maximum", overlapJSON: `,"overlapSeconds":86401`, wantStatus: http.StatusBadRequest},
		{name: "minimum", overlapJSON: `,"overlapSeconds":60`, wantStatus: http.StatusCreated, wantDuration: time.Minute, wantServiceRun: true},
		{name: "maximum", overlapJSON: `,"overlapSeconds":86400`, wantStatus: http.StatusCreated, wantDuration: 24 * time.Hour, wantServiceRun: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			workspace := mid.WorkspaceInfo{ID: uuid.New(), Slug: "acme", UserRole: string(mid.RoleAdmin)}
			ctx := managementContext(t, platformauth.NewHumanActor(uuid.New()), workspace)
			spy := &managementServiceSpy{}
			request := jsonRequest(
				http.MethodPost,
				"/workspaces/acme/oauth-applications/app/secrets/rotate",
				`{"expiresAt":"`+expiresAt.Format(time.RFC3339)+`"`+test.overlapJSON+`}`,
			)
			request.SetPathValue("applicationId", uuid.NewString())
			recorder := httptest.NewRecorder()

			require.NoError(t, New(spy).RotateClientSecret(ctx, recorder, request))
			require.Equal(t, test.wantStatus, recorder.Code)
			if test.wantServiceRun {
				require.Equal(t, 1, spy.rotateCalls)
				require.Equal(t, test.wantDuration, spy.lastRotateInput.Overlap)
				require.Equal(t, "no-store", recorder.Header().Get("Cache-Control"))
				require.Equal(t, "no-cache", recorder.Header().Get("Pragma"))
			} else {
				require.Zero(t, spy.rotateCalls)
			}
		})
	}
}

func TestMalformedRouteIdentifiersFailBeforeService(t *testing.T) {
	t.Parallel()

	workspace := mid.WorkspaceInfo{ID: uuid.New(), Slug: "acme", UserRole: string(mid.RoleAdmin)}
	ctx := managementContext(t, platformauth.NewHumanActor(uuid.New()), workspace)
	tests := []struct {
		name   string
		invoke func(*Handlers, context.Context, http.ResponseWriter) error
		called func(*managementServiceSpy) int
	}{
		{
			name: "application id",
			invoke: func(handlers *Handlers, ctx context.Context, writer http.ResponseWriter) error {
				request := httptest.NewRequest(http.MethodGet, "/workspaces/acme/oauth-applications/not-a-uuid/secrets", nil)
				request.SetPathValue("applicationId", "not-a-uuid")
				return handlers.ListClientSecrets(ctx, writer, request)
			},
			called: func(spy *managementServiceSpy) int { return spy.listSecretsCalls },
		},
		{
			name: "zero secret id",
			invoke: func(handlers *Handlers, ctx context.Context, writer http.ResponseWriter) error {
				request := httptest.NewRequest(http.MethodDelete, "/workspaces/acme/oauth-applications/app/secrets/secret", nil)
				request.SetPathValue("applicationId", uuid.NewString())
				request.SetPathValue("secretId", uuid.Nil.String())
				return handlers.RevokeClientSecret(ctx, writer, request)
			},
			called: func(spy *managementServiceSpy) int { return spy.revokeSecretCalls },
		},
		{
			name: "installation id",
			invoke: func(handlers *Handlers, ctx context.Context, writer http.ResponseWriter) error {
				request := jsonRequest(http.MethodPut, "/workspaces/acme/oauth-application-installations/nope", `{"resource":"/api/v1","scopes":["stories:write"]}`)
				request.SetPathValue("installationId", "nope")
				return handlers.UpdateApplicationInstallation(ctx, writer, request)
			},
			called: func(spy *managementServiceSpy) int { return spy.updateInstallationCalls },
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			spy := &managementServiceSpy{}
			recorder := httptest.NewRecorder()
			require.NoError(t, test.invoke(New(spy), ctx, recorder))
			require.Equal(t, http.StatusBadRequest, recorder.Code)
			require.Zero(t, test.called(spy))
		})
	}
}

func TestManagementBoundaryRejectsUnauthorizedActorsBeforeService(t *testing.T) {
	t.Parallel()

	validBody := `{"name":"Automation","expiresAt":"2026-09-30T12:00:00Z","secretExpiresAt":"2026-09-01T12:00:00Z"}`
	tests := []struct {
		name    string
		context func(*testing.T, mid.WorkspaceInfo) context.Context
		wrap    func(*logger.Logger, web.Handler) web.Handler
	}{
		{
			name: "machine actor",
			context: func(t *testing.T, workspace mid.WorkspaceInfo) context.Context {
				actor, err := platformauth.NewActor(
					uuid.New(), platformauth.PrincipalServiceAccount, uuid.New(),
					platformauth.MustScopeSet(platformauth.ScopeIntegrationsManage),
					platformauth.UnrestrictedTeamAccess(),
				)
				require.NoError(t, err)
				actor, err = actor.WithWorkspace(workspace.ID)
				require.NoError(t, err)
				ctx, err := platformauth.SetActor(context.Background(), actor)
				require.NoError(t, err)
				return ctx
			},
		},
		{
			name: "cross workspace actor",
			context: func(t *testing.T, workspace mid.WorkspaceInfo) context.Context {
				ctx := managementContext(t, platformauth.NewHumanActor(uuid.New()), workspace)
				actor, err := platformauth.GetActor(ctx)
				require.NoError(t, err)
				actor, err = actor.WithWorkspace(uuid.New())
				require.NoError(t, err)
				ctx, err = platformauth.SetActor(ctx, actor)
				require.NoError(t, err)
				return ctx
			},
		},
		{
			name: "non admin",
			context: func(t *testing.T, workspace mid.WorkspaceInfo) context.Context {
				workspace.UserRole = string(mid.RoleMember)
				return managementContext(t, platformauth.NewHumanActor(uuid.New()), workspace)
			},
			wrap: func(log *logger.Logger, handler web.Handler) web.Handler {
				return mid.RequireMinimumRole(log, mid.RoleAdmin)(handler)
			},
		},
		{
			name: "missing integration scope",
			context: func(t *testing.T, workspace mid.WorkspaceInfo) context.Context {
				actor, err := platformauth.NewActor(
					uuid.New(), platformauth.PrincipalHumanUser, uuid.Nil,
					platformauth.MustScopeSet(platformauth.ScopeStoriesRead),
					platformauth.UnrestrictedTeamAccess(),
				)
				require.NoError(t, err)
				return managementContext(t, actor, workspace)
			},
			wrap: func(_ *logger.Logger, handler web.Handler) web.Handler {
				return mid.RequireScopes(platformauth.ScopeIntegrationsManage)(handler)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			workspace := mid.WorkspaceInfo{ID: uuid.New(), Slug: "acme", UserRole: string(mid.RoleAdmin)}
			spy := &managementServiceSpy{}
			handler := web.Handler(New(spy).CreateManagedApplication)
			log := testLogger()
			if test.wrap != nil {
				handler = test.wrap(log, handler)
			}
			request := jsonRequest(http.MethodPost, "/workspaces/acme/oauth-applications", validBody)
			recorder := httptest.NewRecorder()

			require.NoError(t, handler(test.context(t, workspace), recorder, request))
			require.Equal(t, http.StatusForbidden, recorder.Code)
			require.Zero(t, spy.createCalls)
		})
	}
}

func TestWorkspaceAccessIsPreservedForTheServiceRecheck(t *testing.T) {
	t.Parallel()

	workspace := mid.WorkspaceInfo{ID: uuid.New(), Slug: "acme", UserRole: string(mid.RoleAdmin)}
	userID := uuid.New()
	ctx := managementContext(t, platformauth.NewHumanActor(userID), workspace)
	spy := &managementServiceSpy{}
	recorder := httptest.NewRecorder()

	require.NoError(t, New(spy).ListManagedApplications(
		ctx,
		recorder,
		httptest.NewRequest(http.MethodGet, "/workspaces/acme/oauth-applications", nil),
	))
	require.Equal(t, http.StatusOK, recorder.Code)
	require.Equal(t, 1, spy.listApplicationsCalls)
	require.Equal(t, workspace.ID, spy.lastAccess.WorkspaceID)
	require.Equal(t, authorization.WorkspaceRoleAdmin, spy.lastAccess.WorkspaceRole)
	require.Equal(t, userID, spy.lastAccess.Actor.PrincipalID)
	require.Equal(t, platformauth.PrincipalHumanUser, spy.lastAccess.Actor.Kind)
}

func TestClientSecretMetadataNeverLeaksStoredOrPlaintextSecretMaterial(t *testing.T) {
	t.Parallel()

	plaintext := "f41_ocs_v1_abcdef_show-once"
	metadata := developeroauthdomain.ClientSecret{
		ID: uuid.New(), ApplicationID: uuid.New(), LookupPrefix: "abcdef012345",
		ExpiresAt: time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC),
		CreatedAt: time.Date(2026, 8, 28, 12, 0, 0, 0, time.UTC),
	}
	listed, err := json.Marshal(clientSecretModels([]developeroauthdomain.ClientSecret{metadata}))
	require.NoError(t, err)
	require.NotContains(t, string(listed), plaintext)
	require.NotContains(t, string(listed), `"secret":`)
	require.NotContains(t, strings.ToLower(string(listed)), "digest")
	require.Contains(t, string(listed), `"prefix":"abcdef012345"`)

	cutoff := time.Date(2026, 8, 28, 12, 5, 0, 0, time.UTC)
	issued, err := json.Marshal(issuedClientSecretModel(developeroauthdomain.IssuedClientSecret{
		Secret: metadata, Plaintext: developeroauthdomain.NewPlaintextSecret(plaintext),
		PreviousSecretOverlapExpiresAt: &cutoff,
	}))
	require.NoError(t, err)
	require.Equal(t, 1, strings.Count(string(issued), plaintext))
	require.Equal(t, 1, strings.Count(string(issued), `"secret":`))
	require.Contains(t, string(issued), `"previousSecretOverlapExpiresAt":"2026-08-28T12:05:00Z"`)
}

func TestCreatedManagedApplicationSecretResponseCannotBeCached(t *testing.T) {
	t.Parallel()

	workspace := mid.WorkspaceInfo{ID: uuid.New(), Slug: "acme", UserRole: string(mid.RoleAdmin)}
	ctx := managementContext(t, platformauth.NewHumanActor(uuid.New()), workspace)
	spy := &managementServiceSpy{createResult: developeroauthdomain.IssuedManagedApplication{
		Secret: developeroauthdomain.IssuedClientSecret{
			Plaintext: developeroauthdomain.NewPlaintextSecret("f41_ocs_v1_abcdef_show-once"),
		},
	}}
	request := jsonRequest(
		http.MethodPost,
		"/workspaces/acme/oauth-applications",
		`{"name":"Automation","expiresAt":"2026-09-30T12:00:00Z","secretExpiresAt":"2026-09-01T12:00:00Z"}`,
	)
	recorder := httptest.NewRecorder()

	require.NoError(t, New(spy).CreateManagedApplication(ctx, recorder, request))
	require.Equal(t, http.StatusCreated, recorder.Code)
	require.Equal(t, "no-store", recorder.Header().Get("Cache-Control"))
	require.Equal(t, "no-cache", recorder.Header().Get("Pragma"))
}

func managementContext(
	t *testing.T,
	actor platformauth.Actor,
	workspace mid.WorkspaceInfo,
) context.Context {
	t.Helper()

	ctx, err := platformauth.SetActor(context.Background(), actor)
	require.NoError(t, err)
	request := httptest.NewRequest(http.MethodGet, "/workspaces/"+workspace.Slug, nil)
	request.SetPathValue("workspaceSlug", workspace.Slug)
	recorder := httptest.NewRecorder()
	var resolved context.Context
	resolver := workspaceResolverStub{workspace: workspace}
	handler := mid.Workspace(testLogger(), resolver)(func(ctx context.Context, _ http.ResponseWriter, _ *http.Request) error {
		resolved = ctx
		return nil
	})

	require.NoError(t, handler(ctx, recorder, request))
	require.NotNil(t, resolved)
	return resolved
}

func jsonRequest(method, path, body string) *http.Request {
	request := httptest.NewRequest(method, path, strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	return request
}

func testLogger() *logger.Logger {
	return logger.NewWithText(io.Discard, slog.LevelError, "developeroauth-http-test")
}
