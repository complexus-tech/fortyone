package attachmentdomain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrNotFound      = errors.New("attachment not found")
	ErrNotConfigured = errors.New("attachment repository is not configured")
	ErrStateConflict = errors.New("attachment processing state changed")
)

type ScanStatus string

const (
	ScanStatusUnscanned ScanStatus = "unscanned"
	ScanStatusPending   ScanStatus = "pending"
	ScanStatusClean     ScanStatus = "clean"
	ScanStatusInfected  ScanStatus = "infected"
	ScanStatusFailed    ScanStatus = "failed"
)

type OptimizationStatus string

const (
	OptimizationNotRequested OptimizationStatus = "not_requested"
	OptimizationQueued       OptimizationStatus = "queued"
	OptimizationProcessing   OptimizationStatus = "processing"
	OptimizationSucceeded    OptimizationStatus = "succeeded"
	OptimizationSkipped      OptimizationStatus = "skipped"
	OptimizationFailed       OptimizationStatus = "failed"
)

type Attachment struct {
	ID                       uuid.UUID
	Filename                 string
	BlobName                 string
	Size                     int64
	MimeType                 string
	UploadedBy               uuid.UUID
	WorkspaceID              uuid.UUID
	CreatedAt                time.Time
	UpdatedAt                time.Time
	ScanStatus               ScanStatus
	ScanCompletedAt          *time.Time
	ScanFailureReason        string
	OptimizationStatus       OptimizationStatus
	OptimizationAttempts     int32
	OptimizationStartedAt    *time.Time
	OptimizationCompletedAt  *time.Time
	OptimizationLeaseExpires *time.Time
	OptimizationLastError    string
}

func (attachment Attachment) AvailableForDownload() bool {
	switch attachment.ScanStatus {
	case ScanStatusPending, ScanStatusInfected, ScanStatusFailed:
		return false
	default:
		return true
	}
}
