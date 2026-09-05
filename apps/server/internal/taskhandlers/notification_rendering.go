package taskhandlers

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	stdhtml "html"
	"net/url"
	"sort"
	"strings"

	notificationsdomain "github.com/complexus-tech/projects-api/internal/modules/notifications/domain"
	"github.com/complexus-tech/projects-api/pkg/emailcopy"
	"github.com/complexus-tech/projects-api/pkg/emailthread"
	"github.com/complexus-tech/projects-api/pkg/mailer"
	"github.com/google/uuid"
)

func generateNotificationDigestCopy(ctx context.Context, generator emailcopy.Generator, input notificationDigestCopyInput) (notificationDigestCopy, error) {
	if generator == nil {
		return input.Fallback, nil
	}
	generated, err := generator.Generate(ctx, input.Request)
	if err != nil {
		return input.Fallback, fmt.Errorf("generate notification digest copy: %w", err)
	}
	resolved, err := buildGeneratedNotificationDigestCopy(input, generated)
	if err != nil {
		return input.Fallback, fmt.Errorf("resolve notification digest copy: %w", err)
	}
	return resolved, nil
}

func renderNotificationDigestCopy(copy notificationDigestCopy) string {
	textStyle := mailer.EmailStyleString("notificationText")
	listStyle := mailer.EmailStyleString("notificationList")
	firstItemStyle := mailer.EmailStyleString("notificationItemFirst")
	defaultItemStyle := mailer.EmailStyleString("notificationItem")
	linkStyle := mailer.EmailStyleString("notificationLink")
	messageStyle := mailer.EmailStyleString("notificationMessage")

	content := fmt.Sprintf(`
		<div style="%s">
			<p style="%s">%s</p>
			<div style="%s">
	`, textStyle, textStyle, stdhtml.EscapeString(copy.Intro), listStyle)
	for index, row := range copy.Rows {
		itemStyle := defaultItemStyle
		if index == 0 {
			itemStyle = firstItemStyle
		}
		rowHTML := stdhtml.EscapeString(row.Text)
		if row.URL != "" && row.Label != "" {
			escapedLabel := stdhtml.EscapeString(row.Label)
			rowHTML = strings.Replace(rowHTML, escapedLabel, fmt.Sprintf(
				`<a href="%s" style="%s">%s</a>`,
				stdhtml.EscapeString(row.URL),
				linkStyle,
				escapedLabel,
			), 1)
		}
		content += fmt.Sprintf(`
			<div style="%s">
				<p style="%s">%s</p>
			</div>
		`, itemStyle, messageStyle, rowHTML)
	}
	return content + "</div></div>"
}

func pluralWord(count int, singular, plural string) string {
	if count == 1 {
		return singular
	}
	return plural
}

func nonEmptyStrings(values ...string) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func strategyValue(value *float64) string {
	if value == nil {
		return ""
	}
	return fmt.Sprintf("%g", *value)
}

func containsValue(values []string, expected string) bool {
	for _, value := range values {
		if value == expected {
			return true
		}
	}
	return false
}

func notificationEmailDestination(entityType string, entityID uuid.UUID, feedbackSlug string, notificationID uuid.UUID, workspaceURL string) (string, string) {
	if entityType == "feedback" && feedbackSlug != "" {
		return fmt.Sprintf("%s/feedback/%s", workspaceURL, url.PathEscape(feedbackSlug)), "feedback"
	}
	return fmt.Sprintf(
		"%s/notifications/%s?entityId=%s&entityType=%s",
		workspaceURL,
		notificationID.String(),
		url.QueryEscape(entityID.String()),
		url.QueryEscape(entityType),
	), notificationEntityLabel(entityType)
}

func notificationEntityLabel(entityType string) string {
	switch entityType {
	case "objective", "key_result":
		return "strategy"
	case "strategy":
		return "strategy"
	case "story":
		return "work"
	default:
		return "update"
	}
}

func feedbackOnlyDigest(items []NotificationEmailDigestItem) bool {
	if len(items) == 0 {
		return false
	}
	for _, item := range items {
		if item.EntityType != "feedback" || item.FeedbackSlug == "" {
			return false
		}
	}
	return true
}

func renderNotificationDigestPlainText(copy notificationDigestCopy) string {
	sections := nonEmptyStrings(copy.Heading, copy.Intro)
	for _, row := range copy.Rows {
		rowText := strings.TrimSpace(row.Text)
		if rowText == "" {
			continue
		}
		if row.URL != "" {
			rowText += "\n" + row.URL
		}
		sections = append(sections, rowText)
	}
	if copy.CTA.Label != "" && copy.CTA.URL != "" {
		sections = append(sections, copy.CTA.Label+": "+copy.CTA.URL)
	}
	return strings.Join(sections, "\n\n")
}

