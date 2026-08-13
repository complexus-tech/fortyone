package feedbackhttp

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	feedback "github.com/complexus-tech/projects-api/internal/modules/feedback/service"
	teams "github.com/complexus-tech/projects-api/internal/modules/teams/service"
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
	defaultTeamFeedbackPageSize    = 25
	maxTeamFeedbackPageSize        = 50
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

func validatePublicItemBotTrap(input AppCreatePublicItem) error {
	if strings.TrimSpace(input.Website) != "" {
		return errors.New("invalid feedback submission")
	}
	return nil
}

type profileImageResolver interface {
	ResolveProfileImageURL(ctx context.Context, avatar string, expiry time.Duration) (string, error)
}

type teamAccessService interface {
	GetByID(ctx context.Context, teamID, workspaceID, userID uuid.UUID) (teams.CoreTeam, error)
}

type Handlers struct {
	feedback      *feedback.Service
	teams         teamAccessService
	profileImages profileImageResolver
	log           *logger.Logger
}

func New(service *feedback.Service, teamAccess teamAccessService, profileImages profileImageResolver, log *logger.Logger) *Handlers {
	return &Handlers{feedback: service, teams: teamAccess, profileImages: profileImages, log: log}
}

func (h *Handlers) authorizeTeam(ctx context.Context, workspaceID, teamID, userID uuid.UUID) error {
	if h.teams == nil {
		return errors.New("team access service is required")
	}
	if _, err := h.teams.GetByID(ctx, teamID, workspaceID, userID); err != nil {
		if h.log != nil {
			h.log.Warn(ctx, "feedback team access denied", "team_id", teamID, "user_id", userID, "error", err)
		}
		if errors.Is(err, teams.ErrTeamNotFound) {
			return feedback.ErrNotFound
		}
		return err
	}
	return nil
}

