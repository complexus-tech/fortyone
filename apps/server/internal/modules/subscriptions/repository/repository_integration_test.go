//go:build integration

package subscriptionsrepository

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	subscriptionsdomain "github.com/complexus-tech/projects-api/internal/modules/subscriptions/domain"
	"github.com/complexus-tech/projects-api/internal/testkit"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestRepositoryScopesInvoicesAndReadsToWorkspace(t *testing.T) {
	postgres := testkit.NewPostgres(t)
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()
	repository := New(postgres.Pool)

	workspaceA := insertSubscriptionWorkspace(t, ctx, postgres.Pool, "a")
	workspaceB := insertSubscriptionWorkspace(t, ctx, postgres.Pool, "b")
	insertSubscription(t, ctx, postgres.Pool, workspaceA, "cus_a", "sub_a", "active")
	insertSubscription(t, ctx, postgres.Pool, workspaceB, "cus_b", "sub_b", "active")
	providerTime := time.Date(2026, time.August, 28, 11, 0, 0, 0, time.UTC)

	conflictingCustomer := subscriptionSnapshot("cus_b", "sub_a", subscriptionsdomain.StatusPastDue)
	mutation, err := repository.ApplyStripeSubscriptionSnapshot(
		ctx,
		conflictingCustomer,
		stripeCursor("evt_customer_conflict", providerTime, subscriptionsdomain.StripeEventPrioritySnapshot),
	)
	if !errors.Is(err, subscriptionsdomain.ErrProviderIdentityConflict) || mutation.Applied {
		t.Fatalf("cross-workspace customer snapshot mutation = %+v, %v", mutation, err)
	}
	if err := repository.UpdateWorkspaceSubscription(ctx, workspaceA, conflictingCustomer); !errors.Is(err, subscriptionsdomain.ErrProviderIdentityConflict) {
		t.Fatalf("manual cross-workspace customer sync error = %v", err)
	}

	_, err = repository.UpsertStripeSubscription(
		ctx,
		workspaceB,
		subscriptionSnapshot("cus_a", "sub_new_for_b", subscriptionsdomain.StatusActive),
		stripeCursor("evt_customer_reassignment", providerTime, subscriptionsdomain.StripeEventPriorityCreated),
	)
	if !errors.Is(err, subscriptionsdomain.ErrProviderIdentityConflict) {
		t.Fatalf("cross-workspace customer reassignment error = %v", err)
	}
	_, err = repository.UpsertStripeSubscription(
		ctx,
		workspaceB,
		subscriptionSnapshot("cus_b", "sub_a", subscriptionsdomain.StatusActive),
		stripeCursor("evt_subscription_reassignment", providerTime, subscriptionsdomain.StripeEventPriorityCreated),
	)
	if !errors.Is(err, subscriptionsdomain.ErrProviderIdentityConflict) {
		t.Fatalf("cross-workspace subscription reassignment error = %v", err)
	}
	storedA, err := repository.GetSubscriptionByWorkspaceID(ctx, workspaceA)
	if err != nil || storedA.StripeCustomerID != "cus_a" || storedA.SubscriptionStatus == nil || *storedA.SubscriptionStatus != subscriptionsdomain.StatusActive {
		t.Fatalf("workspace A subscription after identity conflicts = %#v, %v", storedA, err)
	}

	invoice := subscriptionsdomain.SubscriptionInvoice{
		WorkspaceID: workspaceA, StripeInvoiceID: "in_tenant_bound", AmountPaid: 42.50,
		InvoiceDate: time.Date(2026, time.August, 28, 12, 0, 0, 0, time.UTC),
		Status:      "paid", SeatsCount: 3,
	}
	if err := repository.UpsertStripeInvoice(ctx, "cus_b", invoice); !errors.Is(err, subscriptionsdomain.ErrSubscriptionNotFound) {
		t.Fatalf("invoice with mismatched customer binding error = %v", err)
	}
	if err := repository.UpsertStripeInvoice(ctx, "cus_a", invoice); err != nil {
		t.Fatalf("upsert workspace invoice: %v", err)
	}

	invoicesA, err := repository.GetInvoicesByWorkspaceID(ctx, workspaceA)
	if err != nil || len(invoicesA) != 1 || invoicesA[0].StripeInvoiceID != invoice.StripeInvoiceID {
		t.Fatalf("workspace A invoices = %#v, %v", invoicesA, err)
	}
	invoicesB, err := repository.GetInvoicesByWorkspaceID(ctx, workspaceB)
	if err != nil || len(invoicesB) != 0 {
		t.Fatalf("workspace B invoices = %#v, %v", invoicesB, err)
	}

	invoice.WorkspaceID = workspaceB
	if err := repository.UpsertStripeInvoice(ctx, "cus_b", invoice); !errors.Is(err, subscriptionsdomain.ErrProviderIdentityConflict) {
		t.Fatalf("cross-workspace invoice reassignment error = %v", err)
	}
}

