package messaging

import (
	"context"
	"fmt"
	"sort"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

type nativeStoryUpdateArguments struct {
	Description       *string `json:"description"`
	AssigneeID        *string `json:"assignee_id"`
	CommandIntent     string  `json:"command_intent"`
	StatusAction      string  `json:"status_action"`
	ExpectedUpdatedAt *string `json:"expected_updated_at"`
}

const nativeMutationUnchangedKind = "native_story_mutation_unchanged"
const nativeStoryMutationReason = "Updated with Maya from a comment."

type nativeMutationUnchanged struct {
	Kind string `json:"kind"`
	Text string `json:"text"`
}

func (e *FortyOneToolExecutor) prepareNativeUpdate(ctx context.Context, scope ToolScope, story messagingStory, assignee string, statusID *uuid.UUID, args nativeStoryUpdateArguments) (storyMutationClaims, error) {
	claims := storyMutationClaims{StatusID: statusID}
	if scope.SourceScope == nil {
		if args.Description != nil || args.AssigneeID != nil || args.CommandIntent != "" || args.StatusAction != "" || args.ExpectedUpdatedAt != nil || assignee == assigneeActionNamed {
			return claims, fmt.Errorf("%w: native update fields require a story-comment source", ErrInvalidToolArguments)
		}
		return claims, nil
	}
	if args.CommandIntent != "direct_command" && args.CommandIntent != "proposal" {
		return claims, fmt.Errorf("%w: command_intent must be direct_command or proposal", ErrInvalidToolArguments)
	}
	claims.SourceScope = scope.SourceScope
	claims.CommandIntent = args.CommandIntent
	if args.ExpectedUpdatedAt != nil {
		expected, err := time.Parse(time.RFC3339Nano, *args.ExpectedUpdatedAt)
		if err != nil {
			return claims, fmt.Errorf("%w: expected_updated_at must be the exact story read timestamp", ErrInvalidToolArguments)
		}
		if !story.UpdatedAt.Equal(expected) {
			return claims, fmt.Errorf("%w: story changed since it was read", ErrStaleMutation)
		}
	}
	if args.Description != nil {
		if args.ExpectedUpdatedAt == nil {
			return claims, fmt.Errorf("%w: description rewrite requires expected_updated_at from get_story", ErrInvalidToolArguments)
		}
		if story.Description != nil && utf8.RuneCountInString(*story.Description) > maximumNativeDescriptionRunes {
			return claims, fmt.Errorf("%w: existing description exceeds the complete rewrite context limit", ErrInvalidToolArguments)
		}
		text, markup, err := rewriteDescription(story.DescriptionHTML, *args.Description)
		if err != nil {
			return claims, err
		}
		if story.Description == nil || *story.Description != text || story.DescriptionHTML == nil || *story.DescriptionHTML != markup {
			claims.Description, claims.DescriptionHTML = &text, &markup
			claims.ChangedFields = append(claims.ChangedFields, "description")
		}
	}
	if assignee == assigneeActionNamed {
		id, err := parseOptionalUUID(args.AssigneeID, "assignee_id")
		if err != nil {
			return claims, err
		}
		if id == nil || e.users == nil {
			return claims, fmt.Errorf("%w: named assignment requires an active team member", ErrInvalidToolArguments)
		}
		member, err := e.activeTeamMemberByID(ctx, scope.WorkspaceID, story.Team, *id)
		if err != nil {
			return claims, err
		}
		if member == nil {
			return claims, fmt.Errorf("%w: assignee is not an active team member", ErrInvalidToolArguments)
		}
		claims.AssigneeID = id
		if story.Assignee == nil || *story.Assignee != *id {
			claims.ChangedFields = append(claims.ChangedFields, "assignee_id")
		}
	} else if args.AssigneeID != nil {
		return claims, fmt.Errorf("%w: assignee_id requires named assignment", ErrInvalidToolArguments)
	}
	if args.StatusAction != "" && args.StatusAction != "unchanged" {
		if args.StatusAction != "close" && args.StatusAction != "reopen" {
			return claims, fmt.Errorf("%w: invalid status_action", ErrInvalidToolArguments)
		}
		if statusID != nil {
			return claims, fmt.Errorf("%w: status_action and status_id cannot be combined", ErrInvalidToolArguments)
		}
		category := "completed"
		if args.StatusAction == "reopen" {
			category = "unstarted"
		}
		statuses, err := e.nativeStatuses(ctx, scope)
		if err != nil {
			return claims, err
		}
		for _, status := range statuses {
			if story.Status != nil && status.ID == *story.Status && ((args.StatusAction == "close" && status.Category == "completed") || (args.StatusAction == "reopen" && !statusIsClosed(status.Category))) {
				claims.StatusID = story.Status
				break
			}
		}
		sort.Slice(statuses, func(i, j int) bool {
			if statuses[i].IsDefault != statuses[j].IsDefault {
				return statuses[i].IsDefault
			}
			if statuses[i].OrderIndex != statuses[j].OrderIndex {
				return statuses[i].OrderIndex < statuses[j].OrderIndex
			}
			return statuses[i].ID.String() < statuses[j].ID.String()
		})
		for _, status := range statuses {
			if claims.StatusID != nil {
				break
			}
			if status.Category == category {
				id := status.ID
				claims.StatusID = &id
				break
			}
		}
		if claims.StatusID == nil {
			return claims, fmt.Errorf("%w: current team has no %s status", ErrInvalidToolArguments, category)
		}
	}
	if err := e.validateNativeUpdateTargets(ctx, scope, claims); err != nil {
		return claims, err
	}
	return claims, nil
}

func (e *FortyOneToolExecutor) nativeStatuses(ctx context.Context, scope ToolScope) ([]messagingState, error) {
	if e.states == nil {
		return nil, fmt.Errorf("%w: status catalog is unavailable", ErrInvalidToolArguments)
	}
	_, joined, err := e.joinedTeams(ctx, scope)
	if err != nil {
		return nil, err
	}
	statuses, _, err := e.scopedStatuses(ctx, scope, joined)
	return statuses, err
}

func (e *FortyOneToolExecutor) validateNativeUpdateTargets(ctx context.Context, scope ToolScope, claims storyMutationClaims) error {
	if scope.SourceScope == nil {
		return nil
	}
	if claims.AssigneeID != nil {
		if e.users == nil {
			return fmt.Errorf("%w: member catalog is unavailable", ErrInvalidConfirmation)
		}
		member, err := e.activeTeamMemberByID(ctx, scope.WorkspaceID, scope.SourceScope.TeamID, *claims.AssigneeID)
		if err != nil {
			return err
		}
		if member == nil {
			return fmt.Errorf("%w: assignee is no longer an active team member", ErrInvalidConfirmation)
		}
	}
	if claims.StatusID != nil {
		statuses, err := e.nativeStatuses(ctx, scope)
		if err != nil {
			return err
		}
		for _, status := range statuses {
			if status.ID == *claims.StatusID {
				return nil
			}
		}
		return fmt.Errorf("%w: status is outside the current team", ErrInvalidConfirmation)
	}
	return nil
}

func addNativeStoryUpdates(updates map[string]any, story messagingStory, claims storyMutationClaims) {
	if claims.Description != nil && claims.DescriptionHTML != nil {
		if story.Description == nil || *story.Description != *claims.Description || story.DescriptionHTML == nil || *story.DescriptionHTML != *claims.DescriptionHTML {
			updates["description"], updates["description_html"] = *claims.Description, *claims.DescriptionHTML
		}
	}
	if claims.AssigneeAction == assigneeActionNamed && claims.AssigneeID != nil && (story.Assignee == nil || *story.Assignee != *claims.AssigneeID) {
		updates["assignee_id"] = *claims.AssigneeID
	}
}

func validateNativeClaims(claims storyMutationClaims) error {
	if claims.SourceScope == nil {
		if claims.Description != nil || claims.DescriptionHTML != nil || claims.AssigneeID != nil || claims.CommandIntent != "" || claims.Review != "" || len(claims.ChangedFields) != 0 || (claims.Operation == StoryMutationUpdate && claims.AssigneeAction == assigneeActionNamed) {
			return fmt.Errorf("%w: native fields require a source binding", ErrInvalidConfirmation)
		}
		return nil
	}
	if _, err := normalizedSourceScope(claims.SourceScope); err != nil {
		return fmt.Errorf("%w: invalid proposal source", ErrInvalidConfirmation)
	}
	if claims.Operation != StoryMutationUpdate || claims.StoryID == nil || *claims.StoryID != claims.SourceScope.StoryID || claims.TeamID != claims.SourceScope.TeamID {
		return fmt.Errorf("%w: native proposal must update its current story", ErrInvalidConfirmation)
	}
	if claims.CommandIntent != "direct_command" && claims.CommandIntent != "proposal" {
		return fmt.Errorf("%w: invalid native command intent", ErrInvalidConfirmation)
	}
	if claims.Review == "" || len(claims.Review) > 64<<10 || claims.SprintID != nil || claims.ObjectiveID != nil || claims.KeyResultID != nil || claims.LabelIDs != nil {
		return fmt.Errorf("%w: native proposal lacks a bounded review or contains unsupported target fields", ErrInvalidConfirmation)
	}
	if (claims.Description == nil) != (claims.DescriptionHTML == nil) || (claims.DescriptionHTML != nil && len(*claims.DescriptionHTML) > maximumNativeDescriptionHTMLBytes) {
		return fmt.Errorf("%w: description representations must remain paired and bounded", ErrInvalidConfirmation)
	}
	if (claims.AssigneeAction == assigneeActionNamed) != (claims.AssigneeID != nil) || (claims.AssigneeID != nil && *claims.AssigneeID == uuid.Nil) {
		return fmt.Errorf("%w: invalid named assignee claim", ErrInvalidConfirmation)
	}
	if len(claims.ChangedFields) == 0 {
		return fmt.Errorf("%w: native update has no changed fields", ErrInvalidConfirmation)
	}
	seen := make(map[string]bool, len(claims.ChangedFields))
	for _, field := range claims.ChangedFields {
		switch field {
		case "description", "title", "priority", "assignee_id", "status_id", "start_date", "end_date", "estimated_duration_minutes", "minimum_focus_block_minutes", "auto_scheduling_enabled", "auto_scheduling_locked":
		default:
			return fmt.Errorf("%w: unsupported native update field", ErrInvalidConfirmation)
		}
		if seen[field] {
			return fmt.Errorf("%w: duplicate native update field", ErrInvalidConfirmation)
		}
		seen[field] = true
	}
	return nil
}
