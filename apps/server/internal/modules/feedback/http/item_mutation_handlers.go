package feedbackhttp

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	attachments "github.com/complexus-tech/projects-api/internal/modules/attachments/service"
	feedback "github.com/complexus-tech/projects-api/internal/modules/feedback/service"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/validate"
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
		WorkspaceID:     workspace.ID,
		PortalID:        input.PortalID,
		BoardID:         input.BoardID,
		AuthorID:        userID,
		Title:           input.Title,
		Description:     input.Description,
		DescriptionHTML: input.DescriptionHTML,
	})
	if err != nil {
		return web.RespondError(ctx, w, err, httpStatus(err))
	}
	item.AuthorAvatar = h.resolveAuthorAvatar(ctx, item.AuthorAvatar, make(map[string]*string))
	return web.Respond(ctx, w, toAppItem(item, nil, nil), http.StatusCreated)
}

func (h *Handlers) CreatePublicItem(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	if strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
		return h.createPublicItemWithAttachments(ctx, w, r, feedback.SubmissionSourcePortal)
	}
	return h.createPublicItem(ctx, w, r, feedback.SubmissionSourcePortal)
}

func (h *Handlers) CreateWidgetItem(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	if strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
		return h.createPublicItemWithAttachments(ctx, w, r, feedback.SubmissionSourceWidget)
	}
	return h.createPublicItem(ctx, w, r, feedback.SubmissionSourceWidget)
}

func (h *Handlers) createPublicItemWithAttachments(
	ctx context.Context,
	w http.ResponseWriter,
	r *http.Request,
	source string,
) error {
	if h.attachments == nil {
		return web.RespondError(ctx, w, errors.New("feedback attachment service is unavailable"), http.StatusServiceUnavailable)
	}
	const maxFiles = 5
	const multipartOverheadAllowance int64 = 1 << 20
	if err := web.ParseMultipartForm(w, r, maxFiles*validate.MaxAttachmentSize+multipartOverheadAllowance); err != nil {
		return web.RespondError(ctx, w, fmt.Errorf("invalid feedback upload: %w", err), http.StatusBadRequest)
	}
	defer func() {
		if err := web.RemoveMultipartForm(r); err != nil && h.log != nil {
			h.log.Warn(ctx, "failed to remove feedback upload temporary files", "error", err)
		}
	}()
	files := r.MultipartForm.File["files"]
	if len(files) == 0 || len(files) > maxFiles {
		return web.RespondError(ctx, w, feedback.ErrInvalidInput, http.StatusBadRequest)
	}
	boardID, err := uuid.Parse(r.FormValue("boardId"))
	if err != nil {
		return web.RespondError(ctx, w, feedback.ErrInvalidInput, http.StatusBadRequest)
	}
	input := AppCreatePublicItem{
		BoardID: boardID, Title: r.FormValue("title"), Description: r.FormValue("description"),
		DescriptionHTML: r.FormValue("descriptionHTML"), ParticipationIntent: r.FormValue("participationIntent"),
		Website: r.FormValue("website"),
	}
	if err := validatePublicItemBotTrap(input); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	userID, _ := mid.GetUserID(ctx)
	var participant *feedback.CoreParticipant
	if input.ParticipationIntent == feedback.ParticipationIntentVerifiedGuest || input.ParticipationIntent == feedback.ParticipationIntentExternal {
		resolved, resolveErr := h.resolvePublicParticipant(ctx, r, input.ParticipationIntent)
		if resolveErr != nil {
			return web.RespondError(ctx, w, resolveErr, httpStatus(resolveErr))
		}
		participant = &resolved.Participant
	}
	portalSlug := web.Params(r, "portalSlug")
	portal, err := h.feedback.GetPublicPortal(ctx, portalSlug)
	if err != nil {
		return web.RespondError(ctx, w, err, httpStatus(err))
	}
	uploaded := make([]attachments.FileInfo, 0, len(files))
	cleanupUploads := func() {
		for _, file := range uploaded {
			if cleanupErr := h.attachments.DeleteOrphanedMedia(ctx, file.ID, portal.WorkspaceID); cleanupErr != nil && h.log != nil {
				h.log.Warn(ctx, "failed to clean up feedback attachment", "error", cleanupErr, "attachment_id", file.ID)
			}
		}
	}
	for _, header := range files {
		file, openErr := header.Open()
		if openErr != nil {
			cleanupUploads()
			return web.RespondError(ctx, w, openErr, http.StatusBadRequest)
		}
		fileInfo, uploadErr := h.attachments.UploadAttachment(ctx, file, header, userID, portal.WorkspaceID)
		_ = file.Close()
		if uploadErr != nil {
			cleanupUploads()
			return web.RespondError(ctx, w, uploadErr, http.StatusBadRequest)
		}
		uploaded = append(uploaded, fileInfo)
	}
	result, err := h.feedback.CreatePublicItem(ctx, feedback.CorePublicItemInput{
		PortalSlug: portalSlug, BoardID: input.BoardID, AuthorID: userID, Title: input.Title,
		Description: input.Description, DescriptionHTML: input.DescriptionHTML, Source: source,
		ParticipationIntent: input.ParticipationIntent, Participant: participant,
	})
	if err != nil {
		cleanupUploads()
		return web.RespondError(ctx, w, err, httpStatus(err))
	}
	appAttachments := make([]AppItemAttachment, 0, len(uploaded))
	for _, file := range uploaded {
		linked, linkErr := h.feedback.AttachPublicItemFile(
			ctx, portalSlug, result.Item.ID, file.ID, userID, participant, input.ParticipationIntent,
		)
		if linkErr != nil {
			return web.RespondError(ctx, w, linkErr, httpStatus(linkErr))
		}
		appAttachments = append(appAttachments, toAppItemAttachment(portalSlug, linked))
	}
	result.Item.AuthorAvatar = h.resolveAuthorAvatar(ctx, result.Item.AuthorAvatar, make(map[string]*string))
	item := toAppItem(result.Item, nil, nil, appAttachments...)
	item.Anonymous = result.Anonymous
	item.ParticipantKind = result.ParticipantKind
	item.Following = result.Following
	return web.Respond(ctx, w, item, http.StatusCreated)
}

func (h *Handlers) ResolvePublicItemAttachment(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	if h.attachments == nil {
		return web.RespondError(ctx, w, errors.New("feedback attachment service is unavailable"), http.StatusServiceUnavailable)
	}
	itemID, itemErr := uuid.Parse(web.Params(r, "itemId"))
	attachmentID, attachmentErr := uuid.Parse(web.Params(r, "attachmentId"))
	if itemErr != nil || attachmentErr != nil {
		return web.RespondError(ctx, w, feedback.ErrInvalidInput, http.StatusBadRequest)
	}
	attachment, err := h.feedback.GetPublicItemAttachment(ctx, web.Params(r, "portalSlug"), itemID, attachmentID)
	if err != nil {
		return web.RespondError(ctx, w, err, httpStatus(err))
	}
	file, err := h.attachments.ResolveAttachmentAccessURL(ctx, attachment.ID, attachment.WorkspaceID, 5*time.Minute)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusNotFound)
	}
	w.Header().Set("Cache-Control", "private, no-store")
	w.Header().Set("Referrer-Policy", "no-referrer")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	http.Redirect(w, r, file.URL, http.StatusTemporaryRedirect)
	return nil
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
		DescriptionHTML:     input.DescriptionHTML,
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
