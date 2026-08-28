package subscriptions

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	subscriptionsdomain "github.com/complexus-tech/projects-api/internal/modules/subscriptions/domain"
	"github.com/complexus-tech/projects-api/pkg/tasks"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/stripe/stripe-go/v82"
	"go.opentelemetry.io/otel"
)

const stripeWebhookTaskRetention = 7 * 24 * time.Hour

func (s *Service) handleCheckoutSessionCompleted(ctx context.Context, event stripe.Event) (WebhookOutcome, error) {
	ctx, span := otel.Tracer("subscriptions.service").Start(ctx, "subscriptions.CheckoutCompleted")
	defer span.End()

	var session stripe.CheckoutSession
	if err := json.Unmarshal(event.Data.Raw, &session); err != nil {
		return WebhookOutcome{}, fmt.Errorf("decode checkout session: %w", err)
	}
	if session.Subscription == nil || session.Subscription.ID == "" {
		return WebhookOutcome{}, fmt.Errorf("checkout session %s has no subscription", session.ID)
	}

	stripeSubscription, err := s.stripeClient.Subscriptions.Get(session.Subscription.ID, expandedSubscriptionParams(ctx))
	if err != nil {
		return WebhookOutcome{}, fmt.Errorf("fetch checkout subscription: %w", err)
	}
	snapshot, err := subscriptionSnapshot(stripeSubscription)
	if err != nil {
		return WebhookOutcome{}, err
	}
	mutation, err := s.repo.ApplyStripeSubscriptionSnapshot(ctx, snapshot, stripeEventCursor(event, subscriptionsdomain.StripeEventPrioritySnapshot))
	if err != nil {
		return WebhookOutcome{}, fmt.Errorf("apply checkout subscription: %w", err)
	}
	return mutationOutcome(mutation), nil
}

func (s *Service) handleSubscriptionUpdated(ctx context.Context, event stripe.Event) (WebhookOutcome, error) {
	ctx, span := otel.Tracer("subscriptions.service").Start(ctx, "subscriptions.SubscriptionUpdated")
	defer span.End()

	var delivered stripe.Subscription
	if err := json.Unmarshal(event.Data.Raw, &delivered); err != nil {
		return WebhookOutcome{}, fmt.Errorf("decode subscription update: %w", err)
	}
	if delivered.ID == "" {
		return WebhookOutcome{}, ErrInvalidSubscription
	}

	// Stripe does not guarantee delivery ordering. Read the current provider
	// snapshot instead of trusting an older event body, then fence the write by
	// the durable event cursor.
	current, err := s.stripeClient.Subscriptions.Get(delivered.ID, expandedSubscriptionParams(ctx))
	if err != nil {
		return WebhookOutcome{}, fmt.Errorf("fetch current Stripe subscription: %w", err)
	}
	snapshot, err := subscriptionSnapshot(current)
	if err != nil {
		return WebhookOutcome{}, err
	}
	mutation, err := s.repo.ApplyStripeSubscriptionSnapshot(ctx, snapshot, stripeEventCursor(event, subscriptionsdomain.StripeEventPrioritySnapshot))
	if err != nil {
		return WebhookOutcome{}, fmt.Errorf("apply Stripe subscription update: %w", err)
	}
	return mutationOutcome(mutation), nil
}

func (s *Service) handleSubscriptionDeleted(ctx context.Context, event stripe.Event) (WebhookOutcome, error) {
	var deleted stripe.Subscription
	if err := json.Unmarshal(event.Data.Raw, &deleted); err != nil {
		return WebhookOutcome{}, fmt.Errorf("decode subscription deletion: %w", err)
	}
	if deleted.ID == "" {
		return WebhookOutcome{}, ErrInvalidSubscription
	}
	mutation, err := s.repo.ApplyStripeSubscriptionDeletion(
		ctx,
		deleted.ID,
		stripeEventCursor(event, subscriptionsdomain.StripeEventPriorityDeleted),
	)
	if err != nil {
		return WebhookOutcome{}, fmt.Errorf("apply Stripe subscription deletion: %w", err)
	}
	return mutationOutcome(mutation), nil
}

