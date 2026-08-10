package slack

import (
	"context"
	"time"

	integrationrequests "github.com/complexus-tech/projects-api/internal/modules/integrationrequests/service"
	messagingrepository "github.com/complexus-tech/projects-api/internal/modules/messaging/repository"
	slackrepository "github.com/complexus-tech/projects-api/internal/modules/slack/repository"
	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	"github.com/complexus-tech/projects-api/pkg/tasks"
	"github.com/google/uuid"
)

const (
	SourceTypeSlackMessage = "slack_message"
)

type Repository interface {
	FindWorkspaceBySlug(ctx context.Context, slug string) (slackrepository.WorkspaceRecord, error)
	FindWorkspaceByID(ctx context.Context, workspaceID uuid.UUID) (slackrepository.WorkspaceRecord, error)
	FindTeamByCode(ctx context.Context, workspaceID uuid.UUID, code string) (slackrepository.TeamRecord, error)
	FindTeamByID(ctx context.Context, workspaceID, teamID uuid.UUID) (slackrepository.TeamRecord, error)
	ListWorkspaceTeams(ctx context.Context, workspaceID uuid.UUID) ([]slackrepository.TeamRecord, error)
	ListWorkspaceTeamsForUser(ctx context.Context, workspaceID, userID uuid.UUID) ([]slackrepository.TeamRecord, error)
	ListTeamStatuses(ctx context.Context, teamID uuid.UUID) ([]slackrepository.StatusRecord, error)
	ListTeamMembers(ctx context.Context, teamID uuid.UUID) ([]slackrepository.TeamMemberRecord, error)
	ListTeamLabels(ctx context.Context, workspaceID, teamID uuid.UUID) ([]slackrepository.LabelRecord, error)
	FindTeamMemberByID(ctx context.Context, teamID, userID uuid.UUID) (slackrepository.TeamMemberRecord, error)
	FindTeamLabelByID(ctx context.Context, workspaceID, teamID, labelID uuid.UUID) (slackrepository.LabelRecord, error)
	FindTeamObjectiveByID(ctx context.Context, workspaceID, teamID, objectiveID uuid.UUID) (slackrepository.ObjectiveRecord, error)
	SearchTeamMembers(ctx context.Context, teamID uuid.UUID, query string, limit int) ([]slackrepository.TeamMemberRecord, error)
	SearchTeamLabels(ctx context.Context, workspaceID, teamID uuid.UUID, query string, limit int) ([]slackrepository.LabelRecord, error)
	SearchTeamObjectives(ctx context.Context, workspaceID, teamID uuid.UUID, query string, limit int) ([]slackrepository.ObjectiveRecord, error)
	CreateStoryLink(ctx context.Context, storyID uuid.UUID, sourceKey, title, linkURL string) error
	ListWorkspaceMembersForSlackLinking(ctx context.Context, workspaceID uuid.UUID) ([]slackrepository.WorkspaceMemberRecord, error)
	UpsertSlackUserLinks(ctx context.Context, workspaceID, slackWorkspaceID uuid.UUID, slackTeamID string, links []slackrepository.SlackUserLinkUpsert) error
	FindLinkedUserIDBySlackUser(ctx context.Context, workspaceID uuid.UUID, slackTeamID, slackUserID string) (*uuid.UUID, error)
	GetWorkspaceBySlackTeamID(ctx context.Context, slackTeamID string) (slackrepository.WorkspaceRecord, error)
	UpsertSlackWorkspace(ctx context.Context, workspaceID, installedByUserID uuid.UUID, payload slackrepository.OAuthInstallPayload) (slackrepository.SlackWorkspaceRecord, error)
	GetSlackWorkspace(ctx context.Context, workspaceID uuid.UUID) (slackrepository.SlackWorkspaceRecord, error)
	GetSlackWorkspaceByTeamID(ctx context.Context, slackTeamID string) (slackrepository.SlackWorkspaceRecord, error)
	DisconnectSlackWorkspace(ctx context.Context, workspaceID uuid.UUID) (slackrepository.SlackUninstallRecord, error)
	EnqueueSlackUninstall(ctx context.Context, input slackrepository.SlackUninstallInput) (slackrepository.SlackUninstallRecord, error)
	ClaimSlackUninstall(ctx context.Context, id uuid.UUID) (slackrepository.SlackUninstallRecord, bool, error)
	CompleteSlackUninstall(ctx context.Context, id uuid.UUID, message string) error
	FailSlackUninstall(ctx context.Context, id uuid.UUID, message string, nextAttemptAt *time.Time) error
	UpsertChannels(ctx context.Context, workspaceID, slackWorkspaceID uuid.UUID, channels []slackrepository.SlackChannelPayload) error
	ListChannels(ctx context.Context, workspaceID uuid.UUID) ([]slackrepository.SlackChannelRecord, error)
	InsertRequestLog(ctx context.Context, entry slackrepository.SlackRequestLogInsert) error
	ListRequestLogs(ctx context.Context, workspaceID uuid.UUID, limit int) ([]slackrepository.SlackRequestLogRecord, error)
	FindFirstStatusByCategory(ctx context.Context, teamID uuid.UUID, category string) (*uuid.UUID, error)
	HasSlackUserOnboardingReceipt(ctx context.Context, workspaceID uuid.UUID, slackTeamID, slackUserID string) (bool, error)
}

