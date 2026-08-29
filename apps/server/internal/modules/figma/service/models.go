package figma

import (
	"context"
	"time"

	figmadomain "github.com/complexus-tech/projects-api/internal/modules/figma/domain"
	"github.com/complexus-tech/projects-api/internal/platform/credentialvault"
	"github.com/complexus-tech/projects-api/internal/platform/integrations"
	"github.com/complexus-tech/projects-api/internal/platform/webhooks"
	"github.com/complexus-tech/projects-api/pkg/tasks"
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
	Credentials  CredentialVault
	// WebhookPayloadSecret protects exact provider bytes in the shared durable
	// inbox. It is separate from retained OAuth credentials and provider
	// passcodes.
	WebhookPayloadSecret string
	WebhookGateway       *webhooks.Gateway
	WebhookInbox         webhooks.Inbox
	WebhookPayloads      WebhookPayloadOpener
}

type Token = figmadomain.Token
type Connection = figmadomain.Connection
type OAuthState = figmadomain.OAuthState
type Artifact = figmadomain.Artifact
type StoryHandoffStatus = figmadomain.StoryHandoffStatus
type StoryLink = figmadomain.StoryLink

type Integration struct {
	Configured bool        `json:"configured"`
	Connection *Connection `json:"connection"`
}

type Webhook = figmadomain.Webhook
type WebhookEvent = figmadomain.WebhookEvent
type FlexibleInt64 = figmadomain.FlexibleInt64

type Repository interface {
	SaveOAuthState(ctx context.Context, state OAuthState) error
	ConsumeOAuthState(ctx context.Context, stateHash string, now time.Time) (OAuthState, error)
	UpsertConnection(ctx context.Context, connection Connection) (Connection, error)
	GetConnection(ctx context.Context, workspaceID uuid.UUID) (Connection, error)
	UpdateConnectionCredential(
		ctx context.Context,
		connectionID, installationGeneration uuid.UUID,
		previousPayload, nextPayload string,
		expiresAt time.Time,
	) (bool, error)
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
	GetCurrentWebhook(
		ctx context.Context,
		connectionID, installationGeneration uuid.UUID,
		figmaWebhookID int64,
	) (Webhook, error)
	FindWebhook(ctx context.Context, connectionID uuid.UUID, fileKey, eventType string) (Webhook, error)
	ListWebhooks(ctx context.Context, connectionID uuid.UUID) ([]Webhook, error)
	DeactivateWebhook(ctx context.Context, figmaWebhookID int64) error
	ListCredentialsForRewrap(
		ctx context.Context,
		after *uuid.UUID,
		limit int32,
	) ([]figmadomain.Credential, error)
	RewrapCredential(
		ctx context.Context,
		record figmadomain.Credential,
		nextPayload string,
	) (bool, error)
}

// CredentialVault is the narrow shared secret capability used by the Figma
// adapter. The provider never receives keyring configuration or key material.
type CredentialVault interface {
	Seal(binding credentialvault.Context, plaintext []byte) (string, error)
	Open(binding credentialvault.Context, envelope string) (credentialvault.Secret, error)
	Rewrap(binding credentialvault.Context, envelope string) (credentialvault.RewrapResult, error)
	ActiveKeyRef() (credentialvault.KeyRef, error)
}

type WebhookQueue interface {
	EnqueueFigmaWebhook(ctx context.Context, payload tasks.FigmaWebhookPayload) error
}

type WebhookPayloadOpener interface {
	Open(record webhooks.Record, value string) ([]byte, error)
}

type WebhookProcessor interface {
	ProcessWebhook(ctx context.Context, provider integrations.ProviderKey, inboxID uuid.UUID) error
}

type StoryService interface {
	Get(ctx context.Context, id, workspaceID uuid.UUID) (Story, error)
	CreateExternal(ctx context.Context, actorID uuid.UUID, story NewStory, workspaceID uuid.UUID) (Story, error)
	RecordActivity(ctx context.Context, activity StoryActivity) error
}

type Story struct {
	ID         uuid.UUID
	SequenceID int
	TeamCode   string
	Title      string
}

type NewStory struct {
	Title           string
	Description     *string
	DescriptionHTML *string
	TeamID          uuid.UUID
	StatusID        *uuid.UUID
	ReporterID      *uuid.UUID
	Priority        string
}

type StoryActivity struct {
	StoryID     uuid.UUID
	ActorID     uuid.UUID
	Type        string
	Field       string
	Previous    string
	Current     string
	WorkspaceID uuid.UUID
}

type CreateStoryInput struct {
	URL           string
	WorkspaceSlug string
	TeamID        uuid.UUID
	StatusID      *uuid.UUID
	Title         *string
	Description   *string
}
