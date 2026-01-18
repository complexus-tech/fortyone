package brevo

import (
	"context"
	"fmt"
)

// EmailNotificationParams represents parameters for notification emails
type EmailNotificationParams struct {
	Subject             string `json:"SUBJECT"`
	UserName            string `json:"USER_NAME"`
	ActorName           string `json:"ACTOR_NAME"`
	UserEmail           string `json:"USER_EMAIL"`
	WorkspaceURL        string `json:"WORKSPACE_URL"`
	WorkspaceName       string `json:"WORKSPACE_NAME"`
	NotificationTitle   string `json:"NOTIFICATION_TITLE"`
	NotificationMessage string `json:"NOTIFICATION_MESSAGE"`
	NotificationType    string `json:"NOTIFICATION_TYPE"`
}

// SendEmailNotification sends a notification email
func (service *Service) SendEmailNotification(ctx context.Context, templateId int64, params EmailNotificationParams) error {
	templateParams := map[string]any{
		"USER_NAME":            params.UserName,
		"ACTOR_NAME":           params.ActorName,
		"USER_EMAIL":           params.UserEmail,
		"WORKSPACE_NAME":       params.WorkspaceName,
		"WORKSPACE_URL":        params.WorkspaceURL,
		"NOTIFICATION_TITLE":   params.NotificationTitle,
		"NOTIFICATION_MESSAGE": params.NotificationMessage,
		"NOTIFICATION_TYPE":    params.NotificationType,
	}

	subject := fmt.Sprintf("Update from %s", params.ActorName)
	if params.Subject != "" {
		subject = params.Subject
	}

	req := SendTemplatedEmailRequest{
		TemplateID: templateId,
		To: []EmailRecipient{
			{
				Email: params.UserEmail,
				Name:  params.UserName,
			},
		},
		Subject: subject,
		Params:  templateParams,
		Tags:    []string{"notification", params.NotificationType},
	}

	return service.SendTemplatedEmail(ctx, req)
}
