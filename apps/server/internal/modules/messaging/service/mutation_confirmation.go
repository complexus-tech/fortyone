package messaging

import (
	"context"
	"errors"
	"fmt"
	"strings"

	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/google/uuid"
)

// ConfirmStoryMutation applies a signed proposal after an explicit provider
// confirmation. It re-authorizes the actor, workspace, joined-team membership,
// and the current channel audience before every write.
func (e *FortyOneToolExecutor) ConfirmStoryMutation(ctx context.Context, scope ToolScope, token string) (StoryMutationResult, error) {
	if e.mutations == nil {
		return StoryMutationResult{}, fmt.Errorf("%w: story mutations are disabled", ErrInvalidConfirmation)
	}
	if err := validateToolScope(&scope); err != nil {
		return StoryMutationResult{}, err
	}
	if !scope.AllowMutations {
		return StoryMutationResult{}, ErrMutationNotAllowed
	}
	ctx = platformauth.SetUserID(ctx, scope.UserID)
	if confirmationID, opaque, err := batchConfirmationID(token); err != nil {
		return StoryMutationResult{}, err
	} else if opaque {
		return e.confirmStoryBatch(ctx, scope, confirmationID, token)
	}
	claims, err := e.mutations.verifyClaims(token)
	if err != nil {
		return StoryMutationResult{}, err
	}
	if claims.WorkspaceID != scope.WorkspaceID || claims.UserID != scope.UserID {
		return StoryMutationResult{}, fmt.Errorf("%w: confirmation identity does not match", ErrInvalidConfirmation)
	}
	result, duplicate, err := e.mutations.store.ApplyStoryMutationConfirmation(
		ctx,
		storyMutationConfirmationBinding(claims, token),
		e.mutations.now().UTC(),
		func(applyCtx context.Context) (StoryMutationResult, error) {
			_, joinedByID, err := e.joinedTeams(applyCtx, scope)
			if err != nil {
				return StoryMutationResult{}, err
			}
			team, allowed := joinedByID[claims.TeamID]
			if !allowed {
				return StoryMutationResult{}, fmt.Errorf("%w: %s", ErrTeamNotAccessible, claims.TeamID)
			}

			switch claims.Operation {
			case StoryMutationCreate:
				return e.mutations.confirmCreate(applyCtx, scope, team, claims)
			case StoryMutationUpdate:
				return e.mutations.confirmUpdate(applyCtx, scope, team, claims)
			case StoryMutationComment:
				return e.mutations.confirmComment(applyCtx, scope, team, claims)
			case StoryMutationRelation:
				return e.mutations.confirmRelationship(applyCtx, scope, team, claims)
			default:
				return StoryMutationResult{}, fmt.Errorf("%w: unsupported operation", ErrInvalidConfirmation)
			}
		},
	)
	if err != nil {
		return StoryMutationResult{}, err
	}
	if duplicate {
		result.Status = storyMutationStatusAlreadyApplied
	}
	return result, nil
}

