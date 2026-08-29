package slack

import (
	"context"
	"crypto/rand"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	slackdomain "github.com/complexus-tech/projects-api/internal/modules/slack/domain"
	platformclock "github.com/complexus-tech/projects-api/internal/platform/clock"
	"github.com/complexus-tech/projects-api/internal/platform/oauthstate"
	"github.com/complexus-tech/projects-api/internal/platform/safecast"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/google/uuid"
)

var (
	ErrNotFound                        = slackdomain.ErrNotFound
	ErrForbidden                       = slackdomain.ErrForbidden
	ErrConflict                        = slackdomain.ErrConflict
	ErrInvalidInput                    = slackdomain.ErrInvalidInput
	ErrSlackNotConfigured              = errors.New("slack integration is not configured")
	ErrSlackSigningSecretNotConfigured = errors.New("slack signing secret is not configured")
	ErrSlackRequestExpired             = errors.New("slack request timestamp is too old")
	ErrSlackInvalidSignature           = errors.New("invalid slack request signature")
	ErrSlackNoWorkspaceLinked          = errors.New("slack workspace is not connected")
	ErrSlackNoTeamsAvailable           = errors.New("no teams are available in this workspace")
	ErrSlackTeamSelectionRequired      = errors.New("team selection is required")
	ErrSlackUserNotLinked              = errors.New("slack user is not linked to a fortyone user")
	ErrSlackTeamNotAvailable           = errors.New("selected team is not available to the slack user")
	ErrSlackInteractionActorMismatch   = errors.New("slack interaction actor does not match the modal source")
	ErrSlackEventRuntimeNotConfigured  = errors.New("slack event runtime is not configured")
	ErrSlackInvalidEventPayload        = errors.New("invalid slack event payload")
	ErrSlackMemberLinkingScopesMissing = errors.New("slack member linking requires users:read and users:read.email; update the Slack connection")
	slackMrkdwnTextEscaper             = strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;")
)

const (
	slackProviderMessaging   = "slack"
	slackNoncePurposeOAuth   = "oauth_install"
	slackNoncePurposeAccount = "account_link"
	slackOAuthNonceTTL       = 10 * time.Minute
	slackAccountLinkNonceTTL = 15 * time.Minute
	slackOpaqueNonceSize     = oauthstate.TokenSize
	slackPriorityNoPriority  = "No Priority"
	slackRequestStatusValue  = "__fortyone_request__"
	slackStatusKindRequest   = "request"
	slackStatusKindStory     = "story"

	modalBlockTeam        = "team"
	modalBlockTitle       = "title"
	modalBlockDescription = "description"
	modalBlockStatus      = "status"
	modalBlockPriority    = "priority"
	modalBlockAssignee    = "assignee"
	modalBlockLabels      = "labels"
	modalBlockObjective   = "objective"

	modalActionTeamSelect        = "team_select"
	modalActionTitleInput        = "title_input"
	modalActionDescriptionInput  = "description_input"
	modalActionStatusSelect      = "status_select"
	modalActionPrioritySelect    = "priority_select"
	modalActionAssigneeSelect    = "assignee_select"
	modalActionLabelsMultiSelect = "labels_multi_select"
	modalActionObjectiveSelect   = "objective_select"

	modalTitleMaxRunes       = 255
	modalDescriptionMaxRunes = 3000
	modalMetadataMaxBytes    = 3000
	modalSourceTextMaxRunes  = 1000

	slackInteractiveWorkTimeout   = 2500 * time.Millisecond
	slackModalHydrationTimeout    = 8 * time.Second
	slackWorkObjectTriggerTimeout = 2500 * time.Millisecond
	slackFailureFeedbackTimeout   = 2 * time.Second
)

type Service struct {
	log                      *logger.Logger
	repo                     Repository
	requests                 RequestStore
	stories                  StoryService
	cfg                      Config
	client                   *http.Client
	clock                    Clock
	random                   io.Reader
	nonces                   NonceStore
	eventGateway             EventGateway
	eventInbox               EventInbox
	outbound                 OutboundStore
	credentials              *credentialCodec
	webClient                *slackWebClient
	mutationConfirmer        storyMutationConfirmer
	objectiveReader          SlackObjectiveReader
	sprintReader             SlackSprintReader
	workObjectTriggerTimeout time.Duration
}

