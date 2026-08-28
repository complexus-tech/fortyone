package messaging

import ()

func nullableUUID(description string) map[string]any {
	return map[string]any{"type": []string{"string", "null"}, "description": description}
}

func nullableDate(description string) map[string]any {
	return map[string]any{"type": []string{"string", "null"}, "description": description}
}

func nullableMinutes(description string) map[string]any {
	return map[string]any{
		"type":        []string{"integer", "null"},
		"description": description,
		"minimum":     1,
		"maximum":     maximumEstimatedDurationMinutes,
	}
}

func nullableBoolean(description string) map[string]any {
	return map[string]any{"type": []string{"boolean", "null"}, "description": description}
}

func storyTimeAction(description string) map[string]any {
	return map[string]any{
		"type":        "string",
		"description": description,
		"enum":        []string{storyTimeActionUnchanged, storyTimeActionSet, storyTimeActionClear},
	}
}

func storyMutationToolDefinitions() []ToolDefinition {
	nullablePriority := map[string]any{
		"type":        []string{"string", "null"},
		"description": "A FortyOne priority, or null as described by this tool.",
		"enum":        []any{"No Priority", "Low", "Medium", "High", "Urgent", nil},
	}
	nullableUpdatePriority := map[string]any{
		"type":        []string{"string", "null"},
		"description": "The replacement priority only when the user explicitly asked to change priority; otherwise null to leave the current priority unchanged. Never default an update to No Priority.",
		"enum":        []any{"No Priority", "Low", "Medium", "High", "Urgent", nil},
	}
	return []ToolDefinition{
		{
			Type:        "function",
			Name:        toolCreateStory,
			Description: "Prepare a story creation proposal only when the user explicitly asks to create one and the team and title are unambiguous. This tool never writes; FortyOne will require explicit user confirmation.",
			Strict:      true,
			Parameters: strictObjectSchema(map[string]any{
				"team_id": map[string]any{
					"type":        "string",
					"description": "An exact team UUID returned by list_teams.",
				},
				"title": map[string]any{
					"type":        "string",
					"description": "The exact story title requested by the user.",
					"minLength":   1,
					"maxLength":   maximumStoryTitleRunes,
				},
				"priority": nullablePriority,
				"assignee": map[string]any{
					"type":        "string",
					"description": "Use me only if the user explicitly asks to assign it to themselves; otherwise use unassigned.",
					"enum":        []string{assigneeActionMe, assigneeActionUnassigned},
				},
				"estimated_duration_minutes":  nullableMinutes("The total time needed in minutes when explicitly known, or null when unspecified."),
				"minimum_focus_block_minutes": nullableMinutes("The smallest useful uninterrupted block in minutes, no greater than estimated_duration_minutes; use null when duration is unspecified."),
				"auto_scheduling_enabled": map[string]any{
					"type":        "boolean",
					"description": "Enable Maya auto-scheduling only when the user explicitly requests it; otherwise false.",
				},
			}, []string{"team_id", "title", "priority", "assignee", "estimated_duration_minutes", "minimum_focus_block_minutes", "auto_scheduling_enabled"}),
		},
		{
			Type:        "function",
			Name:        toolCreateStories,
			Description: "Prepare one confirmation proposal for 1-10 distinct stories in one team when the user explicitly asks to turn a conversation into multiple action items. This tool never writes. Use only explicit assignee UUIDs returned by list_team_members; otherwise pass null. Source attribution is supplied by the server and is not a tool argument.",
			Strict:      true,
			Parameters: strictObjectSchema(map[string]any{
				"team_id": map[string]any{
					"type":        "string",
					"description": "An exact team UUID returned by list_teams.",
				},
				"stories": map[string]any{
					"type":     "array",
					"minItems": 1,
					"maxItems": maximumBatchStoryCount,
					"items": strictObjectSchema(map[string]any{
						"title": map[string]any{
							"type":      "string",
							"minLength": 1,
							"maxLength": maximumStoryTitleRunes,
						},
						"description": map[string]any{
							"type":        []string{"string", "null"},
							"maxLength":   maximumBatchDescriptionRunes,
							"description": "Concise supporting context derived from the visible conversation, or null.",
						},
						"priority": nullablePriority,
						"assignee_id": map[string]any{
							"type":        []string{"string", "null"},
							"description": "An exact active member UUID returned by list_team_members only when assignment is explicit, otherwise null.",
						},
						"estimated_duration_minutes":  nullableMinutes("The total time needed in minutes when explicitly known, or null when unspecified."),
						"minimum_focus_block_minutes": nullableMinutes("The smallest useful uninterrupted block in minutes, no greater than estimated_duration_minutes; use null when duration is unspecified."),
						"auto_scheduling_enabled": map[string]any{
							"type":        "boolean",
							"description": "Enable Maya auto-scheduling only when explicitly requested for this story; otherwise false.",
						},
					}, []string{"title", "description", "priority", "assignee_id", "estimated_duration_minutes", "minimum_focus_block_minutes", "auto_scheduling_enabled"}),
				},
			}, []string{"team_id", "stories"}),
		},
		{
			Type:        "function",
			Name:        toolUpdateStory,
			Description: "Prepare a story update proposal only when the target story and requested fields are unambiguous. Time fields use explicit unchanged, set, or clear actions so null is never ambiguous. This tool never writes; FortyOne will require explicit user confirmation.",
			Strict:      true,
			Parameters: strictObjectSchema(map[string]any{
				"story_id": map[string]any{
					"type":        []string{"string", "null"},
					"description": "An exact story UUID returned by a read tool, or null when story_reference is provided.",
				},
				"story_reference": map[string]any{
					"type":        []string{"string", "null"},
					"description": "An exact human reference such as WEB-123, or null when story_id is provided. Provide exactly one target.",
				},
				"title": map[string]any{
					"type":        []string{"string", "null"},
					"description": "The replacement title, or null to leave it unchanged.",
					"maxLength":   maximumStoryTitleRunes,
				},
				"priority": nullableUpdatePriority,
				"assignee": map[string]any{
					"type":        "string",
					"description": "Whether to leave the assignee unchanged, assign the current user, or unassign the story.",
					"enum":        []string{assigneeActionUnchanged, assigneeActionMe, assigneeActionUnassigned},
				},
				"status_id":                   nullableUUID("A visible status UUID, or null to leave unchanged."),
				"sprint_id":                   nullableUUID("A visible sprint UUID, or null to leave unchanged."),
				"objective_id":                nullableUUID("A visible objective UUID, or null to leave unchanged."),
				"key_result_id":               nullableUUID("A visible key result UUID, or null to leave unchanged."),
				"start_date":                  nullableDate("A start date in YYYY-MM-DD or RFC3339, or null to leave unchanged."),
				"end_date":                    nullableDate("An end date in YYYY-MM-DD or RFC3339, or null to leave unchanged."),
				"label_ids":                   map[string]any{"type": []string{"array", "null"}, "description": "The complete replacement list of visible label UUIDs, or null to leave labels unchanged.", "items": map[string]any{"type": "string"}},
				"estimated_duration_action":   storyTimeAction("Use unchanged to preserve the current time needed, set to replace it with estimated_duration_minutes, or clear to remove it. Clearing time needed also clears its minimum focus block."),
				"estimated_duration_minutes":  nullableMinutes("The replacement total time needed when estimated_duration_action is set; otherwise null."),
				"minimum_focus_block_action":  storyTimeAction("Use unchanged to preserve the current minimum focus block, set to replace it with minimum_focus_block_minutes, or clear to remove it."),
				"minimum_focus_block_minutes": nullableMinutes("The replacement minimum uninterrupted block when minimum_focus_block_action is set; otherwise null. It cannot exceed the resulting time needed."),
				"auto_scheduling_enabled":     nullableBoolean("Set true or false only when the user explicitly asks to change auto-scheduling; otherwise null."),
				"auto_scheduling_locked":      nullableBoolean("Set true or false only when the user explicitly asks to lock or unlock Maya's schedule; otherwise null. Locking requires auto-scheduling to remain enabled."),
			}, []string{"story_id", "story_reference", "title", "priority", "assignee", "status_id", "sprint_id", "objective_id", "key_result_id", "start_date", "end_date", "label_ids", "estimated_duration_action", "estimated_duration_minutes", "minimum_focus_block_action", "minimum_focus_block_minutes", "auto_scheduling_enabled", "auto_scheduling_locked"}),
		},
		{
			Type: "function", Name: toolAddComment,
			Description: "Prepare a comment proposal for an accessible story. The comment and optional mentions are written only after explicit user confirmation.", Strict: true,
			Parameters: strictObjectSchema(map[string]any{
				"story_id":        map[string]any{"type": []string{"string", "null"}},
				"story_reference": map[string]any{"type": []string{"string", "null"}},
				"body":            map[string]any{"type": "string", "minLength": 1, "maxLength": maxStoryDescriptionRunes},
				"mention_ids":     map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "maxItems": 20},
			}, []string{"story_id", "story_reference", "body", "mention_ids"}),
		},
		{
			Type: "function", Name: toolAddRelationship,
			Description: "Prepare a proposal to relate two accessible stories. The relationship is written only after explicit user confirmation.", Strict: true,
			Parameters: strictObjectSchema(map[string]any{
				"from_story_id":        map[string]any{"type": []string{"string", "null"}},
				"from_story_reference": map[string]any{"type": []string{"string", "null"}},
				"to_story_id":          map[string]any{"type": []string{"string", "null"}},
				"to_story_reference":   map[string]any{"type": []string{"string", "null"}},
				"association_type":     map[string]any{"type": "string", "enum": []string{"blocking", "related", "duplicate"}},
			}, []string{"from_story_id", "from_story_reference", "to_story_id", "to_story_reference", "association_type"}),
		},
	}
}
