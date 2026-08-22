package agentreadinesshttp

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	keyresults "github.com/complexus-tech/projects-api/internal/modules/keyresults/service"
	reports "github.com/complexus-tech/projects-api/internal/modules/reports/service"
	teams "github.com/complexus-tech/projects-api/internal/modules/teams/service"
	"github.com/google/uuid"
	mcpauth "github.com/modelcontextprotocol/go-sdk/auth"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func (h *Handler) addTools(server *mcp.Server) {
	read, write := annotations(true, true), annotations(false, false)
	archive := annotations(false, true)
	destructive := true
	archive.DestructiveHint = &destructive
	addSafeTool(h, server, tool("list_workspaces", "List FortyOne workspaces", "Lists a page of organizations or workspaces the connected user can access. Use this first when the user has not identified a workspace.", read), h.listWorkspaces)
	addSafeTool(h, server, tool("list_teams", "List workspace teams", "Lists a page of accessible teams or groups and their codes. Use joinedOnly for teams the connected user belongs to.", read), h.listTeams)
	addSafeTool(h, server, tool("list_story_statuses", "List story workflow statuses", "Lists a page of statuses for a team's stories, tasks, issues, tickets, or work items. Use this to resolve requests such as mark done, start, cancel, pause, or move to backlog before calling update_story.", read), h.listStoryStatuses)
	addSafeTool(h, server, tool("list_stories", "List stories, tasks, or issues", "Lists a page of stories, tasks, issues, tickets, or work items with scheduling, ownership, sprint, objective, key-result, status, and updated_at data. Use search to resolve work by title. For 'my work', set assignedToMe. For work due today or on another day, resolve that day to YYYY-MM-DD and set dueOn.", read), h.listStories)
	addSafeTool(h, server, tool("create_story", "Create a story, task, or issue", "Creates a user-approved story, task, issue, ticket, or work item. Preserve requested duration, focus block, auto-scheduling, dates, estimate, labels, parent, sprint, objective, and key-result links. Generate an opaque idempotencyKey for each approved creation and reuse it only when retrying that same call. Never call until the user approves the final values.", write), h.createStory)
	addSafeTool(h, server, tool("update_story", "Update a story, task, or issue", "Updates user-approved fields on an existing story, task, issue, ticket, or work item. Read the story first and pass its updated_at as expectedUpdatedAt. Dates may be cleared with an empty string. Duration is expressed in minutes. Never call until the user approves the exact changes.", write), h.updateStory)
	addSafeTool(h, server, tool("list_story_comments", "List story comments", "Lists a page of comments and replies on a story, task, issue, ticket, or work item.", read), h.listStoryComments)
	addSafeTool(h, server, tool("add_story_comment", "Add a story comment", "Adds a user-approved comment or reply to a story, task, issue, ticket, or work item. Never call until the user approves the exact comment.", write), h.addStoryComment)
	addSafeTool(h, server, tool("set_story_archived", "Archive or restore a story", "Archives or restores a user-approved story, task, issue, ticket, or work item. Use archived=true for requests to remove or archive work, preserving it for recovery. Never call until the user approves the action.", archive), h.setStoryArchived)
	addSafeTool(h, server, tool("list_sprints", "List sprints or iterations", "Lists a page of sprints, iterations, or delivery cycles, optionally filtered by team or search text.", read), h.listSprints)
	addSafeTool(h, server, tool("create_sprint", "Create a sprint or iteration", "Creates a user-approved sprint, iteration, or delivery cycle with a goal and date range.", write), h.createSprint)
	addSafeTool(h, server, tool("analyze_sprint", "Analyze a sprint or iteration", "Analyzes a sprint, iteration, or cycle and returns completion, status, story breakdown, and a page of burndown and team-allocation details.", read), h.analyzeSprint)
	addSafeTool(h, server, tool("list_objectives", "List objectives, projects, or goals", "Lists a page of objectives, projects, or goals with delivery forecast and linked-work progress, optionally filtered by team or search text.", read), h.listObjectives)
	addSafeTool(h, server, tool("list_objective_statuses", "List objective workflow statuses", "Lists a page of workspace objective statuses. Use this to resolve requested project, goal, or objective status changes before calling update_objective.", read), h.listObjectiveStatuses)
	addSafeTool(h, server, tool("create_objective", "Create an objective, project, or goal", "Creates a user-approved objective when the user asks for an objective, project, or goal, applying the workspace default status when omitted.", write), h.createObjective)
	addSafeTool(h, server, tool("update_objective", "Update an objective, project, or goal", "Updates user-approved fields on an existing objective, project, or goal. Read it first and pass its UpdatedAt as expectedUpdatedAt. Never call until the user approves the exact changes.", write), h.updateObjective)
	addSafeTool(h, server, tool("analyze_objective", "Analyze an objective, project, or goal", "Analyzes an objective, project, or goal and returns a page of progress, priority, allocation, and chart details.", read), h.analyzeObjective)
	addSafeTool(h, server, tool("list_key_results", "List key results or targets", "Lists a page of measurable key results, KRs, outcomes, measures, or targets for accessible objectives.", read), h.listKeyResults)
	addSafeTool(h, server, tool("create_key_result", "Create a key result or target", "Creates a user-approved percentage, number, or boolean key result, KR, outcome, measure, or target for an objective.", write), h.createKeyResult)
	addSafeTool(h, server, tool("update_key_result", "Update a key result or target", "Updates user-approved fields on an existing key result, KR, outcome, measure, or target. Read it first and pass its UpdatedAt as expectedUpdatedAt. Never call until the user approves the exact changes.", write), h.updateKeyResult)
	addSafeTool(h, server, tool("analyze_work", "Analyze work and delivery", "Analyzes workspace work, workload, delivery status, requests, and risks across stories, tasks, issues, sprints, objectives, projects, and goals for a date range. Member, team, and risk details are paginated.", read), h.analyzeWork)
}

