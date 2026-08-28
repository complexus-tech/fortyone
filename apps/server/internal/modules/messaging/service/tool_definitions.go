package messaging

import (
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/google/uuid"
)

func fortyOneToolDefinitions() []ToolDefinition {
	nullableLimit := func() map[string]any {
		return map[string]any{
			"type":        []string{"integer", "null"},
			"description": "Maximum number of results, or null for the default.",
			"minimum":     1,
			"maximum":     maxToolLimit,
		}
	}
	nullableTeamID := func() map[string]any {
		return map[string]any{
			"type":        []string{"string", "null"},
			"description": "A team UUID returned by list_teams, or null for all joined teams.",
		}
	}

	return []ToolDefinition{
		{
			Type:        "function",
			Name:        toolListTeams,
			Description: "List only the FortyOne teams the current user has joined, including whether shared cross-assignee work is authorized in this conversation.",
			Strict:      true,
			Parameters:  strictObjectSchema(map[string]any{}, []string{}),
		},
		{
			Type:        "function",
			Name:        toolListMyTasks,
			Description: "List active tasks assigned to the current user across only their joined teams. Completed, cancelled, deleted, and archived work is excluded.",
			Strict:      true,
			Parameters: strictObjectSchema(map[string]any{
				"limit": nullableLimit(),
			}, []string{"limit"}),
		},
		{
			Type:        "function",
			Name:        toolSearchWork,
			Description: "Search task and objective titles within only the current user's joined teams.",
			Strict:      true,
			Parameters: strictObjectSchema(map[string]any{
				"query": map[string]any{
					"type":        "string",
					"description": "Plain-language text to search for.",
					"minLength":   1,
					"maxLength":   maxSearchRunes,
				},
				"team_id": nullableTeamID(),
				"kind": map[string]any{
					"type":        []string{"string", "null"},
					"description": "Limit the search to stories or objectives, or null/all for both.",
					"enum":        []any{"all", "stories", "objectives", nil},
				},
				"limit": nullableLimit(),
			}, []string{"query", "team_id", "kind", "limit"}),
		},
		{
			Type:        "function",
			Name:        toolListObjectives,
			Description: "List objectives within only the current user's joined teams.",
			Strict:      true,
			Parameters: strictObjectSchema(map[string]any{
				"team_id": nullableTeamID(),
				"query": map[string]any{
					"type":        []string{"string", "null"},
					"description": "Optional objective-name search text, or null.",
					"maxLength":   maxSearchRunes,
				},
				"limit": nullableLimit(),
			}, []string{"team_id", "query", "limit"}),
		},
	}
}

func completedTaskToolDefinitions() []ToolDefinition {
	nullableDate := func(description string) map[string]any {
		return map[string]any{
			"type":        []string{"string", "null"},
			"description": description,
			"pattern":     "^\\d{4}-\\d{2}-\\d{2}$",
		}
	}
	nullableLimit := map[string]any{
		"type":        []string{"integer", "null"},
		"description": "Maximum number of results, or null for the default.",
		"minimum":     1,
		"maximum":     maxToolLimit,
	}

	return []ToolDefinition{
		{
			Type:        "function",
			Name:        toolListCompleted,
			Description: "List tasks assigned to the current user that were completed within a local calendar date range. Omit both dates to use today.",
			Strict:      true,
			Parameters: strictObjectSchema(map[string]any{
				"start_date": nullableDate("Local start date in YYYY-MM-DD format, or null for today."),
				"end_date":   nullableDate("Local end date in YYYY-MM-DD format, or null for the start date."),
				"limit":      nullableLimit,
			}, []string{"start_date", "end_date", "limit"}),
		},
	}
}

func operationalToolDefinitions(includeStoryReader bool) []ToolDefinition {
	nullableLimit := func() map[string]any {
		return map[string]any{
			"type":        []string{"integer", "null"},
			"description": "Maximum number of results, or null for the default.",
			"minimum":     1,
			"maximum":     maxToolLimit,
		}
	}
	nullableTeamID := func() map[string]any {
		return map[string]any{
			"type":        []string{"string", "null"},
			"description": "A team UUID returned by list_teams, or null for every team visible in this conversation.",
		}
	}

	definitions := []ToolDefinition{
		{
			Type:        "function",
			Name:        toolListStatuses,
			Description: "List human-readable task statuses from only the teams visible to the current user in this conversation.",
			Strict:      true,
			Parameters: strictObjectSchema(map[string]any{
				"team_id": nullableTeamID(),
				"limit":   nullableLimit(),
			}, []string{"team_id", "limit"}),
		},
		{
			Type:        "function",
			Name:        toolListTeamMembers,
			Description: "List active human members of one team visible to the current user in this conversation. Results contain no email addresses.",
			Strict:      true,
			Parameters: strictObjectSchema(map[string]any{
				"team_id": map[string]any{
					"type":        "string",
					"description": "An exact team UUID returned by list_teams.",
				},
				"query": map[string]any{
					"type":        []string{"string", "null"},
					"description": "Optional display-name or username search text, or null.",
					"maxLength":   maxSearchRunes,
				},
				"limit": nullableLimit(),
			}, []string{"team_id", "query", "limit"}),
		},
	}
	if includeStoryReader {
		definitions = append(definitions, ToolDefinition{
			Type:        "function",
			Name:        toolGetStory,
			Description: "Get current details for one active task by its human-readable reference, such as WEB-123, only when its team is visible in this conversation.",
			Strict:      true,
			Parameters: strictObjectSchema(map[string]any{
				"story_reference": map[string]any{
					"type":        "string",
					"description": "An exact human-readable task reference such as WEB-123.",
					"minLength":   1,
					"maxLength":   maxStoryReferenceRunes,
				},
			}, []string{"story_reference"}),
		})
	}
	return definitions
}

