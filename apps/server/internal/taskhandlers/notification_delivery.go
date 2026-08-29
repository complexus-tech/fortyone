package taskhandlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	stdhtml "html"
	"strings"

	notificationsdomain "github.com/complexus-tech/projects-api/internal/modules/notifications/domain"
	"github.com/complexus-tech/projects-api/pkg/mailer"
	"github.com/google/uuid"
	htmlparser "golang.org/x/net/html"
)

// parseNotificationMessage converts the template and variables into readable text
func parseNotificationMessage(msg NotificationMessage) ParsedMessage {
	template := notificationPlainText(msg.Template, maxNotificationMessageRunes)
	htmlTemplate := stdhtml.EscapeString(template)

	// Replace template variables with actual values
	for key, variable := range msg.Variables {
		placeholder := "{" + key + "}"
		value := notificationPlainText(variable.Value, maxNotificationMessageRunes)

		// Create HTML version with styling based on variable type
		var htmlValue string
		switch variable.Type {
		case "actor":
			htmlValue = fmt.Sprintf("<strong style=\"%s\">%s</strong>", mailer.EmailStyleString("detailValue"), stdhtml.EscapeString(value))
		case "field":
			htmlValue = fmt.Sprintf("<em style=\"%s\">%s</em>", mailer.EmailStyleString("detailValue"), stdhtml.EscapeString(value))
		case "assignee", "value", "date":
			htmlValue = fmt.Sprintf("<strong style=\"%s\">%s</strong>", mailer.EmailStyleString("detailValue"), stdhtml.EscapeString(value))
		default:
			htmlValue = stdhtml.EscapeString(value)
		}

		template = strings.ReplaceAll(template, placeholder, value)
		htmlTemplate = strings.ReplaceAll(htmlTemplate, placeholder, htmlValue)
	}

	return ParsedMessage{
		Text: notificationPlainText(template, maxNotificationMessageRunes),
		HTML: htmlTemplate,
	}
}

func notificationPlainText(value string, maxRunes int) string {
	tokenizer := htmlparser.NewTokenizer(strings.NewReader(value))
	var text strings.Builder
	for {
		switch tokenizer.Next() {
		case htmlparser.ErrorToken:
			plain := strings.Join(strings.Fields(text.String()), " ")
			runes := []rune(plain)
			if maxRunes > 0 && len(runes) > maxRunes {
				return strings.TrimSpace(string(runes[:maxRunes-1])) + "…"
			}
			return plain
		case htmlparser.TextToken:
			if text.Len() > 0 {
				text.WriteByte(' ')
			}
			text.Write(tokenizer.Text())
		}
	}
}

// getNotificationEmailData retrieves all required data for sending notification email in a single query
func (h *handlers) getNotificationEmailData(ctx context.Context, query notificationsdomain.EmailNotificationQuery) (*NotificationEmailData, error) {
	if h.notificationDeliveries == nil {
		return nil, errors.New("notification delivery store is unavailable")
	}
	delivery, err := h.notificationDeliveries.GetEmailDelivery(ctx, query)
	if err != nil || delivery == nil {
		return nil, err
	}
	return &NotificationEmailData{
		NotificationID: delivery.NotificationID, RecipientID: delivery.RecipientID,
		WorkspaceID: delivery.WorkspaceID, NotificationType: string(delivery.NotificationType),
		EntityType: string(delivery.EntityType), EntityID: delivery.EntityID,
		Title: delivery.Title, Message: delivery.Message, UserEmail: delivery.UserEmail,
		UserName: delivery.UserName, ActorName: delivery.ActorName,
		WorkspaceName: delivery.WorkspaceName, WorkspaceSlug: delivery.WorkspaceSlug,
		WorkspaceRole: delivery.WorkspaceRole, EmailEnabled: delivery.EmailEnabled,
		FeedbackSlug: delivery.FeedbackSlug,
	}, nil
}