func tool(name, title, description string, a *mcp.ToolAnnotations) *mcp.Tool {
	return &mcp.Tool{Name: name, Title: title, Description: description, Annotations: a}
}
func annotations(readOnly, idempotent bool) *mcp.ToolAnnotations {
	destructive, openWorld := false, false
	return &mcp.ToolAnnotations{ReadOnlyHint: readOnly, IdempotentHint: idempotent, DestructiveHint: &destructive, OpenWorldHint: &openWorld}
}

func (h *Handler) listWorkspaces(ctx context.Context, _ *mcp.CallToolRequest, in workspaceListInput) (*mcp.CallToolResult, any, error) {
	userID, err := mcpUserID(ctx)
	if err != nil {
		return nil, nil, err
	}
	items, err := h.cfg.Workspaces.List(ctx, userID)
	if err != nil {
		return nil, nil, err
	}
	page, pageSize, offset, limit := normalizePagination(in.Page, in.PageSize)
	items, hasMore := pageSlice(items, offset, limit, pageSize)
	out := make([]map[string]any, 0, len(items))
	for _, item := range items {
		out = append(out, map[string]any{"id": item.ID, "slug": item.Slug, "name": item.Name, "role": item.UserRole})
	}
	return nil, paginatedResult("workspaces", out, page, pageSize, hasMore), nil
}

func (h *Handler) listTeams(ctx context.Context, _ *mcp.CallToolRequest, in teamListInput) (*mcp.CallToolResult, any, error) {
	workspaceID, userID, err := h.authorizeWorkspace(ctx, in.WorkspaceID)
	if err != nil {
		return nil, nil, err
	}
	page, pageSize, offset, limit := normalizePagination(in.Page, in.PageSize)
	items, err := h.cfg.Teams.List(ctx, workspaceID, userID, teams.CoreListTeamsFilter{JoinedOnly: in.JoinedOnly, Search: strings.TrimSpace(in.Search), Limit: limit, Offset: offset})
	if err != nil {
		return nil, nil, err
	}
	items, hasMore := trimPage(items, pageSize)
	out := make([]map[string]any, 0, len(items))
	for _, item := range items {
		out = append(out, map[string]any{"id": item.ID, "name": item.Name, "code": item.Code, "sprintsEnabled": item.SprintsEnabled})
	}
	return nil, paginatedResult("teams", out, page, pageSize, hasMore), nil
}

