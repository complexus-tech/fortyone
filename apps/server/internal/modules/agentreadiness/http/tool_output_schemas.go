package agentreadinesshttp

type toolJSONSchema = map[string]any

func outputSchemaForTool(name string) toolJSONSchema {
	switch name {
	case "list_workspaces":
		return paginatedCollectionOutput("workspaces", workspaceOutputSchema())
	case "list_teams":
		return paginatedCollectionOutput("teams", teamOutputSchema())
	case "list_story_statuses", "list_objective_statuses":
		return paginatedCollectionOutput("statuses", statusOutputSchema())
	case "list_stories":
		return paginatedCollectionOutput("stories", storyListOutputSchema())
	case "create_story":
		return objectOutput(toolJSONSchema{
			"id":                       stringOutput("Story UUID"),
			"sequenceId":               integerOutput("Team story sequence number"),
			"teamCode":                 stringOutput("Team code used in the human-readable story identifier"),
			"title":                    stringOutput("Story title"),
			"createdNow":               booleanOutput("False when an idempotent retry returned the previously created story"),
			"estimatedDurationMinutes": nullableIntegerOutput("Focused work duration in minutes"),
			"minimumFocusBlockMinutes": nullableIntegerOutput("Minimum focus block in minutes"),
			"autoSchedulingEnabled":    booleanOutput("Whether automatic scheduling is enabled"),
		}, "id", "sequenceId", "teamCode", "title", "createdNow", "estimatedDurationMinutes", "minimumFocusBlockMinutes", "autoSchedulingEnabled")
	case "update_story":
		return wrappedEntityOutput("story", storyDetailOutputSchema())
	case "list_story_comments":
		return paginatedCollectionOutput("comments", commentOutputSchema())
	case "add_story_comment":
		return wrappedEntityOutput("comment", commentOutputSchema())
	case "set_story_archived":
		return objectOutput(toolJSONSchema{
			"id":       stringOutput("Story UUID"),
			"archived": booleanOutput("Current archived state"),
			"changed":  booleanOutput("Whether this call changed the archived state"),
		}, "id", "archived", "changed")
	case "list_sprints":
		return paginatedCollectionOutput("sprints", sprintOutputSchema())
	case "create_sprint":
		return wrappedEntityOutput("sprint", sprintOutputSchema())
	case "analyze_sprint", "analyze_objective":
		return analysisOutput(false)
	case "list_objectives":
		return paginatedCollectionOutput("objectives", objectiveOutputSchema())
	case "create_objective", "update_objective":
		return wrappedEntityOutput("objective", objectiveOutputSchema())
	case "list_key_results":
		return objectOutput(toolJSONSchema{
			"keyResults": arrayOutput(keyResultListOutputSchema(), "Key results on this page"),
			"totalCount": integerOutput("Total matching key results"),
			"page":       integerOutput("Current page number"),
			"pageSize":   integerOutput("Maximum results on this page"),
			"hasMore":    booleanOutput("Whether another page is available"),
		}, "keyResults", "totalCount", "page", "pageSize", "hasMore")
	case "create_key_result", "update_key_result":
		return wrappedEntityOutput("keyResult", keyResultOutputSchema())
	case "analyze_work":
		return analysisOutput(true)
	default:
		panic("missing MCP output schema for tool " + name)
	}
}

func paginatedCollectionOutput(collection string, itemSchema toolJSONSchema) toolJSONSchema {
	return objectOutput(toolJSONSchema{
		collection: arrayOutput(itemSchema, "Results on this page"),
		"page":     integerOutput("Current page number"),
		"pageSize": integerOutput("Maximum results on this page"),
		"hasMore":  booleanOutput("Whether another page is available"),
	}, collection, "page", "pageSize", "hasMore")
}

func wrappedEntityOutput(name string, entitySchema toolJSONSchema) toolJSONSchema {
	return objectOutput(toolJSONSchema{name: entitySchema}, name)
}

