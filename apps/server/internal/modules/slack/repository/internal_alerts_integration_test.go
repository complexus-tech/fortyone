//go:build integration

package slackrepository

import (
	"sync"
	"testing"
	"time"

	slackdomain "github.com/complexus-tech/projects-api/internal/modules/slack/domain"
	subscriptionsdomain "github.com/complexus-tech/projects-api/internal/modules/subscriptions/domain"
	subscriptionsrepository "github.com/complexus-tech/projects-api/internal/modules/subscriptions/repository"
	usersdomain "github.com/complexus-tech/projects-api/internal/modules/users/domain"
	usersrepository "github.com/complexus-tech/projects-api/internal/modules/users/repository"
	"github.com/complexus-tech/projects-api/internal/testkit"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestInternalAlertSignupTransactionsAndConcurrentClaims(t *testing.T) {
	db := testkit.NewPostgres(t)
	ctx := t.Context()
	users := usersrepository.New(db.Pool)
	user, err := users.Create(ctx, usersdomain.User{Username: "newuser", Email: "new@example.test", FullName: "New User", Timezone: "UTC", LastLoginAt: time.Now()})
	require.NoError(t, err)
	_, err = users.Create(ctx, usersdomain.User{Username: "again", Email: user.Email, FullName: "Again", Timezone: "UTC", LastLoginAt: time.Now()})
	require.ErrorIs(t, err, usersdomain.ErrEmailTaken)
	external := usersdomain.ExternalIdentityInput{Provider: "microsoft", Issuer: "https://login.microsoftonline.com/test/v2.0", Subject: "new-subject", Email: "google@example.test", FullName: "Google User", Timezone: "UTC"}
	first, err := users.ResolveExternalIdentity(ctx, external)
	require.NoError(t, err)
	require.True(t, first.Created)
	second, err := users.ResolveExternalIdentity(ctx, external)
	require.NoError(t, err)
	require.False(t, second.Created)
	// Linking a provider to an existing account is not another signup.
	external.Subject, external.Email = "existing-subject", user.Email
	linked, err := users.ResolveExternalIdentity(ctx, external)
	require.NoError(t, err)
	require.False(t, linked.Created)
	// A rejected identity link rolls the account and its alert back together.
	invalidIdentity := external
	invalidIdentity.Provider, invalidIdentity.Subject, invalidIdentity.Email = "unsupported", "rollback", "rollback@example.test"
	_, err = users.ResolveExternalIdentity(ctx, invalidIdentity)
	require.Error(t, err)
	// Creating any number of workspaces produces no account alerts.
	for i := 0; i < 10; i++ {
		_, err = db.Pool.Exec(ctx, `INSERT INTO workspaces(name,slug) VALUES ('Workspace',$1)`, uuid.NewString())
		require.NoError(t, err)
	}
	var count int
	require.NoError(t, db.Pool.QueryRow(ctx, `SELECT count(*) FROM internal_slack_alerts`).Scan(&count))
	require.Equal(t, 2, count)
	store := New(db.Pool)
	results := make(chan *slackdomain.InternalAlert, 8)
	errs := make(chan error, 8)
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			alert, err := store.ClaimInternalAlert(ctx, "T015B85FC6R", "C014XSVSRF1")
			results <- alert
			errs <- err
		}()
	}
	wg.Wait()
	close(results)
	close(errs)
	for err := range errs {
		require.NoError(t, err)
	}
	claims := make(map[uuid.UUID]*slackdomain.InternalAlert)
	for alert := range results {
		if alert != nil {
			require.NotContains(t, claims, alert.ID)
			claims[alert.ID] = alert
		}
	}
	require.Len(t, claims, 2)
	for _, alert := range claims {
		// A stale claim cannot acknowledge a newer owner's delivery.
		_, err = db.Pool.Exec(ctx, `UPDATE internal_slack_alerts SET lease_until=now()-interval '1 second' WHERE id=$1`, alert.ID)
		require.NoError(t, err)
		wrongDestination, err := store.ClaimInternalAlert(ctx, "TOTHER", "COTHER")
		require.NoError(t, err)
		require.Nil(t, wrongDestination)
		recovered, err := store.ClaimInternalAlert(ctx, "T015B85FC6R", "C014XSVSRF1")
		require.NoError(t, err)
		require.NotNil(t, recovered)
		require.Equal(t, alert.ID, recovered.ID)
		require.ErrorIs(t, store.CompleteInternalAlert(ctx, *alert, "stale"), slackdomain.ErrInternalAlertLeaseLost)
		require.NoError(t, store.CompleteInternalAlert(ctx, *recovered, "171.234"))
		var payload string
		require.NoError(t, db.Pool.QueryRow(ctx, `SELECT CAST(payload AS text) FROM internal_slack_alerts WHERE id=$1`, alert.ID).Scan(&payload))
		require.Equal(t, "{}", payload)
	}
	empty, err := store.ClaimInternalAlert(ctx, "T015B85FC6R", "C014XSVSRF1")
	require.NoError(t, err)
	require.Nil(t, empty)
}

