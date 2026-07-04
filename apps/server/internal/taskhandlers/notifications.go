package taskhandlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"html"
	"strings"
	"time"

	"github.com/complexus-tech/projects-api/pkg/mailer"
	"github.com/complexus-tech/projects-api/pkg/tasks"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/lib/pq"
)

type Variable struct {
	Value string `json:"value"`
	Type  string `json:"type"`
}

type NotificationMessage struct {
	Template  string              `json:"template"`
	Variables map[string]Variable `json:"variables"`
}

// NotificationEmailData represents all data needed for sending notification emails
type NotificationEmailData struct {
	NotificationID   uuid.UUID       `db:"notification_id"`
	NotificationType string          `db:"type"`
	Title            string          `db:"title"`
	Message          json.RawMessage `db:"message"`
	UserEmail        string          `db:"user_email"`
	UserName         string          `db:"user_name"`
	ActorName        string          `db:"actor_name"`
	WorkspaceName    string          `db:"workspace_name"`
	WorkspaceSlug    string          `db:"workspace_slug"`
	EmailEnabled     bool            `db:"email_enabled"`
}

type NotificationEmailDigestItem struct {
	NotificationID   uuid.UUID       `db:"notification_id"`
	NotificationType string          `db:"type"`
	Title            string          `db:"title"`
	Message          json.RawMessage `db:"message"`
	CreatedAt        time.Time       `db:"created_at"`
	ActorName        string          `db:"actor_name"`
}

type NotificationEmailDigestData struct {
	UserEmail     string
	UserName      string
	WorkspaceName string
	WorkspaceSlug string
	Items         []NotificationEmailDigestItem
}

type notificationEmailDigestRow struct {
	NotificationID   uuid.UUID       `db:"notification_id"`
	NotificationType string          `db:"type"`
	Title            string          `db:"title"`
	Message          json.RawMessage `db:"message"`
	CreatedAt        time.Time       `db:"created_at"`
	UserEmail        string          `db:"user_email"`
	UserName         string          `db:"user_name"`
	ActorName        string          `db:"actor_name"`
	WorkspaceName    string          `db:"workspace_name"`
	WorkspaceSlug    string          `db:"workspace_slug"`
}

// ParsedMessage represents the final parsed notification message
type ParsedMessage struct {
	Text string
	HTML string
}

// parseNotificationMessage converts the template and variables into readable text
func parseNotificationMessage(msg NotificationMessage) ParsedMessage {
	template := msg.Template
	htmlTemplate := msg.Template

	// Replace template variables with actual values
	for key, variable := range msg.Variables {
		placeholder := "{" + key + "}"
		value := variable.Value

		// Create HTML version with styling based on variable type
		var htmlValue string
		switch variable.Type {
		case "actor":
			htmlValue = fmt.Sprintf("<strong style=\"%s\">%s</strong>", mailer.EmailStyleString("detailValue"), html.EscapeString(value))
		case "field":
			htmlValue = fmt.Sprintf("<em style=\"%s\">%s</em>", mailer.EmailStyleString("detailValue"), html.EscapeString(value))
		case "value", "date":
			htmlValue = fmt.Sprintf("<strong style=\"%s\">%s</strong>", mailer.EmailStyleString("detailValue"), html.EscapeString(value))
		default:
			htmlValue = html.EscapeString(value)
		}

		template = strings.ReplaceAll(template, placeholder, value)
		htmlTemplate = strings.ReplaceAll(htmlTemplate, placeholder, htmlValue)
	}

	return ParsedMessage{
		Text: template,
		HTML: htmlTemplate,
	}
}

// getNotificationEmailData retrieves all required data for sending notification email in a single query
func (h *handlers) getNotificationEmailData(ctx context.Context, notificationID uuid.UUID) (*NotificationEmailData, error) {
	query := `
		SELECT
			n.notification_id,
			n.type,
			n.title,
			n.message,
			u.email AS user_email,
			COALESCE(NULLIF(u.full_name, ''), u.username) AS user_name,
			COALESCE(NULLIF(actor_u.full_name, ''), actor_u.username) AS actor_name,
			w.name AS workspace_name,
			w.slug AS workspace_slug,
			CAST(COALESCE(np.preferences -> CAST(n.type AS TEXT) ->> 'email', 'true') AS BOOLEAN) AS email_enabled
		FROM
			notifications n
			INNER JOIN users u ON n.recipient_id = u.user_id
			INNER JOIN workspaces w ON n.workspace_id = w.workspace_id
			INNER JOIN users actor_u ON n.actor_id = actor_u.user_id
			LEFT JOIN notification_preferences np ON n.recipient_id = np.user_id
			AND n.workspace_id = np.workspace_id
		WHERE
			n.notification_id = :notification_id
			AND n.read_at IS NULL
			AND n.email_sent_at IS NULL
			AND u.is_active = true
			AND u.is_system = false
			AND NULLIF(TRIM(u.email), '') IS NOT NULL;
		`

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

func (h *handlers) getNotificationEmailDigestData(ctx context.Context, recipientID, workspaceID uuid.UUID) (*NotificationEmailDigestData, error) {
	query := `
		SELECT
			n.notification_id,
			n.type,
			n.title,
			n.message,
			n.created_at,
			u.email AS user_email,
			COALESCE(NULLIF(u.full_name, ''), u.username) AS user_name,
			COALESCE(NULLIF(actor_u.full_name, ''), actor_u.username) AS actor_name,
			w.name AS workspace_name,
			w.slug AS workspace_slug
		FROM
			notifications n
			INNER JOIN users u ON n.recipient_id = u.user_id
			INNER JOIN workspaces w ON n.workspace_id = w.workspace_id
			INNER JOIN users actor_u ON n.actor_id = actor_u.user_id
			LEFT JOIN notification_preferences np ON n.recipient_id = np.user_id
				AND n.workspace_id = np.workspace_id
		WHERE
			n.recipient_id = :recipient_id
			AND n.workspace_id = :workspace_id
			AND n.read_at IS NULL
			AND n.email_sent_at IS NULL
			AND u.is_active = true
			AND u.is_system = false
			AND NULLIF(TRIM(u.email), '') IS NOT NULL
			AND CAST(COALESCE(np.preferences -> CAST(n.type AS TEXT) ->> 'email', 'true') AS BOOLEAN) = true
		ORDER BY n.created_at ASC;
	`

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
			Title:            row.Title,
			Message:          row.Message,
			CreatedAt:        row.CreatedAt,
			ActorName:        row.ActorName,
		}
	}

	return &NotificationEmailDigestData{
		UserEmail:     rows[0].UserEmail,
		UserName:      rows[0].UserName,
		WorkspaceName: rows[0].WorkspaceName,
		WorkspaceSlug: rows[0].WorkspaceSlug,
		Items:         items,
	}, nil
}

