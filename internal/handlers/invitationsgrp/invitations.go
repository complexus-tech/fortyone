package invitationsgrp

import (
	"context"
	"errors"
	"net/http"

	"github.com/complexus-tech/projects-api/internal/core/invitations"
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
}

func New(invitations *invitations.Service) *Handlers {
	return &Handlers{
		invitations: invitations,
	}
}

// CreateInvitation creates a new workspace invitation
func (h *Handlers) CreateInvitation(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "handlers.invitations.CreateInvitation")
	defer span.End()

	workspaceID, err := uuid.Parse(web.Params(r, "workspaceId"))
	if err != nil {
		return web.RespondError(ctx, w, ErrInvalidWorkspaceID, http.StatusBadRequest)
	}

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	var req AppNewInvitation
	if err := web.Decode(r, &req); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	invitation, err := h.invitations.CreateInvitation(ctx, workspaceID, userID, req.Email, req.Role, req.TeamIDs)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	span.AddEvent("invitation created", trace.WithAttributes(
		attribute.String("workspace_id", workspaceID.String()),
		attribute.String("email", req.Email),
	))

	return web.Respond(ctx, w, toAppInvitation(invitation), http.StatusCreated)
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

// CreateInvitationLink creates a new workspace invitation link
func (h *Handlers) CreateInvitationLink(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "handlers.invitations.CreateInvitationLink")
	defer span.End()

	workspaceID, err := uuid.Parse(web.Params(r, "workspaceId"))
	if err != nil {
		return web.RespondError(ctx, w, ErrInvalidWorkspaceID, http.StatusBadRequest)
	}

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	var req AppNewInvitationLink
	if err := web.Decode(r, &req); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	link, err := h.invitations.CreateInvitationLink(ctx, workspaceID, userID, req.Role, req.TeamIDs, req.MaxUses, req.ExpiresAt)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	span.AddEvent("invitation link created", trace.WithAttributes(
		attribute.String("workspace_id", workspaceID.String()),
	))

	return web.Respond(ctx, w, toAppInvitationLink(link), http.StatusCreated)
}

// ListInvitationLinks returns all active invitation links for a workspace
func (h *Handlers) ListInvitationLinks(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "handlers.invitations.ListInvitationLinks")
	defer span.End()

	workspaceID, err := uuid.Parse(web.Params(r, "workspaceId"))
	if err != nil {
		return web.RespondError(ctx, w, ErrInvalidWorkspaceID, http.StatusBadRequest)
	}

	links, err := h.invitations.ListInvitationLinks(ctx, workspaceID)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	return web.Respond(ctx, w, toAppInvitationLinks(links), http.StatusOK)
}

// RevokeInvitationLink revokes a workspace invitation link
func (h *Handlers) RevokeInvitationLink(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "handlers.invitations.RevokeInvitationLink")
	defer span.End()

	linkID, err := uuid.Parse(web.Params(r, "id"))
	if err != nil {
		return web.RespondError(ctx, w, ErrInvalidInvitationID, http.StatusBadRequest)
	}

	if err := h.invitations.RevokeInvitationLink(ctx, linkID); err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	span.AddEvent("invitation link revoked", trace.WithAttributes(
		attribute.String("link_id", linkID.String()),
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

// GetInvitationLink retrieves and validates an invitation link by token
func (h *Handlers) GetInvitationLink(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "handlers.invitations.GetInvitationLink")
	defer span.End()

	token := web.Params(r, "token")
	if token == "" {
		return web.RespondError(ctx, w, errors.New("token is required"), http.StatusBadRequest)
	}

	link, err := h.invitations.GetInvitationLink(ctx, token)
	if err != nil {
		switch err {
		case invitations.ErrInvitationNotFound:
			return web.RespondError(ctx, w, err, http.StatusNotFound)
		case invitations.ErrInvitationExpired:
			return web.RespondError(ctx, w, err, http.StatusGone)
		case invitations.ErrInvitationRevoked:
			return web.RespondError(ctx, w, err, http.StatusGone)
		case invitations.ErrMaxUsesReached:
			return web.RespondError(ctx, w, err, http.StatusGone)
		default:
			return web.RespondError(ctx, w, err, http.StatusInternalServerError)
		}
	}

	return web.Respond(ctx, w, toAppInvitationLink(link), http.StatusOK)
}

// UseInvitationLink increments the usage count for an invitation link
func (h *Handlers) UseInvitationLink(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "handlers.invitations.UseInvitationLink")
	defer span.End()

	linkID, err := uuid.Parse(web.Params(r, "id"))
	if err != nil {
		return web.RespondError(ctx, w, ErrInvalidInvitationID, http.StatusBadRequest)
	}

	if err := h.invitations.UseInvitationLink(ctx, linkID); err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	span.AddEvent("invitation link used", trace.WithAttributes(
		attribute.String("link_id", linkID.String()),
	))

	return web.Respond(ctx, w, nil, http.StatusNoContent)
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
