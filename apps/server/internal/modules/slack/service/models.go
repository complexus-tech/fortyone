package slack

import (
	"context"
	"time"

	slackdomain "github.com/complexus-tech/projects-api/internal/modules/slack/domain"
	platformclock "github.com/complexus-tech/projects-api/internal/platform/clock"
	"github.com/complexus-tech/projects-api/internal/platform/credentialvault"
	"github.com/complexus-tech/projects-api/internal/platform/integrations"
	"github.com/complexus-tech/projects-api/internal/platform/webhooks"
	"github.com/complexus-tech/projects-api/pkg/tasks"
	"github.com/google/uuid"
)

const (
	SourceTypeSlackMessage = "slack_message"
)

type WorkspaceDirectory interface {
	FindWorkspaceBySlug(ctx context.Context, slug string) (slackdomain.Workspace, error)
	FindWorkspaceByID(ctx context.Context, workspaceID uuid.UUID) (slackdomain.Workspace, error)
	FindTeamByCode(ctx context.Context, workspaceID uuid.UUID, code string) (slackdomain.Team, error)
	FindTeamByID(ctx context.Context, workspaceID, teamID uuid.UUID) (slackdomain.Team, error)
	ListWorkspaceTeams(ctx context.Context, workspaceID uuid.UUID) ([]slackdomain.Team, error)
	ListWorkspaceTeamsForUser(ctx context.Context, workspaceID, userID uuid.UUID) ([]slackdomain.Team, error)
	ListTeamStatuses(ctx context.Context, teamID uuid.UUID) ([]slackdomain.Status, error)
	ListTeamMembers(ctx context.Context, teamID uuid.UUID) ([]slackdomain.TeamMember, error)
	ListTeamLabels(ctx context.Context, workspaceID, teamID uuid.UUID) ([]slackdomain.Label, error)
	FindTeamMemberByID(ctx context.Context, teamID, userID uuid.UUID) (slackdomain.TeamMember, error)
	FindTeamLabelByID(ctx context.Context, workspaceID, teamID, labelID uuid.UUID) (slackdomain.Label, error)
	FindTeamObjectiveByID(ctx context.Context, workspaceID, teamID, objectiveID uuid.UUID) (slackdomain.Objective, error)
	SearchTeamMembers(ctx context.Context, teamID uuid.UUID, query string, limit int) ([]slackdomain.TeamMember, error)
	SearchTeamLabels(ctx context.Context, workspaceID, teamID uuid.UUID, query string, limit int) ([]slackdomain.Label, error)
	SearchTeamObjectives(ctx context.Context, workspaceID, teamID uuid.UUID, query string, limit int) ([]slackdomain.Objective, error)
	CreateStoryLink(ctx context.Context, storyID uuid.UUID, sourceKey, title, linkURL string) error
}

type UserLinkStore interface {
	ListWorkspaceMembersForSlackLinking(ctx context.Context, workspaceID uuid.UUID) ([]slackdomain.WorkspaceMember, error)
	UpsertSlackUserLinks(ctx context.Context, workspaceID, slackWorkspaceID uuid.UUID, slackTeamID string, links []slackdomain.UserLinkUpsert) error
	FindLinkedUserIDBySlackUser(ctx context.Context, workspaceID uuid.UUID, slackTeamID, slackUserID string) (*uuid.UUID, error)
	FindSlackUserLinkByUser(ctx context.Context, workspaceID uuid.UUID, slackTeamID string, userID uuid.UUID) (*slackdomain.UserLink, error)
	DeleteSlackUserLink(ctx context.Context, workspaceID uuid.UUID, slackTeamID, slackUserID string, userID uuid.UUID) (bool, error)
}

type InstallationStore interface {
	GetWorkspaceBySlackTeamID(ctx context.Context, slackTeamID string) (slackdomain.Workspace, error)
	UpsertSlackWorkspace(ctx context.Context, workspaceID, installedByUserID uuid.UUID, payload slackdomain.OAuthInstallation) (slackdomain.Installation, error)
	GetSlackWorkspace(ctx context.Context, workspaceID uuid.UUID) (slackdomain.Installation, error)
	GetSlackWorkspaceByTeamID(ctx context.Context, slackTeamID string) (slackdomain.Installation, error)
	DisconnectSlackWorkspace(ctx context.Context, command slackdomain.DisconnectInstallationCommand) (slackdomain.Uninstall, error)
	EnqueueSlackUninstall(ctx context.Context, input slackdomain.EnqueueUninstall) (slackdomain.Uninstall, error)
	ClaimSlackUninstall(ctx context.Context, id uuid.UUID) (slackdomain.Uninstall, bool, error)
	CompleteSlackUninstall(ctx context.Context, id uuid.UUID, message string) error
	FailSlackUninstall(ctx context.Context, id uuid.UUID, message string, nextAttemptAt *time.Time) error
}

type ChannelStore interface {
	UpsertChannels(ctx context.Context, command slackdomain.SyncChannelsCommand) error
	ListChannels(ctx context.Context, workspaceID uuid.UUID) ([]slackdomain.Channel, error)
}

