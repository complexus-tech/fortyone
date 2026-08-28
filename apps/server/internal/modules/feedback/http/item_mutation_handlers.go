package feedbackhttp

import (
	"context"
	"net/http"
	"strings"

	feedback "github.com/complexus-tech/projects-api/internal/modules/feedback/service"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

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
	query, err := parseCandidateItemsQuery(r.URL.Query())
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	page, err := h.feedback.ListMergeCandidates(ctx, workspace.ID, sourceItemID, query.Search, query.Limit)
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
	query, err := parseCandidateItemsQuery(r.URL.Query())
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	page, err := h.feedback.ListPortalItemCandidates(ctx, workspace.ID, portalID, query.Search, query.Limit)
	if err != nil {
		return web.RespondError(ctx, w, err, httpStatus(err))
	}
	return web.Respond(ctx, w, toAppMergeCandidatesPage(page), http.StatusOK)
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