type EventQueue interface {
	EnqueueSlackEvent(ctx context.Context, payload tasks.SlackEventPayload) error
}

type EventInbox interface {
	RegisterInboundEvent(ctx context.Context, input messagingrepository.InboundEventInput) (messagingrepository.InboundEventRecord, bool, error)
	MarkInboundEventQueued(ctx context.Context, id uuid.UUID) error
	FindConversation(ctx context.Context, input messagingrepository.ConversationInput) (messagingrepository.ConversationRecord, error)
}

type OutboundStore interface {
	StartOutboundDelivery(ctx context.Context, input messagingrepository.OutboundDeliveryInput) (messagingrepository.OutboundDeliveryRecord, bool, error)
	SetOutboundDeliveryContent(ctx context.Context, id uuid.UUID, content string) error
	CompleteOutboundDelivery(ctx context.Context, id uuid.UUID, externalMessageID string) error
	FailOutboundDelivery(ctx context.Context, id uuid.UUID, message string) error
	CancelOutboundDelivery(ctx context.Context, id uuid.UUID, message string) error
}

type NonceStore interface {
	CreateNonce(ctx context.Context, input messagingrepository.NonceInput) error
	ConsumeNonce(ctx context.Context, input messagingrepository.NonceConsumeInput) (messagingrepository.NonceRecord, error)
}

type RequestStore interface {
	UpsertPending(ctx context.Context, input integrationrequests.CoreUpsertRequestInput) (integrationrequests.CoreIntegrationRequest, error)
	GetForUser(ctx context.Context, workspaceID, requestID, userID uuid.UUID) (integrationrequests.CoreIntegrationRequest, error)
	BindProviderThread(ctx context.Context, input integrationrequests.CoreBindProviderThreadInput) (integrationrequests.CoreProviderThread, error)
	HasAuthorizedProviderThread(ctx context.Context, input integrationrequests.CoreProviderThreadMatchInput) (bool, error)
	HasCurrentProviderThread(ctx context.Context, input integrationrequests.CoreProviderThreadLookupInput) (bool, error)
	FindProviderThread(ctx context.Context, workspaceID, requestID uuid.UUID, provider string) (integrationrequests.CoreProviderThread, error)
}

type StoryService interface {
	Create(ctx context.Context, ns stories.CoreNewStory, workspaceID uuid.UUID) (stories.CoreSingleStory, error)
}

type Config struct {
	SigningSecret string
	ClientID      string
	ClientSecret  string
	RedirectURL   string
	WebsiteURL    string
	SecretKey     string
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

type CoreIntegration struct {
	SlackWorkspace *CoreSlackWorkspace
	Channels       []CoreSlackChannel
}

type CoreRequestLog struct {
	ID           uuid.UUID
	RequestType  string
	Endpoint     string
	WorkspaceID  *uuid.UUID
	SlackTeamID  *string
	SlackUserID  *string
	SlackChannel *string
	Command      *string
	TriggerID    *string
	RequestBody  *string
	Headers      map[string]string
	ResponseCode int
	Outcome      string
	ErrorMessage *string
	CreatedAt    time.Time
}

type CoreCreateInstallSession struct {
	InstallURL string
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
	StatusKind  string
	StatusID    *uuid.UUID
	Priority    string
	AssigneeID  *uuid.UUID
	LabelIDs    []uuid.UUID
	ObjectiveID *uuid.UUID
	Source      requestSourceContext
	BlockIDs    modalDependentBlockIDs
}

type modalDependentBlockIDs struct {
	Status    string
	Assignee  string
	Labels    string
	Objective string
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
	ResponseURL     string `json:"response_url,omitempty"`
}

type CoreRequestLogInput struct {
	RequestType  string
	Endpoint     string
	RawBody      []byte
	Headers      map[string]string
	ResponseCode int
	Outcome      string
	ErrorMessage string
}

type ProviderAccepter interface {
	AcceptIntegrationRequest(ctx context.Context, request integrationrequests.CoreIntegrationRequest, story stories.CoreSingleStory) error
}

type Clock interface {
	Now() time.Time
}
