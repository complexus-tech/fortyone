package attachmentsrepo

import (
	"time"

	"github.com/complexus-tech/projects-api/internal/core/attachments"
	"github.com/google/uuid"
)

// dbAttachment represents an attachment in the database
type dbAttachment struct {
	ID          uuid.UUID `db:"attachment_id"`
	Filename    string    `db:"filename"`
	Size        int64     `db:"size"`
	MimeType    string    `db:"mime_type"`
	UploadedBy  uuid.UUID `db:"uploaded_by"`
	WorkspaceID uuid.UUID `db:"workspace_id"`
	CreatedAt   time.Time `db:"created_at"`
}

// toCoreAttachment converts a database attachment to a core attachment
func toCoreAttachment(a dbAttachment) attachments.CoreAttachment {
	return attachments.CoreAttachment{
		ID:          a.ID,
		Filename:    a.Filename,
		Size:        a.Size,
		MimeType:    a.MimeType,
		UploadedBy:  a.UploadedBy,
		WorkspaceID: a.WorkspaceID,
		CreatedAt:   a.CreatedAt,
	}
}
