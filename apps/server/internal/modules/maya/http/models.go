package mayahttp

import (
	"encoding/json"
	"time"

	maya "github.com/complexus-tech/projects-api/internal/modules/maya/service"
	"github.com/google/uuid"
)

type AppCreateWorkPlanRequest struct {
	StoryID          uuid.UUID   `json:"storyId"`
	WindowStart      *time.Time  `json:"windowStart"`
	WindowEnd        *time.Time  `json:"windowEnd"`
	DurationMinutes  int         `json:"durationMinutes"`
	CandidateUserIDs []uuid.UUID `json:"candidateUserIds"`
	AutoApply        bool        `json:"autoApply"`
}

type AppWorkPlan struct {
	Run     maya.CoreRun      `json:"run"`
	Actions []maya.CoreAction `json:"actions"`
}

type AppRealtimeSession struct {
	ClientSecret string `json:"clientSecret"`
	ExpiresAt    int64  `json:"expiresAt,omitempty"`
	Model        string `json:"model"`
	Voice        string `json:"voice"`
}

type AppRealtimeToolRequest struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

type AppRealtimeListMyTasksArguments struct {
	IncludeCompleted bool `json:"includeCompleted"`
	Limit            int  `json:"limit"`
}

type AppRealtimeListTeamsArguments struct {
	Search string `json:"search"`
	Limit  int    `json:"limit"`
}

type AppRealtimeListTeamMembersArguments struct {
	TeamName string `json:"teamName"`
	Search   string `json:"search"`
	Limit    int    `json:"limit"`
}

type AppRealtimeSearchArguments struct {
	Query    string `json:"query"`
	Type     string `json:"type"`
	TeamName string `json:"teamName"`
	Limit    int    `json:"limit"`
}

type AppRealtimeListObjectivesArguments struct {
	Search   string `json:"search"`
	TeamName string `json:"teamName"`
	Limit    int    `json:"limit"`
}

type AppRealtimeListKeyResultsArguments struct {
	TeamName string `json:"teamName"`
	Limit    int    `json:"limit"`
}

type AppRealtimeCreateTaskArguments struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	TeamName    string `json:"teamName"`
	Priority    string `json:"priority"`
	Confirmed   bool   `json:"confirmed"`
}

type AppRealtimeToolResponse struct {
	Success              bool                        `json:"success"`
	RequiresConfirmation bool                        `json:"requiresConfirmation,omitempty"`
	NeedsTeam            bool                        `json:"needsTeam,omitempty"`
	Count                int                         `json:"count,omitempty"`
	Message              string                      `json:"message,omitempty"`
	Error                string                      `json:"error,omitempty"`
	Teams                []AppRealtimeVoiceTeam      `json:"teams,omitempty"`
	Stories              []AppRealtimeVoiceStory     `json:"stories,omitempty"`
	Story                *AppRealtimeVoiceStory      `json:"story,omitempty"`
	Objectives           []AppRealtimeVoiceObjective `json:"objectives,omitempty"`
	KeyResults           []AppRealtimeVoiceKeyResult `json:"keyResults,omitempty"`
	Members              []AppRealtimeVoiceMember    `json:"members,omitempty"`
	Terminology          *AppRealtimeTerminology     `json:"terminology,omitempty"`
	Confirmation         *AppRealtimeConfirmation    `json:"confirmation,omitempty"`
}

type AppRealtimeTerminology struct {
	Story      string `json:"story"`
	Stories    string `json:"stories"`
	Sprint     string `json:"sprint"`
	Sprints    string `json:"sprints"`
	Objective  string `json:"objective"`
	Objectives string `json:"objectives"`
	KeyResult  string `json:"keyResult"`
	KeyResults string `json:"keyResults"`
}

type AppRealtimeVoiceTeam struct {
	Name        string `json:"name"`
	Code        string `json:"code,omitempty"`
	MemberCount int    `json:"memberCount,omitempty"`
	IsPrivate   bool   `json:"isPrivate"`
}

type AppRealtimeVoiceStatus struct {
	Name     string `json:"name"`
	Category string `json:"category"`
}

type AppRealtimeVoiceStory struct {
	Reference   string                  `json:"reference,omitempty"`
	Title       string                  `json:"title"`
	Priority    string                  `json:"priority"`
	Team        string                  `json:"team,omitempty"`
	Status      *AppRealtimeVoiceStatus `json:"status,omitempty"`
	StartDate   *time.Time              `json:"startDate,omitempty"`
	EndDate     *time.Time              `json:"endDate,omitempty"`
	CompletedAt *time.Time              `json:"completedAt,omitempty"`
}

type AppRealtimeVoiceObjective struct {
	Name             string     `json:"name"`
	Description      *string    `json:"description,omitempty"`
	Team             string     `json:"team,omitempty"`
	Priority         *string    `json:"priority,omitempty"`
	Health           string     `json:"health,omitempty"`
	StartDate        *time.Time `json:"startDate,omitempty"`
	EndDate          *time.Time `json:"endDate,omitempty"`
	TotalStories     int        `json:"totalStories,omitempty"`
	CompletedStories int        `json:"completedStories,omitempty"`
}

type AppRealtimeVoiceKeyResult struct {
	Name            string     `json:"name"`
	ObjectiveName   string     `json:"objectiveName"`
	Team            string     `json:"team,omitempty"`
	MeasurementType string     `json:"measurementType"`
	StartValue      float64    `json:"startValue"`
	CurrentValue    float64    `json:"currentValue"`
	TargetValue     float64    `json:"targetValue"`
	StartDate       *time.Time `json:"startDate,omitempty"`
	EndDate         *time.Time `json:"endDate,omitempty"`
}

type AppRealtimeVoiceMember struct {
	Name         string     `json:"name"`
	Username     string     `json:"username"`
	Role         string     `json:"role,omitempty"`
	RoleTitle    string     `json:"roleTitle,omitempty"`
	LastActiveAt *time.Time `json:"lastActiveAt,omitempty"`
}

type AppRealtimeConfirmation struct {
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	TeamName    string `json:"teamName,omitempty"`
	Priority    string `json:"priority,omitempty"`
}

type openAIRealtimeClientSecretRequest struct {
	Session openAIRealtimeSessionConfig `json:"session"`
}

type openAIRealtimeSessionConfig struct {
	Type         string                    `json:"type"`
	Model        string                    `json:"model"`
	Instructions string                    `json:"instructions"`
	Audio        openAIRealtimeAudioConfig `json:"audio"`
	Tools        []openAIRealtimeTool      `json:"tools,omitempty"`
	ToolChoice   string                    `json:"tool_choice,omitempty"`
}

type openAIRealtimeAudioConfig struct {
	Output openAIRealtimeAudioOutputConfig `json:"output"`
}

type openAIRealtimeAudioOutputConfig struct {
	Voice string `json:"voice"`
}

type openAIRealtimeTool struct {
	Type        string         `json:"type"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  map[string]any `json:"parameters"`
}

type openAIRealtimeClientSecretResponse struct {
	Value        string                          `json:"value"`
	ExpiresAt    int64                           `json:"expires_at"`
	ClientSecret *openAIRealtimeClientSecretData `json:"client_secret"`
}

type openAIRealtimeClientSecretData struct {
	Value     string `json:"value"`
	ExpiresAt int64  `json:"expires_at"`
}

func toAppWorkPlan(plan maya.WorkPlan) AppWorkPlan {
	return AppWorkPlan{
		Run:     plan.Run,
		Actions: plan.Actions,
	}
}