func (h *Handler) listStoryStatuses(ctx context.Context, _ *mcp.CallToolRequest, in storyStatusListInput) (*mcp.CallToolResult, any, error) {
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
	items, err := h.cfg.States.TeamList(ctx, workspaceID, teamID)
	if err != nil {
		return nil, nil, err
	}
	page, pageSize, offset, limit := normalizePagination(in.Page, in.PageSize)
	items, hasMore := pageSlice(items, offset, limit, pageSize)
	out := make([]map[string]any, 0, len(items))
	for _, item := range items {
		out = append(out, map[string]any{"id": item.ID, "name": item.Name, "category": item.Category, "orderIndex": item.OrderIndex, "isDefault": item.IsDefault, "color": item.Color})
	}
	return nil, paginatedResult("statuses", out, page, pageSize, hasMore), nil
}

func (h *Handler) listStories(ctx context.Context, _ *mcp.CallToolRequest, in storyListInput) (*mcp.CallToolResult, any, error) {
	workspaceID, userID, err := h.authorizeWorkspace(ctx, in.WorkspaceID)
	if err != nil {
		return nil, nil, err
	}
	filters, err := storyListFilters(in, userID)
	if err != nil {
		return nil, nil, err
	}
	page, pageSize, offset, limit := normalizePagination(in.Page, in.PageSize)
	filters["limit"] = limit
	filters["offset"] = offset
	items, err := h.cfg.Stories.List(ctx, workspaceID, filters)
	if err != nil {
		return nil, nil, err
	}
	items, hasMore := trimPage(items, pageSize)
	return nil, paginatedResult("stories", items, page, pageSize, hasMore), nil
}

func (h *Handler) listStoryComments(ctx context.Context, _ *mcp.CallToolRequest, in storyCommentListInput) (*mcp.CallToolResult, any, error) {
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
	page, pageSize, _, _ := normalizePagination(in.Page, in.PageSize)
	items, hasMore, err := h.cfg.Stories.GetComments(ctx, storyID, page, pageSize)
	if err != nil {
		return nil, nil, err
	}
	return nil, paginatedResult("comments", items, page, pageSize, hasMore), nil
}

func (h *Handler) listSprints(ctx context.Context, _ *mcp.CallToolRequest, in objectiveListInput) (*mcp.CallToolResult, any, error) {
	workspaceID, userID, err := h.authorizeWorkspace(ctx, in.WorkspaceID)
	if err != nil {
		return nil, nil, err
	}
	filters, err := listFilters(in.TeamID, in.Search)
	if err != nil {
		return nil, nil, err
	}
	page, pageSize, offset, limit := normalizePagination(in.Page, in.PageSize)
	filters["limit"] = limit
	filters["offset"] = offset
	items, err := h.cfg.Sprints.List(ctx, workspaceID, userID, filters)
	if err != nil {
		return nil, nil, err
	}
	items, hasMore := trimPage(items, pageSize)
	out := make([]map[string]any, 0, len(items))
	for _, item := range items {
		out = append(out, sprintToolResult(item))
	}
	return nil, paginatedResult("sprints", out, page, pageSize, hasMore), nil
}

func (h *Handler) analyzeSprint(ctx context.Context, _ *mcp.CallToolRequest, in entityInput) (*mcp.CallToolResult, any, error) {
	workspaceID, _, err := h.authorizeWorkspace(ctx, in.WorkspaceID)
	if err != nil {
		return nil, nil, err
	}
	id, err := parseRequiredUUID("id", in.ID)
	if err != nil {
		return nil, nil, err
	}
	result, err := h.cfg.Sprints.GetAnalytics(ctx, id, workspaceID)
	if err != nil {
		return nil, nil, err
	}
	page, pageSize, offset, limit := normalizePagination(in.Page, in.PageSize)
	var burndownHasMore, allocationHasMore bool
	result.Burndown, burndownHasMore = pageSlice(result.Burndown, offset, limit, pageSize)
	result.TeamAllocation, allocationHasMore = pageSlice(result.TeamAllocation, offset, limit, pageSize)
	return nil, analysisResult(result, page, pageSize, burndownHasMore || allocationHasMore), nil
}

