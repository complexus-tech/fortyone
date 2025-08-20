package stories

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/complexus-tech/projects-api/internal/core/comments"
	"github.com/complexus-tech/projects-api/internal/core/links"
	"github.com/complexus-tech/projects-api/internal/web/mid"
	"github.com/complexus-tech/projects-api/pkg/events"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/publisher"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

var (
	ErrNotFound = errors.New("story not found")
)

// Repository provides access to the story storage.
type Repository interface {
	Get(ctx context.Context, id uuid.UUID, workspaceId uuid.UUID) (CoreSingleStory, error)
	Delete(ctx context.Context, id uuid.UUID, workspaceId uuid.UUID) error
	BulkDelete(ctx context.Context, ids []uuid.UUID, workspaceId uuid.UUID) error
	HardBulkDelete(ctx context.Context, ids []uuid.UUID, workspaceId uuid.UUID) error
	Restore(ctx context.Context, id uuid.UUID, workspaceId uuid.UUID) error
	BulkRestore(ctx context.Context, ids []uuid.UUID, workspaceId uuid.UUID) error
	BulkUnarchive(ctx context.Context, ids []uuid.UUID, workspaceId uuid.UUID) error
	Update(ctx context.Context, id uuid.UUID, workspaceId uuid.UUID, updates map[string]any) error
	UpdateLabels(ctx context.Context, id uuid.UUID, workspaceId uuid.UUID, labels []uuid.UUID) error
	GetStoryLinks(ctx context.Context, storyID uuid.UUID) ([]links.CoreLink, error)
	Create(ctx context.Context, story *CoreSingleStory) (CoreSingleStory, error)
	GetNextSequenceID(ctx context.Context, teamId uuid.UUID, workspaceId uuid.UUID) (int, func() error, func() error, error)
	MyStories(ctx context.Context, workspaceId uuid.UUID) ([]CoreStoryList, error)
	GetSubStories(ctx context.Context, parentId uuid.UUID, workspaceId uuid.UUID) ([]CoreStoryList, error)
	RecordActivities(ctx context.Context, activities []CoreActivity) ([]CoreActivity, error)
	GetActivities(ctx context.Context, storyID uuid.UUID, page, pageSize int) ([]CoreActivity, bool, error)
	CreateComment(ctx context.Context, comment CoreNewComment) (comments.CoreComment, error)
	GetComments(ctx context.Context, storyID uuid.UUID, page, pageSize int) ([]comments.CoreComment, bool, error)
	GetComment(ctx context.Context, commentID uuid.UUID) (comments.CoreComment, error)
	DuplicateStory(ctx context.Context, originalStoryID uuid.UUID, workspaceId uuid.UUID, userID uuid.UUID) (CoreSingleStory, error)
	CountStoriesInWorkspace(ctx context.Context, workspaceId uuid.UUID) (int, error)
	List(ctx context.Context, workspaceId uuid.UUID, filters map[string]any) ([]CoreStoryList, error)
	ListGroupedStories(ctx context.Context, query CoreStoryQuery) ([]CoreStoryGroup, error)
	ListGroupStories(ctx context.Context, groupKey string, query CoreStoryQuery) ([]CoreStoryList, bool, error)
	ListByCategory(ctx context.Context, workspaceId, userID, teamId uuid.UUID, category string, page, pageSize int) ([]CoreStoryList, bool, error)
	GetStatusCategory(ctx context.Context, statusID string) (string, error)
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
	repo         Repository
	mentionsRepo MentionsRepository
	log          *logger.Logger
	publisher    *publisher.Publisher
}

// New constructs a new stories service instance with the provided repository.
func New(log *logger.Logger, repo Repository, mentionsRepo MentionsRepository, publisher *publisher.Publisher) *Service {
	return &Service{
		repo:         repo,
		mentionsRepo: mentionsRepo,
		log:          log,
		publisher:    publisher,
	}
}

