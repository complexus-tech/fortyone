package messaging

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

func (e *FortyOneToolExecutor) listStatuses(ctx context.Context, scope ToolScope, raw json.RawMessage) (json.RawMessage, error) {
	var args struct {
		TeamID *string `json:"team_id"`
		Limit  *int    `json:"limit"`
	}
	if err := decodeToolArguments(raw, &args, "team_id", "limit"); err != nil {
		return nil, err
	}
	limit, err := normalizedLimit(args.Limit)
	if err != nil {
		return nil, err
	}

	_, joinedByID, err := e.joinedTeams(ctx, scope)
	if err != nil {
		return nil, err
	}
	teamID, err := accessibleTeamID(args.TeamID, joinedByID)
	if err != nil {
		return nil, err
	}
	if len(joinedByID) == 0 {
		return marshalToolResult(listStatusesResult{Statuses: []statusResult{}})
	}

	visible, _, err := e.scopedStatuses(ctx, scope, joinedByID)
	if err != nil {
		return nil, err
	}
	result := listStatusesResult{Statuses: make([]statusResult, 0, min(len(visible), limit))}
	for _, status := range visible {
		if teamID != nil && status.Team != *teamID {
			continue
		}
		team := joinedByID[status.Team]
		result.Total++
		if len(result.Statuses) == limit {
			continue
		}
		result.Statuses = append(result.Statuses, statusResult{
			ID:         status.ID,
			Name:       status.Name,
			Category:   status.Category,
			OrderIndex: status.OrderIndex,
			IsDefault:  status.IsDefault,
			TeamID:     team.ID,
			TeamName:   team.Name,
			TeamCode:   strings.ToUpper(strings.TrimSpace(team.Code)),
		})
	}
	result.Truncated = result.Total > len(result.Statuses)
	return marshalToolResult(result)
}

func (e *FortyOneToolExecutor) listTeamMembers(ctx context.Context, scope ToolScope, raw json.RawMessage) (json.RawMessage, error) {
	var args struct {
		TeamID string  `json:"team_id"`
		Query  *string `json:"query"`
		Limit  *int    `json:"limit"`
	}
	if err := decodeToolArguments(raw, &args, "team_id", "query", "limit"); err != nil {
		return nil, err
	}
	limit, err := normalizedLimit(args.Limit)
	if err != nil {
		return nil, err
	}

	_, joinedByID, err := e.joinedTeams(ctx, scope)
	if err != nil {
		return nil, err
	}
	teamID, err := accessibleTeamID(&args.TeamID, joinedByID)
	if err != nil {
		return nil, err
	}
	team := joinedByID[*teamID]
	query := ""
	if args.Query != nil {
		query = strings.TrimSpace(*args.Query)
		if len([]rune(query)) > maxSearchRunes {
			return nil, fmt.Errorf("%w: query must not exceed %d characters", ErrInvalidToolArguments, maxSearchRunes)
		}
	}

	items, err := e.users.List(ctx, scope.WorkspaceID, messagingUserListFilter{
		TeamID: teamID,
		Search: query,
		Limit:  limit + 1,
	})
	if err != nil {
		return nil, fmt.Errorf("list team members: %w", err)
	}
	result := listTeamMembersResult{
		TeamName: team.Name,
		TeamCode: strings.ToUpper(strings.TrimSpace(team.Code)),
		Members:  make([]teamMemberResult, 0, min(len(items), limit)),
	}
	seen := make(map[uuid.UUID]struct{}, len(items))
	for _, member := range items {
		if member.ID == uuid.Nil || !member.IsActive || member.IsSystem {
			continue
		}
		if _, duplicate := seen[member.ID]; duplicate {
			continue
		}
		seen[member.ID] = struct{}{}
		displayName := memberDisplayName(member)
		username := strings.TrimSpace(member.Username)
		if displayName == "" && username == "" {
			continue
		}
		result.Total++
		if len(result.Members) == limit {
			continue
		}
		result.Members = append(result.Members, teamMemberResult{
			ID:          member.ID,
			DisplayName: displayName,
			Username:    username,
			Active:      true,
			RoleTitle:   memberRoleTitle(member),
		})
	}
	result.Truncated = result.Total > len(result.Members)
	return marshalToolResult(result)
}

