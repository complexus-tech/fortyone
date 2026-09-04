package feedback

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/complexus-tech/projects-api/pkg/events"
	"github.com/google/uuid"
)

func (r *repoStub) ListBoardReviewers(ctx context.Context, workspaceID, boardID uuid.UUID) ([]CoreBoardReviewer, error) {
	return append([]CoreBoardReviewer(nil), r.reviewers...), nil
}

func (r *repoStub) SetBoardReviewer(ctx context.Context, input CoreBoardReviewerInput) (CoreBoardReviewer, error) {
	r.reviewerInputs = append(r.reviewerInputs, input)
	reviewer := CoreBoardReviewer{
		UserID:         input.UserID,
		Name:           "Ada Lovelace",
		Email:          "ada@example.com",
		Role:           "member",
		EmailFrequency: input.EmailFrequency,
	}
	return reviewer, nil
}

func (r *repoStub) ListItems(ctx context.Context, input CoreListItemsInput) (CoreItemsPage, error) {
	r.listItemInputs = append(r.listItemInputs, input)
	result := make([]CoreItem, 0, len(r.items))
	for _, item := range r.items {
		if input.PortalID != uuid.Nil && item.PortalID != input.PortalID {
			continue
		}
		if input.TeamID != nil && item.Board.TeamID != *input.TeamID {
			continue
		}
		if input.ItemID != uuid.Nil && item.ID != input.ItemID {
			continue
		}
		if input.DeletedOnly != (item.DeletedAt != nil) {
			continue
		}
		if input.ViewerID != uuid.Nil && r.reads != nil {
			if readAt, ok := r.reads[feedbackReadKey(item.ID, input.ViewerID)]; ok {
				item.ReadAt = &readAt
			}
		}
		result = append(result, item)
	}
	return CoreItemsPage{Items: result, HasMore: false}, nil
}

func (r *repoStub) GetContributor(ctx context.Context, portalID, authorID uuid.UUID) (CoreContributor, error) {
	for _, contributor := range r.contributors {
		if contributor.ID != authorID {
			continue
		}
		if scopedPortalID, ok := r.contributorPortals[authorID]; ok && scopedPortalID != portalID {
			continue
		}
		return contributor, nil
	}
	return CoreContributor{}, sql.ErrNoRows
}

func (r *repoStub) ContributorExists(ctx context.Context, portalID, authorID uuid.UUID) (bool, error) {
	_, err := r.GetContributor(ctx, portalID, authorID)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	return err == nil, err
}

func (r *repoStub) ListContributorComments(ctx context.Context, input CoreListContributorCommentsInput) (CoreContributorCommentsPage, error) {
	r.contributorCommentInputs = append(r.contributorCommentInputs, input)
	return CoreContributorCommentsPage{
		Comments: append([]CoreContributorComment(nil), r.contributorComments...),
		Page:     input.Page,
		PageSize: input.PageSize,
		HasMore:  r.contributorCommentsHasMore,
	}, nil
}

func (r *repoStub) ListComments(ctx context.Context, portalID uuid.UUID, itemIDs []uuid.UUID) ([]CoreComment, error) {
	r.listCommentItemIDs = append(r.listCommentItemIDs, append([]uuid.UUID(nil), itemIDs...))
	visible := make(map[uuid.UUID]struct{}, len(itemIDs))
	for _, itemID := range itemIDs {
		visible[itemID] = struct{}{}
	}
	result := make([]CoreComment, 0)
	for _, comment := range r.comments {
		if _, ok := visible[comment.ItemID]; ok {
			result = append(result, comment)
		}
	}
	return result, nil
}

func (r *repoStub) ListItemComments(ctx context.Context, workspaceID, itemID uuid.UUID) ([]CoreComment, error) {
	result := make([]CoreComment, 0, len(r.comments))
	for _, comment := range r.comments {
		if comment.WorkspaceID == workspaceID && comment.ItemID == itemID {
			result = append(result, comment)
		}
	}
	return result, nil
}

func (r *repoStub) LinkItemAttachment(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) (CoreItemAttachment, error) {
	return CoreItemAttachment{}, nil
}

func (r *repoStub) GetItemAttachment(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) (CoreItemAttachment, error) {
	return CoreItemAttachment{}, nil
}

