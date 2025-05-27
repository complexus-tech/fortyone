package brevo

import (
	"context"
)

const (
	TemplateOverdueStory    int64 = 1
	TemplateStoryAssigned   int64 = 2
	TemplateStoryCompleted  int64 = 3
	TemplateProjectInvite   int64 = 4
	TemplateWorkspaceInvite int64 = 5
	TemplatePasswordReset   int64 = 6
	TemplateWelcome         int64 = 7
)

// NotificationEmailParams represents common parameters for notification emails
type NotificationEmailParams struct {
	UserName      string `json:"USER_NAME"`
	UserEmail     string `json:"USER_EMAIL"`
	WorkspaceName string `json:"WORKSPACE_NAME"`
	WorkspaceURL  string `json:"WORKSPACE_URL"`
}

// StoryNotificationParams represents parameters specific to story notifications
type StoryNotificationParams struct {
	NotificationEmailParams
	StoryTitle string `json:"STORY_TITLE"`
	StoryURL   string `json:"STORY_URL"`
	DueDate    string `json:"DUE_DATE,omitempty"`
	AssignedBy string `json:"ASSIGNED_BY,omitempty"`
}

// InviteNotificationParams represents parameters for invitation emails
type InviteNotificationParams struct {
	NotificationEmailParams
	InviterName   string `json:"INVITER_NAME"`
	InviteURL     string `json:"INVITE_URL"`
	WorkspaceName string `json:"WORKSPACE_NAME,omitempty"`
	ExpiresAt     string `json:"EXPIRES_AT,omitempty"`
}

// SendStoryAssignedNotification sends a story assignment notification email using a template
func (s *Service) SendStoryAssignedNotification(ctx context.Context, params StoryNotificationParams) (*SendTemplatedEmailResponse, error) {
	templateParams := map[string]any{
		"USER_NAME":      params.UserName,
		"USER_EMAIL":     params.UserEmail,
		"WORKSPACE_NAME": params.WorkspaceName,
		"WORKSPACE_URL":  params.WorkspaceURL,
		"STORY_TITLE":    params.StoryTitle,
		"STORY_URL":      params.StoryURL,
		"ASSIGNED_BY":    params.AssignedBy,
		"DUE_DATE":       params.DueDate,
	}

	req := SendTemplatedEmailRequest{
		TemplateID: TemplateStoryAssigned,
		To: []EmailRecipient{
			{
				Email: params.UserEmail,
				Name:  params.UserName,
			},
		},
		Params: templateParams,
		Tags:   []string{"notification", "story-assigned"},
	}

	return s.SendTemplatedEmail(ctx, req)
}

// SendNotificationEmail is a generic function to send any notification email with custom template and params
func (s *Service) SendNotificationEmail(ctx context.Context, templateID int64, recipientEmail, recipientName string, params map[string]any, tags []string) (*SendTemplatedEmailResponse, error) {
	req := SendTemplatedEmailRequest{
		TemplateID: templateID,
		To: []EmailRecipient{
			{
				Email: recipientEmail,
				Name:  recipientName,
			},
		},
		Params: params,
		Tags:   tags,
	}

	return s.SendTemplatedEmail(ctx, req)
}