func (h *Handlers) authorizeItemTeam(ctx context.Context, workspaceID, itemID, userID uuid.UUID) error {
	item, err := h.feedback.GetItem(ctx, workspaceID, itemID)
	if err != nil {
		return err
	}
	return h.authorizeTeam(ctx, workspaceID, item.Board.TeamID, userID)
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

func (h *Handlers) resolveSimilarItemAvatars(ctx context.Context, items []feedback.CoreSimilarItem) {
	resolvedByAvatar := make(map[string]*string)
	for i := range items {
		items[i].AuthorAvatar = h.resolveAuthorAvatar(ctx, items[i].AuthorAvatar, resolvedByAvatar)
	}
}

func (h *Handlers) GetPortal(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	slug := web.Params(r, "portalSlug")
	input, err := publicItemsInput(r)
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
	input, err := publicItemsInput(r)
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
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	items, err := h.feedback.ListPublicSimilarItems(
		ctx,
		web.Params(r, "portalSlug"),
		r.URL.Query().Get("title"),
		r.URL.Query().Get("description"),
		limit,
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
	page, pageSize := publicContributorPagination(r)
	activity, err := h.feedback.ListContributorActivity(ctx, userID, r.URL.Query().Get("type"), page, pageSize)
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
	contributor, err := h.feedback.GetWorkspacePublicContributor(
		ctx,
		web.Params(r, "workspaceSlug"),
		web.Params(r, "portalSlug"),
		authorID,
	)
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
	page, pageSize := publicContributorPagination(r)
	comments, err := h.feedback.ListPublicContributorComments(
		ctx,
		web.Params(r, "portalSlug"),
		authorID,
		page,
		pageSize,
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
	page, pageSize := publicContributorPagination(r)
	comments, err := h.feedback.ListWorkspacePublicContributorComments(
		ctx,
		web.Params(r, "workspaceSlug"),
		web.Params(r, "portalSlug"),
		authorID,
		page,
		pageSize,
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
	return web.Respond(ctx, w, AppContributorCommentsResponse{
		Comments: comments,
		Pagination: AppItemsPagination{
			Page:     page.Page,
			PageSize: page.PageSize,
			HasMore:  page.HasMore,
			NextPage: nextPage,
		},
	}, http.StatusOK)
}

func publicContributorID(r *http.Request) (uuid.UUID, error) {
	authorID, err := uuid.Parse(web.Params(r, "authorId"))
	if err != nil || authorID == uuid.Nil {
		return uuid.Nil, fmt.Errorf("%w: contributor id must be a non-nil UUID", feedback.ErrInvalidInput)
	}
	return authorID, nil
}

func publicContributorPagination(r *http.Request) (int, int) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("pageSize"))
	return page, pageSize
}

func publicItemsInput(r *http.Request) (feedback.CorePortalSnapshotInput, error) {
	query := r.URL.Query()
	view := query.Get("view")
	if view != "" && view != "summary" {
		return feedback.CorePortalSnapshotInput{}, fmt.Errorf("%w: view must be summary when provided", feedback.ErrInvalidInput)
	}
	page, _ := strconv.Atoi(query.Get("page"))
	pageSize, _ := strconv.Atoi(query.Get("pageSize"))
	input := feedback.CorePortalSnapshotInput{
		Status:      query.Get("status"),
		Search:      query.Get("search"),
		Sort:        query.Get("sort"),
		Page:        page,
		PageSize:    pageSize,
		SummaryOnly: view == "summary",
	}
	if boardID := query.Get("boardId"); boardID != "" {
		parsed, err := uuid.Parse(boardID)
		if err != nil {
			return feedback.CorePortalSnapshotInput{}, fmt.Errorf("%w: boardId must be a non-nil UUID", feedback.ErrInvalidInput)
		}
		if parsed == uuid.Nil {
			return feedback.CorePortalSnapshotInput{}, fmt.Errorf("%w: boardId must be a non-nil UUID", feedback.ErrInvalidInput)
		}
		input.BoardID = &parsed
	}
	if itemID := query.Get("itemId"); itemID != "" {
		parsed, err := uuid.Parse(itemID)
		if err != nil || parsed == uuid.Nil {
			return feedback.CorePortalSnapshotInput{}, fmt.Errorf("%w: itemId must be a non-nil UUID", feedback.ErrInvalidInput)
		}
		input.ItemID = parsed
	}
	if authorID := query.Get("authorId"); authorID != "" {
		parsed, err := uuid.Parse(authorID)
		if err != nil || parsed == uuid.Nil {
			return feedback.CorePortalSnapshotInput{}, fmt.Errorf("%w: authorId must be a non-nil UUID", feedback.ErrInvalidInput)
		}
		input.AuthorID = parsed
	}
	return input, nil
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
	page, pageSize := teamFeedbackPagination(r)
	status := strings.TrimSpace(r.URL.Query().Get("status"))
	if status == "" {
		status = "active"
	}
	if status == feedback.ListStatusTrashed && workspace.UserRole != string(mid.RoleAdmin) {
		return web.RespondError(ctx, w, errors.New("you don't have permission to access trashed feedback"), http.StatusForbidden)
	}
	search := strings.TrimSpace(r.URL.Query().Get("search"))
	itemsPage, err := h.feedback.ListTeamItems(ctx, workspace.ID, teamID, userID, status, search, page, pageSize)
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
			Page:     page,
			PageSize: pageSize,
			HasMore:  itemsPage.HasMore,
			NextPage: page + 1,
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
	return h.createPublicItem(ctx, w, r, feedback.SubmissionSourcePortal)
}

func (h *Handlers) CreateWidgetItem(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	return h.createPublicItem(ctx, w, r, feedback.SubmissionSourceWidget)
}

func (h *Handlers) createPublicItem(ctx context.Context, w http.ResponseWriter, r *http.Request, source string) error {
	userID, _ := mid.GetUserID(ctx)
	var input AppCreatePublicItem
	if status, err := decodePublicRequest(w, r, &input, publicFeedbackItemBodyLimit); err != nil {
		return web.RespondError(ctx, w, err, status)
	}
	if err := validatePublicItemBotTrap(input); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	var participant *feedback.CoreParticipant
	if input.ParticipationIntent == feedback.ParticipationIntentVerifiedGuest || input.ParticipationIntent == feedback.ParticipationIntentExternal {
		resolved, err := h.resolvePublicParticipant(ctx, r, input.ParticipationIntent)
		if err != nil {
			return web.RespondError(ctx, w, err, httpStatus(err))
		}
		participant = &resolved.Participant
	}
	result, err := h.feedback.CreatePublicItem(ctx, feedback.CorePublicItemInput{
		PortalSlug:          web.Params(r, "portalSlug"),
		BoardID:             input.BoardID,
		AuthorID:            userID,
		Title:               input.Title,
		Description:         input.Description,
		Source:              source,
		ParticipationIntent: input.ParticipationIntent,
		Participant:         participant,
	})
	if err != nil {
		return web.RespondError(ctx, w, err, httpStatus(err))
	}
	item := result.Item
	item.AuthorAvatar = h.resolveAuthorAvatar(ctx, item.AuthorAvatar, make(map[string]*string))
	response := toAppItem(item, nil, nil)
	response.ParticipantKind = result.ParticipantKind
	response.Following = result.Following
	if result.Anonymous {
		w.Header().Set("Cache-Control", "no-store")
		response.Anonymous = true
	}
	return web.Respond(ctx, w, response, http.StatusCreated)
}

func (h *Handlers) UpdateItemStatus(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
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
	var input AppUpdateItemStatus
	if err := web.Decode(r, &input); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	item, err := h.feedback.UpdateItemStatus(ctx, workspace.ID, itemID, feedback.CoreUpdateItemStatusInput{
		Status:         input.Status,
		RoadmapSummary: input.RoadmapSummary,
		ActorID:        userID,
	})
	if err != nil {
		return web.RespondError(ctx, w, err, httpStatus(err))
	}
	item.AuthorAvatar = h.resolveAuthorAvatar(ctx, item.AuthorAvatar, make(map[string]*string))
	return web.Respond(ctx, w, toAppItem(item, nil, nil), http.StatusOK)
}

func (h *Handlers) MergeItem(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	actorID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	sourceItemID, err := uuid.Parse(web.Params(r, "sourceItemId"))
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	var input AppMergeItemInput
	if err := web.Decode(r, &input); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	result, err := h.feedback.MergeItems(ctx, feedback.CoreMergeItemInput{
		WorkspaceID: workspace.ID, SourceItemID: sourceItemID, TargetItemID: input.TargetItemID, ActorID: actorID,
	})
	if err != nil {
		return web.RespondError(ctx, w, err, httpStatus(err))
	}
	return web.Respond(ctx, w, AppMergeItemResult{
		SourceItemID: result.SourceItemID, TargetItemID: result.TargetItemID, PortalID: result.PortalID,
		MergedAt: result.MergedAt, MergedByUserID: result.MergedByUserID,
		MovedFollowerCount: result.MovedFollowerCount, MovedUpdateLinkCount: result.MovedUpdateLinkCount,
		MovedStoryLinkCount: result.MovedStoryLinkCount, Target: toAppItem(result.Target, nil, nil),
	}, http.StatusOK)
}

func (h *Handlers) ListMergeCandidates(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	sourceItemID, err := uuid.Parse(web.Params(r, "sourceItemId"))
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	limit, err := parseCandidateLimit(r)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	page, err := h.feedback.ListMergeCandidates(ctx, workspace.ID, sourceItemID, r.URL.Query().Get("search"), limit)
	if err != nil {
		return web.RespondError(ctx, w, err, httpStatus(err))
	}
	return web.Respond(ctx, w, toAppMergeCandidatesPage(page), http.StatusOK)
}

func (h *Handlers) ListPortalItemCandidates(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	portalID, err := uuid.Parse(web.Params(r, "portalId"))
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	limit, err := parseCandidateLimit(r)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	page, err := h.feedback.ListPortalItemCandidates(ctx, workspace.ID, portalID, r.URL.Query().Get("search"), limit)
	if err != nil {
		return web.RespondError(ctx, w, err, httpStatus(err))
	}
	return web.Respond(ctx, w, toAppMergeCandidatesPage(page), http.StatusOK)
}

func parseCandidateLimit(r *http.Request) (int, error) {
	limit := 30
	if raw := strings.TrimSpace(r.URL.Query().Get("limit")); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed <= 0 {
			return 0, feedback.ErrInvalidInput
		}
		limit = parsed
	}
	return limit, nil
}

func toAppMergeCandidatesPage(page feedback.CoreMergeCandidatesPage) AppMergeCandidatesPage {
	candidates := make([]AppMergeCandidate, 0, len(page.Candidates))
	for _, candidate := range page.Candidates {
		candidates = append(candidates, AppMergeCandidate{
			ID: candidate.ID, Slug: candidate.Slug, Title: candidate.Title, Status: candidate.Status,
			VoteCount: candidate.VoteCount, CommentCount: candidate.CommentCount,
		})
	}
	return AppMergeCandidatesPage{Candidates: candidates, HasMore: page.HasMore}
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
	if err := h.authorizeItemTeam(ctx, workspace.ID, itemID, userID); err != nil {
		return web.RespondError(ctx, w, err, httpStatus(err))
	}
	var input AppCreateComment
	if err := web.Decode(r, &input); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	comment, err := h.feedback.CreateComment(ctx, feedback.CoreCommentInput{
		WorkspaceID: workspace.ID,
		ItemID:      itemID,
		AuthorID:    userID,
		ParentID:    input.ParentID,
		Body:        input.Body,
	})
	if err != nil {
		return web.RespondError(ctx, w, err, httpStatus(err))
	}
	comment.AuthorAvatar = h.resolveAuthorAvatar(ctx, comment.AuthorAvatar, make(map[string]*string))
	return web.Respond(ctx, w, toAppComment(comment), http.StatusCreated)
}

func (h *Handlers) CreatePublicComment(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	itemID, err := uuid.Parse(web.Params(r, "itemId"))
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	var input AppCreateComment
	if status, err := decodePublicRequest(w, r, &input, publicFeedbackCommentBodyLimit); err != nil {
		return web.RespondError(ctx, w, err, status)
	}
	resolved, err := h.resolvePublicParticipant(ctx, r, strings.ToLower(strings.TrimSpace(input.ParticipationIntent)))
	if err != nil {
		return web.RespondError(ctx, w, err, httpStatus(err))
	}
	var participant *feedback.CoreParticipant
	if resolved.Participant.Kind != feedback.ContributorKindAccount {
		participant = &resolved.Participant
	}
	comment, err := h.feedback.CreatePublicComment(ctx, feedback.CorePublicCommentInput{
		PortalSlug:  web.Params(r, "portalSlug"),
		ItemID:      itemID,
		AuthorID:    resolved.AccountID,
		Participant: participant,
		ParentID:    input.ParentID,
		Body:        input.Body,
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
	if err := h.authorizeItemTeam(ctx, workspace.ID, itemID, userID); err != nil {
		return web.RespondError(ctx, w, err, httpStatus(err))
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
	itemID, err := uuid.Parse(web.Params(r, "itemId"))
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	var input AppVoteInput
	if status, err := decodePublicRequest(w, r, &input, publicFeedbackVoteBodyLimit); err != nil {
		return web.RespondError(ctx, w, err, status)
	}
	resolved, err := h.resolvePublicParticipant(ctx, r, strings.ToLower(strings.TrimSpace(input.ParticipationIntent)))
	if err != nil {
		return web.RespondError(ctx, w, err, httpStatus(err))
	}
	var participant *feedback.CoreParticipant
	if resolved.Participant.Kind != feedback.ContributorKindAccount {
		participant = &resolved.Participant
	}
	result, err := h.feedback.TogglePublicVote(ctx, feedback.CorePublicVoteInput{
		PortalSlug:  web.Params(r, "portalSlug"),
		ItemID:      itemID,
		UserID:      resolved.AccountID,
		Participant: participant,
		Vote:        input.Vote,
	})
	if err != nil {
		return web.RespondError(ctx, w, err, httpStatus(err))
	}
	return web.Respond(ctx, w, AppVoteResult{Vote: result.Vote, Voted: result.Vote == 1, VoteCount: result.VoteCount, ParticipantKind: resolved.Participant.Kind}, http.StatusOK)
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
	if err := h.authorizeItemTeam(ctx, workspace.ID, itemID, userID); err != nil {
		return web.RespondError(ctx, w, err, httpStatus(err))
	}
	var input AppCreateStoryFromItem
	if err := web.Decode(r, &input); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	result, err := h.feedback.CreateStoryFromItem(ctx, workspace.ID, itemID, userID, feedback.CoreCreateStoryInput{
		TeamID:   input.TeamID,
		StoryID:  input.StoryID,
		StatusID: input.StatusID,
	})
	if err != nil {
		return web.RespondError(ctx, w, err, httpStatus(err))
	}
	status := http.StatusCreated
	if !result.Created {
		status = http.StatusOK
	}
	return web.Respond(ctx, w, AppCreateStoryResult{ItemID: result.ItemID, StoryID: result.StoryID, LinkID: result.LinkID, Created: result.Created}, status)
}

func teamFeedbackPagination(r *http.Request) (int, int) {
	page := 1
	pageSize := defaultTeamFeedbackPageSize
	if parsed, err := strconv.Atoi(r.URL.Query().Get("page")); err == nil && parsed > 0 {
		page = parsed
	}
	if parsed, err := strconv.Atoi(r.URL.Query().Get("pageSize")); err == nil && parsed > 0 {
		pageSize = parsed
	}
	if pageSize > maxTeamFeedbackPageSize {
		pageSize = maxTeamFeedbackPageSize
	}
	return page, pageSize
}

func httpStatus(err error) int {
	switch {
	case errors.Is(err, feedback.ErrNotFound):
		return http.StatusNotFound
	case errors.Is(err, feedback.ErrAlreadyPlanned):
		return http.StatusConflict
	case errors.Is(err, feedback.ErrBoardExists):
		return http.StatusConflict
	case errors.Is(err, feedback.ErrStoryManaged):
		return http.StatusConflict
	case errors.Is(err, feedback.ErrDuplicateItem):
		return http.StatusConflict
	case errors.Is(err, feedback.ErrParticipationNotAllowed):
		return http.StatusForbidden
	case errors.Is(err, feedback.ErrContributorBlocked), errors.Is(err, feedback.ErrWidgetOriginNotAllowed):
		return http.StatusForbidden
	case errors.Is(err, feedback.ErrAuthenticationRequired), errors.Is(err, feedback.ErrContributorSessionInvalid), errors.Is(err, feedback.ErrWidgetAssertionInvalid):
		return http.StatusUnauthorized
	case errors.Is(err, feedback.ErrVerificationExpired), errors.Is(err, feedback.ErrVerificationConsumed):
		return http.StatusGone
	case errors.Is(err, feedback.ErrVerificationAttempts):
		return http.StatusTooManyRequests
	case errors.Is(err, feedback.ErrWidgetAssertionReplayed):
		return http.StatusConflict
	case errors.Is(err, feedback.ErrMergeConflict):
		return http.StatusConflict
	case errors.Is(err, feedback.ErrFeatureUnavailable):
		return http.StatusServiceUnavailable
	case errors.Is(err, feedback.ErrTeamMismatch):
		return http.StatusBadRequest
	case errors.Is(err, feedback.ErrInvalidInput):
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}
