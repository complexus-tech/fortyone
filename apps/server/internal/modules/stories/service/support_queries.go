package stories

import (
	"context"
	"errors"
	"fmt"
	"strings"

	storydomain "github.com/complexus-tech/projects-api/internal/modules/stories/domain"
	apptracing "github.com/complexus-tech/projects-api/pkg/tracing"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// GetStoryLinks returns the links for a story after applying the same scoped
// visibility check as the primary story read.
func (s *Service) GetStoryLinks(ctx context.Context, storyID, workspaceID uuid.UUID) ([]storydomain.StoryLink, error) {
	repository, ok := s.repo.(storySupportReadRepository)
	if !ok {
		if _, err := s.getVisibleStory(ctx, storyID, workspaceID); err != nil {
			return nil, err
		}
		legacy, legacyOK := s.repo.(interface {
			GetStoryLinks(context.Context, uuid.UUID, uuid.UUID) ([]storydomain.StoryLink, error)
		})
		if !legacyOK {
			return nil, errors.New("story repository does not support scoped link reads")
		}
		return legacy.GetStoryLinks(ctx, storyID, workspaceID)
	}
	scope, err := readScopeFromContext(ctx, workspaceID)
	if err != nil {
		return nil, err
	}
	return repository.ListVisibleStoryLinks(ctx, scope, storyID)
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
	scope, err := readScopeFromContext(ctx, workspaceID)
	if err != nil {
		return nil, err
	}
	return repo.FindFirstStatusByCategory(ctx, scope, teamID, strings.TrimSpace(category))
}

// GetActivitiesWithUser returns the activities for a story with user details and pagination.
func (s *Service) GetActivitiesWithUser(ctx context.Context, storyID, workspaceID uuid.UUID, page, pageSize int) ([]CoreActivityWithUser, bool, error) {
	s.log.Info(ctx, "business.core.activities.GetActivitiesWithUser")
	ctx, span := apptracing.AddSpan(ctx, storyServiceTracer, "business.core.activities.GetActivitiesWithUser")
	defer span.End()

	repository, ok := s.repo.(storySupportReadRepository)
	if !ok {
		if _, err := s.getVisibleStory(ctx, storyID, workspaceID); err != nil {
			return nil, false, err
		}
		legacy, legacyOK := s.repo.(legacyActivityReader)
		if !legacyOK {
			return nil, false, errors.New("story repository does not support scoped activity reads")
		}
		return legacy.GetActivitiesWithUser(ctx, storyID, workspaceID, page, pageSize)
	}
	scope, err := readScopeFromContext(ctx, workspaceID)
	if err != nil {
		return nil, false, err
	}
	rows, hasMore, err := repository.ListVisibleStoryActivities(ctx, scope, storyID, page, pageSize)
	if err != nil {
		span.RecordError(err)
		return nil, false, err
	}

	activities := make([]CoreActivityWithUser, len(rows))
	for index, row := range rows {
		activities[index] = coreActivityWithUser(row)
	}
	span.AddEvent("activities with user details retrieved.", trace.WithAttributes(
		attribute.Int("activity.count", len(rows)),
		attribute.Int("page", page),
		attribute.Int("pageSize", pageSize),
		attribute.Bool("has.more", hasMore),
	))

	return activities, hasMore, nil
}

// DuplicateStory creates a copy of an existing story.
func (s *Service) DuplicateStory(ctx context.Context, originalStoryID uuid.UUID, workspaceId uuid.UUID, userID uuid.UUID) (CoreSingleStory, error) {
	s.log.Info(ctx, "business.core.stories.DuplicateStory")
	ctx, span := apptracing.AddSpan(ctx, storyServiceTracer, "business.core.stories.DuplicateStory")
	defer span.End()
	if repository, migrated := s.repo.(storyDuplicationRepository); migrated {
		duplicatedStory, err := s.duplicateStoryTyped(
			ctx, repository, originalStoryID, workspaceId, userID,
		)
		if err != nil {
			span.RecordError(err)
			return CoreSingleStory{}, fmt.Errorf("failed to duplicate story: %w", err)
		}
		applySingleStoryEstimateLabels(&duplicatedStory)
		span.AddEvent("Story duplicated.", trace.WithAttributes(
			attribute.String("original_story.id", originalStoryID.String()),
			attribute.String("new_story.id", duplicatedStory.ID.String()),
		))
		return duplicatedStory, nil
	}

	legacy, ok := s.repo.(legacyDuplicateStoryRepository)
	if !ok {
		return CoreSingleStory{}, errors.New("story repository does not support duplication")
	}
	duplicatedStory, err := legacy.DuplicateStory(ctx, originalStoryID, workspaceId, userID)
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
