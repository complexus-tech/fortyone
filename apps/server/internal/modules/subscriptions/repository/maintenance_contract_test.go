package subscriptionsrepository

import (
	"context"
	"math"
	"os"
	"strings"
	"testing"
	"time"

	subscriptionssql "github.com/complexus-tech/projects-api/internal/modules/subscriptions/repository/sqlc"
	"github.com/complexus-tech/projects-api/internal/platform/safecast"
	"github.com/stretchr/testify/require"
)

func TestStripeWebhookMaintenanceQueryPurgesOnlyBoundedTerminalReceipts(t *testing.T) {
	t.Parallel()

	source, err := os.ReadFile("queries/maintenance.sql")
	if err != nil {
		t.Fatalf("read subscription maintenance queries: %v", err)
	}
	query := strings.Join(strings.Fields(strings.ToLower(string(source))), " ")

	for _, contract := range []string{
		"-- name: purgeterminalstripewebhookevents :execrows",
		"with candidates as materialized",
		"webhook.processing_state = 'processed'",
		"webhook.processed_at at time zone 'utc' < sqlc.arg(terminal_before)",
		"webhook.processing_state = 'failed'",
		"webhook.failed_at < sqlc.arg(terminal_before)",
		"case webhook.processing_state when 'processed' then webhook.processed_at at time zone 'utc' else webhook.failed_at end, webhook.event_id",
		"limit cast(sqlc.arg(batch_size) as integer)",
		"for update of webhook skip locked",
		"delete from public.stripe_webhook_events as webhook using candidates",
		"webhook.processing_state in ('processed', 'failed')",
	} {
		if !strings.Contains(query, contract) {
			t.Errorf("Stripe webhook maintenance query is missing contract %q", contract)
		}
	}

	if strings.Contains(query, "'processing'") {
		t.Fatal("Stripe webhook maintenance must never select or delete processing leases")
	}
	if strings.Contains(query, "coalesce(") {
		t.Fatal("terminal webhook retention must use state-specific terminal timestamps")
	}
	if strings.Contains(query, "current_timestamp") || strings.Contains(query, "now()") {
		t.Fatal("Stripe webhook retention cutoff must be supplied by the application clock")
	}
}

func TestStripeWebhookMaintenanceRepositoryMapsUTCAndBatchToSQLC(t *testing.T) {
	t.Parallel()

	queries := &stripeWebhookMaintenanceQueries{rows: 4}
	repository := &Repository{queries: queries}
	terminalBefore := time.Date(2026, time.July, 29, 10, 15, 0, 0, time.FixedZone("CAT", 2*60*60))

	deleted, err := repository.PurgeTerminalStripeWebhookEvents(context.Background(), terminalBefore, 500)

	require.NoError(t, err)
	require.Equal(t, int64(4), deleted)
	require.Equal(t, 1, queries.calls)
	require.NotNil(t, queries.params.TerminalBefore)
	require.Equal(t, terminalBefore.UTC(), *queries.params.TerminalBefore)
	require.Equal(t, int32(500), queries.params.BatchSize)
}

func TestStripeWebhookMaintenanceRepositoryRejectsBatchOutsideSQLCRange(t *testing.T) {
	t.Parallel()

	queries := &stripeWebhookMaintenanceQueries{}
	repository := &Repository{queries: queries}

	_, err := repository.PurgeTerminalStripeWebhookEvents(
		context.Background(),
		time.Date(2026, time.July, 29, 8, 15, 0, 0, time.UTC),
		int(math.MaxInt32)+1,
	)

	require.ErrorIs(t, err, safecast.ErrOutOfRange)
	require.Zero(t, queries.calls)
}

type stripeWebhookMaintenanceQueries struct {
	subscriptionssql.Querier
	params subscriptionssql.PurgeTerminalStripeWebhookEventsParams
	rows   int64
	err    error
	calls  int
}

func (queries *stripeWebhookMaintenanceQueries) PurgeTerminalStripeWebhookEvents(
	_ context.Context,
	params subscriptionssql.PurgeTerminalStripeWebhookEventsParams,
) (int64, error) {
	queries.calls++
	queries.params = params
	return queries.rows, queries.err
}
