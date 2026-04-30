package integrationrequests

import (
	"context"
	"time"

	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	"github.com/google/uuid"
)

const (
	ProviderGitHub   = "github"
	ProviderSlack    = "slack"
	ProviderIntercom = "intercom"

	SourceTypeIssue = "issue"

	StatusPending  = "pending"
	StatusAccepted = "accepted"
	StatusDeclined = "declined"
)

type CoreIntegrationRequest struct {
	ID               uuid.UUID
	WorkspaceID      uuid.UUID
	TeamID           uuid.UUID
	Provider         string
	SourceType       string
	SourceExternalID string
	SourceNumber     *int
	SourceURL        *string
	Title            string
	Description      *string
	Status           string
	Metadata         map[string]any
	AcceptedStoryID  *uuid.UUID
	AcceptedByUserID *uuid.UUID
	AcceptedAt       *time.Time
	DeclinedByUserID *uuid.UUID
	DeclinedAt       *time.Time
	CreatedByUserID  *uuid.UUID
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type CoreUpsertRequestInput struct {
	WorkspaceID      uuid.UUID
	TeamID           uuid.UUID
	Provider         string
	SourceType       string
	SourceExternalID string
	SourceNumber     *int
	SourceURL        *string
	Title            string
	Description      *string
	Metadata         map[string]any
	CreatedByUserID  *uuid.UUID
}

type CoreListRequestsFilter struct {
	Status string
}

type StoryService interface {
	CreateExternal(ctx context.Context, actorID uuid.UUID, ns stories.CoreNewStory, workspaceID uuid.UUID) (stories.CoreSingleStory, error)
}

type ProviderAccepter interface {
	AcceptIntegrationRequest(ctx context.Context, request CoreIntegrationRequest, story stories.CoreSingleStory) error
}
