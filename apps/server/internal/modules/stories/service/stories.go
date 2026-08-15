package stories

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	comments "github.com/complexus-tech/projects-api/internal/modules/comments/service"
	links "github.com/complexus-tech/projects-api/internal/modules/links/service"
	"github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/complexus-tech/projects-api/pkg/events"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/publisher"
	"github.com/complexus-tech/projects-api/pkg/tasks"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

var (
	ErrNotFound                   = errors.New("story not found")
	ErrInvalidStoryReference      = errors.New("invalid story reference")
	ErrInvalidStoryMediaReference = errors.New("invalid story media reference")
	ErrObjectiveKeyResultMismatch = errors.New("key result does not belong to objective")
	ErrInvalidStoryLabels         = errors.New("one or more labels do not belong to the story's workspace and team")
	ErrStoryChanged               = errors.New("story changed before the update was applied")
)

// Repository provides access to the story storage.
type Repository interface {
	Get(ctx context.Context, id uuid.UUID, workspaceId uuid.UUID) (CoreSingleStory, error)
	Delete(ctx context.Context, id uuid.UUID, workspaceId uuid.UUID) error
	BulkDelete(ctx context.Context, ids []uuid.UUID, workspaceId uuid.UUID) error
	HardBulkDelete(ctx context.Context, ids []uuid.UUID, workspaceId uuid.UUID) ([]uuid.UUID, error)
	Restore(ctx context.Context, id uuid.UUID, workspaceId uuid.UUID) error
	BulkRestore(ctx context.Context, ids []uuid.UUID, workspaceId uuid.UUID) error
	BulkArchive(ctx context.Context, ids []uuid.UUID, workspaceId uuid.UUID) error
	BulkUnarchive(ctx context.Context, ids []uuid.UUID, workspaceId uuid.UUID) error
	Update(ctx context.Context, id uuid.UUID, workspaceId uuid.UUID, updates map[string]any) error
	UpdateWithMediaReconciliation(ctx context.Context, id uuid.UUID, workspaceId uuid.UUID, updates map[string]any, referencedAttachmentIDs []uuid.UUID) ([]uuid.UUID, error)
	UpdateLabels(ctx context.Context, id uuid.UUID, workspaceId uuid.UUID, labels []uuid.UUID) error
	GetStoryLinks(ctx context.Context, storyID uuid.UUID) ([]links.CoreLink, error)
	Create(ctx context.Context, story *CoreSingleStory) (CoreSingleStory, error)
	GetNextSequenceID(ctx context.Context, teamId uuid.UUID, workspaceId uuid.UUID) (int, func() error, func() error, error)
	MyStories(ctx context.Context, workspaceId uuid.UUID) ([]CoreStoryList, error)
	GetSubStories(ctx context.Context, parentId uuid.UUID, workspaceId uuid.UUID) ([]CoreStoryList, error)
	RecordActivities(ctx context.Context, activities []CoreActivity) ([]CoreActivity, error)
	GetActivitiesWithUser(ctx context.Context, storyID uuid.UUID, page, pageSize int) ([]CoreActivityWithUser, bool, error)
	CreateComment(ctx context.Context, comment CoreNewComment) (comments.CoreComment, error)
	GetComments(ctx context.Context, storyID uuid.UUID, page, pageSize int) ([]comments.CoreComment, bool, error)
	GetComment(ctx context.Context, commentID uuid.UUID) (comments.CoreComment, error)
	DuplicateStory(ctx context.Context, originalStoryID uuid.UUID, workspaceId uuid.UUID, userID uuid.UUID) (CoreSingleStory, error)
	CountStoriesInWorkspace(ctx context.Context, workspaceId uuid.UUID) (int, error)
	List(ctx context.Context, workspaceId uuid.UUID, filters map[string]any) ([]CoreStoryList, error)
	ListGroupedStories(ctx context.Context, query CoreStoryQuery) ([]CoreStoryGroup, error)
	ListGroupStories(ctx context.Context, groupKey string, query CoreStoryQuery) ([]CoreStoryList, bool, error)
	ListByCategory(ctx context.Context, workspaceId, userID, teamId uuid.UUID, category string, page, pageSize int, showSubStories bool) ([]CoreStoryList, bool, error)
	GetStatusCategory(ctx context.Context, statusID string) (string, error)
	QueryByRef(ctx context.Context, workspaceId uuid.UUID, teamCode string, sequenceID int) (CoreSingleStory, error)
	AddAssociation(ctx context.Context, fromID, toID uuid.UUID, associationType string, workspaceID uuid.UUID) (CoreStoryAssociation, error)
	UpdateAssociation(ctx context.Context, associationID, fromID, toID uuid.UUID, associationType string, workspaceID uuid.UUID) (CoreStoryAssociation, error)
	RemoveAssociation(ctx context.Context, associationID, workspaceID uuid.UUID) (CoreStoryAssociation, error)
	GetTeamEstimateScheme(ctx context.Context, teamID, workspaceID uuid.UUID) (string, error)
	ResolveKeyResult(ctx context.Context, keyResultID, workspaceID uuid.UUID) (CoreKeyResultReference, error)
	GetCollaborators(ctx context.Context, storyID, workspaceID uuid.UUID) ([]uuid.UUID, error)
	SetCollaborators(ctx context.Context, storyID, workspaceID uuid.UUID, collaboratorIDs []uuid.UUID) error
	SetWatching(ctx context.Context, storyID, workspaceID, userID uuid.UUID, watching bool) error
	GetNotificationAudience(ctx context.Context, storyID, workspaceID uuid.UUID) ([]uuid.UUID, error)
}

type idempotentCreateRepository interface {
	CreateIdempotent(ctx context.Context, story *CoreSingleStory) (CoreSingleStory, bool, error)
}

type conditionalUpdateRepository interface {
	UpdateIfUnchanged(ctx context.Context, id, workspaceID uuid.UUID, expectedUpdatedAt time.Time, updates map[string]any) (bool, error)
}

type defaultStatusRepository interface {
	FindFirstStatusByCategory(ctx context.Context, teamID, workspaceID uuid.UUID, category string) (*uuid.UUID, error)
}

// MentionsRepository provides access to comment mentions storage.
type MentionsRepository interface {
	SaveMentions(ctx context.Context, commentID uuid.UUID, userIDs []uuid.UUID) error
	DeleteMentions(ctx context.Context, commentID uuid.UUID) error
	GetMentions(ctx context.Context, commentID uuid.UUID) ([]uuid.UUID, error)
}

type CoreSingleStoryWithSubs struct {
	CoreSingleStory
	SubStories []CoreStoryList `json:"subStories"`
}

// Service provides story-related operations.
type Service struct {
	repo           Repository
	mentionsRepo   MentionsRepository
	log            *logger.Logger
	publisher      eventPublisher
	tasksService   *tasks.Service
	mayaAssignment *mayaAssignmentPolicy
}

type eventPublisher interface {
	Publish(ctx context.Context, event events.Event) error
}

type createOptions struct {
	publishEvents     bool
	enqueueGitHubSync bool
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
}

type commentOptions struct {
	actorID uuid.UUID
}

// New constructs a new stories service instance with the provided repository.
func New(log *logger.Logger, repo Repository, mentionsRepo MentionsRepository, eventPublisher *publisher.Publisher, tasksService *tasks.Service) *Service {
	service := &Service{
		repo:         repo,
		mentionsRepo: mentionsRepo,
		log:          log,
		tasksService: tasksService,
	}
	if eventPublisher != nil {
		service.publisher = eventPublisher
	}
	return service
}

// Create creates a new story.
func (s *Service) Create(ctx context.Context, ns CoreNewStory, workspaceId uuid.UUID) (CoreSingleStory, error) {
	actorID := uuid.Nil
	if ns.Reporter != nil {
		actorID = *ns.Reporter
	}
	return s.createWithOptions(ctx, ns, workspaceId, actorID, createOptions{
		publishEvents:     true,
		enqueueGitHubSync: true,
	})
}

func (s *Service) CreateExternal(ctx context.Context, actorID uuid.UUID, ns CoreNewStory, workspaceID uuid.UUID) (CoreSingleStory, error) {
	if ns.Reporter == nil {
		ns.Reporter = &actorID
	}
	return s.createWithOptions(ctx, ns, workspaceID, actorID, createOptions{})
}

