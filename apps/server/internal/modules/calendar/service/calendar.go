package calendar

import (
	"context"
	"crypto/rand"
	"time"

	calendardomain "github.com/complexus-tech/projects-api/internal/modules/calendar/domain"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/google/uuid"
)

var (
	ErrCalendarNotConfigured           = calendardomain.ErrCalendarNotConfigured
	ErrInvalidCalendarState            = calendardomain.ErrInvalidCalendarState
	ErrCalendarNotFound                = calendardomain.ErrCalendarNotFound
	ErrCalendarAccessDenied            = calendardomain.ErrCalendarAccessDenied
	ErrCalendarCredentialsIncomplete   = calendardomain.ErrCalendarCredentialsIncomplete
	ErrCalendarEventNotFound           = calendardomain.ErrCalendarEventNotFound
	ErrCalendarSyncSuperseded          = calendardomain.ErrCalendarSyncSuperseded
	ErrInvalidScheduleRange            = calendardomain.ErrInvalidScheduleRange
	ErrInvalidScheduleBlock            = calendardomain.ErrInvalidScheduleBlock
	ErrCalendarScheduleConflict        = calendardomain.ErrCalendarScheduleConflict
	ErrCalendarScheduleStalePlan       = calendardomain.ErrCalendarScheduleStalePlan
	ErrCalendarScheduleBlockNotFound   = calendardomain.ErrCalendarScheduleBlockNotFound
	ErrManagedScheduleBlock            = calendardomain.ErrManagedScheduleBlock
	ErrCalendarCleanupPending          = calendardomain.ErrCalendarCleanupPending
	ErrCalendarAccountChangePending    = calendardomain.ErrCalendarAccountChangePending
	ErrCalendarFullSyncRequired        = calendardomain.ErrCalendarFullSyncRequired
	ErrInvalidCalendarNotification     = calendardomain.ErrInvalidCalendarNotification
	ErrCalendarWebhookNotConfigured    = calendardomain.ErrCalendarWebhookNotConfigured
	ErrCalendarReauthorizationRequired = calendardomain.ErrCalendarReauthorizationRequired
)

const (
	connectStateTTL                    = 10 * time.Minute
	defaultSyncLookback                = -7 * 24 * time.Hour
	defaultSyncLookahead               = 90 * 24 * time.Hour
	googleWatchTTL                     = 7 * 24 * time.Hour
	googleWatchRenewalWindow           = 24 * time.Hour
	maximumScheduleEventOutboxAttempts = 8
)

type CalendarTaskQueue interface {
	EnqueueCalendarSync(ctx context.Context, connectionID uuid.UUID) error
}

type CalendarScheduleTaskQueue interface {
	EnqueueCalendarScheduleReconcile(ctx context.Context, userID uuid.UUID) error
}

type StoryScheduleTaskQueue interface {
	EnqueueStoryScheduleReconcile(ctx context.Context, workspaceID, storyID uuid.UUID) error
}

type CalendarUpdatePublisher interface {
	PublishCalendarUpdated(ctx context.Context, workspaceID, userID, connectionID uuid.UUID, syncedAt time.Time) error
}