func TestRepositoryAppliesStripeSubscriptionEventsMonotonically(t *testing.T) {
	postgres := testkit.NewPostgres(t)
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()
	repository := New(postgres.Pool)
	workspaceID := insertSubscriptionWorkspace(t, ctx, postgres.Pool, "ordering")

	baseTime := time.Date(2026, time.August, 28, 12, 0, 0, 0, time.UTC)
	snapshot := subscriptionSnapshot("cus_ordering", "sub_ordering", subscriptionsdomain.StatusActive)
	created, err := repository.UpsertStripeSubscription(ctx, workspaceID, snapshot, stripeCursor("evt_z_create", baseTime, subscriptionsdomain.StripeEventPriorityCreated))
	if err != nil || !created.Applied {
		t.Fatalf("create subscription mutation = %+v, %v", created, err)
	}

	snapshot.Status = subscriptionsdomain.StatusPastDue
	currentSnapshot, err := repository.ApplyStripeSubscriptionSnapshot(ctx, snapshot, stripeCursor("evt_a_snapshot", baseTime, subscriptionsdomain.StripeEventPrioritySnapshot))
	if err != nil || !currentSnapshot.Applied {
		t.Fatalf("same-second current snapshot mutation = %+v, %v", currentSnapshot, err)
	}

	snapshot.Status = subscriptionsdomain.StatusActive
	lateCreation, err := repository.UpsertStripeSubscription(ctx, workspaceID, snapshot, stripeCursor("evt_zz_late_create", baseTime, subscriptionsdomain.StripeEventPriorityCreated))
	if err != nil || lateCreation.Applied {
		t.Fatalf("same-second lower-priority creation mutation = %+v, %v", lateCreation, err)
	}

	snapshot.Status = subscriptionsdomain.StatusUnpaid
	newerSnapshot, err := repository.ApplyStripeSubscriptionSnapshot(ctx, snapshot, stripeCursor("evt_newer_snapshot", baseTime.Add(time.Second), subscriptionsdomain.StripeEventPrioritySnapshot))
	if err != nil || !newerSnapshot.Applied {
		t.Fatalf("newer current snapshot mutation = %+v, %v", newerSnapshot, err)
	}
	snapshot.Status = subscriptionsdomain.StatusActive
	olderSnapshot, err := repository.ApplyStripeSubscriptionSnapshot(ctx, snapshot, stripeCursor("evt_older_snapshot", baseTime, subscriptionsdomain.StripeEventPrioritySnapshot))
	if err != nil || olderSnapshot.Applied {
		t.Fatalf("older current snapshot mutation = %+v, %v", olderSnapshot, err)
	}

	deletedAt := baseTime.Add(2 * time.Second)
	deleted, err := repository.ApplyStripeSubscriptionDeletion(ctx, snapshot.StripeSubscriptionID, stripeCursor("evt_deleted", deletedAt, subscriptionsdomain.StripeEventPriorityDeleted))
	if err != nil || !deleted.Applied {
		t.Fatalf("delete mutation = %+v, %v", deleted, err)
	}
	lateSnapshot, err := repository.ApplyStripeSubscriptionSnapshot(ctx, snapshot, stripeCursor("evt_same_second_update", deletedAt, subscriptionsdomain.StripeEventPrioritySnapshot))
	if err != nil || lateSnapshot.Applied {
		t.Fatalf("same-second lower-priority mutation = %+v, %v", lateSnapshot, err)
	}
	lateCreation, err = repository.UpsertStripeSubscription(ctx, workspaceID, snapshot, stripeCursor("evt_after_delete_create", deletedAt, subscriptionsdomain.StripeEventPriorityCreated))
	if err != nil || lateCreation.Applied {
		t.Fatalf("same-second creation after deletion mutation = %+v, %v", lateCreation, err)
	}

	stored, err := repository.GetSubscriptionByWorkspaceID(ctx, workspaceID)
	if err != nil || stored.SubscriptionStatus == nil || *stored.SubscriptionStatus != subscriptionsdomain.StatusCanceled {
		t.Fatalf("stored subscription = %#v, %v", stored, err)
	}
}

