package agentreadinesshttp

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"strings"
	"testing"

	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

func TestToolErrorsDoNotExposeBackendInternals(t *testing.T) {
	t.Parallel()
	var logs bytes.Buffer
	handler := &Handler{cfg: Config{Log: logger.NewWithJSON(&logs, slog.LevelDebug, "agent-readiness-test")}}
	rawError := errors.New(`ERROR: column reference "workspace_id" is ambiguous (SQLSTATE 42702)`)
	wrapped := safeToolHandler(handler, "list_stories", func(context.Context, *mcp.CallToolRequest, emptyInput) (*mcp.CallToolResult, any, error) {
		return nil, nil, rawError
	})

	_, _, clientErr := wrapped(context.Background(), nil, emptyInput{})
	require.Error(t, clientErr)
	require.Contains(t, clientErr.Error(), "FortyOne couldn't complete this request right now")
	require.Contains(t, clientErr.Error(), "Reference:")
	for _, internalDetail := range []string{"workspace_id", "ambiguous", "SQLSTATE", "42702", "column reference"} {
		require.NotContains(t, clientErr.Error(), internalDetail)
	}

	logOutput := logs.String()
	require.Contains(t, logOutput, `"msg":"MCP tool failed"`)
	require.Contains(t, logOutput, `"tool":"list_stories"`)
	require.Contains(t, logOutput, "workspace_id")
	require.Contains(t, logOutput, "SQLSTATE 42702")
}

func TestToolInputErrorsRemainActionableAndAreLogged(t *testing.T) {
	t.Parallel()
	var logs bytes.Buffer
	handler := &Handler{cfg: Config{Log: logger.NewWithJSON(&logs, slog.LevelDebug, "agent-readiness-test")}}
	wrapped := safeToolHandler(handler, "create_story", func(context.Context, *mcp.CallToolRequest, emptyInput) (*mcp.CallToolResult, any, error) {
		return nil, nil, invalidToolInput("title is required")
	})

	_, _, clientErr := wrapped(context.Background(), nil, emptyInput{})
	require.EqualError(t, clientErr, "title is required")
	require.Contains(t, logs.String(), `"msg":"MCP tool request rejected"`)
	require.Contains(t, logs.String(), `"tool":"create_story"`)
}

func TestMCPWireResultDoesNotExposeBackendInternals(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	handler := &Handler{}
	server := mcp.NewServer(&mcp.Implementation{Name: "safe-error-test", Version: "1.0.0"}, nil)
	definition := &mcp.Tool{Name: "failing_tool", Title: "Fail safely", Description: "Test tool.", Annotations: annotations(true, true), OutputSchema: objectValueOutput("Test output")}
	addSafeTool(handler, server, definition, func(context.Context, *mcp.CallToolRequest, emptyInput) (*mcp.CallToolResult, any, error) {
		return nil, nil, errors.New(`ERROR: relation "private_table" does not exist (SQLSTATE 42P01)`)
	})

	serverTransport, clientTransport := mcp.NewInMemoryTransports()
	serverSession, err := server.Connect(ctx, serverTransport, nil)
	require.NoError(t, err)
	t.Cleanup(func() { _ = serverSession.Close() })
	client := mcp.NewClient(&mcp.Implementation{Name: "safe-error-client", Version: "1.0.0"}, nil)
	clientSession, err := client.Connect(ctx, clientTransport, nil)
	require.NoError(t, err)
	t.Cleanup(func() { _ = clientSession.Close() })

	result, err := clientSession.CallTool(ctx, &mcp.CallToolParams{Name: "failing_tool", Arguments: map[string]any{}})
	require.NoError(t, err)
	require.True(t, result.IsError)
	require.Len(t, result.Content, 1)
	content, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok, "unexpected MCP error content type %T", result.Content[0])
	require.Contains(t, content.Text, "FortyOne couldn't complete this request right now")
	require.Contains(t, content.Text, "Reference:")
	for _, internalDetail := range []string{"private_table", "relation", "SQLSTATE", "42P01"} {
		require.NotContains(t, content.Text, internalDetail)
	}
}

