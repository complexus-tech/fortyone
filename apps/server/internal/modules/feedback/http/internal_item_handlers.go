package feedbackhttp

import (
	"context"
	"errors"
	"net/http"

	feedback "github.com/complexus-tech/projects-api/internal/modules/feedback/service"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

func (h *Handlers) ListPortals(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	portals, err := h.feedback.ListPortals(ctx, feedback.CoreWorkspacePortalInput{
		WorkspaceID:   workspace.ID,
		WorkspaceName: workspace.Name,
		WorkspaceSlug: workspace.Slug,
	})
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}
	response := make([]AppPortal, 0, len(portals))
	for _, portal := range portals {
		appPortal := toAppPortal(portal)
		boards, err := h.feedback.ListPortalBoards(ctx, workspace.ID, portal.ID)
		if err != nil {
			return web.RespondError(ctx, w, err, httpStatus(err))
		}
		appPortal.Boards = make([]AppBoard, 0, len(boards))
		for _, board := range boards {
			appPortal.Boards = append(appPortal.Boards, toAppBoard(board))
		}
		response = append(response, appPortal)
	}
	return web.Respond(ctx, w, response, http.StatusOK)
}

func (h *Handlers) ListTeamItems(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	teamID, err := uuid.Parse(web.Params(r, "teamId"))
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	if err := h.authorizeTeam(ctx, workspace.ID, teamID, userID); err != nil {
		return web.RespondError(ctx, w, err, httpStatus(err))
	}
	query, err := parseTeamFeedbackListQuery(r.URL.Query())
	if err != nil {
		return web.RespondError(ctx, w, err, httpStatus(err))
	}
	if query.Status == feedback.ListStatusTrashed && workspace.UserRole != string(mid.RoleAdmin) {
		return web.RespondError(ctx, w, errors.New("you don't have permission to access trashed feedback"), http.StatusForbidden)
	}
	itemsPage, err := h.feedback.ListTeamItems(
		ctx, workspace.ID, teamID, userID, query.Status, query.Search,
		query.Pagination.Page, query.Pagination.PageSize,
	)
	if err != nil {
		return web.RespondError(ctx, w, err, httpStatus(err))
	}
	resolvedByAvatar := make(map[string]*string)
	items := make([]AppItem, 0, len(itemsPage.Items))
	for _, item := range itemsPage.Items {
		item.AuthorAvatar = h.resolveAuthorAvatar(ctx, item.AuthorAvatar, resolvedByAvatar)
		links := make([]AppStoryLink, 0, len(item.StoryLinks))
		for _, link := range item.StoryLinks {
			links = append(links, toAppStoryLink(link))
		}
		items = append(items, toAppItem(item, []AppComment{}, links))
	}
	return web.Respond(ctx, w, AppTeamFeedbackResponse{
		Feedback: items,
		Pagination: AppItemsPagination{
			Page:     query.Pagination.Page,
			PageSize: query.Pagination.PageSize,
			HasMore:  itemsPage.HasMore,
			NextPage: query.Pagination.Page + 1,
		},
	}, http.StatusOK)
}

func (h *Handlers) TrashItem(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	itemID, err := uuid.Parse(web.Params(r, "itemId"))
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	if err := h.feedback.TrashItem(ctx, workspace.ID, itemID); err != nil {
		return web.RespondError(ctx, w, err, httpStatus(err))
	}
	return web.Respond(ctx, w, nil, http.StatusNoContent)
}

func (h *Handlers) RestoreItem(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	itemID, err := uuid.Parse(web.Params(r, "itemId"))
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	if err := h.feedback.RestoreItem(ctx, workspace.ID, itemID); err != nil {
		return web.RespondError(ctx, w, err, httpStatus(err))
	}
	return web.Respond(ctx, w, nil, http.StatusNoContent)
}

func (h *Handlers) GetItem(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	itemID, err := uuid.Parse(web.Params(r, "itemId"))
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	details, err := h.feedback.GetItemDetails(ctx, workspace.ID, itemID, userID)
	if err != nil {
		return web.RespondError(ctx, w, err, httpStatus(err))
	}
	if err := h.authorizeTeam(ctx, workspace.ID, details.Item.Board.TeamID, userID); err != nil {
		return web.RespondError(ctx, w, err, httpStatus(err))
	}
	resolvedByAvatar := make(map[string]*string)
	details.Item.AuthorAvatar = h.resolveAuthorAvatar(ctx, details.Item.AuthorAvatar, resolvedByAvatar)
	comments := make([]AppComment, 0, len(details.Comments))
	for _, comment := range details.Comments {
		comment.AuthorAvatar = h.resolveAuthorAvatar(ctx, comment.AuthorAvatar, resolvedByAvatar)
		comments = append(comments, toAppComment(comment))
	}
	links := make([]AppStoryLink, 0, len(details.StoryLinks))
	for _, link := range details.StoryLinks {
		links = append(links, toAppStoryLink(link))
	}
	return web.Respond(ctx, w, toAppItem(details.Item, comments, links), http.StatusOK)
}

