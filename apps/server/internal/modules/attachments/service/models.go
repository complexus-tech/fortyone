package attachments

import (
	"time"

	"github.com/google/uuid"
)

// CoreAttachment represents an attachment in the core layer
type CoreAttachment struct {
	ID          uuid.UUID
	Filename    string
	BlobName    string
	Size        int64
	MimeType    string
	UploadedBy  uuid.UUID
	WorkspaceID uuid.UUID
	CreatedAt   time.Time
}

// CoreNewAttachment represents a new attachment
type CoreNewAttachment struct {
	Filename    string
	BlobName    string
	Size        int64
	MimeType    string
	UploadedBy  uuid.UUID
	WorkspaceID uuid.UUID
}

// FileInfo contains information about a file for responses
type FileInfo struct {
	ID         uuid.UUID `json:"id"`
	Filename   string    `json:"filename"`
	BlobName   string    `json:"-"`
	Size       int64     `json:"size"`
	MimeType   string    `json:"mimeType"`
	URL        string    `json:"url"`
	CreatedAt  time.Time `json:"createdAt"`
	UploadedBy uuid.UUID `json:"uploadedBy"`
}
