package feedback

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

func (s *Service) ToggleVote(ctx context.Context, workspaceID, itemID, userID uuid.UUID, vote int) (CoreVoteResult, error) {
	if workspaceID == uuid.Nil || itemID == uuid.Nil || userID == uuid.Nil {
		return CoreVoteResult{}, invalidInput("workspace, feedback, and user are required")
	}
	if vote != -1 && vote != 1 {
		return CoreVoteResult{}, invalidInput("feedback vote must be either -1 or 1")
	}
	item, err := s.getInternalItem(ctx, workspaceID, itemID, userID)
	if err != nil {
		return CoreVoteResult{}, err
	}
	if item.MergedIntoItemID != nil {
		return CoreVoteResult{}, ErrMergeConflict
	}
	scope, scoped, err := s.coreScope(ctx, workspaceID, userID)
	if err != nil {
		return CoreVoteResult{}, err
	}
	if scoped {
		return s.scopedCoreRepo.ToggleVoteScoped(ctx, scope, itemID, vote)
	}
	return s.repo.ToggleVote(ctx, workspaceID, itemID, userID, vote)
}

func (s *Service) TogglePublicVote(ctx context.Context, input CorePublicVoteInput) (CoreVoteResult, error) {
	portal, err := s.repo.GetPortalBySlug(ctx, strings.TrimSpace(input.PortalSlug))
	if err != nil {
		return CoreVoteResult{}, err
	}
	item, err := s.repo.GetItemByPortal(ctx, portal.ID, input.ItemID)
	if err != nil {
		return CoreVoteResult{}, err
	}
	if item.WorkspaceID != portal.WorkspaceID {
		return CoreVoteResult{}, ErrNotFound
	}
	if item.MergedIntoItemID != nil {
		return CoreVoteResult{}, ErrMergeConflict
	}
	if input.Participant != nil {
		participant := *input.Participant
		if participant.PortalID != portal.ID || participant.ID == uuid.Nil || participant.BlockedAt != nil || participant.Kind == ContributorKindAnonymous {
			return CoreVoteResult{}, ErrAuthenticationRequired
		}
		if portal.ParticipationMode == ParticipationModeAccountRequired {
			return CoreVoteResult{}, ErrParticipationNotAllowed
		}
		if input.Vote != -1 && input.Vote != 1 {
			return CoreVoteResult{}, invalidInput("feedback vote must be either -1 or 1")
		}
		return s.nextRepo.ToggleContributorVote(ctx, CoreContributorVoteInput{
			WorkspaceID: portal.WorkspaceID,
			ItemID:      item.ID,
			Participant: participant,
			Vote:        input.Vote,
		})
	}
	if input.UserID == uuid.Nil {
		return CoreVoteResult{}, ErrAuthenticationRequired
	}

	return s.ToggleVote(ctx, portal.WorkspaceID, item.ID, input.UserID, input.Vote)
}

func (s *Service) LinkStory(ctx context.Context, input CoreStoryLinkInput) (CoreStoryLink, error) {
	if input.WorkspaceID == uuid.Nil || input.ItemID == uuid.Nil || input.StoryID == uuid.Nil || input.CreatedByUserID == uuid.Nil {
		return CoreStoryLink{}, invalidInput("workspace, feedback, story, and actor are required")
	}
	if input.Relationship == "" {
		input.Relationship = RelationshipLinked
	}
	item, err := s.getInternalItem(ctx, input.WorkspaceID, input.ItemID, input.CreatedByUserID)
	if err != nil {
		return CoreStoryLink{}, err
	}
	if item.MergedIntoItemID != nil {
		return CoreStoryLink{}, ErrMergeConflict
	}
	scope, scoped, err := s.coreScope(ctx, input.WorkspaceID, input.CreatedByUserID)
	if err != nil {
		return CoreStoryLink{}, err
	}
	if scoped {
		return s.scopedCoreRepo.LinkStoryScoped(ctx, scope, input)
	}
	return s.repo.LinkStory(ctx, input)
}

