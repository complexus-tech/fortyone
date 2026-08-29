package slack

import (
	"github.com/complexus-tech/projects-api/internal/platform/webhooks"
	"github.com/google/uuid"
)

// inboundEventRecord is the Slack-owned processing identity passed to the
// assistant and delivery use cases. Durable state stays in the shared webhook
// inbox; this value intentionally contains no provider payload.
type inboundEventRecord struct {
	ID                uuid.UUID
	WorkspaceID       *uuid.UUID
	InstallationID    *uuid.UUID
	InstallGeneration *uuid.UUID
	AttemptCount      int
}

func inboundReceipt(record webhooks.Record) inboundEventRecord {
	return inboundEventRecord{
		ID:                record.ID,
		WorkspaceID:       optionalUUID(record.WorkspaceID),
		InstallationID:    optionalUUID(record.InstallationID),
		InstallGeneration: optionalUUID(record.InstallationGeneration),
		AttemptCount:      int(record.AttemptCount),
	}
}

func optionalUUID(value uuid.UUID) *uuid.UUID {
	if value == uuid.Nil {
		return nil
	}
	copy := value
	return &copy
}
