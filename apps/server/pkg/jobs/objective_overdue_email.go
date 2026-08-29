package jobs

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"strconv"
	"strings"
	"time"

	"github.com/complexus-tech/projects-api/pkg/emailcopy"
	"github.com/complexus-tech/projects-api/pkg/emailthread"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/mailer"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

var errObjectiveOverdueEmailDelivery = errors.New("objective overdue email delivery failed")

// sendObjectiveOverdueEmailForLead sends an email to a lead about their overdue objectives
func sendObjectiveOverdueEmailForLead(ctx context.Context, log *logger.Logger, mailerService mailer.Service, copyGenerator emailcopy.Generator, threader emailthread.GuidancePreparer, objectives []OverdueObjective) error {
	ctx, span := web.AddSpan(ctx, "jobs.sendObjectiveOverdueEmailForLead")
	defer span.End()

	if len(objectives) == 0 {
		return nil
	}

	if strings.TrimSpace(objectives[0].LeadEmail) == "" {
		log.Info(ctx, "Skipping objective overdue email because lead email is empty", "lead_id", objectives[0].LeadUserID)
		return nil
	}

	// Group objectives by deadline status
	var dueSoonObjectives, dueTodayObjectives, overdueObjectives []OverdueObjective

	for _, objective := range objectives {
		// Check if objective has key results (even if objective itself isn't overdue)
		hasKeyResults := false
		if keyResults := parseKeyResults(objective.KeyResults); len(keyResults) > 0 {
			hasKeyResults = true
		}

		switch objective.DeadlineStatus {
		case "due_in_7_days", "due_tomorrow":
			dueSoonObjectives = append(dueSoonObjectives, objective)
		case "due_today":
			dueTodayObjectives = append(dueTodayObjectives, objective)
		case "overdue":
			overdueObjectives = append(overdueObjectives, objective)
		case "future":
			// If objective is on schedule but has key results, add it to due soon
			if hasKeyResults {
				dueSoonObjectives = append(dueSoonObjectives, objective)
			}
		}
	}

	// Use data from first objective for common fields
	firstObjective := objectives[0]
	workspaceURL := fmt.Sprintf("https://%s.fortyone.app", firstObjective.WorkspaceSlug)

	totalCount := len(dueSoonObjectives) + len(dueTodayObjectives) + len(overdueObjectives)
	itemText := "objective"
	if totalCount > 1 {
		itemText = "objectives"
	}
	title := fmt.Sprintf("%d %s need attention", totalCount, itemText)
	heading := title
	emailContent := formatObjectiveOverdueEmailContent(firstObjective, dueSoonObjectives, dueTodayObjectives, overdueObjectives, workspaceURL)
	emailContent = appendGuidanceReplyPrompt(emailContent, "I’m Maya, your AI agent. Reply to this email with the latest health, value, or blocker, and I’ll help you keep the objective current.")
	ctaURL := fmt.Sprintf("%s/roadmap", workspaceURL)
	ctaLabel := "View objectives"

	if copyGenerator != nil {
		orderedObjectives := make([]OverdueObjective, 0, len(objectives))
		orderedObjectives = append(orderedObjectives, overdueObjectives...)
		orderedObjectives = append(orderedObjectives, dueTodayObjectives...)
		orderedObjectives = append(orderedObjectives, dueSoonObjectives...)
		request, destinations := objectiveOverdueEmailCopyRequest(orderedObjectives, workspaceURL, ctaURL)
		generated, err := copyGenerator.Generate(ctx, request)
		if err != nil {
			log.Warn(ctx, "Falling back to deterministic objective guidance copy", "lead_id", firstObjective.LeadUserID, "workspace_id", firstObjective.WorkspaceID, "error", err)
		} else if generatedContent, renderErr := renderGeneratedEmailContent(generated, destinations); renderErr != nil {
			log.Warn(ctx, "Falling back to deterministic objective guidance copy after render validation", "lead_id", firstObjective.LeadUserID, "workspace_id", firstObjective.WorkspaceID, "error", renderErr)
		} else if generatedCTALabel, generatedCTAURL, ok := generatedPrimaryCTA(generated, destinations); !ok {
			log.Warn(ctx, "Falling back to deterministic objective guidance copy because no trusted CTA was generated", "lead_id", firstObjective.LeadUserID, "workspace_id", firstObjective.WorkspaceID)
		} else {
			title = generated.Subject.Text
			heading = generated.H1.Text
			emailContent = generatedContent
			ctaLabel = generatedCTALabel
			ctaURL = generatedCTAURL
		}
	}

	data := map[string]any{
		"UserName":                 firstObjective.LeadName,
		"UserEmail":                firstObjective.LeadEmail,
		"WorkspaceName":            firstObjective.WorkspaceName,
		"WorkspaceURL":             workspaceURL,
		"NotificationTitle":        heading,
		"NotificationMessage":      emailContent,
		"NotificationType":         "reminders",
		"NotificationCTAURL":       ctaURL,
		"NotificationCTALabel":     ctaLabel,
		"NotificationsSettingsURL": fmt.Sprintf("%s/settings/account/notifications", workspaceURL),
	}
	messageID := guidanceEmailMessageID("objective-guidance", firstObjective.WorkspaceID, firstObjective.LeadUserID, time.Now())
	targets := make([]emailthread.TargetContext, 0, len(objectives)*2)
	for _, objective := range objectives {
		targets = append(targets, emailthread.TargetContext{
			Kind:        "objective",
			ID:          objective.ID,
			TeamID:      objective.TeamID,
			DisplayName: objective.Name,
		})
		for _, keyResult := range parseKeyResults(objective.KeyResults) {
			if keyResult.ID == uuid.Nil {
				continue
			}
			targets = append(targets, emailthread.TargetContext{
				Kind:        "key_result",
				ID:          keyResult.ID,
				TeamID:      objective.TeamID,
				ParentID:    objective.ID,
				DisplayName: keyResult.Name,
			})
		}
	}
	threadContext, err := emailthread.EncodeThreadContext(emailthread.ThreadContext{
		Source:        "objective_guidance",
		WorkspaceSlug: firstObjective.WorkspaceSlug,
		Targets:       targets,
	})
	if err != nil {
		return err
	}
	plainText := guidancePlainText(heading, emailContent, ctaLabel, ctaURL)
	replyTo, err := prepareGuidanceThread(ctx, threader, emailthread.GuidanceInput{
		WorkspaceID:       firstObjective.WorkspaceID,
		UserID:            firstObjective.LeadUserID,
		RecipientEmail:    firstObjective.LeadEmail,
		ExternalThreadID:  messageID,
		InternetMessageID: messageID,
		Subject:           title,
		Content:           plainText,
		Context:           threadContext,
	})
	if err != nil {
		return fmt.Errorf("prepare objective guidance reply thread: %w", err)
	}

	if err := mailerService.SendTemplated(ctx, mailer.TemplatedEmail{
		To:            []string{firstObjective.LeadEmail},
		Template:      "notifications/notification",
		Subject:       title,
		Data:          data,
		PlainTextBody: plainText,
		Sender:        mailer.SenderProfileMaya,
		ReplyTo:       replyTo,
		MessageID:     messageID,
	}); err != nil {
		span.RecordError(errObjectiveOverdueEmailDelivery, trace.WithAttributes(
			attribute.String("lead_id", firstObjective.LeadUserID.String()),
			attribute.String("workspace_id", firstObjective.WorkspaceID.String()),
			attribute.String("error.type", fmt.Sprintf("%T", err)),
		))
		return fmt.Errorf("failed to send objective overdue email: %w", err)
	}

	log.Info(ctx, "Successfully sent objective overdue email",
		"lead_id", firstObjective.LeadUserID,
		"workspace_id", firstObjective.WorkspaceID,
		"total_objectives", totalCount)

	span.AddEvent("email prepared", trace.WithAttributes(
		attribute.String("lead_id", firstObjective.LeadUserID.String()),
		attribute.String("workspace_id", firstObjective.WorkspaceID.String()),
		attribute.Int("objectives.count", len(objectives)),
	))

	return nil
}

