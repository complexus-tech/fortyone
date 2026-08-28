package stories

import (
	"context"
	"errors"
	"fmt"

	"github.com/complexus-tech/projects-api/internal/platform/auth"
	apptracing "github.com/complexus-tech/projects-api/pkg/tracing"
	"github.com/google/uuid"
)

// AddAssociation adds an association between two stories.
func (s *Service) AddAssociation(ctx context.Context, fromID, toID uuid.UUID, associationType string, workspaceID uuid.UUID) (CoreStoryAssociation, error) {
	s.log.Info(ctx, "business.core.stories.AddAssociation")
	ctx, span := apptracing.AddSpan(ctx, storyServiceTracer, "business.core.stories.AddAssociation")
	defer span.End()

	// Validate inputs
	if fromID == toID {
		return CoreStoryAssociation{}, fmt.Errorf("cannot associate story with itself")
	}
	if repository, migrated := s.repo.(storyAssociationMutationRepository); migrated {
		association, err := s.addAssociationTyped(
			ctx, repository, fromID, toID, associationType, workspaceID,
		)
		if err != nil {
			span.RecordError(err)
		}
		return association, err
	}

	legacy, ok := s.repo.(legacyAssociationRepository)
	if !ok {
		return CoreStoryAssociation{}, errors.New("story repository does not support associations")
	}
	assoc, err := legacy.AddAssociation(ctx, fromID, toID, associationType, workspaceID)
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
	ctx, span := apptracing.AddSpan(ctx, storyServiceTracer, "business.core.stories.UpdateAssociation")
	defer span.End()

	if fromID == toID {
		return CoreStoryAssociation{}, fmt.Errorf("cannot associate story with itself")
	}
	if repository, migrated := s.repo.(storyAssociationMutationRepository); migrated {
		association, err := s.updateAssociationTyped(
			ctx, repository, associationID, fromID, toID, associationType, workspaceID,
		)
		if err != nil {
			span.RecordError(err)
		}
		return association, err
	}

	legacy, ok := s.repo.(legacyAssociationRepository)
	if !ok {
		return CoreStoryAssociation{}, errors.New("story repository does not support associations")
	}
	assoc, err := legacy.UpdateAssociation(ctx, associationID, fromID, toID, associationType, workspaceID)
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
	ctx, span := apptracing.AddSpan(ctx, storyServiceTracer, "business.core.stories.RemoveAssociation")
	defer span.End()
	if repository, migrated := s.repo.(storyAssociationMutationRepository); migrated {
		err := s.removeAssociationTyped(ctx, repository, associationID, workspaceID)
		if err != nil {
			span.RecordError(err)
		}
		return err
	}

	legacy, ok := s.repo.(legacyAssociationRepository)
	if !ok {
		return errors.New("story repository does not support associations")
	}
	assoc, err := legacy.RemoveAssociation(ctx, associationID, workspaceID)
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
	return s.recordActivities(ctx, activities)
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
	case "auto_scheduling_enabled":
		return story.AutoSchedulingEnabled
	case "auto_scheduling_locked":
		return story.AutoSchedulingLocked
	case "auto_scheduling_status":
		return story.AutoSchedulingStatus
	case "auto_scheduling_reason":
		return story.AutoSchedulingReason
	case "auto_scheduling_updated_at":
		return story.AutoSchedulingUpdatedAt
	default:
		return nil
	}
}
