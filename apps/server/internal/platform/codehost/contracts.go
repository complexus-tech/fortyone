// Package codehost defines the provider-neutral capability contracts shared by
// first-party source-control adapters. Provider SDK and webhook payload types
// must stay in their adapter packages.
package codehost

import (
	"context"
	"errors"
	"time"

	"github.com/complexus-tech/projects-api/internal/platform/integrations"
	"github.com/google/uuid"
)

var (
	ErrInvalidInput          = errors.New("code-host input is invalid")
	ErrCapabilityUnsupported = errors.New("code-host capability is unsupported")
	ErrAuthentication        = errors.New("code-host authentication failed")
	ErrRateLimited           = errors.New("code-host provider rate limited the request")
	ErrGrantRevoked          = errors.New("code-host installation grant is revoked")
	ErrNotFound              = errors.New("code-host resource was not found")
)

type Capability string

const (
	CapabilityInstallationAuth     Capability = "installation_auth"
	CapabilityRepositoryCatalog    Capability = "repository_catalog"
	CapabilityWorkItemWriter       Capability = "work_item_writer"
	CapabilityCommentWriter        Capability = "comment_writer"
	CapabilityWebhookNormalization Capability = "webhook_normalization"
)

type Capabilities map[Capability]bool

func (capabilities Capabilities) Require(capability Capability) error {
	if !capabilities[capability] {
		return ErrCapabilityUnsupported
	}
	return nil
}

type InstallationRef struct {
	Provider               integrations.ProviderKey
	WorkspaceID            uuid.UUID
	InstallationID         uuid.UUID
	ExternalInstallationID string
	Generation             uuid.UUID
}

type RepositoryRef struct {
	ExternalID    string
	Owner         string
	Name          string
	FullName      string
	WebURL        string
	DefaultBranch string
	Private       bool
	Archived      bool
}

type Cursor struct {
	Value string
	Limit int
}

type RepositoryPage struct {
	Repositories []RepositoryRef
	NextCursor   string
}

type WorkItemState string

const (
	WorkItemStateOpen   WorkItemState = "open"
	WorkItemStateClosed WorkItemState = "closed"
)

type WorkItem struct {
	ExternalID string
	Number     int64
	Repository RepositoryRef
	Title      string
	Body       string
	State      WorkItemState
	WebURL     string
}

type CreateWorkItem struct {
	Repository RepositoryRef
	Title      string
	Body       string
}

type Comment struct {
	ExternalID  string
	WorkItem    WorkItem
	AuthorID    string
	AuthorLogin string
	Body        string
	WebURL      string
	CreatedAt   time.Time
}

type AddComment struct {
	WorkItem WorkItem
	Body     string
}

type EventKind string

const (
	EventWorkItemChanged EventKind = "work_item.changed"
	EventCommentCreated  EventKind = "comment.created"
	EventPush            EventKind = "repository.push"
	EventMergeRequest    EventKind = "merge_request.changed"
)

type NormalizedEvent struct {
	Provider             integrations.ProviderKey
	DeliveryID           string
	Kind                 EventKind
	Action               string
	ExternalRepositoryID string
	ExternalActorID      string
	WorkItem             *WorkItem
	Comment              *Comment
}

type InstallationAuthenticator interface {
	Authorize(ctx context.Context, installation InstallationRef) error
}

type RepositoryCatalog interface {
	ListRepositories(ctx context.Context, installation InstallationRef, cursor Cursor) (RepositoryPage, error)
}

type WorkItemWriter interface {
	CreateWorkItem(ctx context.Context, installation InstallationRef, command CreateWorkItem) (WorkItem, error)
}

type CommentWriter interface {
	AddComment(ctx context.Context, installation InstallationRef, command AddComment) (Comment, error)
}

type WebhookNormalizer interface {
	NormalizeWebhook(ctx context.Context, deliveryID, eventType string, body []byte) (NormalizedEvent, error)
}

type Adapter interface {
	Provider() integrations.ProviderKey
	Capabilities() Capabilities
}
