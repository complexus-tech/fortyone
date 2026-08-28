package stories

import (
	"context"
	"fmt"
	"time"

	storydomain "github.com/complexus-tech/projects-api/internal/modules/stories/domain"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/google/uuid"
)

type storyAssociationMutationRepository interface {
	GetStoryForMutation(context.Context, storydomain.MutationScope, uuid.UUID) (storydomain.Story, error)
	PrepareStoryAssociationMutation(
		context.Context,
		storydomain.MutationScope,
		uuid.UUID,
	) (storydomain.AssociationSnapshot, error)
	ApplyStoryAssociationMutation(
		context.Context,
		storydomain.AssociationMutationCommand,
	) (storydomain.AssociationSnapshot, error)
}

type storyAssociationIntegrationPayload struct {
	StoryID                 uuid.UUID                             `json:"storyId"`
	WorkspaceID             uuid.UUID                             `json:"workspaceId"`
	AssociationID           uuid.UUID                             `json:"associationId"`
	Action                  storydomain.AssociationMutationAction `json:"action"`
	FromStoryID             uuid.UUID                             `json:"fromStoryId"`
	ToStoryID               uuid.UUID                             `json:"toStoryId"`
	AssociationType         string                                `json:"associationType"`
	PreviousFromStoryID     *uuid.UUID                            `json:"previousFromStoryId,omitempty"`
	PreviousToStoryID       *uuid.UUID                            `json:"previousToStoryId,omitempty"`
	PreviousAssociationType *string                               `json:"previousAssociationType,omitempty"`
}

func (s *Service) addAssociationTyped(
	ctx context.Context,
	repository storyAssociationMutationRepository,
	fromID, toID uuid.UUID,
	associationType string,
	workspaceID uuid.UUID,
) (CoreStoryAssociation, error) {
	scope, err := mutationScope(ctx, workspaceID, uuid.Nil, platformauth.PrincipalHumanUser)
	if err != nil {
		return CoreStoryAssociation{}, mapStoryMutationError(err)
	}
	return s.applyAssociationTyped(ctx, repository, scope, storydomain.AssociationMutationAdd,
		storydomain.AssociationSnapshot{
			ID: uuid.New(), FromStoryID: fromID, ToStoryID: toID, Type: associationType,
		}, nil)
}

func (s *Service) updateAssociationTyped(
	ctx context.Context,
	repository storyAssociationMutationRepository,
	associationID, fromID, toID uuid.UUID,
	associationType string,
	workspaceID uuid.UUID,
) (CoreStoryAssociation, error) {
	scope, err := mutationScope(ctx, workspaceID, uuid.Nil, platformauth.PrincipalHumanUser)
	if err != nil {
		return CoreStoryAssociation{}, mapStoryMutationError(err)
	}
	expected, err := repository.PrepareStoryAssociationMutation(ctx, scope, associationID)
	if err != nil {
		return CoreStoryAssociation{}, mapStoryMutationError(err)
	}
	next := storydomain.AssociationSnapshot{
		ID: associationID, FromStoryID: fromID, ToStoryID: toID, Type: associationType,
	}
	return s.applyAssociationTyped(
		ctx, repository, scope, storydomain.AssociationMutationUpdate, next, &expected,
	)
}

func (s *Service) removeAssociationTyped(
	ctx context.Context,
	repository storyAssociationMutationRepository,
	associationID, workspaceID uuid.UUID,
) error {
	scope, err := mutationScope(ctx, workspaceID, uuid.Nil, platformauth.PrincipalHumanUser)
	if err != nil {
		return mapStoryMutationError(err)
	}
	expected, err := repository.PrepareStoryAssociationMutation(ctx, scope, associationID)
	if err != nil {
		return mapStoryMutationError(err)
	}
	_, err = s.applyAssociationTyped(
		ctx, repository, scope, storydomain.AssociationMutationRemove, expected, &expected,
	)
	return err
}