func TestRegisteredToolsCoverNaturalLanguageScenarios(t *testing.T) {
	t.Parallel()
	tools := listToolsForTest(t)

	scenarios := []struct {
		question string
		tool     string
		cue      string
	}{
		{question: "Which FortyOne workspaces or organizations can I access?", tool: "list_workspaces", cue: "organizations"},
		{question: "Which teams or groups do I belong to?", tool: "list_teams", cue: "groups"},
		{question: "Which statuses can I move this task to?", tool: "list_story_statuses", cue: "mark done"},
		{question: "Show me my work in the Art Circles team.", tool: "list_stories", cue: "my work"},
		{question: "Show my tasks and issues due today.", tool: "list_stories", cue: "due today"},
		{question: "Create a 30-minute task and auto-schedule it.", tool: "create_story", cue: "task"},
		{question: "Create an issue in the current sprint.", tool: "create_story", cue: "issue"},
		{question: "Change this task's due date to August 31 and duration to six hours.", tool: "update_story", cue: "duration"},
		{question: "Show me the comments on this task.", tool: "list_story_comments", cue: "comments"},
		{question: "Add this comment to the issue.", tool: "add_story_comment", cue: "comment"},
		{question: "Archive this task so we can restore it later.", tool: "set_story_archived", cue: "recovery"},
		{question: "Show the team's sprints, iterations, or cycles.", tool: "list_sprints", cue: "cycles"},
		{question: "Create a two-week iteration for this team.", tool: "create_sprint", cue: "iteration"},
		{question: "How is this sprint going?", tool: "analyze_sprint", cue: "burndown"},
		{question: "Show our projects, goals, and objectives.", tool: "list_objectives", cue: "projects"},
		{question: "Which statuses can this goal use?", tool: "list_objective_statuses", cue: "status"},
		{question: "Create a project or goal for the launch.", tool: "create_objective", cue: "project"},
		{question: "Mark this goal as at risk and move its target date.", tool: "update_objective", cue: "updates"},
		{question: "Is this goal on track?", tool: "analyze_objective", cue: "goal"},
		{question: "Show the KRs, outcomes, measures, or targets for this goal.", tool: "list_key_results", cue: "targets"},
		{question: "Create a percentage target for this objective.", tool: "create_key_result", cue: "target"},
		{question: "Update this KR's current value.", tool: "update_key_result", cue: "updates"},
		{question: "Analyze our workload, delivery risks, and status this month.", tool: "analyze_work", cue: "risks"},
	}

	covered := make(map[string]bool)
	for _, scenario := range scenarios {
		t.Run(scenario.question, func(t *testing.T) {
			definition, ok := tools[scenario.tool]
			require.True(t, ok, "expected tool %q to be registered", scenario.tool)
			require.Contains(t, strings.ToLower(definition.Description), strings.ToLower(scenario.cue))
			covered[scenario.tool] = true
		})
	}

	require.Len(t, tools, 21)
	require.Len(t, covered, 21, "every registered tool must have at least one natural-language scenario")
	requireToolProperties(t, tools["list_workspaces"], "page", "pageSize")
	requireToolProperties(t, tools["list_teams"], "workspaceId", "search", "page", "pageSize")
	requireToolProperties(t, tools["list_story_statuses"], "workspaceId", "teamId", "page", "pageSize")
	requireToolProperties(t, tools["list_stories"], "workspaceId", "teamId", "assignedToMe", "dueOn", "search", "page", "pageSize")
	requireToolProperties(t, tools["create_story"], "title", "estimatedDurationMinutes", "minimumFocusBlockMinutes", "autoSchedulingEnabled", "idempotencyKey", "confirmed")
	requireToolRequiredProperties(t, tools["create_story"], "workspaceId", "teamId", "title", "idempotencyKey", "confirmed")
	requireToolProperties(t, tools["update_story"], "id", "expectedUpdatedAt", "endDate", "estimatedDurationMinutes", "confirmed")
	requireToolProperties(t, tools["list_story_comments"], "workspaceId", "storyId", "page", "pageSize")
	requireToolProperties(t, tools["add_story_comment"], "workspaceId", "storyId", "comment", "parentId", "confirmed")
	requireToolProperties(t, tools["set_story_archived"], "workspaceId", "id", "archived", "confirmed")
	require.NotNil(t, tools["set_story_archived"].Annotations)
	require.NotNil(t, tools["set_story_archived"].Annotations.DestructiveHint)
	require.True(t, *tools["set_story_archived"].Annotations.DestructiveHint)
	requireToolProperties(t, tools["list_sprints"], "workspaceId", "search", "page", "pageSize")
	requireToolProperties(t, tools["analyze_sprint"], "workspaceId", "id", "page", "pageSize")
	requireToolProperties(t, tools["list_objectives"], "workspaceId", "search", "page", "pageSize")
	requireToolProperties(t, tools["list_objective_statuses"], "workspaceId", "page", "pageSize")
	requireToolProperties(t, tools["analyze_objective"], "workspaceId", "id", "page", "pageSize")
	requireToolProperties(t, tools["list_key_results"], "workspaceId", "page", "pageSize")
	requireToolProperties(t, tools["create_objective"], "name", "teamId", "startDate", "endDate", "confirmed")
	requireToolProperties(t, tools["update_objective"], "id", "expectedUpdatedAt", "health", "endDate", "confirmed")
	requireToolProperties(t, tools["update_key_result"], "id", "expectedUpdatedAt", "currentValue", "targetValue", "confirmed")
	requireToolProperties(t, tools["analyze_work"], "workspaceId", "startDate", "endDate", "page", "pageSize")
}

