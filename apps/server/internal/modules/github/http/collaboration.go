package githubhttp

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	github "github.com/complexus-tech/projects-api/internal/modules/github/service"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/internal/platform/idempotency"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

func (h *Handlers) GetStoryGitHubLinks(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	storyID, err := uuid.Parse(web.Params(r, "storyId"))
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	links, err := h.service.GetStoryGitHubLinks(ctx, workspace.ID, storyID)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}
	return web.Respond(ctx, w, links, http.StatusOK)
}

func (h *Handlers) DeleteStoryGitHubLink(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	linkID, err := uuid.Parse(web.Params(r, "linkId"))
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	if err := h.service.DeleteStoryGitHubLink(ctx, workspace.ID, linkID); err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}
	return web.Respond(ctx, w, nil, http.StatusNoContent)
}

func (h *Handlers) GetStoryGitHubComments(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	storyID, err := uuid.Parse(web.Params(r, "storyId"))
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	comments, err := h.service.GetStoryGitHubComments(ctx, workspace.ID, storyID)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}
	return web.Respond(ctx, w, comments, http.StatusOK)
}

func (h *Handlers) PostStoryGitHubComment(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	storyID, err := uuid.Parse(web.Params(r, "storyId"))
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	var input AppPostGitHubCommentRequest
	if err := web.Decode(r, &input); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	authorName := "Someone"
	if h.users != nil {
		name, err := h.users.GetUserName(ctx, userID)
		if err != nil {
			return web.RespondError(ctx, w, err, http.StatusInternalServerError)
		}
		if strings.TrimSpace(name) != "" {
			authorName = strings.TrimSpace(name)
		}
	}
	commentID, err := githubCommentIdempotencyID(r, workspace.ID, userID, storyID, "stories.github-comment.create")
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	if err := h.service.PostCommentToGitHub(ctx, workspace.ID, storyID, userID, commentID, authorName, input.Body); err != nil {
		return web.RespondError(ctx, w, err, githubCollaborationErrorStatus(err))
	}
	return web.Respond(ctx, w, nil, http.StatusOK)
}

func (h *Handlers) GetRequestGitHubComments(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	requestID, err := uuid.Parse(web.Params(r, "requestId"))
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	comments, err := h.service.GetRequestGitHubComments(ctx, workspace.ID, requestID)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}
	return web.Respond(ctx, w, comments, http.StatusOK)
}

func (h *Handlers) PostRequestGitHubComment(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	requestID, err := uuid.Parse(web.Params(r, "requestId"))
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	var input AppPostGitHubCommentRequest
	if err := web.Decode(r, &input); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	authorName := "Someone"
	if h.users != nil {
		name, err := h.users.GetUserName(ctx, userID)
		if err != nil {
			return web.RespondError(ctx, w, err, http.StatusInternalServerError)
		}
		if strings.TrimSpace(name) != "" {
			authorName = strings.TrimSpace(name)
		}
	}
	commentID, err := githubCommentIdempotencyID(r, workspace.ID, userID, requestID, "integration-requests.github-comment.create")
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	if err := h.service.PostRequestCommentToGitHub(ctx, workspace.ID, requestID, userID, commentID, authorName, input.Body); err != nil {
		return web.RespondError(ctx, w, err, githubCollaborationErrorStatus(err))
	}
	return web.Respond(ctx, w, nil, http.StatusOK)
}

func (h *Handlers) CreateUserLinkSession(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	var input AppCreateUserLinkSessionRequest
	if err := web.Decode(r, &input); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	session, err := h.service.CreateUserLinkSession(ctx, userID, input.ReturnTo)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	return web.Respond(ctx, w, AppCreateUserLinkSession{State: session.State}, http.StatusOK)
}

func (h *Handlers) LinkGitHubUser(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	var input AppLinkGitHubUserRequest
	if err := web.Decode(r, &input); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	if err := h.service.LinkGitHubUser(ctx, userID, input.Code, input.State); err != nil {
		return web.RespondError(ctx, w, err, githubCollaborationErrorStatus(err))
	}
	return web.Respond(ctx, w, nil, http.StatusOK)
}

func (h *Handlers) UnlinkGitHubUser(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	if err := h.service.UnlinkGitHubUser(ctx, userID); err != nil {
		return web.RespondError(ctx, w, err, githubCollaborationErrorStatus(err))
	}
	return web.Respond(ctx, w, nil, http.StatusNoContent)
}

func githubCommentIdempotencyID(
	r *http.Request,
	workspaceID, actorID, resourceID uuid.UUID,
	operation string,
) (*uuid.UUID, error) {
	values := r.Header.Values("Idempotency-Key")
	if len(values) == 0 {
		return nil, nil
	}
	if len(values) != 1 {
		return nil, idempotency.ErrInvalidKey
	}
	rawKey := values[0]
	if _, err := idempotency.ParseKey(rawKey); err != nil {
		return nil, err
	}
	identity := strings.Join([]string{
		operation,
		workspaceID.String(),
		actorID.String(),
		resourceID.String(),
		rawKey,
	}, "\x00")
	commentID := uuid.NewSHA1(uuid.NameSpaceOID, []byte(identity))
	if commentID == uuid.Nil {
		return nil, fmt.Errorf("%w: could not derive comment identity", idempotency.ErrInvalidKey)
	}
	return &commentID, nil
}

func githubCollaborationErrorStatus(err error) int {
	switch {
	case errors.Is(err, github.ErrNoLinkedGitHubIssues):
		return http.StatusNotFound
	case errors.Is(err, github.ErrGitHubCommentKeyConflict):
		return http.StatusConflict
	case errors.Is(err, github.ErrGitHubOAuthStateInvalid),
		errors.Is(err, github.ErrGitHubOAuthStateBinding),
		errors.Is(err, github.ErrGitHubOAuthCodeInvalid),
		errors.Is(err, github.ErrGitHubOAuthCodeRejected),
		errors.Is(err, idempotency.ErrInvalidKey):
		return http.StatusBadRequest
	case errors.Is(err, github.ErrGitHubAppNotConfigured),
		errors.Is(err, github.ErrGitHubOAuthNotConfigured),
		errors.Is(err, github.ErrGitHubOAuthExchangeUnavailable),
		errors.Is(err, github.ErrGitHubOAuthRevocationUnavailable):
		return http.StatusServiceUnavailable
	default:
		return http.StatusInternalServerError
	}
}
