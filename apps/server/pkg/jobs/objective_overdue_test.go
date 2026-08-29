package jobs

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/mailer"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

type failingObjectiveMailer struct {
	err error
}

func (m failingObjectiveMailer) Send(context.Context, mailer.Email) error {
	return m.err
}

func (m failingObjectiveMailer) SendTemplated(context.Context, mailer.TemplatedEmail) error {
	return m.err
}

func TestWaitForNextOverdueGuidanceBatchReturnsWhenContextIsCancelled(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	startedAt := time.Now()
	err := waitForNextOverdueGuidanceBatch(ctx)

	require.ErrorIs(t, err, context.Canceled)
	require.Less(t, time.Since(startedAt), overdueGuidanceBatchDelay)
}

func TestParseKeyResultsRejectsMalformedOrPartiallyDecodedJSON(t *testing.T) {
	t.Parallel()

	malformed := `[{"name":"must not survive"}, invalid]`
	require.Nil(t, parseKeyResults(malformed))
	require.Nil(t, parseKeyResults(""))
	require.Nil(t, parseKeyResults("[]"))

	keyResultID := uuid.New()
	parsed := parseKeyResults(fmt.Sprintf(`[{"id":%q,"name":"Typed result"}]`, keyResultID))
	require.Len(t, parsed, 1)
	require.Equal(t, keyResultID, parsed[0].ID)
	require.Equal(t, "Typed result", parsed[0].Name)
}

func TestObjectiveOverdueDeliveryOmitsRecipientEmailFromObservability(t *testing.T) {
	t.Parallel()

	const recipientEmail = "private-objective-lead@example.com"
	leadID := uuid.New()
	workspaceID := uuid.New()

	spanRecorder := tracetest.NewSpanRecorder()
	tracerProvider := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(spanRecorder))
	t.Cleanup(func() {
		require.NoError(t, tracerProvider.Shutdown(context.Background()))
	})
	ctx, _ := tracerProvider.Tracer("objective-overdue-test").Start(context.Background(), "delivery")

	var logOutput bytes.Buffer
	log := logger.NewWithJSON(&logOutput, slog.LevelDebug, "objective-overdue-test")
	mailerService := &recordingMailer{}
	objective := OverdueObjective{
		ID:             uuid.New(),
		Name:           "Protect launch readiness",
		EndDate:        time.Date(2026, time.August, 28, 0, 0, 0, 0, time.UTC),
		LeadUserID:     leadID,
		LeadEmail:      recipientEmail,
		LeadName:       "Objective lead",
		WorkspaceID:    workspaceID,
		WorkspaceName:  "Product",
		WorkspaceSlug:  "product",
		TeamID:         uuid.New(),
		DeadlineStatus: "due_today",
		KeyResults:     "[]",
	}

	err := sendObjectiveOverdueEmailForLead(ctx, log, mailerService, nil, nil, []OverdueObjective{objective})
	require.NoError(t, err)
	require.Len(t, mailerService.templatedEmails, 1)
	require.NotContains(t, logOutput.String(), recipientEmail)
	require.Contains(t, logOutput.String(), leadID.String())

	endedSpans := spanRecorder.Ended()
	require.Len(t, endedSpans, 1)
	for _, event := range endedSpans[0].Events() {
		for _, attribute := range event.Attributes {
			require.NotEqual(t, "lead_email", string(attribute.Key))
			require.False(t, strings.Contains(fmt.Sprint(attribute.Value.AsInterface()), recipientEmail))
		}
	}
}

func TestObjectiveOverdueDeliverySanitizesProviderErrorsInTraces(t *testing.T) {
	t.Parallel()

	const recipientEmail = "private-objective-lead@example.com"
	spanRecorder := tracetest.NewSpanRecorder()
	tracerProvider := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(spanRecorder))
	t.Cleanup(func() {
		require.NoError(t, tracerProvider.Shutdown(context.Background()))
	})
	ctx, _ := tracerProvider.Tracer("objective-overdue-test").Start(context.Background(), "delivery")
	objective := OverdueObjective{
		ID:             uuid.New(),
		Name:           "Protect launch readiness",
		EndDate:        time.Date(2026, time.August, 28, 0, 0, 0, 0, time.UTC),
		LeadUserID:     uuid.New(),
		LeadEmail:      recipientEmail,
		LeadName:       "Objective lead",
		WorkspaceID:    uuid.New(),
		WorkspaceName:  "Product",
		WorkspaceSlug:  "product",
		TeamID:         uuid.New(),
		DeadlineStatus: "due_today",
		KeyResults:     "[]",
	}
	providerErr := errors.New("provider rejected recipient " + recipientEmail)

	err := sendObjectiveOverdueEmailForLead(
		ctx,
		logger.NewWithJSON(&bytes.Buffer{}, slog.LevelDebug, "objective-overdue-test"),
		failingObjectiveMailer{err: providerErr},
		nil,
		nil,
		[]OverdueObjective{objective},
	)
	require.ErrorIs(t, err, providerErr)

	endedSpans := spanRecorder.Ended()
	require.Len(t, endedSpans, 1)
	for _, event := range endedSpans[0].Events() {
		for _, attribute := range event.Attributes {
			require.False(t, strings.Contains(fmt.Sprint(attribute.Value.AsInterface()), recipientEmail))
		}
	}
}