func TestInternalAlertPaymentAtomicityAndInvoiceDeduplication(t *testing.T) {
	db := testkit.NewPostgres(t)
	ctx := t.Context()
	workspaceID := uuid.New()
	_, err := db.Pool.Exec(ctx, `INSERT INTO workspaces(workspace_id,name,slug) VALUES ($1,'Customer Workspace','customer-workspace')`, workspaceID)
	require.NoError(t, err)
	_, err = db.Pool.Exec(ctx, `INSERT INTO workspace_subscriptions(workspace_id,stripe_customer_id,stripe_subscription_id,subscription_status,subscription_tier,seat_count) VALUES ($1,'cus_internal_test','sub_internal_test','active','pro',2)`, workspaceID)
	require.NoError(t, err)
	repo := subscriptionsrepository.New(db.Pool)
	invoice := subscriptionsdomain.SubscriptionInvoice{WorkspaceID: workspaceID, StripeInvoiceID: "in_internal_test", AmountPaid: 49, InvoiceDate: time.Now().UTC(), Status: "paid", SeatsCount: 2}
	require.NoError(t, repo.UpsertStripeInvoice(ctx, "cus_internal_test", invoice))
	var count int
	require.NoError(t, db.Pool.QueryRow(ctx, `SELECT count(*) FROM internal_slack_alerts`).Scan(&count))
	require.Zero(t, count)
	invoice.CollectedPayment = &subscriptionsdomain.CollectedPayment{AmountMinor: 4900, Currency: "usd", Email: "buyer@example.test"}
	// A forged customer-to-workspace binding must create neither an invoice nor an alert.
	badInvoice := invoice
	badInvoice.StripeInvoiceID = "in_wrong_customer"
	require.Error(t, repo.UpsertStripeInvoice(ctx, "cus_wrong", badInvoice))
	// Outbox failure must roll back the invoice write, leaving Stripe free to retry.
	_, err = db.Pool.Exec(ctx, `ALTER TABLE internal_slack_alerts ADD CONSTRAINT test_reject_payment CHECK (kind <> 'payment_received')`)
	require.NoError(t, err)
	failedInvoice := invoice
	failedInvoice.StripeInvoiceID = "in_rollback"
	require.Error(t, repo.UpsertStripeInvoice(ctx, "cus_internal_test", failedInvoice))
	require.NoError(t, db.Pool.QueryRow(ctx, `SELECT count(*) FROM subscription_invoices WHERE stripe_invoice_id='in_rollback'`).Scan(&count))
	require.Zero(t, count)
	_, err = db.Pool.Exec(ctx, `ALTER TABLE internal_slack_alerts DROP CONSTRAINT test_reject_payment`)
	require.NoError(t, err)
	for i := 0; i < 3; i++ {
		require.NoError(t, repo.UpsertStripeInvoice(ctx, "cus_internal_test", invoice))
	}
	invoice.CollectedPayment = nil // a later invoice.paid must not reset the receipt
	require.NoError(t, repo.UpsertStripeInvoice(ctx, "cus_internal_test", invoice))
	require.NoError(t, db.Pool.QueryRow(ctx, `SELECT count(*) FROM internal_slack_alerts`).Scan(&count))
	require.Equal(t, 1, count)
	var payload string
	require.NoError(t, db.Pool.QueryRow(ctx, `SELECT CAST(payload AS text) FROM internal_slack_alerts`).Scan(&payload))
	require.Contains(t, payload, `"amount_minor": 4900`)
	require.Contains(t, payload, `"currency": "usd"`)
	require.Contains(t, payload, "Customer Workspace")
	store := New(db.Pool)
	alert, err := store.ClaimInternalAlert(ctx, "T015B85FC6R", "C014XSVSRF1")
	require.NoError(t, err)
	require.NotNil(t, alert)
	require.NoError(t, store.RetryInternalAlert(ctx, *alert, time.Now().Add(time.Hour)))
	pending, err := store.ClaimInternalAlert(ctx, "T015B85FC6R", "C014XSVSRF1")
	require.NoError(t, err)
	require.Nil(t, pending)
}