func analysisOutput(includeGuidance bool) toolJSONSchema {
	properties := toolJSONSchema{
		"analysis": objectValueOutput("FortyOne analysis with aggregate metrics and paginated detail sections"),
		"page":     integerOutput("Current detail page number"),
		"pageSize": integerOutput("Maximum rows in each detail section"),
		"hasMore":  booleanOutput("Whether any detail section has another page"),
	}
	required := []string{"analysis", "page", "pageSize", "hasMore"}
	if includeGuidance {
		properties["guidance"] = stringOutput("Instructions for interpreting and acting on the analysis")
		required = append(required, "guidance")
	}
	return objectOutput(properties, required...)
}

func workspaceOutputSchema() toolJSONSchema {
	return entityOutput("FortyOne workspace", toolJSONSchema{
		"id":   stringOutput("Workspace UUID"),
		"slug": stringOutput("Workspace URL slug"),
		"name": stringOutput("Workspace name"),
		"role": stringOutput("Connected user's workspace role"),
	}, "id", "slug", "name", "role")
}

func teamOutputSchema() toolJSONSchema {
	return entityOutput("FortyOne team", toolJSONSchema{
		"id":             stringOutput("Team UUID"),
		"name":           stringOutput("Team name"),
		"code":           stringOutput("Short team code"),
		"sprintsEnabled": booleanOutput("Whether the team uses sprints"),
	}, "id", "name", "code", "sprintsEnabled")
}

func statusOutputSchema() toolJSONSchema {
	return entityOutput("Workflow status", toolJSONSchema{
		"id":         stringOutput("Status UUID"),
		"name":       stringOutput("Display name"),
		"category":   stringOutput("Workflow category such as backlog, unstarted, started, completed, or cancelled"),
		"orderIndex": integerOutput("Display order"),
		"isDefault":  booleanOutput("Whether this is the default status"),
		"color":      stringOutput("Status color"),
	}, "id", "name", "category", "orderIndex", "isDefault", "color")
}

func storyListOutputSchema() toolJSONSchema {
	return entityOutput("Story, task, issue, ticket, or work item", toolJSONSchema{
		"id":          stringOutput("Story UUID"),
		"sequence_id": integerOutput("Team story sequence number"),
		"title":       stringOutput("Story title"),
		"team_id":     stringOutput("Team UUID"),
		"priority":    stringOutput("Story priority"),
		"created_at":  stringOutput("Creation timestamp"),
		"updated_at":  stringOutput("Version timestamp to pass as expectedUpdatedAt when editing"),
	}, "id", "sequence_id", "title", "team_id", "priority", "created_at", "updated_at")
}

func storyDetailOutputSchema() toolJSONSchema {
	return entityOutput("Updated story", toolJSONSchema{
		"id":         stringOutput("Story UUID"),
		"sequenceId": integerOutput("Team story sequence number"),
		"title":      stringOutput("Story title"),
		"teamId":     stringOutput("Team UUID"),
		"priority":   stringOutput("Story priority"),
		"createdAt":  stringOutput("Creation timestamp"),
		"updatedAt":  stringOutput("Current version timestamp"),
	}, "id", "sequenceId", "title", "teamId", "priority", "createdAt", "updatedAt")
}

func commentOutputSchema() toolJSONSchema {
	return entityOutput("Story comment", toolJSONSchema{
		"comment_id":   stringOutput("Comment UUID"),
		"story_id":     stringOutput("Story UUID"),
		"commenter_id": stringOutput("Author UUID"),
		"content":      stringOutput("Comment text"),
		"created_at":   stringOutput("Creation timestamp"),
		"updated_at":   stringOutput("Last update timestamp"),
	}, "comment_id", "story_id", "commenter_id", "content", "created_at", "updated_at")
}