func (h *handlers) getNotificationEmailDigestData(ctx context.Context, recipientID, workspaceID uuid.UUID) (*NotificationEmailDigestData, error) {
	if h.notificationDeliveries == nil {
		return nil, errors.New("notification delivery store is unavailable")
	}
	delivery, err := h.notificationDeliveries.ListEmailDigest(ctx, notificationsdomain.DeliveryScope{
		RecipientID: recipientID,
		WorkspaceID: workspaceID,
	})
	if err != nil || delivery == nil {
		return nil, err
	}
	items := make([]NotificationEmailDigestItem, len(delivery.Items))
	for i, row := range delivery.Items {
		items[i] = NotificationEmailDigestItem{
			NotificationID:   row.NotificationID,
			NotificationType: string(row.NotificationType),
			EntityType:       string(row.EntityType),
			EntityID:         row.EntityID,
			Title:            row.Title,
			Message:          row.Message,
			CreatedAt:        row.CreatedAt,
			ActorName:        row.ActorName,
			FeedbackSlug:     row.FeedbackSlug,
		}
	}

	return &NotificationEmailDigestData{
		RecipientID:   delivery.RecipientID,
		WorkspaceID:   delivery.WorkspaceID,
		UserEmail:     delivery.UserEmail,
		UserName:      delivery.UserName,
		WorkspaceName: delivery.WorkspaceName,
		WorkspaceSlug: delivery.WorkspaceSlug,
		WorkspaceRole: delivery.WorkspaceRole,
		Items:         items,
	}, nil
}

func (h *handlers) filterStrategyDigestForCurrentAccess(ctx context.Context, data *NotificationEmailDigestData) ([]uuid.UUID, error) {
	if data == nil || data.WorkspaceRole == "admin" {
		return nil, nil
	}

	containsWeeklyStrategy := false
	for _, item := range data.Items {
		if item.EntityType != "strategy" {
			continue
		}
		var message NotificationMessage
		if err := json.Unmarshal(item.Message, &message); err != nil {
			return nil, fmt.Errorf("unmarshal strategy notification %s: %w", item.NotificationID, err)
		}
		if message.Strategy != nil && message.Strategy.Kind == "weekly_check_in" && message.Strategy.WeeklyCheckIn != nil {
			containsWeeklyStrategy = true
			break
		}
	}
	if !containsWeeklyStrategy {
		return nil, nil
	}

	if h.notificationDeliveries == nil {
		return nil, errors.New("notification delivery store is unavailable")
	}
	teamIDs, err := h.notificationDeliveries.ListDeliveryTeamIDs(ctx, notificationsdomain.DeliveryScope{
		RecipientID: data.RecipientID,
		WorkspaceID: data.WorkspaceID,
	})
	if err != nil {
		return nil, fmt.Errorf("load current notification team access: %w", err)
	}
	allowedTeams := make(map[uuid.UUID]struct{}, len(teamIDs))
	for _, teamID := range teamIDs {
		allowedTeams[teamID] = struct{}{}
	}

	items := make([]NotificationEmailDigestItem, 0, len(data.Items))
	suppressed := make([]uuid.UUID, 0)
	for _, item := range data.Items {
		filtered, include, err := filterWeeklyStrategyItemForTeams(item, allowedTeams)
		if err != nil {
			return nil, err
		}
		if !include {
			suppressed = append(suppressed, item.NotificationID)
			continue
		}
		items = append(items, filtered)
	}
	data.Items = items
	return suppressed, nil
}

func filterWeeklyStrategyItemForTeams(item NotificationEmailDigestItem, allowedTeams map[uuid.UUID]struct{}) (NotificationEmailDigestItem, bool, error) {
	if item.EntityType != "strategy" {
		return item, true, nil
	}
	var message NotificationMessage
	if err := json.Unmarshal(item.Message, &message); err != nil {
		return NotificationEmailDigestItem{}, false, fmt.Errorf("unmarshal strategy notification %s: %w", item.NotificationID, err)
	}
	if message.Strategy == nil || message.Strategy.Kind != "weekly_check_in" || message.Strategy.WeeklyCheckIn == nil {
		return item, true, nil
	}

	weekly := message.Strategy.WeeklyCheckIn
	objectives := make([]strategyObjectiveSnapshot, 0, len(weekly.Objectives))
	for _, objective := range weekly.Objectives {
		if _, ok := allowedTeams[objective.TeamID]; ok {
			objectives = append(objectives, objective)
		}
	}
	keyResults := make([]strategyKeyResultSnapshot, 0, len(weekly.KeyResults))
	for _, keyResult := range weekly.KeyResults {
		if _, ok := allowedTeams[keyResult.TeamID]; ok {
			keyResults = append(keyResults, keyResult)
		}
	}
	if len(objectives) == 0 && len(keyResults) == 0 {
		hasAllowedSignals := false
		for _, teamCounts := range weekly.TeamCounts {
			if _, ok := allowedTeams[teamCounts.TeamID]; ok && (teamCounts.Counts.AtRiskObjectives > 0 ||
				teamCounts.Counts.StaleObjectives > 0 ||
				teamCounts.Counts.StaleKeyResults > 0) {
				hasAllowedSignals = true
				break
			}
		}
		if !hasAllowedSignals {
			return NotificationEmailDigestItem{}, false, nil
		}
	}

	weekly.Objectives = objectives
	weekly.KeyResults = keyResults
	hadTeamCounts := len(weekly.TeamCounts) > 0
	weekly.Counts = weeklyStrategyCountsForAllowedTeams(weekly, allowedTeams)
	weekly.TeamCounts = filterWeeklyStrategyTeamCounts(weekly.TeamCounts, allowedTeams)
	if hadTeamCounts {
		weekly.OmittedDetails = weeklyStrategyOmittedDetailsForTeams(weekly.TeamCounts)
	} else {
		// Legacy snapshots did not preserve per-team aggregates. Once access is
		// re-evaluated, their global omitted count cannot be attributed safely.
		weekly.OmittedDetails = nil
	}
	encoded, err := json.Marshal(message)
	if err != nil {
		return NotificationEmailDigestItem{}, false, fmt.Errorf("marshal filtered strategy notification %s: %w", item.NotificationID, err)
	}
	item.Message = encoded
	return item, true, nil
}