func (e *FortyOneToolExecutor) confirmStoryBatch(
	ctx context.Context,
	scope ToolScope,
	confirmationID uuid.UUID,
	token string,
) (StoryMutationResult, error) {
	binding := StoryMutationConfirmationBinding{
		ConfirmationID: confirmationID,
		WorkspaceID:    scope.WorkspaceID,
		UserID:         scope.UserID,
		TokenHash:      storyMutationTokenHash(token),
	}
	record, err := e.mutations.store.LoadStoryMutationConfirmation(ctx, binding)
	if err != nil {
		return StoryMutationResult{}, err
	}
	if record.Operation != StoryMutationCreateBatch || record.TeamID == uuid.Nil {
		return StoryMutationResult{}, fmt.Errorf("%w: opaque token is not a batch story proposal", ErrInvalidConfirmation)
	}
	switch record.Status {
	case StoryMutationConfirmationCancelled:
		return StoryMutationResult{}, ErrCancelledConfirmation
	case StoryMutationConfirmationExpired:
		return StoryMutationResult{}, ErrExpiredConfirmation
	case StoryMutationConfirmationPending, StoryMutationConfirmationApplied:
	default:
		return StoryMutationResult{}, fmt.Errorf("%w: unsupported batch confirmation status %q", ErrInvalidConfirmation, record.Status)
	}
	if len(record.Proposal) == 0 {
		if record.Status == StoryMutationConfirmationApplied && record.Result != nil && record.LastError == "" {
			result := *record.Result
			result.Status = storyMutationStatusAlreadyApplied
			return result, nil
		}
		return StoryMutationResult{}, fmt.Errorf("%w: batch proposal is no longer available", ErrInvalidConfirmation)
	}
	proposal, err := decodeBatchStoryProposal(record.Proposal)
	if err != nil {
		return StoryMutationResult{}, err
	}

	result, duplicate, err := e.mutations.store.ApplyStoryMutationConfirmation(
		ctx,
		binding,
		e.mutations.now().UTC(),
		func(applyCtx context.Context) (StoryMutationResult, error) {
			_, joinedByID, err := e.joinedTeams(applyCtx, scope)
			if err != nil {
				return StoryMutationResult{}, err
			}
			team, allowed := joinedByID[record.TeamID]
			if !allowed {
				return StoryMutationResult{}, fmt.Errorf("%w: %s", ErrTeamNotAccessible, record.TeamID)
			}
			return e.mutations.confirmCreateBatch(applyCtx, e, scope, team, confirmationID, proposal)
		},
	)
	if err != nil {
		return result, err
	}
	if duplicate {
		result.Status = storyMutationStatusAlreadyApplied
	}
	return result, nil
}

// CancelStoryMutation atomically consumes a pending proposal without invoking
// its write callback. Only the workspace/user identity bound into the signed
// token can cancel it; a later Confirm therefore cannot mutate.
func (e *FortyOneToolExecutor) CancelStoryMutation(
	ctx context.Context,
	scope ToolScope,
	token string,
) (StoryMutationCancellationResult, error) {
	if e.mutations == nil {
		return StoryMutationCancellationResult{}, fmt.Errorf("%w: story mutations are disabled", ErrInvalidConfirmation)
	}
	if err := validateToolScope(&scope); err != nil {
		return StoryMutationCancellationResult{}, err
	}
	if confirmationID, opaque, err := batchConfirmationID(token); err != nil {
		return StoryMutationCancellationResult{}, err
	} else if opaque {
		return e.mutations.store.CancelStoryMutationConfirmation(
			ctx,
			StoryMutationConfirmationBinding{
				ConfirmationID: confirmationID,
				WorkspaceID:    scope.WorkspaceID,
				UserID:         scope.UserID,
				TokenHash:      storyMutationTokenHash(token),
			},
			e.mutations.now().UTC(),
		)
	}
	claims, err := e.mutations.verifyClaims(token)
	if err != nil {
		return StoryMutationCancellationResult{}, err
	}
	if claims.WorkspaceID != scope.WorkspaceID || claims.UserID != scope.UserID {
		return StoryMutationCancellationResult{}, fmt.Errorf("%w: confirmation identity does not match", ErrInvalidConfirmation)
	}
	return e.mutations.store.CancelStoryMutationConfirmation(
		ctx,
		storyMutationConfirmationBinding(claims, token),
		e.mutations.now().UTC(),
	)
}