func (h *Handler) listObjectives(ctx context.Context, _ *mcp.CallToolRequest, in objectiveListInput) (*mcp.CallToolResult, any, error) {
	workspaceID, userID, err := h.authorizeWorkspace(ctx, in.WorkspaceID)
	if err != nil {
		return nil, nil, err
	}
	filters, err := listFilters(in.TeamID, in.Search)
	if err != nil {
		return nil, nil, err
	}
	page, pageSize, offset, limit := normalizePagination(in.Page, in.PageSize)
	filters["limit"] = limit
	filters["offset"] = offset
	items, err := h.cfg.Objectives.List(ctx, workspaceID, userID, filters)
	if err != nil {
		return nil, nil, err
	}
	items, hasMore := trimPage(items, pageSize)
	out := make([]map[string]any, 0, len(items))
	for _, item := range items {
		out = append(out, objectiveToolResult(item))
	}
	return nil, paginatedResult("objectives", out, page, pageSize, hasMore), nil
}

func (h *Handler) listObjectiveStatuses(ctx context.Context, _ *mcp.CallToolRequest, in objectiveStatusListInput) (*mcp.CallToolResult, any, error) {
	workspaceID, _, err := h.authorizeWorkspace(ctx, in.WorkspaceID)
	if err != nil {
		return nil, nil, err
	}
	items, err := h.cfg.ObjectiveStatuses.List(ctx, workspaceID)
	if err != nil {
		return nil, nil, err
	}
	page, pageSize, offset, limit := normalizePagination(in.Page, in.PageSize)
	items, hasMore := pageSlice(items, offset, limit, pageSize)
	out := make([]map[string]any, 0, len(items))
	for _, item := range items {
		out = append(out, map[string]any{"id": item.ID, "name": item.Name, "category": item.Category, "orderIndex": item.OrderIndex, "isDefault": item.IsDefault, "color": item.Color})
	}
	return nil, paginatedResult("statuses", out, page, pageSize, hasMore), nil
}

func (h *Handler) analyzeObjective(ctx context.Context, _ *mcp.CallToolRequest, in entityInput) (*mcp.CallToolResult, any, error) {
	workspaceID, _, err := h.authorizeWorkspace(ctx, in.WorkspaceID)
	if err != nil {
		return nil, nil, err
	}
	id, err := parseRequiredUUID("id", in.ID)
	if err != nil {
		return nil, nil, err
	}
	result, err := h.cfg.Objectives.GetAnalytics(ctx, id, workspaceID)
	if err != nil {
		return nil, nil, err
	}
	page, pageSize, offset, limit := normalizePagination(in.Page, in.PageSize)
	var priorityHasMore, allocationHasMore, progressHasMore bool
	result.PriorityBreakdown, priorityHasMore = pageSlice(result.PriorityBreakdown, offset, limit, pageSize)
	result.TeamAllocation, allocationHasMore = pageSlice(result.TeamAllocation, offset, limit, pageSize)
	result.ProgressChart, progressHasMore = pageSlice(result.ProgressChart, offset, limit, pageSize)
	return nil, analysisResult(result, page, pageSize, priorityHasMore || allocationHasMore || progressHasMore), nil
}

func (h *Handler) listKeyResults(ctx context.Context, _ *mcp.CallToolRequest, in keyResultListInput) (*mcp.CallToolResult, any, error) {
	workspaceID, userID, err := h.authorizeWorkspace(ctx, in.WorkspaceID)
	if err != nil {
		return nil, nil, err
	}
	objectiveIDs, err := parseUUIDs(in.ObjectiveIDs)
	if err != nil {
		return nil, nil, err
	}
	page, pageSize, _, _ := normalizePagination(in.Page, in.PageSize)
	result, err := h.cfg.KeyResults.ListPaginated(ctx, keyresults.CoreKeyResultFilters{WorkspaceID: workspaceID, CurrentUserID: userID, ObjectiveIDs: objectiveIDs, Page: page, PageSize: pageSize})
	if err != nil {
		return nil, nil, err
	}
	out := make([]map[string]any, 0, len(result.KeyResults))
	for _, item := range result.KeyResults {
		out = append(out, keyResultWithObjectiveToolResult(item))
	}
	return nil, map[string]any{"keyResults": out, "totalCount": result.TotalCount, "page": result.Page, "pageSize": result.PageSize, "hasMore": result.HasMore}, nil
}

