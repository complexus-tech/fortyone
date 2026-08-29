package apiv1http

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"time"

	openapiv1 "github.com/complexus-tech/projects-api/internal/generated/openapi/v1"
	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/complexus-tech/projects-api/internal/platform/idempotency"
	"github.com/complexus-tech/projects-api/internal/platform/safecast"
	"github.com/google/uuid"
)

const jsonContentType = "application/json"

func (s *server) CreateStory(ctx context.Context, request openapiv1.CreateStoryRequestObject) (openapiv1.CreateStoryResponseObject, error) {
	actor, problem := actorFor(ctx, request.WorkspaceId, platformauth.ScopeStoriesWrite)
	if problem == nil {
		problem = requireStoryWriter(actor)
	}
	if problem != nil {
		return createStoryFailure(ctx, problem, 0), nil
	}
	if request.Body == nil {
		return createStoryFailure(ctx, &failure{http.StatusBadRequest, "invalid_request", "The request body is required."}, 0), nil
	}
	rawBody, ok := exactRequestBody(ctx)
	if !ok {
		s.log.Error(ctx, "public API exact request body is unavailable")
		return createStoryFailure(ctx, statusFailure(http.StatusInternalServerError), 0), nil
	}
	input, err := storyCreateInput(*request.Body, actor, request.Params.IdempotencyKey)
	if err != nil {
		return createStoryFailure(ctx, &failure{http.StatusBadRequest, "invalid_request", "The request is invalid."}, 0), nil
	}

	key, err := idempotency.ParseKey(request.Params.IdempotencyKey)
	if err != nil {
		return createStoryFailure(ctx, &failure{http.StatusBadRequest, "invalid_idempotency_key", "Idempotency-Key is invalid."}, 0), nil
	}
	scope, err := idempotency.NewScope(actor, idempotency.MethodPost, s.createStoryOperation)
	if err != nil {
		s.log.Error(ctx, "failed to construct story-create idempotency scope")
		return createStoryFailure(ctx, statusFailure(http.StatusInternalServerError), 0), nil
	}
	begin, err := s.idempotency.Begin(ctx, scope, key, rawBody)
	if err != nil {
		s.log.Error(ctx, "failed to begin story-create idempotency receipt")
		return createStoryFailure(ctx, statusFailure(http.StatusServiceUnavailable), 1), nil
	}
	switch begin.State {
	case idempotency.BeginStateCompleted:
		return replayCreateStoryResponse(begin.Replay)
	case idempotency.BeginStateInProgress:
		retryAfter := retryAfterSeconds(begin.RetryAt)
		return createStoryFailure(ctx, &failure{
			http.StatusConflict,
			"idempotency_in_progress",
			"A request with this Idempotency-Key is still in progress.",
		}, retryAfter), nil
	case idempotency.BeginStateConflict:
		return createStoryFailure(ctx, &failure{
			http.StatusConflict,
			"idempotency_key_reused",
			"Idempotency-Key was already used with a different request.",
		}, 0), nil
	case idempotency.BeginStateNew:
		// Continue below while holding the receipt lease.
	default:
		s.log.Error(ctx, "idempotency service returned an unknown begin state")
		return createStoryFailure(ctx, statusFailure(http.StatusInternalServerError), 0), nil
	}

	story, createErr := s.stories.Create(ctx, input, request.WorkspaceId)
	if createErr != nil {
		problem = classifyFailure(createErr)
		if problem.status == http.StatusNotFound {
			problem = &failure{http.StatusBadRequest, "invalid_reference", "A referenced resource is unavailable in this workspace."}
		}
		response := errorResponse(ctx, problem)
		if err := s.completeCreateStoryReceipt(ctx, begin.Lease, problem.status, response); err != nil {
			s.log.Error(ctx, "failed to complete failed story-create idempotency receipt")
			return createStoryFailure(ctx, statusFailure(http.StatusServiceUnavailable), 1), nil
		}
		return createStoryFailure(ctx, problem, 0), nil
	}

	body := openapiv1.ComponentsResourcesStoryResponse{Data: storyDetailModel(story)}
	if err := s.completeCreateStoryReceipt(ctx, begin.Lease, http.StatusCreated, body); err != nil {
		s.log.Error(ctx, "failed to complete successful story-create idempotency receipt")
		return createStoryFailure(ctx, statusFailure(http.StatusServiceUnavailable), 1), nil
	}
	return openapiv1.CreateStory201JSONResponse{Body: body}, nil
}

func storyCreateInput(body openapiv1.ComponentsResourcesCreateStoryRequest, actor platformauth.Actor, rawKey string) (stories.CoreNewStory, error) {
	priority := "No Priority"
	if body.Priority != nil {
		priority = string(*body.Priority)
	}
	var estimateValue *int16
	if body.EstimateValue != nil {
		value, err := safecast.Int32ToInt16(*body.EstimateValue)
		if err != nil {
			return stories.CoreNewStory{}, err
		}
		estimateValue = &value
	}
	autoScheduling := false
	if body.AutoSchedulingEnabled != nil {
		autoScheduling = *body.AutoSchedulingEnabled
	}
	var labelIDs []uuid.UUID
	if body.LabelIds != nil {
		labelIDs = append(labelIDs, (*body.LabelIds)...)
	}
	creationKey := publicStoryCreationKey(actor, rawKey)
	return stories.CoreNewStory{
		Title: body.Title, Team: body.TeamId, Description: body.Description,
		Parent: body.ParentId, Objective: body.ObjectiveId, Status: body.StatusId,
		Assignee: body.AssigneeId, Sprint: body.SprintId, KeyResult: body.KeyResultId,
		Priority: priority, EstimateValue: estimateValue,
		EstimatedDurationMinutes: body.EstimatedDurationMinutes,
		MinimumFocusBlockMinutes: body.MinimumFocusBlockMinutes,
		AutoSchedulingEnabled:    autoScheduling, LabelIDs: labelIDs,
		StartDate: body.StartDate, EndDate: body.EndDate, CreationKey: &creationKey,
	}, nil
}