func (m *storyMutationExecutor) confirmCreate(
	ctx context.Context,
	scope ToolScope,
	team messagingTeam,
	claims storyMutationClaims,
) (StoryMutationResult, error) {
	if claims.Title == nil || claims.Priority == nil || claims.StoryID != nil || claims.ExpectedUpdatedAt != nil {
		return StoryMutationResult{}, fmt.Errorf("%w: malformed create proposal", ErrInvalidConfirmation)
	}
	statusID, err := m.stories.FindFirstStatusByCategory(ctx, team.ID, scope.WorkspaceID, "unstarted")
	if err != nil {
		return StoryMutationResult{}, fmt.Errorf("resolve default story status: %w", err)
	}
	if statusID == nil {
		return StoryMutationResult{}, errors.New("team has no unstarted story status")
	}

	var assigneeID *uuid.UUID
	if claims.AssigneeAction == assigneeActionMe {
		userID := scope.UserID
		assigneeID = &userID
	} else if claims.AssigneeAction != assigneeActionUnassigned {
		return StoryMutationResult{}, fmt.Errorf("%w: invalid create assignee action", ErrInvalidConfirmation)
	}
	creationKey := "messaging:create-story:" + claims.ConfirmationID.String()
	story, err := m.stories.CreateExternalUserAction(ctx, scope.UserID, messagingNewStory{
		Title:                    *claims.Title,
		Status:                   statusID,
		Assignee:                 assigneeID,
		Reporter:                 &scope.UserID,
		Priority:                 *claims.Priority,
		Team:                     team.ID,
		EstimatedDurationMinutes: claims.EstimatedDurationMinutes,
		MinimumFocusBlockMinutes: claims.MinimumFocusBlockMinutes,
		AutoSchedulingEnabled:    optionalBoolValue(claims.AutoSchedulingEnabled),
		CreationKey:              &creationKey,
	}, scope.WorkspaceID)
	if err != nil {
		return StoryMutationResult{}, fmt.Errorf("create confirmed story: %w", err)
	}
	status := storyMutationStatusAlreadyApplied
	if story.CreatedNow {
		status = storyMutationStatusApplied
	}
	return storyMutationResult(status, StoryMutationCreate, story, team.Code), nil
}

func (m *storyMutationExecutor) confirmCreateBatch(
	ctx context.Context,
	executor *FortyOneToolExecutor,
	scope ToolScope,
	team messagingTeam,
	confirmationID uuid.UUID,
	proposal batchStoryMutationProposal,
) (StoryMutationResult, error) {
	result := StoryMutationResult{
		Status:    storyMutationStatusPartial,
		Operation: StoryMutationCreateBatch,
		TeamID:    team.ID,
		Items:     make([]StoryMutationItemResult, 0, len(proposal.Items)),
	}
	if confirmationID == uuid.Nil || team.ID == uuid.Nil {
		return result, fmt.Errorf("%w: malformed batch create proposal", ErrInvalidConfirmation)
	}

	// Validate every mutable dependency before the first write. A removed team
	// member or malformed later item must never leave a partially-created batch.
	statusID, err := m.stories.FindFirstStatusByCategory(ctx, team.ID, scope.WorkspaceID, "unstarted")
	if err != nil {
		return result, fmt.Errorf("resolve default story status: %w", err)
	}
	if statusID == nil || *statusID == uuid.Nil {
		return result, errors.New("team has no unstarted story status")
	}
	validated := make([]batchStoryMutationItem, len(proposal.Items))
	for index, item := range proposal.Items {
		title, err := normalizedStoryTitle(item.Title)
		if err != nil || title != item.Title {
			return result, fmt.Errorf("%w: invalid title in batch item %d", ErrInvalidConfirmation, index)
		}
		description, err := normalizedBatchStoryDescriptionValue(item.Description)
		if err != nil || description != item.Description {
			return result, fmt.Errorf("%w: invalid description in batch item %d", ErrInvalidConfirmation, index)
		}
		priority, err := normalizedStoryPriority(&item.Priority, "")
		if err != nil || priority != item.Priority {
			return result, fmt.Errorf("%w: invalid priority in batch item %d", ErrInvalidConfirmation, index)
		}
		if err := validateStoryTimeContract(item.EstimatedDurationMinutes, item.MinimumFocusBlockMinutes); err != nil {
			return result, fmt.Errorf("%w: invalid time contract in batch item %d: %v", ErrInvalidConfirmation, index, err)
		}
		if err := validateStoryAutoSchedulingContract(item.AutoSchedulingEnabled, false, storyAutoSchedulingStatusOff); err != nil {
			return result, fmt.Errorf("%w: invalid auto-scheduling contract in batch item %d: %v", ErrInvalidConfirmation, index, err)
		}
		if item.AssigneeID != nil {
			if executor.users == nil {
				return result, fmt.Errorf("%w: named assignees are unavailable", ErrInvalidConfirmation)
			}
			member, err := executor.activeTeamMemberByID(ctx, scope.WorkspaceID, team.ID, *item.AssigneeID)
			if err != nil {
				return result, err
			}
			if member == nil {
				return result, fmt.Errorf("%w: batch item %d assignee is no longer an active team member", ErrInvalidConfirmation, index)
			}
		}
		validated[index] = batchStoryMutationItem{
			Title:                    title,
			Description:              description,
			Priority:                 priority,
			AssigneeID:               cloneUUIDPointer(item.AssigneeID),
			EstimatedDurationMinutes: cloneIntPointer(item.EstimatedDurationMinutes),
			MinimumFocusBlockMinutes: cloneIntPointer(item.MinimumFocusBlockMinutes),
			AutoSchedulingEnabled:    item.AutoSchedulingEnabled,
		}
	}

	result.Status = storyMutationStatusAlreadyApplied
	for index, item := range validated {
		description := batchStoryDescription(item.Description, proposal.SourceURL)
		var descriptionPointer *string
		if description != "" {
			descriptionPointer = &description
		}
		creationKey := fmt.Sprintf("messaging:create-story:%s:%d", confirmationID, index)
		story, err := m.stories.CreateExternalUserAction(ctx, scope.UserID, messagingNewStory{
			Title:                    item.Title,
			Description:              descriptionPointer,
			Status:                   statusID,
			Assignee:                 cloneUUIDPointer(item.AssigneeID),
			Reporter:                 &scope.UserID,
			Priority:                 item.Priority,
			Team:                     team.ID,
			EstimatedDurationMinutes: item.EstimatedDurationMinutes,
			MinimumFocusBlockMinutes: item.MinimumFocusBlockMinutes,
			AutoSchedulingEnabled:    item.AutoSchedulingEnabled,
			CreationKey:              &creationKey,
		}, scope.WorkspaceID)
		if err != nil {
			result.Status = storyMutationStatusPartial
			return result, fmt.Errorf("create confirmed batch story %d: %w", index, err)
		}
		itemStatus := storyMutationStatusAlreadyApplied
		if story.CreatedNow {
			itemStatus = storyMutationStatusApplied
			result.Status = storyMutationStatusApplied
		}
		result.Items = append(result.Items, StoryMutationItemResult{
			Index:                    index,
			Status:                   itemStatus,
			StoryID:                  story.ID,
			Reference:                storyReference(team.Code, story.SequenceID),
			TeamID:                   team.ID,
			Title:                    story.Title,
			Priority:                 story.Priority,
			AssigneeID:               cloneUUIDPointer(story.Assignee),
			EstimatedDurationMinutes: cloneIntPointer(story.EstimatedDurationMinutes),
			MinimumFocusBlockMinutes: cloneIntPointer(story.MinimumFocusBlockMinutes),
			AutoSchedulingEnabled:    story.AutoSchedulingEnabled,
			AutoSchedulingLocked:     story.AutoSchedulingLocked,
			AutoSchedulingStatus:     story.AutoSchedulingStatus,
			AutoSchedulingReason:     story.AutoSchedulingReason,
			AutoSchedulingUpdatedAt:  story.AutoSchedulingUpdatedAt,
		})
	}
	return result, nil
}

