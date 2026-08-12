package taskhandlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	stdhtml "html"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/complexus-tech/projects-api/pkg/emailcopy"
	"github.com/complexus-tech/projects-api/pkg/mailer"
	"github.com/complexus-tech/projects-api/pkg/tasks"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/lib/pq"
	htmlparser "golang.org/x/net/html"
)

const (
	maxNotificationMessageRunes     = 600
	maxNotificationDigestRows       = 12
	maxNotificationDigestDetailRows = maxNotificationDigestRows - 1
)

const (
	digestActionPrimary  = "digest_primary"
	digestActionStrategy = "strategy"
)

type Variable struct {
	Value string `json:"value"`
	Type  string `json:"type"`
}

type NotificationMessage struct {
	Template  string                        `json:"template"`
	Variables map[string]Variable           `json:"variables"`
	Strategy  *strategyNotificationSnapshot `json:"strategy,omitempty"`
}

type strategyNotificationSnapshot struct {
	Version        int                             `json:"version"`
	Kind           string                          `json:"kind"`
	GeneratedAt    time.Time                       `json:"generatedAt"`
	Planning       *strategyPlanningSnapshot       `json:"planning,omitempty"`
	WeeklyCheckIn  *strategyWeeklyCheckInSnapshot  `json:"weeklyCheckIn,omitempty"`
	MonthlySummary *strategyMonthlySummarySnapshot `json:"monthlySummary,omitempty"`
}

type strategyPlanningSnapshot struct {
	Period          string    `json:"period"`
	StartsAt        time.Time `json:"startsAt"`
	DaysUntil       int       `json:"daysUntil"`
	HasUltimateGoal bool      `json:"hasUltimateGoal"`
	PillarCount     int       `json:"pillarCount"`
	ObjectiveCount  int       `json:"objectiveCount"`
	MissingElements []string  `json:"missingElements"`
}

type strategyWeeklyCheckInSnapshot struct {
	StaleAfterDays int                                          `json:"staleAfterDays"`
	Counts         strategyWeeklyCheckInCounts                  `json:"counts"`
	TeamCounts     []strategyWeeklyCheckInTeamCountsSnapshot    `json:"teamCounts,omitempty"`
	Objectives     []strategyObjectiveSnapshot                  `json:"objectives"`
	KeyResults     []strategyKeyResultSnapshot                  `json:"keyResults"`
	OmittedDetails *strategyWeeklyCheckInOmittedDetailsSnapshot `json:"omittedDetails,omitempty"`
}

type strategyWeeklyCheckInTeamCountsSnapshot struct {
	TeamID         uuid.UUID                                    `json:"teamId"`
	Counts         strategyWeeklyCheckInCounts                  `json:"counts"`
	OmittedDetails *strategyWeeklyCheckInOmittedDetailsSnapshot `json:"omittedDetails,omitempty"`
}

type strategyWeeklyCheckInOmittedDetailsSnapshot struct {
	Objectives int `json:"objectives"`
	KeyResults int `json:"keyResults"`
}

type strategyWeeklyCheckInCounts struct {
	AtRiskObjectives int `json:"atRiskObjectives"`
	StaleObjectives  int `json:"staleObjectives"`
	StaleKeyResults  int `json:"staleKeyResults"`
	UniqueObjectives int `json:"uniqueObjectives"`
}

type strategyObjectiveSnapshot struct {
	ID        uuid.UUID                        `json:"id"`
	TeamID    uuid.UUID                        `json:"teamId"`
	Name      string                           `json:"name"`
	Health    *string                          `json:"health,omitempty"`
	Status    *strategyObjectiveStatusSnapshot `json:"status,omitempty"`
	StartDate *time.Time                       `json:"startDate,omitempty"`
	EndDate   *time.Time                       `json:"endDate,omitempty"`
	UpdatedAt time.Time                        `json:"updatedAt"`
	Reasons   []string                         `json:"reasons"`
}

type strategyObjectiveStatusSnapshot struct {
	ID       uuid.UUID `json:"id"`
	Name     string    `json:"name"`
	Category string    `json:"category"`
}

type strategyKeyResultSnapshot struct {
	ID              uuid.UUID                        `json:"id"`
	ObjectiveID     uuid.UUID                        `json:"objectiveId"`
	TeamID          uuid.UUID                        `json:"teamId"`
	Name            string                           `json:"name"`
	ObjectiveName   string                           `json:"objectiveName"`
	ObjectiveHealth *string                          `json:"objectiveHealth,omitempty"`
	ObjectiveStatus *strategyObjectiveStatusSnapshot `json:"objectiveStatus,omitempty"`
	MeasurementType string                           `json:"measurementType"`
	StartValue      *float64                         `json:"startValue,omitempty"`
	CurrentValue    *float64                         `json:"currentValue,omitempty"`
	TargetValue     *float64                         `json:"targetValue,omitempty"`
	StartDate       *time.Time                       `json:"startDate,omitempty"`
	EndDate         *time.Time                       `json:"endDate,omitempty"`
	UpdatedAt       time.Time                        `json:"updatedAt"`
	Reasons         []string                         `json:"reasons"`
}

type strategyMonthlySummarySnapshot struct {
	PeriodStart          time.Time `json:"periodStart"`
	PeriodEnd            time.Time `json:"periodEnd"`
	PillarCount          int       `json:"pillarCount"`
	PillarsNeedingReview int       `json:"pillarsNeedingReview"`
	ObjectiveCount       int       `json:"objectiveCount"`
	AtRiskObjectives     int       `json:"atRiskObjectives"`
	UnalignedObjectives  int       `json:"unalignedObjectives"`
	KeyResultCount       int       `json:"keyResultCount"`
	KeyResultProgress    *float64  `json:"keyResultProgress,omitempty"`
	CompletedStories     int       `json:"completedStories"`
}

// NotificationEmailData represents all data needed for sending notification emails
type NotificationEmailData struct {
	NotificationID   uuid.UUID       `db:"notification_id"`
	RecipientID      uuid.UUID       `db:"recipient_id"`
	WorkspaceID      uuid.UUID       `db:"workspace_id"`
	NotificationType string          `db:"type"`
	EntityType       string          `db:"entity_type"`
	EntityID         uuid.UUID       `db:"entity_id"`
	Title            string          `db:"title"`
	Message          json.RawMessage `db:"message"`
	UserEmail        string          `db:"user_email"`
	UserName         string          `db:"user_name"`
	ActorName        string          `db:"actor_name"`
	WorkspaceName    string          `db:"workspace_name"`
	WorkspaceSlug    string          `db:"workspace_slug"`
	WorkspaceRole    string          `db:"workspace_role"`
	EmailEnabled     bool            `db:"email_enabled"`
	FeedbackSlug     string          `db:"feedback_slug"`
}

type NotificationEmailDigestItem struct {
	NotificationID   uuid.UUID       `db:"notification_id"`
	NotificationType string          `db:"type"`
	EntityType       string          `db:"entity_type"`
	EntityID         uuid.UUID       `db:"entity_id"`
	Title            string          `db:"title"`
	Message          json.RawMessage `db:"message"`
	CreatedAt        time.Time       `db:"created_at"`
	ActorName        string          `db:"actor_name"`
	FeedbackSlug     string          `db:"feedback_slug"`
}

type NotificationEmailDigestData struct {
	RecipientID   uuid.UUID
	WorkspaceID   uuid.UUID
	UserEmail     string
	UserName      string
	WorkspaceName string
	WorkspaceSlug string
	WorkspaceRole string
	Items         []NotificationEmailDigestItem
}

