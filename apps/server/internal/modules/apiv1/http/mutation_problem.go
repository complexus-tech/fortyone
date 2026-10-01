package apiv1http

import (
	"context"
	"encoding/json"
	"net/http"

	openapiv1 "github.com/complexus-tech/projects-api/internal/generated/openapi/v1"
)

// mutationProblem shares the canonical error envelope across the new writes.
// It implements each generated response port without duplicating header logic.
type mutationProblem struct {
	status int
	body   openapiv1.ComponentsCommonErrorResponse
}

func mutationFailure(ctx context.Context, problem *failure) mutationProblem {
	return mutationProblem{status: problem.status, body: errorResponse(ctx, problem)}
}

func (problem mutationProblem) VisitUpdateStoryResponse(writer http.ResponseWriter) error {
	return problem.write(writer)
}

func (problem mutationProblem) VisitCreateStoryCommentResponse(writer http.ResponseWriter) error {
	return problem.write(writer)
}

func (problem mutationProblem) write(writer http.ResponseWriter) error {
	writer.Header().Set("Content-Type", jsonContentType)
	writer.Header().Set("X-Request-ID", problem.body.Error.RequestId)
	writer.WriteHeader(problem.status)
	return json.NewEncoder(writer).Encode(problem.body)
}
