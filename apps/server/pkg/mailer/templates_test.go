package mailer

import (
	"bytes"
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestTemplatesRenderApprovedEmailSystem(t *testing.T) {
	rendered := renderTemplateForTest(t, "auth/verification", map[string]any{
		"Subject":         "Your login link for FortyOne",
		"VerificationURL": "https://projects.fortyone.app/login/verify",
		"ExpiresIn":       "10 minutes",
		"OTP":             "827657",
	})

	assertContains(t, rendered, "fonts.googleapis.com/css2?family=Geist:wght@400;500;600")
	assertContains(t, rendered, "@font-face")
	assertContains(t, rendered, "font-family: 'Geist'")
	assertContains(t, rendered, `font-family: "Geist", Helvetica, Arial, sans-serif`)
	assertContains(t, rendered, "background-color: #ffffff")
	assertContains(t, rendered, ".email-title")
	assertContains(t, rendered, `style="margin: 0 0 14px; color: #6f6c67; font-family: Geist, Helvetica, Arial, sans-serif; font-size: 12px; font-weight: 600; line-height: 1.4; text-transform: uppercase;"`)
	assertContains(t, rendered, `style="margin: 0; max-width: 520px; color: #111111; font-family: Geist, Helvetica, Arial, sans-serif; font-size: 30px; font-weight: 600; line-height: 1.16;"`)
	assertContains(t, rendered, `style="margin: 0 0 18px; color: #303030; font-family: Geist, Helvetica, Arial, sans-serif; font-size: 15px; line-height: 1.62;"`)
	assertContains(t, rendered, `style="display: inline-block; padding: 13px 22px; border-radius: 8px; background-color: #111111; color: #ffffff; font-family: Geist, Helvetica, Arial, sans-serif; font-size: 15px; font-weight: 600; line-height: 1.35; text-align: center; text-decoration: none;"`)
	assertContains(t, rendered, "font-weight: 600")
	assertContains(t, rendered, "font-size: 30px")
	assertContains(t, rendered, "padding: 13px 22px")
	assertContains(t, rendered, ".verification-code")
	assertContains(t, rendered, ".email-body .verification-code")
	assertContains(t, rendered, `font-family: "SFMono-Regular", "Roboto Mono"`)
	assertContains(t, rendered, `style="margin: 0; color: #111111; font-family: SFMono-Regular, Consolas, monospace; font-size: 30px; font-weight: 600; line-height: 1.1;"`)
	assertContains(t, rendered, ".email-body .security-note")
	assertContains(t, rendered, "FortyOne by Complexus LLC")

	if strings.Contains(rendered, "Inter") {
		t.Fatalf("email template should not reference Inter")
	}
	if strings.Contains(rendered, "Jost") {
		t.Fatalf("email template should not reference Jost")
	}
	if strings.Contains(rendered, "wght@400;500;600;700") {
		t.Fatalf("email template should not load bold font weight")
	}
	if strings.Contains(rendered, "min-height: 48px") {
		t.Fatalf("email buttons should not use min-height because it adds extra height on top of padding in email clients")
	}
}

func TestWorkspaceLinkEmailsRenderUnderlinedLinks(t *testing.T) {
	for _, templateName := range []string{
		"workspaces/deletion_scheduled_confirmation",
		"workspaces/deletion_scheduled_notification",
	} {
		t.Run(templateName, func(t *testing.T) {
			rendered := renderTemplateForTest(t, templateName, map[string]any{
				"Subject":       "Workspace deletion scheduled",
				"WorkspaceName": "Art Circles",
				"WorkspaceURL":  "https://projects.fortyone.app/workspaces/art-circles",
				"RestoreURL":    "https://projects.fortyone.app/workspaces/art-circles/settings",
				"DeletionTime":  "7 days",
				"ActorName":     "Joseph Mukorivo",
				"ActorEmail":    "joseph@example.com",
			})

			assertContains(t, rendered, "Workspace link")
			assertContains(t, rendered, "class=\"workspace-link\"")
			assertContains(t, rendered, `style="display: inline-block; overflow-wrap: anywhere; color: #111111; font-family: Geist, Helvetica, Arial, sans-serif; font-size: 14px; line-height: 1.5; text-decoration: underline; text-decoration-thickness: 1px; text-underline-offset: 3px;"`)
			assertContains(t, rendered, "margin-bottom: 4px")
			assertContains(t, rendered, "https://projects.fortyone.app/workspaces/art-circles")
		})
	}
}

func TestNotificationEmailRendersInlineMessageStyles(t *testing.T) {
	rendered := renderTemplateForTest(t, "notifications/notification", map[string]any{
		"NotificationTitle":        "3 tasks need attention",
		"UserName":                 "Joseph Mukorivo",
		"NotificationMessage":      `<h3 style="margin: 0 0 12px; color: #111111; font-family: Geist, Helvetica, Arial, sans-serif; font-size: 17px; font-weight: 600; line-height: 1.3;">What's coming up</h3><p style="margin: 0 0 12px; color: #303030; font-family: Geist, Helvetica, Arial, sans-serif; font-size: 15px; line-height: 1.62;">You have 3 tasks that need attention.</p>`,
		"WorkspaceName":            "Art Circles",
		"NotificationCTAURL":       "https://projects.fortyone.app/work",
		"NotificationCTALabel":     "View my work",
		"NotificationsSettingsURL": "https://projects.fortyone.app/settings/notifications",
	})

	assertContains(t, rendered, `style="margin: 28px 0; padding: 22px 24px; border-radius: 8px; background-color: #f7f7f7; color: #303030; font-family: Geist, Helvetica, Arial, sans-serif; font-size: 15px; line-height: 1.62;"`)
	assertContains(t, rendered, `style="margin: 0 0 12px; color: #111111; font-family: Geist, Helvetica, Arial, sans-serif; font-size: 17px; font-weight: 600; line-height: 1.3;"`)
	assertContains(t, rendered, `style="margin: 0 0 12px; color: #303030; font-family: Geist, Helvetica, Arial, sans-serif; font-size: 15px; line-height: 1.62;"`)
}

func TestAllEmailTemplatesRenderWithApprovedLayout(t *testing.T) {
	testCases := map[string]map[string]any{
		"auth/verification": {
			"VerificationURL": "https://projects.fortyone.app/login/verify",
			"ExpiresIn":       "10 minutes",
			"OTP":             "827657",
		},
		"auth/verification_mobile": {
			"ExpiresIn": "10 minutes",
			"OTP":       "827657",
		},
		"notifications/notification": {
			"ActorName":                "",
			"NotificationTitle":        "3 tasks need attention",
			"UserName":                 "Joseph Mukorivo",
			"NotificationMessage":      `<h3>What's coming up</h3><p>You have 3 tasks that need attention.</p>`,
			"WorkspaceName":            "Art Circles",
			"NotificationCTAURL":       "https://projects.fortyone.app/work",
			"NotificationCTALabel":     "View my work",
			"NotificationsSettingsURL": "https://projects.fortyone.app/settings/notifications",
		},
		"invites/invitation": {
			"WorkspaceName":   "Art Circles",
			"ExpiresIn":       "7 days",
			"Role":            "Editor",
			"InviterName":     "Joseph Mukorivo",
			"VerificationURL": "https://projects.fortyone.app/invites/accept",
		},
		"invites/acceptance": {
			"InviteeName":   "Maya Chen",
			"WorkspaceName": "Art Circles",
			"Role":          "Editor",
			"LoginURL":      "https://projects.fortyone.app",
		},
		"users/inactivity_warning": {
			"UserName": "Joseph",
			"LoginURL": "https://projects.fortyone.app/login",
		},
		"workspaces/inactivity_warning": {
			"WorkspaceName": "Art Circles",
			"WorkspaceURL":  "https://projects.fortyone.app/workspaces/art-circles",
		},
		"workspaces/deletion_scheduled_confirmation": {
			"WorkspaceName": "Art Circles",
			"DeletionTime":  "7 days",
			"RestoreURL":    "https://projects.fortyone.app/workspaces/art-circles/settings",
			"WorkspaceURL":  "https://projects.fortyone.app/workspaces/art-circles",
		},
		"workspaces/deletion_scheduled_notification": {
			"WorkspaceName": "Art Circles",
			"ActorName":     "Joseph Mukorivo",
			"ActorEmail":    "joseph@example.com",
			"DeletionTime":  "7 days",
			"RestoreURL":    "https://projects.fortyone.app/workspaces/art-circles/settings",
			"WorkspaceURL":  "https://projects.fortyone.app/workspaces/art-circles",
		},
		"workspaces/restored_confirmation": {
			"WorkspaceName": "Art Circles",
			"WorkspaceURL":  "https://projects.fortyone.app/workspaces/art-circles",
			"ActorName":     "Joseph Mukorivo",
			"ActorEmail":    "joseph@example.com",
		},
		"workspaces/restored_notification": {
			"WorkspaceName": "Art Circles",
			"WorkspaceURL":  "https://projects.fortyone.app/workspaces/art-circles",
			"ActorName":     "Joseph Mukorivo",
			"ActorEmail":    "joseph@example.com",
		},
	}

	for templateName, data := range testCases {
		t.Run(templateName, func(t *testing.T) {
			rendered := renderTemplateForTest(t, templateName, data)

			assertContains(t, rendered, "class=\"email-shell\"")
			assertContains(t, rendered, "class=\"email-inner\"")
			assertContains(t, rendered, "class=\"email-footer\"")
			assertContains(t, rendered, "FortyOne by Complexus LLC")
		})
	}
}

func renderTemplateForTest(t *testing.T, templateName string, data map[string]any) string {
	t.Helper()

	templatesDir := findTemplatesDirForTest(t)
	basePath := filepath.Join(templatesDir, "layouts", "base.html")
	contentPath := filepath.Join(templatesDir, templateName+".html")

	tmpl, err := template.New("").Funcs(template.FuncMap{
		"formatDate": func(t time.Time) string {
			return t.Format("January 2, 2006")
		},
		"safeHTML": func(value string) template.HTML {
			return template.HTML(value)
		},
		"emailStyle": emailStyle,
	}).ParseFiles(basePath, contentPath)
	if err != nil {
		t.Fatalf("parse template %s: %v", templateName, err)
	}

	renderData := map[string]any{
		"Year":        2026,
		"LogoURL":     defaultLogoURL,
		"CompanyName": defaultCompanyName,
		"Subject":     "FortyOne",
	}
	for key, value := range data {
		renderData[key] = value
	}

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "base", renderData); err != nil {
		t.Fatalf("render template %s: %v", templateName, err)
	}

	return buf.String()
}

func findTemplatesDirForTest(t *testing.T) string {
	t.Helper()

	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}

	for {
		candidate := filepath.Join(wd, "templates")
		if _, err := os.Stat(filepath.Join(candidate, "layouts", "base.html")); err == nil {
			return candidate
		}

		parent := filepath.Dir(wd)
		if parent == wd {
			t.Fatalf("templates directory not found")
		}
		wd = parent
	}
}

func assertContains(t *testing.T, value string, expected string) {
	t.Helper()

	if !strings.Contains(value, expected) {
		t.Fatalf("expected rendered template to contain %q", expected)
	}
}
