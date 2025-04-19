package subscriptionsgrp

import (
	"context"
	"errors"
	"io"
	"net/http"

	"github.com/complexus-tech/projects-api/internal/core/subscriptions"
	"github.com/complexus-tech/projects-api/internal/core/users"
	"github.com/complexus-tech/projects-api/internal/core/workspaces"
	"github.com/complexus-tech/projects-api/internal/web/mid"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

var (
	ErrInvalidWorkspaceID = errors.New("workspace id is not in its proper form")
)

type Handlers struct {
	subscriptions *subscriptions.Service
	users         *users.Service
	workspaces    *workspaces.Service
	log           *logger.Logger
}

func New(subscriptions *subscriptions.Service, users *users.Service, workspaces *workspaces.Service, log *logger.Logger) *Handlers {
	return &Handlers{
		subscriptions: subscriptions,
		users:         users,
		workspaces:    workspaces,
		log:           log,
	}
}

// CreateCheckoutSession generates a new Stripe checkout session
func (h *Handlers) CreateCheckoutSession(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "handlers.subscriptions.CreateCheckoutSession")
	defer span.End()

	workspaceIdParam := web.Params(r, "workspaceId")
	workspaceId, err := uuid.Parse(workspaceIdParam)
	if err != nil {
		span.RecordError(err)
		web.RespondError(ctx, w, ErrInvalidWorkspaceID, http.StatusBadRequest)
		return nil
	}

	userId, err := mid.GetUserID(ctx)
	if err != nil {
		span.RecordError(err)
		web.RespondError(ctx, w, err, http.StatusInternalServerError)
		return nil
	}
	user, err := h.users.GetUser(ctx, userId)
	if err != nil {
		span.RecordError(err)
		web.RespondError(ctx, w, err, http.StatusInternalServerError)
		return nil
	}

	workspace, err := h.workspaces.Get(ctx, workspaceId, userId)
	if err != nil {
		span.RecordError(err)
		web.RespondError(ctx, w, err, http.StatusInternalServerError)
		return nil
	}

	var req AppCheckoutRequest
	if err := web.Decode(r, &req); err != nil {
		span.RecordError(err)
		web.RespondError(ctx, w, err, http.StatusBadRequest)
		return nil
	}

	url, err := h.subscriptions.CreateCheckoutSession(ctx, workspaceId, req.PriceLookupKey, user.Email, workspace.Name)
	if err != nil {
		if errors.Is(err, subscriptions.ErrWorkspaceHasActiveSub) {
			web.RespondError(ctx, w, err, http.StatusBadRequest)
			return nil
		}
		span.RecordError(err)
		web.RespondError(ctx, w, err, http.StatusInternalServerError)
		return nil
	}

	resp := AppCheckoutResponse{
		URL: url,
	}

	web.Respond(ctx, w, resp, http.StatusOK)
	return nil
}

// AddSeat increases the number of seats in a subscription
func (h *Handlers) AddSeat(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "handlers.subscriptions.AddSeat")
	defer span.End()

	workspaceIdParam := web.Params(r, "workspaceId")
	workspaceId, err := uuid.Parse(workspaceIdParam)
	if err != nil {
		span.RecordError(err)
		web.RespondError(ctx, w, ErrInvalidWorkspaceID, http.StatusBadRequest)
		return nil
	}

	err = h.subscriptions.AddSeatToSubscription(ctx, workspaceId)
	if err != nil {
		span.RecordError(err)
		web.RespondError(ctx, w, err, http.StatusInternalServerError)
		return nil
	}

	web.Respond(ctx, w, nil, http.StatusOK)
	return nil
}