func (r *repoStub) ListItemAttachments(context.Context, uuid.UUID, []uuid.UUID) ([]CoreItemAttachment, error) {
	return []CoreItemAttachment{}, nil
}

func (r *repoStub) GetComment(ctx context.Context, workspaceID, itemID, commentID uuid.UUID) (CoreComment, error) {
	for _, comment := range r.comments {
		if comment.WorkspaceID == workspaceID && comment.ItemID == itemID && comment.ID == commentID {
			return comment, nil
		}
	}
	return CoreComment{}, sql.ErrNoRows
}

func (r *repoStub) ListStoryLinks(ctx context.Context, portalID uuid.UUID, itemIDs []uuid.UUID) ([]CoreStoryLink, error) {
	r.listStoryLinkItemIDs = append(r.listStoryLinkItemIDs, append([]uuid.UUID(nil), itemIDs...))
	visible := make(map[uuid.UUID]struct{}, len(itemIDs))
	for _, itemID := range itemIDs {
		visible[itemID] = struct{}{}
	}
	result := make([]CoreStoryLink, 0)
	for _, link := range r.storyLinks {
		if _, ok := visible[link.ItemID]; ok {
			result = append(result, link)
		}
	}
	return result, nil
}

func (r *repoStub) ListItemStoryLinks(ctx context.Context, workspaceID, itemID uuid.UUID) ([]CoreStoryLink, error) {
	result := make([]CoreStoryLink, 0, len(r.storyLinks))
	for _, link := range r.storyLinks {
		if link.WorkspaceID == workspaceID && link.ItemID == itemID {
			result = append(result, link)
		}
	}
	return result, nil
}

func (r *repoStub) ListStoryFeedbackLinks(ctx context.Context, workspaceID, storyID uuid.UUID) ([]CoreStoryFeedbackLink, error) {
	result := make([]CoreStoryFeedbackLink, 0)
	for _, link := range r.storyLinks {
		if link.WorkspaceID != workspaceID || link.StoryID != storyID || !link.IsPrimary {
			continue
		}
		for _, item := range r.items {
			if item.ID == link.ItemID {
				result = append(result, CoreStoryFeedbackLink{
					ID:            link.ID,
					WorkspaceID:   workspaceID,
					ItemID:        item.ID,
					StoryID:       storyID,
					TeamID:        item.Board.TeamID,
					FeedbackTitle: item.Title,
					Relationship:  link.Relationship,
					IsPrimary:     true,
					CreatedAt:     link.CreatedAt,
				})
			}
		}
	}
	return result, nil
}

func (r *repoStub) GetItem(ctx context.Context, workspaceID, itemID uuid.UUID) (CoreItem, error) {
	for _, item := range r.items {
		if item.WorkspaceID == workspaceID && item.ID == itemID && item.DeletedAt == nil {
			return item, nil
		}
	}
	return CoreItem{}, sql.ErrNoRows
}

func (r *repoStub) GetItemReadAt(ctx context.Context, workspaceID, itemID, userID uuid.UUID) (*time.Time, error) {
	if r.reads == nil {
		return nil, nil
	}
	readAt, ok := r.reads[feedbackReadKey(itemID, userID)]
	if !ok {
		return nil, nil
	}
	return &readAt, nil
}

func (r *repoStub) ListTeamSummaries(ctx context.Context, workspaceID, userID uuid.UUID) ([]CoreTeamSummary, error) {
	byTeam := make(map[uuid.UUID]*CoreTeamSummary)
	for _, board := range r.boards {
		if board.WorkspaceID != workspaceID {
			continue
		}
		byTeam[board.TeamID] = &CoreTeamSummary{TeamID: board.TeamID, Enabled: true}
	}
	for _, item := range r.items {
		summary := byTeam[item.Board.TeamID]
		if summary == nil || item.WorkspaceID != workspaceID {
			continue
		}
		summary.TotalCount++
		if _, read := r.reads[feedbackReadKey(item.ID, userID)]; !read {
			summary.UnreadCount++
		}
	}
	result := make([]CoreTeamSummary, 0, len(byTeam))
	for _, summary := range byTeam {
		result = append(result, *summary)
	}
	return result, nil
}

