package brevo

import (
	"context"
)

// EmailNotificationParams represents parameters for notification emails
type EmailNotificationParams struct {
	UserName      string `json:"NAME"`
	UserEmail     string `json:"USER_EMAIL"`
	WorkspaceURL  string `json:"WORKSPACE_URL"`
	WorkspaceName string `json:"WORKSPACE_NAME"`
}

// SendEmailNotification sends a notification email
func (service *Service) SendEmailNotification(ctx context.Context, templateId int64, params EmailNotificationParams) error {
	templateParams := map[string]any{
		"USER_NAME":      params.UserName,
		"USER_EMAIL":     params.UserEmail,
		"WORKSPACE_NAME": params.WorkspaceName,
		"WORKSPACE_URL":  params.WorkspaceURL,
	}

	req := SendTemplatedEmailRequest{
		TemplateID: templateId,
		To: []EmailRecipient{
			{
				Email: params.UserEmail,
				Name:  params.UserName,
			},
		},
		Params: templateParams,
		Tags:   []string{"notification", "general"},
	}

	return service.SendTemplatedEmail(ctx, req)
}
