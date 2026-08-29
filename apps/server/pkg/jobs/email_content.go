package jobs

import (
	"context"
	"fmt"
	"html"
	"strings"
	"sync"
	"time"

	"github.com/complexus-tech/projects-api/pkg/emailcopy"
	"github.com/complexus-tech/projects-api/pkg/emailthread"
	"github.com/complexus-tech/projects-api/pkg/mailer"
	"github.com/google/uuid"
	htmlparser "golang.org/x/net/html"
)

const guidanceEmailBatchConcurrency = 6

const guidanceEmailRecipientAttempts = 2

func nonEmptyFactTokens(values ...string) []string {
	tokens := make([]string, 0, len(values))
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			tokens = append(tokens, trimmed)
		}
	}
	return tokens
}

// deadlineSemanticFactTokens protects the relationship between a deadline's
// label, distance, and date. Literal numeric validation alone cannot detect a
// model swapping two dates or assigning a day count to the wrong field.
func deadlineSemanticFactTokens(deadlineStatus string, daysDifference int, formattedDate string) []string {
	switch deadlineStatus {
	case "overdue":
		return nonEmptyFactTokens(
			fmt.Sprintf("%d %s overdue", daysDifference, pluralize(daysDifference, "day", "days")),
			"due date is "+formattedDate,
		)
	case "due_today":
		return nonEmptyFactTokens("due today, " + formattedDate)
	case "due_tomorrow":
		return nonEmptyFactTokens("due tomorrow, " + formattedDate)
	case "future":
		return nonEmptyFactTokens("on schedule")
	default:
		return nonEmptyFactTokens("due on " + formattedDate)
	}
}

type emailCopyDestination struct {
	Label string
	URL   string
}

type guidanceEmailBatchResult struct {
	Processed bool
	Sent      bool
	Retryable bool
	Err       error
}

func guidanceEmailBatchFailureCount(results []guidanceEmailBatchResult) int {
	failures := 0
	for _, result := range results {
		if result.Err != nil {
			failures++
		}
	}
	return failures
}

// processGuidanceEmailRecipient retries only failures that happened before a
// guidance message could be durably frozen. Once thread preparation or SMTP
// begins, regenerating Luna copy under the same Message-ID could diverge from
// immutable conversation history, so callers deliberately leave those errors
// non-retryable and surface them for operational follow-up.
func processGuidanceEmailRecipient(
	ctx context.Context,
	process func(context.Context) guidanceEmailBatchResult,
) guidanceEmailBatchResult {
	var result guidanceEmailBatchResult
	for attempt := 1; attempt <= guidanceEmailRecipientAttempts; attempt++ {
		result = process(ctx)
		if result.Err == nil || !result.Retryable || ctx.Err() != nil {
			return result
		}
		if attempt < guidanceEmailRecipientAttempts {
			timer := time.NewTimer(time.Duration(attempt) * 100 * time.Millisecond)
			select {
			case <-ctx.Done():
				timer.Stop()
				return guidanceEmailBatchResult{Err: ctx.Err()}
			case <-timer.C:
			}
		}
	}
	return result
}

// processGuidanceEmailBatch bounds the latency and load of per-recipient copy
// generation. Each result index corresponds to the same input index, which
// keeps counters and error reporting deterministic without shared mutations.
func processGuidanceEmailBatch[T any](
	ctx context.Context,
	items []T,
	process func(context.Context, T) guidanceEmailBatchResult,
) ([]guidanceEmailBatchResult, error) {
	results := make([]guidanceEmailBatchResult, len(items))
	if len(items) == 0 {
		return results, nil
	}

	workerCount := min(guidanceEmailBatchConcurrency, len(items))
	indexes := make(chan int)
	var workers sync.WaitGroup
	workers.Add(workerCount)
	for range workerCount {
		go func() {
			defer workers.Done()
			for index := range indexes {
				if ctx.Err() != nil {
					continue
				}
				results[index] = process(ctx, items[index])
			}
		}()
	}

enqueue:
	for index := range items {
		select {
		case indexes <- index:
		case <-ctx.Done():
			break enqueue
		}
	}
	close(indexes)
	workers.Wait()
	return results, ctx.Err()
}

func guidanceEmailMessageID(kind string, workspaceID, userID uuid.UUID, deliveryDate time.Time) string {
	return fmt.Sprintf("<%s-%s-%s-%s@fortyone.app>", kind, workspaceID, userID, deliveryDate.UTC().Format("2006-01-02"))
}

func prepareGuidanceThread(
	ctx context.Context,
	threader emailthread.GuidancePreparer,
	input emailthread.GuidanceInput,
) (string, error) {
	if threader == nil {
		return "", nil
	}
	prepared, err := threader.PrepareGuidance(ctx, input)
	if err != nil {
		return "", err
	}
	return prepared.ReplyTo, nil
}

func guidancePlainText(heading, htmlContent, ctaLabel, ctaURL string) string {
	plain := emailHTMLPlainText(htmlContent)
	sections := nonEmptyFactTokens(strings.TrimSpace(heading), plain)
	if label := strings.TrimSpace(ctaLabel); label != "" {
		if url := strings.TrimSpace(ctaURL); url != "" {
			sections = append(sections, label+": "+url)
		} else {
			sections = append(sections, label)
		}
	}
	return strings.Join(sections, "\n\n")
}