func (h *Handlers) GetPrivateAuthor(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	itemID, err := uuid.Parse(web.Params(r, "itemId"))
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	author, err := h.feedback.GetPrivateAuthor(ctx, workspace.ID, itemID)
	if err != nil {
		return web.RespondError(ctx, w, err, httpStatus(err))
	}
	return web.Respond(ctx, w, AppPrivateAuthor{
		ContributorID: author.ContributorID,
		UserID:        author.UserID,
		Kind:          author.Kind,
		DisplayName:   author.DisplayName,
		Email:         author.Email,
		AvatarURL:     author.AvatarURL,
		PublicMasked:  author.PublicMasked,
	}, http.StatusOK)
}

func (h *Handlers) ResolveCanonicalItem(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	canonical, err := h.feedback.ResolveCanonicalItem(ctx, web.Params(r, "portalSlug"), web.Params(r, "itemReference"))
	if err != nil {
		return web.RespondError(ctx, w, err, httpStatus(err))
	}
	return web.Respond(ctx, w, AppCanonicalItem{
		ItemID: canonical.ItemID, ItemSlug: canonical.ItemSlug, Merged: canonical.Merged,
	}, http.StatusOK)
}

func (h *Handlers) ListTeamSummaries(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	summaries, err := h.feedback.ListTeamSummaries(ctx, workspace.ID, userID)
	if err != nil {
		return web.RespondError(ctx, w, err, httpStatus(err))
	}
	response := make([]AppTeamFeedbackSummary, 0, len(summaries))
	for _, summary := range summaries {
		response = append(response, AppTeamFeedbackSummary{
			TeamID:      summary.TeamID,
			Enabled:     summary.Enabled,
			TotalCount:  summary.TotalCount,
			UnreadCount: summary.UnreadCount,
		})
	}
	return web.Respond(ctx, w, response, http.StatusOK)
}

func (h *Handlers) MarkItemRead(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	return h.updateItemReadState(ctx, w, r, true)
}

func (h *Handlers) MarkItemUnread(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	return h.updateItemReadState(ctx, w, r, false)
}

func (h *Handlers) updateItemReadState(ctx context.Context, w http.ResponseWriter, r *http.Request, read bool) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	itemID, err := uuid.Parse(web.Params(r, "itemId"))
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	if err := h.authorizeItemTeam(ctx, workspace.ID, itemID, userID); err != nil {
		return web.RespondError(ctx, w, err, httpStatus(err))
	}
	if !read {
		if err := h.feedback.MarkItemUnread(ctx, workspace.ID, itemID, userID); err != nil {
			return web.RespondError(ctx, w, err, httpStatus(err))
		}
		return web.Respond(ctx, w, AppFeedbackReadState{}, http.StatusOK)
	}
	readAt, err := h.feedback.MarkItemRead(ctx, workspace.ID, itemID, userID)
	if err != nil {
		return web.RespondError(ctx, w, err, httpStatus(err))
	}
	return web.Respond(ctx, w, AppFeedbackReadState{ReadAt: readAt}, http.StatusOK)
}

func (h *Handlers) GetStoryFeedbackLinks(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	storyID, err := uuid.Parse(web.Params(r, "storyId"))
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	links, err := h.feedback.ListStoryFeedbackLinks(ctx, workspace.ID, storyID)
	if err != nil {
		return web.RespondError(ctx, w, err, httpStatus(err))
	}
	authorizedTeams := make(map[uuid.UUID]struct{})
	response := make([]AppStoryFeedbackLink, 0, len(links))
	for _, link := range links {
		if _, authorized := authorizedTeams[link.TeamID]; !authorized {
			if err := h.authorizeTeam(ctx, workspace.ID, link.TeamID, userID); err != nil {
				if errors.Is(err, feedback.ErrNotFound) {
					continue
				}
				return web.RespondError(ctx, w, err, httpStatus(err))
			}
			authorizedTeams[link.TeamID] = struct{}{}
		}
		response = append(response, toAppStoryFeedbackLink(link))
	}
	return web.Respond(ctx, w, response, http.StatusOK)
}
