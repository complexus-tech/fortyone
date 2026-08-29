package feedbackhttp

import (
	"context"
	"net/http"

	feedback "github.com/complexus-tech/projects-api/internal/modules/feedback/service"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

func (h *Handlers) UpdatePortal(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	portalID, err := uuid.Parse(web.Params(r, "portalId"))
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	var input AppUpdatePortal
	if err := web.Decode(r, &input); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	portal, err := h.feedback.UpdatePortal(ctx, workspace.ID, portalID, feedback.CorePortalInput{
		IsPublic:            input.IsPublic,
		ParticipationMode:   input.ParticipationMode,
		GuestIdentityPolicy: input.GuestIdentityPolicy,
	})
	if err != nil {
		return web.RespondError(ctx, w, err, httpStatus(err))
	}
	return web.Respond(ctx, w, toAppPortal(portal), http.StatusOK)
}

func (h *Handlers) CreateBoard(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	var input AppCreateBoard
	if err := web.Decode(r, &input); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	board, err := h.feedback.CreateBoard(ctx, feedback.CoreBoardInput{
		WorkspaceID: workspace.ID,
		PortalID:    input.PortalID,
		TeamID:      input.TeamID,
		CreatorID:   userID,
		Name:        input.Name,
		Slug:        input.Slug,
		Color:       input.Color,
		OrderIndex:  input.OrderIndex,
	})
	if err != nil {
		return web.RespondError(ctx, w, err, httpStatus(err))
	}
	return web.Respond(ctx, w, toAppBoard(board), http.StatusCreated)
}

func (h *Handlers) DeleteBoard(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	boardID, err := uuid.Parse(web.Params(r, "boardId"))
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	if err := h.feedback.DeleteBoard(ctx, workspace.ID, boardID); err != nil {
		return web.RespondError(ctx, w, err, httpStatus(err))
	}
	return web.Respond(ctx, w, nil, http.StatusNoContent)
}

func (h *Handlers) ListBoardReviewers(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	boardID, err := uuid.Parse(web.Params(r, "boardId"))
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	reviewers, err := h.feedback.ListBoardReviewers(ctx, workspace.ID, boardID)
	if err != nil {
		return web.RespondError(ctx, w, err, httpStatus(err))
	}
	response := make([]AppBoardReviewer, 0, len(reviewers))
	resolvedByAvatar := make(map[string]*string)
	for _, reviewer := range reviewers {
		reviewer.AvatarURL = h.resolveAuthorAvatar(ctx, reviewer.AvatarURL, resolvedByAvatar)
		response = append(response, toAppBoardReviewer(reviewer))
	}
	return web.Respond(ctx, w, response, http.StatusOK)
}

func (h *Handlers) SetBoardReviewer(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	boardID, err := uuid.Parse(web.Params(r, "boardId"))
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	userID, err := uuid.Parse(web.Params(r, "userId"))
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	var input AppSetBoardReviewer
	if err := web.Decode(r, &input); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	reviewer, err := h.feedback.SetBoardReviewer(ctx, feedback.CoreBoardReviewerInput{
		WorkspaceID:    workspace.ID,
		BoardID:        boardID,
		UserID:         userID,
		EmailFrequency: input.EmailFrequency,
	})
	if err != nil {
		return web.RespondError(ctx, w, err, httpStatus(err))
	}
	reviewer.AvatarURL = h.resolveAuthorAvatar(ctx, reviewer.AvatarURL, make(map[string]*string))
	return web.Respond(ctx, w, toAppBoardReviewer(reviewer), http.StatusOK)
}
