package apiv1http

import (
	"context"
	"net/http"

	openapiv1 "github.com/complexus-tech/projects-api/internal/generated/openapi/v1"
	keyresultsdomain "github.com/complexus-tech/projects-api/internal/modules/keyresults/domain"
	labels "github.com/complexus-tech/projects-api/internal/modules/labels/service"
	objectivesdomain "github.com/complexus-tech/projects-api/internal/modules/objectives/domain"
	sprintdomain "github.com/complexus-tech/projects-api/internal/modules/sprints/domain"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/google/uuid"
)

func (s *server) ListLabels(ctx context.Context, request openapiv1.ListLabelsRequestObject) (openapiv1.ListLabelsResponseObject, error) {
	actor, problem := humanActorFor(ctx, request.WorkspaceId, platformauth.ScopeLabelsRead)
	if problem == nil {
		problem = s.requireVisibleTeam(ctx, actor, request.WorkspaceId, request.Params.TeamId)
	}
	if problem != nil {
		return listLabelsFailure(ctx, problem), nil
	}
	page, problem := decodeOffsetPage(
		s.labelCursors, request.Params.Cursor, request.Params.Limit,
		request.WorkspaceId, actor.PrincipalID, request.Params.TeamId, uuid.Nil,
	)
	if problem != nil {
		return listLabelsFailure(ctx, problem), nil
	}
	resultLimit := page.Limit + 1
	items, err := s.labels.GetLabels(ctx, actor.PrincipalID, request.WorkspaceId, labels.LabelFilters{
		TeamID: &request.Params.TeamId, Limit: &resultLimit, Offset: page.Offset,
	})
	if err != nil {
		return listLabelsFailure(ctx, classifyFailure(err)), nil
	}
	items, hasMore := trimPage(items, page.Limit)
	data := make([]openapiv1.ComponentsResourcesLabel, len(items))
	for index, item := range items {
		data[index] = labelModel(item)
	}
	nextCursor, problem := s.nextOffsetCursor(ctx, s.labelCursors, page, hasMore, page.Offset+page.Limit)
	if problem != nil {
		return listLabelsFailure(ctx, problem), nil
	}
	return openapiv1.ListLabels200JSONResponse{Body: openapiv1.ComponentsResourcesLabelPageResponse{
		Data: data, Meta: pageMeta(hasMore, nextCursor),
	}}, nil
}

func (s *server) ListWorkflowStates(ctx context.Context, request openapiv1.ListWorkflowStatesRequestObject) (openapiv1.ListWorkflowStatesResponseObject, error) {
	actor, problem := humanActorFor(ctx, request.WorkspaceId, platformauth.ScopeStoriesRead)
	if problem == nil {
		problem = s.requireVisibleTeam(ctx, actor, request.WorkspaceId, request.Params.TeamId)
	}
	if problem != nil {
		return listWorkflowStatesFailure(ctx, problem), nil
	}
	page, problem := decodeOffsetPage(
		s.workflowStateCursors, request.Params.Cursor, request.Params.Limit,
		request.WorkspaceId, actor.PrincipalID, request.Params.TeamId, uuid.Nil,
	)
	if problem != nil {
		return listWorkflowStatesFailure(ctx, problem), nil
	}
	items, err := s.workflowStates.TeamListForMember(ctx, request.WorkspaceId, request.Params.TeamId, actor.PrincipalID)
	if err != nil {
		return listWorkflowStatesFailure(ctx, classifyFailure(err)), nil
	}
	items, hasMore := slicePage(items, page.Offset, page.Limit)
	data := make([]openapiv1.ComponentsResourcesWorkflowState, len(items))
	for index, item := range items {
		data[index] = workflowStateModel(item)
	}
	nextCursor, problem := s.nextOffsetCursor(ctx, s.workflowStateCursors, page, hasMore, page.Offset+page.Limit)
	if problem != nil {
		return listWorkflowStatesFailure(ctx, problem), nil
	}
	return openapiv1.ListWorkflowStates200JSONResponse{Body: openapiv1.ComponentsResourcesWorkflowStatePageResponse{
		Data: data, Meta: pageMeta(hasMore, nextCursor),
	}}, nil
}

