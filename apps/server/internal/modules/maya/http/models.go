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

type AppRealtimeCreateTaskArguments struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	TeamName    string `json:"teamName"`
	Priority    string `json:"priority"`
	Confirmed   bool   `json:"confirmed"`
}

type AppRealtimeToolResponse struct {
	Success              bool                     `json:"success"`
	RequiresConfirmation bool                     `json:"requiresConfirmation,omitempty"`
	NeedsTeam            bool                     `json:"needsTeam,omitempty"`
	Count                int                      `json:"count,omitempty"`
	Message              string                   `json:"message,omitempty"`
	Error                string                   `json:"error,omitempty"`
	Teams                []AppRealtimeVoiceTeam   `json:"teams,omitempty"`
	Stories              []AppRealtimeVoiceStory  `json:"stories,omitempty"`
	Story                *AppRealtimeVoiceStory   `json:"story,omitempty"`
	Confirmation         *AppRealtimeConfirmation `json:"confirmation,omitempty"`
}

type AppRealtimeVoiceTeam struct {
	Name string `json:"name"`
	Code string `json:"code,omitempty"`
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
