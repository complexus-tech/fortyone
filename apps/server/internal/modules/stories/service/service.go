package stories

import (
	"context"
	"errors"
	"time"

	storydomain "github.com/complexus-tech/projects-api/internal/modules/stories/domain"
	"github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/complexus-tech/projects-api/pkg/events"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/publisher"
	"github.com/complexus-tech/projects-api/pkg/tasks"
	"github.com/google/uuid"
)

var (
	ErrDeleteForbidden            = errors.New("story deletion is not permitted")
	ErrInvalidStoryReference      = errors.New("invalid story reference")
	ErrInvalidStoryMediaReference = errors.New("invalid story media reference")
	ErrObjectiveKeyResultMismatch = errors.New("key result does not belong to objective")
	ErrInvalidStoryLabels         = errors.New("one or more labels do not belong to the story's workspace and team")
	ErrStoryChanged               = storydomain.ErrStoryChanged
	ErrCommentWriterUnavailable   = errors.New("comment writer is unavailable")
)

// Repository is the aggregate adapter accepted by the compatibility
// constructor. Every use case immediately narrows it to a caller-owned port;
// persistence implementations therefore expose only domain-oriented methods.
type Repository interface{}

type legacyStoryRepository interface {
	Get(ctx context.Context, id uuid.UUID, workspaceId uuid.UUID) (CoreSingleStory, error)
}

type legacyStoryUpdateRepository interface {
	Update(ctx context.Context, id uuid.UUID, workspaceId uuid.UUID, updates map[string]any) error
}

type legacyStoryMediaRepository interface {
	UpdateWithMediaReconciliation(ctx context.Context, id uuid.UUID, workspaceId uuid.UUID, updates map[string]any, referencedAttachmentIDs []uuid.UUID) ([]uuid.UUID, error)
}

type legacyStoryCreateRepository interface {
	Create(ctx context.Context, story *CoreSingleStory) (CoreSingleStory, error)
}

type legacyActivityWriter interface {
	RecordActivities(ctx context.Context, activities []CoreActivity) ([]CoreActivity, error)
}

type legacyActivityReader interface {
	GetActivitiesWithUser(ctx context.Context, storyID, workspaceID uuid.UUID, page, pageSize int) ([]CoreActivityWithUser, bool, error)
}

type legacyDuplicateStoryRepository interface {
	DuplicateStory(ctx context.Context, originalStoryID uuid.UUID, workspaceId uuid.UUID, userID uuid.UUID) (CoreSingleStory, error)
}

type legacyStatusCategoryRepository interface {
	GetStatusCategory(ctx context.Context, statusID string) (string, error)
}

type legacyAssociationRepository interface {
	AddAssociation(ctx context.Context, fromID, toID uuid.UUID, associationType string, workspaceID uuid.UUID) (CoreStoryAssociation, error)
	UpdateAssociation(ctx context.Context, associationID, fromID, toID uuid.UUID, associationType string, workspaceID uuid.UUID) (CoreStoryAssociation, error)
	RemoveAssociation(ctx context.Context, associationID, workspaceID uuid.UUID) (CoreStoryAssociation, error)
}

type legacyEstimateSchemeRepository interface {
	GetTeamEstimateScheme(ctx context.Context, teamID, workspaceID uuid.UUID) (string, error)
}

type legacyKeyResultRepository interface {
	ResolveKeyResult(ctx context.Context, keyResultID, workspaceID uuid.UUID) (CoreKeyResultReference, error)
}

type idempotentCreateRepository interface {
	CreateIdempotent(ctx context.Context, story *CoreSingleStory) (CoreSingleStory, bool, error)
}

type conditionalUpdateRepository interface {
	UpdateIfUnchanged(ctx context.Context, id, workspaceID uuid.UUID, expectedUpdatedAt time.Time, updates map[string]any) (bool, error)
}

type autoSchedulingRepository interface {
	MayaScheduleBlocksExist(ctx context.Context, storyID, workspaceID uuid.UUID) (bool, error)
	UpdateAutoSchedulingStateIfUnchanged(
		ctx context.Context,
		storyID, workspaceID uuid.UUID,
		expectedUpdatedAt time.Time,
		status string,
		reason *string,
		stateUpdatedAt time.Time,
		locked *bool,
	) (bool, error)
}

type authorizedAutoSchedulingRepository interface {
	AuthorizedMayaScheduleBlocksExist(context.Context, storydomain.MutationScope, uuid.UUID) (bool, error)
	UpdateAuthorizedAutoSchedulingStateIfUnchanged(
		context.Context,
		storydomain.MutationScope,
		uuid.UUID,
		time.Time,
		string,
		*string,
		time.Time,
		*bool,
	) (bool, error)
}

type defaultStatusRepository interface {
	FindFirstStatusByCategory(ctx context.Context, scope StoryReadScope, teamID uuid.UUID, category string) (*uuid.UUID, error)
}

// CommentCreator is a compatibility port for callers that still reach comment
// creation through the stories service. Persistence and authorization are
// owned entirely by the comments module.
type CommentCreator interface {
	CreateComment(context.Context, CreateCommentCommand) (CoreComment, error)
}

type CoreSingleStoryWithSubs struct {
	CoreSingleStory
	SubStories []CoreStoryList `json:"subStories"`
}

// Service provides story-related operations.
type Service struct {
	repo                      Repository
	commentCreator            CommentCreator
	log                       *logger.Logger
	publisher                 eventPublisher
	tasksService              *tasks.Service
	mayaActorID               uuid.UUID
	mayaAssignment            *mayaAssignmentPolicy
	autoSchedulingEligibility AutoSchedulingEligibilityChecker
}

type eventPublisher interface {
	Publish(ctx context.Context, event events.Event) error
}

type createOptions struct {
	publishEvents         bool
	enqueueGitHubSync     bool
	mutationEventDelivery mutationEventDelivery
	traceStoryTitle       bool
	actorKind             auth.PrincipalKind
}

type updateOptions struct {
	publishEvents bool
	// publishStatusEvents lets provider-originated status transitions reach
	// downstream consumers without enabling outbound provider synchronization.
	publishStatusEvents      bool
	enqueueGitHubSync        bool
	recordDescriptionUpdates bool
	activityReason           string
	reconcileMedia           bool
	referencedMediaIDs       []uuid.UUID
	orphanedMediaIDs         *[]uuid.UUID
	expectedUpdatedAt        *time.Time
	eventSource              events.StoryUpdateSource
	eventReason              string
	eventSchedule            *events.StoryScheduleTransition
	automationMutation       bool
	actorKind                auth.PrincipalKind
}

// New constructs a new stories service instance with the provided repository.
func New(log *logger.Logger, repo Repository, eventPublisher *publisher.Publisher, tasksService *tasks.Service) *Service {
	service := &Service{
		repo:         repo,
		log:          log,
		tasksService: tasksService,
	}
	if eventPublisher != nil {
		service.publisher = eventPublisher
	}
	return service
}

func optionalUUIDUpdate(value any) (*uuid.UUID, bool) {
	switch typed := value.(type) {
	case nil:
		return nil, true
	case uuid.UUID:
		if typed == uuid.Nil {
			return nil, true
		}
		return &typed, true
	case *uuid.UUID:
		if typed == nil || *typed == uuid.Nil {
			return nil, true
		}
		return typed, true
	default:
		return nil, false
	}
}

func sameOptionalUUID(left, right *uuid.UUID) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return *left == *right
}
