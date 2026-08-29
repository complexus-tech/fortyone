// Package domain defines Figma design-context data without coupling adapters
// to the HTTP or application-service packages.
package domain

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
)

type Token struct {
	AccessToken  string    `json:"accessToken"`
	RefreshToken string    `json:"refreshToken"`
	TokenType    string    `json:"tokenType"`
	ExpiresAt    time.Time `json:"expiresAt"`
}

type Connection struct {
	ID                     uuid.UUID `json:"id"`
	WorkspaceID            uuid.UUID `json:"workspaceId"`
	FigmaUserID            string    `json:"figmaUserId"`
	Email                  *string   `json:"email"`
	Handle                 *string   `json:"handle"`
	CredentialPayload      string    `json:"-"`
	CredentialVersion      int16     `json:"-"`
	InstallationGeneration uuid.UUID `json:"-"`
	Scopes                 []string  `json:"scopes"`
	ExpiresAt              time.Time `json:"expiresAt"`
	ConnectedByUserID      uuid.UUID `json:"connectedByUserId"`
	IsActive               bool      `json:"isActive"`
	CreatedAt              time.Time `json:"createdAt"`
	UpdatedAt              time.Time `json:"updatedAt"`
}

type OAuthState struct {
	StateHash     string
	WorkspaceID   uuid.UUID
	UserID        uuid.UUID
	WorkspaceSlug string
	CodeVerifier  string
	ExpiresAt     time.Time
}

type Artifact struct {
	FileKey      string          `json:"fileKey"`
	NodeID       *string         `json:"nodeId"`
	OriginalURL  string          `json:"originalUrl"`
	CanonicalURL string          `json:"canonicalUrl"`
	FileName     string          `json:"fileName"`
	NodeName     *string         `json:"nodeName"`
	NodeType     *string         `json:"nodeType"`
	ThumbnailURL *string         `json:"thumbnailUrl"`
	Version      *string         `json:"version"`
	LastModified *time.Time      `json:"lastModified"`
	TextContent  []string        `json:"textContent,omitempty"`
	Metadata     json.RawMessage `json:"metadata"`
}

type StoryHandoffStatus struct {
	StoryID uuid.UUID `json:"storyId"`
	Status  string    `json:"status"`
}

type StoryLink struct {
	ID              uuid.UUID  `json:"id"`
	WorkspaceID     uuid.UUID  `json:"workspaceId"`
	StoryID         uuid.UUID  `json:"storyId"`
	CreatedByUserID uuid.UUID  `json:"createdByUserId"`
	StoryLinkID     *uuid.UUID `json:"storyLinkId"`
	Artifact        Artifact   `json:"artifact"`
	DevStatus       *string    `json:"devStatus"`
	DevResourceID   *string    `json:"devResourceId"`
	LastSyncedAt    time.Time  `json:"lastSyncedAt"`
	UnavailableAt   *time.Time `json:"unavailableAt"`
	CreatedAt       time.Time  `json:"createdAt"`
	UpdatedAt       time.Time  `json:"updatedAt"`
}

type Webhook struct {
	ID                     uuid.UUID
	ConnectionID           uuid.UUID
	WorkspaceID            uuid.UUID
	InstallationGeneration uuid.UUID
	FileKey                string
	EventType              string
	FigmaWebhookID         int64
	PasscodeHash           string
	IsActive               bool
}

type WebhookEvent struct {
	EventType     string        `json:"event_type"`
	FileKey       string        `json:"file_key"`
	FileName      string        `json:"file_name"`
	NodeID        string        `json:"node_id"`
	Status        string        `json:"status"`
	ChangeMessage string        `json:"change_message"`
	Passcode      string        `json:"passcode"`
	Timestamp     string        `json:"timestamp"`
	WebhookID     FlexibleInt64 `json:"webhook_id"`
}

// FlexibleInt64 accepts the string and numeric webhook IDs Figma has emitted
// across versions of its webhook payloads.
type FlexibleInt64 int64

func (value *FlexibleInt64) UnmarshalJSON(data []byte) error {
	data = bytes.TrimSpace(data)
	if len(data) == 0 {
		return errors.New("empty integer")
	}
	if data[0] == '"' {
		var text string
		if err := json.Unmarshal(data, &text); err != nil {
			return err
		}
		parsed, err := strconv.ParseInt(text, 10, 64)
		if err != nil {
			return fmt.Errorf("parse quoted integer: %w", err)
		}
		*value = FlexibleInt64(parsed)
		return nil
	}
	parsed, err := strconv.ParseInt(string(data), 10, 64)
	if err != nil {
		return fmt.Errorf("parse integer: %w", err)
	}
	*value = FlexibleInt64(parsed)
	return nil
}
