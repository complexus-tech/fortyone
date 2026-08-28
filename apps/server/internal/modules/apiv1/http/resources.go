package apiv1http

import (
	"context"
	"net/http"

	openapiv1 "github.com/complexus-tech/projects-api/internal/generated/openapi/v1"
	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	teams "github.com/complexus-tech/projects-api/internal/modules/teams/service"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

func (s *server) GetWorkspace(ctx context.Context, request openapiv1.GetWorkspaceRequestObject) (openapiv1.GetWorkspaceResponseObject, error) {
	actor, problem := actorFor(ctx, request.WorkspaceId, platformauth.ScopeWorkspacesRead)
	if problem == nil {
		problem = requireUserCredential(actor)
	}
	if problem != nil {
		return getWorkspaceFailure(ctx, problem), nil
	}
	workspace, err := s.workspaces.Get(ctx, request.WorkspaceId, actor.PrincipalID)
	if err != nil {
		return getWorkspaceFailure(ctx, classifyFailure(err)), nil
	}
	return openapiv1.GetWorkspace200JSONResponse{
		Body: openapiv1.ComponentsResourcesWorkspaceResponse{Data: workspaceModel(workspace)},
	}, nil
}

func (s *server) ListTeams(ctx context.Context, request openapiv1.ListTeamsRequestObject) (openapiv1.ListTeamsResponseObject, error) {
	actor, problem := actorFor(ctx, request.WorkspaceId, platformauth.ScopeTeamsRead)
	if problem == nil {
		problem = requireUserCredential(actor)
	}
	if problem != nil {
		return listTeamsFailure(ctx, problem), nil
	}
	page, problem := s.teamPage(request, actor)
	if problem != nil {
		return listTeamsFailure(ctx, problem), nil
	}
	items, err := s.teams.List(ctx, request.WorkspaceId, actor.PrincipalID, teams.CoreListTeamsFilter{
		Limit: page.Limit + 1, Offset: page.Offset, JoinedOnly: true,
	})
	if err != nil {
		return listTeamsFailure(ctx, classifyFailure(err)), nil
	}
	visible := make([]openapiv1.ComponentsResourcesTeam, 0, min(len(items), page.Limit))
	for _, team := range items {
		if actor.TeamAccess.Allows(team.ID) {
			visible = append(visible, teamModel(team))
		}
		if len(visible) == page.Limit {
			break
		}
	}
	hasMore := len(items) > page.Limit
	nextCursor, problem := s.nextOffsetCursor(ctx, s.teamCursors, page, hasMore, page.Offset+page.Limit)
	if problem != nil {
		return listTeamsFailure(ctx, problem), nil
	}
	return openapiv1.ListTeams200JSONResponse{Body: openapiv1.ComponentsResourcesTeamPageResponse{
		Data: visible, Meta: openapiv1.ComponentsCommonPageMeta{HasMore: hasMore, NextCursor: nextCursor},
	}}, nil
}

func (s *server) ListStories(ctx context.Context, request openapiv1.ListStoriesRequestObject) (openapiv1.ListStoriesResponseObject, error) {
	actor, problem := actorFor(ctx, request.WorkspaceId, platformauth.ScopeStoriesRead)
	if problem == nil {
		problem = requireUserCredential(actor)
	}
	if problem != nil {
		return listStoriesFailure(ctx, problem), nil
	}
	page, problem := s.storyPage(request, actor)
	if problem != nil {
		return listStoriesFailure(ctx, problem), nil
	}
	filters := stories.CoreStoryFilters{Limit: page.Limit + 1, Offset: page.Offset}
	if request.Params.TeamId != nil {
		if !actor.TeamAccess.Allows(*request.Params.TeamId) {
			return listStoriesFailure(ctx, &failure{http.StatusForbidden, "team_access_denied", "The credential is not allowed to access this team."}), nil
		}
		filters.TeamIDs = []uuid.UUID{*request.Params.TeamId}
	}
	items, err := s.stories.List(ctx, request.WorkspaceId, filters)
	if err != nil {
		return listStoriesFailure(ctx, classifyFailure(err)), nil
	}
	hasMore := len(items) > page.Limit
	if hasMore {
		items = items[:page.Limit]
	}
	data := make([]openapiv1.ComponentsResourcesStory, len(items))
	for index, story := range items {
		data[index] = storyListModel(story)
	}
	nextCursor, problem := s.nextOffsetCursor(ctx, s.storyCursors, page, hasMore, page.Offset+page.Limit)
	if problem != nil {
		return listStoriesFailure(ctx, problem), nil
	}
	return openapiv1.ListStories200JSONResponse{Body: openapiv1.ComponentsResourcesStoryPageResponse{
		Data: data, Meta: openapiv1.ComponentsCommonPageMeta{HasMore: hasMore, NextCursor: nextCursor},
	}}, nil
}

func (s *server) GetStory(ctx context.Context, request openapiv1.GetStoryRequestObject) (openapiv1.GetStoryResponseObject, error) {
	actor, problem := actorFor(ctx, request.WorkspaceId, platformauth.ScopeStoriesRead)
	if problem == nil {
		problem = requireUserCredential(actor)
	}
	if problem != nil {
		return getStoryFailure(ctx, problem), nil
	}
	story, err := s.stories.Get(ctx, request.StoryId, request.WorkspaceId)
	if err != nil {
		return getStoryFailure(ctx, classifyFailure(err)), nil
	}
	if !actor.TeamAccess.Allows(story.Team) {
		return getStoryFailure(ctx, &failure{http.StatusNotFound, "resource_not_found", "The requested resource was not found."}), nil
	}
	return openapiv1.GetStory200JSONResponse{
		Body: openapiv1.ComponentsResourcesStoryResponse{Data: storyDetailModel(story)},
	}, nil
}

func (s *server) teamPage(request openapiv1.ListTeamsRequestObject, actor platformauth.Actor) (cursorPage, *failure) {
	return decodeOffsetPage(s.teamCursors, request.Params.Cursor, request.Params.Limit, request.WorkspaceId, actor.PrincipalID, uuid.Nil, uuid.Nil)
}

func (s *server) storyPage(request openapiv1.ListStoriesRequestObject, actor platformauth.Actor) (cursorPage, *failure) {
	teamID := uuid.Nil
	if request.Params.TeamId != nil {
		teamID = *request.Params.TeamId
	}
	return decodeOffsetPage(s.storyCursors, request.Params.Cursor, request.Params.Limit, request.WorkspaceId, actor.PrincipalID, teamID, uuid.Nil)
}

func decodeOffsetPage(
	codec interface {
		Decode(string) (cursorPage, error)
	}, cursor *string, requestedLimit *int,
	workspaceID, principalID, teamID, resourceID uuid.UUID,
) (cursorPage, *failure) {
	limit := defaultPageLimit
	if requestedLimit != nil {
		limit = *requestedLimit
	}
	if limit < 1 || limit > maximumPageLimit {
		return cursorPage{}, invalidCursor()
	}
	page := cursorPage{
		Version: 1, WorkspaceID: workspaceID, PrincipalID: principalID,
		Limit: limit, TeamID: teamID, ResourceID: resourceID,
	}
	if cursor == nil {
		return page, nil
	}
	decoded, err := codec.Decode(*cursor)
	if err != nil || decoded.Version != 1 || decoded.WorkspaceID != workspaceID || decoded.PrincipalID != principalID ||
		decoded.TeamID != teamID || decoded.ResourceID != resourceID || decoded.Offset < 0 || decoded.Offset > maximumPageOffset ||
		decoded.Limit < 1 || decoded.Limit > maximumPageLimit || (requestedLimit != nil && decoded.Limit != limit) {
		return cursorPage{}, invalidCursor()
	}
	return decoded, nil
}

func (s *server) nextOffsetCursor(
	ctx context.Context,
	codec interface {
		Encode(cursorPage) (string, error)
	}, page cursorPage, hasMore bool, nextOffset int,
) (*string, *failure) {
	if !hasMore {
		return nil, nil
	}
	page.Offset = nextOffset
	encoded, err := codec.Encode(page)
	if err != nil {
		s.log.Error(ctx, "failed to encode public API cursor")
		return nil, statusFailure(http.StatusInternalServerError)
	}
	return &encoded, nil
}

func invalidCursor() *failure {
	return &failure{http.StatusBadRequest, "invalid_cursor", "The cursor is invalid or does not match this request."}
}

func commonError(ctx context.Context, problem *failure) (openapiv1.ComponentsCommonErrorResponse, openapiv1.ComponentsCommonBadRequestResponseHeaders) {
	requestID := web.GetRequestID(ctx)
	return errorResponse(ctx, problem), openapiv1.ComponentsCommonBadRequestResponseHeaders{XRequestID: &requestID}
}

func getWorkspaceFailure(ctx context.Context, problem *failure) openapiv1.GetWorkspaceResponseObject {
	body, headers := commonError(ctx, problem)
	switch problem.status {
	case http.StatusBadRequest:
		return openapiv1.GetWorkspace400JSONResponse{ComponentsCommonBadRequestJSONResponse: openapiv1.ComponentsCommonBadRequestJSONResponse{Body: body, Headers: headers}}
	case http.StatusForbidden:
		return openapiv1.GetWorkspace403JSONResponse{ComponentsCommonForbiddenJSONResponse: openapiv1.ComponentsCommonForbiddenJSONResponse{Body: body, Headers: openapiv1.ComponentsCommonForbiddenResponseHeaders(headers)}}
	case http.StatusNotFound:
		return openapiv1.GetWorkspace404JSONResponse{ComponentsCommonNotFoundJSONResponse: openapiv1.ComponentsCommonNotFoundJSONResponse{Body: body, Headers: openapiv1.ComponentsCommonNotFoundResponseHeaders(headers)}}
	default:
		return openapiv1.GetWorkspace500JSONResponse{ComponentsCommonInternalErrorJSONResponse: openapiv1.ComponentsCommonInternalErrorJSONResponse{Body: body, Headers: openapiv1.ComponentsCommonInternalErrorResponseHeaders(headers)}}
	}
}

func listTeamsFailure(ctx context.Context, problem *failure) openapiv1.ListTeamsResponseObject {
	body, headers := commonError(ctx, problem)
	switch problem.status {
	case http.StatusBadRequest:
		return openapiv1.ListTeams400JSONResponse{ComponentsCommonBadRequestJSONResponse: openapiv1.ComponentsCommonBadRequestJSONResponse{Body: body, Headers: headers}}
	case http.StatusForbidden:
		return openapiv1.ListTeams403JSONResponse{ComponentsCommonForbiddenJSONResponse: openapiv1.ComponentsCommonForbiddenJSONResponse{Body: body, Headers: openapiv1.ComponentsCommonForbiddenResponseHeaders(headers)}}
	default:
		return openapiv1.ListTeams500JSONResponse{ComponentsCommonInternalErrorJSONResponse: openapiv1.ComponentsCommonInternalErrorJSONResponse{Body: body, Headers: openapiv1.ComponentsCommonInternalErrorResponseHeaders(headers)}}
	}
}

func listStoriesFailure(ctx context.Context, problem *failure) openapiv1.ListStoriesResponseObject {
	body, headers := commonError(ctx, problem)
	switch problem.status {
	case http.StatusBadRequest:
		return openapiv1.ListStories400JSONResponse{ComponentsCommonBadRequestJSONResponse: openapiv1.ComponentsCommonBadRequestJSONResponse{Body: body, Headers: headers}}
	case http.StatusForbidden:
		return openapiv1.ListStories403JSONResponse{ComponentsCommonForbiddenJSONResponse: openapiv1.ComponentsCommonForbiddenJSONResponse{Body: body, Headers: openapiv1.ComponentsCommonForbiddenResponseHeaders(headers)}}
	default:
		return openapiv1.ListStories500JSONResponse{ComponentsCommonInternalErrorJSONResponse: openapiv1.ComponentsCommonInternalErrorJSONResponse{Body: body, Headers: openapiv1.ComponentsCommonInternalErrorResponseHeaders(headers)}}
	}
}

func getStoryFailure(ctx context.Context, problem *failure) openapiv1.GetStoryResponseObject {
	body, headers := commonError(ctx, problem)
	switch problem.status {
	case http.StatusForbidden:
		return openapiv1.GetStory403JSONResponse{ComponentsCommonForbiddenJSONResponse: openapiv1.ComponentsCommonForbiddenJSONResponse{Body: body, Headers: openapiv1.ComponentsCommonForbiddenResponseHeaders(headers)}}
	case http.StatusNotFound:
		return openapiv1.GetStory404JSONResponse{ComponentsCommonNotFoundJSONResponse: openapiv1.ComponentsCommonNotFoundJSONResponse{Body: body, Headers: openapiv1.ComponentsCommonNotFoundResponseHeaders(headers)}}
	default:
		return openapiv1.GetStory500JSONResponse{ComponentsCommonInternalErrorJSONResponse: openapiv1.ComponentsCommonInternalErrorJSONResponse{Body: body, Headers: openapiv1.ComponentsCommonInternalErrorResponseHeaders(headers)}}
	}
}