// Create creates a new story.
func (s *Service) Create(ctx context.Context, ns CoreNewStory, workspaceId uuid.UUID) (CoreSingleStory, error) {
	s.log.Info(ctx, "business.core.stories.create")
	ctx, span := web.AddSpan(ctx, "business.core.stories.Create")
	defer span.End()

	story := toCoreSingleStory(ns, workspaceId)

	cs, err := s.repo.Create(ctx, &story)
	if err != nil {
		span.RecordError(err)
		return CoreSingleStory{}, err
	}

	// Record in the activity log
	ca := CoreActivity{
		StoryID:      cs.ID,
		Type:         "create",
		Field:        "story",
		CurrentValue: cs.Title,
		UserID:       *ns.Reporter,
		WorkspaceID:  workspaceId,
	}
	if _, err := s.repo.RecordActivities(ctx, []CoreActivity{ca}); err != nil {
		span.RecordError(err)
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
	return nil
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

	return story, nil
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
	s.log.Info(ctx, "business.core.stories.Update")
	ctx, span := web.AddSpan(ctx, "business.core.stories.Update")
	defer span.End()

	// Get the actor ID from context
	actorID, _ := mid.GetUserID(ctx)

	story, err := s.repo.Get(ctx, storyID, workspaceID)
	if err != nil {
		span.RecordError(err)
		s.log.Error(ctx, "failed to get story", "error", err)
		return err
	}

	// Handle auto-completion logic if status is being updated
	if newStatusID, hasStatusUpdate := updates["status_id"]; hasStatusUpdate {
		if err := s.handleCompletionStatusChange(ctx, story, newStatusID, updates); err != nil {
			s.log.Error(ctx, "failed to handle completion status change", "error", err)
			// Don't fail the entire update - log and continue
		}
	}

	// Update the story
	if err := s.repo.Update(ctx, storyID, workspaceID, updates); err != nil {
		span.RecordError(err)
		return err
	}
	ca := []CoreActivity{}

	for field, value := range updates {
		currentValue := fmt.Sprintf("%v", value)
		na := CoreActivity{
			StoryID:      storyID,
			Type:         "update",
			Field:        field,
			CurrentValue: currentValue,
			UserID:       actorID,
			WorkspaceID:  workspaceID,
		}
		// ignore if field contains description
		if strings.Contains(field, "description") {
			continue
		}
		ca = append(ca, na)
	}
	if _, err := s.repo.RecordActivities(ctx, ca); err != nil {
		span.RecordError(err)
	}

	span.AddEvent("story updated", trace.WithAttributes(
		attribute.String("story.id", storyID.String()),
	))

	// Always publish event - let notification rules decide what to do
	payload := events.StoryUpdatedPayload{
		StoryID:     storyID,
		WorkspaceID: workspaceID,
		Updates:     updates,
		AssigneeID:  story.Assignee, // Current assignee before update
	}

	// Publish event
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

	return nil
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
func (s *Service) HardBulkDelete(ctx context.Context, ids []uuid.UUID, workspaceId uuid.UUID) error {
	s.log.Info(ctx, fmt.Sprintf("Hard bulk deleting stories: %v", ids), "story_ids", ids)
	ctx, span := web.AddSpan(ctx, "business.core.stories.HardBulkDelete")
	defer span.End()

	if err := s.repo.HardBulkDelete(ctx, ids, workspaceId); err != nil {
		s.log.Error(ctx, fmt.Sprintf("Failed to hard bulk delete stories: %s", err),
			"story_ids", ids, "error", err)
		span.RecordError(err)
		return err
	}

	s.log.Info(ctx, fmt.Sprintf("Successfully hard bulk deleted stories: %v", ids),
		"story_ids", ids)
	span.AddEvent("Stories hard bulk deleted.", trace.WithAttributes(
		attribute.Int("stories.count", len(ids))))

	return nil
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

// GetActivities returns the activities for a story with pagination.
func (s *Service) GetActivities(ctx context.Context, storyID uuid.UUID, page, pageSize int) ([]CoreActivity, bool, error) {
	s.log.Info(ctx, "business.core.activities.GetActivities")
	ctx, span := web.AddSpan(ctx, "business.core.activities.GetActivities")
	defer span.End()

	activities, hasMore, err := s.repo.GetActivities(ctx, storyID, page, pageSize)
	if err != nil {
		span.RecordError(err)
		return nil, false, err
	}

	span.AddEvent("activities retrieved.", trace.WithAttributes(
		attribute.Int("activity.count", len(activities)),
		attribute.Int("page", page),
		attribute.Int("pageSize", pageSize),
		attribute.Bool("has.more", hasMore),
	))

	return activities, hasMore, nil
}

// CreateComment creates a comment for a story.
func (s *Service) CreateComment(ctx context.Context, workspaceID uuid.UUID, cnc CoreNewComment) (comments.CoreComment, error) {
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

	// Get actor ID from context
	actorID, _ := mid.GetUserID(ctx)

	// Publish events based on comment type
	if cnc.Parent != nil {
		// This is a reply - get parent comment details
		parentComment, err := s.repo.GetComment(ctx, *cnc.Parent)
		if err != nil {
			s.log.Error(ctx, "failed to get parent comment for notification", "error", err, "parent_id", *cnc.Parent)
		} else {
			// Publish comment reply event
			payload := events.CommentRepliedPayload{
				CommentID:       comment.ID,
				ParentCommentID: *cnc.Parent,
				ParentAuthorID:  parentComment.UserID,
				StoryID:         cnc.StoryID,
				StoryTitle:      story.Title,
				WorkspaceID:     story.Workspace,
				Content:         cnc.Comment,
				Mentions:        cnc.Mentions,
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
			CommentID:   comment.ID,
			StoryID:     cnc.StoryID,
			StoryTitle:  story.Title,
			AssigneeID:  story.Assignee,
			WorkspaceID: story.Workspace,
			Content:     cnc.Comment,
			Mentions:    cnc.Mentions,
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

	return stories, hasMore, nil
}

// ListByCategory returns stories filtered by category with pagination
func (s *Service) ListByCategory(ctx context.Context, workspaceId, userID, teamId uuid.UUID, category string, page, pageSize int) ([]CoreStoryList, bool, error) {
	ctx, span := web.AddSpan(ctx, "business.services.stories.ListByCategory")
	defer span.End()

	stories, hasMore, err := s.repo.ListByCategory(ctx, workspaceId, userID, teamId, category, page, pageSize)
	if err != nil {
		return nil, false, fmt.Errorf("listing stories by category: %w", err)
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
