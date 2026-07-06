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
	ClientSecret        string    `json:"clientSecret"`
	SessionID           uuid.UUID `json:"sessionId"`
	ExpiresAt           int64     `json:"expiresAt,omitempty"`
	Model               string    `json:"model"`
	Voice               string    `json:"voice"`
	MaxSessionSeconds   int       `json:"maxSessionSeconds"`
	RemainingSeconds    int       `json:"remainingSeconds"`
	MonthlyLimitSeconds int       `json:"monthlyLimitSeconds"`
}

type AppRealtimeEndSessionRequest struct {
	SessionID uuid.UUID `json:"sessionId"`
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
	Title         string `json:"title"`
	Description   string `json:"description"`
	TeamName      string `json:"teamName"`
	AssigneeName  string `json:"assigneeName"`
	AssignToMe    bool   `json:"assignToMe"`
	Priority      string `json:"priority"`
	EstimateValue *int16 `json:"estimateValue"`
	StartDate     string `json:"startDate"`
	EndDate       string `json:"endDate"`
	BlockedByRef  string `json:"blockedByRef"`
	BlockingRef   string `json:"blockingRef"`
	RelatedRef    string `json:"relatedRef"`
	Confirmed     bool   `json:"confirmed"`
}

type AppRealtimeToolResponse struct {
	Success              bool                        `json:"success"`
	RequiresConfirmation bool                        `json:"requiresConfirmation,omitempty"`
	NeedsTeam            bool                        `json:"needsTeam,omitempty"`
	NeedsAssignee        bool                        `json:"needsAssignee,omitempty"`
	NeedsStoryReference  bool                        `json:"needsStoryReference,omitempty"`
	Count                int                         `json:"count,omitempty"`
	Message              string                      `json:"message,omitempty"`
	Error                string                      `json:"error,omitempty"`
	Teams                []AppRealtimeVoiceTeam      `json:"teams,omitempty"`
	Stories              []AppRealtimeVoiceStory     `json:"stories,omitempty"`
	Story                *AppRealtimeVoiceStory      `json:"story,omitempty"`
	Objectives           []AppRealtimeVoiceObjective `json:"objectives,omitempty"`
	KeyResults           []AppRealtimeVoiceKeyResult `json:"keyResults,omitempty"`
	Members              []AppRealtimeVoiceMember    `json:"members,omitempty"`
	User                 *AppRealtimeVoiceUser       `json:"user,omitempty"`
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
	Reference     string                  `json:"reference,omitempty"`
	Title         string                  `json:"title"`
	Priority      string                  `json:"priority"`
	EstimateLabel *string                 `json:"estimateLabel,omitempty"`
	EstimateValue *int16                  `json:"estimateValue,omitempty"`
	Team          string                  `json:"team,omitempty"`
	Assignee      string                  `json:"assignee,omitempty"`
	Status        *AppRealtimeVoiceStatus `json:"status,omitempty"`
	BlockedBy     string                  `json:"blockedBy,omitempty"`
	Blocking      string                  `json:"blocking,omitempty"`
	Related       string                  `json:"related,omitempty"`
	StartDate     *time.Time              `json:"startDate,omitempty"`
	EndDate       *time.Time              `json:"endDate,omitempty"`
	CompletedAt   *time.Time              `json:"completedAt,omitempty"`
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

type AppRealtimeVoiceUser struct {
	Name     string `json:"name"`
	Username string `json:"username"`
	Timezone string `json:"timezone"`
	Today    string `json:"today"`
	Now      string `json:"now"`
}

type AppRealtimeConfirmation struct {
	Title         string `json:"title"`
	Description   string `json:"description,omitempty"`
	TeamName      string `json:"teamName,omitempty"`
	AssigneeName  string `json:"assigneeName,omitempty"`
	Priority      string `json:"priority,omitempty"`
	EstimateValue *int16 `json:"estimateValue,omitempty"`
	StartDate     string `json:"startDate,omitempty"`
	EndDate       string `json:"endDate,omitempty"`
	BlockedByRef  string `json:"blockedByRef,omitempty"`
	BlockingRef   string `json:"blockingRef,omitempty"`
	RelatedRef    string `json:"relatedRef,omitempty"`
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
	Input  openAIRealtimeAudioInputConfig  `json:"input"`
	Output openAIRealtimeAudioOutputConfig `json:"output"`
}

type openAIRealtimeAudioInputConfig struct {
	TurnDetection openAIRealtimeTurnDetectionConfig `json:"turn_detection"`
}

type openAIRealtimeTurnDetectionConfig struct {
	Type              string  `json:"type"`
	Threshold         float64 `json:"threshold"`
	PrefixPaddingMs   int     `json:"prefix_padding_ms"`
	SilenceDurationMs int     `json:"silence_duration_ms"`
	CreateResponse    bool    `json:"create_response"`
	InterruptResponse bool    `json:"interrupt_response"`
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