type objectiveEmailCopyFact struct {
	fact        emailcopy.Fact
	destination emailCopyDestination
}

func objectiveOverdueEmailCopyRequest(objectives []OverdueObjective, workspaceURL, ctaURL string) (emailcopy.Request, map[string]emailCopyDestination) {
	firstObjective := objectives[0]
	availableFacts := make([]objectiveEmailCopyFact, 0, len(objectives)*2)
	for _, objective := range objectives {
		objectiveReference := "objective:" + objective.ID.String()
		availableFacts = append(availableFacts, objectiveEmailCopyFact{
			fact: emailcopy.Fact{
				ReferenceID:     objectiveReference,
				Text:            objectiveEmailCopyFactText(objective),
				EntityTokens:    []string{objective.Name},
				ProtectedTokens: objectiveDeadlineProtectedTokens(objective),
				Required:        true,
			},
			destination: emailCopyDestination{
				Label: objective.Name,
				URL:   objectiveEmailURL(workspaceURL, objective.TeamID, objective.ID),
			},
		})
	}

	// Preserve at least one row per objective before using the remaining email
	// budget for key-result detail.
	for _, objective := range objectives {
		for _, keyResult := range parseKeyResults(objective.KeyResults) {
			if keyResult.ID == uuid.Nil {
				continue
			}
			keyResultReference := "key_result:" + keyResult.ID.String()
			availableFacts = append(availableFacts, objectiveEmailCopyFact{
				fact: emailcopy.Fact{
					ReferenceID:     keyResultReference,
					Text:            keyResultEmailCopyFactText(keyResult),
					EntityTokens:    []string{keyResult.Name},
					ProtectedTokens: keyResultDeadlineProtectedTokens(keyResult),
					Required:        true,
				},
				destination: emailCopyDestination{
					Label: keyResult.Name,
					URL:   keyResultEmailURL(workspaceURL, objective.TeamID, objective.ID, keyResult.ID),
				},
			})
		}
	}

	itemLimit := maxGuidanceEmailRows - 1
	if len(availableFacts) < itemLimit {
		itemLimit = len(availableFacts)
	}
	hiddenCount := len(availableFacts) - itemLimit
	summaryText := fmt.Sprintf("There are %d objectives that need attention, represented by %d objective and key-result signals.", len(objectives), len(availableFacts))
	summaryTokens := nonEmptyFactTokens(
		fmt.Sprintf("%d objectives that need attention", len(objectives)),
		fmt.Sprintf("%d objective and key-result signals", len(availableFacts)),
	)
	if hiddenCount > 0 {
		summaryText = fmt.Sprintf("There are %d objectives that need attention, represented by %d objective and key-result signals. This email includes %d signals; %d more %s available in objectives.", len(objectives), len(availableFacts), itemLimit, hiddenCount, pluralize(hiddenCount, "is", "are"))
		summaryTokens = append(summaryTokens,
			fmt.Sprintf("includes %d signals", itemLimit),
			fmt.Sprintf("%d more %s available in objectives", hiddenCount, pluralize(hiddenCount, "is", "are")),
		)
	}

	facts := []emailcopy.Fact{
		{
			ReferenceID:  "workspace_context",
			Text:         fmt.Sprintf("The workspace is named %s.", firstObjective.WorkspaceName),
			EntityTokens: []string{firstObjective.WorkspaceName},
		},
		{ReferenceID: "objective_summary", Text: summaryText, ProtectedTokens: summaryTokens, Required: true},
	}
	destinations := make(map[string]emailCopyDestination, itemLimit+1)
	for _, availableFact := range availableFacts[:itemLimit] {
		facts = append(facts, availableFact.fact)
		destinations[availableFact.fact.ReferenceID] = availableFact.destination
	}

	const actionReference = "review_objectives"
	destinations[actionReference] = emailCopyDestination{Label: "Objectives", URL: ctaURL}
	return emailcopy.Request{
		SafetyIdentifier: firstObjective.LeadUserID.String(),
		Purpose:          "Help an objective lead protect outcomes by reviewing objectives and key results with approaching or missed due dates.",
		ProductVoice:     mayaGuidanceProductVoice,
		Facts:            facts,
		Actions: []emailcopy.Action{
			{ReferenceID: actionReference, Description: "Open objectives to review progress, dates, and next actions."},
		},
		IncludeReplyPrompt: true,
	}, destinations
}