func strictObjectSchema(properties map[string]any, required []string) map[string]any {
	return map[string]any{
		"type":                 "object",
		"properties":           properties,
		"required":             required,
		"additionalProperties": false,
	}
}

type teamResult struct {
	ID                uuid.UUID `json:"id"`
	Name              string    `json:"name"`
	Code              string    `json:"code"`
	IsPrivate         bool      `json:"is_private"`
	MemberCount       int       `json:"member_count"`
	SprintsEnabled    bool      `json:"sprints_enabled"`
	SharedWorkEnabled bool      `json:"shared_work_enabled"`
}

type listTeamsResult struct {
	Total     int          `json:"total"`
	Truncated bool         `json:"truncated"`
	Teams     []teamResult `json:"teams"`
}

type taskResult struct {
	ID                       uuid.UUID  `json:"id"`
	Reference                string     `json:"reference"`
	URL                      string     `json:"url,omitempty"`
	Title                    string     `json:"title"`
	TeamID                   uuid.UUID  `json:"team_id"`
	TeamName                 string     `json:"team_name"`
	TeamCode                 string     `json:"team_code"`
	StatusID                 *uuid.UUID `json:"status_id"`
	StatusName               string     `json:"status_name"`
	StatusCategory           string     `json:"status_category"`
	AssigneeID               *uuid.UUID `json:"assignee_id"`
	AssigneeName             string     `json:"assignee_name"`
	AssigneeUsername         string     `json:"assignee_username"`
	Priority                 string     `json:"priority"`
	EstimateLabel            *string    `json:"estimate_label"`
	EstimateValue            *int16     `json:"estimate_value"`
	EstimateScheme           string     `json:"estimate_scheme"`
	EstimatedDurationMinutes *int       `json:"estimated_duration_minutes"`
	MinimumFocusBlockMinutes *int       `json:"minimum_focus_block_minutes"`
	AutoSchedulingEnabled    bool       `json:"auto_scheduling_enabled"`
	AutoSchedulingLocked     bool       `json:"auto_scheduling_locked"`
	AutoSchedulingStatus     string     `json:"auto_scheduling_status"`
	AutoSchedulingReason     *string    `json:"auto_scheduling_reason,omitempty"`
	AutoSchedulingUpdatedAt  *time.Time `json:"auto_scheduling_updated_at,omitempty"`
	EndDate                  *time.Time `json:"end_date"`
	CompletedAt              *time.Time `json:"completed_at,omitempty"`
	UpdatedAt                time.Time  `json:"updated_at"`
}

type listTasksResult struct {
	Total     int          `json:"total"`
	Truncated bool         `json:"truncated"`
	Tasks     []taskResult `json:"tasks"`
}

type searchStoryResult struct {
	ID                       uuid.UUID  `json:"id"`
	Reference                string     `json:"reference"`
	URL                      string     `json:"url,omitempty"`
	Title                    string     `json:"title"`
	TeamID                   uuid.UUID  `json:"team_id"`
	StatusID                 *uuid.UUID `json:"status_id"`
	Priority                 string     `json:"priority"`
	EstimateLabel            *string    `json:"estimate_label"`
	EstimateValue            *int16     `json:"estimate_value"`
	EstimateScheme           string     `json:"estimate_scheme"`
	EstimatedDurationMinutes *int       `json:"estimated_duration_minutes"`
	MinimumFocusBlockMinutes *int       `json:"minimum_focus_block_minutes"`
	AutoSchedulingEnabled    bool       `json:"auto_scheduling_enabled"`
	AutoSchedulingLocked     bool       `json:"auto_scheduling_locked"`
	AutoSchedulingStatus     string     `json:"auto_scheduling_status"`
	AutoSchedulingReason     *string    `json:"auto_scheduling_reason,omitempty"`
	AutoSchedulingUpdatedAt  *time.Time `json:"auto_scheduling_updated_at,omitempty"`
	UpdatedAt                time.Time  `json:"updated_at"`
}

