package taskhandlers

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"time"

	feedback "github.com/complexus-tech/projects-api/internal/modules/feedback/domain"
	notifications "github.com/complexus-tech/projects-api/internal/modules/notifications/domain"
	"github.com/complexus-tech/projects-api/pkg/jobs"
	"github.com/complexus-tech/projects-api/pkg/mailer"
	"github.com/complexus-tech/projects-api/pkg/tasks"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

// Older queued single-notification jobs join the same weekly workspace summary.
func (h *handlers) HandleNotificationEmail(ctx context.Context, t *asynq.Task) error {
	return h.HandleNotificationEmailDigest(ctx, t)
}

func (h *handlers) HandleNotificationEmailDigest(ctx context.Context, t *asynq.Task) error {
	return h.handleNotificationEmailDigestAt(ctx, t, time.Now().UTC())
}

func (h *handlers) handleNotificationEmailDigestAt(ctx context.Context, t *asynq.Task, now time.Time) error {
	var payload tasks.NotificationEmailDigestPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("unmarshal notification digest: %w: %w", err, asynq.SkipRetry)
	}
	scope := notifications.DeliveryScope{RecipientID: payload.RecipientID, WorkspaceID: payload.WorkspaceID}
	if err := scope.Validate(); err != nil {
		return fmt.Errorf("invalid digest scope: %w: %w", err, asynq.SkipRetry)
	}
	recipient, err := h.routineRecipient(ctx, scope)
	if err != nil {
		return err
	}
	timezone := "UTC"
	if recipient != nil {
		timezone = recipient.Timezone
	}
	date, ready := routineDeliveryDate(now, timezone)
	if h.routineDeliveries != nil && !ready {
		// The hourly sweep picks up pending content after the weekly delivery window opens.
		return nil
	}

	claimID := uuid.Nil
	accepted := false
	if h.routineDeliveries != nil {
		claimID, err = h.routineDeliveries.ClaimRoutine(ctx, notifications.RoutineClaim{
			RecipientID: scope.RecipientID, WorkspaceID: scope.WorkspaceID,
			Key: routineDeliveryKey(date), Kind: "activity", LocalDate: date, Now: now,
		})
		if err != nil || claimID == uuid.Nil {
			return err
		}
		defer func() {
			if !accepted {
				if err := h.failRoutine(ctx, claimID); err != nil {
					h.log.Error(ctx, "Release routine email claim", "error", err)
				}
			}
		}()
	}

	// Read under the claim. Completion covers this exact content, so an update
	// arriving during SMTP remains pending for next week's summary.
	data, err := h.getNotificationEmailDigestData(ctx, scope.RecipientID, scope.WorkspaceID)
	if err != nil {
		return err
	}
	if data == nil {
		data = routineDigestData(scope, recipient)
	}
	snapshots := notificationEmailSnapshots(data.Items)
	suppressed, err := h.filterStrategyDigestForCurrentAccess(ctx, data)
	if err != nil {
		return err
	}
	batch, err := h.prepareRoutineFeedback(ctx, recipient, now)
	if err != nil {
		return err
	}
	if batch != nil {
		defer func() {
			if !accepted {
				stateCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
				defer cancel()
				if err := h.feedbackDigest.FailDigestDelivery(stateCtx, batch.Completion.DeliveryID, "workspace summary was not sent"); err != nil {
					h.log.Error(ctx, "Release summary feedback claim", "error", err)
				}
			}
		}()
	}
	if len(data.Items) == 0 && batch == nil {
		if claimID == uuid.Nil {
			return h.markNotificationEmailSnapshotsSent(ctx, scope, snapshots)
		}
		err := h.completeRoutine(ctx, notifications.RoutineCompletion{ID: claimID, Scope: scope, NotificationIDs: suppressed, NotificationSnapshots: snapshots, Now: now})
		accepted = err == nil // An empty summary does not consume the weekly send.
		return err
	}

	email, guidanceDay, err := h.buildRoutineSummary(ctx, *data, batch, date, now)
	if err != nil {
		return err
	}
	if claimID != uuid.Nil {
		// Fence the SMTP attempt durably. An accepted send whose completion
		// fails must not be reclaimed and sent again within the same week.
		if err := h.routineDeliveries.BeginRoutineSend(ctx, claimID, scope, now); err != nil {
			return fmt.Errorf("begin routine send: %w", err)
		}
	}
	if err := h.mailerService.SendTemplated(ctx, email); err != nil {
		return fmt.Errorf("send workspace summary: %w", err)
	}
	accepted = true // Never release claims after SMTP has accepted the email.
	if claimID == uuid.Nil {
		return h.markNotificationEmailSnapshotsSent(ctx, scope, snapshots)
	}
	ids := make([]uuid.UUID, 0, len(data.Items)+len(suppressed))
	for _, item := range data.Items {
		ids = append(ids, item.NotificationID)
	}
	completion := notifications.RoutineCompletion{
		ID: claimID, Scope: scope, NotificationIDs: append(ids, suppressed...),
		NotificationSnapshots: snapshots, GuidanceDate: guidanceDay, Sent: true, Now: now,
	}
	if batch != nil {
		completion.FeedbackDigest = &batch.Completion
	}
	return h.completeRoutine(ctx, completion)
}

