package teamshttp

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net/http"
	"net/url"

	teams "github.com/complexus-tech/projects-api/internal/modules/teams/service"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/internal/platform/pagination"
	"github.com/complexus-tech/projects-api/pkg/cache"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

var (
	ErrInvalidWorkspaceID = errors.New("workspace id is not in its proper form")
	ErrInvalidTeamID      = errors.New("team id is not in its proper form")
)

type Handlers struct {
	teams *teams.Service
	cache *cache.Service
}

func parseListTeamsQuery(values url.Values) (teams.CoreListTeamsFilter, *pagination.OffsetParams, error) {
	search, _, err := web.OptionalTextQueryParameter(
		values, "search", web.DefaultMaxQueryParameterBytes, web.DefaultMaxQueryParameterBytes,
	)
	if err != nil {
		return teams.CoreListTeamsFilter{}, nil, err
	}
	joinedOnly, _, err := web.OptionalBooleanQueryParameter(values, "joinedOnly")
	if err != nil {
		return teams.CoreListTeamsFilter{}, nil, err
	}
	filter := teams.CoreListTeamsFilter{
		Search:     search,
		JoinedOnly: joinedOnly,
	}
	if !pagination.OffsetRequested(values) {
		return filter, nil, nil
	}
	params, err := pagination.ParseOffsetQuery(values, pagination.OffsetQueryConfig{
		DefaultPageSize: pagination.DefaultMenuPageSize,
		MaximumPageSize: pagination.MaximumPageSize,
		MaximumOffset:   math.MaxInt32,
	})
	if err != nil {
		return teams.CoreListTeamsFilter{}, nil, err
	}
	return filter, &params, nil
}

func New(teams *teams.Service, cacheService *cache.Service) *Handlers {
	return &Handlers{
		teams: teams,
		cache: cacheService,
	}
}

func (h *Handlers) List(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	filter, params, err := parseListTeamsQuery(r.URL.Query())
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	if params != nil {
		page, pageSize := params.Page, params.PageSize
		filter.Limit = pageSize + 1
		filter.Offset = params.Offset()

		teams, err := h.teams.List(ctx, workspace.ID, userID, filter)
		if err != nil {
			return web.RespondError(ctx, w, err, http.StatusInternalServerError)
		}

		hasMore := len(teams) > pageSize
		if hasMore {
			teams = teams[:pageSize]
		}

		web.Respond(ctx, w, toAppTeamsResponse(teams, page, pageSize, hasMore), http.StatusOK)
		return nil
	}

	teams, err := h.teams.List(ctx, workspace.ID, userID, filter)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}
	web.Respond(ctx, w, toAppTeams(teams), http.StatusOK)
	return nil
}

func (h *Handlers) GetByID(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "handlers.teams.GetByID")
	defer span.End()

	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	teamIDParam := web.Params(r, "id")
	teamID, err := uuid.Parse(teamIDParam)
	if err != nil {
		return web.RespondError(ctx, w, ErrInvalidTeamID, http.StatusBadRequest)
	}

	team, err := h.teams.GetByID(ctx, teamID, workspace.ID, userID)
	if err != nil {
		if err.Error() == "team not found" {
			return web.RespondError(ctx, w, err, http.StatusNotFound)
		}
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	span.AddEvent("team retrieved.", trace.WithAttributes(
		attribute.String("team_id", teamID.String()),
		attribute.String("workspace_id", workspace.ID.String()),
		attribute.String("user_id", userID.String()),
	))

	return web.Respond(ctx, w, toAppTeams([]teams.CoreTeam{team})[0], http.StatusOK)
}

func (h *Handlers) ListPublicTeams(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	filter, params, err := parseListTeamsQuery(r.URL.Query())
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	if params != nil {
		page, pageSize := params.Page, params.PageSize
		filter.Limit = pageSize + 1
		filter.Offset = params.Offset()

		teams, err := h.teams.ListPublicTeams(ctx, workspace.ID, userID, filter)
		if err != nil {
			return web.RespondError(ctx, w, err, http.StatusInternalServerError)
		}

		hasMore := len(teams) > pageSize
		if hasMore {
			teams = teams[:pageSize]
		}

		web.Respond(ctx, w, toAppTeamsResponse(teams, page, pageSize, hasMore), http.StatusOK)
		return nil
	}

	teams, err := h.teams.ListPublicTeams(ctx, workspace.ID, userID, filter)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	web.Respond(ctx, w, toAppTeams(teams), http.StatusOK)
	return nil
}