func (h *Handler) analyzeWork(ctx context.Context, _ *mcp.CallToolRequest, in analysisInput) (*mcp.CallToolResult, any, error) {
	workspaceID, _, err := h.authorizeWorkspace(ctx, in.WorkspaceID)
	if err != nil {
		return nil, nil, err
	}
	teamIDs, err := parseUUIDs(in.TeamIDs)
	if err != nil {
		return nil, nil, err
	}
	assigneeIDs, err := parseUUIDs(in.AssigneeIDs)
	if err != nil {
		return nil, nil, err
	}
	sprintIDs, err := parseUUIDs(in.SprintIDs)
	if err != nil {
		return nil, nil, err
	}
	objectiveIDs, err := parseUUIDs(in.ObjectiveIDs)
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
	result, err := h.cfg.Reports.GetPulseReport(ctx, workspaceID, reports.ReportFilters{TeamIDs: teamIDs, AssigneeIDs: assigneeIDs, SprintIDs: sprintIDs, ObjectiveIDs: objectiveIDs, StartDate: startDate, EndDate: endDate})
	if err != nil {
		return nil, nil, err
	}
	page, pageSize, offset, limit := normalizePagination(in.Page, in.PageSize)
	hasMore := false
	result.Workload.Members, hasMore = pageSlice(result.Workload.Members, offset, limit, pageSize)
	result.Workload.Teams, hasMore = pageSliceWithMore(result.Workload.Teams, offset, limit, pageSize, hasMore)
	result.Workload.Risks.OverloadedMembers, hasMore = pageSliceWithMore(result.Workload.Risks.OverloadedMembers, offset, limit, pageSize, hasMore)
	result.Workload.Risks.OverdueMembers, hasMore = pageSliceWithMore(result.Workload.Risks.OverdueMembers, offset, limit, pageSize, hasMore)
	result.Risks, hasMore = pageSliceWithMore(result.Risks, offset, limit, pageSize, hasMore)
	response := analysisResult(result, page, pageSize, hasMore)
	response["guidance"] = "Treat counts and risks as current FortyOne data. Ask before changing work, ownership, dates, or scope."
	return nil, response, nil
}

func (h *Handler) authorizeWorkspace(ctx context.Context, raw string) (uuid.UUID, uuid.UUID, error) {
	userID, err := mcpUserID(ctx)
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}
	workspaceID, err := parseRequiredUUID("workspaceId", raw)
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}
	if _, err := h.cfg.Workspaces.Get(ctx, workspaceID, userID); err != nil {
		return uuid.Nil, uuid.Nil, err
	}
	return workspaceID, userID, nil
}
func mcpUserID(ctx context.Context) (uuid.UUID, error) {
	info := mcpauth.TokenInfoFromContext(ctx)
	if info == nil {
		return uuid.Nil, errors.New("MCP authentication context is missing")
	}
	id, err := uuid.Parse(info.UserID)
	if err != nil {
		return uuid.Nil, errors.New("MCP user identity is invalid")
	}
	return id, nil
}
func listFilters(teamID, search string) (map[string]any, error) {
	result := map[string]any{}
	if teamID != "" {
		id, err := parseRequiredUUID("teamId", teamID)
		if err != nil {
			return nil, err
		}
		result["team_id"] = id
	}
	if strings.TrimSpace(search) != "" {
		result["search"] = strings.TrimSpace(search)
	}
	return result, nil
}
func parseRequiredUUID(name, raw string) (uuid.UUID, error) {
	id, err := uuid.Parse(strings.TrimSpace(raw))
	if err != nil {
		return uuid.Nil, invalidToolInputf("%s must be a valid UUID", name)
	}
	return id, nil
}
func optionalUUID(raw string) (*uuid.UUID, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, nil
	}
	id, err := uuid.Parse(raw)
	if err != nil {
		return nil, invalidToolInput("must be a valid UUID")
	}
	return &id, nil
}
func parseUUIDs(values []string) ([]uuid.UUID, error) {
	result := make([]uuid.UUID, 0, len(values))
	for _, value := range values {
		id, err := uuid.Parse(value)
		if err != nil {
			return nil, invalidToolInputf("%q is not a valid UUID", value)
		}
		result = append(result, id)
	}
	return result, nil
}
func storyListFilters(input storyListInput, userID uuid.UUID) (map[string]any, error) {
	filters := map[string]any{"current_user_id": userID}
	if search := strings.TrimSpace(input.Search); search != "" {
		filters["title_contains"] = search
	}
	if input.AssignedToMe {
		filters["assigned_to_me"] = true
	}
	if strings.TrimSpace(input.DueOn) != "" {
		dueOn, err := optionalDate(input.DueOn)
		if err != nil {
			return nil, fmt.Errorf("dueOn: %w", err)
		}
		filters["deadline_after"] = *dueOn
		filters["deadline_before"] = *dueOn
	}
	fields := []struct {
		name     string
		value    string
		queryKey string
		plural   bool
	}{
		{name: "teamId", value: input.TeamID, queryKey: "team_ids", plural: true},
		{name: "sprintId", value: input.SprintID, queryKey: "sprint_ids", plural: true},
		{name: "objectiveId", value: input.ObjectiveID, queryKey: "objective_id"},
		{name: "assigneeId", value: input.AssigneeID, queryKey: "assignee_ids", plural: true},
		{name: "statusId", value: input.StatusID, queryKey: "status_ids", plural: true},
		{name: "keyResultId", value: input.KeyResultID, queryKey: "key_result_id"},
	}

	for _, field := range fields {
		if strings.TrimSpace(field.value) == "" {
			continue
		}
		id, err := parseRequiredUUID(field.name, field.value)
		if err != nil {
			return nil, err
		}
		if field.plural {
			filters[field.queryKey] = []uuid.UUID{id}
		} else {
			filters[field.queryKey] = id
		}
	}

	return filters, nil
}