func objectiveEmailCopyFactText(objective OverdueObjective) string {
	objectiveContext := fmt.Sprintf("The objective %s", objective.Name)
	switch objective.DeadlineStatus {
	case "overdue":
		return fmt.Sprintf("%s is %d %s overdue; its due date is %s.", objectiveContext, objective.DaysDifference, pluralize(objective.DaysDifference, "day", "days"), objective.EndDate.Format("January 2, 2006"))
	case "due_today":
		return fmt.Sprintf("%s is due today, %s.", objectiveContext, objective.EndDate.Format("January 2, 2006"))
	case "due_tomorrow":
		return fmt.Sprintf("%s is due tomorrow, %s.", objectiveContext, objective.EndDate.Format("January 2, 2006"))
	case "future":
		return fmt.Sprintf("%s is on schedule, but at least one of its key results needs attention.", objectiveContext)
	default:
		return fmt.Sprintf("%s is due on %s.", objectiveContext, objective.EndDate.Format("January 2, 2006"))
	}
}

func objectiveDeadlineProtectedTokens(objective OverdueObjective) []string {
	return deadlineSemanticFactTokens(
		objective.DeadlineStatus,
		objective.DaysDifference,
		objective.EndDate.Format("January 2, 2006"),
	)
}

func keyResultEmailCopyFactText(keyResult OverdueKeyResult) string {
	keyResultContext := fmt.Sprintf("The key result %s", keyResult.Name)
	if progress := keyResultProgressText(keyResult); progress != "" {
		keyResultContext += " " + progress
	} else {
		keyResultContext += " is incomplete"
	}
	endDate := keyResult.EndDate
	if parsedEndDate, err := time.Parse("2006-01-02", keyResult.EndDate); err == nil {
		endDate = parsedEndDate.Format("January 2, 2006")
	}

	switch keyResult.DeadlineStatus {
	case "overdue":
		return fmt.Sprintf("%s and is %d %s overdue; its due date is %s.", keyResultContext, keyResult.DaysDifference, pluralize(keyResult.DaysDifference, "day", "days"), endDate)
	case "due_today":
		return fmt.Sprintf("%s and is due today, %s.", keyResultContext, endDate)
	case "due_tomorrow":
		return fmt.Sprintf("%s and is due tomorrow, %s.", keyResultContext, endDate)
	default:
		return fmt.Sprintf("%s and is due on %s.", keyResultContext, endDate)
	}
}