func TestStripeWebhookClaimsSerializeConcurrentDeliveries(t *testing.T) {
	postgres := testkit.NewPostgres(t)
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()
	repository := New(postgres.Pool)
	attemptedAt := time.Date(2026, time.August, 28, 13, 0, 0, 0, time.UTC)

	const deliveries = 8
	start := make(chan struct{})
	claims := make(chan subscriptionsdomain.WebhookClaim, deliveries)
	errorsByDelivery := make(chan error, deliveries)
	var waitGroup sync.WaitGroup
	for range deliveries {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			<-start
			claim, err := repository.ClaimWebhookEvent(ctx, "evt_concurrent", "invoice.paid", attemptedAt, attemptedAt.Add(5*time.Minute))
			if err != nil {
				errorsByDelivery <- err
				return
			}
			claims <- claim
		}()
	}
	close(start)
	waitGroup.Wait()
	close(claims)
	close(errorsByDelivery)
	for err := range errorsByDelivery {
		t.Fatalf("concurrent claim: %v", err)
	}

	var acquired subscriptionsdomain.WebhookClaim
	acquiredCount := 0
	inProgressCount := 0
	for claim := range claims {
		switch claim.Disposition {
		case subscriptionsdomain.WebhookClaimAcquired:
			acquired = claim
			acquiredCount++
		case subscriptionsdomain.WebhookClaimInProgress:
			inProgressCount++
		}
	}
	if acquiredCount != 1 || inProgressCount != deliveries-1 {
		t.Fatalf("acquired=%d in_progress=%d", acquiredCount, inProgressCount)
	}
	workspaceID := insertSubscriptionWorkspace(t, ctx, postgres.Pool, "webhook")
	if err := repository.MarkWebhookEventProcessed(ctx, "evt_concurrent", acquired.LeaseToken, subscriptionsdomain.WebhookOutcome{
		Result: subscriptionsdomain.WebhookResultHandled, WorkspaceID: &workspaceID,
	}, attemptedAt.Add(time.Second)); err != nil {
		t.Fatalf("complete webhook: %v", err)
	}
	duplicate, err := repository.ClaimWebhookEvent(ctx, "evt_concurrent", "invoice.paid", attemptedAt.Add(2*time.Second), attemptedAt.Add(7*time.Minute))
	if err != nil || duplicate.Disposition != subscriptionsdomain.WebhookClaimAlreadyProcessed {
		t.Fatalf("terminal duplicate = %+v, %v", duplicate, err)
	}
}