func TestEveryRegisteredToolDeclaresItsStructuredOutput(t *testing.T) {
	t.Parallel()
	tools := listToolsForTest(t)
	expectedRootProperties := map[string][]string{
		"list_workspaces":         {"workspaces", "page", "pageSize", "hasMore"},
		"list_teams":              {"teams", "page", "pageSize", "hasMore"},
		"list_story_statuses":     {"statuses", "page", "pageSize", "hasMore"},
		"list_stories":            {"stories", "page", "pageSize", "hasMore"},
		"create_story":            {"id", "sequenceId", "teamCode", "title", "createdNow"},
		"update_story":            {"story"},
		"list_story_comments":     {"comments", "page", "pageSize", "hasMore"},
		"add_story_comment":       {"comment"},
		"set_story_archived":      {"id", "archived", "changed"},
		"list_sprints":            {"sprints", "page", "pageSize", "hasMore"},
		"create_sprint":           {"sprint"},
		"analyze_sprint":          {"analysis", "page", "pageSize", "hasMore"},
		"list_objectives":         {"objectives", "page", "pageSize", "hasMore"},
		"list_objective_statuses": {"statuses", "page", "pageSize", "hasMore"},
		"create_objective":        {"objective"},
		"update_objective":        {"objective"},
		"analyze_objective":       {"analysis", "page", "pageSize", "hasMore"},
		"list_key_results":        {"keyResults", "totalCount", "page", "pageSize", "hasMore"},
		"create_key_result":       {"keyResult"},
		"update_key_result":       {"keyResult"},
		"analyze_work":            {"analysis", "guidance", "page", "pageSize", "hasMore"},
	}

	require.Len(t, tools, len(expectedRootProperties))
	for name, expectedProperties := range expectedRootProperties {
		definition := tools[name]
		require.NotNil(t, definition, "tool %q is not registered", name)
		schema, ok := definition.OutputSchema.(map[string]any)
		require.True(t, ok, "tool %q output schema has unexpected type %T", name, definition.OutputSchema)
		require.Equal(t, "object", schema["type"], "tool %q output must be a structured object", name)
		properties, ok := schema["properties"].(map[string]any)
		require.True(t, ok, "tool %q output schema has no properties", name)
		for _, property := range expectedProperties {
			require.Contains(t, properties, property, "tool %q output schema is missing %q", name, property)
		}
	}
}

func TestEveryRegisteredToolSerializesItsOutputSchema(t *testing.T) {
	t.Parallel()
	for name, definition := range listToolsForTest(t) {
		wire, err := json.Marshal(definition)
		require.NoError(t, err)

		var serialized map[string]any
		require.NoError(t, json.Unmarshal(wire, &serialized))
		require.Contains(t, serialized, "outputSchema", "tool %q omitted outputSchema from the MCP wire definition", name)
	}
}

