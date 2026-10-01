package domain

import (
	"encoding/json"
	"github.com/google/uuid"
	"time"
)

// ImportReceipt contains source identifiers and write results, never an upload
// or provider credential. Canonical source lifecycle metadata is retained here.
type ImportReceipt struct {
	WorkspaceID     uuid.UUID       `json:"-"`
	TeamID          uuid.UUID       `json:"teamId"`
	CreationKey     string          `json:"-"`
	Provider        string          `json:"provider"`
	SourceDigest    string          `json:"sourceDigest"`
	SourceNamespace *string         `json:"sourceNamespace"`
	SourceKey       string          `json:"sourceKey"`
	StoryID         *uuid.UUID      `json:"storyId"`
	Created         bool            `json:"created"`
	ErrorCode       *string         `json:"errorCode"`
	ErrorMessage    *string         `json:"errorMessage"`
	SourceMetadata  json.RawMessage `json:"sourceMetadata,omitempty"`
	UpdatedAt       time.Time       `json:"updatedAt"`
}
