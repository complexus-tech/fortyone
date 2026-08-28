package notifications

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

type NotificationMessage struct {
	Template  string                        `json:"template"`
	Variables map[string]Variable           `json:"variables"`
	Strategy  *StrategyNotificationSnapshot `json:"strategy,omitempty"`
}

func (message NotificationMessage) Public() NotificationMessage {
	message.Strategy = nil
	return message
}

type Variable struct {
	Value string `json:"value"`
	Type  string `json:"type"`
}

const (
	StrategyNotificationKindPlanningReminder = "planning_reminder"
	StrategyNotificationKindWeeklyCheckIn    = "weekly_check_in"
	StrategyNotificationKindMonthlySummary   = "monthly_summary"

	StrategySignalReasonAtRisk     = "at_risk"
	StrategySignalReasonStale      = "stale"
	StrategySignalReasonIncomplete = "incomplete"

	StrategyMissingElementUltimateGoal = "ultimate_goal"
	StrategyMissingElementPillars      = "strategic_pillars"
	StrategyMissingElementObjectives   = "objectives"
)

type StrategyNotificationSnapshot struct {
	Version        int                             `json:"version"`
	Kind           string                          `json:"kind"`
	GeneratedAt    time.Time                       `json:"generatedAt"`
	Planning       *StrategyPlanningSnapshot       `json:"planning,omitempty"`
	WeeklyCheckIn  *StrategyWeeklyCheckInSnapshot  `json:"weeklyCheckIn,omitempty"`
	MonthlySummary *StrategyMonthlySummarySnapshot `json:"monthlySummary,omitempty"`
}

type StrategyPlanningSnapshot struct {
	Period          string    `json:"period"`
	StartsAt        time.Time `json:"startsAt"`
	DaysUntil       int       `json:"daysUntil"`
	HasUltimateGoal bool      `json:"hasUltimateGoal"`
	PillarCount     int       `json:"pillarCount"`
	ObjectiveCount  int       `json:"objectiveCount"`
	MissingElements []string  `json:"missingElements"`
}

type StrategyWeeklyCheckInSnapshot struct {
	StaleAfterDays int                                          `json:"staleAfterDays"`
	Counts         StrategyWeeklyCheckInCounts                  `json:"counts"`
	TeamCounts     []StrategyWeeklyCheckInTeamCountsSnapshot    `json:"teamCounts,omitempty"`
	Objectives     []StrategyObjectiveSnapshot                  `json:"objectives"`
	KeyResults     []StrategyKeyResultSnapshot                  `json:"keyResults"`
	OmittedDetails *StrategyWeeklyCheckInOmittedDetailsSnapshot `json:"omittedDetails,omitempty"`
}

type StrategyWeeklyCheckInTeamCountsSnapshot struct {
	TeamID         uuid.UUID                                    `json:"teamId"`
	Counts         StrategyWeeklyCheckInCounts                  `json:"counts"`
	OmittedDetails *StrategyWeeklyCheckInOmittedDetailsSnapshot `json:"omittedDetails,omitempty"`
}

type StrategyWeeklyCheckInOmittedDetailsSnapshot struct {
	Objectives int `json:"objectives"`
	KeyResults int `json:"keyResults"`
}

type StrategyWeeklyCheckInCounts struct {
	AtRiskObjectives int `json:"atRiskObjectives"`
	StaleObjectives  int `json:"staleObjectives"`
	StaleKeyResults  int `json:"staleKeyResults"`
	UniqueObjectives int `json:"uniqueObjectives"`
}

type StrategyObjectiveSnapshot struct {
	ID        uuid.UUID                        `json:"id"`
	TeamID    uuid.UUID                        `json:"teamId"`
	Name      string                           `json:"name"`
	Health    *string                          `json:"health,omitempty"`
	Status    *StrategyObjectiveStatusSnapshot `json:"status,omitempty"`
	StartDate *time.Time                       `json:"startDate,omitempty"`
	EndDate   *time.Time                       `json:"endDate,omitempty"`
	UpdatedAt time.Time                        `json:"updatedAt"`
	Reasons   []string                         `json:"reasons"`
}

type StrategyObjectiveStatusSnapshot struct {
	ID       uuid.UUID `json:"id"`
	Name     string    `json:"name"`
	Category string    `json:"category"`
}