func (h *Handlers) Create(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "handlers.teams.Create")
	defer span.End()

	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	var input AppNewTeam
	if err := web.Decode(r, &input); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	team := teams.CoreTeam{
		Name:      input.Name,
		Code:      input.Code,
		Color:     input.Color,
		IsPrivate: input.IsPrivate,
		Workspace: workspace.ID,
	}

	result, err := h.teams.Create(ctx, team)
	if err != nil {
		if err == teams.ErrTeamCodeExists {
			return web.RespondError(ctx, w, err, http.StatusConflict)
		}
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	if err := h.teams.AddMember(ctx, result.ID, userID, workspace.ID); err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	span.AddEvent("team created.", trace.WithAttributes(
		attribute.String("team_id", result.ID.String()),
		attribute.String("workspace_id", workspace.ID.String()),
	))

	return web.Respond(ctx, w, toAppTeams([]teams.CoreTeam{result})[0], http.StatusCreated)
}

func (h *Handlers) Update(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "handlers.teams.Update")
	defer span.End()

	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	teamIDParam := web.Params(r, "id")
	teamID, err := uuid.Parse(teamIDParam)
	if err != nil {
		return web.RespondError(ctx, w, ErrInvalidTeamID, http.StatusBadRequest)
	}

	var input AppUpdateTeam
	if err := web.Decode(r, &input); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	team := teams.CoreTeam{
		Name:      input.Name,
		Code:      input.Code,
		Color:     input.Color,
		Workspace: workspace.ID,
	}

	if input.IsPrivate != nil {
		team.IsPrivate = *input.IsPrivate
	}

	result, err := h.teams.Update(ctx, teamID, team)
	if err != nil {
		if err == teams.ErrTeamCodeExists {
			return web.RespondError(ctx, w, err, http.StatusConflict)
		}
		if errors.Is(err, teams.ErrTeamNotFound) {
			return web.RespondError(ctx, w, err, http.StatusNotFound)
		}
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	span.AddEvent("team updated.", trace.WithAttributes(
		attribute.String("team_id", teamID.String()),
		attribute.String("workspace_id", workspace.ID.String()),
	))

	return web.Respond(ctx, w, toAppTeams([]teams.CoreTeam{result})[0], http.StatusOK)
}

func (h *Handlers) Delete(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "handlers.teams.Delete")
	defer span.End()

	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	teamIDParam := web.Params(r, "id")
	teamID, err := uuid.Parse(teamIDParam)
	if err != nil {
		return web.RespondError(ctx, w, ErrInvalidTeamID, http.StatusBadRequest)
	}

	if err := h.teams.Delete(ctx, teamID, workspace.ID); err != nil {
		if errors.Is(err, teams.ErrTeamNotFound) {
			return web.RespondError(ctx, w, err, http.StatusNotFound)
		}
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	span.AddEvent("team deleted.", trace.WithAttributes(
		attribute.String("team_id", teamID.String()),
		attribute.String("workspace_id", workspace.ID.String()),
	))

	return web.Respond(ctx, w, nil, http.StatusNoContent)
}