type notificationDigestCopy struct {
	Subject             string
	Heading             string
	Intro               string
	Rows                []notificationDigestCopyRow
	CTA                 notificationDigestCopyCTA
	Sender              mailer.SenderProfile
	NotificationsURL    string
	HasStrategySnapshot bool
}

type notificationDigestCopyRow struct {
	Text  string
	Label string
	URL   string
}

type notificationDigestCopyCTA struct {
	Label string
	URL   string
}

type notificationDigestCopyInput struct {
	Request             emailcopy.Request
	Actions             map[string]string
	FactActions         map[string]string
	FactLabels          map[string]string
	Fallback            notificationDigestCopy
	HasStrategySnapshot bool
	NotificationsURL    string
}

type notificationEmailDigestRow struct {
	NotificationID   uuid.UUID       `db:"notification_id"`
	RecipientID      uuid.UUID       `db:"recipient_id"`
	WorkspaceID      uuid.UUID       `db:"workspace_id"`
	NotificationType string          `db:"type"`
	EntityType       string          `db:"entity_type"`
	EntityID         uuid.UUID       `db:"entity_id"`
	Title            string          `db:"title"`
	Message          json.RawMessage `db:"message"`
	CreatedAt        time.Time       `db:"created_at"`
	UserEmail        string          `db:"user_email"`
	UserName         string          `db:"user_name"`
	ActorName        string          `db:"actor_name"`
	WorkspaceName    string          `db:"workspace_name"`
	WorkspaceSlug    string          `db:"workspace_slug"`
	WorkspaceRole    string          `db:"workspace_role"`
	FeedbackSlug     string          `db:"feedback_slug"`
}

