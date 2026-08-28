package taskhandlers

import (
	"encoding/json"
	"time"

	"github.com/complexus-tech/projects-api/pkg/emailcopy"
	"github.com/complexus-tech/projects-api/pkg/mailer"
	"github.com/google/uuid"
)

const (
	maxNotificationMessageRunes     = 600
	maxNotificationDigestRows       = 12
	maxNotificationDigestDetailRows = maxNotificationDigestRows - 1
)

const (
	digestActionPrimary  = "digest_primary"
	digestActionStrategy = "strategy"
)

type Variable struct {
	Value string `json:"value"`
	Type  string `json:"type"`
}

type NotificationMessage struct {
	Template  string                        `json:"template"`
	Variables map[string]Variable           `json:"variables"`
	Strategy  *strategyNotificationSnapshot `json:"strategy,omitempty"`
}

type strategyNotificationSnapshot struct {
	Version        int                             `json:"version"`
	Kind           string                          `json:"kind"`
	GeneratedAt    time.Time                       `json:"generatedAt"`
	Planning       *strategyPlanningSnapshot       `json:"planning,omitempty"`
	WeeklyCheckIn  *strategyWeeklyCheckInSnapshot  `json:"weeklyCheckIn,omitempty"`
	MonthlySummary *strategyMonthlySummarySnapshot `json:"monthlySummary,omitempty"`
}

type strategyPlanningSnapshot struct {
	Period          string    `json:"period"`
	StartsAt        time.Time `json:"startsAt"`
	DaysUntil       int       `json:"daysUntil"`
	HasUltimateGoal bool      `json:"hasUltimateGoal"`
	PillarCount     int       `json:"pillarCount"`
	ObjectiveCount  int       `json:"objectiveCount"`
	MissingElements []string  `json:"missingElements"`
}

type strategyWeeklyCheckInSnapshot struct {
	StaleAfterDays int                                          `json:"staleAfterDays"`
	Counts         strategyWeeklyCheckInCounts                  `json:"counts"`
	TeamCounts     []strategyWeeklyCheckInTeamCountsSnapshot    `json:"teamCounts,omitempty"`
	Objectives     []strategyObjectiveSnapshot                  `json:"objectives"`
	KeyResults     []strategyKeyResultSnapshot                  `json:"keyResults"`
	OmittedDetails *strategyWeeklyCheckInOmittedDetailsSnapshot `json:"omittedDetails,omitempty"`
}

type strategyWeeklyCheckInTeamCountsSnapshot struct {
	TeamID         uuid.UUID                                    `json:"teamId"`
	Counts         strategyWeeklyCheckInCounts                  `json:"counts"`
	OmittedDetails *strategyWeeklyCheckInOmittedDetailsSnapshot `json:"omittedDetails,omitempty"`
}

type strategyWeeklyCheckInOmittedDetailsSnapshot struct {
	Objectives int `json:"objectives"`
	KeyResults int `json:"keyResults"`
}

type strategyWeeklyCheckInCounts struct {
	AtRiskObjectives int `json:"atRiskObjectives"`
	StaleObjectives  int `json:"staleObjectives"`
	StaleKeyResults  int `json:"staleKeyResults"`
	UniqueObjectives int `json:"uniqueObjectives"`
}

type strategyObjectiveSnapshot struct {
	ID        uuid.UUID                        `json:"id"`
	TeamID    uuid.UUID                        `json:"teamId"`
	Name      string                           `json:"name"`
	Health    *string                          `json:"health,omitempty"`
	Status    *strategyObjectiveStatusSnapshot `json:"status,omitempty"`
	StartDate *time.Time                       `json:"startDate,omitempty"`
	EndDate   *time.Time                       `json:"endDate,omitempty"`
	UpdatedAt time.Time                        `json:"updatedAt"`
	Reasons   []string                         `json:"reasons"`
}