func storyURL(scope ToolScope, reference string) string {
	base, err := url.Parse(strings.TrimRight(strings.TrimSpace(scope.WebsiteURL), "/"))
	if err != nil || base.Hostname() == "" || strings.TrimSpace(scope.WorkspaceSlug) == "" || strings.TrimSpace(reference) == "" {
		return ""
	}
	base.Path = path.Join("/", scope.WorkspaceSlug, "work", reference)
	if !strings.EqualFold(base.Hostname(), "localhost") && !strings.EqualFold(base.Hostname(), "127.0.0.1") {
		base.Path = path.Join("/", "work", reference)
		if !strings.HasPrefix(base.Hostname(), scope.WorkspaceSlug+".") {
			base.Host = scope.WorkspaceSlug + "." + base.Host
		}
	}
	return base.String()
}

type searchObjectiveResult struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	ShortSummary *string   `json:"short_summary"`
	TeamID       uuid.UUID `json:"team_id"`
	StatusID     uuid.UUID `json:"status_id"`
	Priority     *string   `json:"priority"`
	Health       *string   `json:"health"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type searchWorkResult struct {
	Stories    []searchStoryResult     `json:"stories"`
	Objectives []searchObjectiveResult `json:"objectives"`
}

type objectiveResult struct {
	ID               uuid.UUID  `json:"id"`
	SequenceID       int        `json:"sequence_id"`
	Name             string     `json:"name"`
	ShortSummary     *string    `json:"short_summary"`
	TeamID           uuid.UUID  `json:"team_id"`
	StatusID         uuid.UUID  `json:"status_id"`
	Priority         *string    `json:"priority"`
	Health           *string    `json:"health"`
	StartDate        *time.Time `json:"start_date"`
	EndDate          *time.Time `json:"end_date"`
	KeyResultCount   int        `json:"key_result_count"`
	TotalStories     int        `json:"total_stories"`
	CompletedStories int        `json:"completed_stories"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

type listObjectivesResult struct {
	Count      int               `json:"count"`
	Objectives []objectiveResult `json:"objectives"`
}

type statusResult struct {
	ID         uuid.UUID `json:"id"`
	Name       string    `json:"name"`
	Category   string    `json:"category"`
	OrderIndex int       `json:"order_index"`
	IsDefault  bool      `json:"is_default"`
	TeamID     uuid.UUID `json:"team_id"`
	TeamName   string    `json:"team_name"`
	TeamCode   string    `json:"team_code"`
}

type listStatusesResult struct {
	Total     int            `json:"total"`
	Truncated bool           `json:"truncated"`
	Statuses  []statusResult `json:"statuses"`
}

type teamMemberResult struct {
	ID          uuid.UUID `json:"id"`
	DisplayName string    `json:"display_name"`
	Username    string    `json:"username"`
	Active      bool      `json:"active"`
	RoleTitle   string    `json:"role_title"`
}

type listTeamMembersResult struct {
	TeamName  string             `json:"team_name"`
	TeamCode  string             `json:"team_code"`
	Total     int                `json:"total"`
	Truncated bool               `json:"truncated"`
	Members   []teamMemberResult `json:"members"`
}

type storyDetailsResult struct {
	ID                       uuid.UUID  `json:"id"`
	Reference                string     `json:"reference"`
	URL                      string     `json:"url,omitempty"`
	Title                    string     `json:"title"`
	Description              *string    `json:"description"`
	DescriptionTruncated     bool       `json:"description_truncated"`
	TeamID                   uuid.UUID  `json:"team_id"`
	TeamName                 string     `json:"team_name"`
	TeamCode                 string     `json:"team_code"`
	StatusID                 *uuid.UUID `json:"status_id"`
	StatusName               string     `json:"status_name"`
	StatusCategory           string     `json:"status_category"`
	AssigneeID               *uuid.UUID `json:"assignee_id"`
	AssigneeName             string     `json:"assignee_name"`
	AssigneeUsername         string     `json:"assignee_username"`
	Priority                 string     `json:"priority"`
	EstimateLabel            *string    `json:"estimate_label"`
	EstimateValue            *int16     `json:"estimate_value"`
	EstimateScheme           string     `json:"estimate_scheme"`
	EstimatedDurationMinutes *int       `json:"estimated_duration_minutes"`
	MinimumFocusBlockMinutes *int       `json:"minimum_focus_block_minutes"`
	AutoSchedulingEnabled    bool       `json:"auto_scheduling_enabled"`
	AutoSchedulingLocked     bool       `json:"auto_scheduling_locked"`
	AutoSchedulingStatus     string     `json:"auto_scheduling_status"`
	AutoSchedulingReason     *string    `json:"auto_scheduling_reason,omitempty"`
	AutoSchedulingUpdatedAt  *time.Time `json:"auto_scheduling_updated_at,omitempty"`
	SprintName               *string    `json:"sprint_name"`
	StartDate                *time.Time `json:"start_date"`
	EndDate                  *time.Time `json:"end_date"`
	CompletedAt              *time.Time `json:"completed_at"`
	UpdatedAt                time.Time  `json:"updated_at"`
}
