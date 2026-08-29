package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

func (m *storyMutationExecutor) proposeUpdate(
	ctx context.Context,
	executor *FortyOneToolExecutor,
	scope ToolScope,
	raw json.RawMessage,
) (json.RawMessage, error) {
	var args struct {
		StoryID                  *string   `json:"story_id"`
		StoryReference           *string   `json:"story_reference"`
		Title                    *string   `json:"title"`
		Priority                 *string   `json:"priority"`
		Assignee                 string    `json:"assignee"`
		StatusID                 *string   `json:"status_id"`
		SprintID                 *string   `json:"sprint_id"`
		ObjectiveID              *string   `json:"objective_id"`
		KeyResultID              *string   `json:"key_result_id"`
		StartDate                *string   `json:"start_date"`
		EndDate                  *string   `json:"end_date"`
		LabelIDs                 *[]string `json:"label_ids"`
		EstimatedDurationAction  string    `json:"estimated_duration_action"`
		EstimatedDurationMinutes *int      `json:"estimated_duration_minutes"`
		MinimumFocusBlockAction  string    `json:"minimum_focus_block_action"`
		MinimumFocusBlockMinutes *int      `json:"minimum_focus_block_minutes"`
		AutoSchedulingEnabled    *bool     `json:"auto_scheduling_enabled"`
		AutoSchedulingLocked     *bool     `json:"auto_scheduling_locked"`
	}
	if err := decodeToolArguments(
		raw,
		&args,
		"story_id",
		"story_reference",
		"title",
		"priority",
		"assignee",
		"estimated_duration_action",
		"estimated_duration_minutes",
		"minimum_focus_block_action",
		"minimum_focus_block_minutes",
	); err != nil {
		return nil, err
	}
	if (args.StoryID == nil) == (args.StoryReference == nil) {
		return nil, fmt.Errorf("%w: provide exactly one of story_id or story_reference", ErrInvalidToolArguments)
	}
	if args.Assignee != assigneeActionUnchanged && args.Assignee != assigneeActionMe && args.Assignee != assigneeActionUnassigned {
		return nil, fmt.Errorf("%w: assignee must be unchanged, me, or unassigned", ErrInvalidToolArguments)
	}

	_, joinedByID, err := executor.joinedTeams(ctx, scope)
	if err != nil {
		return nil, err
	}
	story, err := m.resolveUpdateStory(ctx, scope, joinedByID, args.StoryID, args.StoryReference)
	if err != nil {
		return nil, err
	}
	team, allowed := joinedByID[story.Team]
	if !allowed || story.Workspace != scope.WorkspaceID {
		return nil, fmt.Errorf("%w: %s", ErrTeamNotAccessible, story.Team)
	}

	var title *string
	if args.Title != nil {
		value, err := normalizedStoryTitle(*args.Title)
		if err != nil {
			return nil, err
		}
		title = &value
	}
	var priority *string
	if args.Priority != nil {
		value, err := normalizedStoryPriority(args.Priority, "")
		if err != nil {
			return nil, err
		}
		priority = &value
	}
	statusID, err := parseOptionalUUID(args.StatusID, "status_id")
	if err != nil {
		return nil, err
	}
	sprintID, err := parseOptionalUUID(args.SprintID, "sprint_id")
	if err != nil {
		return nil, err
	}
	objectiveID, err := parseOptionalUUID(args.ObjectiveID, "objective_id")
	if err != nil {
		return nil, err
	}
	keyResultID, err := parseOptionalUUID(args.KeyResultID, "key_result_id")
	if err != nil {
		return nil, err
	}
	var labelIDs []uuid.UUID
	if args.LabelIDs != nil {
		labelIDs, err = parseUUIDList(*args.LabelIDs, "label_ids")
		if err != nil {
			return nil, err
		}
	}
	startDate, err := parseOptionalDate(args.StartDate, "start_date", scope.Timezone)
	if err != nil {
		return nil, err
	}
	endDate, err := parseOptionalDate(args.EndDate, "end_date", scope.Timezone)
	if err != nil {
		return nil, err
	}
	changedFields, err := proposedChangedFields(
		story,
		scope.UserID,
		title,
		priority,
		args.Assignee,
		statusID,
		sprintID,
		objectiveID,
		keyResultID,
		startDate,
		endDate,
		storyTimeMutation{
			estimatedDurationAction:  args.EstimatedDurationAction,
			estimatedDurationMinutes: args.EstimatedDurationMinutes,
			minimumFocusBlockAction:  args.MinimumFocusBlockAction,
			minimumFocusBlockMinutes: args.MinimumFocusBlockMinutes,
		},
		storyAutoSchedulingMutation{
			enabled: args.AutoSchedulingEnabled,
			locked:  args.AutoSchedulingLocked,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidToolArguments, err)
	}
	if args.LabelIDs != nil && !sameUUIDSet(story.Labels, labelIDs) {
		changedFields = append(changedFields, "labels")
	}
	if len(changedFields) == 0 {
		return nil, fmt.Errorf("%w: update_story must contain at least one effective change", ErrInvalidToolArguments)
	}

	confirmationID, err := uuid.NewRandomFromReader(m.random)
	if err != nil {
		return nil, fmt.Errorf("generate story mutation confirmation ID: %w", err)
	}
	now := m.now().UTC()
	expectedUpdatedAt := story.UpdatedAt.UTC()
	claims := storyMutationClaims{
		Version:                  storyMutationTokenVersion,
		ConfirmationID:           confirmationID,
		Operation:                StoryMutationUpdate,
		WorkspaceID:              scope.WorkspaceID,
		UserID:                   scope.UserID,
		TeamID:                   story.Team,
		StoryID:                  &story.ID,
		ExpectedUpdatedAt:        &expectedUpdatedAt,
		Title:                    title,
		Priority:                 priority,
		AssigneeAction:           args.Assignee,
		StatusID:                 statusID,
		SprintID:                 sprintID,
		ObjectiveID:              objectiveID,
		KeyResultID:              keyResultID,
		StartDate:                startDate,
		EndDate:                  endDate,
		LabelIDs:                 labelIDs,
		EstimatedDurationAction:  args.EstimatedDurationAction,
		EstimatedDurationMinutes: args.EstimatedDurationMinutes,
		MinimumFocusBlockAction:  args.MinimumFocusBlockAction,
		MinimumFocusBlockMinutes: args.MinimumFocusBlockMinutes,
		AutoSchedulingEnabled:    cloneBoolPointer(args.AutoSchedulingEnabled),
		AutoSchedulingLocked:     cloneBoolPointer(args.AutoSchedulingLocked),
		ExpiresAt:                now.Add(storyMutationConfirmationTTL),
	}
	previewTitle := story.Title
	if title != nil {
		previewTitle = *title
	}
	reference := storyReference(team.Code, story.SequenceID)
	storyIDCopy := story.ID
	return m.marshalProposal(ctx, claims, StoryMutationPreview{
		StoryID:                  &storyIDCopy,
		Reference:                reference,
		TeamID:                   team.ID,
		TeamName:                 team.Name,
		TeamCode:                 strings.ToUpper(team.Code),
		Title:                    previewTitle,
		Priority:                 priority,
		AssigneeAction:           args.Assignee,
		ChangedFields:            changedFields,
		EstimatedDurationMinutes: args.EstimatedDurationMinutes,
		MinimumFocusBlockMinutes: args.MinimumFocusBlockMinutes,
		AutoSchedulingEnabled:    cloneBoolPointer(args.AutoSchedulingEnabled),
		AutoSchedulingLocked:     cloneBoolPointer(args.AutoSchedulingLocked),
	}, fmt.Sprintf("Update %s?", reference))
}