func (s *Service) applyAssociationTyped(
	ctx context.Context,
	repository storyAssociationMutationRepository,
	scope storydomain.MutationScope,
	action storydomain.AssociationMutationAction,
	association storydomain.AssociationSnapshot,
	expected *storydomain.AssociationSnapshot,
) (CoreStoryAssociation, error) {
	storyByID := make(map[uuid.UUID]storydomain.Story, 2)
	titles := make(map[uuid.UUID]string, 4)
	if expected != nil {
		titles[expected.FromStoryID] = expected.FromStoryTitle
		titles[expected.ToStoryID] = expected.ToStoryTitle
	}
	if action != storydomain.AssociationMutationRemove {
		for _, storyID := range []uuid.UUID{association.FromStoryID, association.ToStoryID} {
			story, err := repository.GetStoryForMutation(ctx, scope, storyID)
			if err != nil {
				return CoreStoryAssociation{}, mapStoryMutationError(err)
			}
			storyByID[storyID] = story
			titles[storyID] = story.Title
		}
	}
	association.FromStoryTitle = titles[association.FromStoryID]
	association.ToStoryTitle = titles[association.ToStoryID]

	changedAt := time.Now().UTC()
	affectedStoryIDs := associationAffectedStoryIDs(action, association, expected)
	previousFromStoryID, previousToStoryID := associationPreviousStoryIDs(expected)
	events := make([]storydomain.MutationEvent, 0, len(affectedStoryIDs))
	for _, storyID := range affectedStoryIDs {
		event, err := newStoryMutationEvent(
			scope,
			storyID,
			storydomain.MutationEventStoryUpdated,
			storyAssociationIntegrationPayload{
				StoryID: storyID, WorkspaceID: scope.WorkspaceID, AssociationID: association.ID,
				Action: action, FromStoryID: association.FromStoryID, ToStoryID: association.ToStoryID,
				AssociationType:     association.Type,
				PreviousFromStoryID: previousFromStoryID, PreviousToStoryID: previousToStoryID,
				PreviousAssociationType: associationPreviousType(expected),
			},
			changedAt,
		)
		if err != nil {
			return CoreStoryAssociation{}, err
		}
		events = append(events, event)
	}
	activities, err := s.associationMutationActivities(
		scope, action, association, expected, affectedStoryIDs, titles, changedAt,
	)
	if err != nil {
		return CoreStoryAssociation{}, err
	}

	result, err := repository.ApplyStoryAssociationMutation(ctx, storydomain.AssociationMutationCommand{
		Scope: scope, Action: action, Association: association, Expected: expected,
		OccurredAt: changedAt, Events: events, Activities: activities,
	})
	if err != nil {
		return CoreStoryAssociation{}, mapStoryMutationError(err)
	}
	response := associationSnapshotToCore(result)
	if target, exists := storyByID[result.ToStoryID]; exists {
		response.Story = storyToList(target)
	}
	return response, nil
}

func (s *Service) associationMutationActivities(
	scope storydomain.MutationScope,
	action storydomain.AssociationMutationAction,
	association storydomain.AssociationSnapshot,
	expected *storydomain.AssociationSnapshot,
	storyIDs []uuid.UUID,
	titles map[uuid.UUID]string,
	changedAt time.Time,
) ([]storydomain.MutationActivity, error) {
	if scope.ActivityUser == nil {
		return nil, nil
	}
	reason := associationMutationReason(action)
	activities := make([]storydomain.MutationActivity, 0, len(storyIDs))
	for _, storyID := range storyIDs {
		current, hasCurrent := associationReferenceForStory(association, storyID)
		previous, hasPrevious := associationReferenceForExpected(expected, storyID)
		if action == storydomain.AssociationMutationRemove {
			hasCurrent = false
		}
		reference := current
		if !hasCurrent {
			reference = previous
		}
		if reference.storyID == uuid.Nil {
			return nil, fmt.Errorf("%w: association activity reference is incomplete", ErrInvalidStoryMutation)
		}
		currentValue := titles[reference.storyID]
		if currentValue == "" {
			currentValue = reference.storyID.String()
		}
		oldValue, newValue := associationActivityValues(
			action, previous, hasPrevious, current, hasCurrent,
		)
		activity, err := newStoryMutationActivity(
			scope, storyID, "update", reference.field, currentValue,
			oldValue, newValue, &reason, changedAt,
		)
		if err != nil {
			return nil, err
		}
		if activity != nil {
			activities = append(activities, *activity)
		}
	}
	return activities, nil
}

type associationActivityReference struct {
	storyID uuid.UUID
	field   string
	label   string
}

func associationReferenceForExpected(
	expected *storydomain.AssociationSnapshot,
	storyID uuid.UUID,
) (associationActivityReference, bool) {
	if expected == nil {
		return associationActivityReference{}, false
	}
	return associationReferenceForStory(*expected, storyID)
}