func sprintOutputSchema() toolJSONSchema {
	return entityOutput("Sprint or iteration", toolJSONSchema{
		"id":        stringOutput("Sprint UUID"),
		"name":      stringOutput("Sprint name"),
		"teamId":    stringOutput("Team UUID"),
		"startDate": stringOutput("Sprint start date"),
		"endDate":   stringOutput("Sprint end date"),
		"createdAt": stringOutput("Creation timestamp"),
		"updatedAt": stringOutput("Last update timestamp"),
	}, "id", "name", "teamId", "startDate", "endDate", "createdAt", "updatedAt")
}

func objectiveOutputSchema() toolJSONSchema {
	return entityOutput("Objective, project, or goal", toolJSONSchema{
		"id":         stringOutput("Objective UUID"),
		"sequenceId": integerOutput("Workspace objective sequence number"),
		"name":       stringOutput("Objective name"),
		"teamId":     stringOutput("Team UUID"),
		"statusId":   stringOutput("Objective status UUID"),
		"isPrivate":  booleanOutput("Whether the objective is private"),
		"createdAt":  stringOutput("Creation timestamp"),
		"updatedAt":  stringOutput("Version timestamp to pass as expectedUpdatedAt when editing"),
	}, "id", "sequenceId", "name", "teamId", "statusId", "isPrivate", "createdAt", "updatedAt")
}

func keyResultOutputSchema() toolJSONSchema {
	return entityOutput("Key result or target", toolJSONSchema{
		"id":              stringOutput("Key-result UUID"),
		"sequenceId":      integerOutput("Key-result sequence number"),
		"objectiveId":     stringOutput("Parent objective UUID"),
		"name":            stringOutput("Key-result name"),
		"measurementType": stringOutput("percentage, number, or boolean"),
		"startValue":      numberOutput("Starting measurement"),
		"currentValue":    numberOutput("Current measurement"),
		"targetValue":     numberOutput("Target measurement"),
		"createdAt":       stringOutput("Creation timestamp"),
		"updatedAt":       stringOutput("Version timestamp to pass as expectedUpdatedAt when editing"),
	}, "id", "sequenceId", "objectiveId", "name", "measurementType", "startValue", "currentValue", "targetValue", "createdAt", "updatedAt")
}

func keyResultListOutputSchema() toolJSONSchema {
	schema := keyResultOutputSchema()
	properties := schema["properties"].(toolJSONSchema)
	properties["objectiveName"] = stringOutput("Parent objective name")
	properties["teamId"] = stringOutput("Team UUID")
	properties["teamName"] = stringOutput("Team name")
	properties["teamCode"] = stringOutput("Team code")
	schema["required"] = append(schema["required"].([]string), "objectiveName", "teamId", "teamName", "teamCode")
	return schema
}

func objectOutput(properties toolJSONSchema, required ...string) toolJSONSchema {
	return toolJSONSchema{
		"type":                 "object",
		"properties":           properties,
		"required":             required,
		"additionalProperties": false,
	}
}

func entityOutput(description string, properties toolJSONSchema, required ...string) toolJSONSchema {
	schema := objectOutput(properties, required...)
	schema["description"] = description
	schema["additionalProperties"] = true
	return schema
}

func stringOutput(description string) toolJSONSchema {
	return toolJSONSchema{"type": "string", "description": description}
}

func integerOutput(description string) toolJSONSchema {
	return toolJSONSchema{"type": "integer", "description": description}
}

func nullableIntegerOutput(description string) toolJSONSchema {
	return toolJSONSchema{"type": []string{"integer", "null"}, "description": description}
}

func numberOutput(description string) toolJSONSchema {
	return toolJSONSchema{"type": "number", "description": description}
}

func booleanOutput(description string) toolJSONSchema {
	return toolJSONSchema{"type": "boolean", "description": description}
}

func arrayOutput(items toolJSONSchema, description string) toolJSONSchema {
	return toolJSONSchema{"type": "array", "items": items, "description": description}
}

func objectValueOutput(description string) toolJSONSchema {
	return toolJSONSchema{"type": "object", "description": description, "additionalProperties": true}
}