// CreateExternalUserAction creates a story from a user-initiated external
// surface such as Slack. Unlike provider ingestion, this preserves the normal
// FortyOne event and GitHub-sync side effects while retaining the explicit
// actor and external idempotency key.
func (s *Service) CreateExternalUserAction(ctx context.Context, actorID uuid.UUID, ns CoreNewStory, workspaceID uuid.UUID) (CoreSingleStory, error) {
	if ns.Reporter == nil {
		ns.Reporter = &actorID
	}
	return s.createWithOptions(ctx, ns, workspaceID, actorID, createOptions{
		publishEvents:     true,
		enqueueGitHubSync: true,
	})
}

func (s *Service) createWithOptions(ctx context.Context, ns CoreNewStory, workspaceId, actorID uuid.UUID, options createOptions) (CoreSingleStory, error) {
	s.log.Info(ctx, "business.core.stories.create")
	ctx, span := web.AddSpan(ctx, "business.core.stories.Create")
	defer span.End()

	if actorID == uuid.Nil {
		if ns.Reporter == nil {
			return CoreSingleStory{}, fmt.Errorf("actor ID is required to create a story")
		}
		actorID = *ns.Reporter
	}
	if ns.Reporter == nil {
		ns.Reporter = &actorID
	}
	if ns.KeyResult != nil {
		keyResult, err := s.repo.ResolveKeyResult(ctx, *ns.KeyResult, workspaceId)
		if err != nil {
			return CoreSingleStory{}, fmt.Errorf("resolve key result objective: %w", err)
		}
		if ns.Objective != nil && *ns.Objective != keyResult.ObjectiveID {
			return CoreSingleStory{}, fmt.Errorf(
				"%w: key result %s belongs to objective %s, not %s",
				ErrObjectiveKeyResultMismatch,
				*ns.KeyResult,
				keyResult.ObjectiveID,
				*ns.Objective,
			)
		}
		ns.Objective = &keyResult.ObjectiveID
	}

	story := toCoreSingleStory(ns, workspaceId)
	estimateScheme, err := s.repo.GetTeamEstimateScheme(ctx, ns.Team, workspaceId)
	if err != nil {
		span.RecordError(err)
		return CoreSingleStory{}, err
	}

	story.EstimateScheme = estimateScheme
	if err := ValidateEstimateValue(estimateScheme, ns.EstimateValue); err != nil {
		span.RecordError(err)
		if ns.EstimateValue != nil {
			return CoreSingleStory{}, fmt.Errorf("%w. If this work is larger than the max estimate, split it into smaller stories", err)
		}
		return CoreSingleStory{}, err
	}
	story.EstimateValue = ns.EstimateValue
	story.EstimateLabel = EstimateLabelFromValue(estimateScheme, ns.EstimateValue)
	if err := ValidateStoryTimeContract(ns.EstimatedDurationMinutes, ns.MinimumFocusBlockMinutes); err != nil {
		span.RecordError(err)
		return CoreSingleStory{}, err
	}
	if err := s.validateMayaAssignment(ctx, story, nil, actorID); err != nil {
		span.RecordError(err)
		return CoreSingleStory{}, err
	}

	created := true
	var cs CoreSingleStory
	if ns.CreationKey != nil {
		idempotentRepo, ok := s.repo.(idempotentCreateRepository)
		if !ok {
			return CoreSingleStory{}, errors.New("story repository does not support idempotent creation")
		}
		cs, created, err = idempotentRepo.CreateIdempotent(ctx, &story)
	} else {
		cs, err = s.repo.Create(ctx, &story)
	}
	if err != nil {
		span.RecordError(err)
		return CoreSingleStory{}, err
	}
	cs.EstimateScheme = estimateScheme
	cs.EstimateLabel = EstimateLabelFromValue(estimateScheme, cs.EstimateValue)
	cs.CreatedNow = created
	if !created {
		return cs, nil
	}

	// Record in the activity log
	ca := CoreActivity{
		StoryID:      cs.ID,
		Type:         "create",
		Field:        "story",
		CurrentValue: cs.Title,
		NewValue:     cs.Title,
		UserID:       actorID,
		WorkspaceID:  workspaceId,
	}
	if _, err := s.repo.RecordActivities(ctx, []CoreActivity{ca}); err != nil {
		span.RecordError(err)
	}

	if options.publishEvents {
		payload := events.StoryCreatedPayload{
			StoryID:     cs.ID,
			WorkspaceID: workspaceId,
			Title:       cs.Title,
			AssigneeID:  cs.Assignee,
			ReporterID:  *ns.Reporter,
		}

		event := events.Event{
			Type:      events.StoryCreated,
			Payload:   payload,
			Timestamp: time.Now(),
			ActorID:   actorID,
		}

		if err := s.publisher.Publish(context.Background(), event); err != nil {
			s.log.Error(ctx, "failed to publish story created event", "error", err)
			// Don't return error as this is not critical
		}
	}
	if options.enqueueGitHubSync {
		s.enqueueGitHubStorySync(ctx, cs.ID, workspaceId)
	}
	span.AddEvent("story created.", trace.WithAttributes(
		attribute.String("story.title", cs.Title),
	))
	return cs, nil
}

// UpdateLabels replaces the labels for a story.
func (s *Service) UpdateLabels(ctx context.Context, id uuid.UUID, workspaceId uuid.UUID, labels []uuid.UUID) error {
	s.log.Info(ctx, "business.core.stories.UpdateLabels")
	ctx, span := web.AddSpan(ctx, "business.core.stories.UpdateLabels")
	defer span.End()

	if err := s.repo.UpdateLabels(ctx, id, workspaceId, labels); err != nil {
		span.RecordError(err)
		return err
	}
	actorID, _ := auth.GetUserID(ctx)
	if err := s.RecordActivity(ctx, CoreActivity{
		StoryID:      id,
		Type:         "update",
		Field:        "labels",
		CurrentValue: s.formatLabelActivityValue(labels),
		NewValue:     labels,
		UserID:       actorID,
		WorkspaceID:  workspaceId,
	}); err != nil {
		span.RecordError(err)
	}
	return nil
}

// UpdateCollaborators replaces the active collaborators for a story.
func (s *Service) UpdateCollaborators(ctx context.Context, storyID, workspaceID uuid.UUID, collaboratorIDs []uuid.UUID) error {
	actorID, _ := auth.GetUserID(ctx)

	story, err := s.repo.Get(ctx, storyID, workspaceID)
	if err != nil {
		return err
	}

	previousIDs, err := s.repo.GetCollaborators(ctx, storyID, workspaceID)
	if err != nil {
		return err
	}
	nextIDs := uniqueUUIDs(collaboratorIDs)
	if sameUUIDSet(previousIDs, nextIDs) {
		return nil
	}

	if err := s.repo.SetCollaborators(ctx, storyID, workspaceID, nextIDs); err != nil {
		return err
	}

	activity := CoreActivity{
		StoryID:      storyID,
		Type:         "update",
		Field:        "collaborator_ids",
		CurrentValue: s.formatValue(nextIDs),
		OldValue:     previousIDs,
		NewValue:     nextIDs,
		UserID:       actorID,
		WorkspaceID:  workspaceID,
	}
	if _, err := s.repo.RecordActivities(ctx, []CoreActivity{activity}); err != nil {
		s.log.Error(ctx, "failed to record collaborator activity", "error", err, "story_id", storyID)
	}

	audienceIDs, err := s.repo.GetNotificationAudience(ctx, storyID, workspaceID)
	audienceResolved := err == nil
	if err != nil {
		s.log.Error(ctx, "failed to load story notification audience", "error", err, "story_id", storyID)
		audienceIDs = nil
	}
	if s.publisher != nil {
		event := events.Event{
			Type: events.StoryUpdated,
			Payload: events.StoryUpdatedPayload{
				StoryID:                 storyID,
				WorkspaceID:             workspaceID,
				Updates:                 map[string]any{"collaborator_ids": nextIDs},
				AssigneeID:              story.Assignee,
				AudienceIDs:             audienceIDs,
				AudienceResolved:        audienceResolved,
				PreviousCollaboratorIDs: previousIDs,
			},
			Timestamp: time.Now(),
			ActorID:   actorID,
		}
		if err := s.publisher.Publish(context.Background(), event); err != nil {
			s.log.Error(ctx, "failed to publish collaborators updated event", "error", err)
		}
	}
	return nil
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

// SetWatching updates the current user's explicit or muted story subscription.
func (s *Service) SetWatching(ctx context.Context, storyID, workspaceID, userID uuid.UUID, watching bool) error {
	return s.repo.SetWatching(ctx, storyID, workspaceID, userID, watching)
}

// GetNotificationAudience returns the non-muted users following a story.
func (s *Service) GetNotificationAudience(ctx context.Context, storyID, workspaceID uuid.UUID) ([]uuid.UUID, error) {
	return s.repo.GetNotificationAudience(ctx, storyID, workspaceID)
}

// GetStoryLinks returns the links for a story.
func (s *Service) GetStoryLinks(ctx context.Context, storyID uuid.UUID) ([]links.CoreLink, error) {
	s.log.Info(ctx, "business.core.stories.GetStoryLinks")
	ctx, span := web.AddSpan(ctx, "business.core.stories.GetStoryLinks")
	defer span.End()

	links, err := s.repo.GetStoryLinks(ctx, storyID)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	span.AddEvent("links retrieved.", trace.WithAttributes(
		attribute.Int("link.count", len(links)),
	))

	return links, nil
}

// MyStories returns a list of stories.
func (s *Service) MyStories(ctx context.Context, workspaceId uuid.UUID) ([]CoreStoryList, error) {
	s.log.Info(ctx, "business.core.stories.list")
	ctx, span := web.AddSpan(ctx, "business.core.stories.List")
	defer span.End()

	stories, err := s.repo.MyStories(ctx, workspaceId)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}
	if err := s.enrichStoryListEstimates(ctx, workspaceId, stories); err != nil {
		span.RecordError(err)
		return nil, err
	}

	span.AddEvent("stories retrieved.", trace.WithAttributes(
		attribute.Int("story.count", len(stories)),
	))
	return stories, nil
}

