package messaging

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"
	"time"

	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	teams "github.com/complexus-tech/projects-api/internal/modules/teams/service"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/google/uuid"
)

const (
	storyMutationConfirmationKind = "story_mutation_confirmation_required"
	storyMutationTokenVersion     = 1
	storyMutationConfirmationTTL  = 10 * time.Minute
	maximumStoryMutationTokenSize = 2_000
	maximumStoryTitleRunes        = 255
	maximumBatchStoryCount        = 10
	maximumBatchDescriptionRunes  = 2_000
	batchStoryProposalVersion     = 1
	batchStoryTokenBytes          = 32

	storyMutationStatusApplied        = "applied"
	storyMutationStatusAlreadyApplied = "already_applied"
	storyMutationStatusPartial        = "partial"

	assigneeActionUnchanged  = "unchanged"
	assigneeActionMe         = "me"
	assigneeActionNamed      = "named"
	assigneeActionUnassigned = "unassigned"

	storyTimeActionUnchanged = "unchanged"
	storyTimeActionSet       = "set"
	storyTimeActionClear     = "clear"
)

var storyPriorities = map[string]struct{}{
	"No Priority": {},
	"Low":         {},
	"Medium":      {},
	"High":        {},
	"Urgent":      {},
}

type storyMutationExecutor struct {
	stories StoryMutationService
	store   StoryMutationConfirmationStore
	key     []byte
	now     func() time.Time
	random  io.Reader
}

type storyMutationClaims struct {
	Version                  int                    `json:"v"`
	ConfirmationID           uuid.UUID              `json:"i"`
	Operation                StoryMutationOperation `json:"o"`
	WorkspaceID              uuid.UUID              `json:"w"`
	UserID                   uuid.UUID              `json:"u"`
	TeamID                   uuid.UUID              `json:"t"`
	StoryID                  *uuid.UUID             `json:"s,omitempty"`
	ExpectedUpdatedAt        *time.Time             `json:"e,omitempty"`
	Title                    *string                `json:"n,omitempty"`
	Priority                 *string                `json:"p,omitempty"`
	AssigneeAction           string                 `json:"a"`
	StatusID                 *uuid.UUID             `json:"st,omitempty"`
	SprintID                 *uuid.UUID             `json:"sp,omitempty"`
	ObjectiveID              *uuid.UUID             `json:"oj,omitempty"`
	KeyResultID              *uuid.UUID             `json:"k,omitempty"`
	StartDate                *time.Time             `json:"sd,omitempty"`
	EndDate                  *time.Time             `json:"ed,omitempty"`
	EstimatedDurationAction  string                 `json:"da,omitempty"`
	EstimatedDurationMinutes *int                   `json:"du,omitempty"`
	MinimumFocusBlockAction  string                 `json:"fa,omitempty"`
	MinimumFocusBlockMinutes *int                   `json:"fb,omitempty"`
	AutoSchedulingEnabled    *bool                  `json:"ae,omitempty"`
	AutoSchedulingLocked     *bool                  `json:"al,omitempty"`
	Comment                  *string                `json:"c,omitempty"`
	MentionIDs               []uuid.UUID            `json:"m,omitempty"`
	LabelIDs                 []uuid.UUID            `json:"l,omitempty"`
	RelationStoryID          *uuid.UUID             `json:"r,omitempty"`
	RelationType             string                 `json:"rt,omitempty"`
	ExpiresAt                time.Time              `json:"x"`
}

type storyTimeMutation struct {
	estimatedDurationAction  string
	estimatedDurationMinutes *int
	minimumFocusBlockAction  string
	minimumFocusBlockMinutes *int
}

type storyAutoSchedulingMutation struct {
	enabled *bool
	locked  *bool
}

type storyMutationConfirmationToolResult struct {
	Kind         string                    `json:"kind"`
	Confirmation StoryMutationConfirmation `json:"confirmation"`
}

type batchStoryMutationProposal struct {
	Version   int                      `json:"version"`
	SourceURL string                   `json:"source_url,omitempty"`
	Items     []batchStoryMutationItem `json:"items"`
}

type batchStoryMutationItem struct {
	Title                    string     `json:"title"`
	Description              string     `json:"description,omitempty"`
	Priority                 string     `json:"priority"`
	AssigneeID               *uuid.UUID `json:"assignee_id,omitempty"`
	EstimatedDurationMinutes *int       `json:"estimated_duration_minutes,omitempty"`
	MinimumFocusBlockMinutes *int       `json:"minimum_focus_block_minutes,omitempty"`
	AutoSchedulingEnabled    bool       `json:"auto_scheduling_enabled"`
}

func newStoryMutationExecutor(
	storiesService StoryMutationService,
	secret string,
	store StoryMutationConfirmationStore,
) *storyMutationExecutor {
	key := sha256.Sum256([]byte("fortyone:messaging:story-mutation:v1\x00" + secret))
	return &storyMutationExecutor{
		stories: storiesService,
		store:   store,
		key:     append([]byte(nil), key[:]...),
		now:     time.Now,
		random:  rand.Reader,
	}
}

func (m *storyMutationExecutor) proposeCreate(
	ctx context.Context,
	executor *FortyOneToolExecutor,
	scope ToolScope,
	raw json.RawMessage,
) (json.RawMessage, error) {
	var args struct {
		TeamID                   string  `json:"team_id"`
		Title                    string  `json:"title"`
		Priority                 *string `json:"priority"`
		Assignee                 string  `json:"assignee"`
		EstimatedDurationMinutes *int    `json:"estimated_duration_minutes"`
		MinimumFocusBlockMinutes *int    `json:"minimum_focus_block_minutes"`
		AutoSchedulingEnabled    bool    `json:"auto_scheduling_enabled"`
	}
	if err := decodeToolArguments(raw, &args, "team_id", "title", "priority", "assignee"); err != nil {
		return nil, err
	}

	_, joinedByID, err := executor.joinedTeams(ctx, scope)
	if err != nil {
		return nil, err
	}
	teamID, err := parseAccessibleTeamID(args.TeamID, joinedByID)
	if err != nil {
		return nil, err
	}
	title, err := normalizedStoryTitle(args.Title)
	if err != nil {
		return nil, err
	}
	priority, err := normalizedStoryPriority(args.Priority, "No Priority")
	if err != nil {
		return nil, err
	}
	if args.Assignee != assigneeActionMe && args.Assignee != assigneeActionUnassigned {
		return nil, fmt.Errorf("%w: assignee must be me or unassigned", ErrInvalidToolArguments)
	}
	if err := stories.ValidateStoryTimeContract(args.EstimatedDurationMinutes, args.MinimumFocusBlockMinutes); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidToolArguments, err)
	}
	if err := stories.ValidateStoryAutoSchedulingContract(args.AutoSchedulingEnabled, false, stories.AutoSchedulingStatusOff); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidToolArguments, err)
	}

	team := joinedByID[teamID]
	confirmationID, err := uuid.NewRandomFromReader(m.random)
	if err != nil {
		return nil, fmt.Errorf("generate story mutation confirmation ID: %w", err)
	}
	now := m.now().UTC()
	claims := storyMutationClaims{
		Version:                  storyMutationTokenVersion,
		ConfirmationID:           confirmationID,
		Operation:                StoryMutationCreate,
		WorkspaceID:              scope.WorkspaceID,
		UserID:                   scope.UserID,
		TeamID:                   teamID,
		Title:                    &title,
		Priority:                 &priority,
		AssigneeAction:           args.Assignee,
		EstimatedDurationMinutes: args.EstimatedDurationMinutes,
		MinimumFocusBlockMinutes: args.MinimumFocusBlockMinutes,
		AutoSchedulingEnabled:    boolPointer(args.AutoSchedulingEnabled),
		ExpiresAt:                now.Add(storyMutationConfirmationTTL),
	}
	return m.marshalProposal(ctx, claims, StoryMutationPreview{
		TeamID:                   team.ID,
		TeamName:                 team.Name,
		TeamCode:                 strings.ToUpper(team.Code),
		Title:                    title,
		Priority:                 &priority,
		AssigneeAction:           args.Assignee,
		EstimatedDurationMinutes: args.EstimatedDurationMinutes,
		MinimumFocusBlockMinutes: args.MinimumFocusBlockMinutes,
		AutoSchedulingEnabled:    boolPointer(args.AutoSchedulingEnabled),
	}, fmt.Sprintf("Create %q in %s (%s)?", title, team.Name, strings.ToUpper(team.Code)))
}

