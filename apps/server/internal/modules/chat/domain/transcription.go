package chatdomain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrTranscriptionUnavailable = errors.New("voice note transcription is unavailable")
	ErrTranscriptionAccess      = errors.New("voice note transcription requires Maya access for this workspace")
)

type Transcription struct {
	Status    string     `json:"status"`
	Text      string     `json:"text,omitempty"`
	Error     string     `json:"error,omitempty"`
	UpdatedAt *time.Time `json:"updatedAt,omitempty"`
}

type TranscriptionJob struct {
	AttachmentID, WorkspaceID, ConversationID, RequesterID, LeaseToken uuid.UUID
}

type TranscriptionAudio struct {
	BlobName, FileName, MediaType string
	SizeBytes                     int64
}