func (m *storyMutationExecutor) proposeRelationship(ctx context.Context, executor *FortyOneToolExecutor, scope ToolScope, raw json.RawMessage) (json.RawMessage, error) {
	var args struct {
		FromStoryID        *string `json:"from_story_id"`
		FromStoryReference *string `json:"from_story_reference"`
		ToStoryID          *string `json:"to_story_id"`
		ToStoryReference   *string `json:"to_story_reference"`
		AssociationType    string  `json:"association_type"`
	}
	if err := decodeToolArguments(raw, &args, "from_story_id", "from_story_reference", "to_story_id", "to_story_reference", "association_type"); err != nil {
		return nil, err
	}
	if (args.FromStoryID == nil) == (args.FromStoryReference == nil) || (args.ToStoryID == nil) == (args.ToStoryReference == nil) {
		return nil, fmt.Errorf("%w: provide exactly one identifier for both stories", ErrInvalidToolArguments)
	}
	if args.AssociationType != "blocking" && args.AssociationType != "related" && args.AssociationType != "duplicate" {
		return nil, fmt.Errorf("%w: association_type must be blocking, related, or duplicate", ErrInvalidToolArguments)
	}
	_, joinedByID, err := executor.joinedTeams(ctx, scope)
	if err != nil {
		return nil, err
	}
	from, err := m.resolveUpdateStory(ctx, scope, joinedByID, args.FromStoryID, args.FromStoryReference)
	if err != nil {
		return nil, err
	}
	to, err := m.resolveUpdateStory(ctx, scope, joinedByID, args.ToStoryID, args.ToStoryReference)
	if err != nil {
		return nil, err
	}
	if from.ID == to.ID {
		return nil, fmt.Errorf("%w: stories must be different", ErrInvalidToolArguments)
	}
	confirmationID, err := uuid.NewRandomFromReader(m.random)
	if err != nil {
		return nil, fmt.Errorf("generate relationship confirmation ID: %w", err)
	}
	now := m.now().UTC()
	claims := storyMutationClaims{Version: storyMutationTokenVersion, ConfirmationID: confirmationID, Operation: StoryMutationRelation, WorkspaceID: scope.WorkspaceID, UserID: scope.UserID, TeamID: from.Team, StoryID: &from.ID, RelationStoryID: &to.ID, RelationType: args.AssociationType, ExpiresAt: now.Add(storyMutationConfirmationTTL)}
	team := joinedByID[from.Team]
	return m.marshalProposal(ctx, claims, StoryMutationPreview{StoryID: &from.ID, Reference: storyReference(team.Code, from.SequenceID), TeamID: team.ID, TeamName: team.Name, TeamCode: strings.ToUpper(team.Code), Title: from.Title, ChangedFields: []string{"relationship:" + args.AssociationType}}, fmt.Sprintf("Add a %s relationship from %s to %s?", args.AssociationType, storyReference(team.Code, from.SequenceID), storyReference(joinedByID[to.Team].Code, to.SequenceID)))
}

