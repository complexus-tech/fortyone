package attachments

import (
	"time"

	"github.com/google/uuid"
)

// CoreAttachment represents an attachment in the core layer
type CoreAttachment struct {
	ID          uuid.UUID
	Filename    string
	Size        int64
	MimeType    string
	UploadedBy  uuid.UUID
	WorkspaceID uuid.UUID
	CreatedAt   time.Time
}

// CoreNewAttachment represents a new attachment
type CoreNewAttachment struct {
	Filename    string
	Size        int64
	MimeType    string
	UploadedBy  uuid.UUID
	WorkspaceID uuid.UUID
}

// FileInfo contains information about a file for responses
type FileInfo struct {
	ID        uuid.UUID `json:"id"`
	Filename  string    `json:"filename"`
	Size      int64     `json:"size"`
	MimeType  string    `json:"mimeType"`
	URL       string    `json:"url"`
	CreatedAt time.Time `json:"createdAt"`
}

type AttachmentResponse struct {
	ID        uuid.UUID `json:"id"`
	Filename  string    `json:"filename"`
	Size      int64     `json:"size"`
	MimeType  string    `json:"mimeType"`
	URL       string    `json:"url"`
	CreatedAt time.Time `json:"createdAt"`
}

// AttachmentsResponse represents a list of attachments
type AttachmentsResponse struct {
	Attachments []AttachmentResponse `json:"attachments"`
}

// ToAppAttachment converts a core file info to an attachment response
func ToAppAttachment(file FileInfo) AttachmentResponse {
	return AttachmentResponse(file)
}

// ToAttachmentsResponse converts a list of core file infos to an attachments response
func ToAppAttachments(files []FileInfo) AttachmentsResponse {
	resp := AttachmentsResponse{
		Attachments: make([]AttachmentResponse, len(files)),
	}

	for i, file := range files {
		resp.Attachments[i] = ToAppAttachment(file)
	}

	return resp
}