func publicStoryCreationKey(actor platformauth.Actor, rawKey string) string {
	digest := sha256.Sum256([]byte(rawKey))
	identityID := actor.PrincipalID
	if actor.Kind == platformauth.PrincipalOAuthApplication {
		identityID = actor.CredentialID
	}
	return fmt.Sprintf("api-v1:%s:%s:%x", actor.Kind, identityID, digest)
}

func (s *server) completeCreateStoryReceipt(ctx context.Context, lease idempotency.Lease, status int, body any) error {
	encoded, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("encode story-create receipt response: %w", err)
	}
	// Generated response writers append a newline through json.Encoder. Store it
	// too so a completed retry reproduces the original response bytes.
	encoded = append(encoded, '\n')
	response, err := idempotency.NewResponse(status, encoded, jsonContentType)
	if err != nil {
		return err
	}
	return s.idempotency.Complete(ctx, lease, response)
}

func replayCreateStoryResponse(replay idempotency.Response) (openapiv1.CreateStoryResponseObject, error) {
	if replay.ContentType() != jsonContentType {
		return nil, errors.New("story-create idempotency receipt has an unsupported content type")
	}
	switch replay.StatusCode() {
	case http.StatusCreated:
		var body openapiv1.ComponentsResourcesStoryResponse
		if err := json.Unmarshal(replay.Body(), &body); err != nil {
			return nil, fmt.Errorf("decode story-create success receipt: %w", err)
		}
		return openapiv1.CreateStory201JSONResponse{Body: body}, nil
	case http.StatusBadRequest, http.StatusForbidden, http.StatusConflict,
		http.StatusInternalServerError, http.StatusServiceUnavailable:
		var body openapiv1.ComponentsCommonErrorResponse
		if err := json.Unmarshal(replay.Body(), &body); err != nil {
			return nil, fmt.Errorf("decode story-create error receipt: %w", err)
		}
		return createStoryErrorResponse(replay.StatusCode(), body, 0)
	default:
		return nil, errors.New("story-create idempotency receipt has an unsupported status")
	}
}

func createStoryFailure(ctx context.Context, problem *failure, retryAfter int) openapiv1.CreateStoryResponseObject {
	response, err := createStoryErrorResponse(problem.status, errorResponse(ctx, problem), retryAfter)
	if err != nil {
		fallback := statusFailure(http.StatusInternalServerError)
		response, _ = createStoryErrorResponse(fallback.status, errorResponse(ctx, fallback), 0)
	}
	return response
}

func createStoryErrorResponse(status int, body openapiv1.ComponentsCommonErrorResponse, retryAfter int) (openapiv1.CreateStoryResponseObject, error) {
	requestID := body.Error.RequestId
	switch status {
	case http.StatusBadRequest:
		return openapiv1.CreateStory400JSONResponse{ComponentsCommonBadRequestJSONResponse: openapiv1.ComponentsCommonBadRequestJSONResponse{
			Body: body, Headers: openapiv1.ComponentsCommonBadRequestResponseHeaders{XRequestID: &requestID},
		}}, nil
	case http.StatusForbidden:
		return openapiv1.CreateStory403JSONResponse{ComponentsCommonForbiddenJSONResponse: openapiv1.ComponentsCommonForbiddenJSONResponse{
			Body: body, Headers: openapiv1.ComponentsCommonForbiddenResponseHeaders{XRequestID: &requestID},
		}}, nil
	case http.StatusConflict:
		var retry *int
		if retryAfter > 0 {
			retry = &retryAfter
		}
		return openapiv1.CreateStory409JSONResponse{ComponentsCommonConflictJSONResponse: openapiv1.ComponentsCommonConflictJSONResponse{
			Body: body, Headers: openapiv1.ComponentsCommonConflictResponseHeaders{XRequestID: &requestID, RetryAfter: retry},
		}}, nil
	case http.StatusServiceUnavailable:
		var retry *int
		if retryAfter > 0 {
			retry = &retryAfter
		}
		return openapiv1.CreateStory503JSONResponse{ComponentsCommonServiceUnavailableJSONResponse: openapiv1.ComponentsCommonServiceUnavailableJSONResponse{
			Body: body, Headers: openapiv1.ComponentsCommonServiceUnavailableResponseHeaders{XRequestID: &requestID, RetryAfter: retry},
		}}, nil
	case http.StatusInternalServerError:
		return openapiv1.CreateStory500JSONResponse{ComponentsCommonInternalErrorJSONResponse: openapiv1.ComponentsCommonInternalErrorJSONResponse{
			Body: body, Headers: openapiv1.ComponentsCommonInternalErrorResponseHeaders{XRequestID: &requestID},
		}}, nil
	default:
		return nil, fmt.Errorf("unsupported story-create response status %d", status)
	}
}

func retryAfterSeconds(retryAt time.Time) int {
	seconds := int(math.Ceil(time.Until(retryAt).Seconds()))
	if seconds < 1 {
		return 1
	}
	return seconds
}
