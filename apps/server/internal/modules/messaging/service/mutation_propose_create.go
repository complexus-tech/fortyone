package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

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
	if err := validateStoryTimeContract(args.EstimatedDurationMinutes, args.MinimumFocusBlockMinutes); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidToolArguments, err)
	}
	if err := validateStoryAutoSchedulingContract(args.AutoSchedulingEnabled, false, storyAutoSchedulingStatusOff); err != nil {
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
		if err := validateStoryTimeContract(item.EstimatedDurationMinutes, item.MinimumFocusBlockMinutes); err != nil {
			return nil, fmt.Errorf("%w: stories[%d]: %v", ErrInvalidToolArguments, index, err)
		}
		if err := validateStoryAutoSchedulingContract(item.AutoSchedulingEnabled, false, storyAutoSchedulingStatusOff); err != nil {
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
