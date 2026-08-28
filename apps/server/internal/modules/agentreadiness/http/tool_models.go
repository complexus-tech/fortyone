package agentreadinesshttp

type emptyInput struct{}

type workspaceListInput struct {
	Page     int `json:"page,omitempty" jsonschema:"Page number, starting at 1"`
	PageSize int `json:"pageSize,omitempty" jsonschema:"Results per page, from 1 to 100"`
}

type teamListInput struct {
	WorkspaceID string `json:"workspaceId"`
	JoinedOnly  bool   `json:"joinedOnly,omitempty"`
	Search      string `json:"search,omitempty"`
	Page        int    `json:"page,omitempty" jsonschema:"Page number, starting at 1"`
	PageSize    int    `json:"pageSize,omitempty" jsonschema:"Results per page, from 1 to 100"`
}

type storyStatusListInput struct {
	WorkspaceID string `json:"workspaceId"`
	TeamID      string `json:"teamId"`
	Page        int    `json:"page,omitempty"`
	PageSize    int    `json:"pageSize,omitempty"`
}

type objectiveStatusListInput struct {
	WorkspaceID string `json:"workspaceId"`
	Page        int    `json:"page,omitempty"`
	PageSize    int    `json:"pageSize,omitempty"`
}
type storyListInput struct {
	WorkspaceID  string `json:"workspaceId"`
	TeamID       string `json:"teamId,omitempty"`
	SprintID     string `json:"sprintId,omitempty"`
	ObjectiveID  string `json:"objectiveId,omitempty"`
	AssigneeID   string `json:"assigneeId,omitempty"`
	AssignedToMe bool   `json:"assignedToMe,omitempty" jsonschema:"Only stories assigned to the connected user"`
	DueOn        string `json:"dueOn,omitempty" jsonschema:"Only stories with this end date, formatted as YYYY-MM-DD"`
	Search       string `json:"search,omitempty" jsonschema:"Case-insensitive title search"`
	StatusID     string `json:"statusId,omitempty"`
	KeyResultID  string `json:"keyResultId,omitempty"`
	Page         int    `json:"page,omitempty" jsonschema:"Page number, starting at 1"`
	PageSize     int    `json:"pageSize,omitempty" jsonschema:"Results per page, from 1 to 100"`
}
type createStoryInput struct {
	WorkspaceID              string   `json:"workspaceId"`
	TeamID                   string   `json:"teamId"`
	Title                    string   `json:"title"`
	Description              string   `json:"description,omitempty"`
	StatusID                 string   `json:"statusId,omitempty"`
	AssigneeID               string   `json:"assigneeId,omitempty"`
	Priority                 string   `json:"priority,omitempty" jsonschema:"No Priority, Low, Medium, High, or Urgent"`
	EstimateValue            *int16   `json:"estimateValue,omitempty"`
	EstimatedDurationMinutes *int     `json:"estimatedDurationMinutes,omitempty" jsonschema:"Total focused work duration in minutes"`
	MinimumFocusBlockMinutes *int     `json:"minimumFocusBlockMinutes,omitempty" jsonschema:"Minimum calendar focus block in minutes"`
	AutoSchedulingEnabled    bool     `json:"autoSchedulingEnabled,omitempty"`
	StartDate                string   `json:"startDate,omitempty" jsonschema:"YYYY-MM-DD"`
	EndDate                  string   `json:"endDate,omitempty" jsonschema:"YYYY-MM-DD"`
	SprintID                 string   `json:"sprintId,omitempty"`
	ObjectiveID              string   `json:"objectiveId,omitempty"`
	KeyResultID              string   `json:"keyResultId,omitempty"`
	ParentID                 string   `json:"parentId,omitempty"`
	LabelIDs                 []string `json:"labelIds,omitempty"`
	IdempotencyKey           string   `json:"idempotencyKey" jsonschema:"Opaque unique value generated for this approved creation; reuse only when retrying the same request"`
	Confirmed                bool     `json:"confirmed" jsonschema:"True only after the user approves creation"`
}

