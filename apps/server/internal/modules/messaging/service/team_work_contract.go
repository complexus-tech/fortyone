package messaging

import (
	"github.com/google/uuid"
)

func teamWorkToolDefinitions() []ToolDefinition {
	nullableDate := func(description string) map[string]any {
		return map[string]any{
			"type":        []string{"string", "null"},
			"description": description,
			"pattern":     "^\\d{4}-\\d{2}-\\d{2}$",
		}
	}
	nullableLimit := func(description string) map[string]any {
		return map[string]any{
			"type":        []string{"integer", "null"},
			"description": description,
			"minimum":     1,
			"maximum":     maxToolLimit,
		}
	}

	return []ToolDefinition{
		{
			Type: "function",
			Name: toolListTeamWork,
			Description: "List work currently assigned to the authenticated user, selected active members, or all active members of exactly one authorized team. " +
				"Cross-assignee requests are limited to server-authorized shared teams and return a structured denied result when shared access is unavailable. Large all-member reports disclose assignee truncation. Completed work is filtered and ordered by completed_at and grouped by the current assignee; it does not identify who moved the task to Done.",
			Strict: true,
			Parameters: strictObjectSchema(map[string]any{
				"team_id": map[string]any{
					"type":        "string",
					"description": "One exact team UUID returned by list_teams.",
				},
				"assignee_scope": map[string]any{
					"type":        "string",
					"description": "Use me for the authenticated user's work, selected for exact member IDs, or all for every active human team member.",
					"enum":        []string{teamWorkAssigneeMe, teamWorkAssigneeSelected, teamWorkAssigneeAll},
				},
				"assignee_ids": map[string]any{
					"type":        []string{"array", "null"},
					"description": "Exact member UUIDs returned by list_team_members when assignee_scope is selected; otherwise null.",
					"items": map[string]any{
						"type": "string",
					},
					"maxItems": maxSelectedAssignees,
				},
				"mode": map[string]any{
					"type":        "string",
					"description": "in_progress means status category started; active includes backlog, unstarted, started, and paused; completed uses completed_at; due uses end_date and excludes completed work.",
					"enum":        []string{teamWorkModeInProgress, teamWorkModeActive, teamWorkModeCompleted, teamWorkModeDue},
				},
				"start_date": nullableDate("Local start date for completed or due work in YYYY-MM-DD format; null defaults to today or the supplied end_date. Must be null for active and in_progress."),
				"end_date":   nullableDate("Local end date for completed or due work in YYYY-MM-DD format; null defaults to the supplied start_date or today. Must be null for active and in_progress."),
				"group_by": map[string]any{
					"type":        "string",
					"description": "Use assignee to return per-person groups, or none for one flat task list.",
					"enum":        []string{teamWorkGroupAssignee, teamWorkGroupNone},
				},
				"limit":              nullableLimit("Maximum tasks across the whole response, or null for the default."),
				"limit_per_assignee": nullableLimit("Maximum tasks in each assignee group, or null for the default. Must be null when group_by is none."),
			}, []string{
				"team_id",
				"assignee_scope",
				"assignee_ids",
				"mode",
				"start_date",
				"end_date",
				"group_by",
				"limit",
				"limit_per_assignee",
			}),
		},
	}
}

type teamWorkAssigneeGroup struct {
	AssigneeID       uuid.UUID    `json:"assignee_id"`
	AssigneeName     string       `json:"assignee_name"`
	AssigneeUsername string       `json:"assignee_username"`
	Total            int          `json:"total"`
	TotalIsExact     bool         `json:"total_is_exact"`
	Returned         int          `json:"returned"`
	Truncated        bool         `json:"truncated"`
	Tasks            []taskResult `json:"tasks"`
}

type listTeamWorkResult struct {
	Team               teamResult              `json:"team"`
	Access             string                  `json:"access"`
	AccessReason       string                  `json:"access_reason,omitempty"`
	AssigneeScope      string                  `json:"assignee_scope"`
	Mode               string                  `json:"mode"`
	GroupBy            string                  `json:"group_by"`
	StartDate          *string                 `json:"start_date,omitempty"`
	EndDate            *string                 `json:"end_date,omitempty"`
	Limit              int                     `json:"limit"`
	LimitPerAssignee   int                     `json:"limit_per_assignee,omitempty"`
	Total              int                     `json:"total"`
	TotalIsExact       bool                    `json:"total_is_exact"`
	Returned           int                     `json:"returned"`
	Truncated          bool                    `json:"truncated"`
	AssigneesTruncated bool                    `json:"assignees_truncated"`
	AssigneeTotal      int                     `json:"assignee_total,omitempty"`
	AssigneeReturned   int                     `json:"assignee_returned,omitempty"`
	Tasks              []taskResult            `json:"tasks,omitempty"`
	Groups             []teamWorkAssigneeGroup `json:"groups,omitempty"`
}