func (m *storyMutationExecutor) proposeCreateBatch(
	ctx context.Context,
	executor *FortyOneToolExecutor,
	scope ToolScope,
	raw json.RawMessage,
) (json.RawMessage, error) {
	var args struct {
		TeamID  string `json:"team_id"`
		Stories []struct {
			Title                    string  `json:"title"`
			Description              *string `json:"description"`
			Priority                 *string `json:"priority"`
			AssigneeID               *string `json:"assignee_id"`
			EstimatedDurationMinutes *int    `json:"estimated_duration_minutes"`
			MinimumFocusBlockMinutes *int    `json:"minimum_focus_block_minutes"`
			AutoSchedulingEnabled    bool    `json:"auto_scheduling_enabled"`
		} `json:"stories"`
	}
	if err := decodeToolArguments(raw, &args, "team_id", "stories"); err != nil {
		return nil, err
	}
	if len(args.Stories) == 0 || len(args.Stories) > maximumBatchStoryCount {
		return nil, fmt.Errorf("%w: stories must contain 1-%d items", ErrInvalidToolArguments, maximumBatchStoryCount)
	}

	_, joinedByID, err := executor.joinedTeams(ctx, scope)
	if err != nil {
		return nil, err
	}
	teamID, err := parseAccessibleTeamID(args.TeamID, joinedByID)
	if err != nil {
		return nil, err
	}
	team := joinedByID[teamID]

	proposal := batchStoryMutationProposal{
		Version:   batchStoryProposalVersion,
		SourceURL: scope.SourceURL,
		Items:     make([]batchStoryMutationItem, 0, len(args.Stories)),
	}
	previews := make([]StoryMutationPreview, 0, len(args.Stories))
	for index, item := range args.Stories {
		title, err := normalizedStoryTitle(item.Title)
		if err != nil {
			return nil, fmt.Errorf("%w: stories[%d].title: %v", ErrInvalidToolArguments, index, err)
		}
		description, err := normalizedBatchStoryDescription(item.Description)
		if err != nil {
			return nil, fmt.Errorf("%w: stories[%d].description: %v", ErrInvalidToolArguments, index, err)
		}
		priority, err := normalizedStoryPriority(item.Priority, "No Priority")
		if err != nil {
			return nil, fmt.Errorf("%w: stories[%d].priority: %v", ErrInvalidToolArguments, index, err)
		}
		if err := stories.ValidateStoryTimeContract(item.EstimatedDurationMinutes, item.MinimumFocusBlockMinutes); err != nil {
			return nil, fmt.Errorf("%w: stories[%d]: %v", ErrInvalidToolArguments, index, err)
		}
		if err := stories.ValidateStoryAutoSchedulingContract(item.AutoSchedulingEnabled, false, stories.AutoSchedulingStatusOff); err != nil {
			return nil, fmt.Errorf("%w: stories[%d]: %v", ErrInvalidToolArguments, index, err)
		}

		var assigneeID *uuid.UUID
		assigneeName := "Unassigned"
		if item.AssigneeID != nil {
			if executor.users == nil {
				return nil, fmt.Errorf("%w: named assignees are unavailable", ErrInvalidToolArguments)
			}
			parsed, err := parseRequiredUUID(*item.AssigneeID, "assignee_id")
			if err != nil {
				return nil, fmt.Errorf("%w: stories[%d].assignee_id: %v", ErrInvalidToolArguments, index, err)
			}
			member, err := executor.activeTeamMemberByID(ctx, scope.WorkspaceID, teamID, parsed)
			if err != nil {
				return nil, err
			}
			if member == nil {
				return nil, fmt.Errorf("%w: stories[%d].assignee_id must identify an active member of %s", ErrInvalidToolArguments, index, team.Name)
			}
			assigneeID = &parsed
			assigneeName = memberDisplayName(*member)
			if assigneeName == "" {
				assigneeName = strings.TrimSpace(member.Username)
			}
		}

		combinedDescription := batchStoryDescription(description, proposal.SourceURL)
		proposal.Items = append(proposal.Items, batchStoryMutationItem{
			Title:                    title,
			Description:              description,
			Priority:                 priority,
			AssigneeID:               assigneeID,
			EstimatedDurationMinutes: item.EstimatedDurationMinutes,
			MinimumFocusBlockMinutes: item.MinimumFocusBlockMinutes,
			AutoSchedulingEnabled:    item.AutoSchedulingEnabled,
		})
		assigneeAction := assigneeActionUnassigned
		if assigneeID != nil {
			assigneeAction = assigneeActionNamed
		}
		previews = append(previews, StoryMutationPreview{
			TeamID:                   team.ID,
			TeamName:                 team.Name,
			TeamCode:                 strings.ToUpper(team.Code),
			Title:                    title,
			Description:              combinedDescription,
			SourceURL:                proposal.SourceURL,
			Priority:                 &priority,
			AssigneeID:               assigneeID,
			AssigneeName:             assigneeName,
			AssigneeAction:           assigneeAction,
			EstimatedDurationMinutes: item.EstimatedDurationMinutes,
			MinimumFocusBlockMinutes: item.MinimumFocusBlockMinutes,
			AutoSchedulingEnabled:    boolPointer(item.AutoSchedulingEnabled),
		})
	}

	confirmationID, err := uuid.NewRandomFromReader(m.random)
	if err != nil {
		return nil, fmt.Errorf("generate batch story confirmation ID: %w", err)
	}
	token, err := m.newBatchToken(confirmationID)
	if err != nil {
		return nil, err
	}
	persistedProposal, err := json.Marshal(proposal)
	if err != nil {
		return nil, fmt.Errorf("encode batch story proposal: %w", err)
	}
	now := m.now().UTC()
	expiresAt := now.Add(storyMutationConfirmationTTL)
	if err := m.store.RegisterStoryMutationConfirmation(ctx, StoryMutationConfirmationStateInput{
		ConfirmationID: confirmationID,
		WorkspaceID:    scope.WorkspaceID,
		UserID:         scope.UserID,
		TeamID:         team.ID,
		Operation:      StoryMutationCreateBatch,
		TokenHash:      storyMutationTokenHash(token),
		Proposal:       persistedProposal,
		ExpiresAt:      expiresAt,
	}); err != nil {
		return nil, fmt.Errorf("register batch story confirmation: %w", err)
	}

	return marshalToolResult(storyMutationConfirmationToolResult{
		Kind: storyMutationConfirmationKind,
		Confirmation: StoryMutationConfirmation{
			Operation: StoryMutationCreateBatch,
			Token:     token,
			ExpiresAt: expiresAt,
			Prompt:    batchStoryConfirmationPrompt(team, previews),
			Stories:   previews,
		},
	})
}

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
	joinedByID map[uuid.UUID]teams.CoreTeam,
	storyIDRaw, storyReferenceRaw *string,
) (stories.CoreSingleStory, error) {
	if storyIDRaw != nil {
		storyID, err := parseRequiredUUID(*storyIDRaw, "story_id")
		if err != nil {
			return stories.CoreSingleStory{}, err
		}
		story, err := m.stories.Get(ctx, storyID, scope.WorkspaceID)
		if err != nil {
			return stories.CoreSingleStory{}, fmt.Errorf("load story for update proposal: %w", err)
		}
		return story, nil
	}

	reference, teamCode, err := normalizedStoryReference(*storyReferenceRaw)
	if err != nil {
		return stories.CoreSingleStory{}, err
	}
	var expectedTeamID uuid.UUID
	for teamID, team := range joinedByID {
		if strings.EqualFold(team.Code, teamCode) {
			expectedTeamID = teamID
			break
		}
	}
	if expectedTeamID == uuid.Nil {
		return stories.CoreSingleStory{}, fmt.Errorf("%w: team code %s", ErrTeamNotAccessible, teamCode)
	}
	story, err := m.stories.QueryByRef(ctx, scope.WorkspaceID, reference)
	if err != nil {
		return stories.CoreSingleStory{}, fmt.Errorf("load story %s for update proposal: %w", reference, err)
	}
	if story.Team != expectedTeamID {
		return stories.CoreSingleStory{}, fmt.Errorf("%w: story team does not match reference", ErrInvalidToolArguments)
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
	team teams.CoreTeam,
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
	story, err := m.stories.CreateExternalUserAction(ctx, scope.UserID, stories.CoreNewStory{
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
	team teams.CoreTeam,
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
		if err := stories.ValidateStoryTimeContract(item.EstimatedDurationMinutes, item.MinimumFocusBlockMinutes); err != nil {
			return result, fmt.Errorf("%w: invalid time contract in batch item %d: %v", ErrInvalidConfirmation, index, err)
		}
		if err := stories.ValidateStoryAutoSchedulingContract(item.AutoSchedulingEnabled, false, stories.AutoSchedulingStatusOff); err != nil {
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
		story, err := m.stories.CreateExternalUserAction(ctx, scope.UserID, stories.CoreNewStory{
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
	team teams.CoreTeam,
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
			if errors.Is(err, stories.ErrStoryChanged) {
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

func (m *storyMutationExecutor) confirmComment(ctx context.Context, scope ToolScope, team teams.CoreTeam, claims storyMutationClaims) (StoryMutationResult, error) {
	if claims.StoryID == nil || claims.Comment == nil || strings.TrimSpace(*claims.Comment) == "" {
		return StoryMutationResult{}, fmt.Errorf("%w: malformed comment proposal", ErrInvalidConfirmation)
	}
	comment, err := m.stories.CreateCommentExternal(ctx, scope.UserID, scope.WorkspaceID, stories.CoreNewComment{StoryID: *claims.StoryID, UserID: scope.UserID, Comment: *claims.Comment, Mentions: claims.MentionIDs})
	if err != nil {
		return StoryMutationResult{}, fmt.Errorf("create confirmed story comment: %w", err)
	}
	commentID := comment.ID
	return StoryMutationResult{Status: storyMutationStatusApplied, Operation: StoryMutationComment, StoryID: *claims.StoryID, CommentID: &commentID, TeamID: team.ID}, nil
}

func (m *storyMutationExecutor) confirmRelationship(ctx context.Context, scope ToolScope, team teams.CoreTeam, claims storyMutationClaims) (StoryMutationResult, error) {
	if claims.StoryID == nil || claims.RelationStoryID == nil || claims.RelationType == "" {
		return StoryMutationResult{}, fmt.Errorf("%w: malformed relationship proposal", ErrInvalidConfirmation)
	}
	association, err := m.stories.AddAssociation(ctx, *claims.StoryID, *claims.RelationStoryID, claims.RelationType, scope.WorkspaceID)
	if err != nil {
		return StoryMutationResult{}, fmt.Errorf("create confirmed story relationship: %w", err)
	}
	return StoryMutationResult{Status: storyMutationStatusApplied, Operation: StoryMutationRelation, StoryID: *claims.StoryID, AssociationID: &association.ID, TeamID: team.ID}, nil
}

func (m *storyMutationExecutor) signClaims(claims storyMutationClaims) (string, error) {
	var encoded bytes.Buffer
	encoder := json.NewEncoder(&encoded)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(claims); err != nil {
		return "", fmt.Errorf("encode story mutation confirmation: %w", err)
	}
	payload := bytes.TrimSuffix(encoded.Bytes(), []byte("\n"))
	signature := hmac.New(sha256.New, m.key)
	_, _ = signature.Write(payload)
	token := base64.RawURLEncoding.EncodeToString(payload) + "." + base64.RawURLEncoding.EncodeToString(signature.Sum(nil))
	if len(token) > maximumStoryMutationTokenSize {
		return "", fmt.Errorf("story mutation confirmation exceeds %d characters", maximumStoryMutationTokenSize)
	}
	return token, nil
}

func (m *storyMutationExecutor) newBatchToken(confirmationID uuid.UUID) (string, error) {
	payload := make([]byte, batchStoryTokenBytes)
	copy(payload, confirmationID[:])
	if _, err := io.ReadFull(m.random, payload[len(confirmationID):]); err != nil {
		return "", fmt.Errorf("generate opaque batch confirmation token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(payload), nil
}

func batchConfirmationID(token string) (uuid.UUID, bool, error) {
	token = strings.TrimSpace(token)
	if strings.Contains(token, ".") {
		return uuid.Nil, false, nil
	}
	if token == "" || len(token) > maximumStoryMutationTokenSize {
		return uuid.Nil, true, fmt.Errorf("%w: token is missing or too large", ErrInvalidConfirmation)
	}
	payload, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil || len(payload) != batchStoryTokenBytes || base64.RawURLEncoding.EncodeToString(payload) != token {
		return uuid.Nil, true, fmt.Errorf("%w: opaque token format", ErrInvalidConfirmation)
	}
	confirmationID, err := uuid.FromBytes(payload[:16])
	if err != nil || confirmationID == uuid.Nil {
		return uuid.Nil, true, fmt.Errorf("%w: opaque token identifier", ErrInvalidConfirmation)
	}
	return confirmationID, true, nil
}

func (m *storyMutationExecutor) verifyClaims(token string) (storyMutationClaims, error) {
	token = strings.TrimSpace(token)
	if token == "" || len(token) > maximumStoryMutationTokenSize {
		return storyMutationClaims{}, fmt.Errorf("%w: token is missing or too large", ErrInvalidConfirmation)
	}
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return storyMutationClaims{}, fmt.Errorf("%w: token format", ErrInvalidConfirmation)
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil || base64.RawURLEncoding.EncodeToString(payload) != parts[0] {
		return storyMutationClaims{}, fmt.Errorf("%w: token payload", ErrInvalidConfirmation)
	}
	providedSignature, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil || base64.RawURLEncoding.EncodeToString(providedSignature) != parts[1] {
		return storyMutationClaims{}, fmt.Errorf("%w: token signature", ErrInvalidConfirmation)
	}
	expectedSignature := hmac.New(sha256.New, m.key)
	_, _ = expectedSignature.Write(payload)
	if !hmac.Equal(providedSignature, expectedSignature.Sum(nil)) {
		return storyMutationClaims{}, fmt.Errorf("%w: token signature", ErrInvalidConfirmation)
	}

	var claims storyMutationClaims
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&claims); err != nil {
		return storyMutationClaims{}, fmt.Errorf("%w: token claims", ErrInvalidConfirmation)
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return storyMutationClaims{}, fmt.Errorf("%w: trailing token claims", ErrInvalidConfirmation)
	}
	if claims.Version != storyMutationTokenVersion || claims.ConfirmationID == uuid.Nil || claims.WorkspaceID == uuid.Nil || claims.UserID == uuid.Nil || claims.TeamID == uuid.Nil || claims.ExpiresAt.IsZero() {
		return storyMutationClaims{}, fmt.Errorf("%w: incomplete token claims", ErrInvalidConfirmation)
	}
	normalizeLegacyUpdateTimeActions(&claims)
	if err := validateStoryMutationClaims(claims); err != nil {
		return storyMutationClaims{}, err
	}
	return claims, nil
}

func storyMutationConfirmationBinding(claims storyMutationClaims, token string) StoryMutationConfirmationBinding {
	return StoryMutationConfirmationBinding{
		ConfirmationID: claims.ConfirmationID,
		WorkspaceID:    claims.WorkspaceID,
		UserID:         claims.UserID,
		TokenHash:      storyMutationTokenHash(token),
	}
}

func storyMutationTokenHash(token string) []byte {
	digest := sha256.Sum256([]byte(strings.TrimSpace(token)))
	return append([]byte(nil), digest[:]...)
}

func decodeBatchStoryProposal(raw json.RawMessage) (batchStoryMutationProposal, error) {
	if len(raw) == 0 {
		return batchStoryMutationProposal{}, fmt.Errorf("%w: batch proposal is missing", ErrInvalidConfirmation)
	}
	var proposal batchStoryMutationProposal
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&proposal); err != nil {
		return batchStoryMutationProposal{}, fmt.Errorf("%w: decode batch proposal: %v", ErrInvalidConfirmation, err)
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return batchStoryMutationProposal{}, fmt.Errorf("%w: trailing batch proposal data", ErrInvalidConfirmation)
	}
	if proposal.Version != batchStoryProposalVersion || len(proposal.Items) == 0 || len(proposal.Items) > maximumBatchStoryCount {
		return batchStoryMutationProposal{}, fmt.Errorf("%w: unsupported or empty batch proposal", ErrInvalidConfirmation)
	}
	sourceURL, err := normalizedSourceURL(proposal.SourceURL)
	if err != nil || sourceURL != proposal.SourceURL {
		return batchStoryMutationProposal{}, fmt.Errorf("%w: invalid batch source URL", ErrInvalidConfirmation)
	}
	for index, item := range proposal.Items {
		title, titleErr := normalizedStoryTitle(item.Title)
		description, descriptionErr := normalizedBatchStoryDescriptionValue(item.Description)
		priority, priorityErr := normalizedStoryPriority(&item.Priority, "")
		if titleErr != nil || title != item.Title || descriptionErr != nil || description != item.Description || priorityErr != nil || priority != item.Priority {
			return batchStoryMutationProposal{}, fmt.Errorf("%w: invalid batch item %d", ErrInvalidConfirmation, index)
		}
		if item.AssigneeID != nil && *item.AssigneeID == uuid.Nil {
			return batchStoryMutationProposal{}, fmt.Errorf("%w: invalid batch item %d assignee", ErrInvalidConfirmation, index)
		}
	}
	return proposal, nil
}

func normalizedBatchStoryDescription(raw *string) (string, error) {
	if raw == nil {
		return "", nil
	}
	return normalizedBatchStoryDescriptionValue(*raw)
}

func normalizedBatchStoryDescriptionValue(raw string) (string, error) {
	description := strings.TrimSpace(raw)
	if len([]rune(description)) > maximumBatchDescriptionRunes {
		return "", fmt.Errorf("description must not exceed %d characters", maximumBatchDescriptionRunes)
	}
	return description, nil
}

func batchStoryDescription(description, sourceURL string) string {
	description = strings.TrimSpace(description)
	sourceURL = strings.TrimSpace(sourceURL)
	if sourceURL == "" {
		return description
	}
	if description == "" {
		return "Source: " + sourceURL
	}
	return description + "\n\nSource: " + sourceURL
}

func batchStoryConfirmationPrompt(team teams.CoreTeam, previews []StoryMutationPreview) string {
	var prompt strings.Builder
	fmt.Fprintf(&prompt, "Create %d stories in %s (%s)?", len(previews), confirmationPromptText(team.Name), confirmationPromptText(strings.ToUpper(team.Code)))
	for index, preview := range previews {
		assignee := preview.AssigneeName
		if strings.TrimSpace(assignee) == "" {
			assignee = "Unassigned"
		}
		priority := "No Priority"
		if preview.Priority != nil {
			priority = *preview.Priority
		}
		fmt.Fprintf(&prompt, "\n%d. %s — %s, %s", index+1, confirmationPromptText(preview.Title), confirmationPromptText(assignee), confirmationPromptText(priority))
	}
	if len(previews) > 0 && previews[0].SourceURL != "" {
		prompt.WriteString("\n\nThe supporting descriptions and a link to this Slack thread will be attached to the stories.")
	} else {
		prompt.WriteString("\n\nThe supporting descriptions shown in the draft will be attached to the stories.")
	}
	return prompt.String()
}

func confirmationPromptText(value string) string {
	value = strings.Join(strings.Fields(value), " ")
	return strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;").Replace(value)
}

func cloneUUIDPointer(value *uuid.UUID) *uuid.UUID {
	if value == nil {
		return nil
	}
	copy := *value
	return &copy
}

func cloneIntPointer(value *int) *int {
	if value == nil {
		return nil
	}
	copy := *value
	return &copy
}

func boolPointer(value bool) *bool {
	return &value
}

func cloneBoolPointer(value *bool) *bool {
	if value == nil {
		return nil
	}
	copy := *value
	return &copy
}

func optionalBoolValue(value *bool) bool {
	return value != nil && *value
}

func normalizeLegacyUpdateTimeActions(claims *storyMutationClaims) {
	if claims == nil || claims.Operation != StoryMutationUpdate {
		return
	}
	if claims.EstimatedDurationAction == "" {
		claims.EstimatedDurationAction = storyTimeActionUnchanged
		if claims.EstimatedDurationMinutes != nil {
			claims.EstimatedDurationAction = storyTimeActionSet
		}
	}
	if claims.MinimumFocusBlockAction == "" {
		claims.MinimumFocusBlockAction = storyTimeActionUnchanged
		if claims.MinimumFocusBlockMinutes != nil {
			claims.MinimumFocusBlockAction = storyTimeActionSet
		}
	}
}

func validateStoryTimeMutationClaims(mutation storyTimeMutation) error {
	estimatedDurationMinutes, err := resolvedStoryTimeValue(
		nil,
		mutation.estimatedDurationAction,
		mutation.estimatedDurationMinutes,
		"estimated_duration",
	)
	if err != nil {
		return err
	}
	minimumFocusBlockMinutes, err := resolvedStoryTimeValue(
		nil,
		mutation.minimumFocusBlockAction,
		mutation.minimumFocusBlockMinutes,
		"minimum_focus_block",
	)
	if err != nil {
		return err
	}

	switch {
	case mutation.estimatedDurationAction == storyTimeActionClear && mutation.minimumFocusBlockAction == storyTimeActionSet:
		return stories.ErrFocusBlockRequiresDuration
	case mutation.estimatedDurationAction == storyTimeActionSet && mutation.minimumFocusBlockAction == storyTimeActionSet:
		return stories.ValidateStoryTimeContract(estimatedDurationMinutes, minimumFocusBlockMinutes)
	default:
		return nil
	}
}

func validateStoryMutationClaims(claims storyMutationClaims) error {
	if claims.Title != nil {
		title, err := normalizedStoryTitle(*claims.Title)
		if err != nil || title != *claims.Title {
			return fmt.Errorf("%w: invalid title claim", ErrInvalidConfirmation)
		}
	}
	if claims.Priority != nil {
		priority, err := normalizedStoryPriority(claims.Priority, "")
		if err != nil || priority != *claims.Priority {
			return fmt.Errorf("%w: invalid priority claim", ErrInvalidConfirmation)
		}
	}

	switch claims.Operation {
	case StoryMutationCreate:
		if claims.Title == nil || claims.Priority == nil || claims.StoryID != nil || claims.ExpectedUpdatedAt != nil {
			return fmt.Errorf("%w: malformed create claims", ErrInvalidConfirmation)
		}
		if claims.AutoSchedulingLocked != nil {
			return fmt.Errorf("%w: create claims cannot set auto-scheduling lock state", ErrInvalidConfirmation)
		}
		if claims.AssigneeAction != assigneeActionMe && claims.AssigneeAction != assigneeActionUnassigned {
			return fmt.Errorf("%w: invalid create assignee claim", ErrInvalidConfirmation)
		}
		if err := stories.ValidateStoryTimeContract(claims.EstimatedDurationMinutes, claims.MinimumFocusBlockMinutes); err != nil {
			return fmt.Errorf("%w: invalid create time claim: %v", ErrInvalidConfirmation, err)
		}
		if err := stories.ValidateStoryAutoSchedulingContract(optionalBoolValue(claims.AutoSchedulingEnabled), false, stories.AutoSchedulingStatusOff); err != nil {
			return fmt.Errorf("%w: invalid create auto-scheduling claim: %v", ErrInvalidConfirmation, err)
		}
	case StoryMutationUpdate:
		if claims.StoryID == nil || *claims.StoryID == uuid.Nil || claims.ExpectedUpdatedAt == nil || claims.ExpectedUpdatedAt.IsZero() {
			return fmt.Errorf("%w: malformed update claims", ErrInvalidConfirmation)
		}
		if claims.AssigneeAction != assigneeActionUnchanged && claims.AssigneeAction != assigneeActionMe && claims.AssigneeAction != assigneeActionUnassigned {
			return fmt.Errorf("%w: invalid update assignee claim", ErrInvalidConfirmation)
		}
		if err := validateStoryTimeMutationClaims(storyTimeMutation{
			estimatedDurationAction:  claims.EstimatedDurationAction,
			estimatedDurationMinutes: claims.EstimatedDurationMinutes,
			minimumFocusBlockAction:  claims.MinimumFocusBlockAction,
			minimumFocusBlockMinutes: claims.MinimumFocusBlockMinutes,
		}); err != nil {
			return fmt.Errorf("%w: invalid update time claim: %v", ErrInvalidConfirmation, err)
		}
		if claims.Title == nil && claims.Priority == nil && claims.AssigneeAction == assigneeActionUnchanged && claims.StatusID == nil && claims.SprintID == nil && claims.ObjectiveID == nil && claims.KeyResultID == nil && claims.StartDate == nil && claims.EndDate == nil && claims.LabelIDs == nil && claims.EstimatedDurationAction == storyTimeActionUnchanged && claims.MinimumFocusBlockAction == storyTimeActionUnchanged && claims.AutoSchedulingEnabled == nil && claims.AutoSchedulingLocked == nil {
			return fmt.Errorf("%w: empty update claims", ErrInvalidConfirmation)
		}
	case StoryMutationComment:
		if claims.StoryID == nil || claims.Comment == nil || strings.TrimSpace(*claims.Comment) == "" {
			return fmt.Errorf("%w: malformed comment claims", ErrInvalidConfirmation)
		}
	case StoryMutationRelation:
		if claims.StoryID == nil || claims.RelationStoryID == nil || claims.RelationType == "" || *claims.StoryID == *claims.RelationStoryID {
			return fmt.Errorf("%w: malformed relationship claims", ErrInvalidConfirmation)
		}
		if claims.RelationType != "blocking" && claims.RelationType != "related" && claims.RelationType != "duplicate" {
			return fmt.Errorf("%w: invalid relationship type", ErrInvalidConfirmation)
		}
	default:
		return fmt.Errorf("%w: unsupported operation", ErrInvalidConfirmation)
	}
	return nil
}

func sameUUIDSet(left, right []uuid.UUID) bool {
	if len(left) != len(right) {
		return false
	}
	seen := make(map[uuid.UUID]struct{}, len(left))
	for _, value := range left {
		seen[value] = struct{}{}
	}
	for _, value := range right {
		if _, ok := seen[value]; !ok {
			return false
		}
	}
	return true
}

func mutationConfirmationFromToolResult(raw json.RawMessage) (*StoryMutationConfirmation, bool, error) {
	var header struct {
		Kind string `json:"kind"`
	}
	if err := json.Unmarshal(raw, &header); err != nil || header.Kind != storyMutationConfirmationKind {
		return nil, false, nil
	}
	var result storyMutationConfirmationToolResult
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, true, fmt.Errorf("%w: decode mutation confirmation: %v", ErrMalformedResponse, err)
	}
	confirmation := result.Confirmation
	if confirmation.Token == "" || confirmation.Prompt == "" || confirmation.ExpiresAt.IsZero() {
		return nil, true, fmt.Errorf("%w: incomplete mutation confirmation", ErrMalformedResponse)
	}
	if confirmation.Operation == StoryMutationCreateBatch {
		if len(confirmation.Stories) == 0 || len(confirmation.Stories) > maximumBatchStoryCount {
			return nil, true, fmt.Errorf("%w: incomplete batch mutation confirmation", ErrMalformedResponse)
		}
		teamID := confirmation.Stories[0].TeamID
		if teamID == uuid.Nil {
			return nil, true, fmt.Errorf("%w: incomplete batch mutation confirmation", ErrMalformedResponse)
		}
		for _, story := range confirmation.Stories[1:] {
			if story.TeamID != teamID {
				return nil, true, fmt.Errorf("%w: batch mutation spans multiple teams", ErrMalformedResponse)
			}
		}
	} else if confirmation.Story.TeamID == uuid.Nil {
		return nil, true, fmt.Errorf("%w: incomplete mutation confirmation", ErrMalformedResponse)
	}
	return &confirmation, true, nil
}

func validateToolScope(scope *ToolScope) error {
	if scope.WorkspaceID == uuid.Nil || scope.UserID == uuid.Nil {
		return fmt.Errorf("%w: workspace and user are required", ErrInvalidRequest)
	}
	allowedTeamIDs, err := normalizedAllowedTeamIDs(scope.AllowedTeamIDs)
	if err != nil {
		return err
	}
	scope.AllowedTeamIDs = allowedTeamIDs
	scope.SourceURL, err = normalizedSourceURL(scope.SourceURL)
	if err != nil {
		return err
	}
	return nil
}

func parseAccessibleTeamID(raw string, joined map[uuid.UUID]teams.CoreTeam) (uuid.UUID, error) {
	teamID, err := parseRequiredUUID(raw, "team_id")
	if err != nil {
		return uuid.Nil, err
	}
	if _, ok := joined[teamID]; !ok {
		return uuid.Nil, fmt.Errorf("%w: %s", ErrTeamNotAccessible, teamID)
	}
	return teamID, nil
}

func parseRequiredUUID(raw, field string) (uuid.UUID, error) {
	value, err := uuid.Parse(strings.TrimSpace(raw))
	if err != nil || value == uuid.Nil {
		return uuid.Nil, fmt.Errorf("%w: %s must be a UUID", ErrInvalidToolArguments, field)
	}
	return value, nil
}

func normalizedStoryTitle(raw string) (string, error) {
	title := strings.TrimSpace(raw)
	if title == "" || len([]rune(title)) > maximumStoryTitleRunes {
		return "", fmt.Errorf("%w: title must contain 1-%d characters", ErrInvalidToolArguments, maximumStoryTitleRunes)
	}
	return title, nil
}

func normalizedStoryReference(raw string) (string, string, error) {
	reference := strings.ToUpper(strings.TrimSpace(raw))
	separator := strings.LastIndex(reference, "-")
	if separator < 1 || separator == len(reference)-1 {
		return "", "", fmt.Errorf("%w: story_reference must use TEAM-123 format", ErrInvalidToolArguments)
	}
	teamCode := strings.TrimSpace(reference[:separator])
	sequenceRaw := strings.TrimSpace(reference[separator+1:])
	sequenceID, err := strconv.Atoi(sequenceRaw)
	if teamCode == "" || err != nil || sequenceID < 1 {
		return "", "", fmt.Errorf("%w: story_reference must use TEAM-123 format", ErrInvalidToolArguments)
	}
	return fmt.Sprintf("%s-%d", teamCode, sequenceID), teamCode, nil
}

func normalizedStoryPriority(raw *string, fallback string) (string, error) {
	if raw == nil {
		return fallback, nil
	}
	priority := strings.TrimSpace(*raw)
	if _, ok := storyPriorities[priority]; !ok {
		return "", fmt.Errorf("%w: unsupported priority %q", ErrInvalidToolArguments, priority)
	}
	return priority, nil
}

func proposedChangedFields(
	story stories.CoreSingleStory,
	userID uuid.UUID,
	title, priority *string,
	assigneeAction string,
	statusID, sprintID, objectiveID, keyResultID *uuid.UUID,
	startDate, endDate *time.Time,
	timeMutation storyTimeMutation,
	autoSchedulingMutation storyAutoSchedulingMutation,
) ([]string, error) {
	updates, err := desiredStoryUpdates(
		story,
		userID,
		title,
		priority,
		assigneeAction,
		statusID,
		sprintID,
		objectiveID,
		keyResultID,
		startDate,
		endDate,
		timeMutation,
		autoSchedulingMutation,
	)
	if err != nil {
		return nil, err
	}
	fields := make([]string, 0, len(updates))
	for _, field := range []string{"title", "priority", "assignee_id", "status_id", "sprint_id", "objective_id", "key_result_id", "start_date", "end_date", "estimated_duration_minutes", "minimum_focus_block_minutes", "auto_scheduling_enabled", "auto_scheduling_locked"} {
		if _, changed := updates[field]; changed {
			fields = append(fields, field)
		}
	}
	return fields, nil
}

func desiredStoryUpdates(
	story stories.CoreSingleStory,
	userID uuid.UUID,
	title, priority *string,
	assigneeAction string,
	statusID, sprintID, objectiveID, keyResultID *uuid.UUID,
	startDate, endDate *time.Time,
	timeMutation storyTimeMutation,
	autoSchedulingMutation storyAutoSchedulingMutation,
) (map[string]any, error) {
	updates := make(map[string]any, 3)
	if title != nil && story.Title != *title {
		updates["title"] = *title
	}
	if priority != nil && story.Priority != *priority {
		updates["priority"] = *priority
	}
	switch assigneeAction {
	case assigneeActionMe:
		if story.Assignee == nil || *story.Assignee != userID {
			updates["assignee_id"] = userID
		}
	case assigneeActionUnassigned:
		if story.Assignee != nil {
			updates["assignee_id"] = nil
		}
	}
	for field, value := range map[string]any{"status_id": statusID, "sprint_id": sprintID, "objective_id": objectiveID, "key_result_id": keyResultID, "start_date": startDate, "end_date": endDate} {
		if value != nil && !storyFieldMatches(story, field, value) {
			updates[field] = value
		}
	}
	timeUpdates, err := desiredStoryTimeUpdates(story, timeMutation)
	if err != nil {
		return nil, err
	}
	for field, value := range timeUpdates {
		updates[field] = value
	}
	autoSchedulingUpdates, err := desiredStoryAutoSchedulingUpdates(story, autoSchedulingMutation)
	if err != nil {
		return nil, err
	}
	for field, value := range autoSchedulingUpdates {
		updates[field] = value
	}
	return updates, nil
}

func desiredStoryAutoSchedulingUpdates(story stories.CoreSingleStory, mutation storyAutoSchedulingMutation) (map[string]any, error) {
	enabled := story.AutoSchedulingEnabled
	if mutation.enabled != nil {
		enabled = *mutation.enabled
	}
	locked := story.AutoSchedulingLocked
	disabling := mutation.enabled != nil && !enabled
	if disabling {
		locked = false
	} else if mutation.locked != nil {
		locked = *mutation.locked
	}
	status := story.AutoSchedulingStatus
	if status == "" {
		status = stories.AutoSchedulingStatusOff
	}
	if mutation.locked != nil && *mutation.locked && !story.AutoSchedulingLocked &&
		(!story.AutoSchedulingEnabled || (status != stories.AutoSchedulingStatusScheduled && status != stories.AutoSchedulingStatusAtRisk)) {
		return nil, fmt.Errorf("%w: %w", ErrInvalidToolArguments, stories.ErrAutoSchedulingLockEmpty)
	}
	if err := stories.ValidateStoryAutoSchedulingContract(enabled, locked, status); err != nil {
		return nil, err
	}

	updates := make(map[string]any, 2)
	if mutation.enabled != nil && enabled != story.AutoSchedulingEnabled {
		updates["auto_scheduling_enabled"] = enabled
	}
	if (mutation.locked != nil || disabling) && locked != story.AutoSchedulingLocked {
		updates["auto_scheduling_locked"] = locked
	}
	return updates, nil
}

func desiredStoryTimeUpdates(story stories.CoreSingleStory, mutation storyTimeMutation) (map[string]any, error) {
	estimatedDurationMinutes, err := resolvedStoryTimeValue(
		story.EstimatedDurationMinutes,
		mutation.estimatedDurationAction,
		mutation.estimatedDurationMinutes,
		"estimated_duration",
	)
	if err != nil {
		return nil, err
	}
	minimumFocusBlockMinutes, err := resolvedStoryTimeValue(
		story.MinimumFocusBlockMinutes,
		mutation.minimumFocusBlockAction,
		mutation.minimumFocusBlockMinutes,
		"minimum_focus_block",
	)
	if err != nil {
		return nil, err
	}

	// A focus-block constraint cannot survive without a duration. Clearing the
	// duration therefore clears the dependent constraint even when its own
	// action is unchanged.
	if mutation.estimatedDurationAction == storyTimeActionClear && mutation.minimumFocusBlockAction == storyTimeActionUnchanged {
		minimumFocusBlockMinutes = nil
	}
	if err := stories.ValidateStoryTimeContract(estimatedDurationMinutes, minimumFocusBlockMinutes); err != nil {
		return nil, err
	}

	updates := make(map[string]any, 2)
	if !sameOptionalInt(story.EstimatedDurationMinutes, estimatedDurationMinutes) {
		updates["estimated_duration_minutes"] = storyTimeUpdateValue(estimatedDurationMinutes)
	}
	if !sameOptionalInt(story.MinimumFocusBlockMinutes, minimumFocusBlockMinutes) {
		updates["minimum_focus_block_minutes"] = storyTimeUpdateValue(minimumFocusBlockMinutes)
	}
	return updates, nil
}

func storyTimeUpdateValue(minutes *int) any {
	if minutes == nil {
		return nil
	}
	return *minutes
}

func resolvedStoryTimeValue(current *int, action string, minutes *int, field string) (*int, error) {
	switch action {
	case storyTimeActionUnchanged:
		if minutes != nil {
			return nil, fmt.Errorf("%s_minutes must be null when %s_action is unchanged", field, field)
		}
		return cloneIntPointer(current), nil
	case storyTimeActionClear:
		if minutes != nil {
			return nil, fmt.Errorf("%s_minutes must be null when %s_action is clear", field, field)
		}
		return nil, nil
	case storyTimeActionSet:
		if minutes == nil {
			return nil, fmt.Errorf("%s_minutes is required when %s_action is set", field, field)
		}
		if *minutes < 1 || *minutes > stories.MaximumEstimatedDurationMinutes {
			return nil, fmt.Errorf("%s_minutes must be between 1 and %d", field, stories.MaximumEstimatedDurationMinutes)
		}
		return cloneIntPointer(minutes), nil
	default:
		return nil, fmt.Errorf("%s_action must be unchanged, set, or clear", field)
	}
}

func sameOptionalInt(left, right *int) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return *left == *right
}

func storyFieldMatches(story stories.CoreSingleStory, field string, value any) bool {
	var current any
	switch field {
	case "status_id":
		current = story.Status
	case "sprint_id":
		current = story.Sprint
	case "objective_id":
		current = story.Objective
	case "key_result_id":
		current = story.KeyResult
	case "start_date":
		current = story.StartDate
	case "end_date":
		current = story.EndDate
	default:
		return false
	}
	return reflect.DeepEqual(current, value)
}

func parseOptionalUUID(raw *string, field string) (*uuid.UUID, error) {
	if raw == nil || strings.TrimSpace(*raw) == "" {
		return nil, nil
	}
	value, err := uuid.Parse(strings.TrimSpace(*raw))
	if err != nil {
		return nil, fmt.Errorf("%w: %s must be a UUID", ErrInvalidToolArguments, field)
	}
	return &value, nil
}

func parseOptionalDate(raw *string, field, timezone string) (*time.Time, error) {
	if raw == nil || strings.TrimSpace(*raw) == "" {
		return nil, nil
	}
	rawValue := strings.TrimSpace(*raw)
	value, err := time.Parse(time.RFC3339, rawValue)
	if err != nil {
		location, locationErr := time.LoadLocation(strings.TrimSpace(timezone))
		if locationErr != nil {
			location = time.UTC
		}
		value, err = time.ParseInLocation("2006-01-02", rawValue, location)
	}
	if err != nil {
		return nil, fmt.Errorf("%w: %s must be YYYY-MM-DD or RFC3339", ErrInvalidToolArguments, field)
	}
	value = value.UTC()
	return &value, nil
}

func parseUUIDList(raw []string, field string) ([]uuid.UUID, error) {
	if raw == nil {
		return nil, nil
	}
	values := make([]uuid.UUID, 0, len(raw))
	seen := make(map[uuid.UUID]struct{}, len(raw))
	for _, item := range raw {
		value, err := uuid.Parse(strings.TrimSpace(item))
		if err != nil || value == uuid.Nil {
			return nil, fmt.Errorf("%w: %s contains an invalid UUID", ErrInvalidToolArguments, field)
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		values = append(values, value)
	}
	return values, nil
}

func storyMutationResult(status string, operation StoryMutationOperation, story stories.CoreSingleStory, teamCode string) StoryMutationResult {
	return StoryMutationResult{
		Status:                   status,
		Operation:                operation,
		StoryID:                  story.ID,
		Reference:                storyReference(strings.ToUpper(teamCode), story.SequenceID),
		TeamID:                   story.Team,
		Title:                    story.Title,
		Priority:                 story.Priority,
		AssigneeID:               story.Assignee,
		EstimatedDurationMinutes: story.EstimatedDurationMinutes,
		MinimumFocusBlockMinutes: story.MinimumFocusBlockMinutes,
		AutoSchedulingEnabled:    story.AutoSchedulingEnabled,
		AutoSchedulingLocked:     story.AutoSchedulingLocked,
		AutoSchedulingStatus:     story.AutoSchedulingStatus,
		AutoSchedulingReason:     story.AutoSchedulingReason,
		AutoSchedulingUpdatedAt:  story.AutoSchedulingUpdatedAt,
	}
}

func nullableUUID(description string) map[string]any {
	return map[string]any{"type": []string{"string", "null"}, "description": description}
}

func nullableDate(description string) map[string]any {
	return map[string]any{"type": []string{"string", "null"}, "description": description}
}

func nullableMinutes(description string) map[string]any {
	return map[string]any{
		"type":        []string{"integer", "null"},
		"description": description,
		"minimum":     1,
		"maximum":     stories.MaximumEstimatedDurationMinutes,
	}
}

func nullableBoolean(description string) map[string]any {
	return map[string]any{"type": []string{"boolean", "null"}, "description": description}
}

func storyTimeAction(description string) map[string]any {
	return map[string]any{
		"type":        "string",
		"description": description,
		"enum":        []string{storyTimeActionUnchanged, storyTimeActionSet, storyTimeActionClear},
	}
}

func storyMutationToolDefinitions() []ToolDefinition {
	nullablePriority := map[string]any{
		"type":        []string{"string", "null"},
		"description": "A FortyOne priority, or null as described by this tool.",
		"enum":        []any{"No Priority", "Low", "Medium", "High", "Urgent", nil},
	}
	return []ToolDefinition{
		{
			Type:        "function",
			Name:        toolCreateStory,
			Description: "Prepare a story creation proposal only when the user explicitly asks to create one and the team and title are unambiguous. This tool never writes; FortyOne will require explicit user confirmation.",
			Strict:      true,
			Parameters: strictObjectSchema(map[string]any{
				"team_id": map[string]any{
					"type":        "string",
					"description": "An exact team UUID returned by list_teams.",
				},
				"title": map[string]any{
					"type":        "string",
					"description": "The exact story title requested by the user.",
					"minLength":   1,
					"maxLength":   maximumStoryTitleRunes,
				},
				"priority": nullablePriority,
				"assignee": map[string]any{
					"type":        "string",
					"description": "Use me only if the user explicitly asks to assign it to themselves; otherwise use unassigned.",
					"enum":        []string{assigneeActionMe, assigneeActionUnassigned},
				},
				"estimated_duration_minutes":  nullableMinutes("The total time needed in minutes when explicitly known, or null when unspecified."),
				"minimum_focus_block_minutes": nullableMinutes("The smallest useful uninterrupted block in minutes, no greater than estimated_duration_minutes; use null when duration is unspecified."),
				"auto_scheduling_enabled": map[string]any{
					"type":        "boolean",
					"description": "Enable Maya auto-scheduling only when the user explicitly requests it; otherwise false.",
				},
			}, []string{"team_id", "title", "priority", "assignee", "estimated_duration_minutes", "minimum_focus_block_minutes", "auto_scheduling_enabled"}),
		},
		{
			Type:        "function",
			Name:        toolCreateStories,
			Description: "Prepare one confirmation proposal for 1-10 distinct stories in one team when the user explicitly asks to turn a conversation into multiple action items. This tool never writes. Use only explicit assignee UUIDs returned by list_team_members; otherwise pass null. Source attribution is supplied by the server and is not a tool argument.",
			Strict:      true,
			Parameters: strictObjectSchema(map[string]any{
				"team_id": map[string]any{
					"type":        "string",
					"description": "An exact team UUID returned by list_teams.",
				},
				"stories": map[string]any{
					"type":     "array",
					"minItems": 1,
					"maxItems": maximumBatchStoryCount,
					"items": strictObjectSchema(map[string]any{
						"title": map[string]any{
							"type":      "string",
							"minLength": 1,
							"maxLength": maximumStoryTitleRunes,
						},
						"description": map[string]any{
							"type":        []string{"string", "null"},
							"maxLength":   maximumBatchDescriptionRunes,
							"description": "Concise supporting context derived from the visible conversation, or null.",
						},
						"priority": nullablePriority,
						"assignee_id": map[string]any{
							"type":        []string{"string", "null"},
							"description": "An exact active member UUID returned by list_team_members only when assignment is explicit, otherwise null.",
						},
						"estimated_duration_minutes":  nullableMinutes("The total time needed in minutes when explicitly known, or null when unspecified."),
						"minimum_focus_block_minutes": nullableMinutes("The smallest useful uninterrupted block in minutes, no greater than estimated_duration_minutes; use null when duration is unspecified."),
						"auto_scheduling_enabled": map[string]any{
							"type":        "boolean",
							"description": "Enable Maya auto-scheduling only when explicitly requested for this story; otherwise false.",
						},
					}, []string{"title", "description", "priority", "assignee_id", "estimated_duration_minutes", "minimum_focus_block_minutes", "auto_scheduling_enabled"}),
				},
			}, []string{"team_id", "stories"}),
		},
		{
			Type:        "function",
			Name:        toolUpdateStory,
			Description: "Prepare a story update proposal only when the target story and requested fields are unambiguous. Time fields use explicit unchanged, set, or clear actions so null is never ambiguous. This tool never writes; FortyOne will require explicit user confirmation.",
			Strict:      true,
			Parameters: strictObjectSchema(map[string]any{
				"story_id": map[string]any{
					"type":        []string{"string", "null"},
					"description": "An exact story UUID returned by a read tool, or null when story_reference is provided.",
				},
				"story_reference": map[string]any{
					"type":        []string{"string", "null"},
					"description": "An exact human reference such as WEB-123, or null when story_id is provided. Provide exactly one target.",
				},
				"title": map[string]any{
					"type":        []string{"string", "null"},
					"description": "The replacement title, or null to leave it unchanged.",
					"maxLength":   maximumStoryTitleRunes,
				},
				"priority": nullablePriority,
				"assignee": map[string]any{
					"type":        "string",
					"description": "Whether to leave the assignee unchanged, assign the current user, or unassign the story.",
					"enum":        []string{assigneeActionUnchanged, assigneeActionMe, assigneeActionUnassigned},
				},
				"status_id":                   nullableUUID("A visible status UUID, or null to leave unchanged."),
				"sprint_id":                   nullableUUID("A visible sprint UUID, or null to leave unchanged."),
				"objective_id":                nullableUUID("A visible objective UUID, or null to leave unchanged."),
				"key_result_id":               nullableUUID("A visible key result UUID, or null to leave unchanged."),
				"start_date":                  nullableDate("A start date in YYYY-MM-DD or RFC3339, or null to leave unchanged."),
				"end_date":                    nullableDate("An end date in YYYY-MM-DD or RFC3339, or null to leave unchanged."),
				"label_ids":                   map[string]any{"type": []string{"array", "null"}, "description": "The complete replacement list of visible label UUIDs, or null to leave labels unchanged.", "items": map[string]any{"type": "string"}},
				"estimated_duration_action":   storyTimeAction("Use unchanged to preserve the current time needed, set to replace it with estimated_duration_minutes, or clear to remove it. Clearing time needed also clears its minimum focus block."),
				"estimated_duration_minutes":  nullableMinutes("The replacement total time needed when estimated_duration_action is set; otherwise null."),
				"minimum_focus_block_action":  storyTimeAction("Use unchanged to preserve the current minimum focus block, set to replace it with minimum_focus_block_minutes, or clear to remove it."),
				"minimum_focus_block_minutes": nullableMinutes("The replacement minimum uninterrupted block when minimum_focus_block_action is set; otherwise null. It cannot exceed the resulting time needed."),
				"auto_scheduling_enabled":     nullableBoolean("Set true or false only when the user explicitly asks to change auto-scheduling; otherwise null."),
				"auto_scheduling_locked":      nullableBoolean("Set true or false only when the user explicitly asks to lock or unlock Maya's schedule; otherwise null. Locking requires auto-scheduling to remain enabled."),
			}, []string{"story_id", "story_reference", "title", "priority", "assignee", "status_id", "sprint_id", "objective_id", "key_result_id", "start_date", "end_date", "label_ids", "estimated_duration_action", "estimated_duration_minutes", "minimum_focus_block_action", "minimum_focus_block_minutes", "auto_scheduling_enabled", "auto_scheduling_locked"}),
		},
		{
			Type: "function", Name: toolAddComment,
			Description: "Prepare a comment proposal for an accessible story. The comment and optional mentions are written only after explicit user confirmation.", Strict: true,
			Parameters: strictObjectSchema(map[string]any{
				"story_id":        map[string]any{"type": []string{"string", "null"}},
				"story_reference": map[string]any{"type": []string{"string", "null"}},
				"body":            map[string]any{"type": "string", "minLength": 1, "maxLength": maxStoryDescriptionRunes},
				"mention_ids":     map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "maxItems": 20},
			}, []string{"story_id", "story_reference", "body", "mention_ids"}),
		},
		{
			Type: "function", Name: toolAddRelationship,
			Description: "Prepare a proposal to relate two accessible stories. The relationship is written only after explicit user confirmation.", Strict: true,
			Parameters: strictObjectSchema(map[string]any{
				"from_story_id":        map[string]any{"type": []string{"string", "null"}},
				"from_story_reference": map[string]any{"type": []string{"string", "null"}},
				"to_story_id":          map[string]any{"type": []string{"string", "null"}},
				"to_story_reference":   map[string]any{"type": []string{"string", "null"}},
				"association_type":     map[string]any{"type": "string", "enum": []string{"blocking", "related", "duplicate"}},
			}, []string{"from_story_id", "from_story_reference", "to_story_id", "to_story_reference", "association_type"}),
		},
	}
}
