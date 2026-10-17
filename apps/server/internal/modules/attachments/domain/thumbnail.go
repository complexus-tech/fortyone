package attachmentdomain

import (
	"time"

	"github.com/google/uuid"
)

// Thumbnail identifies a private, versioned derivative. It does not grant
// access: callers must authorize the original attachment before reading it.
type Thumbnail struct {
	ID            uuid.UUID
	AttachmentID  uuid.UUID
	WorkspaceID   uuid.UUID
	SourceVersion string
	Provider      string
	Container     string
	BlobName      string
	Status        string
	RetryAfter    *time.Time
	ByteSize      *int32
}