func weeklyStrategyOmittedDetailsForTeams(teamCounts []strategyWeeklyCheckInTeamCountsSnapshot) *strategyWeeklyCheckInOmittedDetailsSnapshot {
	omitted := strategyWeeklyCheckInOmittedDetailsSnapshot{}
	for _, teamCount := range teamCounts {
		if teamCount.OmittedDetails == nil {
			continue
		}
		omitted.Objectives += teamCount.OmittedDetails.Objectives
		omitted.KeyResults += teamCount.OmittedDetails.KeyResults
	}
	if omitted.Objectives == 0 && omitted.KeyResults == 0 {
		return nil
	}
	return &omitted
}

func weeklyStrategyCountsForAllowedTeams(weekly *strategyWeeklyCheckInSnapshot, allowedTeams map[uuid.UUID]struct{}) strategyWeeklyCheckInCounts {
	if len(weekly.TeamCounts) > 0 {
		counts := strategyWeeklyCheckInCounts{}
		for _, teamCounts := range weekly.TeamCounts {
			if _, ok := allowedTeams[teamCounts.TeamID]; !ok {
				continue
			}
			counts.AtRiskObjectives += teamCounts.Counts.AtRiskObjectives
			counts.StaleObjectives += teamCounts.Counts.StaleObjectives
			counts.StaleKeyResults += teamCounts.Counts.StaleKeyResults
			counts.UniqueObjectives += teamCounts.Counts.UniqueObjectives
		}
		return counts
	}
	return weeklyStrategyCounts(weekly.Objectives, weekly.KeyResults)
}

func filterWeeklyStrategyTeamCounts(teamCounts []strategyWeeklyCheckInTeamCountsSnapshot, allowedTeams map[uuid.UUID]struct{}) []strategyWeeklyCheckInTeamCountsSnapshot {
	if len(teamCounts) == 0 {
		return nil
	}
	filtered := make([]strategyWeeklyCheckInTeamCountsSnapshot, 0, len(teamCounts))
	for _, teamCount := range teamCounts {
		if _, ok := allowedTeams[teamCount.TeamID]; ok {
			filtered = append(filtered, teamCount)
		}
	}
	return filtered
}

func weeklyStrategyCounts(objectives []strategyObjectiveSnapshot, keyResults []strategyKeyResultSnapshot) strategyWeeklyCheckInCounts {
	counts := strategyWeeklyCheckInCounts{StaleKeyResults: len(keyResults)}
	uniqueObjectives := make(map[uuid.UUID]struct{}, len(objectives)+len(keyResults))
	for _, objective := range objectives {
		uniqueObjectives[objective.ID] = struct{}{}
		for _, reason := range objective.Reasons {
			switch reason {
			case "at_risk":
				counts.AtRiskObjectives++
			case "stale":
				counts.StaleObjectives++
			}
		}
	}
	for _, keyResult := range keyResults {
		uniqueObjectives[keyResult.ObjectiveID] = struct{}{}
	}
	counts.UniqueObjectives = len(uniqueObjectives)
	return counts
}
