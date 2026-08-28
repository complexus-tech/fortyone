package domain

import (
	"time"

	"github.com/google/uuid"
)

// StoryAutoArchiveBatch is one bounded auto-archive transaction evaluated
// against the application-owned UTC clock.
type StoryAutoArchiveBatch struct {
	AsOf      time.Time
	BatchSize int
}

type StoryAutoArchiveResult struct {
	Archived int
}

// StoryAutoCloseBatch is one bounded transition-and-activity transaction.
// SystemUserID is the durable automation actor recorded on every activity.
type StoryAutoCloseBatch struct {
	AsOf         time.Time
	SystemUserID uuid.UUID
	BatchSize    int
}

type StoryAutoCloseResult struct {
	Closed             int
	ActivitiesRecorded int
}

// SprintStoryMigrationBatch is one bounded sprint transition transaction.
// Each migrated story receives both its activity and audit event before the
// transaction can commit.
type SprintStoryMigrationBatch struct {
	AsOf         time.Time
	SystemUserID uuid.UUID
	BatchSize    int
}

type SprintStoryMigrationResult struct {
	Migrated            int
	ActivitiesRecorded  int
	AuditEventsRecorded int
}