type updateStoryInput struct {
	WorkspaceID              string  `json:"workspaceId"`
	ID                       string  `json:"id" jsonschema:"Story UUID returned by list_stories"`
	ExpectedUpdatedAt        string  `json:"expectedUpdatedAt" jsonschema:"RFC3339 updated_at value returned by list_stories; prevents overwriting newer edits"`
	Title                    *string `json:"title,omitempty"`
	Description              *string `json:"description,omitempty"`
	StatusID                 *string `json:"statusId,omitempty"`
	AssigneeID               *string `json:"assigneeId,omitempty" jsonschema:"User UUID, or an empty string to clear"`
	Priority                 *string `json:"priority,omitempty" jsonschema:"No Priority, Low, Medium, High, or Urgent"`
	EstimateValue            *int16  `json:"estimateValue,omitempty" jsonschema:"Team estimate value; this is not a duration"`
	EstimatedDurationMinutes *int    `json:"estimatedDurationMinutes,omitempty" jsonschema:"Total focused work duration in minutes; use 360 for six hours"`
	MinimumFocusBlockMinutes *int    `json:"minimumFocusBlockMinutes,omitempty"`
	AutoSchedulingEnabled    *bool   `json:"autoSchedulingEnabled,omitempty"`
	StartDate                *string `json:"startDate,omitempty" jsonschema:"YYYY-MM-DD, or an empty string to clear"`
	EndDate                  *string `json:"endDate,omitempty" jsonschema:"YYYY-MM-DD, or an empty string to clear"`
	SprintID                 *string `json:"sprintId,omitempty" jsonschema:"Sprint UUID, or an empty string to clear"`
	ObjectiveID              *string `json:"objectiveId,omitempty" jsonschema:"Objective UUID, or an empty string to clear"`
	KeyResultID              *string `json:"keyResultId,omitempty" jsonschema:"Key-result UUID, or an empty string to clear"`
	ParentID                 *string `json:"parentId,omitempty" jsonschema:"Parent story UUID, or an empty string to clear"`
	Confirmed                bool    `json:"confirmed" jsonschema:"True only after the user approves the exact changes"`
}

type storyCommentListInput struct {
	WorkspaceID string `json:"workspaceId"`
	StoryID     string `json:"storyId"`
	Page        int    `json:"page,omitempty"`
	PageSize    int    `json:"pageSize,omitempty"`
}

type addStoryCommentInput struct {
	WorkspaceID string `json:"workspaceId"`
	StoryID     string `json:"storyId"`
	Comment     string `json:"comment"`
	ParentID    string `json:"parentId,omitempty"`
	Confirmed   bool   `json:"confirmed"`
}

type setStoryArchivedInput struct {
	WorkspaceID string `json:"workspaceId"`
	ID          string `json:"id"`
	Archived    bool   `json:"archived" jsonschema:"True to archive; false to restore from the archive"`
	Confirmed   bool   `json:"confirmed"`
}
type createSprintInput struct {
	WorkspaceID string `json:"workspaceId"`
	TeamID      string `json:"teamId"`
	Name        string `json:"name"`
	Goal        string `json:"goal,omitempty"`
	ObjectiveID string `json:"objectiveId,omitempty"`
	StartDate   string `json:"startDate" jsonschema:"YYYY-MM-DD"`
	EndDate     string `json:"endDate" jsonschema:"YYYY-MM-DD"`
	Confirmed   bool   `json:"confirmed"`
}

