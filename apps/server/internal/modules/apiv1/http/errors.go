package apiv1http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	openapiv1 "github.com/complexus-tech/projects-api/internal/generated/openapi/v1"
	keyresults "github.com/complexus-tech/projects-api/internal/modules/keyresults/service"
	labels "github.com/complexus-tech/projects-api/internal/modules/labels/service"
	objectives "github.com/complexus-tech/projects-api/internal/modules/objectives/service"
	outboundwebhooksdomain "github.com/complexus-tech/projects-api/internal/modules/outboundwebhooks/domain"
	sprints "github.com/complexus-tech/projects-api/internal/modules/sprints/service"
	states "github.com/complexus-tech/projects-api/internal/modules/states/service"
	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	teams "github.com/complexus-tech/projects-api/internal/modules/teams/service"
	workspaces "github.com/complexus-tech/projects-api/internal/modules/workspaces/service"
	"github.com/complexus-tech/projects-api/internal/platform/authorization"
	"github.com/complexus-tech/projects-api/pkg/web"
)

type failure struct {
	status  int
	code    string
	message string
}

func classifyFailure(err error) *failure {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, workspaces.ErrNotFound), errors.Is(err, teams.ErrTeamNotFound),
		errors.Is(err, stories.ErrNotFound), errors.Is(err, labels.ErrNotFound),
		errors.Is(err, states.ErrNotFound), errors.Is(err, sprints.ErrNotFound),
		errors.Is(err, objectives.ErrNotFound), errors.Is(err, keyresults.ErrNotFound),
		errors.Is(err, outboundwebhooksdomain.ErrEndpointNotFound):
		return &failure{http.StatusNotFound, "resource_not_found", "The requested resource was not found."}
	case errors.Is(err, stories.ErrStoryReadForbidden),
		errors.Is(err, stories.ErrStoryMutationForbidden),
		errors.Is(err, authorization.ErrWorkspaceAdminRequired),
		errors.Is(err, authorization.ErrInsufficientWorkspaceRole),
		errors.Is(err, sprints.ErrForbidden), errors.Is(err, objectives.ErrForbidden),
		errors.Is(err, keyresults.ErrForbidden),
		errors.Is(err, outboundwebhooksdomain.ErrEndpointOwnerInactive):
		return &failure{http.StatusForbidden, "access_denied", "The credential is not allowed to perform this operation."}
	case errors.Is(err, stories.ErrInvalidStoryReadScope), errors.Is(err, stories.ErrInvalidStoryReadQuery),
		errors.Is(err, labels.ErrInvalidPagination), errors.Is(err, sprints.ErrInvalid),
		errors.Is(err, objectives.ErrInvalid), errors.Is(err, keyresults.ErrInvalid),
		errors.Is(err, stories.ErrInvalidStoryMutation),
		errors.Is(err, stories.ErrInvalidStoryReference),
		errors.Is(err, stories.ErrObjectiveKeyResultMismatch),
		errors.Is(err, stories.ErrInvalidStoryLabels),
		errors.Is(err, stories.ErrInvalidEstimatedDuration),
		errors.Is(err, stories.ErrInvalidMinimumFocusBlock),
		errors.Is(err, stories.ErrEstimatedDurationTooLarge),
		errors.Is(err, stories.ErrMinimumFocusBlockTooLarge),
		errors.Is(err, stories.ErrFocusBlockRequiresDuration),
		errors.Is(err, stories.ErrFocusBlockExceedsDuration),
		errors.Is(err, stories.ErrAutoSchedulingUnavailable),
		errors.Is(err, stories.ErrMayaAssignmentRequiresScheduling),
		errors.Is(err, stories.ErrMayaAssignmentRequiresDuration),
		errors.Is(err, stories.ErrMayaAssignmentRequiresDeliveryDate),
		errors.Is(err, outboundwebhooksdomain.ErrInvalidEndpoint),
		errors.Is(err, outboundwebhooksdomain.ErrInvalidSubscription),
		errors.Is(err, outboundwebhooksdomain.ErrInvalidEventType):
		return &failure{http.StatusBadRequest, "invalid_request", "The request is invalid."}
	case errors.Is(err, stories.ErrStoryChanged),
		errors.Is(err, outboundwebhooksdomain.ErrEndpointConflict), errors.Is(err, outboundwebhooksdomain.ErrEndpointDisabled):
		return &failure{http.StatusConflict, "resource_conflict", "The resource changed or is not in the required state."}
	default:
		return &failure{http.StatusInternalServerError, "internal_error", "The request could not be completed."}
	}
}

func errorResponse(ctx context.Context, problem *failure) openapiv1.ComponentsCommonErrorResponse {
	return openapiv1.ComponentsCommonErrorResponse{Error: openapiv1.ComponentsCommonError{
		Code: problem.code, Message: problem.message, RequestId: web.GetRequestID(ctx),
	}}
}

func writeAPIError(ctx context.Context, writer http.ResponseWriter, _ error, status int) error {
	problem := statusFailure(status)
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(problem.status)
	return json.NewEncoder(writer).Encode(errorResponse(ctx, problem))
}

func statusFailure(status int) *failure {
	switch status {
	case http.StatusBadRequest:
		return &failure{status, "invalid_request", "The request is invalid."}
	case http.StatusUnauthorized:
		return &failure{status, "machine_authentication_required", "A valid machine bearer credential is required."}
	case http.StatusForbidden:
		return &failure{status, "access_denied", "The credential is not allowed to perform this operation."}
	case http.StatusNotFound:
		return &failure{status, "resource_not_found", "The requested resource was not found."}
	case http.StatusMethodNotAllowed:
		return &failure{status, "method_not_allowed", "The request method is not supported."}
	case http.StatusRequestEntityTooLarge:
		return &failure{status, "request_too_large", "The request body is too large."}
	case http.StatusUnsupportedMediaType:
		return &failure{status, "unsupported_media_type", "Content-Type must be application/json."}
	case http.StatusTooManyRequests:
		return &failure{status, "rate_limit_exceeded", "The credential request limit was exceeded."}
	case http.StatusServiceUnavailable:
		return &failure{status, "service_unavailable", "The service is temporarily unavailable."}
	default:
		return &failure{http.StatusInternalServerError, "internal_error", "The request could not be completed."}
	}
}