func (r *repoStub) MarkItemRead(ctx context.Context, workspaceID, itemID, userID uuid.UUID) (time.Time, error) {
	if r.reads == nil {
		r.reads = make(map[string]time.Time)
	}
	key := feedbackReadKey(itemID, userID)
	if readAt, ok := r.reads[key]; ok {
		return readAt, nil
	}
	readAt := time.Now()
	r.reads[key] = readAt
	return readAt, nil
}

func (r *repoStub) MarkItemUnread(ctx context.Context, workspaceID, itemID, userID uuid.UUID) error {
	delete(r.reads, feedbackReadKey(itemID, userID))
	return nil
}

func (r *repoStub) GetItemByPortal(ctx context.Context, portalID, itemID uuid.UUID) (CoreItem, error) {
	for _, item := range r.items {
		if item.PortalID == portalID && item.ID == itemID {
			return item, nil
		}
	}
	return CoreItem{}, sql.ErrNoRows
}

func (r *repoStub) ListSimilarItems(ctx context.Context, portalID uuid.UUID, title, description string, limit int) ([]CoreSimilarItem, error) {
	items := r.similarItems
	if len(items) > limit {
		items = items[:limit]
	}
	return append([]CoreSimilarItem(nil), items...), nil
}

func (r *repoStub) CreateItem(ctx context.Context, input CoreItemInput) (CoreItem, error) {
	r.createdItems = append(r.createdItems, input)
	item := CoreItem{
		ID:          uuid.New(),
		WorkspaceID: input.WorkspaceID,
		PortalID:    input.PortalID,
		BoardID:     input.BoardID,
		AuthorID:    input.AuthorID,
		Title:       input.Title,
		Description: input.Description,
		Slug:        input.Slug,
		Status:      StatusPending,
	}
	r.items = append(r.items, item)
	return item, nil
}

func (r *repoStub) GetOrCreateAccountContributor(ctx context.Context, portalID, userID uuid.UUID) (CoreContributor, error) {
	return CoreContributor{ID: uuid.New(), PortalID: portalID, UserID: userID, Kind: ContributorKindAccount}, nil
}

func (r *repoStub) CreateAnonymousItem(ctx context.Context, input CoreItemInput) (CoreItem, error) {
	r.createdAnonymousItems = append(r.createdAnonymousItems, input)
	input.ContributorID = uuid.New()
	return r.CreateItem(ctx, input)
}

func (r *repoStub) UpdateItemStatus(ctx context.Context, workspaceID, itemID uuid.UUID, input CoreUpdateItemStatusInput) (CoreItem, bool, error) {
	item, err := r.GetItem(ctx, workspaceID, itemID)
	if err != nil {
		return CoreItem{}, false, err
	}
	if !input.AllowLinked {
		for _, link := range item.StoryLinks {
			if link.IsPrimary {
				return CoreItem{}, false, ErrStoryManaged
			}
		}
	}
	statusChanged := item.Status != input.Status
	item.Status = input.Status
	if input.RoadmapSummary != nil {
		item.RoadmapSummary = input.RoadmapSummary
	}
	for index := range r.items {
		if r.items[index].ID == itemID {
			r.items[index] = item
			break
		}
	}
	return item, statusChanged, nil
}

func (r *repoStub) UpdateItemStatusIfUnchanged(
	ctx context.Context,
	workspaceID, itemID uuid.UUID,
	expectedUpdatedAt time.Time,
	input CoreUpdateItemStatusInput,
) (CoreItem, bool, bool, error) {
	item, err := r.GetItem(ctx, workspaceID, itemID)
	if err != nil {
		return CoreItem{}, false, false, err
	}
	if !item.UpdatedAt.Equal(expectedUpdatedAt) {
		return CoreItem{}, false, false, nil
	}
	updated, changed, err := r.UpdateItemStatus(ctx, workspaceID, itemID, input)
	return updated, changed, err == nil, err
}

