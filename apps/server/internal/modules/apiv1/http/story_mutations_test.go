package apiv1http

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	openapiv1 "github.com/complexus-tech/projects-api/internal/generated/openapi/v1"
	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/complexus-tech/projects-api/internal/platform/idempotency"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type storyMutationServiceStub struct {
	created stories.CoreSingleStory
	err     error
	input   stories.CoreNewStory
	calls   int
}

func (stub *storyMutationServiceStub) Get(context.Context, uuid.UUID, uuid.UUID) (stories.CoreSingleStory, error) {
	return stories.CoreSingleStory{}, stories.ErrNotFound
}

func (stub *storyMutationServiceStub) List(context.Context, uuid.UUID, stories.CoreStoryFilters) ([]stories.CoreStoryList, error) {
	return nil, nil
}

func (stub *storyMutationServiceStub) Create(_ context.Context, input stories.CoreNewStory, _ uuid.UUID) (stories.CoreSingleStory, error) {
	stub.calls++
	stub.input = input
	return stub.created, stub.err
}

type recordingIdempotencyManager struct {
	beginResult idempotency.BeginResult
	beginErr    error
	completeErr error
	requestBody []byte
	completed   idempotency.Response
	beginCalls  int
	completions int
}

func (stub *recordingIdempotencyManager) Begin(_ context.Context, _ idempotency.Scope, _ idempotency.Key, body []byte) (idempotency.BeginResult, error) {
	stub.beginCalls++
	stub.requestBody = append([]byte(nil), body...)
	return stub.beginResult, stub.beginErr
}

func (stub *recordingIdempotencyManager) Complete(_ context.Context, _ idempotency.Lease, response idempotency.Response) error {
	stub.completions++
	stub.completed = response
	return stub.completeErr
}

func TestCreateStoryUsesExactReceiptBytesAndSafeDomainCreationKey(t *testing.T) {
	t.Parallel()

	workspaceID := uuid.New()
	teamID := uuid.New()
	storyID := uuid.New()
	actor := testMachineActor(t, platformauth.PrincipalPersonalToken, workspaceID, platformauth.ScopeStoriesWrite)
	storyService := &storyMutationServiceStub{created: stories.CoreSingleStory{
		ID: storyID, Workspace: workspaceID, Team: teamID, TeamCode: "ENG", SequenceID: 41,
		Title: "Ship typed public writes", Priority: "High", AutoSchedulingStatus: stories.AutoSchedulingStatusOff,
		CreatedAt: time.Date(2026, 8, 28, 12, 0, 0, 0, time.UTC), UpdatedAt: time.Date(2026, 8, 28, 12, 0, 0, 0, time.UTC),
	}}
	lease := idempotency.Lease{ReceiptID: uuid.New(), Generation: 1, ExpiresAt: time.Now().Add(time.Minute)}
	receipts := &recordingIdempotencyManager{beginResult: idempotency.BeginResult{State: idempotency.BeginStateNew, Lease: lease}}
	operation, err := idempotency.ParseOperation("stories.create")
	require.NoError(t, err)
	server := &server{stories: storyService, idempotency: receipts, createStoryOperation: operation}

	rawKey := "public-story-create-key-0001"
	rawBody := []byte("{\n  \"title\": \"Ship typed public writes\", \"teamId\": \"" + teamID.String() + "\", \"priority\": \"High\"\n}\n")
	body := openapiv1.ComponentsResourcesCreateStoryRequest{
		Title: "Ship typed public writes", TeamId: teamID,
	}
	priority := openapiv1.ComponentsResourcesCreateStoryRequestPriority("High")
	body.Priority = &priority
	ctx := storyMutationTestContext(t, actor, rawBody)

	response, err := server.CreateStory(ctx, openapiv1.CreateStoryRequestObject{
		WorkspaceId: workspaceID,
		Params:      openapiv1.CreateStoryParams{IdempotencyKey: rawKey},
		Body:        &body,
	})

	require.NoError(t, err)
	created, ok := response.(openapiv1.CreateStory201JSONResponse)
	require.True(t, ok)
	require.Equal(t, storyID, created.Body.Data.Id)
	require.Equal(t, 1, storyService.calls)
	require.Equal(t, rawBody, receipts.requestBody, "receipt hashing must preserve caller bytes")
	require.Equal(t, 1, receipts.completions)
	require.Equal(t, 201, receipts.completed.StatusCode())
	require.Equal(t, jsonContentType, receipts.completed.ContentType())
	require.NotNil(t, storyService.input.CreationKey)
	require.NotContains(t, *storyService.input.CreationKey, rawKey)
	require.True(t, strings.HasPrefix(*storyService.input.CreationKey, "api-v1:personal_token:"))
	require.Equal(t, "High", storyService.input.Priority)

	replayed, err := replayCreateStoryResponse(receipts.completed)
	require.NoError(t, err)
	replayBody, ok := replayed.(openapiv1.CreateStory201JSONResponse)
	require.True(t, ok)
	require.Equal(t, created.Body, replayBody.Body)
}

