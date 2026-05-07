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

	CreateModeCreateTaskNow  = "create_task_now"
	CreateModeSendToRequests = "send_to_requests"
)

type Repository interface {
	FindWorkspaceBySlug(ctx context.Context, slug string) (slackrepository.WorkspaceRecord, error)
	FindWorkspaceByID(ctx context.Context, workspaceID uuid.UUID) (slackrepository.WorkspaceRecord, error)
	FindTeamByCode(ctx context.Context, workspaceID uuid.UUID, code string) (slackrepository.TeamRecord, error)
	FindTeamByID(ctx context.Context, workspaceID, teamID uuid.UUID) (slackrepository.TeamRecord, error)
	ListWorkspaceTeams(ctx context.Context, workspaceID uuid.UUID) ([]slackrepository.TeamRecord, error)
	ListTeamStatuses(ctx context.Context, teamID uuid.UUID) ([]slackrepository.StatusRecord, error)
	ListTeamMembers(ctx context.Context, teamID uuid.UUID) ([]slackrepository.TeamMemberRecord, error)
	ListTeamLabels(ctx context.Context, workspaceID, teamID uuid.UUID) ([]slackrepository.LabelRecord, error)
	GetWorkspaceBySlackTeamID(ctx context.Context, slackTeamID string) (slackrepository.WorkspaceRecord, error)
	GetWorkspaceSettings(ctx context.Context, workspaceID uuid.UUID) (slackrepository.SlackWorkspaceSettingsRecord, error)
	UpdateWorkspaceSettings(ctx context.Context, workspaceID uuid.UUID, defaultCreateMode string) (slackrepository.SlackWorkspaceSettingsRecord, error)
	UpsertSlackWorkspace(ctx context.Context, workspaceID, installedByUserID uuid.UUID, payload slackrepository.OAuthInstallPayload) (slackrepository.SlackWorkspaceRecord, error)
	GetSlackWorkspace(ctx context.Context, workspaceID uuid.UUID) (slackrepository.SlackWorkspaceRecord, error)
	GetSlackWorkspaceByTeamID(ctx context.Context, slackTeamID string) (slackrepository.SlackWorkspaceRecord, error)
	UpsertChannels(ctx context.Context, workspaceID, slackWorkspaceID uuid.UUID, channels []slackrepository.SlackChannelPayload) error
	ListChannels(ctx context.Context, workspaceID uuid.UUID) ([]slackrepository.SlackChannelRecord, error)
	UpsertChannelLink(ctx context.Context, workspaceID uuid.UUID, slackChannelID string, teamID, createdByUserID uuid.UUID) (slackrepository.SlackChannelLinkRecord, error)
	GetChannelLinkByID(ctx context.Context, workspaceID, linkID uuid.UUID) (slackrepository.SlackChannelLinkRecord, error)
	ListChannelLinks(ctx context.Context, workspaceID uuid.UUID) ([]slackrepository.SlackChannelLinkRecord, error)
	DeleteChannelLink(ctx context.Context, workspaceID, linkID uuid.UUID) error
	FindTeamByChannel(ctx context.Context, workspaceID uuid.UUID, slackChannelID string) (slackrepository.TeamRecord, error)
	FindFirstStatusByCategory(ctx context.Context, teamID uuid.UUID, category string) (*uuid.UUID, error)
}

type RequestStore interface {
	UpsertPending(ctx context.Context, input integrationrequests.CoreUpsertRequestInput) (integrationrequests.CoreIntegrationRequest, error)
}

type StoryService interface {
	CreateExternal(ctx context.Context, actorID uuid.UUID, ns stories.CoreNewStory, workspaceID uuid.UUID) (stories.CoreSingleStory, error)
	UpdateLabels(ctx context.Context, id uuid.UUID, workspaceId uuid.UUID, labels []uuid.UUID) error
}

type Config struct {
	SigningSecret string
	ClientID      string
	ClientSecret  string
	RedirectURL   string
	WebsiteURL    string
	SecretKey     string
}

type CoreWorkspaceSettings struct {
	WorkspaceID       uuid.UUID
	DefaultCreateMode string
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type CoreSlackWorkspace struct {
	ID                uuid.UUID
	SlackTeamID       string
	SlackTeamName     string
	SlackTeamDomain   string
	BotUserID         *string
	Scope             *string
	IsActive          bool
	InstalledByUserID *uuid.UUID
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type CoreSlackChannel struct {
	ID             uuid.UUID
	SlackChannelID string
	Name           string
	IsPrivate      bool
	IsArchived     bool
	IsMember       bool
	IsActive       bool
	LastSyncedAt   *time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type CoreSlackChannelLink struct {
	ID             uuid.UUID
	SlackChannelID string
	TeamID         uuid.UUID
	TeamCode       string
	TeamName       string
	TeamColor      string
	IsActive       bool
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type CoreIntegration struct {
	Settings       CoreWorkspaceSettings
	SlackWorkspace *CoreSlackWorkspace
	Channels       []CoreSlackChannel
	ChannelLinks   []CoreSlackChannelLink
}

type CoreCreateInstallSession struct {
	InstallURL string
}

type CoreUpdateWorkspaceSettingsInput struct {
	DefaultCreateMode *string
}

type CoreCreateChannelLinkInput struct {
	SlackChannelID string
	TeamID         uuid.UUID
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
	Title       string
	Description string
	TeamID      uuid.UUID
	StatusID    *uuid.UUID
	Priority    string
	AssigneeID  *uuid.UUID
	LabelIDs    []uuid.UUID
	Source      requestSourceContext
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
