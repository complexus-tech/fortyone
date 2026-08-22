package agentreadinesshttp

import (
	"context"
	"errors"
	"fmt"
	"html"
	"slices"
	"strings"
	"time"

	keyresults "github.com/complexus-tech/projects-api/internal/modules/keyresults/service"
	objectives "github.com/complexus-tech/projects-api/internal/modules/objectives/service"
	objectivestatus "github.com/complexus-tech/projects-api/internal/modules/objectivestatus/service"
	sprints "github.com/complexus-tech/projects-api/internal/modules/sprints/service"
	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	"github.com/google/uuid"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func (h *Handler) createStory(ctx context.Context, _ *mcp.CallToolRequest, in createStoryInput) (*mcp.CallToolResult, any, error) {
	if !in.Confirmed {
		return nil, nil, invalidToolInput("confirmed must be true after the user approves story creation")
	}
	workspaceID, userID, err := h.authorizeWorkspace(ctx, in.WorkspaceID)
	if err != nil {
		return nil, nil, err
	}
	teamID, err := parseRequiredUUID("teamId", in.TeamID)
	if err != nil {
		return nil, nil, err
	}
	if _, err := h.cfg.Teams.GetByID(ctx, teamID, workspaceID, userID); err != nil {
		return nil, nil, err
	}
	if strings.TrimSpace(in.Title) == "" {
		return nil, nil, invalidToolInput("title is required")
	}
	if !slices.Contains([]string{"", "No Priority", "Low", "Medium", "High", "Urgent"}, in.Priority) {
		return nil, nil, invalidToolInput("priority is invalid")
	}
	statusID, err := optionalUUID(in.StatusID)
	if err != nil {
		return nil, nil, fmt.Errorf("statusId: %w", err)
	}
	if statusID == nil {
		available, listErr := h.cfg.States.TeamList(ctx, workspaceID, teamID)
		if listErr != nil {
			return nil, nil, listErr
		}
		for _, status := range available {
			if status.IsDefault || status.Category == "unstarted" {
				value := status.ID
				statusID = &value
				if status.IsDefault {
					break
				}
			}
		}
		if statusID == nil {
			return nil, nil, errors.New("team has no usable default story status")
		}
	} else {
		status, getErr := h.cfg.States.Get(ctx, workspaceID, *statusID)
		if getErr != nil {
			return nil, nil, getErr
		}
		if status.Team != teamID {
			return nil, nil, invalidToolInput("statusId does not belong to teamId")
		}
	}
	startDate, err := optionalDate(in.StartDate)
	if err != nil {
		return nil, nil, fmt.Errorf("startDate: %w", err)
	}
	endDate, err := optionalDate(in.EndDate)
	if err != nil {
		return nil, nil, fmt.Errorf("endDate: %w", err)
	}
	if startDate != nil && endDate != nil && endDate.Before(*startDate) {
		return nil, nil, invalidToolInput("endDate must be on or after startDate")
	}
	if err := stories.ValidateStoryTimeContract(in.EstimatedDurationMinutes, in.MinimumFocusBlockMinutes); err != nil {
		return nil, nil, invalidToolInput(err.Error())
	}
	assignee, err := optionalUUID(in.AssigneeID)
	if err != nil {
		return nil, nil, fmt.Errorf("assigneeId: %w", err)
	}
	objective, err := optionalUUID(in.ObjectiveID)
	if err != nil {
		return nil, nil, fmt.Errorf("objectiveId: %w", err)
	}
	sprint, err := optionalUUID(in.SprintID)
	if err != nil {
		return nil, nil, fmt.Errorf("sprintId: %w", err)
	}
	keyResult, err := optionalUUID(in.KeyResultID)
	if err != nil {
		return nil, nil, fmt.Errorf("keyResultId: %w", err)
	}
	parent, err := optionalUUID(in.ParentID)
	if err != nil {
		return nil, nil, fmt.Errorf("parentId: %w", err)
	}
	if objective != nil {
		linkedObjective, getErr := h.cfg.Objectives.Get(ctx, *objective, workspaceID)
		if getErr != nil {
			return nil, nil, getErr
		}
		if linkedObjective.Team != teamID {
			return nil, nil, invalidToolInput("objectiveId does not belong to teamId")
		}
	}
	if sprint != nil {
		linkedSprint, getErr := h.cfg.Sprints.GetByID(ctx, *sprint, workspaceID)
		if getErr != nil {
			return nil, nil, getErr
		}
		if linkedSprint.Team != teamID {
			return nil, nil, invalidToolInput("sprintId does not belong to teamId")
		}
	}
	if keyResult != nil {
		linkedKeyResult, getErr := h.cfg.KeyResults.Get(ctx, *keyResult, workspaceID)
		if getErr != nil {
			return nil, nil, getErr
		}
		linkedObjective, getErr := h.cfg.Objectives.Get(ctx, linkedKeyResult.ObjectiveID, workspaceID)
		if getErr != nil {
			return nil, nil, getErr
		}
		if linkedObjective.Team != teamID {
			return nil, nil, invalidToolInput("keyResultId does not belong to teamId")
		}
		if objective != nil && *objective != linkedKeyResult.ObjectiveID {
			return nil, nil, invalidToolInput("keyResultId does not belong to objectiveId")
		}
		objective = &linkedKeyResult.ObjectiveID
	}
	if parent != nil {
		linkedParent, getErr := h.cfg.Stories.Get(ctx, *parent, workspaceID)
		if getErr != nil {
			return nil, nil, getErr
		}
		if linkedParent.Team != teamID {
			return nil, nil, invalidToolInput("parentId does not belong to teamId")
		}
	}
	labels, err := parseUUIDs(in.LabelIDs)
	if err != nil {
		return nil, nil, fmt.Errorf("labelIds: %w", err)
	}
	requestKey := strings.TrimSpace(in.IdempotencyKey)
	if requestKey == "" {
		return nil, nil, invalidToolInput("idempotencyKey is required")
	}
	if len(requestKey) > 200 {
		return nil, nil, invalidToolInput("idempotencyKey must be 200 characters or fewer")
	}
	creationKey := "mcp:" + userID.String() + ":" + requestKey
	description := optionalString(in.Description)
	var descriptionHTML *string
	if description != nil {
		value := "<p>" + strings.ReplaceAll(html.EscapeString(*description), "\n", "<br>") + "</p>"
		descriptionHTML = &value
	}
	priority := in.Priority
	if priority == "" {
		priority = "No Priority"
	}
	created, err := h.cfg.Stories.CreateExternalUserAction(ctx, userID, stories.CoreNewStory{Title: strings.TrimSpace(in.Title), Description: description, DescriptionHTML: descriptionHTML, Status: statusID, Assignee: assignee, Reporter: &userID, Priority: priority, EstimateValue: in.EstimateValue, EstimatedDurationMinutes: in.EstimatedDurationMinutes, MinimumFocusBlockMinutes: in.MinimumFocusBlockMinutes, AutoSchedulingEnabled: in.AutoSchedulingEnabled, StartDate: startDate, EndDate: endDate, Sprint: sprint, Objective: objective, KeyResult: keyResult, Parent: parent, LabelIDs: labels, Team: teamID, CreationKey: &creationKey}, workspaceID)
	if err != nil {
		return nil, nil, err
	}
	return nil, map[string]any{"id": created.ID, "sequenceId": created.SequenceID, "teamCode": created.TeamCode, "title": created.Title, "createdNow": created.CreatedNow, "estimatedDurationMinutes": created.EstimatedDurationMinutes, "minimumFocusBlockMinutes": created.MinimumFocusBlockMinutes, "autoSchedulingEnabled": created.AutoSchedulingEnabled}, nil
}

func (h *Handler) updateStory(ctx context.Context, _ *mcp.CallToolRequest, in updateStoryInput) (*mcp.CallToolResult, any, error) {
	if !in.Confirmed {
		return nil, nil, invalidToolInput("confirmed must be true after the user approves the exact story changes")
	}
	workspaceID, userID, err := h.authorizeWorkspace(ctx, in.WorkspaceID)
	if err != nil {
		return nil, nil, err
	}
	storyID, err := parseRequiredUUID("id", in.ID)
	if err != nil {
		return nil, nil, err
	}
	expectedUpdatedAt, err := requiredTimestamp("expectedUpdatedAt", in.ExpectedUpdatedAt)
	if err != nil {
		return nil, nil, err
	}
	current, err := h.cfg.Stories.Get(ctx, storyID, workspaceID)
	if err != nil {
		return nil, nil, err
	}
	if _, err := h.cfg.Teams.GetByID(ctx, current.Team, workspaceID, userID); err != nil {
		return nil, nil, err
	}

	updates := make(map[string]any)
	if in.Title != nil {
		title := strings.TrimSpace(*in.Title)
		if title == "" {
			return nil, nil, invalidToolInput("title cannot be empty")
		}
		updates["title"] = title
	}
	if in.Description != nil {
		description := strings.TrimSpace(*in.Description)
		updates["description"] = description
		updates["description_html"] = "<p>" + strings.ReplaceAll(html.EscapeString(description), "\n", "<br>") + "</p>"
	}
	if in.Priority != nil {
		priority := strings.TrimSpace(*in.Priority)
		if !slices.Contains([]string{"No Priority", "Low", "Medium", "High", "Urgent"}, priority) {
			return nil, nil, invalidToolInput("priority is invalid")
		}
		updates["priority"] = priority
	}
	if in.EstimateValue != nil {
		updates["estimate_unit"] = in.EstimateValue
	}
	if in.EstimatedDurationMinutes != nil {
		updates["estimated_duration_minutes"] = in.EstimatedDurationMinutes
	}
	if in.MinimumFocusBlockMinutes != nil {
		updates["minimum_focus_block_minutes"] = in.MinimumFocusBlockMinutes
	}
	if in.AutoSchedulingEnabled != nil {
		updates["auto_scheduling_enabled"] = *in.AutoSchedulingEnabled
	}
	if in.StatusID != nil {
		statusID, parseErr := parseRequiredUUID("statusId", *in.StatusID)
		if parseErr != nil {
			return nil, nil, parseErr
		}
		status, getErr := h.cfg.States.Get(ctx, workspaceID, statusID)
		if getErr != nil {
			return nil, nil, getErr
		}
		if status.Team != current.Team {
			return nil, nil, invalidToolInput("statusId does not belong to the story's team")
		}
		updates["status_id"] = statusID
	}
	if in.AssigneeID != nil {
		assigneeID, parseErr := optionalUUID(*in.AssigneeID)
		if parseErr != nil {
			return nil, nil, fmt.Errorf("assigneeId: %w", parseErr)
		}
		updates["assignee_id"] = assigneeID
	}
	if in.SprintID != nil {
		sprintID, parseErr := optionalUUID(*in.SprintID)
		if parseErr != nil {
			return nil, nil, fmt.Errorf("sprintId: %w", parseErr)
		}
		if sprintID != nil {
			sprint, getErr := h.cfg.Sprints.GetByID(ctx, *sprintID, workspaceID)
			if getErr != nil {
				return nil, nil, getErr
			}
			if sprint.Team != current.Team {
				return nil, nil, invalidToolInput("sprintId does not belong to the story's team")
			}
		}
		updates["sprint_id"] = sprintID
	}

	effectiveObjective := current.Objective
	if in.ObjectiveID != nil {
		objectiveID, parseErr := optionalUUID(*in.ObjectiveID)
		if parseErr != nil {
			return nil, nil, fmt.Errorf("objectiveId: %w", parseErr)
		}
		if objectiveID != nil {
			objective, getErr := h.cfg.Objectives.Get(ctx, *objectiveID, workspaceID)
			if getErr != nil {
				return nil, nil, getErr
			}
			if objective.Team != current.Team {
				return nil, nil, invalidToolInput("objectiveId does not belong to the story's team")
			}
		}
		effectiveObjective = objectiveID
		updates["objective_id"] = objectiveID
	}

	if in.KeyResultID != nil {
		keyResultID, parseErr := optionalUUID(*in.KeyResultID)
		if parseErr != nil {
			return nil, nil, fmt.Errorf("keyResultId: %w", parseErr)
		}
		updates["key_result_id"] = keyResultID
		if keyResultID != nil {
			keyResult, getErr := h.cfg.KeyResults.Get(ctx, *keyResultID, workspaceID)
			if getErr != nil {
				return nil, nil, getErr
			}
			objective, getErr := h.cfg.Objectives.Get(ctx, keyResult.ObjectiveID, workspaceID)
			if getErr != nil {
				return nil, nil, getErr
			}
			if objective.Team != current.Team {
				return nil, nil, invalidToolInput("keyResultId does not belong to the story's team")
			}
			if effectiveObjective == nil {
				if in.ObjectiveID != nil {
					return nil, nil, invalidToolInput("objectiveId cannot be cleared while a key result is linked")
				}
				effectiveObjective = &keyResult.ObjectiveID
				updates["objective_id"] = effectiveObjective
			} else if *effectiveObjective != keyResult.ObjectiveID {
				return nil, nil, invalidToolInput("keyResultId does not belong to objectiveId")
			}
		}
	}

	if in.ParentID != nil {
		parentID, parseErr := optionalUUID(*in.ParentID)
		if parseErr != nil {
			return nil, nil, fmt.Errorf("parentId: %w", parseErr)
		}
		if parentID != nil {
			if *parentID == storyID {
				return nil, nil, invalidToolInput("parentId cannot reference the story itself")
			}
			parent, getErr := h.cfg.Stories.Get(ctx, *parentID, workspaceID)
			if getErr != nil {
				return nil, nil, getErr
			}
			if parent.Team != current.Team {
				return nil, nil, invalidToolInput("parentId does not belong to the story's team")
			}
		}
		updates["parent_id"] = parentID
	}

	effectiveStart, effectiveEnd := current.StartDate, current.EndDate
	for _, field := range []struct {
		name      string
		raw       *string
		key       string
		effective **time.Time
	}{
		{name: "startDate", raw: in.StartDate, key: "start_date", effective: &effectiveStart},
		{name: "endDate", raw: in.EndDate, key: "end_date", effective: &effectiveEnd},
	} {
		if field.raw == nil {
			continue
		}
		value, parseErr := optionalDate(*field.raw)
		if parseErr != nil {
			return nil, nil, fmt.Errorf("%s: %w", field.name, parseErr)
		}
		*field.effective = value
		updates[field.key] = value
	}
	if effectiveStart != nil && effectiveEnd != nil && effectiveEnd.Before(*effectiveStart) {
		return nil, nil, invalidToolInput("endDate must be on or after startDate")
	}
	effectiveDuration, effectiveFocusBlock := current.EstimatedDurationMinutes, current.MinimumFocusBlockMinutes
	if in.EstimatedDurationMinutes != nil {
		effectiveDuration = in.EstimatedDurationMinutes
	}
	if in.MinimumFocusBlockMinutes != nil {
		effectiveFocusBlock = in.MinimumFocusBlockMinutes
	}
	if err := stories.ValidateStoryTimeContract(effectiveDuration, effectiveFocusBlock); err != nil {
		return nil, nil, invalidToolInput(err.Error())
	}
	if len(updates) == 0 {
		return nil, nil, invalidToolInput("at least one story field must be provided")
	}
	if err := h.cfg.Stories.UpdateExternalUserActionIfUnchanged(ctx, userID, storyID, workspaceID, expectedUpdatedAt, updates); err != nil {
		return nil, nil, err
	}
	updated, err := h.cfg.Stories.Get(ctx, storyID, workspaceID)
	if err != nil {
		return nil, nil, err
	}
	return nil, map[string]any{"story": storyToolResult(updated)}, nil
}

func (h *Handler) addStoryComment(ctx context.Context, _ *mcp.CallToolRequest, in addStoryCommentInput) (*mcp.CallToolResult, any, error) {
	if !in.Confirmed {
		return nil, nil, invalidToolInput("confirmed must be true after the user approves the exact comment")
	}
	workspaceID, userID, err := h.authorizeWorkspace(ctx, in.WorkspaceID)
	if err != nil {
		return nil, nil, err
	}
	storyID, err := parseRequiredUUID("storyId", in.StoryID)
	if err != nil {
		return nil, nil, err
	}
	story, err := h.cfg.Stories.Get(ctx, storyID, workspaceID)
	if err != nil {
		return nil, nil, err
	}
	if _, err := h.cfg.Teams.GetByID(ctx, story.Team, workspaceID, userID); err != nil {
		return nil, nil, err
	}
	commentText := strings.TrimSpace(in.Comment)
	if commentText == "" {
		return nil, nil, invalidToolInput("comment cannot be empty")
	}
	if len(commentText) > 10000 {
		return nil, nil, invalidToolInput("comment must be 10,000 characters or fewer")
	}
	parentID, err := optionalUUID(in.ParentID)
	if err != nil {
		return nil, nil, fmt.Errorf("parentId: %w", err)
	}
	if parentID != nil {
		parent, getErr := h.cfg.Stories.GetComment(ctx, *parentID)
		if getErr != nil {
			return nil, nil, getErr
		}
		if parent.StoryID != storyID {
			return nil, nil, invalidToolInput("parentId does not belong to storyId")
		}
	}
	created, err := h.cfg.Stories.CreateCommentExternal(ctx, userID, workspaceID, stories.CoreNewComment{StoryID: storyID, Parent: parentID, UserID: userID, Comment: commentText})
	if err != nil {
		return nil, nil, err
	}
	return nil, map[string]any{"comment": created}, nil
}

func (h *Handler) setStoryArchived(ctx context.Context, _ *mcp.CallToolRequest, in setStoryArchivedInput) (*mcp.CallToolResult, any, error) {
	if !in.Confirmed {
		return nil, nil, invalidToolInput("confirmed must be true after the user approves archiving or restoring the story")
	}
	workspaceID, userID, err := h.authorizeWorkspace(ctx, in.WorkspaceID)
	if err != nil {
		return nil, nil, err
	}
	storyID, err := parseRequiredUUID("id", in.ID)
	if err != nil {
		return nil, nil, err
	}
	story, err := h.cfg.Stories.Get(ctx, storyID, workspaceID)
	if err != nil {
		return nil, nil, err
	}
	if _, err := h.cfg.Teams.GetByID(ctx, story.Team, workspaceID, userID); err != nil {
		return nil, nil, err
	}
	if (story.ArchivedAt != nil) == in.Archived {
		return nil, map[string]any{"id": storyID, "archived": in.Archived, "changed": false}, nil
	}
	if in.Archived {
		err = h.cfg.Stories.BulkArchive(ctx, []uuid.UUID{storyID}, workspaceID)
	} else {
		err = h.cfg.Stories.BulkUnarchive(ctx, []uuid.UUID{storyID}, workspaceID)
	}
	if err != nil {
		return nil, nil, err
	}
	return nil, map[string]any{"id": storyID, "archived": in.Archived, "changed": true}, nil
}

func (h *Handler) createSprint(ctx context.Context, _ *mcp.CallToolRequest, in createSprintInput) (*mcp.CallToolResult, any, error) {
	if !in.Confirmed {
		return nil, nil, invalidToolInput("confirmed must be true after the user approves sprint creation")
	}
	workspaceID, userID, err := h.authorizeWorkspace(ctx, in.WorkspaceID)
	if err != nil {
		return nil, nil, err
	}
	teamID, err := parseRequiredUUID("teamId", in.TeamID)
	if err != nil {
		return nil, nil, err
	}
	if _, err := h.cfg.Teams.GetByID(ctx, teamID, workspaceID, userID); err != nil {
		return nil, nil, err
	}
	if strings.TrimSpace(in.Name) == "" {
		return nil, nil, invalidToolInput("name is required")
	}
	startDate, err := requiredDate("startDate", in.StartDate)
	if err != nil {
		return nil, nil, err
	}
	endDate, err := requiredDate("endDate", in.EndDate)
	if err != nil {
		return nil, nil, err
	}
	if endDate.Before(startDate) {
		return nil, nil, invalidToolInput("endDate must be on or after startDate")
	}
	objective, err := optionalUUID(in.ObjectiveID)
	if err != nil {
		return nil, nil, fmt.Errorf("objectiveId: %w", err)
	}
	if objective != nil {
		linkedObjective, getErr := h.cfg.Objectives.Get(ctx, *objective, workspaceID)
		if getErr != nil {
			return nil, nil, getErr
		}
		if linkedObjective.Team != teamID {
			return nil, nil, invalidToolInput("objectiveId does not belong to teamId")
		}
	}
	created, err := h.cfg.Sprints.Create(ctx, sprints.CoreNewSprint{Name: strings.TrimSpace(in.Name), Goal: optionalString(in.Goal), Objective: objective, Team: teamID, Workspace: workspaceID, StartDate: startDate, EndDate: endDate}, &userID)
	if err != nil {
		return nil, nil, err
	}
	return nil, map[string]any{"sprint": sprintToolResult(created)}, nil
}

func (h *Handler) createObjective(ctx context.Context, _ *mcp.CallToolRequest, in createObjectiveInput) (*mcp.CallToolResult, any, error) {
	if !in.Confirmed {
		return nil, nil, invalidToolInput("confirmed must be true after the user approves objective creation")
	}
	workspaceID, userID, err := h.authorizeWorkspace(ctx, in.WorkspaceID)
	if err != nil {
		return nil, nil, err
	}
	teamID, err := parseRequiredUUID("teamId", in.TeamID)
	if err != nil {
		return nil, nil, err
	}
	if _, err := h.cfg.Teams.GetByID(ctx, teamID, workspaceID, userID); err != nil {
		return nil, nil, err
	}
	if strings.TrimSpace(in.Name) == "" {
		return nil, nil, invalidToolInput("name is required")
	}
	statusID, err := optionalUUID(in.StatusID)
	if err != nil {
		return nil, nil, fmt.Errorf("statusId: %w", err)
	}
	if statusID == nil {
		statuses, listErr := h.cfg.ObjectiveStatuses.List(ctx, workspaceID)
		if listErr != nil {
			return nil, nil, listErr
		}
		for _, status := range statuses {
			if status.IsDefault || status.Category == "unstarted" {
				value := status.ID
				statusID = &value
				if status.IsDefault {
					break
				}
			}
		}
		if statusID == nil {
			return nil, nil, errors.New("workspace has no usable default objective status")
		}
	} else {
		statuses, listErr := h.cfg.ObjectiveStatuses.List(ctx, workspaceID)
		if listErr != nil {
			return nil, nil, listErr
		}
		if !slices.ContainsFunc(statuses, func(status objectivestatus.CoreObjectiveStatus) bool {
			return status.ID == *statusID
		}) {
			return nil, nil, invalidToolInput("statusId is not an objective status in this workspace")
		}
	}
	lead, err := optionalUUID(in.LeadUserID)
	if err != nil {
		return nil, nil, fmt.Errorf("leadUserId: %w", err)
	}
	startDate, err := optionalDate(in.StartDate)
	if err != nil {
		return nil, nil, err
	}
	endDate, err := optionalDate(in.EndDate)
	if err != nil {
		return nil, nil, err
	}
	if startDate != nil && endDate != nil && endDate.Before(*startDate) {
		return nil, nil, invalidToolInput("endDate must be on or after startDate")
	}
	created, _, err := h.cfg.Objectives.Create(ctx, objectives.CoreNewObjective{Name: strings.TrimSpace(in.Name), Description: optionalString(in.Description), LeadUser: lead, Team: teamID, StartDate: startDate, EndDate: endDate, IsPrivate: in.IsPrivate, Status: *statusID, Priority: optionalString(in.Priority), Color: objectives.DefaultObjectiveColor, CreatedBy: userID}, workspaceID, nil)
	if err != nil {
		return nil, nil, err
	}
	return nil, map[string]any{"objective": objectiveToolResult(created)}, nil
}

func (h *Handler) updateObjective(ctx context.Context, _ *mcp.CallToolRequest, in updateObjectiveInput) (*mcp.CallToolResult, any, error) {
	if !in.Confirmed {
		return nil, nil, invalidToolInput("confirmed must be true after the user approves the exact objective changes")
	}
	workspaceID, userID, err := h.authorizeWorkspace(ctx, in.WorkspaceID)
	if err != nil {
		return nil, nil, err
	}
	objectiveID, err := parseRequiredUUID("id", in.ID)
	if err != nil {
		return nil, nil, err
	}
	expectedUpdatedAt, err := requiredTimestamp("expectedUpdatedAt", in.ExpectedUpdatedAt)
	if err != nil {
		return nil, nil, err
	}
	current, err := h.cfg.Objectives.Get(ctx, objectiveID, workspaceID)
	if err != nil {
		return nil, nil, err
	}
	if _, err := h.cfg.Teams.GetByID(ctx, current.Team, workspaceID, userID); err != nil {
		return nil, nil, err
	}

	updates := make(map[string]any)
	if in.Name != nil {
		name := strings.TrimSpace(*in.Name)
		if name == "" {
			return nil, nil, invalidToolInput("name cannot be empty")
		}
		updates["name"] = name
	}
	if in.Description != nil {
		updates["description"] = strings.TrimSpace(*in.Description)
	}
	if in.LeadUserID != nil {
		lead, parseErr := optionalUUID(*in.LeadUserID)
		if parseErr != nil {
			return nil, nil, fmt.Errorf("leadUserId: %w", parseErr)
		}
		updates["lead_user_id"] = lead
	}
	if in.StatusID != nil {
		statusID, parseErr := parseRequiredUUID("statusId", *in.StatusID)
		if parseErr != nil {
			return nil, nil, parseErr
		}
		statuses, listErr := h.cfg.ObjectiveStatuses.List(ctx, workspaceID)
		if listErr != nil {
			return nil, nil, listErr
		}
		found := false
		for _, status := range statuses {
			if status.ID == statusID {
				found = true
				break
			}
		}
		if !found {
			return nil, nil, invalidToolInput("statusId is not an objective status in this workspace")
		}
		updates["status_id"] = statusID
	}
	if in.Priority != nil {
		updates["priority"] = strings.TrimSpace(*in.Priority)
	}
	if in.Health != nil {
		health := objectives.ObjectiveHealth(strings.TrimSpace(*in.Health))
		if !slices.Contains([]objectives.ObjectiveHealth{objectives.HealthOnTrack, objectives.HealthAtRisk, objectives.HealthOffTrack}, health) {
			return nil, nil, invalidToolInput("health must be On Track, At Risk, or Off Track")
		}
		updates["health"] = health
	}
	if in.IsPrivate != nil {
		updates["is_private"] = *in.IsPrivate
	}
	effectiveStart, effectiveEnd := current.StartDate, current.EndDate
	for _, field := range []struct {
		name      string
		raw       *string
		key       string
		effective **time.Time
	}{
		{name: "startDate", raw: in.StartDate, key: "start_date", effective: &effectiveStart},
		{name: "endDate", raw: in.EndDate, key: "end_date", effective: &effectiveEnd},
	} {
		if field.raw == nil {
			continue
		}
		value, parseErr := optionalDate(*field.raw)
		if parseErr != nil {
			return nil, nil, fmt.Errorf("%s: %w", field.name, parseErr)
		}
		*field.effective = value
		updates[field.key] = value
	}
	if effectiveStart != nil && effectiveEnd != nil && effectiveEnd.Before(*effectiveStart) {
		return nil, nil, invalidToolInput("endDate must be on or after startDate")
	}
	if len(updates) == 0 {
		return nil, nil, invalidToolInput("at least one objective field must be provided")
	}
	if err := h.cfg.Objectives.UpdateExternalUserActionIfUnchanged(ctx, objectiveID, workspaceID, userID, expectedUpdatedAt, strings.TrimSpace(in.Comment), updates); err != nil {
		return nil, nil, err
	}
	updated, err := h.cfg.Objectives.Get(ctx, objectiveID, workspaceID)
	if err != nil {
		return nil, nil, err
	}
	return nil, map[string]any{"objective": objectiveToolResult(updated)}, nil
}

func (h *Handler) createKeyResult(ctx context.Context, _ *mcp.CallToolRequest, in createKeyResultInput) (*mcp.CallToolResult, any, error) {
	if !in.Confirmed {
		return nil, nil, invalidToolInput("confirmed must be true after the user approves key-result creation")
	}
	workspaceID, userID, err := h.authorizeWorkspace(ctx, in.WorkspaceID)
	if err != nil {
		return nil, nil, err
	}
	objectiveID, err := parseRequiredUUID("objectiveId", in.ObjectiveID)
	if err != nil {
		return nil, nil, err
	}
	if strings.TrimSpace(in.Name) == "" {
		return nil, nil, invalidToolInput("name is required")
	}
	objective, err := h.cfg.Objectives.Get(ctx, objectiveID, workspaceID)
	if err != nil {
		return nil, nil, err
	}
	if _, err := h.cfg.Teams.GetByID(ctx, objective.Team, workspaceID, userID); err != nil {
		return nil, nil, err
	}
	if !slices.Contains([]string{"percentage", "number", "boolean"}, in.MeasurementType) {
		return nil, nil, invalidToolInput("measurementType must be percentage, number, or boolean")
	}
	lead, err := optionalUUID(in.LeadUserID)
	if err != nil {
		return nil, nil, err
	}
	contributors, err := parseUUIDs(in.ContributorIDs)
	if err != nil {
		return nil, nil, err
	}
	startDate, err := optionalDate(in.StartDate)
	if err != nil {
		return nil, nil, err
	}
	endDate, err := optionalDate(in.EndDate)
	if err != nil {
		return nil, nil, err
	}
	if startDate != nil && endDate != nil && endDate.Before(*startDate) {
		return nil, nil, invalidToolInput("endDate must be on or after startDate")
	}
	created, err := h.cfg.KeyResults.Create(ctx, keyresults.CoreNewKeyResult{ObjectiveID: objectiveID, Name: strings.TrimSpace(in.Name), MeasurementType: in.MeasurementType, StartValue: in.StartValue, CurrentValue: in.CurrentValue, TargetValue: in.TargetValue, Lead: lead, Contributors: contributors, StartDate: startDate, EndDate: endDate, CreatedBy: userID}, workspaceID)
	if err != nil {
		return nil, nil, err
	}
	return nil, map[string]any{"keyResult": keyResultToolResult(created)}, nil
}

func (h *Handler) updateKeyResult(ctx context.Context, _ *mcp.CallToolRequest, in updateKeyResultInput) (*mcp.CallToolResult, any, error) {
	if !in.Confirmed {
		return nil, nil, invalidToolInput("confirmed must be true after the user approves the exact key-result changes")
	}
	workspaceID, userID, err := h.authorizeWorkspace(ctx, in.WorkspaceID)
	if err != nil {
		return nil, nil, err
	}
	keyResultID, err := parseRequiredUUID("id", in.ID)
	if err != nil {
		return nil, nil, err
	}
	expectedUpdatedAt, err := requiredTimestamp("expectedUpdatedAt", in.ExpectedUpdatedAt)
	if err != nil {
		return nil, nil, err
	}
	current, err := h.cfg.KeyResults.Get(ctx, keyResultID, workspaceID)
	if err != nil {
		return nil, nil, err
	}
	objective, err := h.cfg.Objectives.Get(ctx, current.ObjectiveID, workspaceID)
	if err != nil {
		return nil, nil, err
	}
	if _, err := h.cfg.Teams.GetByID(ctx, objective.Team, workspaceID, userID); err != nil {
		return nil, nil, err
	}

	updates := make(map[string]any)
	if in.Name != nil {
		name := strings.TrimSpace(*in.Name)
		if name == "" {
			return nil, nil, invalidToolInput("name cannot be empty")
		}
		updates["name"] = name
	}
	if in.MeasurementType != nil {
		measurementType := strings.TrimSpace(*in.MeasurementType)
		if !slices.Contains([]string{"percentage", "number", "boolean"}, measurementType) {
			return nil, nil, invalidToolInput("measurementType must be percentage, number, or boolean")
		}
		updates["measurement_type"] = measurementType
	}
	if in.StartValue != nil {
		updates["start_value"] = in.StartValue
	}
	if in.CurrentValue != nil {
		updates["current_value"] = in.CurrentValue
	}
	if in.TargetValue != nil {
		updates["target_value"] = in.TargetValue
	}
	if in.LeadUserID != nil {
		lead, parseErr := optionalUUID(*in.LeadUserID)
		if parseErr != nil {
			return nil, nil, fmt.Errorf("leadUserId: %w", parseErr)
		}
		updates["lead"] = lead
	}
	effectiveStart, effectiveEnd := current.StartDate, current.EndDate
	for _, field := range []struct {
		name      string
		raw       *string
		key       string
		effective **time.Time
	}{
		{name: "startDate", raw: in.StartDate, key: "start_date", effective: &effectiveStart},
		{name: "endDate", raw: in.EndDate, key: "end_date", effective: &effectiveEnd},
	} {
		if field.raw == nil {
			continue
		}
		value, parseErr := optionalDate(*field.raw)
		if parseErr != nil {
			return nil, nil, fmt.Errorf("%s: %w", field.name, parseErr)
		}
		*field.effective = value
		updates[field.key] = value
	}
	if effectiveStart != nil && effectiveEnd != nil && effectiveEnd.Before(*effectiveStart) {
		return nil, nil, invalidToolInput("endDate must be on or after startDate")
	}
	if len(updates) == 0 {
		return nil, nil, invalidToolInput("at least one key-result field must be provided")
	}
	if err := h.cfg.KeyResults.UpdateExternalUserActionIfUnchanged(ctx, keyResultID, workspaceID, userID, expectedUpdatedAt, updates, strings.TrimSpace(in.Comment)); err != nil {
		return nil, nil, err
	}
	updated, err := h.cfg.KeyResults.Get(ctx, keyResultID, workspaceID)
	if err != nil {
		return nil, nil, err
	}
	return nil, map[string]any{"keyResult": keyResultToolResult(updated)}, nil
}