func (s *Service) handleSubscriptionCreated(ctx context.Context, event stripe.Event) (WebhookOutcome, error) {
	var delivered stripe.Subscription
	if err := json.Unmarshal(event.Data.Raw, &delivered); err != nil {
		return WebhookOutcome{}, fmt.Errorf("decode subscription creation: %w", err)
	}
	if delivered.ID == "" {
		return WebhookOutcome{}, ErrInvalidSubscription
	}

	// A creation delivery can arrive after an update from the same Stripe
	// timestamp second. Reconcile from current provider state so the lower
	// creation priority can never restore the older delivered snapshot.
	current, err := s.stripeClient.Subscriptions.Get(delivered.ID, expandedSubscriptionParams(ctx))
	if err != nil {
		return WebhookOutcome{}, fmt.Errorf("fetch current Stripe subscription after creation: %w", err)
	}
	if current == nil || current.Customer == nil || current.Customer.ID == "" {
		return WebhookOutcome{}, ErrInvalidSubscription
	}

	customerParams := &stripe.CustomerParams{}
	customerParams.Context = ctx
	customer, err := s.stripeClient.Customers.Get(current.Customer.ID, customerParams)
	if err != nil {
		return WebhookOutcome{}, fmt.Errorf("fetch Stripe customer: %w", err)
	}
	workspaceID, err := workspaceIDFromCustomer(customer)
	if err != nil {
		return WebhookOutcome{}, err
	}
	snapshot, err := subscriptionSnapshot(current)
	if err != nil {
		return WebhookOutcome{}, err
	}
	mutation, err := s.repo.UpsertStripeSubscription(ctx, workspaceID, snapshot, stripeEventCursor(event, subscriptionsdomain.StripeEventPriorityCreated))
	if err != nil {
		return WebhookOutcome{}, fmt.Errorf("upsert Stripe subscription: %w", err)
	}

	if mutation.Applied {
		s.enqueueTrialEndNotice(ctx, mutation.WorkspaceID, snapshot.StripeSubscriptionID)
	}
	return mutationOutcome(mutation), nil
}

func (s *Service) handleInvoicePaid(ctx context.Context, event stripe.Event) (WebhookOutcome, error) {
	var invoice stripe.Invoice
	if err := json.Unmarshal(event.Data.Raw, &invoice); err != nil {
		return WebhookOutcome{}, fmt.Errorf("decode paid invoice: %w", err)
	}
	if invoice.ID == "" || invoice.Customer == nil || invoice.Customer.ID == "" {
		return WebhookOutcome{}, ErrInvalidInvoice
	}
	customerParams := &stripe.CustomerParams{}
	customerParams.Context = ctx
	customer, err := s.stripeClient.Customers.Get(invoice.Customer.ID, customerParams)
	if err != nil {
		return WebhookOutcome{}, fmt.Errorf("fetch invoice customer: %w", err)
	}
	workspaceID, err := workspaceIDFromCustomer(customer)
	if err != nil {
		return WebhookOutcome{}, err
	}

	invoiceDate := time.Unix(invoice.Created, 0).UTC()
	if invoice.StatusTransitions != nil && invoice.StatusTransitions.PaidAt > 0 {
		invoiceDate = time.Unix(invoice.StatusTransitions.PaidAt, 0).UTC()
	}
	seatCount := 0
	if invoice.Lines != nil && len(invoice.Lines.Data) > 0 && invoice.Lines.Data[0].Quantity > 0 {
		seatCount = int(invoice.Lines.Data[0].Quantity)
	}
	paid := CoreSubscriptionInvoice{
		WorkspaceID: workspaceID, StripeInvoiceID: invoice.ID,
		AmountPaid: float64(invoice.AmountPaid) / 100, InvoiceDate: invoiceDate,
		Status: string(invoice.Status), SeatsCount: seatCount,
		HostedURL: &invoice.HostedInvoiceURL, CustomerName: &customer.Name,
	}
	if err := s.repo.UpsertStripeInvoice(ctx, customer.ID, paid); err != nil {
		return WebhookOutcome{}, fmt.Errorf("upsert paid Stripe invoice: %w", err)
	}
	return WebhookOutcome{Result: WebhookResultHandled, WorkspaceID: &workspaceID}, nil
}