func (m *storyMutationExecutor) proposeComment(ctx context.Context, executor *FortyOneToolExecutor, scope ToolScope, raw json.RawMessage) (json.RawMessage, error) {
	var args struct {
		StoryID        *string  `json:"story_id"`
		StoryReference *string  `json:"story_reference"`
		Body           string   `json:"body"`
		MentionIDs     []string `json:"mention_ids"`
	}
	if err := decodeToolArguments(raw, &args, "story_id", "story_reference", "body", "mention_ids"); err != nil {
		return nil, err
	}
	if (args.StoryID == nil) == (args.StoryReference == nil) {
		return nil, fmt.Errorf("%w: provide exactly one of story_id or story_reference", ErrInvalidToolArguments)
	}
	_, joinedByID, err := executor.joinedTeams(ctx, scope)
	if err != nil {
		return nil, err
	}
	story, err := m.resolveUpdateStory(ctx, scope, joinedByID, args.StoryID, args.StoryReference)
	if err != nil {
		return nil, err
	}
	body := strings.TrimSpace(args.Body)
	if body == "" || len([]rune(body)) > maxStoryDescriptionRunes {
		return nil, fmt.Errorf("%w: body must contain 1-%d characters", ErrInvalidToolArguments, maxStoryDescriptionRunes)
	}
	mentions, err := parseUUIDList(args.MentionIDs, "mention_ids")
	if err != nil {
		return nil, err
	}
	confirmationID, err := uuid.NewRandomFromReader(m.random)
	if err != nil {
		return nil, fmt.Errorf("generate comment confirmation ID: %w", err)
	}
	now := m.now().UTC()
	claims := storyMutationClaims{Version: storyMutationTokenVersion, ConfirmationID: confirmationID, Operation: StoryMutationComment, WorkspaceID: scope.WorkspaceID, UserID: scope.UserID, TeamID: story.Team, StoryID: &story.ID, Comment: &body, MentionIDs: mentions, ExpiresAt: now.Add(storyMutationConfirmationTTL)}
	team := joinedByID[story.Team]
	return m.marshalProposal(ctx, claims, StoryMutationPreview{StoryID: &story.ID, Reference: storyReference(team.Code, story.SequenceID), TeamID: team.ID, TeamName: team.Name, TeamCode: strings.ToUpper(team.Code), Title: story.Title, ChangedFields: []string{"comment"}}, fmt.Sprintf("Add a comment to %s?", storyReference(team.Code, story.SequenceID)))
}