func (h *Handlers) AddMember(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "handlers.teams.AddMember")
	defer span.End()

	var input AppNewTeamMember
	if err := web.Decode(r, &input); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	teamIDParam := web.Params(r, "id")
	teamID, err := uuid.Parse(teamIDParam)
	if err != nil {
		return web.RespondError(ctx, w, ErrInvalidTeamID, http.StatusBadRequest)
	}

	if err := h.teams.AddMember(ctx, teamID, input.UserID, workspace.ID); err != nil {
		if errors.Is(err, teams.ErrTeamMemberExists) {
			return web.RespondError(ctx, w, err, http.StatusConflict)
		}
		if errors.Is(err, teams.ErrTeamNotFound) {
			return web.RespondError(ctx, w, teams.ErrTeamNotFound, http.StatusNotFound)
		}
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	myStoriesCachePattern := fmt.Sprintf(cache.MyStoriesKey+"*", workspace.ID.String())
	if err := h.cache.DeleteByPattern(ctx, myStoriesCachePattern); err != nil {
		span.RecordError(fmt.Errorf("failed to invalidate my-stories cache: %w", err))
	}

	span.AddEvent("team member added.", trace.WithAttributes(
		attribute.String("team_id", teamID.String()),
		attribute.String("workspace_id", workspace.ID.String()),
		attribute.String("user_id", input.UserID.String()),
	))

	team := struct {
		ID uuid.UUID `json:"teamId"`
	}{
		ID: teamID,
	}

	return web.Respond(ctx, w, team, http.StatusOK)
}

// JoinPublicTeam adds only the authenticated actor to an eligible public team.
func (h *Handlers) JoinPublicTeam(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "handlers.teams.JoinPublicTeam")
	defer span.End()

	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	actorID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	teamID, err := uuid.Parse(web.Params(r, "id"))
	if err != nil {
		return web.RespondError(ctx, w, ErrInvalidTeamID, http.StatusBadRequest)
	}

	input := teams.CorePublicTeamJoin{
		TeamID:      teamID,
		ActorID:     actorID,
		WorkspaceID: workspace.ID,
	}
	if err := h.teams.JoinPublicTeam(ctx, input); err != nil {
		switch {
		case errors.Is(err, teams.ErrTeamNotFound):
			return web.RespondError(ctx, w, teams.ErrTeamNotFound, http.StatusNotFound)
		case errors.Is(err, teams.ErrTeamMemberExists):
			return web.RespondError(ctx, w, teams.ErrTeamMemberExists, http.StatusConflict)
		default:
			return web.RespondError(ctx, w, err, http.StatusInternalServerError)
		}
	}

	myStoriesCachePattern := fmt.Sprintf(cache.MyStoriesKey+"*", workspace.ID.String())
	if err := h.cache.DeleteByPattern(ctx, myStoriesCachePattern); err != nil {
		span.RecordError(fmt.Errorf("failed to invalidate my-stories cache: %w", err))
	}

	span.AddEvent("public team joined.", trace.WithAttributes(
		attribute.String("team_id", teamID.String()),
		attribute.String("workspace_id", workspace.ID.String()),
		attribute.String("actor_id", actorID.String()),
	))

	return web.Respond(ctx, w, struct {
		ID uuid.UUID `json:"teamId"`
	}{ID: teamID}, http.StatusOK)
}

func (h *Handlers) RemoveMember(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "handlers.teams.RemoveMember")
	defer span.End()

	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	teamIDParam := web.Params(r, "id")
	teamID, err := uuid.Parse(teamIDParam)
	if err != nil {
		return web.RespondError(ctx, w, ErrInvalidTeamID, http.StatusBadRequest)
	}

	userIDParam := web.Params(r, "userId")
	userID, err := uuid.Parse(userIDParam)
	if err != nil {
		return web.RespondError(ctx, w, errors.New("invalid user id"), http.StatusBadRequest)
	}

	if err := h.teams.RemoveMember(ctx, teamID, userID, workspace.ID); err != nil {
		return web.RespondError(ctx, w, err, teamMembershipMutationStatus(err))
	}
	myStoriesCachePattern := fmt.Sprintf(cache.MyStoriesKey+"*", workspace.ID.String())
	if err := h.cache.DeleteByPattern(ctx, myStoriesCachePattern); err != nil {
		span.RecordError(fmt.Errorf("failed to invalidate my-stories cache: %w", err))
	}

	span.AddEvent("team member removed.", trace.WithAttributes(
		attribute.String("team_id", teamID.String()),
		attribute.String("workspace_id", workspace.ID.String()),
		attribute.String("user_id", userID.String()),
	))

	return web.Respond(ctx, w, nil, http.StatusNoContent)
}

