package taskhandlers

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	notificationsdomain "github.com/complexus-tech/projects-api/internal/modules/notifications/domain"
	"github.com/complexus-tech/projects-api/pkg/mailer"
	"github.com/complexus-tech/projects-api/pkg/tasks"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

// HandleNotificationEmail processes the notification email task.
func (h *handlers) HandleNotificationEmail(ctx context.Context, t *asynq.Task) error {
	// Older queued tasks must join the recipient/workspace batch too.
	return h.HandleNotificationEmailDigest(ctx, t)
}

// HandleNotificationEmailDigest processes a coalesced notification email task.
func (h *handlers) HandleNotificationEmailDigest(ctx context.Context, t *asynq.Task) error {
	return h.handleNotificationEmailDigestAt(ctx, t, time.Now().UTC())
}

func (h *handlers) handleNotificationEmailDigestAt(ctx context.Context, t *asynq.Task, now time.Time) error {
	taskID, _ := asynq.GetTaskID(ctx)
	var p tasks.NotificationEmailDigestPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		h.log.Error(ctx, "Failed to unmarshal NotificationEmailDigestPayload in Handlers", "error", err, "task_id", taskID)
		return fmt.Errorf("unmarshal payload failed: %w: %w", err, asynq.SkipRetry)
	}

	h.log.Info(ctx, "HANDLER: Processing NotificationEmailDigest task",
		"recipient_id", p.RecipientID,
		"workspace_id", p.WorkspaceID,
		"task_id", taskID,
	)

	data, err := h.getNotificationEmailDigestData(ctx, p.RecipientID, p.WorkspaceID)
	if err != nil {
		h.log.Error(ctx, "Failed to get notification email digest data", "error", err, "task_id", taskID)
		return err
	}

	if data == nil || len(data.Items) == 0 {
		h.log.Info(ctx, "No unread unsent notifications for digest - skipping email",
			"recipient_id", p.RecipientID,
			"workspace_id", p.WorkspaceID,
			"task_id", taskID)
		return nil
	}
	// Take a shared claim before re-reading so a briefing cannot consume the
	// same notification snapshot between our eligibility check and send.
	claimID := uuid.Nil
	completed := false
	if h.routineDeliveries != nil {
		claimID, err = h.routineDeliveries.ClaimRoutine(ctx, notificationsdomain.RoutineClaim{RecipientID: p.RecipientID, WorkspaceID: p.WorkspaceID, Key: notificationDigestMessageID(*data), Kind: "activity", LocalDate: now, Now: now})
		if err != nil {
			return err
		}
		if claimID == uuid.Nil {
			return nil
		}
		defer func() {
			if !completed {
				if err := h.failRoutine(ctx, claimID); err != nil {
					h.log.Error(ctx, "Release activity email claim", "error", err)
				}
			}
		}()
		data, err = h.getNotificationEmailDigestData(ctx, p.RecipientID, p.WorkspaceID)
		if err != nil {
			return err
		}
		if data == nil || len(data.Items) == 0 {
			completed = true
			return h.completeRoutine(ctx, notificationsdomain.RoutineCompletion{ID: claimID, Scope: notificationsdomain.DeliveryScope{RecipientID: p.RecipientID, WorkspaceID: p.WorkspaceID}, Now: now})
		}
	}
	suppressedNotificationIDs, err := h.filterStrategyDigestForCurrentAccess(ctx, data)
	if err != nil {
		h.log.Error(ctx, "Failed to filter notification digest for current access", "error", err, "task_id", taskID)
		return err
	}
	if len(data.Items) == 0 {
		if claimID != uuid.Nil {
			completed = true
			return h.completeRoutine(ctx, notificationsdomain.RoutineCompletion{ID: claimID, Scope: notificationsdomain.DeliveryScope{RecipientID: data.RecipientID, WorkspaceID: data.WorkspaceID}, NotificationIDs: suppressedNotificationIDs, Now: now})
		}
		return h.markNotificationsEmailSent(ctx, notificationsdomain.DeliveryScope{RecipientID: data.RecipientID, WorkspaceID: data.WorkspaceID}, suppressedNotificationIDs)
	}

	workspaceURL := fmt.Sprintf("https://%s.fortyone.app", data.WorkspaceSlug)
	copyInput, err := buildNotificationDigestCopyInput(*data, workspaceURL)
	if err != nil {
		h.log.Error(ctx, "Failed to build notification email digest facts", "error", err, "task_id", taskID)
		return err
	}
	digestCopy, copyErr := generateNotificationDigestCopy(ctx, h.emailCopy, copyInput)
	if copyErr != nil {
		h.log.Error(ctx, "Email copy generation failed; using deterministic notification digest copy", "error", copyErr, "task_id", taskID)
	}
	notificationMessage := renderNotificationDigestCopy(digestCopy)
	typedDigest := templateDigest(digestCopy)
	h.resolveDigestAvatars(ctx, &typedDigest)

	guidance, guidanceDay, err := h.activityGuidance(ctx, notificationsdomain.DeliveryScope{RecipientID: data.RecipientID, WorkspaceID: data.WorkspaceID}, now)
	if err != nil {
		return fmt.Errorf("build activity guidance: %w", err)
	}
	if len(guidance.Sections) > 0 {
		digestCopy.Subject = "Your workspace updates · " + data.WorkspaceName
		digestCopy.Heading = "Your workspace updates"
		digestCopy.Sender = mailer.SenderProfileMaya
		digestCopy.CTA.URL, digestCopy.CTA.Label = workspaceURL, "Open workspace"
	}

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
		"NotificationDigest":       typedDigest,
		"NotificationType":         "notification_digest",
		"NotificationCTAURL":       digestCopy.CTA.URL,
		"NotificationCTALabel":     digestCopy.CTA.Label,
		"NotificationsSettingsURL": notificationsSettingsURL,
	}
	if len(guidance.Sections) > 0 {
		mailData["NotificationSections"] = append(guidance.Sections, typedDigest)
	}
	messageID := notificationDigestMessageID(*data)
	plainText := renderNotificationDigestPlainText(digestCopy)
	if len(guidance.Sections) > 0 {
		var summary strings.Builder
		for _, section := range guidance.Sections {
			summary.WriteString(section.Intro + "\n")
			for _, row := range section.Rows {
				summary.WriteString(strings.TrimSpace(row.Label+" "+row.Text) + "\n" + row.URL + "\n\n")
			}
		}
		plainText = summary.String() + plainText
	}
	if notificationsSettingsURL != "" {
		plainText += "\n\nManage notifications: " + notificationsSettingsURL
	}
	replyTo, err := h.prepareNotificationGuidanceThread(ctx, *data, digestCopy, messageID, plainText, guidance.Targets...)
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
		h.log.Error(ctx, "Failed to send notification email digest", "error", err, "task_id", taskID)
		return err
	}

	notificationIDs := make([]uuid.UUID, 0, len(data.Items)+len(suppressedNotificationIDs))
	for _, item := range data.Items {
		notificationIDs = append(notificationIDs, item.NotificationID)
	}
	notificationIDs = append(notificationIDs, suppressedNotificationIDs...)
	completed = true // Once SMTP returns success, never release the claim as a failed send.
	if claimID != uuid.Nil {
		err = h.completeRoutine(ctx, notificationsdomain.RoutineCompletion{ID: claimID, Scope: notificationsdomain.DeliveryScope{RecipientID: data.RecipientID, WorkspaceID: data.WorkspaceID}, NotificationIDs: notificationIDs, GuidanceDate: guidanceDay, Sent: true, Now: now})
	} else {
		err = h.markNotificationsEmailSent(ctx, notificationsdomain.DeliveryScope{RecipientID: data.RecipientID, WorkspaceID: data.WorkspaceID}, notificationIDs)
	}
	if err != nil {
		return err
	}

	h.log.Info(ctx, "HANDLER: Successfully processed NotificationEmailDigest task",
		"recipient_id", p.RecipientID,
		"workspace_id", p.WorkspaceID,
		"notifications_count", len(data.Items),
		"task_id", taskID)
	return nil
}