type Option func(*Service)

func WithEventRuntime(gateway EventGateway, inbox EventInbox) Option {
	return func(service *Service) {
		service.eventGateway = gateway
		service.eventInbox = inbox
		service.outbound, _ = inbox.(OutboundStore)
	}
}

func WithNonceStore(store NonceStore) Option {
	return func(service *Service) {
		service.nonces = store
	}
}

func WithMutationConfirmer(confirmer storyMutationConfirmer) Option {
	return func(service *Service) {
		service.mutationConfirmer = confirmer
	}
}

// WithObjectiveReader enables permission-aware objective Work Object details.
func WithObjectiveReader(reader SlackObjectiveReader) Option {
	return func(service *Service) {
		service.objectiveReader = reader
	}
}

// WithSprintReader enables permission-aware sprint Work Object details.
func WithSprintReader(reader SlackSprintReader) Option {
	return func(service *Service) {
		service.sprintReader = reader
	}
}

func New(log *logger.Logger, repo Repository, requests RequestStore, stories StoryService, cfg Config, options ...Option) *Service {
	service := &Service{
		log:      log,
		repo:     repo,
		requests: requests,
		stories:  stories,
		cfg:      cfg,
		client: &http.Client{
			Timeout: 12 * time.Second,
		},
		clock:                    platformclock.System{},
		random:                   rand.Reader,
		workObjectTriggerTimeout: slackWorkObjectTriggerTimeout,
	}
	service.webClient = newSlackWebClient(service.client)
	if codec, err := newCredentialCodec(cfg.CredentialVault); err == nil {
		service.credentials = codec
	}
	for _, option := range options {
		if option != nil {
			option(service)
		}
	}
	return service
}

func (s *Service) GetIntegration(ctx context.Context, workspaceID, userID uuid.UUID) (CoreIntegration, error) {
	if err := s.requireWorkspaceMember(ctx, workspaceID, userID); err != nil {
		return CoreIntegration{}, err
	}
	integration := CoreIntegration{
		Channels: make([]CoreSlackChannel, 0),
	}

	query := slackdomain.WorkspaceActorQuery{WorkspaceID: workspaceID, ActorID: userID}
	slackWorkspace, err := s.repo.GetSlackWorkspaceForMember(ctx, query)
	if err != nil {
		if isSlackRepositoryNotFound(err) {
			return integration, nil
		}
		return CoreIntegration{}, err
	}
	coreWorkspace := toCoreSlackWorkspace(slackWorkspace)
	integration.SlackWorkspace = &coreWorkspace
	if userID != uuid.Nil {
		link, linkErr := s.repo.FindSlackUserLinkForMember(ctx, query, slackWorkspace.SlackTeamID)
		if linkErr != nil {
			return CoreIntegration{}, linkErr
		}
		if link != nil {
			integration.AccountLink = &CoreSlackAccountLink{
				SlackUserID: link.SlackUserID,
				LinkedVia:   link.LinkedVia,
				LinkedAt:    link.LinkedAt,
			}
		}
	}

	channels, err := s.repo.ListChannelsForMember(ctx, query)
	if err != nil && !isSlackRepositoryNotFound(err) {
		return CoreIntegration{}, err
	}
	if err == nil {
		integration.Channels = toCoreChannels(channels)
	}

	return integration, nil
}

func (s *Service) GetRequestLogs(ctx context.Context, workspaceID, actorID uuid.UUID, limit int) ([]CoreRequestLog, error) {
	if err := s.requireWorkspaceAdmin(ctx, workspaceID, actorID); err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	queryLimit, err := safecast.Int32(limit)
	if err != nil {
		return nil, errors.Join(slackdomain.ErrInvalidInput, err)
	}
	rows, err := s.repo.ListRequestLogsForAdmin(ctx, slackdomain.ListRequestLogsQuery{
		WorkspaceID: workspaceID,
		ActorID:     actorID,
		Limit:       queryLimit,
	})
	if err != nil {
		return nil, err
	}
	logs := make([]CoreRequestLog, 0, len(rows))
	for _, row := range rows {
		logs = append(logs, toCoreRequestLog(row))
	}
	return logs, nil
}

func IsNotFound(err error) bool {
	return isSlackRepositoryNotFound(err)
}