type Repository interface {
	ListConnections(ctx context.Context, workspaceID uuid.UUID, userID *uuid.UUID) ([]CoreConnection, error)
	GetOwnedConnection(ctx context.Context, workspaceID, userID, connectionID uuid.UUID) (CoreConnection, error)
	GetActiveConnection(ctx context.Context, workspaceID, userID uuid.UUID, provider Provider) (CoreConnection, error)
	GetConnection(ctx context.Context, connectionID uuid.UUID) (CoreConnection, error)
	GetScheduleEventDispatchConnection(ctx context.Context, userID uuid.UUID) (CoreConnection, bool, error)
	ListConnectionsNeedingWatch(ctx context.Context, renewBefore time.Time) ([]CoreConnection, error)
	WorkspaceMemberExists(ctx context.Context, workspaceID, userID uuid.UUID) (bool, error)
	UpsertConnection(ctx context.Context, input CoreConnectionUpsert) (CoreConnection, error)
	SetPrimaryConnection(ctx context.Context, workspaceID, userID, connectionID uuid.UUID) (CoreConnection, error)
	UpdateConnectionToken(ctx context.Context, connection CoreConnection, tokenPayload string) error
	BeginConnectionSync(ctx context.Context, connection CoreConnection) (CoreConnection, error)
	RevokeConnection(ctx context.Context, workspaceID, userID, connectionID uuid.UUID) error
	ReplaceCalendarSnapshot(ctx context.Context, connection CoreConnection, snapshot CalendarSyncSnapshot) error
	ApplyCalendarChanges(ctx context.Context, connection CoreConnection, delta CalendarSyncDelta) error
	SetNotificationChannel(ctx context.Context, connection CoreConnection, channel CalendarWatchChannel) error
	ClearNotificationChannel(ctx context.Context, connectionID uuid.UUID) error
	ListCalendarEvents(ctx context.Context, workspaceID, userID uuid.UUID, startAt, endAt time.Time) ([]CoreCalendarEventSummary, error)
	GetCalendarEvent(ctx context.Context, workspaceID, userID, eventID uuid.UUID) (CoreCalendarEvent, error)
	ListBusyWindows(ctx context.Context, workspaceID, userID uuid.UUID, startAt, endAt time.Time) ([]CoreBusyWindow, error)
	ListScheduleBlocks(ctx context.Context, workspaceID, userID uuid.UUID, startAt, endAt time.Time) ([]CoreScheduleBlock, error)
	GetScheduleBlock(ctx context.Context, workspaceID, userID, blockID uuid.UUID) (CoreScheduleBlock, error)
	ScheduleStoryExists(ctx context.Context, workspaceID, userID, storyID uuid.UUID) (bool, error)
	CreateScheduleBlock(ctx context.Context, input CoreScheduleBlockInput) (CoreScheduleBlock, error)
	UpdateScheduleBlock(ctx context.Context, input CoreScheduleBlockInput) (CoreScheduleBlock, error)
	DeleteScheduleBlock(ctx context.Context, workspaceID, userID, blockID uuid.UUID) error
	MarkConnectionSynced(ctx context.Context, workspaceID, connectionID, credentialGeneration uuid.UUID, syncedAt time.Time) error
	MarkConnectionSyncFailed(ctx context.Context, workspaceID, connectionID, credentialGeneration uuid.UUID, message string) error
}

type ScheduleReconciliationRepository interface {
	ListSchedulingBlocksForUser(ctx context.Context, workspaceID, userID uuid.UUID, startAt, endAt time.Time) ([]CoreScheduleBlock, error)
	ListMayaScheduleBlocksForStory(ctx context.Context, workspaceID, userID, storyID uuid.UUID) ([]CoreScheduleBlock, error)
	MayaScheduleOwnershipExists(ctx context.Context, workspaceID, userID, storyID uuid.UUID) (bool, error)
	ReconcileMayaScheduleBlocks(ctx context.Context, input MayaScheduleReconcileInput) (CoreScheduleReconcileResult, error)
	ListReadyScheduleEventOutboxUsers(ctx context.Context, limit int) ([]uuid.UUID, error)
	WithScheduleEventDispatchLock(ctx context.Context, userID uuid.UUID, dispatch func(ScheduleEventOutboxStore) error) error
}

type ScheduleIssueRepository interface {
	ListScheduleIssues(ctx context.Context, workspaceID, userID uuid.UUID) ([]CoreScheduleIssue, error)
}

type ScheduleEventOutboxStore = calendardomain.ScheduleEventOutboxStore

type Config struct {
	SecretKey      string
	WebsiteURL     string
	WebhookURL     string
	WebhookURLs    map[Provider]string
	RequireWebhook bool
	Providers      map[Provider]CalendarProvider
	Tasks          CalendarTaskQueue
	Updates        CalendarUpdatePublisher
}

type Service struct {
	log       *logger.Logger
	repo      Repository
	cfg       Config
	now       func() time.Time
	randBytes func([]byte) (int, error)
}

func New(log *logger.Logger, repo Repository, cfg Config) *Service {
	return &Service{
		log:       log,
		repo:      repo,
		cfg:       cfg,
		now:       time.Now,
		randBytes: rand.Read,
	}
}
