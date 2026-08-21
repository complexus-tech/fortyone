package figma

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	"github.com/google/uuid"
)

const (
	EventFileUpdate          = "FILE_UPDATE"
	EventDevModeStatusUpdate = "DEV_MODE_STATUS_UPDATE"
)

type Config struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	WebhookURL   string
	WebsiteURL   string
	SecretKey    string
}

type Token struct {
	AccessToken  string    `json:"accessToken"`
	RefreshToken string    `json:"refreshToken"`
	TokenType    string    `json:"tokenType"`
	ExpiresAt    time.Time `json:"expiresAt"`
}

type Connection struct {
	ID                uuid.UUID `json:"id"`
	WorkspaceID       uuid.UUID `json:"workspaceId"`
	FigmaUserID       string    `json:"figmaUserId"`
	Email             *string   `json:"email"`
	Handle            *string   `json:"handle"`
	TokenPayload      string    `json:"-"`
	Scopes            []string  `json:"scopes"`
	ExpiresAt         time.Time `json:"expiresAt"`
	ConnectedByUserID uuid.UUID `json:"connectedByUserId"`
	IsActive          bool      `json:"isActive"`
	CreatedAt         time.Time `json:"createdAt"`
	UpdatedAt         time.Time `json:"updatedAt"`
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
	StoryID uuid.UUID `json:"storyId" db:"story_id"`
	Status  string    `json:"status" db:"status"`
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

type Integration struct {
	Configured bool        `json:"configured"`
	Connection *Connection `json:"connection"`
}

type Webhook struct {
	ID             uuid.UUID
	ConnectionID   uuid.UUID
	WorkspaceID    uuid.UUID
	FileKey        string
	EventType      string
	FigmaWebhookID int64
	PasscodeHash   string
	IsActive       bool
}

type WebhookEvent struct {
	EventType     string          `json:"event_type"`
	FileKey       string          `json:"file_key"`
	FileName      string          `json:"file_name"`
	NodeID        string          `json:"node_id"`
	Status        string          `json:"status"`
	ChangeMessage string          `json:"change_message"`
	Passcode      string          `json:"passcode"`
	Timestamp     string          `json:"timestamp"`
	WebhookID     FlexibleInt64   `json:"webhook_id"`
	Raw           json.RawMessage `json:"-"`
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

type Repository interface {
	SaveOAuthState(ctx context.Context, state OAuthState) error
	ConsumeOAuthState(ctx context.Context, stateHash string, now time.Time) (OAuthState, error)
	UpsertConnection(ctx context.Context, connection Connection) (Connection, error)
	GetConnection(ctx context.Context, workspaceID uuid.UUID) (Connection, error)
	UpdateConnectionToken(ctx context.Context, connectionID uuid.UUID, payload string, expiresAt time.Time) error
	Disconnect(ctx context.Context, workspaceID uuid.UUID) error
	ListStoryLinks(ctx context.Context, workspaceID, storyID uuid.UUID) ([]StoryLink, error)
	ListStoryHandoffStatuses(ctx context.Context, workspaceID uuid.UUID) ([]StoryHandoffStatus, error)
	ListLinksByFile(ctx context.Context, workspaceID uuid.UUID, fileKey string) ([]StoryLink, error)
	UpsertStoryLink(ctx context.Context, link StoryLink) (StoryLink, error)
	UpdateStoryLink(ctx context.Context, link StoryLink) error
	GetStoryLink(ctx context.Context, workspaceID, linkID uuid.UUID) (StoryLink, error)
	DeleteStoryLink(ctx context.Context, workspaceID, linkID uuid.UUID) (StoryLink, error)
	SaveWebhook(ctx context.Context, webhook Webhook) error
	GetWebhook(ctx context.Context, figmaWebhookID int64) (Webhook, error)
	RecordWebhookEvent(ctx context.Context, eventKey string, event WebhookEvent) (bool, error)
	FindWebhook(ctx context.Context, connectionID uuid.UUID, fileKey, eventType string) (Webhook, error)
	ListWebhooks(ctx context.Context, connectionID uuid.UUID) ([]Webhook, error)
	DeactivateWebhook(ctx context.Context, figmaWebhookID int64) error
}

type StoryService interface {
	Get(ctx context.Context, id uuid.UUID, workspaceID uuid.UUID) (stories.CoreSingleStory, error)
	CreateExternal(ctx context.Context, actorID uuid.UUID, story stories.CoreNewStory, workspaceID uuid.UUID) (stories.CoreSingleStory, error)
	RecordActivity(ctx context.Context, activity stories.CoreActivity) error
}

type CreateStoryInput struct {
	URL           string
	WorkspaceSlug string
	TeamID        uuid.UUID
	StatusID      *uuid.UUID
	Title         *string
	Description   *string
}
