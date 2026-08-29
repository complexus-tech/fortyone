package attachments

import (
	"time"

	attachmentdomain "github.com/complexus-tech/projects-api/internal/modules/attachments/domain"
	"github.com/google/uuid"
)

// CoreAttachment represents an attachment in the core layer
type CoreAttachment = attachmentdomain.Attachment

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
