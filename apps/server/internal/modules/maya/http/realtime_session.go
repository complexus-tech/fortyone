package mayahttp

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	teams "github.com/complexus-tech/projects-api/internal/modules/teams/service"
	"github.com/google/uuid"
)

func (h *Handlers) createRealtimeClientSecret(ctx context.Context, workspaceID, userID uuid.UUID, sessionRequest AppRealtimeSessionRequest) (AppRealtimeSession, error) {
	terminology := h.realtimeTerminology(ctx, workspaceID)
	workspaceTeams, err := h.teams.List(ctx, workspaceID, userID)
	if err != nil {
		return AppRealtimeSession{}, fmt.Errorf("list teams for realtime context: %w", err)
	}
	currentUser, err := h.currentRealtimeUser(ctx, userID)
	if err != nil {
		return AppRealtimeSession{}, err
	}

	payload := openAIRealtimeClientSecretRequest{
		Session: newRealtimeSessionConfig(terminology, workspaceTeams, currentUser, sessionRequest),
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return AppRealtimeSession{}, fmt.Errorf("marshal realtime session request: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		strings.TrimRight(h.baseURL, "/")+"/realtime/client_secrets",
		bytes.NewReader(body),
	)
	if err != nil {
		return AppRealtimeSession{}, fmt.Errorf("create realtime session request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+h.aiAPIKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("OpenAI-Safety-Identifier", safetyIdentifier(userID))

	res, err := h.client.Do(req)
	if err != nil {
		return AppRealtimeSession{}, fmt.Errorf("call realtime session endpoint: %w", err)
	}
	defer res.Body.Close()

	data, err := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if err != nil {
		return AppRealtimeSession{}, fmt.Errorf("read realtime session response: %w", err)
	}
	if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusMultipleChoices {
		return AppRealtimeSession{}, fmt.Errorf("realtime session endpoint returned %s: %s", res.Status, strings.TrimSpace(string(data)))
	}

	var response openAIRealtimeClientSecretResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return AppRealtimeSession{}, fmt.Errorf("decode realtime session response: %w", err)
	}

	clientSecret := strings.TrimSpace(response.Value)
	expiresAt := response.ExpiresAt
	if clientSecret == "" && response.ClientSecret != nil {
		clientSecret = strings.TrimSpace(response.ClientSecret.Value)
		expiresAt = response.ClientSecret.ExpiresAt
	}
	if clientSecret == "" {
		return AppRealtimeSession{}, errors.New("realtime session response did not include a client secret")
	}

	return AppRealtimeSession{
		ClientSecret: clientSecret,
		ExpiresAt:    expiresAt,
		Model:        defaultRealtimeModel,
		Voice:        defaultRealtimeVoice,
	}, nil
}

func newRealtimeSessionConfig(terminology AppRealtimeTerminology, workspaceTeams []teams.CoreTeam, currentUser AppRealtimeVoiceUser, sessionRequest AppRealtimeSessionRequest) openAIRealtimeSessionConfig {
	return openAIRealtimeSessionConfig{
		Type:             "realtime",
		Model:            defaultRealtimeModel,
		Instructions:     realtimeInstructions(terminology, workspaceTeams, currentUser, sessionRequest),
		OutputModalities: []string{"audio"},
		Tools:            realtimeTools(),
		ToolChoice:       "auto",
		Audio: openAIRealtimeAudioConfig{
			Input: openAIRealtimeAudioInputConfig{
				NoiseReduction: openAIRealtimeNoiseReductionConfig{
					Type: "near_field",
				},
				Transcription: openAIRealtimeTranscriptionConfig{
					Language: "en",
					Model:    defaultRealtimeTranscriptionModel,
					Prompt:   realtimeTranscriptionPrompt(terminology, workspaceTeams),
				},
				TurnDetection: openAIRealtimeTurnDetectionConfig{
					Type:              "server_vad",
					Threshold:         0.75,
					PrefixPaddingMs:   300,
					SilenceDurationMs: 700,
					CreateResponse:    true,
					InterruptResponse: true,
				},
			},
			Output: openAIRealtimeAudioOutputConfig{
				Voice: defaultRealtimeVoice,
			},
		},
	}
}

func realtimeInstructions(terminology AppRealtimeTerminology, workspaceTeams []teams.CoreTeam, currentUser AppRealtimeVoiceUser, sessionRequest AppRealtimeSessionRequest) string {
	instructions := []string{
		"You are Maya, FortyOne's AI agent for project management.",
		"Your job is to help users manage work in FortyOne: work items, teams, priorities, assignments, workload, objectives, key results, activity, and workspace insights.",
		"In voice mode, be concise, natural, and direct. Prefer one to three spoken sentences unless the user asks for detail.",
		"Sound warm, sharp, curious, and genuinely enjoyable to talk to. Let personality come through as natural, context-dependent banter rather than a scripted joke.",
		"Use more playful energy for casual conversation and a lighter touch for professional or operational requests. Keep confirmations, failures, permissions, and sensitive topics straightforward and respectful.",
		"Avoid puns, dad jokes, forced analogies, corporate wordplay, fixed joke templates, and unrelated quips.",
		"Stay focused on project management inside FortyOne. Briefly redirect off-topic requests back to project-management help.",
		"Use available tools whenever facts, permissions, current state, IDs, or state changes are involved.",
		fmt.Sprintf("The current authenticated user is %s (@%s). When the user says me, my, or assign to me, resolve that to this user.", currentUser.Name, currentUser.Username),
		fmt.Sprintf("The user's timezone is %s. Today is %s and the current local time is %s. Interpret relative dates like today, tomorrow, this Friday, and next week in this timezone.", currentUser.Timezone, currentUser.Today, currentUser.Now),
		fmt.Sprintf("Use this workspace's preferred terminology when speaking: stories are called %q/%q, sprints are called %q/%q, objectives are called %q/%q, and key results are called %q/%q.", terminology.Story, terminology.Stories, terminology.Sprint, terminology.Sprints, terminology.Objective, terminology.Objectives, terminology.KeyResult, terminology.KeyResults),
		"Understand all common aliases even when you do not speak them back: story, task, issue, work item, objective, goal, project, key result, milestone, focus area, KPI, sprint, cycle, and iteration.",
		"Use get_context when you need current terminology or team context.",
		"Use list_my_tasks when the user asks about their assigned work, current work, plate, priorities, deadlines, overdue work, what they have today, or what to focus on.",
		"Use list_teams or list_team_members for team questions.",
		"Use search_work when the user asks to find or look up work by name, description, topic, or keyword.",
		fmt.Sprintf("Use list_objectives for %s/%s questions and list_key_results for %s/%s questions.", terminology.Objective, terminology.Objectives, terminology.KeyResult, terminology.KeyResults),
		fmt.Sprintf("Use create_task when the user asks you to create a %s, task, story, issue, or work item.", terminology.Story),
		"Use navigate to open FortyOne pages or records, and set_theme to change the application's appearance.",
		"Use get_story and update_story for story details and confirmed field changes. Use story_comments to read comments or add one after confirmation.",
		"Use sprints for running sprint lists and sprint summaries, workload for workload or capacity questions, recent_activity for recent workspace changes, notifications for notification questions and confirmed read actions, customer_feedback for customer feedback, and workspace_briefing for a concise operational overview.",
		"When the user clearly ends the conversation with phrases like bye, goodbye, that's all, thanks that's all, or talk later, say a brief goodbye and call end_conversation.",
		"Do not guess teams, statuses, permissions, or results. Ask a short clarifying question when the target is ambiguous.",
		teamSelectionInstruction(workspaceTeams),
		"Never expose raw UUIDs. Use human-readable names and story references.",
		"Keep tool usage internal. Do not mention tool names, parameters, or implementation details to the user.",
		"Never claim an action succeeded unless the tool result clearly shows success.",
		fmt.Sprintf("For %s creation: gather the title and target team if needed, resolve assignees from team members, convert natural dates to startDate/endDate, draft a concise title and useful description, ask for explicit confirmation, then call create_task with confirmed=true only after the user confirms the exact %s.", terminology.Story, terminology.Story),
		"For assignment during creation: set assignToMe=true when the user says me, myself, or assign to me. Set assigneeName when the user names another person; the backend resolves that name against team members.",
		"For estimates during creation: set estimateValue only when the user gives a numeric estimate such as 1, 2, 3, 5, or 8. If the estimate is non-numeric or unclear, ask a short clarifying question.",
		"For blockers and related work during creation: set blockedByRef when the new item is blocked by existing work, blockingRef when the new item blocks existing work, and relatedRef for related existing work. Use a human-readable story reference or title; the backend resolves it.",
		"If a tool returns requiresConfirmation, ask the requested confirmation in plain language. If the user confirms, repeat the exact same action details with confirmed=true and the returned confirmationToken. Never invent or reuse a token for different details.",
		"If a tool returns needsTeam, ask the requested clarification in plain language.",
		"If a tool returns needsAssignee, ask which team member should be assigned.",
		"If a tool returns needsStoryReference, ask which existing work item the user meant, using the returned references and titles.",
		"If a tool fails, repeat the useful error briefly. Do not invent a fallback workflow.",
	}

	if currentPath := strings.TrimSpace(sessionRequest.CurrentPath); currentPath != "" {
		instructions = append(instructions, fmt.Sprintf("The user started voice mode from the FortyOne path %q. Use it only to resolve references such as this page or this story when the conversation supports that interpretation.", currentPath))
	}
	if recentConversation := realtimeConversationContext(sessionRequest.Messages); recentConversation != "" {
		instructions = append(instructions, "Continue naturally from this recent typed and voice conversation. Do not repeat information the user already received:\n"+recentConversation)
	}

	return strings.Join(instructions, " ")
}

func realtimeConversationContext(messages []AppRealtimeConversationMessage) string {
	if len(messages) == 0 {
		return ""
	}

	lines := make([]string, 0, len(messages))
	for _, message := range messages {
		text := strings.TrimSpace(message.Text)
		if text == "" {
			continue
		}
		speaker := "User"
		if message.Role == "assistant" {
			speaker = "Maya"
		}
		lines = append(lines, speaker+": "+text)
	}
	return strings.Join(lines, "\n")
}

func realtimeTranscriptionPrompt(terminology AppRealtimeTerminology, workspaceTeams []teams.CoreTeam) string {
	terms := []string{
		"FortyOne",
		"Maya",
		terminology.Story,
		terminology.Stories,
		terminology.Sprint,
		terminology.Sprints,
		terminology.Objective,
		terminology.Objectives,
		terminology.KeyResult,
		terminology.KeyResults,
	}
	for _, team := range workspaceTeams {
		if name := strings.TrimSpace(team.Name); name != "" {
			terms = append(terms, name)
		}
	}
	return "Expect FortyOne workspace terminology, names, and references including: " + strings.Join(terms, ", ") + "."
}

func realtimeTools() []openAIRealtimeTool {
	tools := []openAIRealtimeTool{
		{
			Type:        "function",
			Name:        "end_conversation",
			Description: "End the realtime voice conversation when the user says goodbye or clearly indicates they are done.",
			Parameters: map[string]any{
				"type":                 "object",
				"additionalProperties": false,
				"properties":           map[string]any{},
				"required":             []string{},
			},
		},
		{
			Type:        "function",
			Name:        "get_context",
			Description: "Get the current workspace terminology and the teams available to the user.",
			Parameters: map[string]any{
				"type":                 "object",
				"additionalProperties": false,
				"properties":           map[string]any{},
				"required":             []string{},
			},
		},
		{
			Type:        "function",
			Name:        "list_teams",
			Description: "List teams the current user belongs to. Use for team questions or to resolve team names.",
			Parameters: map[string]any{
				"type":                 "object",
				"additionalProperties": false,
				"properties": map[string]any{
					"search": map[string]any{
						"type":        "string",
						"description": "Optional team name or code search.",
					},
					"limit": map[string]any{
						"type":        "integer",
						"description": "Maximum number of teams to return. Defaults to 25.",
					},
				},
				"required": []string{},
			},
		},
		{
			Type:        "function",
			Name:        "list_team_members",
			Description: "List members of a team. If the user belongs to exactly one team, teamName can be omitted.",
			Parameters: map[string]any{
				"type":                 "object",
				"additionalProperties": false,
				"properties": map[string]any{
					"teamName": map[string]any{
						"type":        "string",
						"description": "Team name or code. Omit only when the current user has one team.",
					},
					"search": map[string]any{
						"type":        "string",
						"description": "Optional member name, username, or email search.",
					},
					"limit": map[string]any{
						"type":        "integer",
						"description": "Maximum number of members to return. Defaults to 25.",
					},
				},
				"required": []string{},
			},
		},
		{
			Type:        "function",
			Name:        "list_my_tasks",
			Description: "List current user's assigned work items. Understands task/story/issue/work item terminology.",
			Parameters: map[string]any{
				"type":                 "object",
				"additionalProperties": false,
				"properties": map[string]any{
					"includeCompleted": map[string]any{
						"type":        "boolean",
						"description": "Whether completed stories should be included. Defaults to false.",
					},
					"limit": map[string]any{
						"type":        "integer",
						"description": "Maximum number of stories to return. Defaults to 10.",
					},
				},
				"required": []string{},
			},
		},
		{
			Type:        "function",
			Name:        "search_work",
			Description: "Search across work items and objectives by topic, name, or description. Use for find/look up questions.",
			Parameters: map[string]any{
				"type":                 "object",
				"additionalProperties": false,
				"properties": map[string]any{
					"query": map[string]any{
						"type":        "string",
						"description": "The text to search for.",
					},
					"type": map[string]any{
						"type":        "string",
						"enum":        []string{"all", "stories", "objectives"},
						"description": "Content type to search. Defaults to all.",
					},
					"teamName": map[string]any{
						"type":        "string",
						"description": "Optional team name or code to restrict the search.",
					},
					"limit": map[string]any{
						"type":        "integer",
						"description": "Maximum number of results to return. Defaults to 10.",
					},
				},
				"required": []string{"query"},
			},
		},
		{
			Type:        "function",
			Name:        "list_objectives",
			Description: "List objectives/goals/projects accessible to the user, respecting the workspace's selected terminology.",
			Parameters: map[string]any{
				"type":                 "object",
				"additionalProperties": false,
				"properties": map[string]any{
					"search": map[string]any{
						"type":        "string",
						"description": "Optional objective/goal/project name search.",
					},
					"teamName": map[string]any{
						"type":        "string",
						"description": "Optional team name or code.",
					},
					"limit": map[string]any{
						"type":        "integer",
						"description": "Maximum number of results to return. Defaults to 10.",
					},
				},
				"required": []string{},
			},
		},
		{
			Type:        "function",
			Name:        "list_key_results",
			Description: "List key results/milestones/focus areas/KPIs accessible to the user, respecting workspace terminology.",
			Parameters: map[string]any{
				"type":                 "object",
				"additionalProperties": false,
				"properties": map[string]any{
					"teamName": map[string]any{
						"type":        "string",
						"description": "Optional team name or code.",
					},
					"limit": map[string]any{
						"type":        "integer",
						"description": "Maximum number of results to return. Defaults to 10.",
					},
				},
				"required": []string{},
			},
		},
		{
			Type:        "function",
			Name:        "create_task",
			Description: "Create a new FortyOne work item after the user has confirmed the exact item. Understands task/story/issue/work item terminology.",
			Parameters: map[string]any{
				"type":                 "object",
				"additionalProperties": false,
				"properties": map[string]any{
					"title": map[string]any{
						"type":        "string",
						"description": "The story title.",
					},
					"description": map[string]any{
						"type":        "string",
						"description": "Optional story description.",
					},
					"teamName": map[string]any{
						"type":        "string",
						"description": "The target team name when the workspace has more than one team or the user mentioned a team.",
					},
					"assigneeName": map[string]any{
						"type":        "string",
						"description": "Optional assignee name, username, or 'me'. The backend resolves this against the selected team's members.",
					},
					"assignToMe": map[string]any{
						"type":        "boolean",
						"description": "Set true when the user asks to assign the item to themselves.",
					},
					"priority": map[string]any{
						"type":        "string",
						"enum":        []string{"No Priority", "Low", "Medium", "High", "Urgent"},
						"description": "Optional story priority. Defaults to No Priority.",
					},
					"estimateValue": map[string]any{
						"type":        "integer",
						"description": "Optional numeric estimate value. Use only when the user gives a numeric estimate.",
					},
					"startDate": map[string]any{
						"type":        "string",
						"description": "Optional start date. Prefer YYYY-MM-DD, but natural phrases like today, tomorrow, this Friday, or next Friday are accepted.",
					},
					"endDate": map[string]any{
						"type":        "string",
						"description": "Optional due/deadline date. Prefer YYYY-MM-DD, but natural phrases like today, tomorrow, this Friday, or next Friday are accepted.",
					},
					"blockedByRef": map[string]any{
						"type":        "string",
						"description": "Optional existing story reference or title that blocks this new story.",
					},
					"blockingRef": map[string]any{
						"type":        "string",
						"description": "Optional existing story reference or title that this new story blocks.",
					},
					"relatedRef": map[string]any{
						"type":        "string",
						"description": "Optional existing story reference or title related to this new story.",
					},
					"confirmed": map[string]any{
						"type":        "boolean",
						"description": "True only after the user explicitly confirms creating this story.",
					},
					"confirmationToken": map[string]any{
						"type":        "string",
						"description": "The exact token returned by the preceding confirmation request. Include only after the user confirms without changing details.",
					},
				},
				"required": []string{"title", "confirmed"},
			},
		},
	}
	return append(tools, realtimeExtendedTools()...)
}