func keyResultDeadlineProtectedTokens(keyResult OverdueKeyResult) []string {
	endDate := keyResult.EndDate
	if parsedEndDate, err := time.Parse("2006-01-02", keyResult.EndDate); err == nil {
		endDate = parsedEndDate.Format("January 2, 2006")
	}
	tokens := deadlineSemanticFactTokens(keyResult.DeadlineStatus, keyResult.DaysDifference, endDate)
	if progress := keyResultProgressText(keyResult); progress != "" {
		tokens = append(tokens, progress)
	} else {
		tokens = append(tokens, "is incomplete")
	}
	return nonEmptyFactTokens(tokens...)
}

func keyResultProgressText(keyResult OverdueKeyResult) string {
	currentValue := strconv.FormatFloat(keyResult.CurrentValue, 'f', -1, 64)
	targetValue := strconv.FormatFloat(keyResult.TargetValue, 'f', -1, 64)
	switch keyResult.MeasurementType {
	case "percentage":
		return fmt.Sprintf("is at %s%% against a %s%% target", currentValue, targetValue)
	case "number":
		return fmt.Sprintf("is at %s against a target of %s", currentValue, targetValue)
	default:
		return ""
	}
}

func objectiveEmailURL(workspaceURL string, teamID, objectiveID uuid.UUID) string {
	return fmt.Sprintf("%s/teams/%s/objectives/%s", strings.TrimRight(workspaceURL, "/"), teamID.String(), objectiveID.String())
}

func keyResultEmailURL(workspaceURL string, teamID, objectiveID, keyResultID uuid.UUID) string {
	return fmt.Sprintf("%s?tab=overview&keyResultId=%s", objectiveEmailURL(workspaceURL, teamID, objectiveID), keyResultID.String())
}