func subscriptionSnapshot(subscription *stripe.Subscription) (subscriptionsdomain.SubscriptionSnapshot, error) {
	if subscription == nil || subscription.ID == "" || subscription.Customer == nil || subscription.Customer.ID == "" {
		return subscriptionsdomain.SubscriptionSnapshot{}, ErrInvalidSubscription
	}
	itemID, seats, tier, interval, billingEndsAt, err := stripeSubscriptionDetails(subscription)
	if err != nil {
		return subscriptionsdomain.SubscriptionSnapshot{}, err
	}
	var trialEnd *time.Time
	if subscription.TrialEnd > 0 {
		value := time.Unix(subscription.TrialEnd, 0).UTC()
		trialEnd = &value
	}
	return subscriptionsdomain.SubscriptionSnapshot{
		StripeCustomerID: subscription.Customer.ID, StripeSubscriptionID: subscription.ID,
		StripeSubscriptionItemID: &itemID, Status: SubscriptionStatus(subscription.Status),
		Tier: tier, SeatCount: seats, TrialEnd: trialEnd,
		BillingInterval: interval, BillingEndsAt: billingEndsAt,
	}, nil
}

func (s *Service) enqueueTrialEndNotice(ctx context.Context, workspaceID uuid.UUID, subscriptionID string) {
	if s.tasksService == nil {
		return
	}
	email, err := s.repo.GetWorkspaceCreatorEmail(ctx, workspaceID)
	if err != nil || email == "" {
		return
	}
	if _, err := s.tasksService.EnqueueWorkspaceTrialEnd(
		tasks.WorkspaceTrialEndPayload{Email: email},
		asynq.TaskID("stripe-subscription-trial-end:"+subscriptionID),
		asynq.Retention(stripeWebhookTaskRetention),
	); err != nil {
		s.log.Error(ctx, "Failed to enqueue workspace trial end notice", "workspace_id", workspaceID, "error", err)
	}
}

func workspaceIDFromCustomer(customer *stripe.Customer) (uuid.UUID, error) {
	if customer == nil {
		return uuid.Nil, errorsInvalidCustomerMetadata("")
	}
	workspaceID, err := uuid.Parse(customer.Metadata["workspace_id"])
	if err != nil {
		return uuid.Nil, errorsInvalidCustomerMetadata(customer.ID)
	}
	return workspaceID, nil
}

func errorsInvalidCustomerMetadata(customerID string) error {
	return fmt.Errorf("stripe customer %s has invalid workspace metadata", customerID)
}

func expandedSubscriptionParams(ctx context.Context) *stripe.SubscriptionParams {
	return &stripe.SubscriptionParams{
		Params: stripe.Params{Context: ctx},
		Expand: []*string{stripe.String("items.data.price"), stripe.String("customer")},
	}
}

func stripeEventCursor(event stripe.Event, priority int16) subscriptionsdomain.StripeEventCursor {
	return subscriptionsdomain.StripeEventCursor{
		CreatedAt: time.Unix(event.Created, 0).UTC(), Priority: priority, EventID: event.ID,
	}
}

func mutationOutcome(mutation subscriptionsdomain.SubscriptionMutation) WebhookOutcome {
	result := WebhookResultHandled
	if !mutation.Applied {
		result = WebhookResultIgnored
	}
	workspaceID := mutation.WorkspaceID
	return WebhookOutcome{Result: result, WorkspaceID: &workspaceID}
}