func (r *repoStub) TrashItem(ctx context.Context, workspaceID, itemID uuid.UUID) error {
	for index, item := range r.items {
		if item.WorkspaceID != workspaceID || item.ID != itemID || item.DeletedAt != nil {
			continue
		}
		if item.MergedIntoItemID != nil {
			return ErrMergeConflict
		}
		for _, possibleSource := range r.items {
			if possibleSource.WorkspaceID == workspaceID && possibleSource.PortalID == item.PortalID &&
				possibleSource.MergedIntoItemID != nil && *possibleSource.MergedIntoItemID == item.ID {
				return ErrMergeConflict
			}
		}
		if len(item.StoryLinks) > 0 && item.StoryLinks[0].IsPrimary {
			return ErrStoryManaged
		}
		deletedAt := time.Now()
		r.items[index].DeletedAt = &deletedAt
		return nil
	}
	return ErrNotFound
}

func (r *repoStub) RestoreItem(ctx context.Context, workspaceID, itemID uuid.UUID) error {
	for index, item := range r.items {
		if item.WorkspaceID == workspaceID && item.ID == itemID && item.DeletedAt != nil {
			r.items[index].DeletedAt = nil
			return nil
		}
	}
	return ErrNotFound
}

func (r *repoStub) CreateComment(ctx context.Context, input CoreCommentInput) (CoreComment, error) {
	comment := CoreComment{ID: uuid.New(), WorkspaceID: input.WorkspaceID, ItemID: input.ItemID, AuthorID: input.AuthorID, ParentID: input.ParentID, Body: input.Body}
	r.comments = append(r.comments, comment)
	return comment, nil
}

func (r *repoStub) ToggleVote(ctx context.Context, workspaceID, itemID, userID uuid.UUID, vote int) (CoreVoteResult, error) {
	return CoreVoteResult{Vote: vote, VoteCount: vote}, nil
}

func (r *repoStub) LinkStory(ctx context.Context, input CoreStoryLinkInput) (CoreStoryLink, error) {
	if r.linkStoryErr != nil {
		err := r.linkStoryErr
		r.linkStoryErr = nil
		if r.linkStoryWinner != nil {
			r.storyLinks = append(r.storyLinks, *r.linkStoryWinner)
		}
		return CoreStoryLink{}, err
	}
	link := CoreStoryLink{ID: uuid.New(), WorkspaceID: input.WorkspaceID, ItemID: input.ItemID, StoryID: input.StoryID, Relationship: input.Relationship, IsPrimary: input.IsPrimary, CreatedByUserID: input.CreatedByUserID}
	r.storyLinks = append(r.storyLinks, link)
	return link, nil
}

func (r *repoStub) FindFirstStatusByCategory(ctx context.Context, teamID uuid.UUID, category string) (*uuid.UUID, error) {
	return &r.statusID, nil
}

func (r *repoStub) GetStatusCategory(ctx context.Context, teamID, statusID uuid.UUID) (string, error) {
	if statusID != r.statusID {
		return "", sql.ErrNoRows
	}
	return "unstarted", nil
}

type storyServiceStub struct {
	created []StoryDraft
	stories map[uuid.UUID]StoryPlan
	deleted []uuid.UUID
}

type eventPublisherStub struct {
	events []events.Event
}

func (p *eventPublisherStub) Publish(_ context.Context, event events.Event) error {
	p.events = append(p.events, event)
	return nil
}

func (s *storyServiceStub) CreateFromFeedback(_ context.Context, workspaceID, _ uuid.UUID, draft StoryDraft) (StoryPlan, error) {
	s.created = append(s.created, draft)
	story := StoryPlan{ID: uuid.New(), WorkspaceID: workspaceID, TeamID: draft.TeamID, StatusID: draft.StatusID}
	if s.stories == nil {
		s.stories = make(map[uuid.UUID]StoryPlan)
	}
	s.stories[story.ID] = story
	return story, nil
}

func (s *storyServiceStub) GetForFeedback(_ context.Context, workspaceID, id uuid.UUID) (StoryPlan, error) {
	story, ok := s.stories[id]
	if !ok || story.WorkspaceID != workspaceID {
		return StoryPlan{}, sql.ErrNoRows
	}
	return story, nil
}

func (s *storyServiceStub) DeleteCreatedFromFeedback(_ context.Context, _ uuid.UUID, id, _ uuid.UUID) error {
	s.deleted = append(s.deleted, id)
	delete(s.stories, id)
	return nil
}