func TestStripeWebhookClaimRejectsEventTypeConfusionAndStaleOwner(t *testing.T) {
	postgres := testkit.NewPostgres(t)
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()
	repository := New(postgres.Pool)
	firstAttempt := time.Date(2026, time.August, 28, 14, 0, 0, 0, time.UTC)
	first, err := repository.ClaimWebhookEvent(ctx, "evt_fenced", "invoice.paid", firstAttempt, firstAttempt.Add(time.Minute))
	if err != nil {
		t.Fatalf("first claim: %v", err)
	}
	if _, err := repository.ClaimWebhookEvent(ctx, "evt_fenced", "customer.subscription.deleted", firstAttempt.Add(time.Second), firstAttempt.Add(time.Minute)); !errors.Is(err, subscriptionsdomain.ErrWebhookEventTypeConflict) {
		t.Fatalf("event type confusion error = %v", err)
	}
	replacement, err := repository.ClaimWebhookEvent(ctx, "evt_fenced", "invoice.paid", firstAttempt.Add(time.Minute), firstAttempt.Add(2*time.Minute))
	if err != nil || replacement.Disposition != subscriptionsdomain.WebhookClaimAcquired {
		t.Fatalf("replacement claim = %+v, %v", replacement, err)
	}
	if err := repository.MarkWebhookEventFailed(ctx, "evt_fenced", first.LeaseToken, subscriptionsdomain.WebhookFailureHandler, firstAttempt.Add(time.Minute)); !errors.Is(err, subscriptionsdomain.ErrWebhookEventClaimLost) {
		t.Fatalf("stale owner error = %v", err)
	}
}

func subscriptionSnapshot(customerID, subscriptionID string, status subscriptionsdomain.SubscriptionStatus) subscriptionsdomain.SubscriptionSnapshot {
	itemID := "si_" + subscriptionID
	return subscriptionsdomain.SubscriptionSnapshot{
		StripeCustomerID: customerID, StripeSubscriptionID: subscriptionID,
		StripeSubscriptionItemID: &itemID, Status: status,
		Tier: subscriptionsdomain.TierPro, SeatCount: 3,
	}
}

func stripeCursor(eventID string, createdAt time.Time, priority int16) subscriptionsdomain.StripeEventCursor {
	return subscriptionsdomain.StripeEventCursor{EventID: eventID, CreatedAt: createdAt, Priority: priority}
}

func insertSubscriptionWorkspace(t *testing.T, ctx context.Context, pool *pgxpool.Pool, label string) uuid.UUID {
	t.Helper()
	id := uuid.New()
	if _, err := pool.Exec(ctx, `INSERT INTO workspaces (workspace_id, name, slug) VALUES ($1, $2, $3)`, id, "Billing "+label, "billing-"+label+"-"+uuid.NewString()); err != nil {
		t.Fatalf("insert workspace: %v", err)
	}
	return id
}

func insertSubscription(t *testing.T, ctx context.Context, pool *pgxpool.Pool, workspaceID uuid.UUID, customerID, subscriptionID, status string) {
	t.Helper()
	if _, err := pool.Exec(ctx, `
		INSERT INTO workspace_subscriptions (
			workspace_id, stripe_customer_id, stripe_subscription_id,
			subscription_status, subscription_tier, seat_count, created_at, updated_at
		) VALUES ($1, $2, $3, $4, 'pro', 2, NOW(), NOW())
	`, workspaceID, customerID, subscriptionID, status); err != nil {
		t.Fatalf("insert subscription: %v", err)
	}
}

func TestMigrationDoesNotRetainRawStripePayloads(t *testing.T) {
	postgres := testkit.NewPostgres(t)
	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	defer cancel()
	var exists bool
	if err := postgres.Pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM information_schema.columns
			WHERE table_schema = 'public'
			  AND table_name = 'stripe_webhook_events'
			  AND column_name = 'payload'
		)
	`).Scan(&exists); err != nil {
		t.Fatalf("inspect webhook schema: %v", err)
	}
	if exists {
		t.Fatal("stripe_webhook_events retains raw signed payloads")
	}
}
