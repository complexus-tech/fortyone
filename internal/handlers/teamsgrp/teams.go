package teamsgrp

import (
	"context"
	"errors"
	"net/http"

	"github.com/complexus-tech/projects-api/internal/core/teams"
	"github.com/complexus-tech/projects-api/internal/web/mid"
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
	// audit  *audit.Service
}

// New constructs a new teams handlers instance.
func New(teams *teams.Service) *Handlers {
	return &Handlers{
		teams: teams,
	}
}

// List returns a list of teams.
func (h *Handlers) List(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspaceIdParam := web.Params(r, "workspaceId")
	workspaceId, err := uuid.Parse(workspaceIdParam)
	if err != nil {
		web.RespondError(ctx, w, ErrInvalidWorkspaceID, http.StatusBadRequest)
		return nil
	}
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	teams, err := h.teams.List(ctx, workspaceId, userID)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}
	web.Respond(ctx, w, toAppTeams(teams), http.StatusOK)
	return nil
}

func (h *Handlers) Create(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "handlers.teams.Create")
	defer span.End()

	workspaceIDParam := web.Params(r, "workspaceId")
	workspaceID, err := uuid.Parse(workspaceIDParam)
	if err != nil {
		return web.RespondError(ctx, w, ErrInvalidWorkspaceID, http.StatusBadRequest)
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
		Workspace: workspaceID,
	}

	result, err := h.teams.Create(ctx, team)
	if err != nil {
		if err == teams.ErrTeamCodeExists {
			return web.RespondError(ctx, w, err, http.StatusConflict)
		}
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	// Add creator as member
	if err := h.teams.AddMember(ctx, result.ID, userID, "admin"); err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	span.AddEvent("team created.", trace.WithAttributes(
		attribute.String("team_id", result.ID.String()),
		attribute.String("workspace_id", workspaceID.String()),
	))

	return web.Respond(ctx, w, toAppTeams([]teams.CoreTeam{result})[0], http.StatusCreated)
}

func (h *Handlers) Update(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "handlers.teams.Update")
	defer span.End()

	workspaceIDParam := web.Params(r, "workspaceId")
	workspaceID, err := uuid.Parse(workspaceIDParam)
	if err != nil {
		return web.RespondError(ctx, w, ErrInvalidWorkspaceID, http.StatusBadRequest)
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
		Workspace: workspaceID,
	}

	if input.IsPrivate != nil {
		team.IsPrivate = *input.IsPrivate
	}

	result, err := h.teams.Update(ctx, teamID, team)
	if err != nil {
		if err == teams.ErrTeamCodeExists {
			return web.RespondError(ctx, w, err, http.StatusConflict)
		}
		if err.Error() == "team not found" {
			return web.RespondError(ctx, w, err, http.StatusNotFound)
		}
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	span.AddEvent("team updated.", trace.WithAttributes(
		attribute.String("team_id", teamID.String()),
		attribute.String("workspace_id", workspaceID.String()),
	))

	return web.Respond(ctx, w, toAppTeams([]teams.CoreTeam{result})[0], http.StatusOK)
}

func (h *Handlers) Delete(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "handlers.teams.Delete")
	defer span.End()

	workspaceIDParam := web.Params(r, "workspaceId")
	workspaceID, err := uuid.Parse(workspaceIDParam)
	if err != nil {
		return web.RespondError(ctx, w, ErrInvalidWorkspaceID, http.StatusBadRequest)
	}

	teamIDParam := web.Params(r, "id")
	teamID, err := uuid.Parse(teamIDParam)
	if err != nil {
		return web.RespondError(ctx, w, ErrInvalidTeamID, http.StatusBadRequest)
	}

	if err := h.teams.Delete(ctx, teamID, workspaceID); err != nil {
		if err.Error() == "team not found" {
			return web.RespondError(ctx, w, err, http.StatusNotFound)
		}
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	span.AddEvent("team deleted.", trace.WithAttributes(
		attribute.String("team_id", teamID.String()),
		attribute.String("workspace_id", workspaceID.String()),
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

	workspaceIDParam := web.Params(r, "workspaceId")
	workspaceID, err := uuid.Parse(workspaceIDParam)
	if err != nil {
		return web.RespondError(ctx, w, ErrInvalidWorkspaceID, http.StatusBadRequest)
	}

	teamIDParam := web.Params(r, "id")
	teamID, err := uuid.Parse(teamIDParam)
	if err != nil {
		return web.RespondError(ctx, w, ErrInvalidTeamID, http.StatusBadRequest)
	}

	// Default to member role if not provided
	role := input.Role
	if role == "" {
		role = "member"
	}

	if err := h.teams.AddMember(ctx, teamID, input.UserID, role); err != nil {
		if err == teams.ErrTeamMemberExists {
			return web.RespondError(ctx, w, err, http.StatusConflict)
		}
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	span.AddEvent("team member added.", trace.WithAttributes(
		attribute.String("team_id", teamID.String()),
		attribute.String("workspace_id", workspaceID.String()),
		attribute.String("user_id", input.UserID.String()),
		attribute.String("role", role),
	))

	// TODO: Send notification to user

	team := struct {
		ID uuid.UUID `json:"teamId"`
	}{
		ID: teamID,
	}

	return web.Respond(ctx, w, team, http.StatusNoContent)
}

func (h *Handlers) RemoveMember(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "handlers.teams.RemoveMember")
	defer span.End()

	workspaceIDParam := web.Params(r, "workspaceId")
	workspaceID, err := uuid.Parse(workspaceIDParam)
	if err != nil {
		return web.RespondError(ctx, w, ErrInvalidWorkspaceID, http.StatusBadRequest)
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

	if err := h.teams.RemoveMember(ctx, teamID, userID, workspaceID); err != nil {
		if err.Error() == "member not found" {
			return web.RespondError(ctx, w, err, http.StatusNotFound)
		}
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	span.AddEvent("team member removed.", trace.WithAttributes(
		attribute.String("team_id", teamID.String()),
		attribute.String("workspace_id", workspaceID.String()),
		attribute.String("user_id", userID.String()),
	))

	return web.Respond(ctx, w, nil, http.StatusNoContent)
}
