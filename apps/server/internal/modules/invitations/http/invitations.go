package invitationshttp

import (
	"context"
	"errors"
	"net/http"

	invitations "github.com/complexus-tech/projects-api/internal/modules/invitations/service"
	users "github.com/complexus-tech/projects-api/internal/modules/users/service"
	"github.com/complexus-tech/projects-api/internal/platform/authorization"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

var (
	ErrInvalidWorkspaceID  = errors.New("workspace id is not in its proper form")
	ErrInvalidInvitationID = errors.New("invitation id is not in its proper form")
)

type Handlers struct {
	invitations *invitations.Service
	users       *users.Service
}

func New(invitations *invitations.Service, users *users.Service) *Handlers {
	return &Handlers{
		invitations: invitations,
		users:       users,
	}
}

func (h *Handlers) CreateBulkInvitations(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "handlers.invitations.CreateBulkInvitations")
	defer span.End()

	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	var req AppNewInvitationBulk
	if err := web.Decode(r, &req); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	requests := make([]invitations.InvitationRequest, len(req.Invitations))
	for i, inv := range req.Invitations {
		requests[i] = invitations.InvitationRequest{
			Email:   inv.Email,
			Role:    inv.Role,
			TeamIDs: inv.TeamIDs,
		}
	}

	results, err := h.invitations.CreateBulkInvitations(ctx, workspace.ID, userID, requests)
	if err != nil {
		return web.RespondError(ctx, w, err, invitationAdminErrorStatus(err, http.StatusInternalServerError))
	}

	span.AddEvent("bulk invitations created", trace.WithAttributes(
		attribute.String("workspace_id", workspace.ID.String()),
		attribute.Int("invitation_count", len(results)),
	))

	return web.Respond(ctx, w, toAppInvitations(results), http.StatusCreated)
}

func (h *Handlers) ListInvitations(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "handlers.invitations.ListInvitations")
	defer span.End()

	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	invitations, err := h.invitations.ListInvitations(ctx, workspace.ID, userID)
	if err != nil {
		return web.RespondError(ctx, w, err, invitationAdminErrorStatus(err, http.StatusInternalServerError))
	}

	return web.Respond(ctx, w, toAppInvitations(invitations), http.StatusOK)
}

func (h *Handlers) RevokeInvitation(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "handlers.invitations.RevokeInvitation")
	defer span.End()

	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	invitationID, err := uuid.Parse(web.Params(r, "id"))
	if err != nil {
		return web.RespondError(ctx, w, ErrInvalidInvitationID, http.StatusBadRequest)
	}

	if err := h.invitations.RevokeInvitation(ctx, workspace.ID, userID, invitationID); err != nil {
		return web.RespondError(ctx, w, err, invitationAdminErrorStatus(err, http.StatusInternalServerError))
	}

	span.AddEvent("invitation revoked", trace.WithAttributes(
		attribute.String("invitation_id", invitationID.String()),
	))

	return web.Respond(ctx, w, nil, http.StatusNoContent)
}

func invitationAdminErrorStatus(err error, fallback int) int {
	switch {
	case errors.Is(err, authorization.ErrWorkspaceAdminRequired):
		return http.StatusForbidden
	case errors.Is(err, invitations.ErrInvalidInvitationRole):
		return http.StatusBadRequest
	case errors.Is(err, invitations.ErrInvalidInvitationEmail):
		return http.StatusBadRequest
	case errors.Is(err, invitations.ErrInvalidInvitationTeam):
		return http.StatusBadRequest
	case errors.Is(err, invitations.ErrTooManyInvitations):
		return http.StatusBadRequest
	case errors.Is(err, invitations.ErrDuplicateInvitation):
		return http.StatusConflict
	case errors.Is(err, invitations.ErrInvitationNotFound):
		return http.StatusNotFound
	default:
		return fallback
	}
}

func (h *Handlers) GetInvitation(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "handlers.invitations.GetInvitation")
	defer span.End()

	token := web.Params(r, "token")
	if token == "" {
		return web.RespondError(ctx, w, errors.New("token is required"), http.StatusBadRequest)
	}

	invitation, err := h.invitations.GetInvitation(ctx, token)
	if err != nil {
		if status, handled, publicError := publicInvitationBearerError(err); handled {
			return web.RespondError(ctx, w, publicError, status)
		}
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	return web.Respond(ctx, w, toAppInvitation(invitation), http.StatusOK)
}

func (h *Handlers) ListUserInvitations(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "handlers.invitations.ListUserInvitations")
	defer span.End()

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	user, err := h.users.GetUser(ctx, userID)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	invitations, err := h.invitations.ListUserInvitations(ctx, user.Email)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	return web.Respond(ctx, w, toAppInvitations(invitations), http.StatusOK)
}

func (h *Handlers) AcceptInvitation(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "handlers.invitations.AcceptInvitation")
	defer span.End()

	token := web.Params(r, "token")
	if token == "" {
		return web.RespondError(ctx, w, errors.New("token is required"), http.StatusBadRequest)
	}

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	if err := h.invitations.AcceptInvitation(ctx, token, userID); err != nil {
		if status, handled, publicError := publicInvitationBearerError(err); handled {
			return web.RespondError(ctx, w, publicError, status)
		}
		switch {
		case errors.Is(err, invitations.ErrInvalidInvitee):
			return web.RespondError(ctx, w, invitations.ErrInvitationNotFound, http.StatusNotFound)
		case errors.Is(err, invitations.ErrAlreadyWorkspaceMember):
			return web.RespondError(ctx, w, err, http.StatusConflict)
		default:
			return web.RespondError(ctx, w, err, http.StatusInternalServerError)
		}
	}

	return web.Respond(ctx, w, nil, http.StatusOK)
}

// publicInvitationBearerError deliberately collapses token lifecycle state so
// callers cannot distinguish an unknown bearer from a known expired, revoked,
// or already-used bearer.
func publicInvitationBearerError(err error) (int, bool, error) {
	switch {
	case errors.Is(err, invitations.ErrInvitationNotFound),
		errors.Is(err, invitations.ErrInvitationExpired),
		errors.Is(err, invitations.ErrInvitationUsed),
		errors.Is(err, invitations.ErrInvitationRevoked):
		return http.StatusNotFound, true, invitations.ErrInvitationNotFound
	default:
		return 0, false, nil
	}
}