func (s *server) ListSprints(ctx context.Context, request openapiv1.ListSprintsRequestObject) (openapiv1.ListSprintsResponseObject, error) {
	actor, problem := humanActorFor(ctx, request.WorkspaceId, platformauth.ScopeSprintsRead)
	teamID := request.Params.TeamId
	if problem == nil {
		teamID, problem = effectiveTeamFilter(actor, teamID)
	}
	if problem != nil {
		return listSprintsFailure(ctx, problem), nil
	}
	page, problem := decodeOffsetPage(
		s.sprintCursors, request.Params.Cursor, request.Params.Limit,
		request.WorkspaceId, actor.PrincipalID, uuidValue(teamID), uuid.Nil,
	)
	if problem != nil {
		return listSprintsFailure(ctx, problem), nil
	}
	items, err := s.sprints.ListQuery(ctx, sprintdomain.ListQuery{
		WorkspaceID: request.WorkspaceId, ActorID: actor.PrincipalID,
		Filter: sprintdomain.ListFilter{TeamID: teamID, Limit: page.Limit + 1, Offset: page.Offset},
	})
	if err != nil {
		return listSprintsFailure(ctx, classifyFailure(err)), nil
	}
	items, hasMore := trimPage(items, page.Limit)
	data := make([]openapiv1.ComponentsResourcesSprint, len(items))
	for index, item := range items {
		data[index] = sprintModel(item)
	}
	nextCursor, problem := s.nextOffsetCursor(ctx, s.sprintCursors, page, hasMore, page.Offset+page.Limit)
	if problem != nil {
		return listSprintsFailure(ctx, problem), nil
	}
	return openapiv1.ListSprints200JSONResponse{Body: openapiv1.ComponentsResourcesSprintPageResponse{
		Data: data, Meta: pageMeta(hasMore, nextCursor),
	}}, nil
}

func (s *server) ListObjectives(ctx context.Context, request openapiv1.ListObjectivesRequestObject) (openapiv1.ListObjectivesResponseObject, error) {
	actor, problem := humanActorFor(ctx, request.WorkspaceId, platformauth.ScopeObjectivesRead)
	teamID := request.Params.TeamId
	if problem == nil {
		teamID, problem = effectiveTeamFilter(actor, teamID)
	}
	if problem != nil {
		return listObjectivesFailure(ctx, problem), nil
	}
	page, problem := decodeOffsetPage(
		s.objectiveCursors, request.Params.Cursor, request.Params.Limit,
		request.WorkspaceId, actor.PrincipalID, uuidValue(teamID), uuid.Nil,
	)
	if problem != nil {
		return listObjectivesFailure(ctx, problem), nil
	}
	items, err := s.objectives.ListIntent(ctx, objectivesdomain.ListQuery{
		WorkspaceID: request.WorkspaceId, ActorID: actor.PrincipalID, TeamID: teamID,
		Limit: page.Limit + 1, Offset: page.Offset,
	})
	if err != nil {
		return listObjectivesFailure(ctx, classifyFailure(err)), nil
	}
	items, hasMore := trimPage(items, page.Limit)
	data := make([]openapiv1.ComponentsResourcesObjective, len(items))
	for index, item := range items {
		data[index] = objectiveModel(item)
	}
	nextCursor, problem := s.nextOffsetCursor(ctx, s.objectiveCursors, page, hasMore, page.Offset+page.Limit)
	if problem != nil {
		return listObjectivesFailure(ctx, problem), nil
	}
	return openapiv1.ListObjectives200JSONResponse{Body: openapiv1.ComponentsResourcesObjectivePageResponse{
		Data: data, Meta: pageMeta(hasMore, nextCursor),
	}}, nil
}

func (s *server) ListKeyResults(ctx context.Context, request openapiv1.ListKeyResultsRequestObject) (openapiv1.ListKeyResultsResponseObject, error) {
	actor, problem := humanActorFor(ctx, request.WorkspaceId, platformauth.ScopeObjectivesRead)
	if problem == nil && request.Params.TeamId != nil && !actor.TeamAccess.Allows(*request.Params.TeamId) {
		problem = teamAccessDenied()
	}
	if problem != nil {
		return listKeyResultsFailure(ctx, problem), nil
	}
	page, problem := decodeOffsetPage(
		s.keyResultCursors, request.Params.Cursor, request.Params.Limit,
		request.WorkspaceId, actor.PrincipalID, uuidValue(request.Params.TeamId), uuidValue(request.Params.ObjectiveId),
	)
	if problem != nil {
		return listKeyResultsFailure(ctx, problem), nil
	}
	filters := keyresultsdomain.Filters{
		WorkspaceID: request.WorkspaceId, CurrentUserID: actor.PrincipalID,
		Page: page.Offset/page.Limit + 1, PageSize: page.Limit,
		OrderBy: "created_at", OrderDirection: "desc",
	}
	if request.Params.TeamId != nil {
		filters.TeamIDs = []uuid.UUID{*request.Params.TeamId}
	}
	if request.Params.ObjectiveId != nil {
		filters.ObjectiveIDs = []uuid.UUID{*request.Params.ObjectiveId}
	}
	result, err := s.keyResults.ListPaginated(ctx, filters)
	if err != nil {
		return listKeyResultsFailure(ctx, classifyFailure(err)), nil
	}
	data := make([]openapiv1.ComponentsResourcesKeyResult, len(result.KeyResults))
	for index, item := range result.KeyResults {
		data[index] = keyResultModel(item)
	}
	nextCursor, problem := s.nextOffsetCursor(ctx, s.keyResultCursors, page, result.HasMore, page.Offset+page.Limit)
	if problem != nil {
		return listKeyResultsFailure(ctx, problem), nil
	}
	return openapiv1.ListKeyResults200JSONResponse{Body: openapiv1.ComponentsResourcesKeyResultPageResponse{
		Data: data, Meta: pageMeta(result.HasMore, nextCursor),
	}}, nil
}

