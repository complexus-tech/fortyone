package feedbackhttp

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	feedback "github.com/complexus-tech/projects-api/internal/modules/feedback/service"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

const (
	avatarAccessURLExpiry          = 24 * time.Hour
	publicFeedbackItemBodyLimit    = 128 << 10
	publicFeedbackCommentBodyLimit = 64 << 10
	publicFeedbackVoteBodyLimit    = 4 << 10
)

func decodePublicRequest(w http.ResponseWriter, r *http.Request, input any, bodyLimit int64) (int, error) {
	r.Body = http.MaxBytesReader(w, r.Body, bodyLimit)
	if err := web.Decode(r, input); err != nil {
		if errors.Is(err, web.ErrRequestBodyTooLarge) {
			return http.StatusRequestEntityTooLarge, err
		}
		return http.StatusBadRequest, err
	}
	return 0, nil
}

type profileImageResolver interface {
	ResolveProfileImageURL(ctx context.Context, avatar string, expiry time.Duration) (string, error)
}

type Handlers struct {
	feedback      *feedback.Service
	profileImages profileImageResolver
	log           *logger.Logger
}

func New(service *feedback.Service, profileImages profileImageResolver, log *logger.Logger) *Handlers {
	return &Handlers{feedback: service, profileImages: profileImages, log: log}
}

func (h *Handlers) resolveAuthorAvatar(
	ctx context.Context,
	avatar *string,
	resolvedByAvatar map[string]*string,
) *string {
	if avatar == nil {
		return nil
	}

	avatarKey := strings.TrimSpace(*avatar)
	if avatarKey == "" {
		return nil
	}
	if resolved, ok := resolvedByAvatar[avatarKey]; ok {
		return resolved
	}
	if h.profileImages == nil {
		resolvedByAvatar[avatarKey] = nil
		return nil
	}

	resolved, err := h.profileImages.ResolveProfileImageURL(ctx, avatarKey, avatarAccessURLExpiry)
	if err != nil {
		if h.log != nil {
			h.log.Warn(ctx, "failed to resolve feedback author avatar", "error", err)
		}
		resolvedByAvatar[avatarKey] = nil
		return nil
	}
	if strings.TrimSpace(resolved) == "" {
		resolvedByAvatar[avatarKey] = nil
		return nil
	}

	resolvedByAvatar[avatarKey] = &resolved
	return &resolved
}

func (h *Handlers) resolvePortalAvatars(ctx context.Context, portal *feedback.CorePortalSnapshot) {
	resolvedByAvatar := make(map[string]*string)
	itemIDs := make(map[uuid.UUID]struct{}, len(portal.Items))

	for i := range portal.Items {
		itemIDs[portal.Items[i].ID] = struct{}{}
		portal.Items[i].AuthorAvatar = h.resolveAuthorAvatar(ctx, portal.Items[i].AuthorAvatar, resolvedByAvatar)
	}
	for i := range portal.Comments {
		if _, visible := itemIDs[portal.Comments[i].ItemID]; !visible {
			continue
		}
		portal.Comments[i].AuthorAvatar = h.resolveAuthorAvatar(ctx, portal.Comments[i].AuthorAvatar, resolvedByAvatar)
	}
}

func (h *Handlers) GetPortal(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	slug := web.Params(r, "portalSlug")
	portal, err := h.feedback.GetPortalSnapshot(ctx, slug)
	if err != nil {
		return web.RespondError(ctx, w, err, httpStatus(err))
	}
	if err := h.applyItemsQuery(ctx, r, &portal); err != nil {
		return web.RespondError(ctx, w, err, httpStatus(err))
	}
	h.resolvePortalAvatars(ctx, &portal)
	return web.Respond(ctx, w, toAppPortalSnapshot(portal), http.StatusOK)
}