// List returns a list of stories for a workspace with additional filters.

// Get returns the story with the specified ID.
func (s *Service) Get(ctx context.Context, id uuid.UUID, workspaceId uuid.UUID) (CoreSingleStory, error) {
	s.log.Info(ctx, "business.core.stories.Get")
	ctx, span := web.AddSpan(ctx, "business.core.stories.Get")
	defer span.End()

	story, err := s.repo.Get(ctx, id, workspaceId)
	if err != nil {
		span.RecordError(err)
		return CoreSingleStory{}, err
	}
	if err := s.enrichSingleStoryEstimate(ctx, workspaceId, &story); err != nil {
		span.RecordError(err)
		return CoreSingleStory{}, err
	}

	return story, nil
}

// FindFirstStatusByCategory returns the first workflow status in a category for
// a team. The workspace constraint prevents a caller from resolving workflow
// state through a cross-workspace team identifier.
func (s *Service) FindFirstStatusByCategory(ctx context.Context, teamID, workspaceID uuid.UUID, category string) (*uuid.UUID, error) {
	repo, ok := s.repo.(defaultStatusRepository)
	if !ok {
		return nil, errors.New("story repository does not support default status lookup")
	}
	if teamID == uuid.Nil || workspaceID == uuid.Nil || strings.TrimSpace(category) == "" {
		return nil, errors.New("team, workspace, and status category are required")
	}
	return repo.FindFirstStatusByCategory(ctx, teamID, workspaceID, strings.TrimSpace(category))
}

// List returns a list of stories for a workspace with additional filters.
func (s *Service) List(ctx context.Context, workspaceId uuid.UUID, filters map[string]any) ([]CoreStoryList, error) {
	s.log.Info(ctx, "business.core.stories.List")
	ctx, span := web.AddSpan(ctx, "business.core.stories.List")
	defer span.End()

	stories, err := s.repo.List(ctx, workspaceId, filters)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}
	if err := s.enrichStoryListEstimates(ctx, workspaceId, stories); err != nil {
		span.RecordError(err)
		return nil, err
	}
	span.AddEvent("stories retrieved.", trace.WithAttributes(
		attribute.Int("story.count", len(stories)),
	))
	return stories, nil
}

// Delete deletes the story with the specified ID.
func (s *Service) Delete(ctx context.Context, id uuid.UUID, workspaceId uuid.UUID) error {
	s.log.Info(ctx, "business.core.stories.Delete")
	ctx, span := web.AddSpan(ctx, "business.core.stories.Delete")
	defer span.End()

	if err := s.repo.Delete(ctx, id, workspaceId); err != nil {
		span.RecordError(err)
		return err
	}
	return nil
}

// Update updates a story.
func (s *Service) Update(ctx context.Context, storyID, workspaceID uuid.UUID, updates map[string]any) error {
	actorID, _ := auth.GetUserID(ctx)
	return s.updateWithOptions(ctx, storyID, workspaceID, actorID, updates, updateOptions{
		publishEvents:            true,
		enqueueGitHubSync:        true,
		recordDescriptionUpdates: false,
	})
}

// UpdateWithMediaReconciliation applies an authoritative description snapshot.
// The caller must only use this after all editor uploads have settled. Ordinary
// updates intentionally do not reconcile media so an older autosave that omits
// a pending upload cannot unlink it.
func (s *Service) UpdateWithMediaReconciliation(
	ctx context.Context,
	storyID, workspaceID uuid.UUID,
	updates map[string]any,
	referencedMediaIDs []uuid.UUID,
) ([]uuid.UUID, error) {
	if storyID == uuid.Nil || workspaceID == uuid.Nil {
		return nil, ErrInvalidStoryMediaReference
	}
	if _, hasDescriptionHTML := updates["description_html"].(string); !hasDescriptionHTML {
		return nil, ErrInvalidStoryMediaReference
	}
	seenMediaIDs := make(map[uuid.UUID]struct{}, len(referencedMediaIDs))
	deduplicatedMediaIDs := make([]uuid.UUID, 0, len(referencedMediaIDs))
	for _, attachmentID := range referencedMediaIDs {
		if attachmentID == uuid.Nil {
			return nil, ErrInvalidStoryMediaReference
		}
		if _, seen := seenMediaIDs[attachmentID]; seen {
			continue
		}
		seenMediaIDs[attachmentID] = struct{}{}
		deduplicatedMediaIDs = append(deduplicatedMediaIDs, attachmentID)
	}

	actorID, _ := auth.GetUserID(ctx)
	orphanedMediaIDs := []uuid.UUID{}
	err := s.updateWithOptions(ctx, storyID, workspaceID, actorID, updates, updateOptions{
		publishEvents:            true,
		enqueueGitHubSync:        true,
		recordDescriptionUpdates: false,
		reconcileMedia:           true,
		referencedMediaIDs:       deduplicatedMediaIDs,
		orphanedMediaIDs:         &orphanedMediaIDs,
	})
	if err != nil {
		return nil, err
	}
	return orphanedMediaIDs, nil
}

func (s *Service) UpdateExternal(ctx context.Context, actorID, storyID, workspaceID uuid.UUID, updates map[string]any) error {
	return s.UpdateExternalWithReason(ctx, actorID, storyID, workspaceID, updates, "")
}

// UpdateExternalWithReason applies a provider-originated update. Actual status
// transitions publish a domain event, but never enqueue a GitHub write-back.
func (s *Service) UpdateExternalWithReason(ctx context.Context, actorID, storyID, workspaceID uuid.UUID, updates map[string]any, reason string) error {
	return s.updateWithOptions(ctx, storyID, workspaceID, actorID, updates, updateOptions{
		recordDescriptionUpdates: true,
		activityReason:           reason,
		publishStatusEvents:      true,
	})
}

// UpdateExternalIfUnchanged applies an integration-originated update only when
// the story still has the version the caller inspected. This closes the race
// between confirmation-time validation and the repository write.
func (s *Service) UpdateExternalIfUnchanged(
	ctx context.Context,
	actorID, storyID, workspaceID uuid.UUID,
	expectedUpdatedAt time.Time,
	updates map[string]any,
) error {
	return s.UpdateExternalWithReasonIfUnchanged(ctx, actorID, storyID, workspaceID, expectedUpdatedAt, updates, "")
}

// UpdateExternalWithReasonIfUnchanged preserves the inspected story version
// across a reason-aware integration update. It prevents an automated actor
// from overwriting a user edit that committed after planning.
func (s *Service) UpdateExternalWithReasonIfUnchanged(
	ctx context.Context,
	actorID, storyID, workspaceID uuid.UUID,
	expectedUpdatedAt time.Time,
	updates map[string]any,
	reason string,
) error {
	if expectedUpdatedAt.IsZero() {
		return errors.New("expected story update time is required")
	}
	expectedUpdatedAt = expectedUpdatedAt.UTC()
	return s.updateWithOptions(ctx, storyID, workspaceID, actorID, updates, updateOptions{
		recordDescriptionUpdates: true,
		activityReason:           reason,
		expectedUpdatedAt:        &expectedUpdatedAt,
		publishStatusEvents:      true,
	})
}