func (e *FortyOneToolExecutor) getStory(ctx context.Context, scope ToolScope, raw json.RawMessage) (json.RawMessage, error) {
	var args struct {
		StoryReference string `json:"story_reference"`
	}
	if err := decodeToolArguments(raw, &args, "story_reference"); err != nil {
		return nil, err
	}

	_, joinedByID, err := e.joinedTeams(ctx, scope)
	if err != nil {
		return nil, err
	}
	storyReference, expectedTeam, err := accessibleStoryReference(args.StoryReference, joinedByID)
	if err != nil {
		return nil, err
	}
	story, err := e.storyReader.QueryByRef(ctx, scope.WorkspaceID, storyReference)
	if err != nil {
		return nil, fmt.Errorf("get story: %w", err)
	}
	if story.Workspace != scope.WorkspaceID || story.Team != expectedTeam.ID {
		return nil, fmt.Errorf("%w: story reference resolved outside the authorized team", ErrTeamNotAccessible)
	}
	if story.DeletedAt != nil || story.ArchivedAt != nil {
		return nil, errors.New("get story: story is deleted or archived")
	}

	_, statusesByID, err := e.scopedStatuses(ctx, scope, joinedByID)
	if err != nil {
		return nil, err
	}
	var statusName string
	var statusCategory string
	if story.Status != nil {
		if status, visible := statusesByID[*story.Status]; visible {
			statusName = status.Name
			statusCategory = status.Category
		}
	}

	var assigneeName string
	var assigneeUsername string
	if story.Assignee != nil {
		member, memberErr := e.activeTeamMemberByID(ctx, scope.WorkspaceID, expectedTeam.ID, *story.Assignee)
		if memberErr != nil {
			return nil, memberErr
		}
		if member != nil {
			assigneeName = memberDisplayName(*member)
			assigneeUsername = strings.TrimSpace(member.Username)
		}
	}
	description, descriptionTruncated := boundedOptionalString(story.Description, maxStoryDescriptionRunes)
	var sprintName *string
	if story.SprintSummary != nil {
		name := strings.TrimSpace(story.SprintSummary.Name)
		if name != "" {
			sprintName = &name
		}
	}

	return marshalToolResult(storyDetailsResult{
		ID:                       story.ID,
		Reference:                storyReference,
		URL:                      storyURL(scope, storyReference),
		Title:                    story.Title,
		Description:              description,
		DescriptionTruncated:     descriptionTruncated,
		TeamID:                   expectedTeam.ID,
		TeamName:                 expectedTeam.Name,
		TeamCode:                 strings.ToUpper(strings.TrimSpace(expectedTeam.Code)),
		StatusID:                 story.Status,
		StatusName:               statusName,
		StatusCategory:           statusCategory,
		AssigneeID:               story.Assignee,
		AssigneeName:             assigneeName,
		AssigneeUsername:         assigneeUsername,
		Priority:                 story.Priority,
		EstimateLabel:            story.EstimateLabel,
		EstimateValue:            story.EstimateValue,
		EstimateScheme:           story.EstimateScheme,
		EstimatedDurationMinutes: story.EstimatedDurationMinutes,
		MinimumFocusBlockMinutes: story.MinimumFocusBlockMinutes,
		AutoSchedulingEnabled:    story.AutoSchedulingEnabled,
		AutoSchedulingLocked:     story.AutoSchedulingLocked,
		AutoSchedulingStatus:     story.AutoSchedulingStatus,
		AutoSchedulingReason:     story.AutoSchedulingReason,
		AutoSchedulingUpdatedAt:  story.AutoSchedulingUpdatedAt,
		SprintName:               sprintName,
		StartDate:                story.StartDate,
		EndDate:                  story.EndDate,
		CompletedAt:              story.CompletedAt,
		UpdatedAt:                story.UpdatedAt,
	})
}

