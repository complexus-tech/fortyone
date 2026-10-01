package apiv1http

import (
	"context"
	"net/http"

	openapiv1 "github.com/complexus-tech/projects-api/internal/generated/openapi/v1"
	comments "github.com/complexus-tech/projects-api/internal/modules/comments/service"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/google/uuid"
)

func (s *server) CreateStoryComment(ctx context.Context, request openapiv1.CreateStoryCommentRequestObject) (openapiv1.CreateStoryCommentResponseObject, error) {
	actor, problem := actorFor(ctx, request.WorkspaceId, platformauth.ScopeCommentsWrite)
	if problem == nil {
		problem = requireUserCredential(actor)
	}
	if problem == nil && !actor.Scopes.Has(platformauth.ScopeStoriesRead) {
		problem = statusFailure(http.StatusForbidden)
	}
	if problem != nil {
		return mutationFailure(ctx, problem), nil
	}
	if request.Body == nil || request.Body.ClientRequestId == uuid.Nil {
		return mutationFailure(ctx, statusFailure(http.StatusBadRequest)), nil
	}
	// The same client UUID in another workspace, task, or principal has its own
	// identity. A rotated credential for the same user preserves safe retries.
	creationID := uuid.NewSHA1(request.WorkspaceId, []byte("api-comment:"+actor.PrincipalID.String()+":"+request.StoryId.String()+":"+request.Body.ClientRequestId.String()))
	if s.commentWriter == nil {
		return mutationFailure(ctx, statusFailure(http.StatusServiceUnavailable)), nil
	}
	created, err := s.commentWriter.CreateComment(ctx, comments.CreateCommentCommand{
		CreationID: &creationID, WorkspaceID: request.WorkspaceId, StoryID: request.StoryId,
		ParentID: request.Body.ParentId, Actor: actor, Content: request.Body.Content,
	})
	if err != nil {
		return mutationFailure(ctx, classifyFailure(err)), nil
	}
	data := openapiv1.ComponentsResourcesComment{
		Id: created.ID, StoryId: created.StoryID, ParentId: created.Parent, AuthorId: created.UserID,
		Content: created.Comment, CreatedAt: created.CreatedAt, UpdatedAt: created.UpdatedAt,
		Replies: []openapiv1.ComponentsResourcesComment{},
	}
	return openapiv1.CreateStoryComment201JSONResponse{Data: data}, nil
}
