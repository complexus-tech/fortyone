package taskhandlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"html"
	"strings"

	"github.com/complexus-tech/projects-api/pkg/brevo"
	"github.com/complexus-tech/projects-api/pkg/tasks"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

type Variable struct {
	Value string `json:"value"`
	Type  string `json:"type"` // "actor", "assignee", "field", "value", "date"
}

type NotificationMessage struct {
	Template  string              `json:"template"`
	Variables map[string]Variable `json:"variables"`
}

// NotificationEmailData represents all data needed for sending notification emails
type NotificationEmailData struct {
	NotificationID   uuid.UUID           `db:"notification_id"`
	NotificationType string              `db:"type"`
	Title            string              `db:"title"`
	Message          NotificationMessage `db:"message"`
	UserEmail        string              `db:"user_email"`
	UserName         string              `db:"user_name"`
	WorkspaceName    string              `db:"workspace_name"`
	WorkspaceSlug    string              `db:"workspace_slug"`
	EmailEnabled     bool                `db:"email_enabled"`
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
			htmlValue = fmt.Sprintf("<strong>%s</strong>", html.EscapeString(value))
		case "field":
			htmlValue = fmt.Sprintf("<em>%s</em>", html.EscapeString(value))
		case "value":
			htmlValue = fmt.Sprintf("<span style='color: #007bff;'>%s</span>", html.EscapeString(value))
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
			u.full_name AS user_name,
			w.name AS workspace_name,
			w.slug AS workspace_slug,
			CAST(COALESCE(np.preferences -> CAST(n.type AS TEXT) ->> 'email', 'true') AS BOOLEAN) AS email_enabled
		FROM
			notifications n
			INNER JOIN users u ON n.recipient_id = u.user_id
			INNER JOIN workspaces w ON n.workspace_id = w.workspace_id
			LEFT JOIN notification_preferences np ON n.recipient_id = np.user_id
			AND n.workspace_id = np.workspace_id
		WHERE
			n.notification_id = :notification_id
			AND n.read_at IS NULL;
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
		h.log.Info(ctx, "Notification not found, already read, or user inactive - skipping email",
			"notification_id", p.NotificationID,
			"task_id", t.ResultWriter().TaskID())
		return nil
	}

	if !data.EmailEnabled {
		h.log.Info(ctx, "Email notifications disabled for this type - skipping",
			"notification_id", p.NotificationID,
			"notification_type", data.NotificationType,
			"task_id", t.ResultWriter().TaskID())
		return nil
	}

	// Parse the notification message
	parsedMessage := parseNotificationMessage(data.Message)

	// Send email with real data
	workspaceURL := fmt.Sprintf("https://%s.complexus.app", data.WorkspaceSlug)

	if err := h.brevoService.SendEmailNotification(ctx, 3, brevo.EmailNotificationParams{
		UserName:            data.UserName,
		UserEmail:           data.UserEmail,
		WorkspaceName:       data.WorkspaceName,
		WorkspaceURL:        workspaceURL,
		NotificationTitle:   data.Title,
		NotificationMessage: parsedMessage.Text,
		NotificationType:    data.NotificationType,
	}); err != nil {
		h.log.Error(ctx, "Failed to send notification email", "error", err, "task_id", t.ResultWriter().TaskID())
		return err
	}

	h.log.Info(ctx, "HANDLER: Successfully processed NotificationEmail task",
		"notification_id", p.NotificationID,
		"user_email", data.UserEmail,
		"parsed_message", parsedMessage.Text,
		"task_id", t.ResultWriter().TaskID())
	return nil
}