func TestEveryOutputSchemaAcceptsItsToolResultShape(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	server := mcp.NewServer(&mcp.Implementation{Name: "output-schema-test", Version: "1.0.0"}, nil)
	for name, output := range sampleToolOutputs() {
		definition := tool(name, name, "Output schema contract test.", annotations(true, true))
		addSafeTool(&Handler{}, server, definition, func(_ context.Context, _ *mcp.CallToolRequest, _ emptyInput) (*mcp.CallToolResult, any, error) {
			return nil, output, nil
		})
	}

	serverTransport, clientTransport := mcp.NewInMemoryTransports()
	serverSession, err := server.Connect(ctx, serverTransport, nil)
	require.NoError(t, err)
	t.Cleanup(func() { _ = serverSession.Close() })
	client := mcp.NewClient(&mcp.Implementation{Name: "output-schema-client", Version: "1.0.0"}, nil)
	clientSession, err := client.Connect(ctx, clientTransport, nil)
	require.NoError(t, err)
	t.Cleanup(func() { _ = clientSession.Close() })

	for name := range sampleToolOutputs() {
		t.Run(name, func(t *testing.T) {
			result, callErr := clientSession.CallTool(ctx, &mcp.CallToolParams{Name: name, Arguments: map[string]any{}})
			require.NoError(t, callErr)
			require.False(t, result.IsError, "output schema rejected %s result: %#v", name, result.Content)
		})
	}
}

func sampleToolOutputs() map[string]map[string]any {
	pagination := func(collection string, items ...any) map[string]any {
		return map[string]any{collection: items, "page": 1, "pageSize": 25, "hasMore": false}
	}
	analysis := func(guidance bool) map[string]any {
		result := map[string]any{"analysis": map[string]any{}, "page": 1, "pageSize": 25, "hasMore": false}
		if guidance {
			result["guidance"] = "Interpretation guidance"
		}
		return result
	}

	return map[string]map[string]any{
		"list_workspaces":         pagination("workspaces", map[string]any{"id": "workspace-id", "slug": "workspace", "name": "Workspace", "role": "member"}),
		"list_teams":              pagination("teams", map[string]any{"id": "team-id", "name": "Engineering", "code": "ENG", "sprintsEnabled": true}),
		"list_story_statuses":     pagination("statuses", sampleStatusOutput()),
		"list_stories":            pagination("stories", map[string]any{"id": "story-id", "sequence_id": 1, "title": "Story", "team_id": "team-id", "priority": "High", "created_at": "2026-08-22T00:00:00Z", "updated_at": "2026-08-22T00:00:00Z"}),
		"create_story":            {"id": "story-id", "sequenceId": 1, "teamCode": "ENG", "title": "Story", "createdNow": true, "estimatedDurationMinutes": nil, "minimumFocusBlockMinutes": nil, "autoSchedulingEnabled": false},
		"update_story":            {"story": map[string]any{"id": "story-id", "sequenceId": 1, "title": "Story", "teamId": "team-id", "priority": "High", "createdAt": "2026-08-22T00:00:00Z", "updatedAt": "2026-08-22T00:00:00Z"}},
		"list_story_comments":     pagination("comments", sampleCommentOutput()),
		"add_story_comment":       {"comment": sampleCommentOutput()},
		"set_story_archived":      {"id": "story-id", "archived": true, "changed": true},
		"list_sprints":            pagination("sprints", sampleSprintOutput()),
		"create_sprint":           {"sprint": sampleSprintOutput()},
		"analyze_sprint":          analysis(false),
		"list_objectives":         pagination("objectives", sampleObjectiveOutput()),
		"list_objective_statuses": pagination("statuses", sampleStatusOutput()),
		"create_objective":        {"objective": sampleObjectiveOutput()},
		"update_objective":        {"objective": sampleObjectiveOutput()},
		"analyze_objective":       analysis(false),
		"list_key_results":        {"keyResults": []any{sampleKeyResultListOutput()}, "totalCount": 1, "page": 1, "pageSize": 25, "hasMore": false},
		"create_key_result":       {"keyResult": sampleKeyResultOutput()},
		"update_key_result":       {"keyResult": sampleKeyResultOutput()},
		"analyze_work":            analysis(true),
	}
}

