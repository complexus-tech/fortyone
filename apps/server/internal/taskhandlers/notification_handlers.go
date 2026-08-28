package taskhandlers

import (
	"context"
	"encoding/json"
	"fmt"

	notificationsdomain "github.com/complexus-tech/projects-api/internal/modules/notifications/domain"
	"github.com/complexus-tech/projects-api/pkg/mailer"
	"github.com/complexus-tech/projects-api/pkg/tasks"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

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
	data, err := h.getNotificationEmailData(ctx, notificationsdomain.EmailNotificationQuery{
		Scope:          notificationsdomain.DeliveryScope{RecipientID: p.RecipientID, WorkspaceID: p.WorkspaceID},
		NotificationID: p.NotificationID,
	})
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
		h.log.Error(ctx, "Failed to unmarshal notification message", "error", err, "notification_id", p.NotificationID)
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
		return h.markNotificationsEmailSent(ctx, notificationsdomain.DeliveryScope{RecipientID: data.RecipientID, WorkspaceID: data.WorkspaceID}, []uuid.UUID{data.NotificationID})
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
	messageID := fmt.Sprintf("<notification-%s@fortyone.app>", data.NotificationID)
	plainText := renderNotificationDigestPlainText(notificationCopy)
	replyTo, err := h.prepareNotificationGuidanceThread(ctx, digestData, notificationCopy, messageID, plainText)
	if err != nil {
		return fmt.Errorf("prepare notification guidance reply thread: %w", err)
	}

	if err := h.mailerService.SendTemplated(ctx, mailer.TemplatedEmail{
		To:            []string{data.UserEmail},
		Template:      "notifications/notification",
		Subject:       notificationCopy.Subject,
		Data:          mailData,
		PlainTextBody: plainText,
		Sender:        notificationCopy.Sender,
		ReplyTo:       replyTo,
		MessageID:     messageID,
	}); err != nil {
		h.log.Error(ctx, "Failed to send notification email", "error", err, "task_id", t.ResultWriter().TaskID())
		return err
	}

	if err := h.markNotificationsEmailSent(ctx, notificationsdomain.DeliveryScope{RecipientID: data.RecipientID, WorkspaceID: data.WorkspaceID}, []uuid.UUID{data.NotificationID}); err != nil {
		return err
	}

	h.log.Info(ctx, "HANDLER: Successfully processed NotificationEmail task",
		"notification_id", p.NotificationID,
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
		return h.markNotificationsEmailSent(ctx, notificationsdomain.DeliveryScope{RecipientID: data.RecipientID, WorkspaceID: data.WorkspaceID}, suppressedNotificationIDs)
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
	messageID := notificationDigestMessageID(*data)
	plainText := renderNotificationDigestPlainText(digestCopy)
	replyTo, err := h.prepareNotificationGuidanceThread(ctx, *data, digestCopy, messageID, plainText)
	if err != nil {
		return fmt.Errorf("prepare notification digest reply thread: %w", err)
	}

	if err := h.mailerService.SendTemplated(ctx, mailer.TemplatedEmail{
		To:            []string{data.UserEmail},
		Template:      "notifications/notification",
		Subject:       digestCopy.Subject,
		Data:          mailData,
		PlainTextBody: plainText,
		Sender:        digestCopy.Sender,
		ReplyTo:       replyTo,
		MessageID:     messageID,
	}); err != nil {
		h.log.Error(ctx, "Failed to send notification email digest", "error", err, "task_id", t.ResultWriter().TaskID())
		return err
	}

	notificationIDs := make([]uuid.UUID, 0, len(data.Items)+len(suppressedNotificationIDs))
	for _, item := range data.Items {
		notificationIDs = append(notificationIDs, item.NotificationID)
	}
	notificationIDs = append(notificationIDs, suppressedNotificationIDs...)
	if err := h.markNotificationsEmailSent(ctx, notificationsdomain.DeliveryScope{RecipientID: data.RecipientID, WorkspaceID: data.WorkspaceID}, notificationIDs); err != nil {
		return err
	}

	h.log.Info(ctx, "HANDLER: Successfully processed NotificationEmailDigest task",
		"recipient_id", p.RecipientID,
		"workspace_id", p.WorkspaceID,
		"notifications_count", len(data.Items),
		"task_id", t.ResultWriter().TaskID())
	return nil
}
