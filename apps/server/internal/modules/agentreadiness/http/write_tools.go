package agentreadinesshttp

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"html"
	"slices"
	"strings"

	keyresults "github.com/complexus-tech/projects-api/internal/modules/keyresults/service"
	objectives "github.com/complexus-tech/projects-api/internal/modules/objectives/service"
	sprints "github.com/complexus-tech/projects-api/internal/modules/sprints/service"
	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func (h *Handler) createStory(ctx context.Context, _ *mcp.CallToolRequest, in createStoryInput) (*mcp.CallToolResult, any, error) {
	if !in.Confirmed {
		return nil, nil, errors.New("confirmed must be true after the user approves story creation")
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
		return nil, nil, errors.New("title is required")
	}
	if !slices.Contains([]string{"", "No Priority", "Low", "Medium", "High", "Urgent"}, in.Priority) {
		return nil, nil, errors.New("priority is invalid")
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
			return nil, nil, errors.New("statusId does not belong to teamId")
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
		return nil, nil, errors.New("endDate must be on or after startDate")
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
	labels, err := parseUUIDs(in.LabelIDs)
	if err != nil {
		return nil, nil, fmt.Errorf("labelIds: %w", err)
	}
	creationKey := strings.TrimSpace(in.IdempotencyKey)
	if creationKey == "" {
		sum := sha256.Sum256([]byte(userID.String() + "\x00" + in.WorkspaceID + "\x00" + in.TeamID + "\x00" + in.Title))
		creationKey = "mcp:" + hex.EncodeToString(sum[:])
	}
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

func (h *Handler) createSprint(ctx context.Context, _ *mcp.CallToolRequest, in createSprintInput) (*mcp.CallToolResult, any, error) {
	if !in.Confirmed {
		return nil, nil, errors.New("confirmed must be true after the user approves sprint creation")
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
		return nil, nil, errors.New("name is required")
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
		return nil, nil, errors.New("endDate must be on or after startDate")
	}
	objective, err := optionalUUID(in.ObjectiveID)
	if err != nil {
		return nil, nil, fmt.Errorf("objectiveId: %w", err)
	}
	created, err := h.cfg.Sprints.Create(ctx, sprints.CoreNewSprint{Name: strings.TrimSpace(in.Name), Goal: optionalString(in.Goal), Objective: objective, Team: teamID, Workspace: workspaceID, StartDate: startDate, EndDate: endDate}, &userID)
	if err != nil {
		return nil, nil, err
	}
	return nil, map[string]any{"sprint": created}, nil
}

func (h *Handler) createObjective(ctx context.Context, _ *mcp.CallToolRequest, in createObjectiveInput) (*mcp.CallToolResult, any, error) {
	if !in.Confirmed {
		return nil, nil, errors.New("confirmed must be true after the user approves objective creation")
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
		return nil, nil, errors.New("name is required")
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
		return nil, nil, errors.New("endDate must be on or after startDate")
	}
	created, _, err := h.cfg.Objectives.Create(ctx, objectives.CoreNewObjective{Name: strings.TrimSpace(in.Name), Description: optionalString(in.Description), LeadUser: lead, Team: teamID, StartDate: startDate, EndDate: endDate, IsPrivate: in.IsPrivate, Status: *statusID, Priority: optionalString(in.Priority), Color: objectives.DefaultObjectiveColor, CreatedBy: userID}, workspaceID, nil)
	if err != nil {
		return nil, nil, err
	}
	return nil, map[string]any{"objective": created}, nil
}

func (h *Handler) createKeyResult(ctx context.Context, _ *mcp.CallToolRequest, in createKeyResultInput) (*mcp.CallToolResult, any, error) {
	if !in.Confirmed {
		return nil, nil, errors.New("confirmed must be true after the user approves key-result creation")
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
		return nil, nil, errors.New("name is required")
	}
	if !slices.Contains([]string{"percentage", "number", "boolean"}, in.MeasurementType) {
		return nil, nil, errors.New("measurementType must be percentage, number, or boolean")
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
	created, err := h.cfg.KeyResults.Create(ctx, keyresults.CoreNewKeyResult{ObjectiveID: objectiveID, Name: strings.TrimSpace(in.Name), MeasurementType: in.MeasurementType, StartValue: in.StartValue, CurrentValue: in.CurrentValue, TargetValue: in.TargetValue, Lead: lead, Contributors: contributors, StartDate: startDate, EndDate: endDate, CreatedBy: userID}, workspaceID)
	if err != nil {
		return nil, nil, err
	}
	return nil, map[string]any{"keyResult": created}, nil
}