func (s *server) ListStoryComments(ctx context.Context, request openapiv1.ListStoryCommentsRequestObject) (openapiv1.ListStoryCommentsResponseObject, error) {
	actor, problem := humanActorFor(ctx, request.WorkspaceId, platformauth.ScopeCommentsRead)
	if problem != nil {
		return listStoryCommentsFailure(ctx, problem), nil
	}
	page, problem := decodeOffsetPage(
		s.commentCursors, request.Params.Cursor, request.Params.Limit,
		request.WorkspaceId, actor.PrincipalID, uuid.Nil, request.StoryId,
	)
	if problem != nil {
		return listStoryCommentsFailure(ctx, problem), nil
	}
	items, hasMore, err := s.storyComments.GetComments(
		ctx, request.StoryId, request.WorkspaceId, page.Offset/page.Limit+1, page.Limit,
	)
	if err != nil {
		return listStoryCommentsFailure(ctx, classifyFailure(err)), nil
	}
	data := make([]openapiv1.ComponentsResourcesComment, len(items))
	for index, item := range items {
		data[index] = commentModel(item)
	}
	nextCursor, problem := s.nextOffsetCursor(ctx, s.commentCursors, page, hasMore, page.Offset+page.Limit)
	if problem != nil {
		return listStoryCommentsFailure(ctx, problem), nil
	}
	return openapiv1.ListStoryComments200JSONResponse{Body: openapiv1.ComponentsResourcesCommentPageResponse{
		Data: data, Meta: pageMeta(hasMore, nextCursor),
	}}, nil
}

func (s *server) GetStoryComment(ctx context.Context, request openapiv1.GetStoryCommentRequestObject) (openapiv1.GetStoryCommentResponseObject, error) {
	_, problem := humanActorFor(ctx, request.WorkspaceId, platformauth.ScopeCommentsRead)
	if problem != nil {
		return getStoryCommentFailure(ctx, problem), nil
	}
	comment, err := s.storyComments.GetComment(ctx, request.CommentId, request.StoryId, request.WorkspaceId)
	if err != nil {
		return getStoryCommentFailure(ctx, classifyFailure(err)), nil
	}
	return openapiv1.GetStoryComment200JSONResponse{
		Body: openapiv1.ComponentsResourcesCommentResponse{Data: commentModel(comment)},
	}, nil
}

func humanActorFor(ctx context.Context, workspaceID uuid.UUID, scope platformauth.Scope) (platformauth.Actor, *failure) {
	actor, problem := actorFor(ctx, workspaceID, scope)
	if problem == nil {
		problem = requireUserCredential(actor)
	}
	return actor, problem
}

func (s *server) requireVisibleTeam(
	ctx context.Context,
	actor platformauth.Actor,
	workspaceID, teamID uuid.UUID,
) *failure {
	if !actor.TeamAccess.Allows(teamID) {
		return teamAccessDenied()
	}
	if _, err := s.teams.GetByID(ctx, teamID, workspaceID, actor.PrincipalID); err != nil {
		if classifyFailure(err).status == http.StatusNotFound {
			return teamAccessDenied()
		}
		return classifyFailure(err)
	}
	return nil
}

func effectiveTeamFilter(actor platformauth.Actor, requested *uuid.UUID) (*uuid.UUID, *failure) {
	if requested != nil {
		if !actor.TeamAccess.Allows(*requested) {
			return nil, teamAccessDenied()
		}
		value := *requested
		return &value, nil
	}
	if actor.TeamAccess.IsUnrestricted() {
		return nil, nil
	}
	teamIDs := actor.TeamAccess.RestrictedTeamIDs()
	switch len(teamIDs) {
	case 0:
		return nil, teamAccessDenied()
	case 1:
		return &teamIDs[0], nil
	default:
		return nil, &failure{
			status: http.StatusBadRequest, code: "team_filter_required",
			message: "A teamId is required when the credential is restricted to multiple teams.",
		}
	}
}

func teamAccessDenied() *failure {
	return &failure{
		status: http.StatusForbidden, code: "team_access_denied",
		message: "The credential is not allowed to access this team.",
	}
}

func uuidValue(value *uuid.UUID) uuid.UUID {
	if value == nil {
		return uuid.Nil
	}
	return *value
}

func trimPage[T any](items []T, limit int) ([]T, bool) {
	if len(items) <= limit {
		return items, false
	}
	return items[:limit], true
}

func slicePage[T any](items []T, offset, limit int) ([]T, bool) {
	if offset >= len(items) {
		return []T{}, false
	}
	end := min(offset+limit, len(items))
	return items[offset:end], end < len(items)
}

func pageMeta(hasMore bool, nextCursor *string) openapiv1.ComponentsCommonPageMeta {
	return openapiv1.ComponentsCommonPageMeta{HasMore: hasMore, NextCursor: nextCursor}
}