type RequestLogStore interface {
	InsertRequestLog(ctx context.Context, entry slackdomain.RequestLogInsert) error
	ListRequestLogsForAdmin(ctx context.Context, query slackdomain.ListRequestLogsQuery) ([]slackdomain.RequestLog, error)
}

// HumanIntegrationStore keeps the live actor check in every dashboard read.
// Provider workers use the separately named installation/directory methods and
// must prove installation generation instead of fabricating a human actor.
type HumanIntegrationStore interface {
	GetSlackWorkspaceForMember(ctx context.Context, query slackdomain.WorkspaceActorQuery) (slackdomain.Installation, error)
	FindSlackUserLinkForMember(ctx context.Context, query slackdomain.WorkspaceActorQuery, slackTeamID string) (*slackdomain.UserLink, error)
	ListChannelsForMember(ctx context.Context, query slackdomain.WorkspaceActorQuery) ([]slackdomain.Channel, error)
}

type OnboardingStore interface {
	FindFirstStatusByCategory(ctx context.Context, teamID uuid.UUID, category string) (*uuid.UUID, error)
	HasSlackUserOnboardingReceipt(ctx context.Context, workspaceID uuid.UUID, slackTeamID, slackUserID string) (bool, error)
}

// Repository is the composition contract. Individual use cases consume the
// capability interfaces above; the constructor accepts the concrete aggregate
// only so bootstrap cannot accidentally provide a partially configured Slack
// persistence adapter.
type Repository interface {
	WorkspaceDirectory
	UserLinkStore
	InstallationStore
	ChannelStore
	HumanIntegrationStore
	RequestLogStore
	OnboardingStore
}

type EventQueue interface {
	EnqueueSlackEvent(ctx context.Context, payload tasks.SlackEventPayload) error
}

type EventGateway interface {
	Receive(ctx context.Context, provider integrations.ProviderKey, request webhooks.SignedRequest) (webhooks.Receipt, error)
}

type EventInbox interface {
	FindConversation(ctx context.Context, input ConversationInput) (ConversationRecord, error)
}

type OutboundStore interface {
	StartOutboundDelivery(ctx context.Context, input OutboundDeliveryInput) (OutboundDeliveryRecord, bool, error)
	SetOutboundDeliveryContent(ctx context.Context, id uuid.UUID, content string) error
	CompleteOutboundDelivery(ctx context.Context, id uuid.UUID, externalMessageID string) error
	FailOutboundDelivery(ctx context.Context, id uuid.UUID, message string) error
	CancelOutboundDelivery(ctx context.Context, id uuid.UUID, message string) error
}

type NonceStore interface {
	CreateNonce(ctx context.Context, input NonceInput) error
	ConsumeNonce(ctx context.Context, input NonceConsumeInput) (NonceRecord, error)
}

type RequestStore interface {
	UpsertPending(ctx context.Context, input UpsertIntegrationRequestInput) (IntegrationRequest, error)
	GetForUser(ctx context.Context, workspaceID, requestID, userID uuid.UUID) (IntegrationRequest, error)
	BindProviderThread(ctx context.Context, input BindProviderThreadInput) (ProviderThread, error)
	HasAuthorizedProviderThread(ctx context.Context, input ProviderThreadMatchInput) (bool, error)
	HasCurrentProviderThread(ctx context.Context, input ProviderThreadLookupInput) (bool, error)
	FindProviderThread(ctx context.Context, workspaceID, requestID uuid.UUID, provider string) (ProviderThread, error)
}

type StoryService interface {
	Create(ctx context.Context, story NewStory, workspaceID uuid.UUID) (Story, error)
}

type Config struct {
	SigningSecret        string
	WebhookPayloadSecret string
	ClientID             string
	ClientSecret         string
	RedirectURL          string
	WebsiteURL           string
	CredentialVault      CredentialVault
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
	AccountLink    *CoreSlackAccountLink
	Channels       []CoreSlackChannel
}

type CoreSlackAccountLink struct {
	SlackUserID string
	LinkedVia   string
	LinkedAt    time.Time
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

type CoreCreateAccountLinkSession struct {
	Linked     bool
	CanLink    bool
	InstallURL string
}

type CoreLinkSlackAccountResult struct {
	AlreadyLinked bool
	SlackUserID   string
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
	AcceptIntegrationRequest(ctx context.Context, request IntegrationRequest, story Story) error
}

type Clock = platformclock.Clock

// CredentialVault is the narrow recoverable-secret capability consumed by the
// Slack adapter. Provider clients never receive vault key material or an
// implementation-specific keyring.
type CredentialVault interface {
	Seal(binding credentialvault.Context, plaintext []byte) (string, error)
	Open(binding credentialvault.Context, envelope string) (credentialvault.Secret, error)
	Rewrap(binding credentialvault.Context, envelope string) (credentialvault.RewrapResult, error)
	ActiveKeyRef() (credentialvault.KeyRef, error)
}