// UpdateExternalUserActionIfUnchanged applies a user-initiated external edit
// with compare-and-swap protection and the same downstream event and GitHub
// synchronization behavior as an in-app update.
func (s *Service) UpdateExternalUserActionIfUnchanged(
	ctx context.Context,
	actorID, storyID, workspaceID uuid.UUID,
	expectedUpdatedAt time.Time,
	updates map[string]any,
) error {
	if expectedUpdatedAt.IsZero() {
		return errors.New("expected story update time is required")
	}
	expectedUpdatedAt = expectedUpdatedAt.UTC()
	return s.updateWithOptions(ctx, storyID, workspaceID, actorID, updates, updateOptions{
		publishEvents:            true,
		enqueueGitHubSync:        true,
		recordDescriptionUpdates: true,
		expectedUpdatedAt:        &expectedUpdatedAt,
	})
}

func (s *Service) RecordActivity(ctx context.Context, activity CoreActivity) error {
	if _, err := s.repo.RecordActivities(ctx, []CoreActivity{activity}); err != nil {
		return err
	}
	return nil
}

func (s *Service) updateWithOptions(ctx context.Context, storyID, workspaceID, actorID uuid.UUID, updates map[string]any, options updateOptions) error {
	s.log.Info(ctx, "business.core.stories.Update")
	ctx, span := web.AddSpan(ctx, "business.core.stories.Update")
	defer span.End()

	// Copy updates to avoid mutating shared maps in bulk updates.
	normalizedUpdates := make(map[string]any, len(updates))
	for field, value := range updates {
		normalizedUpdates[field] = value
	}
	updates = normalizedUpdates

	story, err := s.repo.Get(ctx, storyID, workspaceID)
	if err != nil {
		span.RecordError(err)
		s.log.Error(ctx, "failed to get story", "error", err)
		return err
	}
	activityDisplayValues := make(map[string]string)

	keyResultValue, hasKeyResultUpdate := updates["key_result_id"]
	if hasKeyResultUpdate {
		keyResultID, validKeyResultUpdate := optionalUUIDUpdate(keyResultValue)
		if validKeyResultUpdate && keyResultID != nil {
			keyResult, err := s.repo.ResolveKeyResult(ctx, *keyResultID, workspaceID)
			if err != nil {
				span.RecordError(err)
				return fmt.Errorf("resolve key result objective: %w", err)
			}
			activityDisplayValues["key_result_id"] = keyResult.Name

			if objectiveValue, hasObjectiveUpdate := updates["objective_id"]; hasObjectiveUpdate {
				requestedObjectiveID, validObjectiveUpdate := optionalUUIDUpdate(objectiveValue)
				if !validObjectiveUpdate || requestedObjectiveID == nil || *requestedObjectiveID != keyResult.ObjectiveID {
					return fmt.Errorf(
						"%w: key result %s belongs to objective %s",
						ErrObjectiveKeyResultMismatch,
						*keyResultID,
						keyResult.ObjectiveID,
					)
				}
			}

			updates["objective_id"] = &keyResult.ObjectiveID
		}
	} else if objectiveValue, hasObjectiveUpdate := updates["objective_id"]; hasObjectiveUpdate {
		if objectiveID, validObjectiveUpdate := optionalUUIDUpdate(objectiveValue); validObjectiveUpdate && !sameOptionalUUID(objectiveID, story.Objective) {
			updates["key_result_id"] = nil
		}
	}

	if err := s.applyEstimateUpdate(ctx, workspaceID, story, updates); err != nil {
		span.RecordError(err)
		return err
	}
	if err := applyStoryTimeContractUpdate(story, updates); err != nil {
		span.RecordError(err)
		return err
	}

	for field, value := range updates {
		if s.valuesEqual(s.getOldValue(story, field), value) {
			delete(updates, field)
		}
	}
	if len(updates) == 0 && !options.reconcileMedia {
		return nil
	}
	if len(updates) == 0 {
		orphanedMediaIDs, err := s.repo.UpdateWithMediaReconciliation(
			ctx,
			storyID,
			workspaceID,
			updates,
			options.referencedMediaIDs,
		)
		if err != nil {
			span.RecordError(err)
			return err
		}
		if options.orphanedMediaIDs != nil {
			*options.orphanedMediaIDs = append((*options.orphanedMediaIDs)[:0], orphanedMediaIDs...)
		}
		return nil
	}

	if assigneeID, ok := mayaAssignmentUpdateAssignee(updates); ok {
		updatedStory, err := storyWithAssignee(story, assigneeID)
		if err != nil {
			s.log.Error(ctx, "failed to prepare story for Maya assignment automation", "story_id", storyID, "workspace_id", workspaceID, "error", err)
			return err
		}
		if err := s.validateMayaAssignment(ctx, updatedStory, story.Assignee, actorID); err != nil {
			span.RecordError(err)
			return err
		}
	}

	// Handle auto-completion logic if status is being updated
	if newStatusID, hasStatusUpdate := updates["status_id"]; hasStatusUpdate {
		if err := s.handleCompletionStatusChange(ctx, story, newStatusID, updates); err != nil {
			s.log.Error(ctx, "failed to handle completion status change", "error", err)
			// Don't fail the entire update - log and continue
		}
	}

	// Update the story, reconciling inline media only for an explicitly
	// authoritative editor snapshot.
	var updateErr error
	if options.reconcileMedia {
		var orphanedMediaIDs []uuid.UUID
		orphanedMediaIDs, updateErr = s.repo.UpdateWithMediaReconciliation(
			ctx,
			storyID,
			workspaceID,
			updates,
			options.referencedMediaIDs,
		)
		if updateErr == nil && options.orphanedMediaIDs != nil {
			*options.orphanedMediaIDs = append((*options.orphanedMediaIDs)[:0], orphanedMediaIDs...)
		}
	} else {
		if options.expectedUpdatedAt == nil {
			updateErr = s.repo.Update(ctx, storyID, workspaceID, updates)
		} else {
			conditionalRepo, ok := s.repo.(conditionalUpdateRepository)
			if !ok {
				return errors.New("story repository does not support conditional updates")
			}
			var updated bool
			updated, updateErr = conditionalRepo.UpdateIfUnchanged(
				ctx,
				storyID,
				workspaceID,
				*options.expectedUpdatedAt,
				updates,
			)
			if updateErr == nil && !updated {
				updateErr = ErrStoryChanged
			}
		}
	}
	if updateErr != nil {
		span.RecordError(updateErr)
		return updateErr
	}
	ca := []CoreActivity{}
	activityReason := normalizeActivityReason(options.activityReason)

	for field, value := range updates {
		if strings.Contains(field, "description") && !options.recordDescriptionUpdates {
			continue
		}

		currentValue := s.formatValue(value)
		if displayValue, ok := activityDisplayValues[field]; ok {
			currentValue = displayValue
		}
		na := CoreActivity{
			StoryID:      storyID,
			Type:         "update",
			Field:        field,
			CurrentValue: currentValue,
			OldValue:     s.getOldValue(story, field),
			NewValue:     value,
			Reason:       activityReason,
			UserID:       actorID,
			WorkspaceID:  workspaceID,
		}
		ca = append(ca, na)
	}
	if len(ca) > 0 {
		if _, err := s.repo.RecordActivities(ctx, ca); err != nil {
			span.RecordError(err)
		}
	}

	span.AddEvent("story updated", trace.WithAttributes(
		attribute.String("story.id", storyID.String()),
	))

	_, statusChanged := updates["status_id"]
	if (options.publishEvents || (options.publishStatusEvents && statusChanged)) && s.publisher != nil {
		audienceIDs, audienceErr := s.repo.GetNotificationAudience(ctx, storyID, workspaceID)
		audienceResolved := audienceErr == nil
		if audienceErr != nil {
			s.log.Error(ctx, "failed to load story notification audience", "error", audienceErr, "story_id", storyID)
		}
		payload := events.StoryUpdatedPayload{
			StoryID:          storyID,
			WorkspaceID:      workspaceID,
			Updates:          updates,
			AssigneeID:       story.Assignee, // Current assignee before update
			AudienceIDs:      audienceIDs,
			AudienceResolved: audienceResolved,
		}
		if statusChanged {
			payload.PreviousStatusID = story.Status
		}

		event := events.Event{
			Type:      events.StoryUpdated,
			Payload:   payload,
			Timestamp: time.Now(),
			ActorID:   actorID,
		}

		if err := s.publisher.Publish(context.Background(), event); err != nil {
			s.log.Error(ctx, "failed to publish story updated event", "error", err)
			// Don't return error as this is not critical
		}
	}
	if options.enqueueGitHubSync {
		s.enqueueGitHubStorySync(ctx, storyID, workspaceID)
	}

	return nil
}

