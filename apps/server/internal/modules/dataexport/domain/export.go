package dataexportdomain

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

const (
	MaximumTasks = 10000
	MaximumBytes = 20 * 1024 * 1024
)

var (
	ErrForbidden = errors.New("workspace administrator access is required")
	ErrTooLarge  = errors.New("export exceeds the file limit; export one team at a time")
	ErrInvalid   = errors.New("invalid export scope")
)

type Scope struct {
	ActorID     uuid.UUID
	WorkspaceID uuid.UUID
	TeamID      *uuid.UUID
}

func (scope Scope) Validate() error {
	if scope.ActorID == uuid.Nil || scope.WorkspaceID == uuid.Nil || (scope.TeamID != nil && *scope.TeamID == uuid.Nil) {
		return ErrInvalid
	}
	return nil
}

// Envelope is a versioned work graph. Canonical task content and lifecycle
// provenance live in TaskData; the analysis graph supports import previews.
type Envelope struct {
	Format       string          `json:"format"`
	Version      int             `json:"version"`
	GeneratedAt  time.Time       `json:"generatedAt"`
	Analysis     json.RawMessage `json:"analysis"`
	CustomFields json.RawMessage `json:"customFields"`
	TaskData     json.RawMessage `json:"taskData"`
}

type Snapshot struct {
	Envelope  Envelope
	TaskCount int
}
