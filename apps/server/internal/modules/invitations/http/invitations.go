package invitationshttp

import (
	"context"
	"errors"
	"net/http"

	invitations "github.com/complexus-tech/projects-api/internal/modules/invitations/service"
	users "github.com/complexus-tech/projects-api/internal/modules/users/service"
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
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
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

	invitations, err := h.invitations.ListInvitations(ctx, workspace.ID)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	return web.Respond(ctx, w, toAppInvitations(invitations), http.StatusOK)
}

func (h *Handlers) RevokeInvitation(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "handlers.invitations.RevokeInvitation")
	defer span.End()

	// Workspace is in context for authorization, but not directly used here.
	_, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	invitationID, err := uuid.Parse(web.Params(r, "id"))
	if err != nil {
		return web.RespondError(ctx, w, ErrInvalidInvitationID, http.StatusBadRequest)
	}

	if err := h.invitations.RevokeInvitation(ctx, invitationID); err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	span.AddEvent("invitation revoked", trace.WithAttributes(
		attribute.String("invitation_id", invitationID.String()),
	))

	return web.Respond(ctx, w, nil, http.StatusNoContent)
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
		switch err {
		case invitations.ErrInvitationNotFound:
			return web.RespondError(ctx, w, err, http.StatusNotFound)
		case invitations.ErrInvitationExpired:
			return web.RespondError(ctx, w, err, http.StatusGone)
		case invitations.ErrInvitationUsed:
			return web.RespondError(ctx, w, err, http.StatusGone)
		default:
			return web.RespondError(ctx, w, err, http.StatusInternalServerError)
		}
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
		switch err {
		case invitations.ErrInvitationNotFound:
			return web.RespondError(ctx, w, err, http.StatusNotFound)
		case invitations.ErrInvitationExpired:
			return web.RespondError(ctx, w, err, http.StatusGone)
		case invitations.ErrInvitationUsed:
			return web.RespondError(ctx, w, err, http.StatusGone)
		case invitations.ErrInvalidInvitee:
			return web.RespondError(ctx, w, err, http.StatusForbidden)
		case invitations.ErrAlreadyWorkspaceMember:
			return web.RespondError(ctx, w, err, http.StatusConflict)
		default:
			return web.RespondError(ctx, w, err, http.StatusInternalServerError)
		}
	}

	return web.Respond(ctx, w, nil, http.StatusOK)
}
