package domain

import (
	"time"

	"github.com/google/uuid"
)

// StoryRetentionCursor is the stable keyset position for permanently deleting
// soft-deleted stories in deleted-at and story-ID order.
type StoryRetentionCursor struct {
	DeletedAt time.Time
	StoryID   uuid.UUID
	Valid     bool
}

// StoryRetentionBatch describes one atomic, bounded story and attachment
// retirement transaction.
type StoryRetentionBatch struct {
	DeletedBefore          time.Time
	EnqueuedAt             time.Time
	Cursor                 StoryRetentionCursor
	BatchSize              int
	MaximumAttachmentCount int
	StorageProvider        string
	ContainerName          string
}

// StoryRetentionResult reports only committed database work. CandidateCount
// determines whether the caller should advance the returned cursor.
type StoryRetentionResult struct {
	CandidateCount      int
	DeletedStoryCount   int64
	EnqueuedObjectCount int64
	NextCursor          StoryRetentionCursor
}

// AttachmentObjectDeletionClaimBatch describes one fenced claim of due or
// expired-lease object deletion work.
type AttachmentObjectDeletionClaimBatch struct {
	AsOf               time.Time
	LeaseExpiredBefore time.Time
	ClaimToken         uuid.UUID
	BatchSize          int
}

// AttachmentObjectDeletion is credential-free routing metadata for one
// claimed object. BlobName must never enter logs, traces, metrics, or errors.
type AttachmentObjectDeletion struct {
	OutboxID        uuid.UUID
	AttachmentID    uuid.UUID
	WorkspaceID     uuid.UUID
	StorageProvider string
	ContainerName   string
	BlobName        string
	ClaimToken      uuid.UUID
	AttemptCount    int
}

// AttachmentObjectDeletionCompletion fences a successful provider deletion
// to the worker that currently owns the claim.
type AttachmentObjectDeletionCompletion struct {
	OutboxID    uuid.UUID
	ClaimToken  uuid.UUID
	CompletedAt time.Time
}

// AttachmentObjectDeletionFailure releases a claimed row for a later retry.
// LastError is a bounded safe classification, never a provider response body.
type AttachmentObjectDeletionFailure struct {
	OutboxID      uuid.UUID
	ClaimToken    uuid.UUID
	FailedAt      time.Time
	NextAttemptAt time.Time
	LastError     string
}