func normalizeActivityReason(reason string) *string {
	const maxActivityReasonRunes = 180

	normalized := strings.Join(strings.Fields(reason), " ")
	if normalized == "" {
		return nil
	}

	runes := []rune(normalized)
	if len(runes) > maxActivityReasonRunes {
		normalized = strings.TrimSpace(string(runes[:maxActivityReasonRunes-3])) + "..."
	}
	return &normalized
}

func (s *Service) enqueueGitHubStorySync(ctx context.Context, storyID, workspaceID uuid.UUID) {
	if s.tasksService == nil {
		return
	}
	if _, err := s.tasksService.EnqueueGitHubStorySync(tasks.GitHubStorySyncPayload{
		StoryID:     storyID,
		WorkspaceID: workspaceID,
	}); err != nil {
		s.log.Error(ctx, "failed to enqueue github story sync task", "story_id", storyID, "workspace_id", workspaceID, "error", err)
	}
}

// handleCompletionStatusChange handles auto-setting completed_at based on status category changes
func (s *Service) handleCompletionStatusChange(ctx context.Context, story CoreSingleStory,
	newStatusID any, updates map[string]any) error {

	// Convert status ID to string
	newStatus, ok := newStatusID.(uuid.UUID)
	if !ok {
		return fmt.Errorf("status ID is not a string: %T", newStatusID)
	}

	// Get old status category
	oldCategory, err := s.repo.GetStatusCategory(ctx, story.Status.String())
	if err != nil {
		s.log.Error(ctx, "failed to get old status category", "statusId", *story.Status, "error", err)
		// Continue without old category info
	}

	// Get new status category
	newCategory, err := s.repo.GetStatusCategory(ctx, newStatus.String())
	if err != nil {
		return fmt.Errorf("failed to get new status category: %w", err)
	}

	// Auto-completion logic
	now := time.Now()
	if newCategory == "completed" && oldCategory != "completed" {
		updates["completed_at"] = &now
		s.log.Info(ctx, "auto-completing story", "storyId", story.ID, "oldCategory", oldCategory, "newCategory", newCategory)
	} else if oldCategory == "completed" && newCategory != "completed" {
		updates["completed_at"] = nil
		s.log.Info(ctx, "auto-uncompleting story", "storyId", story.ID, "oldCategory", oldCategory, "newCategory", newCategory)
	}
	// If both old and new are in completed category, do nothing (keep existing completed_at)

	return nil
}

// BulkUpdate updates multiple stories with the same updates in parallel.
func (s *Service) BulkUpdate(ctx context.Context, storyIDs []uuid.UUID, workspaceID uuid.UUID, updates map[string]any) error {
	s.log.Info(ctx, "business.core.stories.BulkUpdate")
	ctx, span := web.AddSpan(ctx, "business.core.stories.BulkUpdate")
	defer span.End()

	if len(storyIDs) == 0 {
		return fmt.Errorf("no story IDs provided")
	}

	span.AddEvent("bulk update started", trace.WithAttributes(
		attribute.Int("story.count", len(storyIDs)),
		attribute.String("workspace.id", workspaceID.String()),
	))
	var wg sync.WaitGroup

	// Channel to collect errors from goroutines
	errChan := make(chan error, len(storyIDs))
	for _, storyID := range storyIDs {
		wg.Add(1)
		go func(id uuid.UUID) {
			defer wg.Done()
			if err := s.Update(ctx, id, workspaceID, updates); err != nil {
				errChan <- fmt.Errorf("failed to update story %s: %w", id, err)
			}
		}(storyID)
	}

	// Wait for all goroutines to complete
	wg.Wait()
	close(errChan)

	// Collect all errors
	var errors []error
	for err := range errChan {
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		span.RecordError(fmt.Errorf("bulk update completed with %d errors", len(errors)))

		var errorMessages []string
		for _, err := range errors {
			errorMessages = append(errorMessages, err.Error())
		}
		return fmt.Errorf("bulk update errors: %s", strings.Join(errorMessages, "; "))
	}

	span.AddEvent("bulk update completed successfully", trace.WithAttributes(
		attribute.Int("stories.updated", len(storyIDs)),
	))

	return nil
}

// BulkDelete deletes the stories with the specified IDs.
func (s *Service) BulkDelete(ctx context.Context, ids []uuid.UUID, workspaceId uuid.UUID) error {
	s.log.Info(ctx, "business.core.stories.BulkDelete")
	ctx, span := web.AddSpan(ctx, "business.core.stories.BulkDelete")
	defer span.End()

	if err := s.repo.BulkDelete(ctx, ids, workspaceId); err != nil {
		span.RecordError(err)
		return err
	}
	return nil
}

// HardBulkDelete performs permanent removal of the stories with the specified IDs.
func (s *Service) HardBulkDelete(ctx context.Context, ids []uuid.UUID, workspaceId uuid.UUID) ([]uuid.UUID, error) {
	s.log.Info(ctx, fmt.Sprintf("Hard bulk deleting stories: %v", ids), "story_ids", ids)
	ctx, span := web.AddSpan(ctx, "business.core.stories.HardBulkDelete")
	defer span.End()

	orphanedAttachmentIDs, err := s.repo.HardBulkDelete(ctx, ids, workspaceId)
	if err != nil {
		s.log.Error(ctx, fmt.Sprintf("Failed to hard bulk delete stories: %s", err),
			"story_ids", ids, "error", err)
		span.RecordError(err)
		return nil, err
	}

	s.log.Info(ctx, fmt.Sprintf("Successfully hard bulk deleted stories: %v", ids),
		"story_ids", ids)
	span.AddEvent("Stories hard bulk deleted.", trace.WithAttributes(
		attribute.Int("stories.count", len(ids))))

	return orphanedAttachmentIDs, nil
}

// Restore restores the story with the specified ID.
func (s *Service) Restore(ctx context.Context, id uuid.UUID, workspaceId uuid.UUID) error {
	s.log.Info(ctx, "business.core.stories.Restore")
	ctx, span := web.AddSpan(ctx, "business.core.stories.Restore")
	defer span.End()

	if err := s.repo.Restore(ctx, id, workspaceId); err != nil {
		span.RecordError(err)
		return err
	}
	return nil
}

// BulkRestore restores the stories with the specified IDs.
func (s *Service) BulkRestore(ctx context.Context, ids []uuid.UUID, workspaceId uuid.UUID) error {
	s.log.Info(ctx, "business.core.stories.BulkRestore")
	ctx, span := web.AddSpan(ctx, "business.core.stories.BulkRestore")
	defer span.End()

	if err := s.repo.BulkRestore(ctx, ids, workspaceId); err != nil {
		span.RecordError(err)
		return err
	}
	return nil
}

// BulkArchive archives the stories with the specified IDs.
func (s *Service) BulkArchive(ctx context.Context, ids []uuid.UUID, workspaceId uuid.UUID) error {
	s.log.Info(ctx, fmt.Sprintf("Bulk archiving stories: %v", ids), "story_ids", ids)
	ctx, span := web.AddSpan(ctx, "business.core.stories.BulkArchive")
	defer span.End()

	if err := s.repo.BulkArchive(ctx, ids, workspaceId); err != nil {
		s.log.Error(ctx, fmt.Sprintf("Failed to bulk archive stories: %s", err),
			"story_ids", ids, "error", err)
		span.RecordError(err)
		return err
	}

	s.log.Info(ctx, fmt.Sprintf("Successfully bulk archived stories: %v", ids),
		"story_ids", ids)
	span.AddEvent("Stories bulk archived.", trace.WithAttributes(
		attribute.Int("stories.count", len(ids))))

	return nil
}

// BulkUnarchive unarchives the stories with the specified IDs.
func (s *Service) BulkUnarchive(ctx context.Context, ids []uuid.UUID, workspaceId uuid.UUID) error {
	s.log.Info(ctx, fmt.Sprintf("Bulk unarchiving stories: %v", ids), "story_ids", ids)
	ctx, span := web.AddSpan(ctx, "business.core.stories.BulkUnarchive")
	defer span.End()

	if err := s.repo.BulkUnarchive(ctx, ids, workspaceId); err != nil {
		s.log.Error(ctx, fmt.Sprintf("Failed to bulk unarchive stories: %s", err),
			"story_ids", ids, "error", err)
		span.RecordError(err)
		return err
	}

	s.log.Info(ctx, fmt.Sprintf("Successfully bulk unarchived stories: %v", ids),
		"story_ids", ids)
	span.AddEvent("Stories bulk unarchived.", trace.WithAttributes(
		attribute.Int("stories.count", len(ids))))

	return nil
}