// ParsedMessage represents the final parsed notification message
type ParsedMessage struct {
	Text string
	HTML string
}

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
func (h *handlers) getNotificationEmailData(ctx context.Context, notificationID uuid.UUID) (*NotificationEmailData, error) {
	query := notificationEmailDataQuery()

	params := map[string]any{
		"notification_id": notificationID,
	}

	stmt, err := h.db.PrepareNamedContext(ctx, query)
	if err != nil {
		h.log.Error(ctx, "Failed to prepare notification email query", "error", err)
		return nil, fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	var data NotificationEmailData
	err = stmt.GetContext(ctx, &data, params)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		h.log.Error(ctx, "Failed to execute notification email query", "error", err, "notification_id", notificationID)
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	return &data, nil
}

func notificationEmailDataQuery() string {
	return `
		SELECT
			n.notification_id,
			n.recipient_id,
			n.workspace_id,
			n.type,
			n.entity_type,
			n.entity_id,
			n.title,
			n.message,
			u.email AS user_email,
			COALESCE(NULLIF(u.full_name, ''), u.username) AS user_name,
			COALESCE(NULLIF(actor_u.full_name, ''), actor_u.username) AS actor_name,
			w.name AS workspace_name,
			w.slug AS workspace_slug,
			COALESCE(CAST(wm.role AS TEXT), '') AS workspace_role,
			COALESCE(fi.slug, '') AS feedback_slug,
			CAST(COALESCE(np.preferences -> CAST(n.type AS TEXT) ->> 'email', 'true') AS BOOLEAN) AS email_enabled
		FROM
			notifications n
			INNER JOIN users u ON n.recipient_id = u.user_id
			INNER JOIN workspaces w ON n.workspace_id = w.workspace_id
			LEFT JOIN workspace_members wm
				ON wm.workspace_id = n.workspace_id
				AND wm.user_id = n.recipient_id
			INNER JOIN users actor_u ON n.actor_id = actor_u.user_id
			LEFT JOIN feedback_items fi ON CAST(n.entity_type AS TEXT) = 'feedback' AND fi.id = n.entity_id
			LEFT JOIN notification_preferences np ON n.recipient_id = np.user_id
			AND n.workspace_id = np.workspace_id
		WHERE
			n.notification_id = :notification_id
			AND n.read_at IS NULL
			AND n.email_sent_at IS NULL
			AND w.deleted_at IS NULL
			AND u.is_active = true
			AND u.is_system = false
			AND (
				CAST(n.entity_type AS TEXT) <> 'feedback'
				OR (fi.id IS NOT NULL AND fi.deleted_at IS NULL)
			)
			AND ` + notificationEmailAccessPredicate() + `
			AND NULLIF(TRIM(u.email), '') IS NOT NULL;
	`
}

func (h *handlers) getNotificationEmailDigestData(ctx context.Context, recipientID, workspaceID uuid.UUID) (*NotificationEmailDigestData, error) {
	query := notificationEmailDigestDataQuery()

	params := map[string]any{
		"recipient_id": recipientID,
		"workspace_id": workspaceID,
	}

	stmt, err := h.db.PrepareNamedContext(ctx, query)
	if err != nil {
		h.log.Error(ctx, "Failed to prepare notification email digest query", "error", err)
		return nil, fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	var rows []notificationEmailDigestRow
	if err := stmt.SelectContext(ctx, &rows, params); err != nil {
		h.log.Error(ctx, "Failed to execute notification email digest query", "error", err, "recipient_id", recipientID, "workspace_id", workspaceID)
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	if len(rows) == 0 {
		return nil, nil
	}

	items := make([]NotificationEmailDigestItem, len(rows))
	for i, row := range rows {
		items[i] = NotificationEmailDigestItem{
			NotificationID:   row.NotificationID,
			NotificationType: row.NotificationType,
			EntityType:       row.EntityType,
			EntityID:         row.EntityID,
			Title:            row.Title,
			Message:          row.Message,
			CreatedAt:        row.CreatedAt,
			ActorName:        row.ActorName,
			FeedbackSlug:     row.FeedbackSlug,
		}
	}

	return &NotificationEmailDigestData{
		RecipientID:   rows[0].RecipientID,
		WorkspaceID:   rows[0].WorkspaceID,
		UserEmail:     rows[0].UserEmail,
		UserName:      rows[0].UserName,
		WorkspaceName: rows[0].WorkspaceName,
		WorkspaceSlug: rows[0].WorkspaceSlug,
		WorkspaceRole: rows[0].WorkspaceRole,
		Items:         items,
	}, nil
}

func notificationEmailDigestDataQuery() string {
	return `
		SELECT
			n.notification_id,
			n.recipient_id,
			n.workspace_id,
			n.type,
			n.entity_type,
			n.entity_id,
			n.title,
			n.message,
			n.created_at,
			u.email AS user_email,
			COALESCE(NULLIF(u.full_name, ''), u.username) AS user_name,
			COALESCE(NULLIF(actor_u.full_name, ''), actor_u.username) AS actor_name,
			w.name AS workspace_name,
			w.slug AS workspace_slug,
			COALESCE(CAST(wm.role AS TEXT), '') AS workspace_role,
			COALESCE(fi.slug, '') AS feedback_slug
		FROM
			notifications n
			INNER JOIN users u ON n.recipient_id = u.user_id
			INNER JOIN workspaces w ON n.workspace_id = w.workspace_id
			LEFT JOIN workspace_members wm
				ON wm.workspace_id = n.workspace_id
				AND wm.user_id = n.recipient_id
			INNER JOIN users actor_u ON n.actor_id = actor_u.user_id
			LEFT JOIN feedback_items fi ON CAST(n.entity_type AS TEXT) = 'feedback' AND fi.id = n.entity_id
			LEFT JOIN notification_preferences np ON n.recipient_id = np.user_id
				AND n.workspace_id = np.workspace_id
		WHERE
			n.recipient_id = :recipient_id
			AND n.workspace_id = :workspace_id
			AND n.read_at IS NULL
			AND n.email_sent_at IS NULL
			AND w.deleted_at IS NULL
			AND u.is_active = true
			AND u.is_system = false
			AND (
				CAST(n.entity_type AS TEXT) <> 'feedback'
				OR (fi.id IS NOT NULL AND fi.deleted_at IS NULL)
			)
			AND ` + notificationEmailAccessPredicate() + `
			AND NULLIF(TRIM(u.email), '') IS NOT NULL
			AND CAST(COALESCE(np.preferences -> CAST(n.type AS TEXT) ->> 'email', 'true') AS BOOLEAN) = true
		ORDER BY n.created_at ASC;
	`
}

func notificationEmailAccessPredicate() string {
	return `(
		CAST(n.entity_type AS TEXT) = 'feedback'
		OR (
			wm.role IN ('admin', 'member', 'guest')
			AND (
				(
					CAST(n.entity_type AS TEXT) = 'story'
					AND EXISTS (
						SELECT 1
						FROM stories story
						WHERE story.id = n.entity_id
							AND story.workspace_id = n.workspace_id
							AND story.deleted_at IS NULL
							AND (
								wm.role = 'admin'
								OR EXISTS (
									SELECT 1
									FROM team_members story_member
									WHERE story_member.team_id = story.team_id
										AND story_member.user_id = n.recipient_id
								)
							)
					)
				)
				OR (
					CAST(n.entity_type AS TEXT) = 'comment'
					AND EXISTS (
						SELECT 1
						FROM story_comments comment
						INNER JOIN stories story ON story.id = comment.story_id
						WHERE comment.comment_id = n.entity_id
							AND story.workspace_id = n.workspace_id
							AND story.deleted_at IS NULL
							AND (
								wm.role = 'admin'
								OR EXISTS (
									SELECT 1
									FROM team_members comment_member
									WHERE comment_member.team_id = story.team_id
										AND comment_member.user_id = n.recipient_id
								)
							)
					)
				)
				OR (
					CAST(n.entity_type AS TEXT) = 'objective'
					AND EXISTS (
						SELECT 1
						FROM objectives objective
						WHERE objective.objective_id = n.entity_id
							AND objective.workspace_id = n.workspace_id
							AND (
								wm.role = 'admin'
								OR EXISTS (
									SELECT 1
									FROM team_members objective_member
									WHERE objective_member.team_id = objective.team_id
										AND objective_member.user_id = n.recipient_id
								)
							)
					)
				)
				OR (
					CAST(n.entity_type AS TEXT) = 'key_result'
					AND EXISTS (
						SELECT 1
						FROM key_results key_result
						INNER JOIN objectives objective ON objective.objective_id = key_result.objective_id
						WHERE key_result.id = n.entity_id
							AND objective.workspace_id = n.workspace_id
							AND (
								wm.role = 'admin'
								OR EXISTS (
									SELECT 1
									FROM team_members key_result_member
									WHERE key_result_member.team_id = objective.team_id
										AND key_result_member.user_id = n.recipient_id
								)
							)
					)
				)
				OR (
					CAST(n.entity_type AS TEXT) = 'strategy'
					AND (
						wm.role = 'admin'
						OR n.message -> 'strategy' ->> 'kind' = 'weekly_check_in'
					)
				)
			)
		)
	)`
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

	var teamIDs []uuid.UUID
	if err := h.db.SelectContext(ctx, &teamIDs, `
		SELECT tm.team_id
		FROM team_members tm
		INNER JOIN teams team ON team.team_id = tm.team_id
		WHERE tm.user_id = $1
			AND team.workspace_id = $2;
	`, data.RecipientID, data.WorkspaceID); err != nil {
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

func buildNotificationDigestSubject(workspaceName string, count int) string {
	if count == 1 {
		return fmt.Sprintf("New activity in %s", workspaceName)
	}
	return fmt.Sprintf("%d updates to review in %s", count, workspaceName)
}

func buildNotificationDigestCopyInput(data NotificationEmailDigestData, workspaceURL string) (notificationDigestCopyInput, error) {
	notificationsURL := workspaceURL + "/notifications"
	fallbackRows := make([]notificationDigestCopyRow, 0, maxNotificationDigestDetailRows+1)
	facts := []emailcopy.Fact{{
		ReferenceID: "digest_context",
		Text: fmt.Sprintf(
			"This email contains %d unread product %s from the workspace %s.",
			len(data.Items),
			pluralWord(len(data.Items), "update", "updates"),
			data.WorkspaceName,
		),
		EntityTokens: nonEmptyStrings(data.WorkspaceName),
		ProtectedTokens: nonEmptyStrings(fmt.Sprintf(
			"%d unread product %s",
			len(data.Items),
			pluralWord(len(data.Items), "update", "updates"),
		)),
	}}
	actions := make([]emailcopy.Action, 0, 1)
	actionURLs := make(map[string]string, len(data.Items)+1)
	factActions := make(map[string]string, len(data.Items))
	factLabels := make(map[string]string, len(data.Items))
	hasStrategySnapshot := false
	detailCount := 0
	omittedNotificationCount := 0

	for index, item := range data.Items {
		var message NotificationMessage
		if err := json.Unmarshal(item.Message, &message); err != nil {
			return notificationDigestCopyInput{}, fmt.Errorf("failed to unmarshal notification message %s: %w", item.NotificationID, err)
		}
		if message.Strategy != nil {
			hasStrategySnapshot = true
		}
		if detailCount >= maxNotificationDigestDetailRows {
			omittedNotificationCount++
			continue
		}

		if message.Strategy != nil && message.Strategy.Version == 1 && message.Strategy.Kind == "weekly_check_in" && message.Strategy.WeeklyCheckIn != nil {
			weeklyFacts, weeklyURLs, weeklyFactActions, weeklyFactLabels, weeklyFallbackRows := buildWeeklyStrategyDigestFacts(
				item,
				*message.Strategy.WeeklyCheckIn,
				workspaceURL,
				maxNotificationDigestDetailRows-detailCount,
			)
			facts = append(facts, weeklyFacts...)
			for referenceID, destination := range weeklyURLs {
				actionURLs[referenceID] = destination
			}
			for referenceID, actionReferenceID := range weeklyFactActions {
				factActions[referenceID] = actionReferenceID
			}
			for referenceID, label := range weeklyFactLabels {
				factLabels[referenceID] = label
			}
			fallbackRows = append(fallbackRows, weeklyFallbackRows...)
			detailCount += len(weeklyFallbackRows)
			continue
		}
		if message.Strategy != nil && message.Strategy.Version == 1 && message.Strategy.Kind == "planning_reminder" && message.Strategy.Planning != nil {
			fact, actionReferenceID, destination, label, fallbackRow := buildStrategyPlanningDigestFact(item, *message.Strategy.Planning, workspaceURL)
			facts = append(facts, fact)
			actionURLs[actionReferenceID] = destination
			factActions[fact.ReferenceID] = actionReferenceID
			factLabels[fact.ReferenceID] = label
			fallbackRows = append(fallbackRows, fallbackRow)
			detailCount++
			continue
		}
		if message.Strategy != nil && message.Strategy.Version == 1 && message.Strategy.Kind == "monthly_summary" && message.Strategy.MonthlySummary != nil {
			fact, actionReferenceID, destination, label, fallbackRow := buildStrategyMonthlyDigestFact(item, *message.Strategy, workspaceURL)
			facts = append(facts, fact)
			actionURLs[actionReferenceID] = destination
			factActions[fact.ReferenceID] = actionReferenceID
			factLabels[fact.ReferenceID] = label
			fallbackRows = append(fallbackRows, fallbackRow)
			detailCount++
			continue
		}
		parsed := parseNotificationMessage(message)
		factReferenceID := fmt.Sprintf("notification_%d", index+1)
		actionReferenceID := fmt.Sprintf("notification_action_%d", index+1)
		destination, _ := notificationEmailDestination(item.EntityType, item.EntityID, item.FeedbackSlug, item.NotificationID, workspaceURL)
		factText := strings.TrimSpace(fmt.Sprintf("%s: %s", notificationPlainText(item.Title, 180), parsed.Text))
		facts = append(facts, emailcopy.Fact{
			ReferenceID:     factReferenceID,
			Text:            factText,
			EntityTokens:    nonEmptyStrings(notificationPlainText(item.Title, 180)),
			ProtectedTokens: notificationSemanticProtectedTokens(message, parsed.Text),
			Required:        true,
		})
		actionURLs[actionReferenceID] = destination
		factActions[factReferenceID] = actionReferenceID
		factLabels[factReferenceID] = notificationPlainText(item.Title, 180)
		fallbackRows = append(fallbackRows, notificationDigestCopyRow{Text: factText, Label: factLabels[factReferenceID], URL: destination})
		detailCount++
	}
	if omittedNotificationCount > 0 {
		const remainingActionReferenceID = "remaining_notifications_action"
		facts = append(facts, emailcopy.Fact{
			ReferenceID:  "remaining_updates",
			Text:         fmt.Sprintf("There are %d additional unread %s available in Notifications.", omittedNotificationCount, pluralWord(omittedNotificationCount, "update", "updates")),
			EntityTokens: []string{"Notifications"},
			ProtectedTokens: nonEmptyStrings(fmt.Sprintf(
				"%d additional unread %s available in Notifications",
				omittedNotificationCount,
				pluralWord(omittedNotificationCount, "update", "updates"),
			)),
			Required: true,
		})
		actionURLs[remainingActionReferenceID] = notificationsURL
		factActions["remaining_updates"] = remainingActionReferenceID
		factLabels["remaining_updates"] = "Notifications"
		fallbackRows = append(fallbackRows, notificationDigestCopyRow{
			Text:  fmt.Sprintf("%d additional unread %s are available in Notifications.", omittedNotificationCount, pluralWord(omittedNotificationCount, "update", "updates")),
			Label: "Notifications",
			URL:   notificationsURL,
		})
	}

	primaryActionReferenceID := digestActionPrimary
	primaryDestination := notificationsURL
	primaryDescription := "Open notifications"
	primaryLabel := "Open notifications"
	if feedbackOnlyDigest(data.Items) {
		primaryDestination = workspaceURL + "/feedback"
		primaryDescription = "Open feedback"
		primaryLabel = "Open feedback"
	}
	if hasStrategySnapshot {
		primaryActionReferenceID = digestActionStrategy
		primaryDestination = workspaceURL + "/strategy"
		primaryDescription = "Review strategy"
		primaryLabel = "Review strategy"
	}
	actions = append(actions, emailcopy.Action{ReferenceID: primaryActionReferenceID, Description: primaryDescription})
	actionURLs[primaryActionReferenceID] = primaryDestination

	fallbackSubject := buildNotificationDigestSubject(data.WorkspaceName, len(data.Items))
	fallbackIntro := "Review the activity below and choose the next useful action."
	if hasStrategySnapshot {
		fallbackSubject = "Your strategy check-in"
		fallbackIntro = "Here are the objectives, key results, and strategy updates that need your attention."
	}
	fallback := notificationDigestCopy{
		Subject:             fallbackSubject,
		Heading:             fallbackSubject,
		Intro:               fallbackIntro,
		Rows:                fallbackRows,
		CTA:                 notificationDigestCopyCTA{Label: primaryLabel, URL: primaryDestination},
		NotificationsURL:    notificationsURL,
		HasStrategySnapshot: hasStrategySnapshot,
	}
	if hasStrategySnapshot {
		fallback.Sender = mailer.SenderProfileMaya
	}

	safetyIdentifier := data.RecipientID.String()
	if data.RecipientID == uuid.Nil {
		safetyIdentifier = data.UserEmail
	}

	return notificationDigestCopyInput{
		Request: emailcopy.Request{
			SafetyIdentifier:   safetyIdentifier,
			Purpose:            "persisted notification email digest",
			ProductVoice:       "Help the recipient understand what changed, why it matters, and choose the next useful action. Keep activity factual; make strategy guidance warm and calm.",
			Facts:              facts,
			Actions:            actions,
			IncludeSenderProse: hasStrategySnapshot,
			IncludeReplyPrompt: false,
		},
		Actions:             actionURLs,
		FactActions:         factActions,
		FactLabels:          factLabels,
		Fallback:            fallback,
		HasStrategySnapshot: hasStrategySnapshot,
		NotificationsURL:    notificationsURL,
	}, nil
}

func notificationVariableValues(variables map[string]Variable) []string {
	keys := make([]string, 0, len(variables))
	for key := range variables {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	values := make([]string, 0, len(keys))
	for _, key := range keys {
		variable := variables[key]
		values = append(values, notificationPlainText(variable.Value, maxNotificationMessageRunes))
	}
	return nonEmptyStrings(values...)
}

// notificationSemanticProtectedTokens keeps the factual roles in activity
// messages bound together. Requiring only the individual variable values
// would allow generated copy to swap an actor, field, assignee, or status while
// still passing literal-token validation. For long comment-style messages, the
// author/action prefix stays exact while the user-authored body remains free to
// be summarized.
func notificationSemanticProtectedTokens(message NotificationMessage, parsedText string) []string {
	values := notificationVariableValues(message.Variables)
	if len(values) == 0 {
		return values
	}

	const maxProtectedActivityRunes = 300
	activity := notificationPlainText(parsedText, 0)
	if len([]rune(activity)) <= maxProtectedActivityRunes {
		return nonEmptyStrings(activity)
	}
	if separator := strings.Index(activity, ":"); separator > 0 {
		semanticPrefix := strings.TrimSpace(activity[:separator])
		if len([]rune(semanticPrefix)) <= maxProtectedActivityRunes {
			return nonEmptyStrings(semanticPrefix)
		}
	}
	return values
}

func buildStrategyPlanningDigestFact(
	item NotificationEmailDigestItem,
	planning strategyPlanningSnapshot,
	workspaceURL string,
) (emailcopy.Fact, string, string, string, notificationDigestCopyRow) {
	ultimateGoalState := "does not yet have an ultimate goal"
	if planning.HasUltimateGoal {
		ultimateGoalState = "has an ultimate goal"
	}
	missingElements := strategyMissingElementLabels(planning.MissingElements)
	factText := fmt.Sprintf(
		"%s: %s starts on %s, in %d days. The strategy foundation %s, has %d strategic pillars, and has %d objectives for that period.",
		item.Title,
		planning.Period,
		planning.StartsAt.UTC().Format("January 2, 2006"),
		planning.DaysUntil,
		ultimateGoalState,
		planning.PillarCount,
		planning.ObjectiveCount,
	)
	if len(missingElements) > 0 {
		factText += " The missing elements are " + strings.Join(missingElements, ", ") + "."
	}
	factReferenceID := "strategy_planning_" + item.NotificationID.String()
	actionReferenceID := "strategy_planning_action_" + item.NotificationID.String()
	destination := workspaceURL + "/strategy"
	label := notificationPlainText(item.Title, 180)
	protectedTokens := nonEmptyStrings(
		fmt.Sprintf("%s starts on %s", planning.Period, planning.StartsAt.UTC().Format("January 2, 2006")),
		fmt.Sprintf("in %d days", planning.DaysUntil),
		"strategy foundation "+ultimateGoalState,
		fmt.Sprintf("has %d strategic pillars", planning.PillarCount),
		fmt.Sprintf("has %d objectives for that period", planning.ObjectiveCount),
	)
	if len(missingElements) > 0 {
		protectedTokens = append(protectedTokens, "missing elements are "+strings.Join(missingElements, ", "))
	}
	fact := emailcopy.Fact{
		ReferenceID:     factReferenceID,
		Text:            factText,
		EntityTokens:    nonEmptyStrings(label, planning.Period),
		ProtectedTokens: protectedTokens,
		Required:        true,
	}
	return fact, actionReferenceID, destination, label, notificationDigestCopyRow{Text: factText, Label: label, URL: destination}
}

func buildStrategyMonthlyDigestFact(
	item NotificationEmailDigestItem,
	snapshot strategyNotificationSnapshot,
	workspaceURL string,
) (emailcopy.Fact, string, string, string, notificationDigestCopyRow) {
	monthly := snapshot.MonthlySummary
	periodLabel := monthly.PeriodStart.UTC().Format("January 2006")
	generatedLabel := snapshot.GeneratedAt.UTC().Format("January 2, 2006")
	periodStart := monthly.PeriodStart.UTC().Format("January 2, 2006")
	periodEnd := monthly.PeriodEnd.UTC().Format("January 2, 2006")
	keyResultSummary := strategyMonthlyKeyResultSummary(*monthly)
	factText := fmt.Sprintf(
		"%s covers %s and was generated on %s. The current snapshot has %d strategic pillars, including %d that need review; %d objectives, including %d at risk or off track and %d without a strategic alignment; and %s. From %s up to %s, %d linked tasks were completed.",
		item.Title,
		periodLabel,
		generatedLabel,
		monthly.PillarCount,
		monthly.PillarsNeedingReview,
		monthly.ObjectiveCount,
		monthly.AtRiskObjectives,
		monthly.UnalignedObjectives,
		keyResultSummary,
		periodStart,
		periodEnd,
		monthly.CompletedStories,
	)
	factReferenceID := "strategy_monthly_" + item.NotificationID.String()
	actionReferenceID := "strategy_monthly_action_" + item.NotificationID.String()
	destination := workspaceURL + "/strategy"
	label := notificationPlainText(item.Title, 180)
	fact := emailcopy.Fact{
		ReferenceID:  factReferenceID,
		Text:         factText,
		EntityTokens: nonEmptyStrings(label, periodLabel),
		ProtectedTokens: nonEmptyStrings(
			"generated on "+generatedLabel,
			fmt.Sprintf("has %d strategic pillars", monthly.PillarCount),
			fmt.Sprintf("including %d that need review", monthly.PillarsNeedingReview),
			fmt.Sprintf("%d objectives", monthly.ObjectiveCount),
			fmt.Sprintf("including %d at risk or off track", monthly.AtRiskObjectives),
			fmt.Sprintf("%d without a strategic alignment", monthly.UnalignedObjectives),
			keyResultSummary,
			fmt.Sprintf("From %s up to %s", periodStart, periodEnd),
			fmt.Sprintf("%d linked tasks were completed", monthly.CompletedStories),
		),
		Required: true,
	}
	return fact, actionReferenceID, destination, label, notificationDigestCopyRow{Text: factText, Label: label, URL: destination}
}

func strategyMonthlyKeyResultSummary(monthly strategyMonthlySummarySnapshot) string {
	switch {
	case monthly.KeyResultProgress != nil && monthly.KeyResultCount > 0:
		return fmt.Sprintf(
			"%.0f%% average key-result progress across %d %s",
			*monthly.KeyResultProgress,
			monthly.KeyResultCount,
			pluralWord(monthly.KeyResultCount, "key result", "key results"),
		)
	case monthly.KeyResultProgress != nil:
		// Snapshots created before keyResultCount was added still carry valid
		// progress; preserve it during the additive schema transition.
		return fmt.Sprintf("%.0f%% average key-result progress", *monthly.KeyResultProgress)
	case monthly.KeyResultCount > 0:
		return fmt.Sprintf(
			"average progress is unavailable for %d %s",
			monthly.KeyResultCount,
			pluralWord(monthly.KeyResultCount, "key result", "key results"),
		)
	default:
		return "there are no key results in the current snapshot"
	}
}

func strategyMissingElementLabels(elements []string) []string {
	labels := make([]string, 0, len(elements))
	for _, element := range elements {
		switch element {
		case "ultimate_goal":
			labels = append(labels, "ultimate goal")
		case "strategic_pillars":
			labels = append(labels, "strategic pillars")
		case "objectives":
			labels = append(labels, "objectives")
		}
	}
	return labels
}

func buildWeeklyStrategyDigestFacts(item NotificationEmailDigestItem, weekly strategyWeeklyCheckInSnapshot, workspaceURL string, rowLimit int) ([]emailcopy.Fact, map[string]string, map[string]string, map[string]string, []notificationDigestCopyRow) {
	facts := make([]emailcopy.Fact, 0, rowLimit)
	actionURLs := make(map[string]string, len(weekly.Objectives)+len(weekly.KeyResults))
	factActions := make(map[string]string, len(weekly.Objectives)+len(weekly.KeyResults)+1)
	factLabels := make(map[string]string, len(weekly.Objectives)+len(weekly.KeyResults))
	fallbackRows := make([]notificationDigestCopyRow, 0, rowLimit)

	summaryReferenceID := fmt.Sprintf("strategy_summary_%s", item.NotificationID.String())
	summaryText := fmt.Sprintf(
		"The weekly strategy check-in has %d at-risk objectives, %d objectives without a recent update, and %d incomplete key results without a recent update; recent means within %d days.",
		weekly.Counts.AtRiskObjectives,
		weekly.Counts.StaleObjectives,
		weekly.Counts.StaleKeyResults,
		weekly.StaleAfterDays,
	)
	protectedSummaryTokens := nonEmptyStrings(
		fmt.Sprintf("%d at-risk objectives", weekly.Counts.AtRiskObjectives),
		fmt.Sprintf("%d objectives without a recent update", weekly.Counts.StaleObjectives),
		fmt.Sprintf("%d incomplete key results without a recent update", weekly.Counts.StaleKeyResults),
		fmt.Sprintf("within %d days", weekly.StaleAfterDays),
	)
	if weekly.OmittedDetails != nil {
		omitted := weekly.OmittedDetails.Objectives + weekly.OmittedDetails.KeyResults
		if omitted > 0 {
			summaryText += fmt.Sprintf(" The saved detail shows a prioritized selection; %d additional %s available in Strategy.", omitted, pluralWord(omitted, "detail is", "details are"))
			protectedSummaryTokens = append(protectedSummaryTokens, fmt.Sprintf("%d additional %s available in Strategy", omitted, pluralWord(omitted, "detail is", "details are")))
		}
	}
	facts = append(facts, emailcopy.Fact{
		ReferenceID:     summaryReferenceID,
		Text:            summaryText,
		ProtectedTokens: protectedSummaryTokens,
		Required:        true,
	})
	factActions[summaryReferenceID] = digestActionStrategy
	fallbackRows = append(fallbackRows, notificationDigestCopyRow{Text: summaryText, URL: workspaceURL + "/strategy"})

	selectedObjectives, selectedKeyResults := selectWeeklyStrategyDetails(weekly.Objectives, weekly.KeyResults, rowLimit-1)

	for index, objective := range selectedObjectives {
		factReferenceID := fmt.Sprintf("strategy_objective_%d_%s", index+1, objective.ID.String())
		actionReferenceID := fmt.Sprintf("strategy_objective_action_%d_%s", index+1, objective.ID.String())
		factText := strategyObjectiveFactText(objective, weekly.StaleAfterDays)
		destination := strategyObjectiveNotificationURL(workspaceURL, item.NotificationID, objective.ID)
		facts = append(facts, emailcopy.Fact{
			ReferenceID:     factReferenceID,
			Text:            factText,
			EntityTokens:    nonEmptyStrings(objective.Name),
			ProtectedTokens: strategyObjectiveProtectedTokens(objective),
			Required:        true,
		})
		actionURLs[actionReferenceID] = destination
		factActions[factReferenceID] = actionReferenceID
		factLabels[factReferenceID] = objective.Name
		fallbackRows = append(fallbackRows, notificationDigestCopyRow{Text: factText, Label: objective.Name, URL: destination})
	}

	for index, keyResult := range selectedKeyResults {
		factReferenceID := fmt.Sprintf("strategy_key_result_%d_%s", index+1, keyResult.ID.String())
		actionReferenceID := fmt.Sprintf("strategy_key_result_action_%d_%s", index+1, keyResult.ID.String())
		factText := strategyKeyResultFactText(keyResult, weekly.StaleAfterDays)
		destination := strategyKeyResultNotificationURL(workspaceURL, item.NotificationID, keyResult.ID, keyResult.ObjectiveID)
		facts = append(facts, emailcopy.Fact{
			ReferenceID:     factReferenceID,
			Text:            factText,
			EntityTokens:    nonEmptyStrings(keyResult.Name, keyResult.ObjectiveName),
			ProtectedTokens: strategyKeyResultProtectedTokens(keyResult),
			Required:        true,
		})
		actionURLs[actionReferenceID] = destination
		factActions[factReferenceID] = actionReferenceID
		factLabels[factReferenceID] = keyResult.Name
		fallbackRows = append(fallbackRows, notificationDigestCopyRow{Text: factText, Label: keyResult.Name, URL: destination})
	}

	return facts, actionURLs, factActions, factLabels, fallbackRows
}

func selectWeeklyStrategyDetails(
	objectives []strategyObjectiveSnapshot,
	keyResults []strategyKeyResultSnapshot,
	limit int,
) ([]strategyObjectiveSnapshot, []strategyKeyResultSnapshot) {
	if limit <= 0 {
		return nil, nil
	}
	objectiveLimit := min(len(objectives), (limit+1)/2)
	keyResultLimit := min(len(keyResults), limit-objectiveLimit)
	objectiveLimit = min(len(objectives), limit-keyResultLimit)
	if remaining := limit - objectiveLimit - keyResultLimit; remaining > 0 {
		keyResultLimit = min(len(keyResults), keyResultLimit+remaining)
	}
	return objectives[:objectiveLimit], keyResults[:keyResultLimit]
}

func strategyObjectiveFactText(objective strategyObjectiveSnapshot, staleAfterDays int) string {
	parts := []string{objective.Name}
	if objective.Health != nil && strings.TrimSpace(*objective.Health) != "" {
		parts = append(parts, "health is "+strings.TrimSpace(*objective.Health))
	}
	if objective.Status != nil && strings.TrimSpace(objective.Status.Name) != "" {
		parts = append(parts, "status is "+strings.TrimSpace(objective.Status.Name))
	}
	if containsValue(objective.Reasons, "stale") {
		parts = append(parts, fmt.Sprintf("has not had a recent update; recent means within %d days", staleAfterDays))
	}
	if !objective.UpdatedAt.IsZero() {
		parts = append(parts, "last updated on "+objective.UpdatedAt.UTC().Format("January 2, 2006"))
	}
	if objective.EndDate != nil {
		parts = append(parts, "ends on "+objective.EndDate.UTC().Format("January 2, 2006"))
	}
	return strings.Join(parts, "; ") + "."
}

func strategyObjectiveProtectedTokens(objective strategyObjectiveSnapshot) []string {
	values := make([]string, 0, 6)
	if objective.Health != nil {
		values = append(values, "health is "+strings.TrimSpace(*objective.Health))
	}
	if objective.Status != nil {
		values = append(values, "status is "+strings.TrimSpace(objective.Status.Name))
	}
	if containsValue(objective.Reasons, "stale") {
		values = append(values, "has not had a recent update")
	}
	if !objective.UpdatedAt.IsZero() {
		values = append(values, "last updated on "+objective.UpdatedAt.UTC().Format("January 2, 2006"))
	}
	if objective.EndDate != nil {
		values = append(values, "ends on "+objective.EndDate.UTC().Format("January 2, 2006"))
	}
	return nonEmptyStrings(values...)
}

func strategyKeyResultFactText(keyResult strategyKeyResultSnapshot, staleAfterDays int) string {
	parts := []string{fmt.Sprintf("%s under %s", keyResult.Name, keyResult.ObjectiveName)}
	if keyResult.ObjectiveHealth != nil && strings.TrimSpace(*keyResult.ObjectiveHealth) != "" {
		parts = append(parts, "objective health is "+strings.TrimSpace(*keyResult.ObjectiveHealth))
	}
	if keyResult.ObjectiveStatus != nil && strings.TrimSpace(keyResult.ObjectiveStatus.Name) != "" {
		parts = append(parts, "objective status is "+strings.TrimSpace(keyResult.ObjectiveStatus.Name))
	}
	if strings.TrimSpace(keyResult.MeasurementType) != "" {
		parts = append(parts, "measurement is "+strings.TrimSpace(keyResult.MeasurementType))
	}
	if keyResult.CurrentValue != nil && keyResult.TargetValue != nil {
		parts = append(parts, fmt.Sprintf("current value is %s and target value is %s", strategyValue(keyResult.CurrentValue), strategyValue(keyResult.TargetValue)))
	}
	parts = append(parts, fmt.Sprintf("is incomplete and has not had a recent update; recent means within %d days", staleAfterDays))
	if !keyResult.UpdatedAt.IsZero() {
		parts = append(parts, "last updated on "+keyResult.UpdatedAt.UTC().Format("January 2, 2006"))
	}
	if keyResult.EndDate != nil {
		parts = append(parts, "ends on "+keyResult.EndDate.UTC().Format("January 2, 2006"))
	}
	return strings.Join(parts, "; ") + "."
}

func strategyKeyResultProtectedTokens(keyResult strategyKeyResultSnapshot) []string {
	values := []string{"is incomplete", "has not had a recent update"}
	if keyResult.ObjectiveHealth != nil {
		values = append(values, "objective health is "+strings.TrimSpace(*keyResult.ObjectiveHealth))
	}
	if keyResult.ObjectiveStatus != nil {
		values = append(values, "objective status is "+strings.TrimSpace(keyResult.ObjectiveStatus.Name))
	}
	if measurementType := strings.TrimSpace(keyResult.MeasurementType); measurementType != "" {
		values = append(values, "measurement is "+measurementType)
	}
	if keyResult.CurrentValue != nil && keyResult.TargetValue != nil {
		values = append(values,
			"current value is "+strategyValue(keyResult.CurrentValue),
			"target value is "+strategyValue(keyResult.TargetValue),
		)
	}
	if !keyResult.UpdatedAt.IsZero() {
		values = append(values, "last updated on "+keyResult.UpdatedAt.UTC().Format("January 2, 2006"))
	}
	if keyResult.EndDate != nil {
		values = append(values, "ends on "+keyResult.EndDate.UTC().Format("January 2, 2006"))
	}
	return nonEmptyStrings(values...)
}

func strategyObjectiveNotificationURL(workspaceURL string, notificationID, objectiveID uuid.UUID) string {
	return fmt.Sprintf(
		"%s/notifications/%s?entityId=%s&entityType=objective",
		workspaceURL,
		notificationID.String(),
		url.QueryEscape(objectiveID.String()),
	)
}

func strategyKeyResultNotificationURL(workspaceURL string, notificationID, keyResultID, objectiveID uuid.UUID) string {
	return fmt.Sprintf(
		"%s/notifications/%s?entityId=%s&entityType=key_result&objectiveId=%s",
		workspaceURL,
		notificationID.String(),
		url.QueryEscape(keyResultID.String()),
		url.QueryEscape(objectiveID.String()),
	)
}

func buildGeneratedNotificationDigestCopy(input notificationDigestCopyInput, output emailcopy.Output) (notificationDigestCopy, error) {
	knownFacts := make(map[string]struct{}, len(input.Request.Facts))
	for _, fact := range input.Request.Facts {
		knownFacts[fact.ReferenceID] = struct{}{}
	}
	rows := make([]notificationDigestCopyRow, 0, len(output.Rows))
	for _, row := range output.Rows {
		if _, exists := knownFacts[row.ReferenceID]; !exists {
			return notificationDigestCopy{}, fmt.Errorf("email copy row %q cites an unknown fact", row.ReferenceID)
		}
		destination := ""
		if actionReferenceID := input.FactActions[row.ReferenceID]; actionReferenceID != "" {
			var ok bool
			destination, ok = input.Actions[actionReferenceID]
			if !ok || strings.TrimSpace(destination) == "" {
				return notificationDigestCopy{}, fmt.Errorf("email copy row %q has no trusted destination", row.ReferenceID)
			}
		}
		rows = append(rows, notificationDigestCopyRow{
			Text:  notificationPlainText(row.Text, 360),
			Label: input.FactLabels[row.ReferenceID],
			URL:   destination,
		})
	}
	if len(rows) == 0 {
		return notificationDigestCopy{}, errors.New("email copy has no rows")
	}

	primaryCTA := notificationDigestCopyCTA{}
	preferredReferenceID := digestActionPrimary
	if input.HasStrategySnapshot {
		preferredReferenceID = digestActionStrategy
	}
	for _, cta := range output.CTAs {
		if cta.ReferenceID != preferredReferenceID {
			continue
		}
		if destination, ok := input.Actions[cta.ReferenceID]; ok {
			primaryCTA = notificationDigestCopyCTA{Label: notificationPlainText(cta.Label, 48), URL: destination}
			break
		}
	}
	if primaryCTA.URL == "" {
		primaryCTA = input.Fallback.CTA
	}

	introParts := []string{notificationPlainText(output.Intro.Text, 420)}
	if output.SenderProse != nil {
		introParts = append(introParts, notificationPlainText(output.SenderProse.Text, 320))
	}
	if output.ReplyPrompt != nil {
		introParts = append(introParts, notificationPlainText(output.ReplyPrompt.Text, 320))
	}

	copy := notificationDigestCopy{
		Subject:             notificationPlainText(output.Subject.Text, 90),
		Heading:             notificationPlainText(output.H1.Text, 110),
		Intro:               strings.Join(nonEmptyStrings(introParts...), " "),
		Rows:                rows,
		CTA:                 primaryCTA,
		NotificationsURL:    input.NotificationsURL,
		HasStrategySnapshot: input.HasStrategySnapshot,
	}
	if input.HasStrategySnapshot {
		copy.Sender = mailer.SenderProfileMaya
	}
	if copy.Subject == "" || copy.Heading == "" || copy.Intro == "" {
		return notificationDigestCopy{}, errors.New("email copy has an empty visible field")
	}
	return copy, nil
}

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
			if strings.Contains(rowHTML, escapedLabel) {
				rowHTML = strings.Replace(rowHTML, escapedLabel, fmt.Sprintf(
					`<a href="%s" style="%s">%s</a>`,
					stdhtml.EscapeString(row.URL),
					linkStyle,
					escapedLabel,
				), 1)
			}
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

func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(*value)
}

func statusName(status *strategyObjectiveStatusSnapshot) string {
	if status == nil {
		return ""
	}
	return strings.TrimSpace(status.Name)
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

func (h *handlers) markNotificationsEmailSent(ctx context.Context, notificationIDs []uuid.UUID) error {
	if len(notificationIDs) == 0 {
		return nil
	}

	query := `
		UPDATE notifications
		SET email_sent_at = CURRENT_TIMESTAMP
		WHERE notification_id = ANY(:notification_ids);
	`

	params := map[string]any{
		"notification_ids": pq.Array(notificationIDs),
	}

	if _, err := h.db.NamedExecContext(ctx, query, params); err != nil {
		h.log.Error(ctx, "Failed to mark notifications as emailed", "error", err)
		return fmt.Errorf("failed to mark notifications as emailed: %w", err)
	}

	return nil
}

// HandleNotificationEmail processes the notification email task.
func (h *handlers) HandleNotificationEmail(ctx context.Context, t *asynq.Task) error {
	var p tasks.NotificationEmailPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		h.log.Error(ctx, "Failed to unmarshal NotificationEmailPayload in Handlers", "error", err, "task_id", t.ResultWriter().TaskID())
		return fmt.Errorf("unmarshal payload failed: %w: %w", err, asynq.SkipRetry)
	}

	h.log.Info(ctx, "HANDLER: Processing NotificationEmail task",
		"notification_id", p.NotificationID,
		"recipient_id", p.RecipientID,
		"workspace_id", p.WorkspaceID,
		"task_id", t.ResultWriter().TaskID(),
	)

	// Single query to get all required data
	data, err := h.getNotificationEmailData(ctx, p.NotificationID)
	if err != nil {
		h.log.Error(ctx, "Failed to get notification data", "error", err, "task_id", t.ResultWriter().TaskID())
		return err
	}

	if data == nil {
		h.log.Info(ctx, "Notification not found, already read, recipient inactive, system recipient, or missing email - skipping email",
			"notification_id", p.NotificationID,
			"task_id", t.ResultWriter().TaskID())
		return nil
	}

	// Unmarshal the raw JSON message into NotificationMessage struct
	var notificationMsg NotificationMessage
	if err := json.Unmarshal(data.Message, &notificationMsg); err != nil {
		h.log.Error(ctx, "Failed to unmarshal notification message", "error", err, "notification_id", p.NotificationID, "raw_message", string(data.Message))
		return fmt.Errorf("failed to unmarshal notification message: %w", err)
	}

	if !data.EmailEnabled {
		h.log.Info(ctx, "Email notifications disabled for this type - skipping",
			"notification_id", p.NotificationID,
			"notification_type", data.NotificationType,
			"task_id", t.ResultWriter().TaskID())
		return nil
	}

	workspaceURL := fmt.Sprintf("https://%s.fortyone.app", data.WorkspaceSlug)
	digestData := NotificationEmailDigestData{
		RecipientID:   data.RecipientID,
		WorkspaceID:   data.WorkspaceID,
		UserEmail:     data.UserEmail,
		UserName:      data.UserName,
		WorkspaceName: data.WorkspaceName,
		WorkspaceSlug: data.WorkspaceSlug,
		WorkspaceRole: data.WorkspaceRole,
		Items: []NotificationEmailDigestItem{{
			NotificationID:   data.NotificationID,
			NotificationType: data.NotificationType,
			EntityType:       data.EntityType,
			EntityID:         data.EntityID,
			Title:            data.Title,
			Message:          data.Message,
			ActorName:        data.ActorName,
			FeedbackSlug:     data.FeedbackSlug,
		}},
	}
	if _, err := h.filterStrategyDigestForCurrentAccess(ctx, &digestData); err != nil {
		return fmt.Errorf("filter notification email for current access: %w", err)
	}
	if len(digestData.Items) == 0 {
		return h.markNotificationsEmailSent(ctx, []uuid.UUID{data.NotificationID})
	}
	copyInput, err := buildNotificationDigestCopyInput(digestData, workspaceURL)
	if err != nil {
		return fmt.Errorf("build notification email copy input: %w", err)
	}
	notificationCopy, copyErr := generateNotificationDigestCopy(ctx, h.emailCopy, copyInput)
	if copyErr != nil {
		h.log.Error(ctx, "Email copy generation failed; using deterministic notification copy", "error", copyErr, "task_id", t.ResultWriter().TaskID())
	}
	notificationMessage := renderNotificationDigestCopy(notificationCopy)

	notificationsSettingsURL := fmt.Sprintf("%s/settings/account/notifications", workspaceURL)
	if data.EntityType == "feedback" && data.FeedbackSlug != "" {
		notificationsSettingsURL = ""
	}

	mailData := map[string]any{
		"UserName":                 data.UserName,
		"ActorName":                data.ActorName,
		"UserEmail":                data.UserEmail,
		"WorkspaceName":            data.WorkspaceName,
		"WorkspaceURL":             workspaceURL,
		"NotificationTitle":        notificationCopy.Heading,
		"NotificationMessage":      notificationMessage,
		"NotificationType":         data.NotificationType,
		"NotificationCTAURL":       notificationCopy.CTA.URL,
		"NotificationCTALabel":     notificationCopy.CTA.Label,
		"NotificationsSettingsURL": notificationsSettingsURL,
	}

	if err := h.mailerService.SendTemplated(ctx, mailer.TemplatedEmail{
		To:       []string{data.UserEmail},
		Template: "notifications/notification",
		Subject:  notificationCopy.Subject,
		Data:     mailData,
		Sender:   notificationCopy.Sender,
	}); err != nil {
		h.log.Error(ctx, "Failed to send notification email", "error", err, "task_id", t.ResultWriter().TaskID())
		return err
	}

	if err := h.markNotificationsEmailSent(ctx, []uuid.UUID{data.NotificationID}); err != nil {
		return err
	}

	h.log.Info(ctx, "HANDLER: Successfully processed NotificationEmail task",
		"notification_id", p.NotificationID,
		"user_email", data.UserEmail,
		"subject", notificationCopy.Subject,
		"task_id", t.ResultWriter().TaskID())
	return nil
}

// HandleNotificationEmailDigest processes a coalesced notification email task.
func (h *handlers) HandleNotificationEmailDigest(ctx context.Context, t *asynq.Task) error {
	var p tasks.NotificationEmailDigestPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		h.log.Error(ctx, "Failed to unmarshal NotificationEmailDigestPayload in Handlers", "error", err, "task_id", t.ResultWriter().TaskID())
		return fmt.Errorf("unmarshal payload failed: %w: %w", err, asynq.SkipRetry)
	}

	h.log.Info(ctx, "HANDLER: Processing NotificationEmailDigest task",
		"recipient_id", p.RecipientID,
		"workspace_id", p.WorkspaceID,
		"task_id", t.ResultWriter().TaskID(),
	)

	data, err := h.getNotificationEmailDigestData(ctx, p.RecipientID, p.WorkspaceID)
	if err != nil {
		h.log.Error(ctx, "Failed to get notification email digest data", "error", err, "task_id", t.ResultWriter().TaskID())
		return err
	}

	if data == nil || len(data.Items) == 0 {
		h.log.Info(ctx, "No unread unsent notifications for digest - skipping email",
			"recipient_id", p.RecipientID,
			"workspace_id", p.WorkspaceID,
			"task_id", t.ResultWriter().TaskID())
		return nil
	}
	suppressedNotificationIDs, err := h.filterStrategyDigestForCurrentAccess(ctx, data)
	if err != nil {
		h.log.Error(ctx, "Failed to filter notification digest for current access", "error", err, "task_id", t.ResultWriter().TaskID())
		return err
	}
	if len(data.Items) == 0 {
		return h.markNotificationsEmailSent(ctx, suppressedNotificationIDs)
	}

	workspaceURL := fmt.Sprintf("https://%s.fortyone.app", data.WorkspaceSlug)
	copyInput, err := buildNotificationDigestCopyInput(*data, workspaceURL)
	if err != nil {
		h.log.Error(ctx, "Failed to build notification email digest facts", "error", err, "task_id", t.ResultWriter().TaskID())
		return err
	}
	digestCopy, copyErr := generateNotificationDigestCopy(ctx, h.emailCopy, copyInput)
	if copyErr != nil {
		h.log.Error(ctx, "Email copy generation failed; using deterministic notification digest copy", "error", copyErr, "task_id", t.ResultWriter().TaskID())
	}
	notificationMessage := renderNotificationDigestCopy(digestCopy)

	notificationsSettingsURL := fmt.Sprintf("%s/settings/account/notifications", workspaceURL)
	if feedbackOnlyDigest(data.Items) {
		notificationsSettingsURL = ""
	}
	mailData := map[string]any{
		"UserName":                 data.UserName,
		"ActorName":                "",
		"UserEmail":                data.UserEmail,
		"WorkspaceName":            data.WorkspaceName,
		"WorkspaceURL":             workspaceURL,
		"NotificationTitle":        digestCopy.Heading,
		"NotificationMessage":      notificationMessage,
		"NotificationType":         "notification_digest",
		"NotificationCTAURL":       digestCopy.CTA.URL,
		"NotificationCTALabel":     digestCopy.CTA.Label,
		"NotificationsSettingsURL": notificationsSettingsURL,
	}

	if err := h.mailerService.SendTemplated(ctx, mailer.TemplatedEmail{
		To:       []string{data.UserEmail},
		Template: "notifications/notification",
		Subject:  digestCopy.Subject,
		Data:     mailData,
		Sender:   digestCopy.Sender,
	}); err != nil {
		h.log.Error(ctx, "Failed to send notification email digest", "error", err, "task_id", t.ResultWriter().TaskID())
		return err
	}

	notificationIDs := make([]uuid.UUID, 0, len(data.Items)+len(suppressedNotificationIDs))
	for _, item := range data.Items {
		notificationIDs = append(notificationIDs, item.NotificationID)
	}
	notificationIDs = append(notificationIDs, suppressedNotificationIDs...)
	if err := h.markNotificationsEmailSent(ctx, notificationIDs); err != nil {
		return err
	}

	h.log.Info(ctx, "HANDLER: Successfully processed NotificationEmailDigest task",
		"recipient_id", p.RecipientID,
		"workspace_id", p.WorkspaceID,
		"user_email", data.UserEmail,
		"notifications_count", len(data.Items),
		"task_id", t.ResultWriter().TaskID())
	return nil
}