func TestCreateStoryRejectsIdempotencyConflictBeforeMutation(t *testing.T) {
	t.Parallel()

	workspaceID := uuid.New()
	teamID := uuid.New()
	actor := testMachineActor(t, platformauth.PrincipalPersonalToken, workspaceID, platformauth.ScopeStoriesWrite)
	storyService := &storyMutationServiceStub{err: errors.New("must not be called")}
	receipts := &recordingIdempotencyManager{beginResult: idempotency.BeginResult{State: idempotency.BeginStateConflict}}
	operation, err := idempotency.ParseOperation("stories.create")
	require.NoError(t, err)
	server := &server{stories: storyService, idempotency: receipts, createStoryOperation: operation}
	body := openapiv1.ComponentsResourcesCreateStoryRequest{Title: "Different request", TeamId: teamID}
	ctx := storyMutationTestContext(t, actor, []byte(`{"title":"Different request"}`))

	response, err := server.CreateStory(ctx, openapiv1.CreateStoryRequestObject{
		WorkspaceId: workspaceID,
		Params:      openapiv1.CreateStoryParams{IdempotencyKey: "public-story-create-key-0001"},
		Body:        &body,
	})

	require.NoError(t, err)
	conflict, ok := response.(openapiv1.CreateStory409JSONResponse)
	require.True(t, ok)
	require.Equal(t, "idempotency_key_reused", conflict.Body.Error.Code)
	require.Zero(t, storyService.calls)
	require.Zero(t, receipts.completions)
}

func TestCreateStoryRejectsOutOfRangeEstimateBeforeReceipt(t *testing.T) {
	t.Parallel()

	workspaceID := uuid.New()
	actor := testMachineActor(t, platformauth.PrincipalPersonalToken, workspaceID, platformauth.ScopeStoriesWrite)
	storyService := &storyMutationServiceStub{err: errors.New("must not be called")}
	receipts := &recordingIdempotencyManager{}
	operation, err := idempotency.ParseOperation("stories.create")
	require.NoError(t, err)
	server := &server{stories: storyService, idempotency: receipts, createStoryOperation: operation}
	overflow := int32(32768)
	body := openapiv1.ComponentsResourcesCreateStoryRequest{Title: "Invalid estimate", TeamId: uuid.New(), EstimateValue: &overflow}
	ctx := storyMutationTestContext(t, actor, []byte(`{"title":"Invalid estimate","estimateValue":32768}`))

	response, err := server.CreateStory(ctx, openapiv1.CreateStoryRequestObject{
		WorkspaceId: workspaceID,
		Params:      openapiv1.CreateStoryParams{IdempotencyKey: "out-of-range-estimate-0001"},
		Body:        &body,
	})

	require.NoError(t, err)
	rejected, ok := response.(openapiv1.CreateStory400JSONResponse)
	require.True(t, ok)
	require.Equal(t, "invalid_request", rejected.Body.Error.Code)
	require.Zero(t, receipts.beginCalls)
	require.Zero(t, storyService.calls)
}

func TestCreateStoryPreservesServiceAccountAttribution(t *testing.T) {
	t.Parallel()

	workspaceID := uuid.New()
	teamID := uuid.New()
	actor := testMachineActor(t, platformauth.PrincipalServiceAccount, workspaceID, platformauth.ScopeStoriesWrite)
	storyService := &storyMutationServiceStub{created: stories.CoreSingleStory{
		ID: uuid.New(), Workspace: workspaceID, Team: teamID, TeamCode: "ENG", SequenceID: 42,
		Title: "Machine-created story", Priority: "No Priority", AutoSchedulingStatus: stories.AutoSchedulingStatusOff,
		CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC(),
	}}
	lease := idempotency.Lease{ReceiptID: uuid.New(), Generation: 1, ExpiresAt: time.Now().Add(time.Minute)}
	receipts := &recordingIdempotencyManager{beginResult: idempotency.BeginResult{State: idempotency.BeginStateNew, Lease: lease}}
	operation, err := idempotency.ParseOperation("stories.create")
	require.NoError(t, err)
	server := &server{stories: storyService, idempotency: receipts, createStoryOperation: operation}
	body := openapiv1.ComponentsResourcesCreateStoryRequest{Title: "Machine-created story", TeamId: teamID}
	key := "service-account-story-key-0001"
	ctx := storyMutationTestContext(t, actor, []byte(`{"title":"Machine-created story"}`))

	response, err := server.CreateStory(ctx, openapiv1.CreateStoryRequestObject{
		WorkspaceId: workspaceID,
		Params:      openapiv1.CreateStoryParams{IdempotencyKey: key},
		Body:        &body,
	})

	require.NoError(t, err)
	_, created := response.(openapiv1.CreateStory201JSONResponse)
	require.True(t, created)
	require.Equal(t, 1, storyService.calls)
	require.Nil(t, storyService.input.Reporter, "machine actors must not be represented as human reporters")
	require.NotNil(t, storyService.input.CreationKey)
	require.True(t, strings.HasPrefix(*storyService.input.CreationKey, "api-v1:service_account:"))
}

func TestCaptureJSONBodyPreservesExactBytes(t *testing.T) {
	t.Parallel()

	raw := []byte("{ \"title\" : \"spacing matters\" }\n")
	request := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(raw))
	var captured []byte
	handler := captureJSONBody(func(ctx context.Context, _ http.ResponseWriter, request *http.Request) error {
		var ok bool
		captured, ok = exactRequestBody(ctx)
		require.True(t, ok)
		reread, err := io.ReadAll(request.Body)
		require.NoError(t, err)
		require.Equal(t, raw, reread)
		return nil
	})
	require.NoError(t, handler(context.Background(), httptest.NewRecorder(), request))
	require.True(t, bytes.Equal(raw, captured))
}

func storyMutationTestContext(t *testing.T, actor platformauth.Actor, rawBody []byte) context.Context {
	t.Helper()
	ctx, err := platformauth.SetActor(context.Background(), actor)
	require.NoError(t, err)
	ctx = web.SetValues(ctx, &web.Values{RequestID: "request-test"})
	return context.WithValue(ctx, requestBodyContextKey{}, append([]byte(nil), rawBody...))
}
