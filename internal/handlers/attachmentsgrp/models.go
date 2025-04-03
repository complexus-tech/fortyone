package attachmentsgrp

import (
	"time"

	"github.com/complexus-tech/projects-api/internal/core/attachments"
	"github.com/google/uuid"
)

// AttachmentResponse represents an attachment response
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

// LinkAttachmentRequest represents a request to link an attachment to a story
type LinkAttachmentRequest struct {
	AttachmentID uuid.UUID `json:"attachmentId" validate:"required"`
}

// NewAttachmentResponse converts a core file info to an attachment response
func NewAttachmentResponse(file attachments.FileInfo) AttachmentResponse {
	return AttachmentResponse{
		ID:        file.ID,
		Filename:  file.Filename,
		Size:      file.Size,
		MimeType:  file.MimeType,
		URL:       file.URL,
		CreatedAt: file.CreatedAt,
	}
}

// NewAttachmentsResponse converts a list of core file infos to an attachments response
func NewAttachmentsResponse(files []attachments.FileInfo) AttachmentsResponse {
	resp := AttachmentsResponse{
		Attachments: make([]AttachmentResponse, len(files)),
	}

	for i, file := range files {
		resp.Attachments[i] = NewAttachmentResponse(file)
	}

	return resp
}
