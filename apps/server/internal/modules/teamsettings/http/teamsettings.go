package teamsettingshttp

import (
	"context"
	"errors"
	"net/http"

	teamsettings "github.com/complexus-tech/projects-api/internal/modules/teamsettings/service"
	"github.com/complexus-tech/projects-api/internal/platform/authorization"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/web"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type Handlers struct {
	teamsettings *teamsettings.Service
}

var (
	ErrInvalidTeamID = errors.New("team id is not in its proper form")
)

func New(teamsettings *teamsettings.Service) *Handlers {
	return &Handlers{
		teamsettings: teamsettings,
	}
}

func (h *Handlers) GetSettings(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "handlers.teamsettings.GetSettings")
	defer span.End()

	access, status, err := accessFromRequest(ctx, r)
	if err != nil {
		return web.RespondError(ctx, w, err, status)
	}

	settings, err := h.teamsettings.GetSettings(ctx, access)
	if err != nil {
		return web.RespondError(ctx, w, err, teamSettingsErrorStatus(err))
	}

	span.AddEvent("team settings retrieved.", trace.WithAttributes(
		attribute.String("team_id", access.TeamID.String()),
		attribute.String("workspace_id", access.WorkspaceID.String()),
	))

	return web.Respond(ctx, w, toAppTeamSettings(settings), http.StatusOK)
}

func (h *Handlers) UpdateSprintSettings(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "handlers.teamsettings.UpdateSprintSettings")
	defer span.End()

	access, status, err := accessFromRequest(ctx, r)
	if err != nil {
		return web.RespondError(ctx, w, err, status)
	}

	var input AppUpdateTeamSprintSettings
	if err := web.Decode(r, &input); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	updates := toCoreUpdateTeamSprintSettings(input)
	result, err := h.teamsettings.UpdateSprintSettings(ctx, access, updates)
	if err != nil {
		return web.RespondError(ctx, w, err, teamSettingsErrorStatus(err))
	}

	span.AddEvent("sprint settings updated.", trace.WithAttributes(
		attribute.String("team_id", access.TeamID.String()),
		attribute.String("workspace_id", access.WorkspaceID.String()),
	))

	return web.Respond(ctx, w, toAppTeamSprintSettings(result), http.StatusOK)
}

func (h *Handlers) UpdateStoryAutomationSettings(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "handlers.teamsettings.UpdateStoryAutomationSettings")
	defer span.End()

	access, status, err := accessFromRequest(ctx, r)
	if err != nil {
		return web.RespondError(ctx, w, err, status)
	}

	var input AppUpdateTeamStoryAutomationSettings
	if err := web.Decode(r, &input); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	updates := toCoreUpdateTeamStoryAutomationSettings(input)
	result, err := h.teamsettings.UpdateStoryAutomationSettings(ctx, access, updates)
	if err != nil {
		return web.RespondError(ctx, w, err, teamSettingsErrorStatus(err))
	}

	span.AddEvent("story automation settings updated.", trace.WithAttributes(
		attribute.String("team_id", access.TeamID.String()),
		attribute.String("workspace_id", access.WorkspaceID.String()),
	))

	return web.Respond(ctx, w, toAppTeamStoryAutomationSettings(result), http.StatusOK)
}

func (h *Handlers) UpdateEstimationSettings(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "handlers.teamsettings.UpdateEstimationSettings")
	defer span.End()

	access, status, err := accessFromRequest(ctx, r)
	if err != nil {
		return web.RespondError(ctx, w, err, status)
	}

	var input AppUpdateTeamEstimationSettings
	if err := web.Decode(r, &input); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	updates := toCoreUpdateTeamEstimationSettings(input)
	result, err := h.teamsettings.UpdateEstimationSettings(ctx, access, updates)
	if err != nil {
		return web.RespondError(ctx, w, err, teamSettingsErrorStatus(err))
	}

	return web.Respond(ctx, w, toAppTeamEstimationSettings(result), http.StatusOK)
}

func accessFromRequest(ctx context.Context, r *http.Request) (teamsettings.Access, int, error) {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return teamsettings.Access{}, http.StatusUnauthorized, err
	}
	actor, err := mid.GetActor(ctx)
	if err != nil {
		return teamsettings.Access{}, http.StatusUnauthorized, err
	}
	teamID, err := web.UUIDPathParameter(r, "teamId")
	if err != nil {
		return teamsettings.Access{}, http.StatusBadRequest, ErrInvalidTeamID
	}
	return teamsettings.Access{
		Actor:         actor,
		WorkspaceRole: authorization.WorkspaceRole(workspace.UserRole),
		WorkspaceID:   workspace.ID,
		TeamID:        teamID,
	}, 0, nil
}

func teamSettingsErrorStatus(err error) int {
	switch {
	case errors.Is(err, teamsettings.ErrInvalidSprintStartDay):
		return http.StatusBadRequest
	case errors.Is(err, teamsettings.ErrInvalidSprintDuration):
		return http.StatusBadRequest
	case errors.Is(err, teamsettings.ErrInvalidWorkingDays):
		return http.StatusBadRequest
	case errors.Is(err, teamsettings.ErrInvalidUpcomingCount):
		return http.StatusBadRequest
	case errors.Is(err, teamsettings.ErrInvalidNextAutoNumber):
		return http.StatusBadRequest
	case errors.Is(err, teamsettings.ErrInvalidCloseMonths):
		return http.StatusBadRequest
	case errors.Is(err, teamsettings.ErrInvalidArchiveMonths):
		return http.StatusBadRequest
	case errors.Is(err, teamsettings.ErrInvalidEstimateScheme):
		return http.StatusBadRequest
	case errors.Is(err, teamsettings.ErrNoSettingsChanges):
		return http.StatusBadRequest
	case errors.Is(err, teamsettings.ErrSprintScheduleConflict):
		return http.StatusConflict
	case errors.Is(err, teamsettings.ErrConcurrentUpdate):
		return http.StatusConflict
	case errors.Is(err, teamsettings.ErrTeamSettingsNotFound):
		return http.StatusNotFound
	case errors.Is(err, teamsettings.ErrTeamMembershipRequired):
		return http.StatusForbidden
	case errors.Is(err, authorization.ErrWorkspaceAdminRequired):
		return http.StatusForbidden
	case errors.Is(err, authorization.ErrInsufficientWorkspaceRole):
		return http.StatusForbidden
	case errors.Is(err, authorization.ErrPrincipalKindDenied):
		return http.StatusForbidden
	case errors.Is(err, authorization.ErrWorkspaceMismatch):
		return http.StatusForbidden
	case errors.Is(err, authorization.ErrCredentialScopeDenied):
		return http.StatusForbidden
	case errors.Is(err, authorization.ErrTeamRestrictionDenied):
		return http.StatusForbidden
	default:
		return http.StatusInternalServerError
	}
}
