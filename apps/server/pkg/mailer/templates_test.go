package mailer

import (
	"bytes"
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTemplatesRenderApprovedEmailSystem(t *testing.T) {
	rendered := renderTemplateForTest(t, "auth/verification", map[string]any{
		"VerificationURL": "https://cloud.fortyone.app/login/verify?token=abc&state=xyz", "ExpiresIn": "10 minutes", "OTP": "827657",
	})
	for _, expected := range []string{defaultLogoURL, "font-family:Inter", "fonts/Inter-Regular.woff2", "font-size: 23px", "line-height: 21px", "font-size:21px!important", "#fff6ef", `bgcolor="#ffffff"`, `class="email-card"`, `<!--[if mso]>`, `v:roundrect`, `width:456px`, "827657", "FortyOne by Complexus LLC", "token=abc&amp;state=xyz"} {
		assertContains(t, rendered, expected)
	}
	for _, rejected := range []string{"Geist", "/images/logo.png", "#ZgotmplZ", "<no value>", `class="eyebrow"`, `src="data:`} {
		assertNotContains(t, rendered, rejected)
	}
}

func TestInvitationArtworkAndDestinations(t *testing.T) {
	for _, tc := range []struct{ name, asset, urlKey string }{
		{"invites/invitation", "invitation.png", "VerificationURL"},
		{"invites/acceptance", "invitation-accepted.png", "LoginURL"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rendered := renderTemplateForTest(t, tc.name, map[string]any{tc.urlKey: "https://cloud.fortyone.app/target?token=original", "WorkspaceName": `Acme <script>alert(1)</script>`, "InviterName": "Joseph", "InviteeName": "Alex", "Role": "Member", "ExpiresIn": "7 days"})
			assertContains(t, rendered, emailAsset(tc.asset))
			assertContains(t, rendered, "max-width:480px")
			assertContains(t, rendered, "width:416px")
			assertContains(t, rendered, `width="126" height="31"`)
			assertContains(t, rendered, "https://cloud.fortyone.app/target?token=original")
			assertContains(t, rendered, "&lt;script&gt;")
			assertNotContains(t, rendered, "<script>")
		})
	}
}

func TestWorkspaceLinkEmailsPreserveRecoveryInstructions(t *testing.T) {
	for _, name := range []string{"workspaces/deletion_scheduled_confirmation", "workspaces/deletion_scheduled_notification"} {
		rendered := renderTemplateForTest(t, name, map[string]any{"WorkspaceName": "Acme", "WorkspaceURL": "https://cloud.fortyone.app/acme", "RestoreURL": "https://cloud.fortyone.app/acme/settings", "DeletionTime": "48 hours", "ActorName": "Joseph", "ActorEmail": "joseph@example.com"})
		for _, expected := range []string{"48 hours", "https://cloud.fortyone.app/acme/settings", "Workspace link", "https://cloud.fortyone.app/acme"} {
			assertContains(t, rendered, expected)
		}
	}
}

func TestNotificationEmailHasOptionalActionAndPreferences(t *testing.T) {
	data := map[string]any{"NotificationTitle": "Your story changed", "UserName": "Alex", "NotificationMessage": "<h3>Onboarding</h3><p>Ready for review.</p>", "NotificationsSettingsURL": "https://cloud.fortyone.app/settings/notifications"}
	rendered := renderTemplateForTest(t, "notifications/notification", data)
	assertContains(t, rendered, "Ready for review.")
	assertContains(t, rendered, EmailStyleString("panelTitle"))
	assertContains(t, rendered, "Manage notification preferences")
	assertNotContains(t, rendered, `class="actions"`)
	assertNotContains(t, rendered, `class="hero"`)
	data["NotificationCTAURL"], data["NotificationCTALabel"] = "https://cloud.fortyone.app/work", "View my work"
	rendered = renderTemplateForTest(t, "notifications/notification", data)
	assertContains(t, rendered, `class="actions"`)
	assertContains(t, rendered, "View my work")
}

func TestAllEmailTemplatesRenderWithApprovedLayout(t *testing.T) {
	testCases := map[string]map[string]any{
		"feedback/verification": {"PortalName": "Acme feedback", "VerificationURL": "https://example.com/verify", "Code": "654321", "ExpiresIn": "10 minutes"},
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
			assertNotContains(t, rendered, `class="eyebrow"`)
		})
	}
}

func renderTemplateForTest(t *testing.T, templateName string, data map[string]any) string {
	t.Helper()

	templatesDir := findTemplatesDirForTest(t)
	basePath := filepath.Join(templatesDir, "layouts", "base.html")
	contentPath := filepath.Join(templatesDir, templateName+".html")

	tmpl, err := template.New("").Funcs(emailTemplateFuncs()).ParseFiles(basePath, contentPath)
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

	prepareTemplateData(templateName, renderData)

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "base", renderData); err != nil {
		t.Fatalf("render template %s: %v", templateName, err)
	}

	if outputDir := os.Getenv("FORTYONE_EMAIL_PREVIEW_DIR"); outputDir != "" {
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			t.Fatal(err)
		}
		path := filepath.Join(outputDir, strings.ReplaceAll(templateName, "/", "-")+".html")
		if err := os.WriteFile(path, buf.Bytes(), 0644); err != nil {
			t.Fatal(err)
		}
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

func assertNotContains(t *testing.T, value string, unexpected string) {
	t.Helper()

	if strings.Contains(value, unexpected) {
		t.Fatalf("expected rendered template not to contain %q", unexpected)
	}
}
