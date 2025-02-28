package invitationsgrp

import (
	"context"
	"errors"
	"net/http"

	"github.com/complexus-tech/projects-api/internal/core/invitations"
	"github.com/complexus-tech/projects-api/internal/core/users"
	"github.com/complexus-tech/projects-api/internal/web/mid"
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

// CreateBulkInvitations creates multiple workspace invitations
func (h *Handlers) CreateBulkInvitations(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "handlers.invitations.CreateBulkInvitations")
	defer span.End()

	workspaceID, err := uuid.Parse(web.Params(r, "workspaceId"))
	if err != nil {
		return web.RespondError(ctx, w, ErrInvalidWorkspaceID, http.StatusBadRequest)
	}

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	var req AppNewInvitationBulk
	if err := web.Decode(r, &req); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	// Convert request to core type
	requests := make([]invitations.InvitationRequest, len(req.Invitations))
	for i, inv := range req.Invitations {
		requests[i] = invitations.InvitationRequest{
			Email:   inv.Email,
			Role:    inv.Role,
			TeamIDs: inv.TeamIDs,
		}
	}

	results, err := h.invitations.CreateBulkInvitations(ctx, workspaceID, userID, requests)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	span.AddEvent("bulk invitations created", trace.WithAttributes(
		attribute.String("workspace_id", workspaceID.String()),
		attribute.Int("invitation_count", len(results)),
	))

	return web.Respond(ctx, w, toAppInvitations(results), http.StatusCreated)
}

// ListInvitations returns all pending invitations for a workspace
func (h *Handlers) ListInvitations(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "handlers.invitations.ListInvitations")
	defer span.End()

	workspaceID, err := uuid.Parse(web.Params(r, "workspaceId"))
	if err != nil {
		return web.RespondError(ctx, w, ErrInvalidWorkspaceID, http.StatusBadRequest)
	}

	invitations, err := h.invitations.ListInvitations(ctx, workspaceID)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	return web.Respond(ctx, w, toAppInvitations(invitations), http.StatusOK)
}

// RevokeInvitation revokes a workspace invitation
func (h *Handlers) RevokeInvitation(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "handlers.invitations.RevokeInvitation")
	defer span.End()

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

// GetInvitation retrieves and validates an invitation by token
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

// ListUserInvitations returns all pending invitations for the authenticated user
func (h *Handlers) ListUserInvitations(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "handlers.invitations.ListUserInvitations")
	defer span.End()

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	// Get user details to get email
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

// AcceptInvitation accepts a workspace invitation
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
