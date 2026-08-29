package objectives

import (
	"context"
	"errors"

	objectivesdomain "github.com/complexus-tech/projects-api/internal/modules/objectives/domain"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/google/uuid"
)

func (service *Service) Get(ctx context.Context, id, workspaceID uuid.UUID) (CoreObjective, error) {
	query := objectivesdomain.GetQuery{ObjectiveID: id, WorkspaceID: workspaceID}
	if _, err := platformauth.GetActor(ctx); errors.Is(err, platformauth.ErrActorNotFound) {
		// Trusted background consumers historically call Get without an actor.
		// This compatibility path remains strictly tenant- and identifier-scoped.
		query.Internal = true
		return service.repo.Get(ctx, query)
	}
	actor, err := actorFor(ctx, workspaceID, uuid.Nil, platformauth.ScopeObjectivesRead)
	if err != nil {
		return CoreObjective{}, err
	}
	query.ActorID = actor.PrincipalID
	return service.repo.Get(ctx, query)
}

func (service *Service) List(
	ctx context.Context,
	workspaceID, userID uuid.UUID,
	filters map[string]any,
) ([]CoreObjective, error) {
	query, err := listQueryFromCompatibilityMap(workspaceID, userID, filters)
	if err != nil {
		return nil, err
	}
	return service.ListIntent(ctx, query)
}

// ListIntent is the typed list boundary used by HTTP and new integrations.
// List remains above as a finite compatibility adapter for older in-process
// callers; maps never cross into the repository.
func (service *Service) ListIntent(
	ctx context.Context,
	query objectivesdomain.ListQuery,
) ([]CoreObjective, error) {
	actor, err := actorFor(ctx, query.WorkspaceID, query.ActorID, platformauth.ScopeObjectivesRead)
	if err != nil {
		return nil, err
	}
	query.ActorID = actor.PrincipalID
	query, err = query.Normalize()
	if err != nil {
		return nil, err
	}
	return service.repo.List(ctx, query)
}

func (service *Service) GetAnalytics(
	ctx context.Context,
	objectiveID, workspaceID uuid.UUID,
) (CoreObjectiveAnalytics, error) {
	actor, err := actorFor(ctx, workspaceID, uuid.Nil, platformauth.ScopeObjectivesRead)
	if err != nil {
		return CoreObjectiveAnalytics{}, err
	}
	return service.repo.GetAnalytics(ctx, objectivesdomain.AnalyticsQuery{
		ObjectiveID: objectiveID, WorkspaceID: workspaceID, ActorID: actor.PrincipalID,
	}, service.now().UTC())
}