// LeaveTeam removes the authenticated actor from a team without accepting a caller-selected user ID.
func (h *Handlers) LeaveTeam(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "handlers.teams.LeaveTeam")
	defer span.End()

	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	actorID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	teamID, err := uuid.Parse(web.Params(r, "id"))
	if err != nil {
		return web.RespondError(ctx, w, ErrInvalidTeamID, http.StatusBadRequest)
	}

	input := teams.CoreTeamSelfLeave{
		TeamID:      teamID,
		ActorID:     actorID,
		WorkspaceID: workspace.ID,
	}
	if err := h.teams.LeaveTeam(ctx, input); err != nil {
		return web.RespondError(ctx, w, err, teamMembershipMutationStatus(err))
	}

	myStoriesCachePattern := fmt.Sprintf(cache.MyStoriesKey+"*", workspace.ID.String())
	if err := h.cache.DeleteByPattern(ctx, myStoriesCachePattern); err != nil {
		span.RecordError(fmt.Errorf("failed to invalidate my-stories cache: %w", err))
	}

	span.AddEvent("team left.", trace.WithAttributes(
		attribute.String("team_id", teamID.String()),
		attribute.String("workspace_id", workspace.ID.String()),
		attribute.String("actor_id", actorID.String()),
	))

	return web.Respond(ctx, w, nil, http.StatusNoContent)
}

func teamMembershipMutationStatus(err error) int {
	if errors.Is(err, teams.ErrTeamNotFound) || errors.Is(err, teams.ErrTeamMemberNotFound) {
		return http.StatusNotFound
	}
	return http.StatusInternalServerError
}

func (h *Handlers) UpdateMemberAIContext(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "handlers.teams.UpdateMemberAIContext")
	defer span.End()

	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	teamIDParam := web.Params(r, "id")
	teamID, err := uuid.Parse(teamIDParam)
	if err != nil {
		return web.RespondError(ctx, w, ErrInvalidTeamID, http.StatusBadRequest)
	}

	userIDParam := web.Params(r, "userId")
	userID, err := uuid.Parse(userIDParam)
	if err != nil {
		return web.RespondError(ctx, w, errors.New("invalid user id"), http.StatusBadRequest)
	}

	var input AppUpdateTeamMemberAIContext
	if err := web.Decode(r, &input); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	if err := h.teams.UpdateMemberAIContext(ctx, teamID, userID, workspace.ID, teams.CoreTeamMemberAIContext{
		RoleTitle:       input.RoleTitle,
		RoleDescription: input.RoleDescription,
	}); err != nil {
		if err.Error() == "team member not found" {
			return web.RespondError(ctx, w, err, http.StatusNotFound)
		}
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	span.AddEvent("team member ai context updated.", trace.WithAttributes(
		attribute.String("team_id", teamID.String()),
		attribute.String("workspace_id", workspace.ID.String()),
		attribute.String("user_id", userID.String()),
	))

	return web.Respond(ctx, w, nil, http.StatusNoContent)
}

// UpdateTeamOrdering updates the user's custom team ordering for a workspace.
func (h *Handlers) UpdateTeamOrdering(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "handlers.teams.UpdateTeamOrdering")
	defer span.End()

	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	var input AppUpdateTeamOrdering
	if err := web.Decode(r, &input); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	// Validate that the user has access to all the teams being ordered
	// This is a basic validation - you might want to add more sophisticated checks
	if len(input.TeamIDs) == 0 {
		return web.RespondError(ctx, w, errors.New("team IDs cannot be empty"), http.StatusBadRequest)
	}

	if err := h.teams.UpdateUserTeamOrdering(ctx, userID, workspace.ID, input.TeamIDs); err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	span.AddEvent("team ordering updated.", trace.WithAttributes(
		attribute.String("workspace_id", workspace.ID.String()),
		attribute.String("user_id", userID.String()),
		attribute.Int("teams_ordered", len(input.TeamIDs)),
	))

	return web.Respond(ctx, w, nil, http.StatusNoContent)
}
