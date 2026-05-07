package slack

import (
	"context"
	"time"

	integrationrequests "github.com/complexus-tech/projects-api/internal/modules/integrationrequests/service"
	slackrepository "github.com/complexus-tech/projects-api/internal/modules/slack/repository"
	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	"github.com/google/uuid"
)

const (
	SourceTypeSlackMessage = "slack_message"
)

type Repository interface {
	FindWorkspaceBySlug(ctx context.Context, slug string) (slackrepository.WorkspaceRecord, error)
	FindTeamByCode(ctx context.Context, workspaceID uuid.UUID, code string) (slackrepository.TeamRecord, error)
}

type RequestStore interface {
	UpsertPending(ctx context.Context, input integrationrequests.CoreUpsertRequestInput) (integrationrequests.CoreIntegrationRequest, error)
}

type Config struct {
	SigningSecret string
	BotToken      string
	WebsiteURL    string
}

type CommandResponse struct {
	ResponseType string `json:"response_type,omitempty"`
	Text         string `json:"text,omitempty"`
}

type InteractionResponse struct {
	StatusCode  int
	ContentType string
	Body        []byte
}

type EventResponse struct {
	Challenge string `json:"challenge,omitempty"`
}

type viewSubmissionData struct {
	WorkspaceSlug string
	TeamCode      string
	Title         string
	Description   string
	Source        requestSourceContext
}

type requestSourceContext struct {
	SlackTeamID     string `json:"slack_team_id,omitempty"`
	SlackTeamDomain string `json:"slack_team_domain,omitempty"`
	SlackChannelID  string `json:"slack_channel_id,omitempty"`
	SlackChannel    string `json:"slack_channel,omitempty"`
	SlackMessageTS  string `json:"slack_message_ts,omitempty"`
	SlackThreadTS   string `json:"slack_thread_ts,omitempty"`
	SlackUserID     string `json:"slack_user_id,omitempty"`
	SlackUsername   string `json:"slack_username,omitempty"`
	SlackText       string `json:"slack_text,omitempty"`
}

type ProviderAccepter interface {
	AcceptIntegrationRequest(ctx context.Context, request integrationrequests.CoreIntegrationRequest, story stories.CoreSingleStory) error
}

type Clock interface {
	Now() time.Time
}