// GetActivitiesWithUser returns the activities for a story with user details and pagination.
func (s *Service) GetActivitiesWithUser(ctx context.Context, storyID uuid.UUID, page, pageSize int) ([]CoreActivityWithUser, bool, error) {
	s.log.Info(ctx, "business.core.activities.GetActivitiesWithUser")
	ctx, span := web.AddSpan(ctx, "business.core.activities.GetActivitiesWithUser")
	defer span.End()

	activities, hasMore, err := s.repo.GetActivitiesWithUser(ctx, storyID, page, pageSize)
	if err != nil {
		span.RecordError(err)
		return nil, false, err
	}

	span.AddEvent("activities with user details retrieved.", trace.WithAttributes(
		attribute.Int("activity.count", len(activities)),
		attribute.Int("page", page),
		attribute.Int("pageSize", pageSize),
		attribute.Bool("has.more", hasMore),
	))

	return activities, hasMore, nil
}

// CreateComment creates a comment for a story.
func (s *Service) CreateComment(ctx context.Context, workspaceID uuid.UUID, cnc CoreNewComment) (comments.CoreComment, error) {
	actorID, _ := auth.GetUserID(ctx)
	return s.createCommentWithOptions(ctx, workspaceID, cnc, commentOptions{actorID: actorID})
}

func (s *Service) CreateCommentExternal(ctx context.Context, actorID uuid.UUID, workspaceID uuid.UUID, cnc CoreNewComment) (comments.CoreComment, error) {
	return s.createCommentWithOptions(ctx, workspaceID, cnc, commentOptions{actorID: actorID})
}

func (s *Service) createCommentWithOptions(ctx context.Context, workspaceID uuid.UUID, cnc CoreNewComment, options commentOptions) (comments.CoreComment, error) {
	s.log.Info(ctx, "business.core.stories.CreateComment")
	ctx, span := web.AddSpan(ctx, "business.core.stories.CreateComment")
	defer span.End()

	// Now get story details for notifications
	story, err := s.repo.Get(ctx, cnc.StoryID, workspaceID)
	if err != nil {
		span.RecordError(err)
		return comments.CoreComment{}, err
	}

	comment, err := s.repo.CreateComment(ctx, cnc)
	if err != nil {
		span.RecordError(err)
		return comments.CoreComment{}, err
	}

	if len(cnc.Mentions) > 0 {
		if err := s.mentionsRepo.SaveMentions(ctx, comment.ID, cnc.Mentions); err != nil {
			s.log.Error(ctx, "failed to save mentions", "error", err, "commentId", comment.ID)
			// Note: We don't return error here to avoid failing comment creation if mentions fail
		}
	}

	actorID := options.actorID
	audienceIDs, audienceErr := s.repo.GetNotificationAudience(ctx, cnc.StoryID, workspaceID)
	audienceResolved := audienceErr == nil
	if audienceErr != nil {
		s.log.Error(ctx, "failed to load story notification audience", "error", audienceErr, "story_id", cnc.StoryID)
	}

	// Publish events based on comment type
	if cnc.Parent != nil {
		// This is a reply - get parent comment details
		parentComment, err := s.repo.GetComment(ctx, *cnc.Parent)
		if err != nil {
			s.log.Error(ctx, "failed to get parent comment for notification", "error", err, "parent_id", *cnc.Parent)
		} else {
			// Publish comment reply event
			payload := events.CommentRepliedPayload{
				CommentID:        comment.ID,
				ParentCommentID:  *cnc.Parent,
				ParentAuthorID:   parentComment.UserID,
				StoryID:          cnc.StoryID,
				StoryTitle:       story.Title,
				WorkspaceID:      story.Workspace,
				Content:          cnc.Comment,
				Mentions:         cnc.Mentions,
				AudienceIDs:      audienceIDs,
				AudienceResolved: audienceResolved,
			}

			event := events.Event{
				Type:      events.CommentReplied,
				Payload:   payload,
				Timestamp: time.Now(),
				ActorID:   actorID,
			}

			if err := s.publisher.Publish(context.Background(), event); err != nil {
				s.log.Error(ctx, "failed to publish comment replied event", "error", err)
			}
		}
	} else {
		// This is a new comment on the story
		payload := events.CommentCreatedPayload{
			CommentID:        comment.ID,
			StoryID:          cnc.StoryID,
			StoryTitle:       story.Title,
			AssigneeID:       story.Assignee,
			WorkspaceID:      story.Workspace,
			Content:          cnc.Comment,
			Mentions:         cnc.Mentions,
			AudienceIDs:      audienceIDs,
			AudienceResolved: audienceResolved,
		}

		event := events.Event{
			Type:      events.CommentCreated,
			Payload:   payload,
			Timestamp: time.Now(),
			ActorID:   actorID,
		}

		if err := s.publisher.Publish(context.Background(), event); err != nil {
			s.log.Error(ctx, "failed to publish comment created event", "error", err)
		}
	}

	// Publish mention events for each mentioned user
	for _, mentionedUserID := range cnc.Mentions {
		payload := events.UserMentionedPayload{
			CommentID:     comment.ID,
			StoryID:       cnc.StoryID,
			StoryTitle:    story.Title,
			WorkspaceID:   story.Workspace,
			MentionedUser: mentionedUserID,
			Content:       cnc.Comment,
		}

		event := events.Event{
			Type:      events.UserMentioned,
			Payload:   payload,
			Timestamp: time.Now(),
			ActorID:   actorID,
		}

		if err := s.publisher.Publish(context.Background(), event); err != nil {
			s.log.Error(ctx, "failed to publish user mentioned event", "error", err, "mentioned_user", mentionedUserID)
		}
	}

	span.AddEvent("comment created.", trace.WithAttributes(
		attribute.String("comment.comment", comment.Comment),
		attribute.Int("mentions.count", len(cnc.Mentions)),
	))

	return comment, nil
}

// GetComments returns the comments for a story with pagination.
func (s *Service) GetComments(ctx context.Context, storyID uuid.UUID, page, pageSize int) ([]comments.CoreComment, bool, error) {
	s.log.Info(ctx, "business.core.stories.GetComments")
	ctx, span := web.AddSpan(ctx, "business.core.stories.GetComments")
	defer span.End()

	comments, hasMore, err := s.repo.GetComments(ctx, storyID, page, pageSize)
	if err != nil {
		s.log.Error(ctx, fmt.Sprintf("failed to get comments: %s", err))
		span.RecordError(err)
		return nil, false, err
	}

	span.AddEvent("comments retrieved.", trace.WithAttributes(
		attribute.Int("comment.count", len(comments)),
		attribute.Int("page", page),
		attribute.Int("pageSize", pageSize),
		attribute.Bool("has.more", hasMore),
	))

	return comments, hasMore, nil
}

// DuplicateStory creates a copy of an existing story.
func (s *Service) DuplicateStory(ctx context.Context, originalStoryID uuid.UUID, workspaceId uuid.UUID, userID uuid.UUID) (CoreSingleStory, error) {
	s.log.Info(ctx, "business.core.stories.DuplicateStory")
	ctx, span := web.AddSpan(ctx, "business.core.stories.DuplicateStory")
	defer span.End()

	duplicatedStory, err := s.repo.DuplicateStory(ctx, originalStoryID, workspaceId, userID)
	if err != nil {
		span.RecordError(err)
		return CoreSingleStory{}, fmt.Errorf("failed to duplicate story: %w", err)
	}
	if err := s.enrichSingleStoryEstimate(ctx, workspaceId, &duplicatedStory); err != nil {
		span.RecordError(err)
		return CoreSingleStory{}, err
	}

	span.AddEvent("Story duplicated.", trace.WithAttributes(
		attribute.String("original_story.id", originalStoryID.String()),
		attribute.String("new_story.id", duplicatedStory.ID.String()),
	))

	return duplicatedStory, nil
}

// CountInWorkspace returns the count of stories in a workspace.
func (s *Service) CountInWorkspace(ctx context.Context, workspaceId uuid.UUID) (int, error) {
	ctx, span := web.AddSpan(ctx, "business.services.stories.CountInWorkspace")
	defer span.End()

	count, err := s.repo.CountStoriesInWorkspace(ctx, workspaceId)
	if err != nil {
		return 0, fmt.Errorf("counting stories in workspace: %w", err)
	}

	return count, nil
}