func (s *Service) CreateStoryFromItem(ctx context.Context, workspaceID, itemID, actorID uuid.UUID, input CoreCreateStoryInput) (CoreCreateStoryResult, error) {
	if s.stories == nil {
		return CoreCreateStoryResult{}, errors.New("story service is required")
	}
	if workspaceID == uuid.Nil || itemID == uuid.Nil || actorID == uuid.Nil || input.TeamID == uuid.Nil {
		return CoreCreateStoryResult{}, invalidInput("workspace, feedback, actor, and team are required")
	}
	item, err := s.getInternalItem(ctx, workspaceID, itemID, actorID)
	if err != nil {
		return CoreCreateStoryResult{}, err
	}
	if item.MergedIntoItemID != nil {
		return CoreCreateStoryResult{}, ErrMergeConflict
	}
	if item.Board.TeamID == uuid.Nil || item.Board.TeamID != input.TeamID {
		return CoreCreateStoryResult{}, ErrTeamMismatch
	}
	existingLinks, err := s.listInternalStoryLinks(ctx, workspaceID, itemID, actorID)
	if err != nil {
		return CoreCreateStoryResult{}, err
	}
	for _, existingLink := range existingLinks {
		if !existingLink.IsPrimary {
			continue
		}
		if input.StoryID != nil && existingLink.StoryID != *input.StoryID {
			return CoreCreateStoryResult{}, ErrAlreadyPlanned
		}
		if _, err := s.UpdateItemStatus(ctx, workspaceID, itemID, CoreUpdateItemStatusInput{Status: item.Status, ActorID: actorID, AllowLinked: true}); err != nil {
			return CoreCreateStoryResult{}, err
		}
		return CoreCreateStoryResult{
			ItemID:  itemID,
			StoryID: existingLink.StoryID,
			LinkID:  existingLink.ID,
			Created: false,
		}, nil
	}

	if input.StoryID != nil {
		story, err := s.stories.GetForFeedback(ctx, workspaceID, *input.StoryID)
		if err != nil {
			return CoreCreateStoryResult{}, err
		}
		if story.WorkspaceID != workspaceID {
			return CoreCreateStoryResult{}, ErrNotFound
		}
		if story.TeamID != item.Board.TeamID {
			return CoreCreateStoryResult{}, ErrTeamMismatch
		}
		if story.DeletedAt != nil {
			return CoreCreateStoryResult{}, fmt.Errorf("%w: deleted stories cannot be linked to feedback", ErrInvalidInput)
		}
		plannedStatus := StatusPlanned
		if story.StatusID != nil {
			category, err := s.repo.GetStatusCategory(ctx, item.Board.TeamID, *story.StatusID)
			if err != nil {
				return CoreCreateStoryResult{}, err
			}
			plannedStatus = feedbackStatusForStoryCategory(category)
		}
		link, err := s.linkInternalStory(ctx, CoreStoryLinkInput{
			WorkspaceID:     workspaceID,
			ItemID:          itemID,
			StoryID:         story.ID,
			Relationship:    RelationshipSolves,
			IsPrimary:       true,
			CreatedByUserID: actorID,
		})
		if err != nil {
			if errors.Is(err, ErrAlreadyPlanned) {
				return s.resolveExistingPlan(ctx, workspaceID, itemID, input.StoryID)
			}
			return CoreCreateStoryResult{}, err
		}
		if _, err := s.UpdateItemStatus(ctx, workspaceID, itemID, CoreUpdateItemStatusInput{Status: plannedStatus, ActorID: actorID, AllowLinked: true}); err != nil {
			return CoreCreateStoryResult{}, err
		}
		return CoreCreateStoryResult{ItemID: itemID, StoryID: story.ID, LinkID: link.ID, Created: false}, nil
	}

	statusID := input.StatusID
	statusCategory := "unstarted"
	if statusID == nil {
		statusID, err = s.repo.FindFirstStatusByCategory(ctx, input.TeamID, "unstarted")
		if err != nil {
			return CoreCreateStoryResult{}, err
		}
		if statusID == nil {
			return CoreCreateStoryResult{}, invalidInput("team has no unstarted status configured")
		}
	} else {
		statusCategory, err = s.repo.GetStatusCategory(ctx, input.TeamID, *statusID)
		if err != nil {
			return CoreCreateStoryResult{}, invalidInput("story status does not belong to the feedback team")
		}
	}
	story, err := s.stories.CreateFromFeedback(ctx, workspaceID, actorID, StoryDraft{
		Title:       item.Title,
		Description: item.Description,
		StatusID:    statusID,
		ReporterID:  actorID,
		TeamID:      item.Board.TeamID,
	})
	if err != nil {
		return CoreCreateStoryResult{}, err
	}
	link, err := s.linkInternalStory(ctx, CoreStoryLinkInput{
		WorkspaceID:     workspaceID,
		ItemID:          itemID,
		StoryID:         story.ID,
		Relationship:    RelationshipCreatedFrom,
		IsPrimary:       true,
		CreatedByUserID: actorID,
	})
	if err != nil {
		if deleteErr := s.stories.DeleteCreatedFromFeedback(
			context.WithoutCancel(ctx),
			workspaceID,
			story.ID,
			actorID,
		); deleteErr != nil {
			return CoreCreateStoryResult{}, errors.Join(err, fmt.Errorf("compensating story delete: %w", deleteErr))
		}
		if errors.Is(err, ErrAlreadyPlanned) {
			return s.resolveExistingPlan(ctx, workspaceID, itemID, nil)
		}
		return CoreCreateStoryResult{}, err
	}
	if _, err := s.UpdateItemStatus(ctx, workspaceID, itemID, CoreUpdateItemStatusInput{Status: feedbackStatusForStoryCategory(statusCategory), ActorID: actorID, AllowLinked: true}); err != nil {
		return CoreCreateStoryResult{}, err
	}
	return CoreCreateStoryResult{ItemID: itemID, StoryID: story.ID, LinkID: link.ID, Created: true}, nil
}

func (s *Service) resolveExistingPlan(ctx context.Context, workspaceID, itemID uuid.UUID, requestedStoryID *uuid.UUID) (CoreCreateStoryResult, error) {
	links, err := s.listInternalStoryLinks(ctx, workspaceID, itemID, uuid.Nil)
	if err != nil {
		return CoreCreateStoryResult{}, err
	}
	for _, link := range links {
		if !link.IsPrimary {
			continue
		}
		if requestedStoryID != nil && link.StoryID != *requestedStoryID {
			return CoreCreateStoryResult{}, ErrAlreadyPlanned
		}
		return CoreCreateStoryResult{
			ItemID:  itemID,
			StoryID: link.StoryID,
			LinkID:  link.ID,
			Created: false,
		}, nil
	}
	return CoreCreateStoryResult{}, ErrAlreadyPlanned
}

func feedbackStatusForStoryCategory(category string) string {
	switch category {
	case "backlog":
		return StatusReviewing
	case "started":
		return StatusInProgress
	case "completed":
		return StatusCompleted
	case "cancelled":
		return StatusClosed
	case "unstarted", "paused":
		return StatusPlanned
	default:
		return StatusPlanned
	}
}