func notificationDigestMessageID(data NotificationEmailDigestData) string {
	ids := make([]string, 0, len(data.Items))
	for _, item := range data.Items {
		ids = append(ids, item.NotificationID.String())
	}
	sort.Strings(ids)
	digest := sha256.Sum256([]byte(data.WorkspaceID.String() + ":" + data.RecipientID.String() + ":" + strings.Join(ids, ",")))
	return "<notification-digest-" + hex.EncodeToString(digest[:16]) + "@fortyone.app>"
}

func notificationGuidanceThreadContext(data NotificationEmailDigestData, additionalTargets ...emailthread.TargetContext) (json.RawMessage, error) {
	targets := make([]emailthread.TargetContext, 0)
	seen := make(map[string]struct{})
	addTarget := func(target emailthread.TargetContext) {
		key := target.Kind + ":" + target.ID.String()
		if target.ID == uuid.Nil {
			return
		}
		if _, exists := seen[key]; exists {
			return
		}
		seen[key] = struct{}{}
		targets = append(targets, target)
	}

	for _, target := range additionalTargets {
		addTarget(target)
	}
	for _, item := range data.Items {
		var message NotificationMessage
		if err := json.Unmarshal(item.Message, &message); err != nil || message.Strategy == nil {
			continue
		}
		if weekly := message.Strategy.WeeklyCheckIn; weekly != nil {
			for _, objective := range weekly.Objectives {
				addTarget(emailthread.TargetContext{
					Kind:        "objective",
					ID:          objective.ID,
					TeamID:      objective.TeamID,
					DisplayName: objective.Name,
				})
			}
			for _, keyResult := range weekly.KeyResults {
				addTarget(emailthread.TargetContext{
					Kind:        "key_result",
					ID:          keyResult.ID,
					TeamID:      keyResult.TeamID,
					ParentID:    keyResult.ObjectiveID,
					DisplayName: keyResult.Name,
				})
			}
		}
	}
	source := "strategy_notification"
	if len(additionalTargets) > 0 {
		source = "workspace_updates"
	}
	return emailthread.EncodeThreadContext(emailthread.ThreadContext{
		Source:        source,
		WorkspaceSlug: data.WorkspaceSlug,
		Targets:       targets,
	})
}

func (h *handlers) prepareNotificationGuidanceThread(
	ctx context.Context,
	data NotificationEmailDigestData,
	copy notificationDigestCopy,
	messageID string,
	plainText string,
	targets ...emailthread.TargetContext,
) (string, error) {
	if (!copy.HasStrategySnapshot && len(targets) == 0) || copy.Sender != mailer.SenderProfileMaya || h.emailThreads == nil {
		return "", nil
	}
	threadContext, err := notificationGuidanceThreadContext(data, targets...)
	if err != nil {
		return "", err
	}
	prepared, err := h.emailThreads.PrepareGuidance(ctx, emailthread.GuidanceInput{
		WorkspaceID:       data.WorkspaceID,
		UserID:            data.RecipientID,
		RecipientEmail:    data.UserEmail,
		ExternalThreadID:  messageID,
		InternetMessageID: messageID,
		Subject:           copy.Subject,
		Content:           plainText,
		Context:           threadContext,
	})
	if err != nil {
		return "", err
	}
	return prepared.ReplyTo, nil
}

func (h *handlers) markNotificationsEmailSent(ctx context.Context, scope notificationsdomain.DeliveryScope, notificationIDs []uuid.UUID) error {
	if len(notificationIDs) == 0 {
		return nil
	}
	if h.notificationDeliveries == nil {
		return errors.New("notification delivery store is unavailable")
	}
	if err := h.notificationDeliveries.MarkEmailSent(ctx, scope, notificationIDs); err != nil {
		h.log.Error(ctx, "Failed to mark notifications as emailed", "error", err)
		return fmt.Errorf("mark notifications as emailed: %w", err)
	}
	return nil
}

func templateDigest(copy notificationDigestCopy) mailer.Digest {
	result := mailer.Digest{Intro: copy.Intro, Rows: make([]mailer.DigestRow, 0, len(copy.Rows))}
	for _, row := range copy.Rows {
		text := strings.TrimPrefix(row.Text, row.Label+": ")
		result.Rows = append(result.Rows, mailer.DigestRow{Label: row.Label, Text: text, URL: row.URL, Actor: row.Actor, Icon: row.Icon, More: row.Label == "Notifications" || (row.Label == "" && strings.Contains(row.Text, "more details in Strategy"))})
	}
	return result
}

func notificationIcon(message NotificationMessage, notificationType string) string {
	for _, variable := range message.Variables {
		if variable.Type == "date" || (variable.Type == "field" && strings.Contains(strings.ToLower(variable.Value), "date")) {
			return "calendar"
		}
	}
	if strings.Contains(notificationType, "comment") || strings.Contains(notificationType, "reply") || strings.Contains(notificationType, "conversation") {
		return "comment"
	}
	return ""
}