func (m *storyMutationExecutor) confirmUpdate(
	ctx context.Context,
	scope ToolScope,
	team messagingTeam,
	claims storyMutationClaims,
) (StoryMutationResult, error) {
	if claims.StoryID == nil || claims.ExpectedUpdatedAt == nil || claims.ConfirmationID == uuid.Nil {
		return StoryMutationResult{}, fmt.Errorf("%w: malformed update proposal", ErrInvalidConfirmation)
	}
	story, err := m.stories.Get(ctx, *claims.StoryID, scope.WorkspaceID)
	if err != nil {
		return StoryMutationResult{}, fmt.Errorf("load story for confirmed update: %w", err)
	}
	if story.Team != team.ID || story.Workspace != scope.WorkspaceID {
		return StoryMutationResult{}, fmt.Errorf("%w: story team does not match proposal", ErrInvalidConfirmation)
	}

	updates, err := desiredStoryUpdates(
		story,
		scope.UserID,
		claims.Title,
		claims.Priority,
		claims.AssigneeAction,
		claims.StatusID,
		claims.SprintID,
		claims.ObjectiveID,
		claims.KeyResultID,
		claims.StartDate,
		claims.EndDate,
		storyTimeMutation{
			estimatedDurationAction:  claims.EstimatedDurationAction,
			estimatedDurationMinutes: claims.EstimatedDurationMinutes,
			minimumFocusBlockAction:  claims.MinimumFocusBlockAction,
			minimumFocusBlockMinutes: claims.MinimumFocusBlockMinutes,
		},
		storyAutoSchedulingMutation{
			enabled: claims.AutoSchedulingEnabled,
			locked:  claims.AutoSchedulingLocked,
		},
	)
	if err != nil {
		return StoryMutationResult{}, fmt.Errorf("%w: invalid confirmed story time update: %v", ErrInvalidConfirmation, err)
	}
	labelsChanged := claims.LabelIDs != nil && !sameUUIDSet(story.Labels, claims.LabelIDs)
	if len(updates) == 0 && !labelsChanged {
		return storyMutationResult(storyMutationStatusAlreadyApplied, StoryMutationUpdate, story, team.Code), nil
	}
	if !story.UpdatedAt.Equal(claims.ExpectedUpdatedAt.UTC()) {
		return StoryMutationResult{}, fmt.Errorf("%w: %s", ErrStaleMutation, storyReference(team.Code, story.SequenceID))
	}
	if len(updates) > 0 {
		if err := m.stories.UpdateExternalUserActionIfUnchanged(ctx, scope.UserID, story.ID, scope.WorkspaceID, claims.ExpectedUpdatedAt.UTC(), updates); err != nil {
			if errors.Is(err, storyChangedError()) {
				return StoryMutationResult{}, fmt.Errorf("%w: %s", ErrStaleMutation, storyReference(team.Code, story.SequenceID))
			}
			return StoryMutationResult{}, fmt.Errorf("update confirmed story: %w", err)
		}
	}
	if labelsChanged {
		if err := m.stories.UpdateLabels(ctx, story.ID, scope.WorkspaceID, claims.LabelIDs); err != nil {
			return StoryMutationResult{}, fmt.Errorf("update confirmed story labels: %w", err)
		}
	}
	updated, err := m.stories.Get(ctx, story.ID, scope.WorkspaceID)
	if err != nil {
		return StoryMutationResult{}, fmt.Errorf("reload confirmed story: %w", err)
	}
	return storyMutationResult(storyMutationStatusApplied, StoryMutationUpdate, updated, team.Code), nil
}

