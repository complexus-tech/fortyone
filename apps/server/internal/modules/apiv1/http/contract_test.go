package apiv1http

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	openapiv1 "github.com/complexus-tech/projects-api/internal/generated/openapi/v1"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestPublicAPIOpenAPIContractIsVersionedAndSecured(t *testing.T) {
	t.Parallel()

	spec, err := openapiv1.GetSpec()
	require.NoError(t, err)
	require.NoError(t, spec.Validate(t.Context()))
	require.Equal(t, "3.1.0", spec.OpenAPI)
	developerBearer := spec.Components.SecuritySchemes["machineBearer"]
	require.NotNil(t, developerBearer, "the preview scheme identifier is compatibility-stable")
	require.NotNil(t, developerBearer.Value)
	require.Equal(t, "FortyOne developer credential", developerBearer.Value.BearerFormat)
	description := strings.ToLower(strings.Join(strings.Fields(spec.Info.Description), " "))
	require.Contains(t, description, "credential recognition does not imply operation authorization")
	require.Contains(t, description, "user-authorized oauth token")
	require.Contains(t, description, "exact `/api/v1` audience")
	require.Contains(t, description, "story creation also")
	require.NotEmpty(t, spec.Security)
	require.Len(t, spec.Paths.Map(), 16)
	for path, item := range spec.Paths.Map() {
		require.True(t, strings.HasPrefix(path, "/api/v1/"), path)
		for method, operation := range item.Operations() {
			require.NotEmpty(t, operation.OperationID, method+" "+path)
			require.NotNil(t, operation.Responses, method+" "+path)
			require.Contains(t, strings.ToLower(operation.Description), "personal access token", method+" "+path)
		}
	}
}

func TestPublicAPIContractRejectsUnknownAndOversizedWebhookBodies(t *testing.T) {
	t.Parallel()

	workspaceID := uuid.New()
	actor := testMachineActor(t, platformauth.PrincipalPersonalToken, workspaceID, platformauth.ScopeWebhooksManage)
	for _, test := range []struct {
		name        string
		body        string
		contentType string
		wantStatus  int
		wantCode    string
	}{
		{name: "unknown property", body: `{"name":"deploys","url":"https://example.com/hook","subscriptions":["story.created"],"secret":"caller-chosen"}`, contentType: "application/json", wantStatus: http.StatusBadRequest, wantCode: "invalid_request"},
		{name: "wrong media type", body: `{}`, contentType: "text/plain", wantStatus: http.StatusUnsupportedMediaType, wantCode: "unsupported_media_type"},
		{name: "oversized", body: `{"name":"` + strings.Repeat("a", int(maximumJSONBytes)) + `"}`, contentType: "application/json", wantStatus: http.StatusRequestEntityTooLarge, wantCode: "request_too_large"},
	} {
		t.Run(test.name, func(t *testing.T) {
			manager := &webhookManagerStub{}
			workspaceReader := &workspaceReaderStub{}
			app := publicAPITestApp(t, &credentialResolverStub{actor: actor}, &rateLimitStoreStub{count: 1}, workspaceReader, manager)
			request := httptest.NewRequest(http.MethodPost, "/api/v1/workspaces/"+workspaceID.String()+"/webhook-endpoints", strings.NewReader(test.body))
			request.Header.Set("Authorization", "Bearer valid-machine-token")
			request.Header.Set("Content-Type", test.contentType)
			recorder := httptest.NewRecorder()

			app.ServeHTTP(recorder, request)

			require.Equal(t, test.wantStatus, recorder.Code, recorder.Body.String())
			require.Contains(t, recorder.Body.String(), `"code":"`+test.wantCode+`"`)
			require.Zero(t, workspaceReader.calls)
			require.Zero(t, manager.calls)
		})
	}
}

func TestPublicAPICursorsAreSignedAndBoundToPrincipalAndFilters(t *testing.T) {
	t.Parallel()

	codec, err := newCursorCodec[cursorPage](strings.Repeat("cursor-secret", 4), "stories")
	require.NoError(t, err)
	workspaceID := uuid.New()
	principalID := uuid.New()
	teamID := uuid.New()
	resourceID := uuid.New()
	limit := 25
	token, err := codec.Encode(cursorPage{
		Version: 1, WorkspaceID: workspaceID, PrincipalID: principalID,
		TeamID: teamID, ResourceID: resourceID, Limit: limit, Offset: 25,
	})
	require.NoError(t, err)

	page, problem := decodeOffsetPage(codec, &token, &limit, workspaceID, principalID, teamID, resourceID)
	require.Nil(t, problem)
	require.Equal(t, 25, page.Offset)

	_, problem = decodeOffsetPage(codec, &token, &limit, workspaceID, uuid.New(), teamID, resourceID)
	require.NotNil(t, problem)
	require.Equal(t, "invalid_cursor", problem.code)
	_, problem = decodeOffsetPage(codec, &token, &limit, workspaceID, principalID, teamID, uuid.New())
	require.NotNil(t, problem)
	require.Equal(t, "invalid_cursor", problem.code)
	parts := strings.Split(token, ".")
	require.Len(t, parts, 4)
	replacement := "A"
	if strings.HasPrefix(parts[3], replacement) {
		replacement = "B"
	}
	parts[3] = replacement + parts[3][1:]
	tampered := strings.Join(parts, ".")
	_, problem = decodeOffsetPage(codec, &tampered, &limit, workspaceID, principalID, teamID, resourceID)
	require.NotNil(t, problem)
	require.Equal(t, "invalid_cursor", problem.code)
}