func buildNotificationDigestSubject(workspaceName string, count int) string {
	itemText := "updates"
	if count == 1 {
		itemText = "update"
	}
	return fmt.Sprintf("%d %s in %s", count, itemText, workspaceName)
}

func formatNotificationDigestMessage(items []NotificationEmailDigestItem, workspaceURL string) (string, error) {
	if len(items) == 0 {
		return "", nil
	}

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
	`, textStyle, textStyle, html.EscapeString("Here's what changed while you were away."), listStyle)

	for index, item := range items {
		var notificationMsg NotificationMessage
		if err := json.Unmarshal(item.Message, &notificationMsg); err != nil {
			return "", fmt.Errorf("failed to unmarshal notification message %s: %w", item.NotificationID, err)
		}

		parsedMessage := parseNotificationMessage(notificationMsg)
		notificationURL := fmt.Sprintf("%s/notifications/%s", workspaceURL, item.NotificationID.String())
		itemStyle := defaultItemStyle
		if index == 0 {
			itemStyle = firstItemStyle
		}
		content += fmt.Sprintf(`
			<div style="%s">
				<p style="%s">%s for task <a href="%s" style="%s">%s</a></p>
			</div>
		`, itemStyle, messageStyle, parsedMessage.HTML, html.EscapeString(notificationURL), linkStyle, html.EscapeString(item.Title))
	}

	return content + "</div></div>", nil
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

	// Parse the notification message
	parsedMessage := parseNotificationMessage(notificationMsg)

	// Send email with real data
	workspaceURL := fmt.Sprintf("https://%s.fortyone.app", data.WorkspaceSlug)

	mailData := map[string]any{
		"UserName":                 data.UserName,
		"ActorName":                data.ActorName,
		"UserEmail":                data.UserEmail,
		"WorkspaceName":            data.WorkspaceName,
		"WorkspaceURL":             workspaceURL,
		"NotificationTitle":        data.Title,
		"NotificationMessage":      parsedMessage.HTML,
		"NotificationType":         data.NotificationType,
		"NotificationCTAURL":       fmt.Sprintf("%s/notifications", workspaceURL),
		"NotificationCTALabel":     "Open notifications",
		"NotificationsSettingsURL": fmt.Sprintf("%s/settings/account/notifications", workspaceURL),
	}

	if err := h.mailerService.SendTemplated(ctx, mailer.TemplatedEmail{
		To:       []string{data.UserEmail},
		Template: "notifications/notification",
		Subject:  data.Title,
		Data:     mailData,
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
		"parsed_message", parsedMessage.Text,
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

	workspaceURL := fmt.Sprintf("https://%s.fortyone.app", data.WorkspaceSlug)
	notificationMessage, err := formatNotificationDigestMessage(data.Items, workspaceURL)
	if err != nil {
		h.log.Error(ctx, "Failed to format notification email digest", "error", err, "task_id", t.ResultWriter().TaskID())
		return err
	}

	subject := buildNotificationDigestSubject(data.WorkspaceName, len(data.Items))
	mailData := map[string]any{
		"UserName":                 data.UserName,
		"ActorName":                "",
		"UserEmail":                data.UserEmail,
		"WorkspaceName":            data.WorkspaceName,
		"WorkspaceURL":             workspaceURL,
		"NotificationTitle":        subject,
		"NotificationMessage":      notificationMessage,
		"NotificationType":         "notification_digest",
		"NotificationCTAURL":       fmt.Sprintf("%s/notifications", workspaceURL),
		"NotificationCTALabel":     "Open notifications",
		"NotificationsSettingsURL": fmt.Sprintf("%s/settings/account/notifications", workspaceURL),
	}

	if err := h.mailerService.SendTemplated(ctx, mailer.TemplatedEmail{
		To:       []string{data.UserEmail},
		Template: "notifications/notification",
		Subject:  subject,
		Data:     mailData,
	}); err != nil {
		h.log.Error(ctx, "Failed to send notification email digest", "error", err, "task_id", t.ResultWriter().TaskID())
		return err
	}

	notificationIDs := make([]uuid.UUID, len(data.Items))
	for i, item := range data.Items {
		notificationIDs[i] = item.NotificationID
	}
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
