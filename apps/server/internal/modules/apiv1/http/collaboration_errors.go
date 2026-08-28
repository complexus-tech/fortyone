package apiv1http

import (
	"context"
	"net/http"

	openapiv1 "github.com/complexus-tech/projects-api/internal/generated/openapi/v1"
)

func listLabelsFailure(ctx context.Context, problem *failure) openapiv1.ListLabelsResponseObject {
	body, headers := commonError(ctx, problem)
	switch problem.status {
	case http.StatusBadRequest:
		return openapiv1.ListLabels400JSONResponse{ComponentsCommonBadRequestJSONResponse: badRequest(body, headers)}
	case http.StatusForbidden:
		return openapiv1.ListLabels403JSONResponse{ComponentsCommonForbiddenJSONResponse: forbidden(body, headers)}
	default:
		return openapiv1.ListLabels500JSONResponse{ComponentsCommonInternalErrorJSONResponse: internalError(body, headers)}
	}
}

func listWorkflowStatesFailure(ctx context.Context, problem *failure) openapiv1.ListWorkflowStatesResponseObject {
	body, headers := commonError(ctx, problem)
	switch problem.status {
	case http.StatusBadRequest:
		return openapiv1.ListWorkflowStates400JSONResponse{ComponentsCommonBadRequestJSONResponse: badRequest(body, headers)}
	case http.StatusForbidden:
		return openapiv1.ListWorkflowStates403JSONResponse{ComponentsCommonForbiddenJSONResponse: forbidden(body, headers)}
	default:
		return openapiv1.ListWorkflowStates500JSONResponse{ComponentsCommonInternalErrorJSONResponse: internalError(body, headers)}
	}
}

func listSprintsFailure(ctx context.Context, problem *failure) openapiv1.ListSprintsResponseObject {
	body, headers := commonError(ctx, problem)
	switch problem.status {
	case http.StatusBadRequest:
		return openapiv1.ListSprints400JSONResponse{ComponentsCommonBadRequestJSONResponse: badRequest(body, headers)}
	case http.StatusForbidden:
		return openapiv1.ListSprints403JSONResponse{ComponentsCommonForbiddenJSONResponse: forbidden(body, headers)}
	default:
		return openapiv1.ListSprints500JSONResponse{ComponentsCommonInternalErrorJSONResponse: internalError(body, headers)}
	}
}

func listObjectivesFailure(ctx context.Context, problem *failure) openapiv1.ListObjectivesResponseObject {
	body, headers := commonError(ctx, problem)
	switch problem.status {
	case http.StatusBadRequest:
		return openapiv1.ListObjectives400JSONResponse{ComponentsCommonBadRequestJSONResponse: badRequest(body, headers)}
	case http.StatusForbidden:
		return openapiv1.ListObjectives403JSONResponse{ComponentsCommonForbiddenJSONResponse: forbidden(body, headers)}
	default:
		return openapiv1.ListObjectives500JSONResponse{ComponentsCommonInternalErrorJSONResponse: internalError(body, headers)}
	}
}

func listKeyResultsFailure(ctx context.Context, problem *failure) openapiv1.ListKeyResultsResponseObject {
	body, headers := commonError(ctx, problem)
	switch problem.status {
	case http.StatusBadRequest:
		return openapiv1.ListKeyResults400JSONResponse{ComponentsCommonBadRequestJSONResponse: badRequest(body, headers)}
	case http.StatusForbidden:
		return openapiv1.ListKeyResults403JSONResponse{ComponentsCommonForbiddenJSONResponse: forbidden(body, headers)}
	default:
		return openapiv1.ListKeyResults500JSONResponse{ComponentsCommonInternalErrorJSONResponse: internalError(body, headers)}
	}
}

func listStoryCommentsFailure(ctx context.Context, problem *failure) openapiv1.ListStoryCommentsResponseObject {
	body, headers := commonError(ctx, problem)
	switch problem.status {
	case http.StatusBadRequest:
		return openapiv1.ListStoryComments400JSONResponse{ComponentsCommonBadRequestJSONResponse: badRequest(body, headers)}
	case http.StatusForbidden:
		return openapiv1.ListStoryComments403JSONResponse{ComponentsCommonForbiddenJSONResponse: forbidden(body, headers)}
	case http.StatusNotFound:
		return openapiv1.ListStoryComments404JSONResponse{ComponentsCommonNotFoundJSONResponse: notFound(body, headers)}
	default:
		return openapiv1.ListStoryComments500JSONResponse{ComponentsCommonInternalErrorJSONResponse: internalError(body, headers)}
	}
}

func getStoryCommentFailure(ctx context.Context, problem *failure) openapiv1.GetStoryCommentResponseObject {
	body, headers := commonError(ctx, problem)
	switch problem.status {
	case http.StatusBadRequest:
		return openapiv1.GetStoryComment400JSONResponse{ComponentsCommonBadRequestJSONResponse: badRequest(body, headers)}
	case http.StatusForbidden:
		return openapiv1.GetStoryComment403JSONResponse{ComponentsCommonForbiddenJSONResponse: forbidden(body, headers)}
	case http.StatusNotFound:
		return openapiv1.GetStoryComment404JSONResponse{ComponentsCommonNotFoundJSONResponse: notFound(body, headers)}
	default:
		return openapiv1.GetStoryComment500JSONResponse{ComponentsCommonInternalErrorJSONResponse: internalError(body, headers)}
	}
}

func badRequest(
	body openapiv1.ComponentsCommonErrorResponse,
	headers openapiv1.ComponentsCommonBadRequestResponseHeaders,
) openapiv1.ComponentsCommonBadRequestJSONResponse {
	return openapiv1.ComponentsCommonBadRequestJSONResponse{Body: body, Headers: headers}
}

func forbidden(
	body openapiv1.ComponentsCommonErrorResponse,
	headers openapiv1.ComponentsCommonBadRequestResponseHeaders,
) openapiv1.ComponentsCommonForbiddenJSONResponse {
	return openapiv1.ComponentsCommonForbiddenJSONResponse{
		Body: body, Headers: openapiv1.ComponentsCommonForbiddenResponseHeaders(headers),
	}
}

func notFound(
	body openapiv1.ComponentsCommonErrorResponse,
	headers openapiv1.ComponentsCommonBadRequestResponseHeaders,
) openapiv1.ComponentsCommonNotFoundJSONResponse {
	return openapiv1.ComponentsCommonNotFoundJSONResponse{
		Body: body, Headers: openapiv1.ComponentsCommonNotFoundResponseHeaders(headers),
	}
}

func internalError(
	body openapiv1.ComponentsCommonErrorResponse,
	headers openapiv1.ComponentsCommonBadRequestResponseHeaders,
) openapiv1.ComponentsCommonInternalErrorJSONResponse {
	return openapiv1.ComponentsCommonInternalErrorJSONResponse{
		Body: body, Headers: openapiv1.ComponentsCommonInternalErrorResponseHeaders(headers),
	}
}