func (m *storyMutationExecutor) confirmComment(ctx context.Context, scope ToolScope, team messagingTeam, claims storyMutationClaims) (StoryMutationResult, error) {
	if claims.StoryID == nil || claims.Comment == nil || strings.TrimSpace(*claims.Comment) == "" {
		return StoryMutationResult{}, fmt.Errorf("%w: malformed comment proposal", ErrInvalidConfirmation)
	}
	comment, err := m.stories.CreateCommentExternal(ctx, scope.UserID, scope.WorkspaceID, messagingNewComment{StoryID: *claims.StoryID, UserID: scope.UserID, Comment: *claims.Comment, Mentions: claims.MentionIDs})
	if err != nil {
		return StoryMutationResult{}, fmt.Errorf("create confirmed story comment: %w", err)
	}
	commentID := comment.ID
	return StoryMutationResult{Status: storyMutationStatusApplied, Operation: StoryMutationComment, StoryID: *claims.StoryID, CommentID: &commentID, TeamID: team.ID}, nil
}

func (m *storyMutationExecutor) confirmRelationship(ctx context.Context, scope ToolScope, team messagingTeam, claims storyMutationClaims) (StoryMutationResult, error) {
	if claims.StoryID == nil || claims.RelationStoryID == nil || claims.RelationType == "" {
		return StoryMutationResult{}, fmt.Errorf("%w: malformed relationship proposal", ErrInvalidConfirmation)
	}
	association, err := m.stories.AddAssociation(ctx, *claims.StoryID, *claims.RelationStoryID, claims.RelationType, scope.WorkspaceID)
	if err != nil {
		return StoryMutationResult{}, fmt.Errorf("create confirmed story relationship: %w", err)
	}
	return StoryMutationResult{Status: storyMutationStatusApplied, Operation: StoryMutationRelation, StoryID: *claims.StoryID, AssociationID: &association.ID, TeamID: team.ID}, nil
}
