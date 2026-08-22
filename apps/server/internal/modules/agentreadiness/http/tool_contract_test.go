package agentreadinesshttp

import (
	"bytes"
	"context"
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
	addSafeTool(handler, server, tool("failing_tool", "Fail safely", "Test tool.", annotations(true, true)), func(context.Context, *mcp.CallToolRequest, emptyInput) (*mcp.CallToolResult, any, error) {
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
		{question: "Show me my work in the Art Circles team.", tool: "list_stories", cue: "my work"},
		{question: "Show my tasks and issues due today.", tool: "list_stories", cue: "due today"},
		{question: "Create a 30-minute task and auto-schedule it.", tool: "create_story", cue: "task"},
		{question: "Create an issue in the current sprint.", tool: "create_story", cue: "issue"},
		{question: "Show the team's sprints, iterations, or cycles.", tool: "list_sprints", cue: "cycles"},
		{question: "Create a two-week iteration for this team.", tool: "create_sprint", cue: "iteration"},
		{question: "How is this sprint going?", tool: "analyze_sprint", cue: "burndown"},
		{question: "Show our projects, goals, and objectives.", tool: "list_objectives", cue: "projects"},
		{question: "Create a project or goal for the launch.", tool: "create_objective", cue: "project"},
		{question: "Is this goal on track?", tool: "analyze_objective", cue: "goal"},
		{question: "Show the KRs, outcomes, measures, or targets for this goal.", tool: "list_key_results", cue: "targets"},
		{question: "Create a percentage target for this objective.", tool: "create_key_result", cue: "target"},
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

	require.Len(t, tools, 13)
	require.Len(t, covered, 13, "every registered tool must have at least one natural-language scenario")
	requireToolProperties(t, tools["list_stories"], "workspaceId", "teamId", "assignedToMe", "dueOn")
	requireToolProperties(t, tools["create_story"], "title", "estimatedDurationMinutes", "minimumFocusBlockMinutes", "autoSchedulingEnabled", "confirmed")
	requireToolProperties(t, tools["create_objective"], "name", "teamId", "startDate", "endDate", "confirmed")
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