func emailHTMLPlainText(value string) string {
	tokenizer := htmlparser.NewTokenizer(strings.NewReader(value))
	var text strings.Builder
	for {
		switch tokenizer.Next() {
		case htmlparser.ErrorToken:
			return strings.Join(strings.Fields(text.String()), " ")
		case htmlparser.TextToken:
			if text.Len() > 0 {
				text.WriteByte(' ')
			}
			text.Write(tokenizer.Text())
		}
	}
}

func formatCompactNotificationRows(intro string, rows []string) string {
	textStyle := mailer.EmailStyleString("notificationText")
	listStyle := mailer.EmailStyleString("notificationList")
	messageStyle := mailer.EmailStyleString("notificationMessage")

	content := fmt.Sprintf(`
		<div style="%s">
			<p style="%s">%s</p>
			<div style="%s">
	`, textStyle, textStyle, html.EscapeString(intro), listStyle)

	for index, row := range rows {
		itemStyle := mailer.EmailStyleString("notificationItem")
		if index == 0 {
			itemStyle = mailer.EmailStyleString("notificationItemFirst")
		}
		content += fmt.Sprintf(`
			<div style="%s">
				<p style="%s">%s</p>
			</div>
		`, itemStyle, messageStyle, row)
	}

	return content + "</div></div>"
}

// capGuidanceEmailDetailRows reserves one row for the summary and bounds the
// remaining product detail to the same budget used by generated guidance copy.
func capGuidanceEmailDetailRows(rows []string) ([]string, int) {
	limit := maxGuidanceEmailRows - 1
	if len(rows) <= limit {
		return rows, 0
	}
	return rows[:limit], len(rows) - limit
}

func formatEmailStrong(value string) string {
	return fmt.Sprintf(`<strong style="%s">%s</strong>`, mailer.EmailStyleString("detailValue"), html.EscapeString(value))
}

func formatEmailLink(url string, label string) string {
	return fmt.Sprintf(`<a href="%s" style="%s">%s</a>`, html.EscapeString(url), mailer.EmailStyleString("notificationLink"), html.EscapeString(label))
}

func renderGeneratedEmailContent(output emailcopy.Output, destinations map[string]emailCopyDestination) (string, error) {
	textStyle := mailer.EmailStyleString("notificationText")
	listStyle := mailer.EmailStyleString("notificationList")
	messageStyle := mailer.EmailStyleString("notificationMessage")

	paragraphs := []string{output.Intro.Text}
	if output.SenderProse != nil {
		paragraphs = append(paragraphs, output.SenderProse.Text)
	}
	if output.FeedbackThemeSummary != nil {
		paragraphs = append(paragraphs, output.FeedbackThemeSummary.Text)
	}
	if output.ReplyPrompt != nil {
		paragraphs = append(paragraphs, output.ReplyPrompt.Text)
	}

	var content strings.Builder
	content.WriteString(fmt.Sprintf(`<div style="%s">`, textStyle))
	for _, paragraph := range paragraphs {
		content.WriteString(fmt.Sprintf(`<p style="%s">%s</p>`, textStyle, html.EscapeString(paragraph)))
	}
	if len(output.Rows) == 0 {
		content.WriteString("</div>")
		return content.String(), nil
	}

	content.WriteString(fmt.Sprintf(`<div style="%s">`, listStyle))
	for index, row := range output.Rows {
		rowHTML := html.EscapeString(row.Text)
		if destination, exists := destinations[row.ReferenceID]; exists && destination.URL != "" && destination.Label != "" {
			escapedLabel := html.EscapeString(destination.Label)
			if !strings.Contains(rowHTML, escapedLabel) {
				return "", fmt.Errorf("generated row %q does not contain its destination label", row.ReferenceID)
			}
			rowHTML = strings.Replace(rowHTML, escapedLabel, formatEmailLink(destination.URL, destination.Label), 1)
		}

		itemStyle := mailer.EmailStyleString("notificationItem")
		if index == 0 {
			itemStyle = mailer.EmailStyleString("notificationItemFirst")
		}
		content.WriteString(fmt.Sprintf(`<div style="%s"><p style="%s">%s</p></div>`, itemStyle, messageStyle, rowHTML))
	}
	content.WriteString("</div></div>")
	return content.String(), nil
}

func appendGuidanceReplyPrompt(content, prompt string) string {
	prompt = strings.TrimSpace(prompt)
	if prompt == "" {
		return content
	}
	return content + fmt.Sprintf(
		`<div style="%s"><p style="%s">%s</p></div>`,
		mailer.EmailStyleString("notificationText"),
		mailer.EmailStyleString("notificationText"),
		html.EscapeString(prompt),
	)
}

func generatedPrimaryCTA(output emailcopy.Output, destinations map[string]emailCopyDestination) (string, string, bool) {
	for _, cta := range output.CTAs {
		destination, exists := destinations[cta.ReferenceID]
		if !exists || strings.TrimSpace(destination.URL) == "" {
			continue
		}
		return cta.Label, destination.URL, true
	}
	return "", "", false
}
