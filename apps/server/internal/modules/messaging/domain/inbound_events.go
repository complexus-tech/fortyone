package messagingdomain

import (
	"time"

	"github.com/google/uuid"
)

type InboundEventInput struct {
	Provider            string
	WorkspaceID         *uuid.UUID
	InstallGeneration   *uuid.UUID
	ExternalWorkspaceID string
	ExternalEventID     string
	EventType           string
	PayloadEncrypted    string
}

type InboundEventRecord struct {
	ID                  uuid.UUID
	WorkspaceID         *uuid.UUID
	InstallGeneration   *uuid.UUID
	ExternalWorkspaceID string
	ExternalEventID     string
	Status              string
	AttemptCount        int
	RecoveryGeneration  int
	ProcessedAt         *time.Time
	PayloadEncrypted    *string
}
