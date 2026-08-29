package subscriptionshttp

import (
	"context"
	"errors"
	"net/http"

	subscriptions "github.com/complexus-tech/projects-api/internal/modules/subscriptions/service"
	users "github.com/complexus-tech/projects-api/internal/modules/users/service"
	workspaces "github.com/complexus-tech/projects-api/internal/modules/workspaces/service"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
)

var (
	ErrInvalidWorkspaceID = errors.New("workspace id is not in its proper form")
)

type Handlers struct {
	subscriptions *subscriptions.Service
	webhookEvents webhookEventService
	users         *users.Service
	workspaces    *workspaces.Service
	log           *logger.Logger
}

type webhookEventService interface {
	HandleWebhookEvent(ctx context.Context, payload []byte, signature string) error
}

func New(subscriptions *subscriptions.Service, users *users.Service, workspaces *workspaces.Service, log *logger.Logger) *Handlers {
	return &Handlers{
		subscriptions: subscriptions,
		webhookEvents: subscriptions,
		users:         users,
		workspaces:    workspaces,
		log:           log,
	}
}

func (h *Handlers) CreateCheckoutSession(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "handlers.subscriptions.CreateCheckoutSession")
	defer span.End()

	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		span.RecordError(err)
		web.RespondError(ctx, w, err, http.StatusUnauthorized)
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

	var req AppCheckoutRequest
	if err := web.Decode(r, &req); err != nil {
		span.RecordError(err)
		web.RespondError(ctx, w, err, http.StatusBadRequest)
		return nil
	}

	url, err := h.subscriptions.CreateCheckoutSession(ctx, workspace.ID, req.PriceLookupKey, user.Email, workspace.Name, req.SuccessURL, req.CancelURL)
	if err != nil {
		if errors.Is(err, subscriptions.ErrWorkspaceHasActiveSub) ||
			errors.Is(err, subscriptions.ErrInvalidBillingRedirect) ||
			errors.Is(err, subscriptions.ErrInvalidPriceLookupKey) {
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

func (h *Handlers) AddSeat(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "handlers.subscriptions.AddSeat")
	defer span.End()

	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		span.RecordError(err)
		web.RespondError(ctx, w, err, http.StatusUnauthorized)
		return nil
	}

	err = h.subscriptions.UpdateSubscriptionSeats(ctx, workspace.ID)
	if err != nil {
		span.RecordError(err)
		web.RespondError(ctx, w, err, http.StatusInternalServerError)
		return nil
	}

	web.Respond(ctx, w, nil, http.StatusOK)
	return nil
}

func (h *Handlers) CreateCustomerPortal(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "handlers.subscriptions.CreateCustomerPortal")
	defer span.End()

	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		span.RecordError(err)
		web.RespondError(ctx, w, err, http.StatusUnauthorized)
		return nil
	}

	var req AppCustomerPortalRequest
	if err := web.Decode(r, &req); err != nil {
		span.RecordError(err)
		web.RespondError(ctx, w, err, http.StatusBadRequest)
		return nil
	}

	url, err := h.subscriptions.CreateCustomerPortalSession(ctx, workspace.ID, req.ReturnURL)
	if err != nil {
		if errors.Is(err, subscriptions.ErrInvalidBillingRedirect) {
			web.RespondError(ctx, w, err, http.StatusBadRequest)
			return nil
		}
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

func (h *Handlers) GetSubscription(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "handlers.subscriptions.GetSubscription")
	defer span.End()

	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		span.RecordError(err)
		web.RespondError(ctx, w, err, http.StatusUnauthorized)
		return nil
	}

	subscription, err := h.subscriptions.GetSubscription(ctx, workspace.ID)
	if err != nil {
		span.RecordError(err)
		if errors.Is(err, subscriptions.ErrSubscriptionNotFound) {
			web.RespondError(ctx, w, err, http.StatusNotFound)
			return nil
		}
		web.RespondError(ctx, w, err, http.StatusInternalServerError)
		return nil
	}

	web.Respond(ctx, w, toAppSubscription(subscription), http.StatusOK)
	return nil
}

func (h *Handlers) GetInvoices(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "handlers.subscriptions.GetInvoices")
	defer span.End()

	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		span.RecordError(err)
		web.RespondError(ctx, w, err, http.StatusUnauthorized)
		return nil
	}

	invoices, err := h.subscriptions.GetInvoices(ctx, workspace.ID)
	if err != nil {
		span.RecordError(err)
		web.RespondError(ctx, w, err, http.StatusInternalServerError)
		return nil
	}

	web.Respond(ctx, w, toAppInvoices(invoices), http.StatusOK)
	return nil
}

func (h *Handlers) HandleWebhook(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "handlers.subscriptions.HandleWebhook")
	defer span.End()

	const maxBodyBytes = int64(64 * 1024)
	body, err := web.ReadBoundedBody(w, r, maxBodyBytes)
	if err != nil {
		span.RecordError(err)
		h.log.Warn(ctx, "Rejected unreadable Stripe webhook body")
		if errors.Is(err, web.ErrRequestBodyTooLarge) {
			web.RespondError(ctx, w, err, http.StatusRequestEntityTooLarge)
			return nil
		}
		web.RespondError(ctx, w, web.HumanizeJSONDecodeError(err), http.StatusBadRequest)
		return nil
	}

	signature := r.Header.Get("Stripe-Signature")
	if signature == "" {
		span.RecordError(errors.New("missing Stripe-Signature header"))
		h.log.Error(ctx, "Missing Stripe-Signature header")
		web.RespondError(ctx, w, errors.New("missing Stripe-Signature header"), http.StatusBadRequest)
		return nil
	}

	err = h.webhookEvents.HandleWebhookEvent(ctx, body, signature)
	if err != nil {
		span.RecordError(err)
		if errors.Is(err, subscriptions.ErrInvalidWebhookSignature) {
			h.log.Warn(ctx, "Rejected Stripe webhook with an invalid signature")
			web.RespondError(ctx, w, subscriptions.ErrInvalidWebhookSignature, http.StatusBadRequest)
			return nil
		}

		// Stripe retries only non-2xx responses. Processing, lease, and durable
		// persistence failures must therefore remain retryable.
		h.log.Error(ctx, "Stripe webhook did not reach a durable terminal state")
		web.RespondError(ctx, w, err, http.StatusInternalServerError)
		return nil
	}

	web.Respond(ctx, w, nil, http.StatusOK)
	return nil
}

func (h *Handlers) ChangeSubscriptionPlan(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "handlers.subscriptions.ChangeSubscriptionPlan")
	defer span.End()

	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		span.RecordError(err)
		web.RespondError(ctx, w, err, http.StatusUnauthorized)
		return nil
	}

	var req AppChangeSubscriptionPlanRequest
	if err := web.Decode(r, &req); err != nil {
		span.RecordError(err)
		web.RespondError(ctx, w, err, http.StatusBadRequest)
		return nil
	}

	err = h.subscriptions.ChangeSubscriptionPlan(ctx, workspace.ID, req.NewLookupKey)
	if err != nil {
		switch {
		case errors.Is(err, subscriptions.ErrNoActiveSubscriptionToChange),
			errors.Is(err, subscriptions.ErrAlreadySubscribedToThisPlan),
			errors.Is(err, subscriptions.ErrInvalidPriceLookupKey):
			web.RespondError(ctx, w, err, http.StatusBadRequest)
		default:
			span.RecordError(err)
			web.RespondError(ctx, w, err, http.StatusInternalServerError)
		}
		return nil
	}

	web.Respond(ctx, w, nil, http.StatusOK)
	return nil
}
