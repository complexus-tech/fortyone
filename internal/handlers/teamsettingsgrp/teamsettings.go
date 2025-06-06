package teamsettingsgrp

import (
	"context"
	"errors"
	"net/http"

	"github.com/complexus-tech/projects-api/internal/core/teamsettings"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type Handlers struct {
	teamsettings *teamsettings.Service
}

var (
	ErrInvalidWorkspaceID = errors.New("workspace id is not in its proper form")
	ErrInvalidTeamID      = errors.New("team id is not in its proper form")
)

// New constructs a new team settings handlers instance.
func New(teamsettings *teamsettings.Service) *Handlers {
	return &Handlers{
		teamsettings: teamsettings,
	}
}

// GetSettings returns the complete team settings.
func (h *Handlers) GetSettings(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "handlers.teamsettings.GetSettings")
	defer span.End()

	workspaceIDParam := web.Params(r, "workspaceId")
	workspaceID, err := uuid.Parse(workspaceIDParam)
	if err != nil {
		return web.RespondError(ctx, w, ErrInvalidWorkspaceID, http.StatusBadRequest)
	}

	teamIDParam := web.Params(r, "teamId")
	teamID, err := uuid.Parse(teamIDParam)
	if err != nil {
		return web.RespondError(ctx, w, ErrInvalidTeamID, http.StatusBadRequest)
	}

	settings, err := h.teamsettings.GetSettings(ctx, teamID, workspaceID)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	span.AddEvent("team settings retrieved.", trace.WithAttributes(
		attribute.String("team_id", teamID.String()),
		attribute.String("workspace_id", workspaceID.String()),
	))

	return web.Respond(ctx, w, toAppTeamSettings(settings), http.StatusOK)
}

// UpdateSprintSettings updates the sprint settings for a team.
func (h *Handlers) UpdateSprintSettings(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "handlers.teamsettings.UpdateSprintSettings")
	defer span.End()

	workspaceIDParam := web.Params(r, "workspaceId")
	workspaceID, err := uuid.Parse(workspaceIDParam)
	if err != nil {
		return web.RespondError(ctx, w, ErrInvalidWorkspaceID, http.StatusBadRequest)
	}

	teamIDParam := web.Params(r, "teamId")
	teamID, err := uuid.Parse(teamIDParam)
	if err != nil {
		return web.RespondError(ctx, w, ErrInvalidTeamID, http.StatusBadRequest)
	}

	var input AppUpdateTeamSprintSettings
	if err := web.Decode(r, &input); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	updates := toCoreUpdateTeamSprintSettings(input)
	result, err := h.teamsettings.UpdateSprintSettings(ctx, teamID, workspaceID, updates)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	span.AddEvent("sprint settings updated.", trace.WithAttributes(
		attribute.String("team_id", teamID.String()),
		attribute.String("workspace_id", workspaceID.String()),
	))

	return web.Respond(ctx, w, toAppTeamSprintSettings(result), http.StatusOK)
}

// UpdateStoryAutomationSettings updates the story automation settings for a team.
func (h *Handlers) UpdateStoryAutomationSettings(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "handlers.teamsettings.UpdateStoryAutomationSettings")
	defer span.End()

	workspaceIDParam := web.Params(r, "workspaceId")
	workspaceID, err := uuid.Parse(workspaceIDParam)
	if err != nil {
		return web.RespondError(ctx, w, ErrInvalidWorkspaceID, http.StatusBadRequest)
	}

	teamIDParam := web.Params(r, "teamId")
	teamID, err := uuid.Parse(teamIDParam)
	if err != nil {
		return web.RespondError(ctx, w, ErrInvalidTeamID, http.StatusBadRequest)
	}

	var input AppUpdateTeamStoryAutomationSettings
	if err := web.Decode(r, &input); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	updates := toCoreUpdateTeamStoryAutomationSettings(input)
	result, err := h.teamsettings.UpdateStoryAutomationSettings(ctx, teamID, workspaceID, updates)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	span.AddEvent("story automation settings updated.", trace.WithAttributes(
		attribute.String("team_id", teamID.String()),
		attribute.String("workspace_id", workspaceID.String()),
	))

	return web.Respond(ctx, w, toAppTeamStoryAutomationSettings(result), http.StatusOK)
}