type StrategyKeyResultSnapshot struct {
	ID              uuid.UUID                        `json:"id"`
	ObjectiveID     uuid.UUID                        `json:"objectiveId"`
	TeamID          uuid.UUID                        `json:"teamId"`
	Name            string                           `json:"name"`
	ObjectiveName   string                           `json:"objectiveName"`
	ObjectiveHealth *string                          `json:"objectiveHealth,omitempty"`
	ObjectiveStatus *StrategyObjectiveStatusSnapshot `json:"objectiveStatus,omitempty"`
	MeasurementType string                           `json:"measurementType"`
	StartValue      *float64                         `json:"startValue,omitempty"`
	CurrentValue    *float64                         `json:"currentValue,omitempty"`
	TargetValue     *float64                         `json:"targetValue,omitempty"`
	StartDate       *time.Time                       `json:"startDate,omitempty"`
	EndDate         *time.Time                       `json:"endDate,omitempty"`
	UpdatedAt       time.Time                        `json:"updatedAt"`
	Reasons         []string                         `json:"reasons"`
}

type StrategyMonthlySummarySnapshot struct {
	PeriodStart          time.Time `json:"periodStart"`
	PeriodEnd            time.Time `json:"periodEnd"`
	PillarCount          int       `json:"pillarCount"`
	PillarsNeedingReview int       `json:"pillarsNeedingReview"`
	ObjectiveCount       int       `json:"objectiveCount"`
	AtRiskObjectives     int       `json:"atRiskObjectives"`
	UnalignedObjectives  int       `json:"unalignedObjectives"`
	KeyResultCount       int       `json:"keyResultCount"`
	KeyResultProgress    *float64  `json:"keyResultProgress,omitempty"`
	CompletedStories     int       `json:"completedStories"`
}

type NewNotification struct {
	DedupeKey   string    `json:"-"`
	RecipientID uuid.UUID `json:"recipient_id"`
	WorkspaceID uuid.UUID `json:"workspace_id"`
	// InAppEnabled is an explicit delivery override. Nil resolves the current
	// typed preference in the same statement that inserts the notification.
	InAppEnabled *bool               `json:"-"`
	Type         NotificationType    `json:"type"`
	EntityType   EntityType          `json:"entity_type"`
	EntityID     uuid.UUID           `json:"entity_id"`
	ActorID      uuid.UUID           `json:"actor_id"`
	Title        string              `json:"title"`
	Message      NotificationMessage `json:"message"`
}

func (notification NewNotification) Validate() error {
	if notification.RecipientID == uuid.Nil || notification.WorkspaceID == uuid.Nil ||
		notification.EntityID == uuid.Nil || notification.ActorID == uuid.Nil {
		return fmt.Errorf("%w: notification IDs are required", ErrInvalid)
	}
	if !notification.Type.Valid() || !notification.EntityType.Valid() ||
		!notification.Type.SupportsEntity(notification.EntityType) {
		return fmt.Errorf("%w: notification type and entity type are incompatible", ErrInvalid)
	}
	if strings.TrimSpace(notification.Title) == "" {
		return fmt.Errorf("%w: notification title is required", ErrInvalid)
	}
	return nil
}

type Notification struct {
	ID           uuid.UUID           `json:"id"`
	RecipientID  uuid.UUID           `json:"recipient_id"`
	WorkspaceID  uuid.UUID           `json:"workspace_id"`
	InAppEnabled bool                `json:"-"`
	Type         NotificationType    `json:"type"`
	EntityType   EntityType          `json:"entity_type"`
	EntityID     uuid.UUID           `json:"entity_id"`
	ActorID      uuid.UUID           `json:"actor_id"`
	Actor        *NotificationActor  `json:"actor,omitempty"`
	Title        string              `json:"title"`
	Message      NotificationMessage `json:"message"`
	CreatedAt    time.Time           `json:"created_at"`
	ReadAt       *time.Time          `json:"read_at"`
}

type NotificationActor struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	FullName  string    `json:"full_name"`
	AvatarURL string    `json:"avatar_url"`
	IsActive  bool      `json:"is_active"`
	IsSystem  bool      `json:"is_system"`
}

func (notification Notification) Public() Notification {
	if notification.EntityType == EntityTypeStrategy && notification.Message.Strategy != nil {
		notification.Message.Template = "Strategy guidance is ready to review."
		notification.Message.Variables = map[string]Variable{}
	}
	notification.Message = notification.Message.Public()
	return notification
}

type PortalNotification struct {
	Notification  Notification
	ActorName     string
	ActorAvatar   *string
	FeedbackTitle string
	FeedbackSlug  string
}

type Preferences struct {
	ID          uuid.UUID     `json:"id"`
	UserID      uuid.UUID     `json:"user_id"`
	WorkspaceID uuid.UUID     `json:"workspace_id"`
	Preferences PreferenceSet `json:"preferences"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
}
