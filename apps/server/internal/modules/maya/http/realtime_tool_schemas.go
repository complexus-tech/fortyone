package mayahttp

func realtimeExtendedTools() []openAIRealtimeTool {
	return []openAIRealtimeTool{
		realtimeTool(
			"navigate",
			"Open a FortyOne page or a specific record. Use human-readable names and references; the backend resolves IDs.",
			map[string]any{
				"targetType": enumProperty("Destination type.", "user-profile", "team", "sprint", "objective", "story", "feedback", "my-work", "summary", "analytics", "sprints", "notifications", "settings", "roadmaps", "billing"),
				"reference":  stringProperty("Human-readable story reference, record name, or person name for a dynamic destination."),
				"teamName":   stringProperty("Optional team name or code for a team-scoped destination."),
				"route":      enumProperty("Optional page within a team.", "stories", "sprints", "objectives", "backlog", "feedback", "requests"),
			},
			[]string{"targetType"},
		),
		realtimeTool(
			"set_theme",
			"Change the FortyOne appearance theme.",
			map[string]any{
				"theme": enumProperty("Theme to apply.", "light", "dark", "system", "toggle"),
			},
			[]string{"theme"},
		),
		realtimeTool(
			"get_story",
			"Get current details for one work item by its human-readable reference or title.",
			map[string]any{
				"reference": stringProperty("Story reference such as ENG-42, or an unambiguous title."),
			},
			[]string{"reference"},
		),
		realtimeTool(
			"update_story",
			"Update supported fields on one story after the user confirms the exact changes.",
			map[string]any{
				"reference":         stringProperty("Story reference such as ENG-42, or an unambiguous title."),
				"title":             stringProperty("Optional new title."),
				"status":            stringProperty("Optional status name or category."),
				"priority":          enumProperty("Optional priority.", "No Priority", "Low", "Medium", "High", "Urgent"),
				"assigneeName":      stringProperty("Optional assignee name or username."),
				"assignToMe":        booleanProperty("True when assigning the story to the current user."),
				"unassign":          booleanProperty("True when removing the current assignee."),
				"estimateValue":     integerProperty("Optional numeric estimate."),
				"clearEstimate":     booleanProperty("True when removing the current estimate."),
				"startDate":         stringProperty("Optional start date, preferably YYYY-MM-DD. Natural relative dates are accepted."),
				"clearStartDate":    booleanProperty("True when removing the current start date."),
				"endDate":           stringProperty("Optional due date, preferably YYYY-MM-DD. Natural relative dates are accepted."),
				"clearEndDate":      booleanProperty("True when removing the current due date."),
				"sprintName":        stringProperty("Optional sprint name to associate with the story."),
				"clearSprint":       booleanProperty("True when removing the story from its sprint."),
				"objectiveName":     stringProperty("Optional objective name to associate with the story."),
				"clearObjective":    booleanProperty("True when removing the story from its objective."),
				"confirmed":         booleanProperty("True only after the user explicitly confirms these exact changes."),
				"confirmationToken": stringProperty("Exact token returned by the preceding confirmation request."),
			},
			[]string{"reference", "confirmed"},
		),
		realtimeTool(
			"story_comments",
			"List comments on a story, or add a comment after the user confirms the exact text and target story.",
			map[string]any{
				"action":            enumProperty("Comment operation.", "list", "add"),
				"reference":         stringProperty("Story reference or unambiguous title."),
				"comment":           stringProperty("Exact comment text when adding."),
				"limit":             integerProperty("Maximum comments to return. Defaults to 10."),
				"confirmed":         booleanProperty("True only after the user explicitly confirms adding the exact comment."),
				"confirmationToken": stringProperty("Exact token returned by the preceding confirmation request."),
			},
			[]string{"action", "reference"},
		),
		realtimeTool(
			"sprints",
			"List running sprints or get the progress summary for one sprint.",
			map[string]any{
				"action":   enumProperty("Sprint operation.", "list_running", "get_summary"),
				"name":     stringProperty("Sprint name for get_summary."),
				"teamName": stringProperty("Optional team name or code."),
				"limit":    integerProperty("Maximum running sprints to return. Defaults to 10."),
			},
			[]string{"action"},
		),
		realtimeTool(
			"workload",
			"Get current workload, capacity risks, overdue work, and unassigned work.",
			map[string]any{
				"teamName": stringProperty("Optional team name or code."),
				"limit":    integerProperty("Maximum at-risk members to return. Defaults to 5."),
			},
			[]string{},
		),
		realtimeTool(
			"recent_activity",
			"Get recent workspace activity visible to the current user.",
			map[string]any{
				"days":  integerProperty("Lookback window in days, from 1 to 90. Defaults to 7."),
				"limit": integerProperty("Maximum activities to return. Defaults to 10."),
			},
			[]string{},
		),
		realtimeTool(
			"notifications",
			"List notifications, get the unread count, or mark one or all as read after confirmation.",
			map[string]any{
				"action":            enumProperty("Notification operation.", "list", "unread_count", "mark_read", "mark_all_read"),
				"title":             stringProperty("Exact notification title for mark_read."),
				"limit":             integerProperty("Maximum notifications to return. Defaults to 10."),
				"confirmed":         booleanProperty("True only after the user confirms the exact read action."),
				"confirmationToken": stringProperty("Exact token returned by the preceding confirmation request."),
			},
			[]string{"action"},
		),
		realtimeTool(
			"customer_feedback",
			"List customer feedback or find one feedback item by title. Can be scoped to a team and status.",
			map[string]any{
				"action":   enumProperty("Feedback operation.", "list", "get"),
				"title":    stringProperty("Feedback title search, required for get."),
				"teamName": stringProperty("Optional team name or code."),
				"status":   enumProperty("Feedback status filter.", "active", "all", "pending", "reviewing", "planned", "in_progress", "completed", "closed"),
				"limit":    integerProperty("Maximum feedback items to return. Defaults to 10."),
			},
			[]string{"action"},
		),
		realtimeTool(
			"workspace_briefing",
			"Get a concise workspace operational briefing covering delivery, risks, objectives, sprints, feedback, and workload.",
			map[string]any{
				"days": integerProperty("Reporting window in days, from 1 to 365. Defaults to 30."),
			},
			[]string{},
		),
	}
}

func realtimeTool(name, description string, properties map[string]any, required []string) openAIRealtimeTool {
	return openAIRealtimeTool{
		Type:        "function",
		Name:        name,
		Description: description,
		Parameters: map[string]any{
			"type":                 "object",
			"additionalProperties": false,
			"properties":           properties,
			"required":             required,
		},
	}
}

func stringProperty(description string) map[string]any {
	return map[string]any{"type": "string", "description": description}
}

func booleanProperty(description string) map[string]any {
	return map[string]any{"type": "boolean", "description": description}
}

func integerProperty(description string) map[string]any {
	return map[string]any{"type": "integer", "description": description}
}

func enumProperty(description string, values ...string) map[string]any {
	return map[string]any{"type": "string", "enum": values, "description": description}
}