func (h *Handlers) GetWorkspacePortal(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspaceSlug := web.Params(r, "workspaceSlug")
	portalSlug := web.Params(r, "portalSlug")
	portal, err := h.feedback.GetWorkspacePortalSnapshot(ctx, workspaceSlug, portalSlug)
	if err != nil {
		return web.RespondError(ctx, w, err, httpStatus(err))
	}
	if err := h.applyItemsQuery(ctx, r, &portal); err != nil {
		return web.RespondError(ctx, w, err, httpStatus(err))
	}
	h.resolvePortalAvatars(ctx, &portal)
	return web.Respond(ctx, w, toAppPortalSnapshot(portal), http.StatusOK)
}

func (h *Handlers) applyItemsQuery(ctx context.Context, r *http.Request, portal *feedback.CorePortalSnapshot) error {
	query := r.URL.Query()
	page, _ := strconv.Atoi(query.Get("page"))
	pageSize, _ := strconv.Atoi(query.Get("pageSize"))
	input := feedback.CoreListItemsInput{
		PortalID: portal.Portal.ID,
		Status:   query.Get("status"),
		Search:   query.Get("search"),
		Sort:     query.Get("sort"),
		Page:     page,
		PageSize: pageSize,
	}
	if boardID := query.Get("boardId"); boardID != "" {
		parsed, err := uuid.Parse(boardID)
		if err != nil {
			return err
		}
		input.BoardID = &parsed
	}
	items, err := h.feedback.ListItems(ctx, input)
	if err != nil {
		return err
	}
	portal.Items = items.Items
	portal.ItemsHasMore = items.HasMore
	return nil
}

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
		Description: input.Description,
		IsPublic:    input.IsPublic,
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
	var input AppCreateBoard
	if err := web.Decode(r, &input); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	board, err := h.feedback.CreateBoard(ctx, feedback.CoreBoardInput{
		WorkspaceID: workspace.ID,
		PortalID:    input.PortalID,
		TeamID:      input.TeamID,
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

func (h *Handlers) CreateItem(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	var input AppCreateItem
	if err := web.Decode(r, &input); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	item, err := h.feedback.CreateItem(ctx, feedback.CoreItemInput{
		WorkspaceID: workspace.ID,
		PortalID:    input.PortalID,
		BoardID:     input.BoardID,
		AuthorID:    userID,
		Title:       input.Title,
		Description: input.Description,
	})
	if err != nil {
		return web.RespondError(ctx, w, err, httpStatus(err))
	}
	item.AuthorAvatar = h.resolveAuthorAvatar(ctx, item.AuthorAvatar, make(map[string]*string))
	return web.Respond(ctx, w, toAppItem(item, nil, nil), http.StatusCreated)
}

func (h *Handlers) CreatePublicItem(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	var input AppCreatePublicItem
	if status, err := decodePublicRequest(w, r, &input, publicFeedbackItemBodyLimit); err != nil {
		return web.RespondError(ctx, w, err, status)
	}
	item, err := h.feedback.CreatePublicItem(ctx, feedback.CorePublicItemInput{
		PortalSlug:  web.Params(r, "portalSlug"),
		BoardID:     input.BoardID,
		AuthorID:    userID,
		Title:       input.Title,
		Description: input.Description,
	})
	if err != nil {
		return web.RespondError(ctx, w, err, httpStatus(err))
	}
	item.AuthorAvatar = h.resolveAuthorAvatar(ctx, item.AuthorAvatar, make(map[string]*string))
	return web.Respond(ctx, w, toAppItem(item, nil, nil), http.StatusCreated)
}

func (h *Handlers) UpdateItemStatus(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	itemID, err := uuid.Parse(web.Params(r, "itemId"))
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	var input AppUpdateItemStatus
	if err := web.Decode(r, &input); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	item, err := h.feedback.UpdateItemStatus(ctx, workspace.ID, itemID, feedback.CoreUpdateItemStatusInput{
		Status:         input.Status,
		RoadmapSummary: input.RoadmapSummary,
	})
	if err != nil {
		return web.RespondError(ctx, w, err, httpStatus(err))
	}
	item.AuthorAvatar = h.resolveAuthorAvatar(ctx, item.AuthorAvatar, make(map[string]*string))
	return web.Respond(ctx, w, toAppItem(item, nil, nil), http.StatusOK)
}

func (h *Handlers) CreateComment(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
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
	var input AppCreateComment
	if err := web.Decode(r, &input); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	comment, err := h.feedback.CreateComment(ctx, feedback.CoreCommentInput{
		WorkspaceID: workspace.ID,
		ItemID:      itemID,
		AuthorID:    userID,
		Body:        input.Body,
	})
	if err != nil {
		return web.RespondError(ctx, w, err, httpStatus(err))
	}
	comment.AuthorAvatar = h.resolveAuthorAvatar(ctx, comment.AuthorAvatar, make(map[string]*string))
	return web.Respond(ctx, w, toAppComment(comment), http.StatusCreated)
}

func (h *Handlers) CreatePublicComment(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	itemID, err := uuid.Parse(web.Params(r, "itemId"))
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	var input AppCreateComment
	if status, err := decodePublicRequest(w, r, &input, publicFeedbackCommentBodyLimit); err != nil {
		return web.RespondError(ctx, w, err, status)
	}
	comment, err := h.feedback.CreatePublicComment(ctx, feedback.CorePublicCommentInput{
		PortalSlug: web.Params(r, "portalSlug"),
		ItemID:     itemID,
		AuthorID:   userID,
		Body:       input.Body,
	})
	if err != nil {
		return web.RespondError(ctx, w, err, httpStatus(err))
	}
	comment.AuthorAvatar = h.resolveAuthorAvatar(ctx, comment.AuthorAvatar, make(map[string]*string))
	return web.Respond(ctx, w, toAppComment(comment), http.StatusCreated)
}

func (h *Handlers) ToggleVote(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
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
	var input AppVoteInput
	if err := web.Decode(r, &input); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	if input.Vote == 0 {
		input.Vote = 1
	}
	result, err := h.feedback.ToggleVote(ctx, workspace.ID, itemID, userID, input.Vote)
	if err != nil {
		return web.RespondError(ctx, w, err, httpStatus(err))
	}
	return web.Respond(ctx, w, AppVoteResult{Vote: result.Vote, Voted: result.Vote == 1, VoteCount: result.VoteCount}, http.StatusOK)
}

func (h *Handlers) TogglePublicVote(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	itemID, err := uuid.Parse(web.Params(r, "itemId"))
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	var input AppVoteInput
	if status, err := decodePublicRequest(w, r, &input, publicFeedbackVoteBodyLimit); err != nil {
		return web.RespondError(ctx, w, err, status)
	}
	result, err := h.feedback.TogglePublicVote(ctx, feedback.CorePublicVoteInput{
		PortalSlug: web.Params(r, "portalSlug"),
		ItemID:     itemID,
		UserID:     userID,
		Vote:       input.Vote,
	})
	if err != nil {
		return web.RespondError(ctx, w, err, httpStatus(err))
	}
	return web.Respond(ctx, w, AppVoteResult{Vote: result.Vote, Voted: result.Vote == 1, VoteCount: result.VoteCount}, http.StatusOK)
}

func (h *Handlers) CreateStoryFromItem(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
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
	var input AppCreateStoryFromItem
	if err := web.Decode(r, &input); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	result, err := h.feedback.CreateStoryFromItem(ctx, workspace.ID, itemID, userID, feedback.CoreCreateStoryInput{
		TeamID:   input.TeamID,
		StatusID: input.StatusID,
	})
	if err != nil {
		return web.RespondError(ctx, w, err, httpStatus(err))
	}
	return web.Respond(ctx, w, AppCreateStoryResult{ItemID: result.ItemID, StoryID: result.StoryID, LinkID: result.LinkID}, http.StatusCreated)
}

func httpStatus(err error) int {
	switch {
	case errors.Is(err, feedback.ErrNotFound):
		return http.StatusNotFound
	default:
		return http.StatusBadRequest
	}
}