func routineDigestData(scope notifications.DeliveryScope, recipient *notifications.RoutineRecipient) *NotificationEmailDigestData {
	data := &NotificationEmailDigestData{RecipientID: scope.RecipientID, WorkspaceID: scope.WorkspaceID}
	if recipient != nil {
		data.UserEmail, data.UserName = recipient.Email, recipient.Name
		data.WorkspaceSlug, data.WorkspaceName = recipient.WorkspaceSlug, recipient.WorkspaceName
	}
	return data
}

func (h *handlers) prepareRoutineFeedback(ctx context.Context, recipient *notifications.RoutineRecipient, now time.Time) (*jobs.FeedbackDigestBatch, error) {
	// A feedback section must commit with the durable weekly delivery claim.
	if recipient == nil || h.feedbackDigest == nil || h.routineDeliveries == nil {
		return nil, nil
	}
	return jobs.PrepareFeedbackDigest(ctx, h.feedbackDigest, feedback.CoreDigestRecipient{
		UserID: recipient.UserID, WorkspaceID: recipient.WorkspaceID,
		UserEmail: recipient.Email, UserName: recipient.Name, Timezone: recipient.Timezone,
		WorkspaceName: recipient.WorkspaceName, WorkspaceSlug: recipient.WorkspaceSlug,
	}, now)
}

func (h *handlers) buildRoutineSummary(ctx context.Context, data NotificationEmailDigestData, batch *jobs.FeedbackDigestBatch, date, now time.Time) (mailer.TemplatedEmail, *time.Time, error) {
	workspaceURL := "https://" + data.WorkspaceSlug + ".fortyone.app"
	copy := notificationDigestCopy{
		Subject: "Your workspace updates · " + data.WorkspaceName, Heading: "Your workspace updates",
		Sender: mailer.SenderProfileMaya, CTA: notificationDigestCopyCTA{Label: "Open workspace", URL: workspaceURL},
	}
	sections := make([]mailer.Digest, 0, 4)
	if len(data.Items) > 0 {
		input, err := buildNotificationDigestCopyInput(data, workspaceURL)
		if err != nil {
			return mailer.TemplatedEmail{}, nil, err
		}
		var copyErr error
		copy, copyErr = generateNotificationDigestCopy(ctx, h.emailCopy, input)
		if copyErr != nil {
			h.log.Error(ctx, "Use deterministic notification summary copy", "error", copyErr)
		}
		activity := templateDigest(copy)
		h.resolveDigestAvatars(ctx, &activity)
		sections = append(sections, activity)
	}
	guidance, guidanceDay, err := h.activityGuidance(ctx, notifications.DeliveryScope{RecipientID: data.RecipientID, WorkspaceID: data.WorkspaceID}, now)
	if err != nil {
		return mailer.TemplatedEmail{}, nil, fmt.Errorf("build summary guidance: %w", err)
	}
	sections = append(guidance.Sections, sections...)
	targets := guidance.Targets
	if batch != nil {
		sections = append(sections, batch.Section)
		targets = append(targets, batch.Targets...)
	}
	if batch != nil || len(guidance.Sections) > 0 {
		copy.Subject, copy.Heading = "Your workspace updates · "+data.WorkspaceName, "Your workspace updates"
		copy.Sender = mailer.SenderProfileMaya
		copy.CTA = notificationDigestCopyCTA{Label: "Open workspace", URL: workspaceURL}
	}
	sections = limitRoutineEmailSections(sections, workspaceURL)
	settingsURL := workspaceURL + "/settings/account/notifications"
	if feedbackOnlyDigest(data.Items) && batch == nil {
		settingsURL = ""
	}
	mailData := map[string]any{
		"UserName": data.UserName, "ActorName": "", "UserEmail": data.UserEmail,
		"WorkspaceName": data.WorkspaceName, "WorkspaceURL": workspaceURL,
		"NotificationTitle": copy.Heading, "NotificationMessage": renderNotificationDigestCopy(copy),
		"NotificationType": "notification_digest", "NotificationCTAURL": copy.CTA.URL,
		"NotificationCTALabel": copy.CTA.Label, "NotificationsSettingsURL": settingsURL,
	}
	if len(sections) == 1 && batch == nil {
		mailData["NotificationDigest"] = sections[0]
	} else {
		mailData["NotificationSections"] = sections
	}
	messageID := notificationDigestMessageID(data)
	if batch != nil {
		digest := sha256.Sum256([]byte(messageID + ":" + date.Format(time.DateOnly) + ":" + batch.Completion.DeliveryID.String()))
		messageID = fmt.Sprintf("<workspace-summary-%x@fortyone.app>", digest[:16])
	}
	plainText := routineSectionsPlainText(sections) + "\n" + copy.CTA.Label + ": " + copy.CTA.URL
	if settingsURL != "" {
		plainText += "\n\nManage notifications: " + settingsURL
	}
	replyTo, err := h.prepareNotificationGuidanceThread(ctx, data, copy, messageID, plainText, targets...)
	if err != nil {
		return mailer.TemplatedEmail{}, nil, fmt.Errorf("prepare summary reply thread: %w", err)
	}
	return mailer.TemplatedEmail{
		To: []string{data.UserEmail}, Template: "notifications/notification", Subject: copy.Subject,
		Data: mailData, PlainTextBody: plainText, Sender: copy.Sender, ReplyTo: replyTo, MessageID: messageID,
	}, guidanceDay, nil
}
