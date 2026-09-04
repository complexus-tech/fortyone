package domain

import (
	"time"

	"github.com/google/uuid"
)

const (
	ImportOperationPending   = "pending"
	ImportOperationCompleted = "completed"
	ImportOperationFailed    = "failed"
)

// ImportOperation is the durable idempotency record for a Google Doc snapshot.
// DocumentID is allocated before provider I/O so every attempt for one key
// converges on the same native document identity.
type ImportOperation struct {
	ID                uuid.UUID
	WorkspaceID       uuid.UUID
	UserID            uuid.UUID
	SourceReferenceID uuid.UUID
	DocumentID        uuid.UUID
	IdempotencyKey    string
	RequestHash       string
	Visibility        string
	AttemptGeneration uuid.UUID
	Status            string
	CreatedAt         time.Time
	UpdatedAt         time.Time
	CompletedAt       *time.Time
}

type ImportFinalization struct {
	Operation       ImportOperation
	AccountID       uuid.UUID
	GrantGeneration uuid.UUID
	TargetType      TargetType
	TargetID        uuid.UUID
	GoogleFileID    string
	SourceVersion   *string
	Title           string
	ContentHTML     string
	ContentText     string
}
