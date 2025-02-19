package stories

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/complexus-tech/projects-api/internal/core/comments"
	"github.com/complexus-tech/projects-api/internal/core/links"
	"github.com/complexus-tech/projects-api/internal/web/mid"
	"github.com/complexus-tech/projects-api/pkg/email"
	"github.com/complexus-tech/projects-api/pkg/events"
	"github.com/complexus-tech/projects-api/pkg/logger"
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
	Restore(ctx context.Context, id uuid.UUID, workspaceId uuid.UUID) error
	BulkRestore(ctx context.Context, ids []uuid.UUID, workspaceId uuid.UUID) error
	Update(ctx context.Context, id uuid.UUID, workspaceId uuid.UUID, updates map[string]any) error
	UpdateLabels(ctx context.Context, id uuid.UUID, workspaceId uuid.UUID, labels []uuid.UUID) error
	GetStoryLinks(ctx context.Context, storyID uuid.UUID) ([]links.CoreLink, error)
	Create(ctx context.Context, story *CoreSingleStory) (CoreSingleStory, error)
	GetNextSequenceID(ctx context.Context, teamId uuid.UUID, workspaceId uuid.UUID) (int, func() error, func() error, error)
	MyStories(ctx context.Context, workspaceId uuid.UUID) ([]CoreStoryList, error)
	List(ctx context.Context, workspaceId uuid.UUID, filters map[string]any) ([]CoreStoryList, error)
	GetSubStories(ctx context.Context, parentId uuid.UUID, workspaceId uuid.UUID) ([]CoreStoryList, error)
	RecordActivities(ctx context.Context, activities []CoreActivity) ([]CoreActivity, error)
	GetActivities(ctx context.Context, storyID uuid.UUID) ([]CoreActivity, error)
	CreateComment(ctx context.Context, comment CoreNewComment) (comments.CoreComment, error)
	GetComments(ctx context.Context, storyID uuid.UUID) ([]comments.CoreComment, error)
}

type CoreSingleStoryWithSubs struct {
	CoreSingleStory
	SubStories []CoreStoryList `json:"subStories"`
}

// Service provides story-related operations.
type Service struct {
	repo      Repository
	log       *logger.Logger
	publisher *events.Publisher
	email     email.Service
}

// New constructs a new stories service instance with the provided repository.
func New(log *logger.Logger, repo Repository, publisher *events.Publisher, emailService email.Service) *Service {
	return &Service{
		repo:      repo,
		log:       log,
		publisher: publisher,
		email:     emailService,
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

	s.email.SendEmail(ctx, email.Email{
		To:      []string{"joseph@complexus.app"},
		Subject: "Test Email",
		Body:    "This is a test email",
		IsHTML:  true,
	})
	s.email.SendTemplatedEmail(ctx, email.TemplatedEmail{
		To:       []string{"joseph@complexus.app"},
		Template: "test",
		Data:     map[string]any{"name": "Joseph"},
	})

	span.AddEvent("stories retrieved.", trace.WithAttributes(
		attribute.Int("story.count", len(stories)),
	))
	return stories, nil
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
		return nil
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

	// Create event payload
	var assigneeID *uuid.UUID
	if assigneeIDRaw, ok := updates["assignee_id"]; ok {
		if id, ok := assigneeIDRaw.(uuid.UUID); ok {
			assigneeID = &id
		}
	}

	// ignore if there in no assignee
	if story.Assignee == nil && assigneeID == nil {
		return nil
	}

	// ignore if assignee is the actor
	if story.Assignee != nil && assigneeID != nil && *story.Assignee == *assigneeID {
		return nil
	}

	payload := events.StoryUpdatedPayload{
		StoryID:     storyID,
		WorkspaceID: workspaceID,
		Updates:     updates,
		AssigneeID:  story.Assignee,
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

// GetActivities returns the activities for a story.
func (s *Service) GetActivities(ctx context.Context, storyID uuid.UUID) ([]CoreActivity, error) {
	s.log.Info(ctx, "business.core.activities.GetActivities")
	ctx, span := web.AddSpan(ctx, "business.core.activities.GetActivities")
	defer span.End()

	activities, err := s.repo.GetActivities(ctx, storyID)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	span.AddEvent("activities retrieved.", trace.WithAttributes(
		attribute.Int("activity.count", len(activities)),
	))

	return activities, nil
}

// CreateComment creates a comment for a story.
func (s *Service) CreateComment(ctx context.Context, cnc CoreNewComment) (comments.CoreComment, error) {
	s.log.Info(ctx, "business.core.stories.CreateComment")
	ctx, span := web.AddSpan(ctx, "business.core.stories.CreateComment")
	defer span.End()

	comment, err := s.repo.CreateComment(ctx, cnc)
	if err != nil {
		span.RecordError(err)
		return comments.CoreComment{}, err
	}

	span.AddEvent("comment created.", trace.WithAttributes(
		attribute.String("comment.comment", comment.Comment),
	))

	return comment, nil
}

// GetComments returns the comments for a story.
func (s *Service) GetComments(ctx context.Context, storyID uuid.UUID) ([]comments.CoreComment, error) {
	s.log.Info(ctx, "business.core.stories.GetComments")
	ctx, span := web.AddSpan(ctx, "business.core.stories.GetComments")
	defer span.End()

	comments, err := s.repo.GetComments(ctx, storyID)
	if err != nil {
		s.log.Error(ctx, fmt.Sprintf("failed to get comments: %s", err))
		span.RecordError(err)
		return nil, err
	}

	span.AddEvent("comments retrieved.", trace.WithAttributes(
		attribute.Int("comment.count", len(comments)),
	))

	return comments, nil
}
