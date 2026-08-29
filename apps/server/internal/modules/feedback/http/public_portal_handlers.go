package feedbackhttp

import (
	"context"
	"fmt"
	"net/http"

	feedback "github.com/complexus-tech/projects-api/internal/modules/feedback/service"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

func (h *Handlers) GetPortal(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	slug := web.Params(r, "portalSlug")
	input, err := parsePublicItemsQuery(r.URL.Query())
	if err != nil {
		return web.RespondError(ctx, w, err, httpStatus(err))
	}
	portal, err := h.feedback.GetPortalSnapshot(ctx, slug, input)
	if err != nil {
		return web.RespondError(ctx, w, err, httpStatus(err))
	}
	h.resolvePortalAvatars(ctx, &portal)
	return web.Respond(ctx, w, toAppPortalSnapshot(portal), http.StatusOK)
}

func (h *Handlers) GetWorkspacePortal(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspaceSlug := web.Params(r, "workspaceSlug")
	portalSlug := web.Params(r, "portalSlug")
	input, err := parsePublicItemsQuery(r.URL.Query())
	if err != nil {
		return web.RespondError(ctx, w, err, httpStatus(err))
	}
	portal, err := h.feedback.GetWorkspacePortalSnapshot(ctx, workspaceSlug, portalSlug, input)
	if err != nil {
		return web.RespondError(ctx, w, err, httpStatus(err))
	}
	h.resolvePortalAvatars(ctx, &portal)
	return web.Respond(ctx, w, toAppPortalSnapshot(portal), http.StatusOK)
}

func (h *Handlers) ListPublicSimilarItems(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	query, err := parsePublicSimilarItemsQuery(r.URL.Query())
	if err != nil {
		return web.RespondError(ctx, w, err, httpStatus(err))
	}
	items, err := h.feedback.ListPublicSimilarItems(
		ctx, web.Params(r, "portalSlug"), query.Title, query.Description, query.Limit,
	)
	if err != nil {
		return web.RespondError(ctx, w, err, httpStatus(err))
	}
	h.resolveSimilarItemAvatars(ctx, items)
	response := make([]AppSimilarItem, 0, len(items))
	for _, item := range items {
		response = append(response, toAppSimilarItem(item))
	}
	return web.Respond(ctx, w, response, http.StatusOK)
}

func (h *Handlers) GetPublicContributor(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	authorID, err := publicContributorID(r)
	if err != nil {
		return web.RespondError(ctx, w, err, httpStatus(err))
	}
	contributor, err := h.feedback.GetPublicContributor(ctx, web.Params(r, "portalSlug"), authorID)
	if err != nil {
		return web.RespondError(ctx, w, err, httpStatus(err))
	}
	return h.respondPublicContributor(ctx, w, contributor)
}

func (h *Handlers) ListContributorActivity(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	query, err := parseContributorActivityQuery(r.URL.Query())
	if err != nil {
		return web.RespondError(ctx, w, err, httpStatus(err))
	}
	activity, err := h.feedback.ListContributorActivity(
		ctx, userID, query.ActivityType, query.Pagination.Page, query.Pagination.PageSize,
	)
	if err != nil {
		return web.RespondError(ctx, w, err, httpStatus(err))
	}
	return web.Respond(ctx, w, toAppContributorActivityPage(activity), http.StatusOK)
}

func (h *Handlers) GetWorkspacePublicContributor(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	authorID, err := publicContributorID(r)
	if err != nil {
		return web.RespondError(ctx, w, err, httpStatus(err))
	}
	contributor, err := h.feedback.GetWorkspacePublicContributor(ctx, web.Params(r, "workspaceSlug"), web.Params(r, "portalSlug"), authorID)
	if err != nil {
		return web.RespondError(ctx, w, err, httpStatus(err))
	}
	return h.respondPublicContributor(ctx, w, contributor)
}

func (h *Handlers) respondPublicContributor(ctx context.Context, w http.ResponseWriter, contributor feedback.CoreContributor) error {
	contributor.AvatarURL = h.resolveAuthorAvatar(ctx, contributor.AvatarURL, make(map[string]*string))
	return web.Respond(ctx, w, toAppContributor(contributor), http.StatusOK)
}

func (h *Handlers) ListPublicContributorComments(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	authorID, err := publicContributorID(r)
	if err != nil {
		return web.RespondError(ctx, w, err, httpStatus(err))
	}
	params, err := parseContributorPagination(r.URL.Query())
	if err != nil {
		return web.RespondError(ctx, w, err, httpStatus(err))
	}
	comments, err := h.feedback.ListPublicContributorComments(
		ctx, web.Params(r, "portalSlug"), authorID, params.Page, params.PageSize,
	)
	if err != nil {
		return web.RespondError(ctx, w, err, httpStatus(err))
	}
	return respondPublicContributorComments(ctx, w, comments)
}

func (h *Handlers) ListWorkspacePublicContributorComments(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	authorID, err := publicContributorID(r)
	if err != nil {
		return web.RespondError(ctx, w, err, httpStatus(err))
	}
	params, err := parseContributorPagination(r.URL.Query())
	if err != nil {
		return web.RespondError(ctx, w, err, httpStatus(err))
	}
	comments, err := h.feedback.ListWorkspacePublicContributorComments(
		ctx, web.Params(r, "workspaceSlug"), web.Params(r, "portalSlug"), authorID,
		params.Page, params.PageSize,
	)
	if err != nil {
		return web.RespondError(ctx, w, err, httpStatus(err))
	}
	return respondPublicContributorComments(ctx, w, comments)
}

func respondPublicContributorComments(ctx context.Context, w http.ResponseWriter, page feedback.CoreContributorCommentsPage) error {
	comments := make([]AppContributorComment, 0, len(page.Comments))
	for _, comment := range page.Comments {
		comments = append(comments, toAppContributorComment(comment))
	}
	nextPage := 0
	if page.HasMore {
		nextPage = page.Page + 1
	}
	return web.Respond(ctx, w, AppContributorCommentsResponse{Comments: comments, Pagination: AppItemsPagination{Page: page.Page, PageSize: page.PageSize, HasMore: page.HasMore, NextPage: nextPage}}, http.StatusOK)
}

func publicContributorID(r *http.Request) (uuid.UUID, error) {
	authorID, err := uuid.Parse(web.Params(r, "authorId"))
	if err != nil || authorID == uuid.Nil {
		return uuid.Nil, fmt.Errorf("%w: contributor id must be a non-nil UUID", feedback.ErrInvalidInput)
	}
	return authorID, nil
}
