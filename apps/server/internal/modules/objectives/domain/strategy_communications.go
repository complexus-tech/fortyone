package domain

import (
	"time"

	"github.com/google/uuid"
)

// StrategyCommunicationCursor identifies the last recipient in the stable
// workspace/user ordering used by strategy communication maintenance reads.
type StrategyCommunicationCursor struct {
	WorkspaceID uuid.UUID
	UserID      uuid.UUID
}

// StrategyCommunicationRecipient is the persistence-neutral identity and
// timezone context needed to evaluate a strategy communication.
type StrategyCommunicationRecipient struct {
	UserID      uuid.UUID
	WorkspaceID uuid.UUID
	Timezone    string
}

// StrategyCommunicationRecipientPage is a bounded recipient page. HasMore is
// derived by the repository from one look-ahead row, so a full terminal page is
// not mistaken for remaining work.
type StrategyCommunicationRecipientPage struct {
	Recipients []StrategyCommunicationRecipient
	HasMore    bool
}

// StrategyCommunicationFoundation contains the planning-period coverage used
// to decide whether an administrator needs a planning reminder.
type StrategyCommunicationFoundation struct {
	HasUltimateGoal bool
	PillarCount     int
	ObjectiveCount  int
}

// StrategyCommunicationMonthlySummary contains the current strategy shape and
// delivery evidence for a completed local calendar month.
type StrategyCommunicationMonthlySummary struct {
	PillarCount          int
	PillarsNeedingReview int
	ObjectiveCount       int
	AtRiskObjectives     int
	UnalignedObjectives  int
	KeyResultCount       int
	KeyResultProgress    *float64
	CompletedStories     int
}

// StrategyWeeklySignalCursor identifies the last raw weekly signal row. The
// ordering fields are normalized by SQL so nullable key-result columns remain
// safe to resume without OFFSET pagination.
type StrategyWeeklySignalCursor struct {
	ObjectiveID       uuid.UUID
	KeyResultNullRank int
	KeyResultID       uuid.UUID
}

// StrategyWeeklySignalRecord is one objective signal and, when present, one
// stale incomplete key result visible to the objective lead.
type StrategyWeeklySignalRecord struct {
	Recipient                StrategyCommunicationRecipient
	ObjectiveID              uuid.UUID
	TeamID                   uuid.UUID
	ObjectiveName            string
	ObjectiveHealth          *string
	ObjectiveStatusID        *uuid.UUID
	ObjectiveStatusName      *string
	ObjectiveStatusCategory  *string
	ObjectiveStartDate       *time.Time
	ObjectiveEndDate         *time.Time
	ObjectiveUpdatedAt       time.Time
	IsStaleObjective         bool
	IsAtRiskObjective        bool
	KeyResultID              *uuid.UUID
	KeyResultName            *string
	KeyResultMeasurementType *string
	KeyResultStartValue      *float64
	KeyResultCurrentValue    *float64
	KeyResultTargetValue     *float64
	KeyResultStartDate       *time.Time
	KeyResultEndDate         *time.Time
	KeyResultUpdatedAt       *time.Time
}

// StrategyWeeklySignalPage is a bounded signal page. NextCursor is present
// exactly when HasMore is true.
type StrategyWeeklySignalPage struct {
	Records    []StrategyWeeklySignalRecord
	NextCursor *StrategyWeeklySignalCursor
	HasMore    bool
}