func associationReferenceForStory(
	association storydomain.AssociationSnapshot,
	storyID uuid.UUID,
) (associationActivityReference, bool) {
	switch storyID {
	case association.FromStoryID:
		return associationActivityReference{
			storyID: association.ToStoryID,
			field:   outgoingAssociationActivityField(association.Type),
			label:   outgoingAssociationActivityLabel(association.Type),
		}, true
	case association.ToStoryID:
		return associationActivityReference{
			storyID: association.FromStoryID,
			field:   incomingAssociationActivityField(association.Type),
			label:   incomingAssociationActivityLabel(association.Type),
		}, true
	default:
		return associationActivityReference{}, false
	}
}

func associationActivityValues(
	action storydomain.AssociationMutationAction,
	previous associationActivityReference,
	hasPrevious bool,
	current associationActivityReference,
	hasCurrent bool,
) (any, any) {
	if action == storydomain.AssociationMutationAdd {
		return nil, current.storyID
	}
	if action == storydomain.AssociationMutationRemove || !hasCurrent {
		return previous.storyID, nil
	}
	var oldValue any
	if hasPrevious {
		switch {
		case previous.field != current.field:
			oldValue = previous.label
		case previous.storyID != current.storyID:
			oldValue = previous.storyID
		}
	}
	return oldValue, current.storyID
}

func associationAffectedStoryIDs(
	action storydomain.AssociationMutationAction,
	association storydomain.AssociationSnapshot,
	expected *storydomain.AssociationSnapshot,
) []uuid.UUID {
	values := []uuid.UUID{association.FromStoryID, association.ToStoryID}
	if action != storydomain.AssociationMutationAdd && expected != nil {
		values = append(values, expected.FromStoryID, expected.ToStoryID)
	}
	seen := make(map[uuid.UUID]struct{}, len(values))
	result := make([]uuid.UUID, 0, len(values))
	for _, value := range values {
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}

func associationPreviousType(expected *storydomain.AssociationSnapshot) *string {
	if expected == nil {
		return nil
	}
	value := expected.Type
	return &value
}

func associationPreviousStoryIDs(expected *storydomain.AssociationSnapshot) (*uuid.UUID, *uuid.UUID) {
	if expected == nil {
		return nil, nil
	}
	fromStoryID, toStoryID := expected.FromStoryID, expected.ToStoryID
	return &fromStoryID, &toStoryID
}

func associationMutationReason(action storydomain.AssociationMutationAction) string {
	switch action {
	case storydomain.AssociationMutationAdd:
		return associationActivityAdded
	case storydomain.AssociationMutationUpdate:
		return associationActivityUpdated
	default:
		return associationActivityRemoved
	}
}

func associationSnapshotToCore(snapshot storydomain.AssociationSnapshot) CoreStoryAssociation {
	return CoreStoryAssociation{
		ID: snapshot.ID, FromStoryID: snapshot.FromStoryID, ToStoryID: snapshot.ToStoryID,
		Type: snapshot.Type, PreviousType: snapshot.PreviousType,
		FromStoryTitle: snapshot.FromStoryTitle, ToStoryTitle: snapshot.ToStoryTitle,
	}
}

func storyToList(story storydomain.Story) CoreStoryList {
	return CoreStoryList{
		ID: story.ID, SequenceID: story.SequenceID, Title: story.Title,
		EstimateLabel: story.EstimateLabel, EstimateValue: story.EstimateValue,
		EstimateScheme:           story.EstimateScheme,
		EstimatedDurationMinutes: story.EstimatedDurationMinutes,
		MinimumFocusBlockMinutes: story.MinimumFocusBlockMinutes,
		AutoSchedulingEnabled:    story.AutoSchedulingEnabled,
		AutoSchedulingLocked:     story.AutoSchedulingLocked,
		AutoSchedulingStatus:     story.AutoSchedulingStatus,
		AutoSchedulingReason:     story.AutoSchedulingReason,
		AutoSchedulingUpdatedAt:  story.AutoSchedulingUpdatedAt,
		Parent:                   story.Parent, Objective: story.Objective, Epic: story.Epic,
		Status: story.Status, Assignee: story.Assignee,
		CollaboratorCount: len(story.Collaborators), Reporter: story.Reporter,
		Priority: story.Priority, Sprint: story.Sprint, SprintSummary: story.SprintSummary,
		KeyResult: story.KeyResult, Team: story.Team, Workspace: story.Workspace,
		StartDate: story.StartDate, EndDate: story.EndDate,
		CreatedAt: story.CreatedAt, UpdatedAt: story.UpdatedAt,
		CompletedAt: story.CompletedAt, DeletedAt: story.DeletedAt, ArchivedAt: story.ArchivedAt,
		Labels:     append([]uuid.UUID(nil), story.Labels...),
		SubStories: append([]storydomain.StoryList(nil), story.SubStories...),
	}
}