// formatObjectiveOverdueEmailContent formats the email content
func formatObjectiveOverdueEmailContent(firstObjective OverdueObjective, dueSoonObjectives, dueTodayObjectives, overdueObjectives []OverdueObjective, workspaceURL string) string {
	totalItems := len(dueSoonObjectives) + len(dueTodayObjectives) + len(overdueObjectives)
	itemText := "objective"
	if totalItems > 1 {
		itemText = "objectives"
	}

	detailRows := make([]string, 0, totalItems)

	if len(dueSoonObjectives) > 0 {
		for _, objective := range dueSoonObjectives {
			// Check if this objective is actually due soon or just has key results
			objectiveLink := formatEmailLink(fmt.Sprintf("%s/teams/%s/objectives/%s", workspaceURL, objective.TeamID.String(), objective.ID.String()), objective.Name)
			if objective.DeadlineStatus == "future" {
				// Objective is on schedule but has key results
				detailRows = append(detailRows, fmt.Sprintf("Objective %s is on schedule, but key results need attention.", objectiveLink))
			} else {
				// Objective is actually due soon
				detailRows = append(detailRows, fmt.Sprintf("Objective %s is due %s.", objectiveLink, html.EscapeString(objective.EndDate.Format("January 2, 2006"))))
			}

			// Add key results if any
			if keyResults := parseKeyResults(objective.KeyResults); len(keyResults) > 0 {
				detailRows = append(detailRows, formatKeyResultEmailRows(keyResults, workspaceURL, objective.TeamID, objective.ID)...)
			}
		}
	}

	if len(dueTodayObjectives) > 0 {
		for _, objective := range dueTodayObjectives {
			objectiveLink := formatEmailLink(fmt.Sprintf("%s/teams/%s/objectives/%s", workspaceURL, objective.TeamID.String(), objective.ID.String()), objective.Name)
			detailRows = append(detailRows, fmt.Sprintf("Objective %s is due today.", objectiveLink))

			// Add key results if any
			if keyResults := parseKeyResults(objective.KeyResults); len(keyResults) > 0 {
				detailRows = append(detailRows, formatKeyResultEmailRows(keyResults, workspaceURL, objective.TeamID, objective.ID)...)
			}
		}
	}

	if len(overdueObjectives) > 0 {
		for _, objective := range overdueObjectives {
			daysText := "day"
			if objective.DaysDifference > 1 {
				daysText = "days"
			}
			objectiveLink := formatEmailLink(fmt.Sprintf("%s/teams/%s/objectives/%s", workspaceURL, objective.TeamID.String(), objective.ID.String()), objective.Name)
			detailRows = append(detailRows, fmt.Sprintf("Objective %s is %s overdue.", objectiveLink, formatEmailStrong(fmt.Sprintf("%d %s", objective.DaysDifference, daysText))))

			// Add key results if any
			if keyResults := parseKeyResults(objective.KeyResults); len(keyResults) > 0 {
				detailRows = append(detailRows, formatKeyResultEmailRows(keyResults, workspaceURL, objective.TeamID, objective.ID)...)
			}
		}
	}

	visibleRows, hiddenCount := capGuidanceEmailDetailRows(detailRows)
	summary := fmt.Sprintf("You have %s that need attention.", formatEmailStrong(fmt.Sprintf("%d %s", totalItems, itemText)))
	if hiddenCount > 0 {
		summary += fmt.Sprintf(
			" They represent %d objective and key-result signals. This email includes %d signals; %d more %s available in objectives.",
			len(detailRows),
			len(visibleRows),
			hiddenCount,
			pluralize(hiddenCount, "is", "are"),
		)
	}
	rows := append([]string{summary}, visibleRows...)

	return formatCompactNotificationRows("Here's what needs attention.", rows)
}

// parseKeyResults parses the JSON string of key results
func parseKeyResults(keyResultsJSON string) []OverdueKeyResult {
	if keyResultsJSON == "" || keyResultsJSON == "[]" {
		return nil
	}

	var keyResults []OverdueKeyResult
	if err := json.Unmarshal([]byte(keyResultsJSON), &keyResults); err != nil {
		// This value comes from an aggregate query. A malformed legacy row must
		// not abort delivery of otherwise valid objective guidance, but it must
		// also never produce partially decoded email content.
		return nil
	}
	return keyResults
}

// formatKeyResultEmailRows formats key results for email display.
func formatKeyResultEmailRows(keyResults []OverdueKeyResult, workspaceURL string, teamID uuid.UUID, objectiveID uuid.UUID) []string {
	if len(keyResults) == 0 {
		return nil
	}

	rows := make([]string, 0, len(keyResults))
	for _, kr := range keyResults {
		// Parse the date string to format it properly
		endDateStr := kr.EndDate // Default to raw string
		if endDate, err := time.Parse("2006-01-02", kr.EndDate); err == nil {
			endDateStr = endDate.Format("Jan 2, 2006")
		}

		// Display all key results that were fetched (they're already filtered in the SQL query)
		statusText := ""
		switch kr.DeadlineStatus {
		case "overdue":
			daysText := "day"
			if kr.DaysDifference > 1 {
				daysText = "days"
			}
			statusText = fmt.Sprintf("%s overdue (due %s)", formatEmailStrong(fmt.Sprintf("%d %s", kr.DaysDifference, daysText)), html.EscapeString(endDateStr))
		case "due_today":
			statusText = fmt.Sprintf("due today (%s)", html.EscapeString(endDateStr))
		case "due_tomorrow":
			statusText = fmt.Sprintf("due tomorrow (%s)", html.EscapeString(endDateStr))
		case "due_in_7_days":
			statusText = fmt.Sprintf("due %s", html.EscapeString(endDateStr))
		case "future":
			statusText = fmt.Sprintf("due %s", html.EscapeString(endDateStr))
		default:
			statusText = fmt.Sprintf("due %s", html.EscapeString(endDateStr))
		}

		rows = append(rows, fmt.Sprintf(
			"Key result %s is %s.",
			formatEmailLink(fmt.Sprintf("%s/teams/%s/objectives/%s", workspaceURL, teamID.String(), objectiveID.String()), kr.Name),
			statusText,
		))
	}
	return rows
}