// ListGroupedStories returns stories grouped by the specified field with limited stories per group
func (s *Service) ListGroupedStories(ctx context.Context, query CoreStoryQuery) ([]CoreStoryGroup, error) {
	ctx, span := web.AddSpan(ctx, "business.services.stories.ListGroupedStories")
	defer span.End()

	groups, err := s.repo.ListGroupedStories(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("listing grouped stories: %w", err)
	}
	for i := range groups {
		if err := s.enrichStoryListEstimates(ctx, query.Filters.WorkspaceID, groups[i].Stories); err != nil {
			return nil, err
		}
	}

	return groups, nil
}

// ListGroupStories returns more stories for a specific group (for load more functionality)
func (s *Service) ListGroupStories(ctx context.Context, groupKey string, query CoreStoryQuery) ([]CoreStoryList, bool, error) {
	ctx, span := web.AddSpan(ctx, "business.services.stories.ListGroupStories")
	defer span.End()

	stories, hasMore, err := s.repo.ListGroupStories(ctx, groupKey, query)
	if err != nil {
		return nil, false, fmt.Errorf("listing group stories: %w", err)
	}
	if err := s.enrichStoryListEstimates(ctx, query.Filters.WorkspaceID, stories); err != nil {
		return nil, false, err
	}

	return stories, hasMore, nil
}

// ListByCategory returns stories filtered by category with pagination
func (s *Service) ListByCategory(ctx context.Context, workspaceId, userID, teamId uuid.UUID, category string, page, pageSize int, showSubStories bool) ([]CoreStoryList, bool, error) {
	ctx, span := web.AddSpan(ctx, "business.services.stories.ListByCategory")
	defer span.End()

	stories, hasMore, err := s.repo.ListByCategory(ctx, workspaceId, userID, teamId, category, page, pageSize, showSubStories)
	if err != nil {
		return nil, false, fmt.Errorf("listing stories by category: %w", err)
	}
	if err := s.enrichStoryListEstimates(ctx, workspaceId, stories); err != nil {
		return nil, false, err
	}

	span.AddEvent("category stories retrieved.", trace.WithAttributes(
		attribute.Int("stories.count", len(stories)),
		attribute.String("category", category),
		attribute.Int("page", page),
		attribute.Int("pageSize", pageSize),
		attribute.Bool("has.more", hasMore),
	))

	return stories, hasMore, nil
}

func (s *Service) formatValue(value any) string {
	if value == nil {
		return "nil"
	}
	switch v := value.(type) {
	case string:
		return v
	case *string:
		if v != nil {
			return *v
		}
		return "nil"
	case *int16:
		if v != nil {
			return fmt.Sprintf("%d", *v)
		}
		return "nil"
	case *int:
		if v != nil {
			return strconv.Itoa(*v)
		}
		return "nil"
	case int:
		return strconv.Itoa(v)
	case int16:
		return fmt.Sprintf("%d", v)
	case *float64:
		if v != nil {
			return fmt.Sprintf("%.2f", *v)
		}
		return "nil"
	case *uuid.UUID:
		if v != nil {
			return v.String()
		}
		return "nil"
	case *time.Time:
		if v != nil {
			return v.Format(time.RFC3339)
		}
		return "nil"
	case time.Time:
		return v.Format(time.RFC3339)
	default:
		return fmt.Sprintf("%v", v)
	}
}

