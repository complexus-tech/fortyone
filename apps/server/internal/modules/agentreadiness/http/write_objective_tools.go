package agentreadinesshttp

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	objectives "github.com/complexus-tech/projects-api/internal/modules/objectives/service"
	objectivestatus "github.com/complexus-tech/projects-api/internal/modules/objectivestatus/service"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

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
