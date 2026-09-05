package mailer

import (
	"bytes"
	"fmt"
	"html/template"
	"net/url"
	"strings"
	"time"

	emailtemplates "github.com/complexus-tech/projects-api/templates"
)

// Versioned public assets must be deployed on the landing site before the
// API/worker release. Keep old versions available for already-delivered mail.
const emailAssetBaseURL = "https://fortyone.app/email-assets/v1/"

func emailAsset(name string) string {
	switch name {
	case "icons/calendar.png", "icons/comment.png", "wordmark.png", "invitation.png", "invitation-accepted.png",
		"fonts/Inter-Regular.woff2", "fonts/Inter-Medium.woff2", "fonts/Inter-SemiBold.woff2", "fonts/Inter-Bold.woff2":
		return emailAssetBaseURL + name
	default:
		return ""
	}
}

func emailTemplateFuncs() template.FuncMap {
	return template.FuncMap{
		// html/template strips literal comments. Fixed delimiters preserve
		// Outlook conditionals while their contents remain contextually escaped.
		"emailActorText": actorText,
		"emailIcon":      func(name string) string { return emailAsset("icons/" + name + ".png") },
		"msoOnly":        func() template.HTML { return template.HTML(`<!--[if mso]>`) },
		"msoEnd":         func() template.HTML { return template.HTML(`<![endif]-->`) },
		"notMSO":         func() template.HTML { return template.HTML(`<!--[if !mso]><!-->`) },
		"notMSOEnd":      func() template.HTML { return template.HTML(`<!--<![endif]-->`) },
		"formatDate":     func(t time.Time) string { return t.Format("January 2, 2006") },
		"safeHTML":       safeEmailHTML, "emailStyle": emailStyle, "emailAsset": emailAsset,
	}
}

// Layout and artwork are owned by the renderer, not caller-supplied metadata.
func prepareTemplateData(name string, data map[string]any) {
	width := 520
	data["EmailHeroURL"], data["EmailHeroAlt"] = "", ""
	actionURL, actionLabel := "", ""
	value := func(key string) string { s, _ := data[key].(string); return s }
	switch name {
	case "invites/invitation":
		width = 480
		data["EmailHeroURL"] = emailAsset("invitation.png")
		data["EmailHeroAlt"] = "Two open doorways in warm peach and orange."
		actionURL, actionLabel = value("VerificationURL"), "Join "+value("WorkspaceName")
	case "invites/acceptance":
		width = 480
		data["EmailHeroURL"] = emailAsset("invitation-accepted.png")
		data["EmailHeroAlt"] = "Two paths come together at an open doorway."
		actionURL, actionLabel = value("LoginURL"), "Open "+value("WorkspaceName")
	case "notifications/notification":
		actionURL, actionLabel = value("NotificationCTAURL"), value("NotificationCTALabel")
	case "auth/verification":
		actionURL, actionLabel = value("VerificationURL"), "Log in to FortyOne"
	case "feedback/verification":
		actionURL, actionLabel = value("VerificationURL"), "Verify email"
	case "users/inactivity_warning":
		actionURL, actionLabel = value("LoginURL"), "Keep my account"
	case "workspaces/inactivity_warning":
		actionURL, actionLabel = value("WorkspaceURL"), "Keep my workspace"
	case "workspaces/deletion_scheduled_confirmation", "workspaces/deletion_scheduled_notification":
		actionURL, actionLabel = value("RestoreURL"), "Open workspace settings"
	case "workspaces/restored_confirmation", "workspaces/restored_notification":
		actionURL, actionLabel = value("WorkspaceURL"), "Open workspace"
	}
	if _, supplied := data["NotificationsSettingsURL"]; !supplied {
		if parsed, err := url.Parse(value("WorkspaceURL")); err == nil && parsed.Scheme == "https" && strings.HasSuffix(parsed.Hostname(), ".fortyone.app") && parsed.User == nil {
			data["NotificationsSettingsURL"] = "https://" + parsed.Host + "/settings/account/notifications"
		}
	}
	data["LogoURL"] = defaultLogoURL
	data["EmailWidth"], data["EmailInnerWidth"] = width, width-64
	data["EmailActionURL"], data["EmailActionLabel"] = actionURL, actionLabel
	data["EmailHasActions"] = actionURL != "" || name == "maya/reply"
}

var mayaReplyTemplate = template.Must(template.Must(template.New("").Funcs(emailTemplateFuncs()).Parse(emailtemplates.BaseLayout)).Parse(`
{{define "content"}}<h1 class="heading" style="{{emailStyle "heading"}}">{{.Subject}}</h1>{{safeHTML .Content}}{{end}}
{{define "actions"}}<p style="{{emailStyle "textNoMargin"}}"><strong>Maya</strong><br><span style="color:#72645d;font-size:14px;">Your AI agent at FortyOne</span></p>{{end}}
`))

// RenderMayaReply shares the transactional shell and sanitizes the fixed HTML
// vocabulary emitted by the email agent; model output cannot supply assets.
func RenderMayaReply(subject, content string, workspaceSlug ...string) (string, error) {
	data := map[string]any{"Subject": strings.TrimSpace(subject), "Content": content, "CompanyName": defaultCompanyName}
	if len(workspaceSlug) > 0 {
		slug := strings.TrimSpace(workspaceSlug[0])
		if slug != "" && strings.Trim(slug, "abcdefghijklmnopqrstuvwxyz0123456789-") == "" {
			data["WorkspaceURL"] = "https://" + slug + ".fortyone.app"
		}
	}
	prepareTemplateData("maya/reply", data)
	var output bytes.Buffer
	if err := mayaReplyTemplate.ExecuteTemplate(&output, "base", data); err != nil {
		return "", fmt.Errorf("render Maya reply: %w", err)
	}
	return output.String(), nil
}