func uniqueUUIDs(values []uuid.UUID) []uuid.UUID {
	seen := make(map[uuid.UUID]struct{}, len(values))
	result := make([]uuid.UUID, 0, len(values))
	for _, value := range values {
		if value == uuid.Nil {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}

func sameUUIDSet(left, right []uuid.UUID) bool {
	if len(left) != len(right) {
		return false
	}
	values := make(map[uuid.UUID]struct{}, len(left))
	for _, value := range left {
		values[value] = struct{}{}
	}
	for _, value := range right {
		if _, exists := values[value]; !exists {
			return false
		}
	}
	return true
}

func (s *Service) valuesEqual(oldValue, newValue any) bool {
	return normalizeComparableValue(oldValue) == normalizeComparableValue(newValue)
}

func normalizeComparableValue(value any) string {
	if value == nil {
		return "nil"
	}

	switch v := value.(type) {
	case string:
		return v
	case *string:
		if v == nil {
			return "nil"
		}
		return *v
	case uuid.UUID:
		return v.String()
	case *uuid.UUID:
		if v == nil {
			return "nil"
		}
		return v.String()
	case time.Time:
		return v.UTC().Format(time.RFC3339Nano)
	case *time.Time:
		if v == nil {
			return "nil"
		}
		return v.UTC().Format(time.RFC3339Nano)
	case int:
		return strconv.Itoa(v)
	case *int:
		if v == nil {
			return "nil"
		}
		return strconv.Itoa(*v)
	case int16:
		return strconv.FormatInt(int64(v), 10)
	case *int16:
		if v == nil {
			return "nil"
		}
		return strconv.FormatInt(int64(*v), 10)
	case bool:
		return strconv.FormatBool(v)
	case *bool:
		if v == nil {
			return "nil"
		}
		return strconv.FormatBool(*v)
	default:
		return fmt.Sprintf("%v", v)
	}
}

// QueryByRef returns a story by team code and sequence ID.
func (s *Service) QueryByRef(ctx context.Context, workspaceId uuid.UUID, storyRef string) (CoreSingleStory, error) {
	s.log.Info(ctx, "business.core.stories.QueryByRef")
	ctx, span := web.AddSpan(ctx, "business.core.stories.QueryByRef")
	defer span.End()

	teamCode, sequenceID, err := s.parseStoryRef(storyRef)
	if err != nil {
		span.RecordError(err)
		return CoreSingleStory{}, err
	}

	story, err := s.repo.QueryByRef(ctx, workspaceId, teamCode, sequenceID)
	if err != nil {
		span.RecordError(err)
		return CoreSingleStory{}, err
	}
	if err := s.enrichSingleStoryEstimate(ctx, workspaceId, &story); err != nil {
		span.RecordError(err)
		return CoreSingleStory{}, err
	}

	return story, nil
}

// parseStoryRef parses a story reference into team code and sequence ID.
func (s *Service) parseStoryRef(storyRef string) (string, int, error) {
	storyRef = strings.ToUpper(strings.ReplaceAll(storyRef, " ", ""))
	storyRef = strings.ReplaceAll(storyRef, "-", "")

	// Split at the transition from letter to digit
	var teamCode, seqStr string
	for i, ch := range storyRef {
		if ch >= '0' && ch <= '9' {
			teamCode = storyRef[:i]
			seqStr = storyRef[i:]
			break
		}
	}

	if teamCode == "" || seqStr == "" {
		return "", 0, fmt.Errorf("%w: %s", ErrInvalidStoryReference, storyRef)
	}

	seqID, err := strconv.Atoi(seqStr)
	if err != nil {
		return "", 0, fmt.Errorf("%w: invalid sequence number in %s", ErrInvalidStoryReference, storyRef)
	}

	return teamCode, seqID, nil
}

func (s *Service) enrichSingleStoryEstimate(ctx context.Context, workspaceID uuid.UUID, story *CoreSingleStory) error {
	schemeCache := map[uuid.UUID]string{}
	scheme, err := s.getEstimateSchemeForTeam(ctx, workspaceID, story.Team, schemeCache)
	if err != nil {
		return err
	}

	story.EstimateScheme = scheme
	story.EstimateLabel = EstimateLabelFromValue(scheme, story.EstimateValue)

	for i := range story.SubStories {
		if err := s.enrichStoryListItemEstimate(ctx, workspaceID, &story.SubStories[i], schemeCache); err != nil {
			return err
		}
	}

	for i := range story.Associations {
		if err := s.enrichStoryListItemEstimate(ctx, workspaceID, &story.Associations[i].Story, schemeCache); err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) enrichStoryListEstimates(ctx context.Context, workspaceID uuid.UUID, stories []CoreStoryList) error {
	schemeCache := map[uuid.UUID]string{}
	for i := range stories {
		if err := s.enrichStoryListItemEstimate(ctx, workspaceID, &stories[i], schemeCache); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) enrichStoryListItemEstimate(ctx context.Context, workspaceID uuid.UUID, story *CoreStoryList, schemeCache map[uuid.UUID]string) error {
	scheme, err := s.getEstimateSchemeForTeam(ctx, workspaceID, story.Team, schemeCache)
	if err != nil {
		return err
	}

	story.EstimateScheme = scheme
	story.EstimateLabel = EstimateLabelFromValue(scheme, story.EstimateValue)

	for i := range story.SubStories {
		if err := s.enrichStoryListItemEstimate(ctx, workspaceID, &story.SubStories[i], schemeCache); err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) getEstimateSchemeForTeam(ctx context.Context, workspaceID, teamID uuid.UUID, schemeCache map[uuid.UUID]string) (string, error) {
	if scheme, ok := schemeCache[teamID]; ok {
		return scheme, nil
	}

	scheme, err := s.repo.GetTeamEstimateScheme(ctx, teamID, workspaceID)
	if err != nil {
		return "", err
	}
	schemeCache[teamID] = scheme
	return scheme, nil
}

// AddAssociation adds an association between two stories.
func (s *Service) AddAssociation(ctx context.Context, fromID, toID uuid.UUID, associationType string, workspaceID uuid.UUID) (CoreStoryAssociation, error) {
	s.log.Info(ctx, "business.core.stories.AddAssociation")
	ctx, span := web.AddSpan(ctx, "business.core.stories.AddAssociation")
	defer span.End()

	// Validate inputs
	if fromID == toID {
		return CoreStoryAssociation{}, fmt.Errorf("cannot associate story with itself")
	}

	assoc, err := s.repo.AddAssociation(ctx, fromID, toID, associationType, workspaceID)
	if err != nil {
		span.RecordError(err)
		return CoreStoryAssociation{}, err
	}
	if err := s.recordAssociationActivities(ctx, assoc, workspaceID, associationActivityAdded); err != nil {
		span.RecordError(err)
	}

	return assoc, nil
}

// UpdateAssociation updates an association between two stories.
func (s *Service) UpdateAssociation(ctx context.Context, associationID, fromID, toID uuid.UUID, associationType string, workspaceID uuid.UUID) (CoreStoryAssociation, error) {
	s.log.Info(ctx, "business.core.stories.UpdateAssociation")
	ctx, span := web.AddSpan(ctx, "business.core.stories.UpdateAssociation")
	defer span.End()

	if fromID == toID {
		return CoreStoryAssociation{}, fmt.Errorf("cannot associate story with itself")
	}

	assoc, err := s.repo.UpdateAssociation(ctx, associationID, fromID, toID, associationType, workspaceID)
	if err != nil {
		return CoreStoryAssociation{}, err
	}
	if err := s.recordAssociationActivities(ctx, assoc, workspaceID, associationActivityUpdated); err != nil {
		span.RecordError(err)
	}

	return assoc, nil
}

// RemoveAssociation removes an association between two stories.
func (s *Service) RemoveAssociation(ctx context.Context, associationID, workspaceID uuid.UUID) error {
	s.log.Info(ctx, "business.core.stories.RemoveAssociation")
	ctx, span := web.AddSpan(ctx, "business.core.stories.RemoveAssociation")
	defer span.End()

	assoc, err := s.repo.RemoveAssociation(ctx, associationID, workspaceID)
	if err != nil {
		span.RecordError(err)
		return err
	}
	if err := s.recordAssociationActivities(ctx, assoc, workspaceID, associationActivityRemoved); err != nil {
		span.RecordError(err)
	}

	return nil
}

func (s *Service) formatLabelActivityValue(labels []uuid.UUID) string {
	if len(labels) == 1 {
		return "1 label"
	}
	return fmt.Sprintf("%d labels", len(labels))
}

const (
	associationActivityAdded   = "association_added"
	associationActivityUpdated = "association_updated"
	associationActivityRemoved = "association_removed"
)

func (s *Service) recordAssociationActivities(ctx context.Context, assoc CoreStoryAssociation, workspaceID uuid.UUID, reason string) error {
	actorID, _ := auth.GetUserID(ctx)
	activityReason := reason
	outgoingOldValue, incomingOldValue := associationOldValues(assoc)
	activities := []CoreActivity{
		{
			StoryID:      assoc.FromStoryID,
			Type:         "update",
			Field:        outgoingAssociationActivityField(assoc.Type),
			CurrentValue: s.associationActivityValue(assoc.ToStoryID, assoc),
			OldValue:     outgoingOldValue,
			NewValue:     assoc.ToStoryID,
			Reason:       &activityReason,
			UserID:       actorID,
			WorkspaceID:  workspaceID,
		},
		{
			StoryID:      assoc.ToStoryID,
			Type:         "update",
			Field:        incomingAssociationActivityField(assoc.Type),
			CurrentValue: s.associationActivityValue(assoc.FromStoryID, assoc),
			OldValue:     incomingOldValue,
			NewValue:     assoc.FromStoryID,
			Reason:       &activityReason,
			UserID:       actorID,
			WorkspaceID:  workspaceID,
		},
	}
	_, err := s.repo.RecordActivities(ctx, activities)
	return err
}

func associationOldValues(assoc CoreStoryAssociation) (any, any) {
	if assoc.PreviousType == nil || *assoc.PreviousType == assoc.Type {
		return nil, nil
	}
	return outgoingAssociationActivityLabel(*assoc.PreviousType), incomingAssociationActivityLabel(*assoc.PreviousType)
}

func (s *Service) associationActivityValue(storyID uuid.UUID, assoc CoreStoryAssociation) string {
	if storyID == assoc.FromStoryID && assoc.FromStoryTitle != "" {
		return assoc.FromStoryTitle
	}
	if storyID == assoc.ToStoryID && assoc.ToStoryTitle != "" {
		return assoc.ToStoryTitle
	}
	if assoc.Story.ID == storyID && assoc.Story.Title != "" {
		return assoc.Story.Title
	}
	return storyID.String()
}

func outgoingAssociationActivityField(associationType string) string {
	switch associationType {
	case "blocking":
		return "blocking_id"
	case "duplicate":
		return "duplicate_id"
	default:
		return "related_id"
	}
}

func incomingAssociationActivityField(associationType string) string {
	switch associationType {
	case "blocking":
		return "blocked_by_id"
	case "duplicate":
		return "duplicated_by_id"
	default:
		return "related_id"
	}
}

func outgoingAssociationActivityLabel(associationType string) string {
	switch associationType {
	case "blocking":
		return "Blocks"
	case "duplicate":
		return "Duplicate of"
	default:
		return "Related to"
	}
}

func incomingAssociationActivityLabel(associationType string) string {
	switch associationType {
	case "blocking":
		return "Blocked by"
	case "duplicate":
		return "Duplicated by"
	default:
		return "Related to"
	}
}

func (s *Service) getOldValue(story CoreSingleStory, field string) any {
	switch field {
	case "title":
		return story.Title
	case "description":
		return story.Description
	case "description_html":
		return story.DescriptionHTML
	case "parent_id":
		return story.Parent
	case "objective_id":
		return story.Objective
	case "status_id":
		return story.Status
	case "assignee_id":
		return story.Assignee
	case "priority":
		return story.Priority
	case "sprint_id":
		return story.Sprint
	case "key_result_id":
		return story.KeyResult
	case "start_date":
		return story.StartDate
	case "end_date":
		return story.EndDate
	case "completed_at":
		return story.CompletedAt
	case "estimate_unit":
		return story.EstimateValue
	case "estimated_duration_minutes":
		return story.EstimatedDurationMinutes
	case "minimum_focus_block_minutes":
		return story.MinimumFocusBlockMinutes
	default:
		return nil
	}
}

func (s *Service) applyEstimateUpdate(ctx context.Context, workspaceID uuid.UUID, story CoreSingleStory, updates map[string]any) error {
	estimateRaw, hasEstimateUpdate := updates["estimate_unit"]
	if !hasEstimateUpdate {
		return nil
	}

	estimateScheme, err := s.repo.GetTeamEstimateScheme(ctx, story.Team, workspaceID)
	if err != nil {
		return err
	}

	var estimateValue *int16
	switch value := estimateRaw.(type) {
	case nil:
		estimateValue = nil
	case *int16:
		estimateValue = value
	case int16:
		estimateValue = &value
	case int:
		normalized := int16(value)
		estimateValue = &normalized
	case float64:
		normalized := int16(value)
		if float64(normalized) != value {
			return fmt.Errorf("invalid estimate value type: %T", estimateRaw)
		}
		estimateValue = &normalized
	default:
		return fmt.Errorf("invalid estimate value type: %T", estimateRaw)
	}

	if err := ValidateEstimateValue(estimateScheme, estimateValue); err != nil {
		if estimateValue != nil {
			return fmt.Errorf("%w. If this work is larger than the max estimate, split it into smaller stories", err)
		}
		return err
	}

	updates["estimate_unit"] = estimateValue
	return nil
}
