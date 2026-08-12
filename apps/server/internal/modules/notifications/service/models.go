package notifications

import (
	"time"

	"github.com/google/uuid"
)

type NotificationMessage struct {
	Template  string                        `json:"template"`
	Variables map[string]Variable           `json:"variables"`
	Strategy  *StrategyNotificationSnapshot `json:"strategy,omitempty"`
}

// Public returns the product-facing message without structured delivery
// evidence. Delivery evidence is persisted for delayed email rendering and may
// contain access-scoped objective or key-result snapshots; it is never part of
// the browser notification contract.
func (message NotificationMessage) Public() NotificationMessage {
	message.Strategy = nil
	return message
}

type Variable struct {
	Value string `json:"value"`
	Type  string `json:"type"` // "actor", "assignee", "field", "value", "date"
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

// StrategyNotificationSnapshot carries a versioned, send-time snapshot of the
// facts behind a strategy notification. The optional pointer keeps existing
// notification messages wire-compatible while allowing email and AI renderers
// to use structured facts instead of reverse-engineering prose.
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

// StrategyWeeklyCheckInOmittedDetailsSnapshot records how many signal details
// were deliberately left out of the bounded notification snapshot. Aggregate
// counts remain complete even when the detail payload is capped.
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

// CoreNewNotification represents a new notification to be created.
type CoreNewNotification struct {
	DedupeKey   string              `json:"-"`
	RecipientID uuid.UUID           `json:"recipient_id"`
	WorkspaceID uuid.UUID           `json:"workspace_id"`
	Type        string              `json:"type"`
	EntityType  string              `json:"entity_type"`
	EntityID    uuid.UUID           `json:"entity_id"`
	ActorID     uuid.UUID           `json:"actor_id"`
	Title       string              `json:"title"`
	Message     NotificationMessage `json:"message"`
}

// CoreNotification represents a notification.
type CoreNotification struct {
	ID          uuid.UUID           `json:"id"`
	RecipientID uuid.UUID           `json:"recipient_id"`
	WorkspaceID uuid.UUID           `json:"workspace_id"`
	Type        string              `json:"type"`
	EntityType  string              `json:"entity_type"`
	EntityID    uuid.UUID           `json:"entity_id"`
	ActorID     uuid.UUID           `json:"actor_id"`
	Title       string              `json:"title"`
	Message     NotificationMessage `json:"message"`
	CreatedAt   time.Time           `json:"created_at"`
	ReadAt      *time.Time          `json:"read_at"`
}

// Public returns a value copy safe for HTTP and realtime clients.
func (notification CoreNotification) Public() CoreNotification {
	// Strategy prose is generated from delivery-only snapshots. Exposing the
	// stored summary after a team or role change can reveal aggregate facts the
	// recipient can no longer access, even when the structured snapshot itself
	// is removed. The title remains the neutral in-app affordance.
	if notification.EntityType == "strategy" && notification.Message.Strategy != nil {
		notification.Message.Template = "Strategy guidance is ready to review."
		notification.Message.Variables = map[string]Variable{}
	}
	notification.Message = notification.Message.Public()
	return notification
}

type CorePortalNotification struct {
	Notification  CoreNotification
	ActorName     string
	ActorAvatar   *string
	FeedbackTitle string
	FeedbackSlug  string
}

// CoreNotificationPreferences represents a user's notification preferences.
type CoreNotificationPreferences struct {
	ID          uuid.UUID              `json:"id"`
	UserID      uuid.UUID              `json:"user_id"`
	WorkspaceID uuid.UUID              `json:"workspace_id"`
	Preferences map[string]interface{} `json:"preferences"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// CoreNotificationPreference represents a single notification preference
// Legacy model - kept for backward compatibility
type CoreNotificationPreference struct {
	ID           uuid.UUID
	UserID       uuid.UUID
	WorkspaceID  uuid.UUID
	Type         string
	EmailEnabled bool
	InAppEnabled bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// NotificationChannels represents the different delivery channels for a notification type
type NotificationChannels struct {
	Email bool `json:"email"`
	InApp bool `json:"in_app"`
}