const (
	defaultMCPPageSize = 25
	maximumMCPPageSize = 100
)

func normalizePagination(rawPage, rawPageSize int) (page, pageSize, offset, limit int) {
	page = rawPage
	if page < 1 {
		page = 1
	}
	pageSize = rawPageSize
	if pageSize < 1 {
		pageSize = defaultMCPPageSize
	}
	if pageSize > maximumMCPPageSize {
		pageSize = maximumMCPPageSize
	}
	limit = pageSize + 1
	maximumInt := int(^uint(0) >> 1)
	maximumPage := ((maximumInt - limit) / pageSize) + 1
	if page > maximumPage {
		page = maximumPage
	}
	offset = (page - 1) * pageSize
	return page, pageSize, offset, limit
}

func trimPage[T any](items []T, pageSize int) ([]T, bool) {
	hasMore := len(items) > pageSize
	if hasMore {
		items = items[:pageSize]
	}
	return items, hasMore
}

func pageSlice[T any](items []T, offset, limit, pageSize int) ([]T, bool) {
	if offset >= len(items) {
		return []T{}, false
	}
	end := min(offset+limit, len(items))
	return trimPage(items[offset:end], pageSize)
}

func pageSliceWithMore[T any](items []T, offset, limit, pageSize int, previousHasMore bool) ([]T, bool) {
	pageItems, hasMore := pageSlice(items, offset, limit, pageSize)
	return pageItems, previousHasMore || hasMore
}

func paginatedResult(name string, items any, page, pageSize int, hasMore bool) map[string]any {
	return map[string]any{name: items, "page": page, "pageSize": pageSize, "hasMore": hasMore}
}

func analysisResult(result any, page, pageSize int, hasMore bool) map[string]any {
	return map[string]any{"analysis": result, "page": page, "pageSize": pageSize, "hasMore": hasMore}
}

func optionalDate(raw string) (*time.Time, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, nil
	}
	value, err := time.Parse("2006-01-02", raw)
	if err != nil {
		return nil, invalidToolInput("date must use YYYY-MM-DD")
	}
	return &value, nil
}
func requiredDate(name, raw string) (time.Time, error) {
	value, err := optionalDate(raw)
	if err != nil {
		return time.Time{}, fmt.Errorf("%s: %w", name, err)
	}
	if value == nil {
		return time.Time{}, invalidToolInputf("%s is required", name)
	}
	return *value, nil
}

func requiredTimestamp(name, raw string) (time.Time, error) {
	value, err := time.Parse(time.RFC3339Nano, strings.TrimSpace(raw))
	if err != nil {
		return time.Time{}, invalidToolInputf("%s must use RFC3339", name)
	}
	return value.UTC(), nil
}
func optionalString(value string) *string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return &value
}
