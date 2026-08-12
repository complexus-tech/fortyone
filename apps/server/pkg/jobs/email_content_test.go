package jobs

import (
	"context"
	"testing"

	"github.com/complexus-tech/projects-api/pkg/emailcopy"
	"github.com/stretchr/testify/require"
)

func TestGuidanceEmailBatchFailureCount(t *testing.T) {
	require.Equal(t, 2, guidanceEmailBatchFailureCount([]guidanceEmailBatchResult{
		{Processed: true},
		{Err: context.DeadlineExceeded},
		{Err: context.Canceled},
	}))
	require.Zero(t, guidanceEmailBatchFailureCount([]guidanceEmailBatchResult{{Processed: true}}))
}

func TestProcessGuidanceEmailRecipientRetriesOnce(t *testing.T) {
	attempts := 0
	result := processGuidanceEmailRecipient(context.Background(), func(context.Context) guidanceEmailBatchResult {
		attempts++
		if attempts == 1 {
			return guidanceEmailBatchResult{Err: context.DeadlineExceeded}
		}
		return guidanceEmailBatchResult{Processed: true, Sent: true}
	})

	require.Equal(t, guidanceEmailRecipientAttempts, attempts)
	require.NoError(t, result.Err)
	require.True(t, result.Sent)
}

func TestProcessGuidanceEmailRecipientStopsAfterBoundedAttempts(t *testing.T) {
	attempts := 0
	result := processGuidanceEmailRecipient(context.Background(), func(context.Context) guidanceEmailBatchResult {
		attempts++
		return guidanceEmailBatchResult{Err: context.DeadlineExceeded}
	})

	require.Equal(t, guidanceEmailRecipientAttempts, attempts)
	require.ErrorIs(t, result.Err, context.DeadlineExceeded)
}

func TestProcessGuidanceEmailBatchBoundsRecipientConcurrency(t *testing.T) {
	items := make([]int, guidanceEmailBatchConcurrency*2)
	started := make(chan struct{}, len(items))
	release := make(chan struct{})
	type batchOutcome struct {
		results []guidanceEmailBatchResult
		err     error
	}
	done := make(chan batchOutcome, 1)

	go func() {
		results, err := processGuidanceEmailBatch(context.Background(), items, func(context.Context, int) guidanceEmailBatchResult {
			started <- struct{}{}
			<-release
			return guidanceEmailBatchResult{Processed: true, Sent: true}
		})
		done <- batchOutcome{results: results, err: err}
	}()

	for range guidanceEmailBatchConcurrency {
		<-started
	}
	select {
	case <-started:
		t.Fatal("guidance batch exceeded its concurrency limit")
	default:
	}
	close(release)

	outcome := <-done
	require.NoError(t, outcome.err)
	require.Len(t, outcome.results, len(items))
	for _, result := range outcome.results {
		require.True(t, result.Processed)
		require.True(t, result.Sent)
	}
}

func TestRenderGeneratedEmailContentLinksOnlyTrustedCanonicalLabels(t *testing.T) {
	theme := &emailcopy.GroundedText{Text: "Requests cluster around faster reporting.", ReferenceIDs: []string{"feedback_a"}}
	output := emailcopy.Output{
		Intro: emailcopy.GroundedText{Text: "Three customer notes point to one useful review.", ReferenceIDs: []string{"summary"}},
		Rows: []emailcopy.Row{{
			ReferenceID: "feedback_a",
			Text:        "Review Faster exports first; it is still pending.",
		}},
		FeedbackThemeSummary: theme,
	}

	rendered, err := renderGeneratedEmailContent(output, map[string]emailCopyDestination{
		"feedback_a": {Label: "Faster exports", URL: "https://acme.fortyone.app/feedback/faster-exports"},
	})

	require.NoError(t, err)
	require.Contains(t, rendered, "Requests cluster around faster reporting.")
	require.Contains(t, rendered, `href="https://acme.fortyone.app/feedback/faster-exports"`)
	require.Contains(t, rendered, ">Faster exports</a>")
}

func TestRenderGeneratedEmailContentRejectsMissingCanonicalLabel(t *testing.T) {
	output := emailcopy.Output{
		Intro: emailcopy.GroundedText{Text: "A useful review is ready.", ReferenceIDs: []string{"summary"}},
		Rows: []emailcopy.Row{{
			ReferenceID: "objective_a",
			Text:        "Review the objective today.",
		}},
	}

	_, err := renderGeneratedEmailContent(output, map[string]emailCopyDestination{
		"objective_a": {Label: "Grow enterprise revenue", URL: "https://acme.fortyone.app/objectives/a"},
	})

	require.ErrorContains(t, err, "does not contain its destination label")
}