type entityInput struct {
	WorkspaceID string `json:"workspaceId"`
	ID          string `json:"id"`
	Page        int    `json:"page,omitempty" jsonschema:"Detail page number, starting at 1"`
	PageSize    int    `json:"pageSize,omitempty" jsonschema:"Detail rows per section, from 1 to 100"`
}
type objectiveListInput struct {
	WorkspaceID string `json:"workspaceId"`
	TeamID      string `json:"teamId,omitempty"`
	Search      string `json:"search,omitempty"`
	Page        int    `json:"page,omitempty" jsonschema:"Page number, starting at 1"`
	PageSize    int    `json:"pageSize,omitempty" jsonschema:"Results per page, from 1 to 100"`
}
type createObjectiveInput struct {
	WorkspaceID string `json:"workspaceId"`
	TeamID      string `json:"teamId"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	LeadUserID  string `json:"leadUserId,omitempty"`
	StatusID    string `json:"statusId,omitempty"`
	Priority    string `json:"priority,omitempty"`
	StartDate   string `json:"startDate,omitempty" jsonschema:"YYYY-MM-DD"`
	EndDate     string `json:"endDate,omitempty" jsonschema:"YYYY-MM-DD"`
	IsPrivate   bool   `json:"isPrivate,omitempty"`
	Confirmed   bool   `json:"confirmed"`
}

type updateObjectiveInput struct {
	WorkspaceID       string  `json:"workspaceId"`
	ID                string  `json:"id" jsonschema:"Objective UUID returned by list_objectives"`
	ExpectedUpdatedAt string  `json:"expectedUpdatedAt" jsonschema:"RFC3339 UpdatedAt value returned by list_objectives"`
	Name              *string `json:"name,omitempty"`
	Description       *string `json:"description,omitempty"`
	LeadUserID        *string `json:"leadUserId,omitempty"`
	StatusID          *string `json:"statusId,omitempty"`
	Priority          *string `json:"priority,omitempty"`
	Health            *string `json:"health,omitempty" jsonschema:"On Track, At Risk, or Off Track"`
	StartDate         *string `json:"startDate,omitempty" jsonschema:"YYYY-MM-DD"`
	EndDate           *string `json:"endDate,omitempty" jsonschema:"YYYY-MM-DD"`
	IsPrivate         *bool   `json:"isPrivate,omitempty"`
	Comment           string  `json:"comment,omitempty"`
	Confirmed         bool    `json:"confirmed"`
}
type keyResultListInput struct {
	WorkspaceID  string   `json:"workspaceId"`
	ObjectiveIDs []string `json:"objectiveIds,omitempty"`
	Page         int      `json:"page,omitempty"`
	PageSize     int      `json:"pageSize,omitempty"`
}
type createKeyResultInput struct {
	WorkspaceID     string   `json:"workspaceId"`
	ObjectiveID     string   `json:"objectiveId"`
	Name            string   `json:"name"`
	MeasurementType string   `json:"measurementType" jsonschema:"percentage, number, or boolean"`
	StartValue      float64  `json:"startValue"`
	CurrentValue    float64  `json:"currentValue"`
	TargetValue     float64  `json:"targetValue"`
	LeadUserID      string   `json:"leadUserId,omitempty"`
	ContributorIDs  []string `json:"contributorIds,omitempty"`
	StartDate       string   `json:"startDate,omitempty" jsonschema:"YYYY-MM-DD"`
	EndDate         string   `json:"endDate,omitempty" jsonschema:"YYYY-MM-DD"`
	Confirmed       bool     `json:"confirmed"`
}

type updateKeyResultInput struct {
	WorkspaceID       string   `json:"workspaceId"`
	ID                string   `json:"id" jsonschema:"Key-result UUID returned by list_key_results"`
	ExpectedUpdatedAt string   `json:"expectedUpdatedAt" jsonschema:"RFC3339 UpdatedAt value returned by list_key_results"`
	Name              *string  `json:"name,omitempty"`
	MeasurementType   *string  `json:"measurementType,omitempty" jsonschema:"percentage, number, or boolean"`
	StartValue        *float64 `json:"startValue,omitempty"`
	CurrentValue      *float64 `json:"currentValue,omitempty"`
	TargetValue       *float64 `json:"targetValue,omitempty"`
	LeadUserID        *string  `json:"leadUserId,omitempty" jsonschema:"User UUID, or an empty string to clear"`
	StartDate         *string  `json:"startDate,omitempty" jsonschema:"YYYY-MM-DD; key-result dates cannot be cleared"`
	EndDate           *string  `json:"endDate,omitempty" jsonschema:"YYYY-MM-DD; key-result dates cannot be cleared"`
	Comment           string   `json:"comment,omitempty"`
	Confirmed         bool     `json:"confirmed"`
}
type analysisInput struct {
	WorkspaceID  string   `json:"workspaceId"`
	TeamIDs      []string `json:"teamIds,omitempty"`
	AssigneeIDs  []string `json:"assigneeIds,omitempty"`
	SprintIDs    []string `json:"sprintIds,omitempty"`
	ObjectiveIDs []string `json:"objectiveIds,omitempty"`
	StartDate    string   `json:"startDate,omitempty" jsonschema:"YYYY-MM-DD"`
	EndDate      string   `json:"endDate,omitempty" jsonschema:"YYYY-MM-DD"`
	Page         int      `json:"page,omitempty" jsonschema:"Detail page number, starting at 1"`
	PageSize     int      `json:"pageSize,omitempty" jsonschema:"Detail rows per section, from 1 to 100"`
}