// CreateCustomerPortal generates a URL for the Stripe Customer Portal
func (h *Handlers) CreateCustomerPortal(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "handlers.subscriptions.CreateCustomerPortal")
	defer span.End()

	workspaceIdParam := web.Params(r, "workspaceId")
	workspaceId, err := uuid.Parse(workspaceIdParam)
	if err != nil {
		span.RecordError(err)
		web.RespondError(ctx, w, ErrInvalidWorkspaceID, http.StatusBadRequest)
		return nil
	}

	var req AppCustomerPortalRequest
	if err := web.Decode(r, &req); err != nil {
		span.RecordError(err)
		web.RespondError(ctx, w, err, http.StatusBadRequest)
		return nil
	}

	url, err := h.subscriptions.CreateCustomerPortalSession(ctx, workspaceId, req.ReturnURL)
	if err != nil {
		span.RecordError(err)
		web.RespondError(ctx, w, err, http.StatusInternalServerError)
		return nil
	}

	resp := AppCustomerPortalResponse{
		URL: url,
	}

	web.Respond(ctx, w, resp, http.StatusOK)
	return nil
}

// GetSubscription returns the current subscription for a workspace
func (h *Handlers) GetSubscription(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "handlers.subscriptions.GetSubscription")
	defer span.End()

	workspaceIdParam := web.Params(r, "workspaceId")
	workspaceId, err := uuid.Parse(workspaceIdParam)
	if err != nil {
		span.RecordError(err)
		web.RespondError(ctx, w, ErrInvalidWorkspaceID, http.StatusBadRequest)
		return nil
	}

	subscription, err := h.subscriptions.GetSubscription(ctx, workspaceId)
	if err != nil {
		span.RecordError(err)
		web.RespondError(ctx, w, err, http.StatusInternalServerError)
		return nil
	}

	web.Respond(ctx, w, toAppSubscription(subscription), http.StatusOK)
	return nil
}

// GetInvoices returns the invoices for a workspace
func (h *Handlers) GetInvoices(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "handlers.subscriptions.GetInvoices")
	defer span.End()

	workspaceIdParam := web.Params(r, "workspaceId")
	workspaceId, err := uuid.Parse(workspaceIdParam)
	if err != nil {
		span.RecordError(err)
		web.RespondError(ctx, w, ErrInvalidWorkspaceID, http.StatusBadRequest)
		return nil
	}

	invoices, err := h.subscriptions.GetInvoices(ctx, workspaceId)
	if err != nil {
		span.RecordError(err)
		web.RespondError(ctx, w, err, http.StatusInternalServerError)
		return nil
	}

	web.Respond(ctx, w, toAppInvoices(invoices), http.StatusOK)
	return nil
}

// HandleWebhook processes Stripe webhook events
func (h *Handlers) HandleWebhook(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "handlers.subscriptions.HandleWebhook")
	defer span.End()

	const MaxBodyBytes = int64(65536)
	r.Body = http.MaxBytesReader(w, r.Body, MaxBodyBytes)
	body, err := io.ReadAll(r.Body)
	if err != nil {
		span.RecordError(err)
		h.log.Error(ctx, "Failed to read webhook body", "error", err)
		web.RespondError(ctx, w, err, http.StatusBadRequest)
		return nil
	}

	// Get the Stripe signature from headers
	signature := r.Header.Get("Stripe-Signature")
	if signature == "" {
		span.RecordError(errors.New("missing Stripe-Signature header"))
		h.log.Error(ctx, "Missing Stripe-Signature header")
		web.RespondError(ctx, w, errors.New("missing Stripe-Signature header"), http.StatusBadRequest)
		return nil
	}

	err = h.subscriptions.HandleWebhookEvent(ctx, body, signature)
	if err != nil {
		span.RecordError(err)
		h.log.Error(ctx, "Failed to handle webhook event", "error", err)
		// Still return 200 to prevent Stripe from retrying failed events with the same payload
		// since we've already logged the error
		if errors.Is(err, subscriptions.ErrFailedToCreateSubscription) {
			web.RespondError(ctx, w, err, http.StatusInternalServerError)
			return nil
		}
		web.Respond(ctx, w, nil, http.StatusOK)
		return nil
	}

	web.Respond(ctx, w, nil, http.StatusOK)
	return nil
}