type strategyObjectiveStatusSnapshot struct {
	ID       uuid.UUID `json:"id"`
	Name     string    `json:"name"`
	Category string    `json:"category"`
}

type strategyKeyResultSnapshot struct {
	ID              uuid.UUID                        `json:"id"`
	ObjectiveID     uuid.UUID                        `json:"objectiveId"`
	TeamID          uuid.UUID                        `json:"teamId"`
	Name            string                           `json:"name"`
	ObjectiveName   string                           `json:"objectiveName"`
	ObjectiveHealth *string                          `json:"objectiveHealth,omitempty"`
	ObjectiveStatus *strategyObjectiveStatusSnapshot `json:"objectiveStatus,omitempty"`
	MeasurementType string                           `json:"measurementType"`
	StartValue      *float64                         `json:"startValue,omitempty"`
	CurrentValue    *float64                         `json:"currentValue,omitempty"`
	TargetValue     *float64                         `json:"targetValue,omitempty"`
	StartDate       *time.Time                       `json:"startDate,omitempty"`
	EndDate         *time.Time                       `json:"endDate,omitempty"`
	UpdatedAt       time.Time                        `json:"updatedAt"`
	Reasons         []string                         `json:"reasons"`
}

type strategyMonthlySummarySnapshot struct {
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

// NotificationEmailData represents all data needed for sending notification emails
type NotificationEmailData struct {
	NotificationID   uuid.UUID       `db:"notification_id"`
	RecipientID      uuid.UUID       `db:"recipient_id"`
	WorkspaceID      uuid.UUID       `db:"workspace_id"`
	NotificationType string          `db:"type"`
	EntityType       string          `db:"entity_type"`
	EntityID         uuid.UUID       `db:"entity_id"`
	Title            string          `db:"title"`
	Message          json.RawMessage `db:"message"`
	UserEmail        string          `db:"user_email"`
	UserName         string          `db:"user_name"`
	ActorName        string          `db:"actor_name"`
	WorkspaceName    string          `db:"workspace_name"`
	WorkspaceSlug    string          `db:"workspace_slug"`
	WorkspaceRole    string          `db:"workspace_role"`
	EmailEnabled     bool            `db:"email_enabled"`
	FeedbackSlug     string          `db:"feedback_slug"`
}

type NotificationEmailDigestItem struct {
	NotificationID   uuid.UUID       `db:"notification_id"`
	NotificationType string          `db:"type"`
	EntityType       string          `db:"entity_type"`
	EntityID         uuid.UUID       `db:"entity_id"`
	Title            string          `db:"title"`
	Message          json.RawMessage `db:"message"`
	CreatedAt        time.Time       `db:"created_at"`
	ActorName        string          `db:"actor_name"`
	FeedbackSlug     string          `db:"feedback_slug"`
}

type NotificationEmailDigestData struct {
	RecipientID   uuid.UUID
	WorkspaceID   uuid.UUID
	UserEmail     string
	UserName      string
	WorkspaceName string
	WorkspaceSlug string
	WorkspaceRole string
	Items         []NotificationEmailDigestItem
}

type notificationDigestCopy struct {
	Subject             string
	Heading             string
	Intro               string
	Rows                []notificationDigestCopyRow
	CTA                 notificationDigestCopyCTA
	Sender              mailer.SenderProfile
	NotificationsURL    string
	HasStrategySnapshot bool
}

type notificationDigestCopyRow struct {
	Text  string
	Label string
	URL   string
}

type notificationDigestCopyCTA struct {
	Label string
	URL   string
}

type notificationDigestCopyInput struct {
	Request             emailcopy.Request
	Actions             map[string]string
	FactActions         map[string]string
	FactLabels          map[string]string
	Fallback            notificationDigestCopy
	HasStrategySnapshot bool
	NotificationsURL    string
}

// ParsedMessage represents the final parsed notification message
type ParsedMessage struct {
	Text string
	HTML string
}
