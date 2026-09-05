package mailer

import (
	"encoding/xml"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEmailAssetsMatchLandingWordmarkAndExist(t *testing.T) {
	root := filepath.Join(findTemplatesDirForTest(t), "..", "..", "landing")
	assets := filepath.Join(root, "public", "email-assets", "v1")
	for _, name := range []string{"wordmark.png", "invitation.png", "invitation-accepted.png", "fonts/Inter-Regular.woff2", "fonts/Inter-Medium.woff2", "fonts/Inter-SemiBold.woff2", "fonts/Inter-Bold.woff2"} {
		info, err := os.Stat(filepath.Join(assets, name))
		require.NoError(t, err)
		require.Positive(t, info.Size())
		require.Equal(t, emailAssetBaseURL+name, emailAsset(name))
	}
	svg, err := os.ReadFile(filepath.Join(assets, "wordmark.svg"))
	require.NoError(t, err)
	var wordmark struct {
		Paths []struct {
			D string `xml:"d,attr"`
		} `xml:"path"`
	}
	require.NoError(t, xml.Unmarshal(svg, &wordmark))
	require.NotEmpty(t, wordmark.Paths)
	landing, err := os.ReadFile(filepath.Join(root, "src", "components", "ui", "logo.tsx"))
	require.NoError(t, err)
	for _, path := range wordmark.Paths {
		require.Contains(t, string(landing), path.D)
	}
}

func TestMayaReplySharesLayoutAndPreservesConfirmationText(t *testing.T) {
	rendered, err := RenderMayaReply("Ready to update Activation.", `<p>Hi Alex,</p><p>Based on your reply, here’s the change I can make.</p><div style="`+EmailStyleString("quietPanel")+`"><p>Activation: At Risk → On Track</p><p>The launch blocker is resolved.</p></div><p>No changes have been made yet.</p><p>Reply CONFIRM to apply this change.</p><p>Reply CANCEL to leave it unchanged.</p>`)
	require.NoError(t, err)
	for _, expected := range []string{defaultLogoURL, `class="email-card"`, "max-width:520px", "line-height: 20px", "CONFIRM", "CANCEL", "No changes have been made yet.", "<!--[if mso]>", "Inter-Regular.woff2"} {
		require.Contains(t, rendered, expected)
	}
	require.NotContains(t, rendered, `class="hero"`)
	require.NotContains(t, rendered, `class="button"`)
	require.NotContains(t, rendered, "#ZgotmplZ")
}

func TestSafeEmailHTMLRestylesLegacyAndUnstyledCopy(t *testing.T) {
	for legacy := range legacyEmailStyles {
		sanitized := string(safeEmailHTML(`<p style="` + legacy + `">Queued copy</p>`))
		require.Contains(t, sanitized, "Inter")
		require.NotContains(t, sanitized, "Geist")
	}
	data := map[string]any{"LogoURL": "https://other.example/logo.png", "EmailWidth": 999, "EmailHeroURL": "https://other.example/image.png"}
	prepareTemplateData("notifications/notification", data)
	require.Equal(t, defaultLogoURL, data["LogoURL"])
	require.Equal(t, 520, data["EmailWidth"])
	require.Empty(t, data["EmailHeroURL"])
	require.Empty(t, emailAsset("../../private"))
	rendered := string(safeEmailHTML(`<h3>Story update</h3><p>Review this.</p>`))
	require.Equal(t, 2, strings.Count(rendered, `style="`))
}

func TestLegacyNotificationContentUsesCompactRowsAndSignature(t *testing.T) {
	var rows strings.Builder
	for _, row := range [][2]string{
		{"2 overdue stories", "Decide what still needs doing and reset any dates that have slipped."},
		{"3 stories due this week", "Choose the most important one to move forward first."},
		{"1 mention waiting for you", "Catch up with your team before planning your next steps."},
	} {
		rows.WriteString(`<div style="` + EmailStyleString("notificationItem") + `"><h3 style="` + EmailStyleString("panelTitle") + `">` + row[0] + `</h3><p style="` + EmailStyleString("textNoMargin") + `">` + row[1] + `</p></div>`)
	}
	rendered := renderTemplateForTest(t, "notifications/notification", map[string]any{
		"Subject": "A little focus for the week.", "NotificationTitle": "A little focus for the week.",
		"UserName": "Alex", "EmailIsMaya": true, "NotificationMessage": rows.String(),
		"NotificationCTAURL": "https://example.com/my-work", "NotificationCTALabel": "Plan my week",
		"NotificationsSettingsURL": "https://example.com/settings/notifications",
	})
	require.Contains(t, rendered, "Your AI agent at FortyOne")
	require.Contains(t, rendered, "padding: 12px 0;")
	require.NotContains(t, rendered, `class="hero"`)
}