func sampleStatusOutput() map[string]any {
	return map[string]any{"id": "status-id", "name": "In progress", "category": "started", "orderIndex": 1, "isDefault": false, "color": "orange"}
}

func sampleCommentOutput() map[string]any {
	return map[string]any{"comment_id": "comment-id", "story_id": "story-id", "commenter_id": "user-id", "content": "Comment", "created_at": "2026-08-22T00:00:00Z", "updated_at": "2026-08-22T00:00:00Z"}
}

func sampleSprintOutput() map[string]any {
	return map[string]any{"id": "sprint-id", "name": "Sprint", "teamId": "team-id", "startDate": "2026-08-22T00:00:00Z", "endDate": "2026-09-05T00:00:00Z", "createdAt": "2026-08-22T00:00:00Z", "updatedAt": "2026-08-22T00:00:00Z"}
}

func sampleObjectiveOutput() map[string]any {
	return map[string]any{"id": "objective-id", "sequenceId": 1, "name": "Objective", "teamId": "team-id", "statusId": "status-id", "isPrivate": false, "createdAt": "2026-08-22T00:00:00Z", "updatedAt": "2026-08-22T00:00:00Z"}
}

func sampleKeyResultOutput() map[string]any {
	return map[string]any{
		"id": "key-result-id", "sequenceId": 1, "objectiveId": "objective-id", "name": "Target",
		"measurementType": "percentage", "startValue": 0.0, "currentValue": 25.0, "targetValue": 100.0,
		"createdAt": "2026-08-22T00:00:00Z", "updatedAt": "2026-08-22T00:00:00Z",
	}
}

func sampleKeyResultListOutput() map[string]any {
	result := sampleKeyResultOutput()
	result["objectiveName"] = "Objective"
	result["teamId"] = "team-id"
	result["teamName"] = "Engineering"
	result["teamCode"] = "ENG"
	return result
}

func listToolsForTest(t *testing.T) map[string]*mcp.Tool {
	t.Helper()
	ctx := context.Background()
	serverTransport, clientTransport := mcp.NewInMemoryTransports()
	serverSession, err := (&Handler{}).newMCPServer().Connect(ctx, serverTransport, nil)
	require.NoError(t, err)
	t.Cleanup(func() { _ = serverSession.Close() })

	client := mcp.NewClient(&mcp.Implementation{Name: "agent-readiness-test", Version: "1.0.0"}, nil)
	clientSession, err := client.Connect(ctx, clientTransport, nil)
	require.NoError(t, err)
	t.Cleanup(func() { _ = clientSession.Close() })
	require.Equal(t, mcpServerVersion, clientSession.InitializeResult().ServerInfo.Version)

	result, err := clientSession.ListTools(ctx, nil)
	require.NoError(t, err)
	tools := make(map[string]*mcp.Tool, len(result.Tools))
	for _, definition := range result.Tools {
		tools[definition.Name] = definition
	}
	return tools
}

func requireToolProperties(t *testing.T, tool *mcp.Tool, names ...string) {
	t.Helper()
	require.NotNil(t, tool)
	schema, ok := tool.InputSchema.(map[string]any)
	require.True(t, ok, "tool %q input schema has unexpected type %T", tool.Name, tool.InputSchema)
	properties, ok := schema["properties"].(map[string]any)
	require.True(t, ok, "tool %q input schema has no properties", tool.Name)
	for _, name := range names {
		require.Contains(t, properties, name, "tool %q is missing %q", tool.Name, name)
	}
}

func requireToolRequiredProperties(t *testing.T, tool *mcp.Tool, names ...string) {
	t.Helper()
	require.NotNil(t, tool)
	schema, ok := tool.InputSchema.(map[string]any)
	require.True(t, ok, "tool %q input schema has unexpected type %T", tool.Name, tool.InputSchema)
	required, ok := schema["required"].([]any)
	require.True(t, ok, "tool %q input schema has no required properties", tool.Name)
	for _, name := range names {
		require.Contains(t, required, name, "tool %q does not require %q", tool.Name, name)
	}
}
