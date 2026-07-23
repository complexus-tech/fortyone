package mayahttp

import (
	"encoding/json"
	"errors"
	"strings"
	"time"
	"unicode/utf8"

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

type AppRealtimeSessionRequest struct {
	CurrentPath string                           `json:"currentPath"`
	Messages    []AppRealtimeConversationMessage `json:"messages"`
}

type AppRealtimeConversationMessage struct {
	Role string `json:"role"`
	Text string `json:"text"`
}

func (r AppRealtimeSessionRequest) Validate() error {
	if utf8.RuneCountInString(r.CurrentPath) > 512 {
		return errors.New("currentPath must be 512 characters or fewer")
	}
	if len(r.Messages) > 24 {
		return errors.New("messages must contain 24 items or fewer")
	}
	for _, message := range r.Messages {
		if message.Role != "user" && message.Role != "assistant" {
			return errors.New("message role must be user or assistant")
		}
		if utf8.RuneCountInString(message.Text) > 4_000 {
			return errors.New("message text must be 4000 characters or fewer")
		}
	}
	return nil
}

type AppRealtimeEndSessionRequest struct {
	SessionID uuid.UUID `json:"sessionId"`
}

type AppRealtimeToolRequest struct {
	SessionID uuid.UUID       `json:"sessionId"`
	CallID    string          `json:"callId"`
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

func (r AppRealtimeToolRequest) Validate() error {
	if r.SessionID == uuid.Nil {
		return errors.New("sessionId is required")
	}
	if strings.TrimSpace(r.CallID) == "" {
		return errors.New("callId is required")
	}
	if utf8.RuneCountInString(r.CallID) > 128 {
		return errors.New("callId must be 128 characters or fewer")
	}
	if strings.TrimSpace(r.Name) == "" {
		return errors.New("name is required")
	}
	if utf8.RuneCountInString(r.Name) > 80 {
		return errors.New("name must be 80 characters or fewer")
	}
	if len(r.Arguments) > 32_000 {
		return errors.New("arguments must be 32000 bytes or fewer")
	}
	return nil
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
	Title             string `json:"title"`
	Description       string `json:"description"`
	TeamName          string `json:"teamName"`
	AssigneeName      string `json:"assigneeName"`
	AssignToMe        bool   `json:"assignToMe"`
	Priority          string `json:"priority"`
	EstimateValue     *int16 `json:"estimateValue"`
	StartDate         string `json:"startDate"`
	EndDate           string `json:"endDate"`
	BlockedByRef      string `json:"blockedByRef"`
	BlockingRef       string `json:"blockingRef"`
	RelatedRef        string `json:"relatedRef"`
	Confirmed         bool   `json:"confirmed"`
	ConfirmationToken string `json:"confirmationToken"`
}

type AppRealtimeNavigateArguments struct {
	TargetType string `json:"targetType"`
	Reference  string `json:"reference"`
	TeamName   string `json:"teamName"`
	Route      string `json:"route"`
}

type AppRealtimeSetThemeArguments struct {
	Theme string `json:"theme"`
}

type AppRealtimeStoryArguments struct {
	Reference string `json:"reference"`
}

type AppRealtimeUpdateStoryArguments struct {
	Reference         string `json:"reference"`
	Title             string `json:"title"`
	Status            string `json:"status"`
	Priority          string `json:"priority"`
	AssigneeName      string `json:"assigneeName"`
	AssignToMe        bool   `json:"assignToMe"`
	Unassign          bool   `json:"unassign"`
	EstimateValue     *int16 `json:"estimateValue"`
	ClearEstimate     bool   `json:"clearEstimate"`
	StartDate         string `json:"startDate"`
	ClearStartDate    bool   `json:"clearStartDate"`
	EndDate           string `json:"endDate"`
	ClearEndDate      bool   `json:"clearEndDate"`
	SprintName        string `json:"sprintName"`
	ClearSprint       bool   `json:"clearSprint"`
	ObjectiveName     string `json:"objectiveName"`
	ClearObjective    bool   `json:"clearObjective"`
	Confirmed         bool   `json:"confirmed"`
	ConfirmationToken string `json:"confirmationToken"`
}

type AppRealtimeCommentsArguments struct {
	Action            string `json:"action"`
	Reference         string `json:"reference"`
	Comment           string `json:"comment"`
	Limit             int    `json:"limit"`
	Confirmed         bool   `json:"confirmed"`
	ConfirmationToken string `json:"confirmationToken"`
}

type AppRealtimeSprintArguments struct {
	Action   string `json:"action"`
	Name     string `json:"name"`
	TeamName string `json:"teamName"`
	Limit    int    `json:"limit"`
}

type AppRealtimeWorkloadArguments struct {
	TeamName string `json:"teamName"`
	Limit    int    `json:"limit"`
}

type AppRealtimeActivityArguments struct {
	Days  int `json:"days"`
	Limit int `json:"limit"`
}

type AppRealtimeNotificationsArguments struct {
	Action            string `json:"action"`
	Title             string `json:"title"`
	Limit             int    `json:"limit"`
	Confirmed         bool   `json:"confirmed"`
	ConfirmationToken string `json:"confirmationToken"`
}

type AppRealtimeFeedbackArguments struct {
	Action   string `json:"action"`
	Title    string `json:"title"`
	TeamName string `json:"teamName"`
	Status   string `json:"status"`
	Limit    int    `json:"limit"`
}

type AppRealtimeWorkspaceBriefingArguments struct {
	Days int `json:"days"`
}

type AppRealtimeToolResponse struct {
	Success              bool                           `json:"success"`
	RequiresConfirmation bool                           `json:"requiresConfirmation,omitempty"`
	NeedsTeam            bool                           `json:"needsTeam,omitempty"`
	NeedsAssignee        bool                           `json:"needsAssignee,omitempty"`
	NeedsStoryReference  bool                           `json:"needsStoryReference,omitempty"`
	Count                int                            `json:"count,omitempty"`
	Message              string                         `json:"message,omitempty"`
	Error                string                         `json:"error,omitempty"`
	Teams                []AppRealtimeVoiceTeam         `json:"teams,omitempty"`
	Stories              []AppRealtimeVoiceStory        `json:"stories,omitempty"`
	Story                *AppRealtimeVoiceStory         `json:"story,omitempty"`
	Objectives           []AppRealtimeVoiceObjective    `json:"objectives,omitempty"`
	KeyResults           []AppRealtimeVoiceKeyResult    `json:"keyResults,omitempty"`
	Members              []AppRealtimeVoiceMember       `json:"members,omitempty"`
	User                 *AppRealtimeVoiceUser          `json:"user,omitempty"`
	Terminology          *AppRealtimeTerminology        `json:"terminology,omitempty"`
	Confirmation         *AppRealtimeConfirmation       `json:"confirmation,omitempty"`
	ConfirmationToken    string                         `json:"confirmationToken,omitempty"`
	ClientAction         *AppRealtimeClientAction       `json:"clientAction,omitempty"`
	Sprints              []AppRealtimeVoiceSprint       `json:"sprints,omitempty"`
	Comments             []AppRealtimeVoiceComment      `json:"comments,omitempty"`
	Activities           []AppRealtimeVoiceActivity     `json:"activities,omitempty"`
	Notifications        []AppRealtimeVoiceNotification `json:"notifications,omitempty"`
	FeedbackItems        []AppRealtimeVoiceFeedbackItem `json:"feedbackItems,omitempty"`
	Workload             *AppRealtimeVoiceWorkload      `json:"workload,omitempty"`
	Briefing             *AppRealtimeVoiceBriefing      `json:"briefing,omitempty"`
}

type AppRealtimeClientAction struct {
	Type  string `json:"type"`
	Path  string `json:"path,omitempty"`
	Theme string `json:"theme,omitempty"`
}

type AppRealtimeVoiceSprint struct {
	Name                 string    `json:"name"`
	Team                 string    `json:"team"`
	Goal                 string    `json:"goal,omitempty"`
	StartDate            time.Time `json:"startDate"`
	EndDate              time.Time `json:"endDate"`
	TotalStories         int       `json:"totalStories"`
	CompletedStories     int       `json:"completedStories"`
	StartedStories       int       `json:"startedStories"`
	CompletionPercentage int       `json:"completionPercentage"`
	Status               string    `json:"status,omitempty"`
}

type AppRealtimeVoiceComment struct {
	Author    string    `json:"author"`
	Comment   string    `json:"comment"`
	CreatedAt time.Time `json:"createdAt"`
}

type AppRealtimeVoiceActivity struct {
	Actor     string    `json:"actor"`
	StoryRef  string    `json:"storyRef,omitempty"`
	Type      string    `json:"type"`
	Field     string    `json:"field,omitempty"`
	Value     string    `json:"value,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
}

type AppRealtimeVoiceNotification struct {
	Title      string    `json:"title"`
	Type       string    `json:"type"`
	EntityType string    `json:"entityType"`
	IsRead     bool      `json:"isRead"`
	CreatedAt  time.Time `json:"createdAt"`
}

type AppRealtimeVoiceFeedbackItem struct {
	Title          string    `json:"title"`
	Description    string    `json:"description,omitempty"`
	Status         string    `json:"status"`
	Team           string    `json:"team,omitempty"`
	Author         string    `json:"author,omitempty"`
	RoadmapSummary string    `json:"roadmapSummary,omitempty"`
	VoteCount      int       `json:"voteCount"`
	CommentCount   int       `json:"commentCount"`
	CreatedAt      time.Time `json:"createdAt"`
}

type AppRealtimeVoiceWorkloadMember struct {
	Name           string `json:"name"`
	OpenStories    int    `json:"openStories"`
	OverdueStories int    `json:"overdueStories"`
	UrgentStories  int    `json:"urgentStories"`
	EstimateTotal  int    `json:"estimateTotal"`
}

type AppRealtimeVoiceWorkload struct {
	TotalOpenStories   int                              `json:"totalOpenStories"`
	TotalEstimate      int                              `json:"totalEstimate"`
	OverdueStories     int                              `json:"overdueStories"`
	UnassignedStories  int                              `json:"unassignedStories"`
	UnestimatedStories int                              `json:"unestimatedStories"`
	AtRiskMembers      []AppRealtimeVoiceWorkloadMember `json:"atRiskMembers,omitempty"`
}

type AppRealtimeVoiceBriefing struct {
	TotalStories      int `json:"totalStories"`
	CompletedStories  int `json:"completedStories"`
	ActiveObjectives  int `json:"activeObjectives"`
	ActiveSprints     int `json:"activeSprints"`
	TeamMembers       int `json:"teamMembers"`
	OverdueStories    int `json:"overdueStories"`
	BlockedStories    int `json:"blockedStories"`
	AtRiskSprints     int `json:"atRiskSprints"`
	AtRiskObjectives  int `json:"atRiskObjectives"`
	FeedbackItems     int `json:"feedbackItems"`
	UnreadFeedback    int `json:"unreadFeedback"`
	OverloadedMembers int `json:"overloadedMembers"`
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
	Description   string                  `json:"description,omitempty"`
	Priority      string                  `json:"priority"`
	EstimateLabel *string                 `json:"estimateLabel,omitempty"`
	EstimateValue *int16                  `json:"estimateValue,omitempty"`
	Team          string                  `json:"team,omitempty"`
	Assignee      string                  `json:"assignee,omitempty"`
	Sprint        string                  `json:"sprint,omitempty"`
	Objective     string                  `json:"objective,omitempty"`
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
	Type             string                    `json:"type"`
	Model            string                    `json:"model"`
	Instructions     string                    `json:"instructions"`
	MaxOutputTokens  int                       `json:"max_output_tokens,omitempty"`
	OutputModalities []string                  `json:"output_modalities,omitempty"`
	Audio            openAIRealtimeAudioConfig `json:"audio"`
	Tools            []openAIRealtimeTool      `json:"tools,omitempty"`
	ToolChoice       string                    `json:"tool_choice,omitempty"`
}

type openAIRealtimeAudioConfig struct {
	Input  openAIRealtimeAudioInputConfig  `json:"input"`
	Output openAIRealtimeAudioOutputConfig `json:"output"`
}

type openAIRealtimeAudioInputConfig struct {
	NoiseReduction openAIRealtimeNoiseReductionConfig `json:"noise_reduction"`
	Transcription  openAIRealtimeTranscriptionConfig  `json:"transcription"`
	TurnDetection  openAIRealtimeTurnDetectionConfig  `json:"turn_detection"`
}

type openAIRealtimeNoiseReductionConfig struct {
	Type string `json:"type"`
}

type openAIRealtimeTranscriptionConfig struct {
	Language string `json:"language"`
	Model    string `json:"model"`
	Prompt   string `json:"prompt"`
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