func (e *FortyOneToolExecutor) scopedStatuses(
	ctx context.Context,
	scope ToolScope,
	joinedByID map[uuid.UUID]messagingTeam,
) ([]messagingState, map[uuid.UUID]messagingState, error) {
	items, err := e.states.List(ctx, scope.WorkspaceID, scope.UserID)
	if err != nil {
		return nil, nil, fmt.Errorf("list statuses: %w", err)
	}
	visible := make([]messagingState, 0, len(items))
	byID := make(map[uuid.UUID]messagingState, len(items))
	for _, status := range items {
		if status.ID == uuid.Nil || status.Workspace != scope.WorkspaceID {
			continue
		}
		if _, allowed := joinedByID[status.Team]; !allowed {
			continue
		}
		if _, duplicate := byID[status.ID]; duplicate {
			continue
		}
		status.Name = strings.TrimSpace(status.Name)
		status.Category = strings.TrimSpace(status.Category)
		if status.Name == "" {
			continue
		}
		visible = append(visible, status)
		byID[status.ID] = status
	}
	return visible, byID, nil
}

func (e *FortyOneToolExecutor) activeTeamMemberByID(
	ctx context.Context,
	workspaceID, teamID, memberID uuid.UUID,
) (*messagingUser, error) {
	items, err := e.users.List(ctx, workspaceID, messagingUserListFilter{TeamID: &teamID})
	if err != nil {
		return nil, fmt.Errorf("load story assignee: %w", err)
	}
	for _, member := range items {
		if member.ID == memberID && member.IsActive && !member.IsSystem {
			copy := member
			return &copy, nil
		}
	}
	return nil, nil
}

func accessibleStoryReference(raw string, joinedByID map[uuid.UUID]messagingTeam) (string, messagingTeam, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" || len([]rune(trimmed)) > maxStoryReferenceRunes {
		return "", messagingTeam{}, fmt.Errorf("%w: story_reference must contain 1-%d characters", ErrInvalidToolArguments, maxStoryReferenceRunes)
	}
	compact := strings.ToUpper(strings.ReplaceAll(strings.ReplaceAll(trimmed, " ", ""), "-", ""))
	digitIndex := -1
	for index, char := range compact {
		if char >= '0' && char <= '9' {
			digitIndex = index
			break
		}
	}
	if digitIndex < 1 || digitIndex == len(compact) {
		return "", messagingTeam{}, fmt.Errorf("%w: story_reference must look like WEB-123", ErrInvalidToolArguments)
	}
	teamCode := compact[:digitIndex]
	sequenceText := compact[digitIndex:]
	for _, char := range sequenceText {
		if char < '0' || char > '9' {
			return "", messagingTeam{}, fmt.Errorf("%w: story_reference must look like WEB-123", ErrInvalidToolArguments)
		}
	}
	sequenceID, err := strconv.Atoi(sequenceText)
	if err != nil || sequenceID < 1 {
		return "", messagingTeam{}, fmt.Errorf("%w: story_reference must contain a positive sequence number", ErrInvalidToolArguments)
	}
	for _, team := range joinedByID {
		normalizedCode := strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(team.Code), "-", ""))
		if normalizedCode == teamCode {
			return fmt.Sprintf("%s-%d", strings.ToUpper(strings.TrimSpace(team.Code)), sequenceID), team, nil
		}
	}
	return "", messagingTeam{}, fmt.Errorf("%w: team code %s", ErrTeamNotAccessible, teamCode)
}

func statusIsClosed(category string) bool {
	switch strings.ToLower(strings.TrimSpace(category)) {
	case "completed", "cancelled":
		return true
	default:
		return false
	}
}

func memberDisplayName(member messagingUser) string {
	if displayName := strings.TrimSpace(member.FullName); displayName != "" {
		return displayName
	}
	return strings.TrimSpace(member.Username)
}

func memberRoleTitle(member messagingUser) string {
	if roleTitle := strings.TrimSpace(member.TeamAIRoleTitle); roleTitle != "" {
		return roleTitle
	}
	return strings.TrimSpace(member.InferredTeamAIRoleTitle)
}

func boundedOptionalString(value *string, maximumRunes int) (*string, bool) {
	if value == nil {
		return nil, false
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil, false
	}
	runes := []rune(trimmed)
	if len(runes) <= maximumRunes {
		return &trimmed, false
	}
	truncated := string(runes[:maximumRunes])
	return &truncated, true
}