func (m *storyMutationExecutor) resolveUpdateStory(
	ctx context.Context,
	scope ToolScope,
	joinedByID map[uuid.UUID]messagingTeam,
	storyIDRaw, storyReferenceRaw *string,
) (messagingStory, error) {
	if storyIDRaw != nil {
		storyID, err := parseRequiredUUID(*storyIDRaw, "story_id")
		if err != nil {
			return messagingStory{}, err
		}
		story, err := m.stories.Get(ctx, storyID, scope.WorkspaceID)
		if err != nil {
			return messagingStory{}, fmt.Errorf("load story for update proposal: %w", err)
		}
		return story, nil
	}

	reference, teamCode, err := normalizedStoryReference(*storyReferenceRaw)
	if err != nil {
		return messagingStory{}, err
	}
	var expectedTeamID uuid.UUID
	for teamID, team := range joinedByID {
		if strings.EqualFold(team.Code, teamCode) {
			expectedTeamID = teamID
			break
		}
	}
	if expectedTeamID == uuid.Nil {
		return messagingStory{}, fmt.Errorf("%w: team code %s", ErrTeamNotAccessible, teamCode)
	}
	story, err := m.stories.QueryByRef(ctx, scope.WorkspaceID, reference)
	if err != nil {
		return messagingStory{}, fmt.Errorf("load story %s for update proposal: %w", reference, err)
	}
	if story.Team != expectedTeamID {
		return messagingStory{}, fmt.Errorf("%w: story team does not match reference", ErrInvalidToolArguments)
	}
	return story, nil
}

func (m *storyMutationExecutor) marshalProposal(
	ctx context.Context,
	claims storyMutationClaims,
	preview StoryMutationPreview,
	prompt string,
) (json.RawMessage, error) {
	token, err := m.signClaims(claims)
	if err != nil {
		return nil, err
	}
	if err := m.store.RegisterStoryMutationConfirmation(ctx, StoryMutationConfirmationStateInput{
		ConfirmationID: claims.ConfirmationID,
		WorkspaceID:    claims.WorkspaceID,
		UserID:         claims.UserID,
		TeamID:         claims.TeamID,
		Operation:      claims.Operation,
		TokenHash:      storyMutationTokenHash(token),
		ExpiresAt:      claims.ExpiresAt,
	}); err != nil {
		return nil, fmt.Errorf("register story mutation confirmation: %w", err)
	}
	return marshalToolResult(storyMutationConfirmationToolResult{
		Kind: storyMutationConfirmationKind,
		Confirmation: StoryMutationConfirmation{
			Operation: claims.Operation,
			Token:     token,
			ExpiresAt: claims.ExpiresAt,
			Prompt:    prompt,
			Story:     preview,
		},
	})
}
