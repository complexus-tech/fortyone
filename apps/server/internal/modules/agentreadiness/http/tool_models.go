package agentreadinesshttp

type emptyInput struct{}

type workspaceInput struct {
	WorkspaceID string `json:"workspaceId"`
}
type teamListInput struct {
	WorkspaceID string `json:"workspaceId"`
	JoinedOnly  bool   `json:"joinedOnly,omitempty"`
}
type storyListInput struct {
	WorkspaceID  string `json:"workspaceId"`
	TeamID       string `json:"teamId,omitempty"`
	SprintID     string `json:"sprintId,omitempty"`
	ObjectiveID  string `json:"objectiveId,omitempty"`
	AssigneeID   string `json:"assigneeId,omitempty"`
	AssignedToMe bool   `json:"assignedToMe,omitempty" jsonschema:"Only stories assigned to the connected user"`
	DueOn        string `json:"dueOn,omitempty" jsonschema:"Only stories with this end date, formatted as YYYY-MM-DD"`
	StatusID     string `json:"statusId,omitempty"`
	KeyResultID  string `json:"keyResultId,omitempty"`
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
	IdempotencyKey           string   `json:"idempotencyKey,omitempty"`
	Confirmed                bool     `json:"confirmed" jsonschema:"True only after the user approves creation"`
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
}
type objectiveListInput struct {
	WorkspaceID string `json:"workspaceId"`
	TeamID      string `json:"teamId,omitempty"`
	Search      string `json:"search,omitempty"`
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
type analysisInput struct {
	WorkspaceID  string   `json:"workspaceId"`
	TeamIDs      []string `json:"teamIds,omitempty"`
	AssigneeIDs  []string `json:"assigneeIds,omitempty"`
	SprintIDs    []string `json:"sprintIds,omitempty"`
	ObjectiveIDs []string `json:"objectiveIds,omitempty"`
	StartDate    string   `json:"startDate,omitempty" jsonschema:"YYYY-MM-DD"`
	EndDate      string   `json:"endDate,omitempty" jsonschema:"YYYY-MM-DD"`
}
