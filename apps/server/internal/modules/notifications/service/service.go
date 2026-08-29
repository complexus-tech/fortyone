package notifications

import (
	"context"

	notificationsdomain "github.com/complexus-tech/projects-api/internal/modules/notifications/domain"
	platformclock "github.com/complexus-tech/projects-api/internal/platform/clock"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/tasks"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
)

// Repository is the caller-owned persistence port. Its commands carry actor,
// workspace/resource scope and finite mutation intent into the SQL boundary.
type Repository interface {
	Create(context.Context, notificationsdomain.NewNotification) (notificationsdomain.Notification, bool, error)
	List(context.Context, notificationsdomain.ListQuery) ([]notificationsdomain.Notification, error)
	CountUnread(context.Context, notificationsdomain.WorkspaceAccess) (int, error)
	Mutate(context.Context, notificationsdomain.NotificationMutation) error
	MutateAll(context.Context, notificationsdomain.WorkspaceMutation) (int, error)
	GetPreferences(context.Context, notificationsdomain.WorkspaceAccess) (notificationsdomain.Preferences, error)
	UpdatePreference(context.Context, notificationsdomain.UpdatePreference) (notificationsdomain.Preferences, error)
	ListPortalFeedback(context.Context, notificationsdomain.PortalListQuery) ([]notificationsdomain.PortalNotification, error)
	CountUnreadPortalFeedback(context.Context, notificationsdomain.PortalAccess) (int, error)
	MarkPortalFeedbackRead(context.Context, notificationsdomain.PortalNotificationMutation) error
	MarkAllPortalFeedbackRead(context.Context, notificationsdomain.PortalMutation) (int, error)
	ListKeyResultAudience(context.Context, notificationsdomain.KeyResultAudienceQuery) ([]notificationsdomain.KeyResultAudienceMember, error)
	GetEmailDelivery(context.Context, notificationsdomain.EmailNotificationQuery) (*notificationsdomain.EmailNotification, error)
	ListEmailDigest(context.Context, notificationsdomain.DeliveryScope) (*notificationsdomain.EmailDigest, error)
	ListDeliveryTeamIDs(context.Context, notificationsdomain.DeliveryScope) ([]uuid.UUID, error)
	MarkEmailSent(context.Context, notificationsdomain.MarkEmailSent) error
}

type TasksService interface {
	EnqueueNotificationEmailDigest(tasks.NotificationEmailDigestPayload, ...asynq.Option) (*asynq.TaskInfo, error)
}

type Option func(*Service)

func WithClock(source platformclock.Clock) Option {
	return func(service *Service) {
		if source != nil {
			service.clock = source
		}
	}
}

type Service struct {
	repo            Repository
	log             *logger.Logger
	redisClient     *redis.Client
	tasksService    TasksService
	clock           platformclock.Clock
	publishRealtime func(context.Context, CoreNotification) error
}

func New(log *logger.Logger, repo Repository, redisClient *redis.Client, tasksService TasksService, options ...Option) *Service {
	service := &Service{
		repo:         repo,
		log:          log,
		redisClient:  redisClient,
		tasksService: tasksService,
		clock:        platformclock.System{},
	}
	for _, option := range options {
		option(service)
	}
	service.publishRealtime = service.publishNotification
	return service
}

var (
	ErrInvalid              = notificationsdomain.ErrInvalid
	ErrNotificationNotFound = notificationsdomain.ErrNotFound
	ErrForbidden            = notificationsdomain.ErrForbidden
	ErrConflict             = notificationsdomain.ErrConflict
)
