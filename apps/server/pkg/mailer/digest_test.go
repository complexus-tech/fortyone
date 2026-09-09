package mailer

import (
	"fmt"
	"strings"
	"testing"
)

func TestAvatarColorsMatchFrontendUnicodeFixtures(t *testing.T) {
	for name, want := range map[string][2]string{
		"Sam Taylor": {"#27AE60", "#111827"}, "Joseph Mukorivo": {"#4A90E2", "#111827"}, "Ana García": {"#E67E22", "#111827"}, "👩🏽‍💻 Dev": {"#8E44AD", "#FFFFFF"}, "İpek": {"#27AE60", "#111827"}, "ΟΣ": {"#E67E22", "#111827"}, "\ufeff Sam   Taylor ": {"#27AE60", "#111827"},
	} {
		if got := avatarColor(name); got != want {
			t.Errorf("avatarColor(%q) = %v, want %v", name, got, want)
		}
	}
}

func TestAvatarMarkupEscapesTextAndUsesPhotoOrInitials(t *testing.T) {
	actor := EmailActor{Name: "Sam Taylor", AvatarURL: `https://example.com/avatar.png?a=1&b=2`}
	rendered := string(actorText("Sam changed the start date <script>alert(1)</script>.", actor))
	for _, expected := range []string{`width="20" height="20"`, `font-size:8px`, `width:4px`, `vertical-align:middle`, `alt="ST"`, `a=1&amp;b=2`, `&lt;script&gt;`} {
		assertContains(t, rendered, expected)
	}
	assertNotContains(t, rendered, "<script>")
	for _, unsafe := range []string{"javascript:alert(1)", "data:image/png;base64,abc", "https://name:secret@example.com/a.png", "profiles/private.png"} {
		actor.AvatarURL = unsafe
		rendered := string(actorText("Sam Taylor replied.", actor))
		assertNotContains(t, rendered, "<img")
		assertContains(t, rendered, ">ST</span>")
		assertContains(t, rendered, "background-color:#27AE60;color:#111827")
	}
}

func TestConsolidatedEmailRenderingAndFooter(t *testing.T) {
	base := "https://product.fortyone.app"
	section := Digest{Intro: "You have 10 unread updates in Product. Here are the first five."}
	names := []string{"Sam Taylor", "Ana García", "Joseph Mukorivo", "Maya Chen", "Alex Lee"}
	for i, name := range names {
		actor := EmailActor{Name: name}
		if i == 0 {
			actor.AvatarURL = "https://fortyone.app/images/avatars/product-lead.png"
		}
		section.Rows = append(section.Rows, DigestRow{Label: []string{"Prepare the launch", "Review onboarding", "Update the help center", "Plan the next sprint", "Check the release notes"}[i], Text: name + " changed the start date to September 8.", URL: fmt.Sprintf("%s/notifications/example-%d", base, i+1), Actor: actor, Icon: "calendar"})
	}
	section.Rows = append(section.Rows, DigestRow{Text: "View 5 more updates →", URL: base + "/notifications", More: true})
	data := map[string]any{"NotificationTitle": "10 updates to review in Product", "NotificationDigest": section, "WorkspaceURL": base, "NotificationCTAURL": base + "/notifications", "NotificationCTALabel": "View notifications"}
	rendered := renderTemplateForTest(t, "notifications/notification", data)
	for _, expected := range []string{"Manage notifications", base + "/settings/account/notifications", "View 5 more updates", "icons/calendar.png", `align="left"`, `align="right"`, `vertical-align:middle`, `font-size:8px`, "A product of Complexus"} {
		assertContains(t, rendered, expected)
	}
	if strings.Index(rendered, "Manage notifications") > strings.Index(rendered, "A product of Complexus") {
		t.Fatal("preferences must precede right-aligned company footer")
	}
	delete(data, "NotificationDigest")
	data["NotificationTitle"] = "Your workspace updates"
	data["EmailIsMaya"] = true
	data["NotificationCTAURL"], data["NotificationCTALabel"] = base, "Open workspace"
	data["NotificationSections"] = []Digest{{Intro: "Your priorities for this week", Rows: []DigestRow{{Label: "Review your assigned work", Text: "2 stories are overdue; 3 stories are due this week.", URL: base + "/my-work?tab=assigned"}}}, section}
	combined := renderTemplateForTest(t, "notifications/notification", data)
	for _, expected := range []string{"Your workspace updates", "Your priorities for this week", "View 5 more updates", "Your AI agent at FortyOne"} {
		assertContains(t, combined, expected)
	}
	// Explicit portal settings remain authoritative; a blank value suppresses
	// the private workspace settings link for external contributors.
	data["NotificationsSettingsURL"] = ""
	assertNotContains(t, renderTemplateForTest(t, "notifications/notification", data), "Manage notifications")
	response, err := RenderMayaReply("Change confirmed", "<p>The start date is now September 8.</p>", "product")
	if err != nil {
		t.Fatal(err)
	}
	assertContains(t, response, base+"/settings/account/notifications")
}

func TestNotificationDigestRendersActivityIcons(t *testing.T) {
	for _, icon := range []string{"calendar", "comment", "status", "priority"} {
		t.Run(icon, func(t *testing.T) {
			rendered := renderTemplateForTest(t, "notifications/notification", map[string]any{
				"NotificationTitle": "1 task updated in Art Circles",
				"NotificationDigest": Digest{Intro: "Here is the latest update for this task.", Rows: []DigestRow{{
					Icon: icon, Label: "Ticketing system mobile app", Text: "hector updated this task",
					URL: "https://art.fortyone.app/notifications/example", Actor: EmailActor{Name: "hector"},
				}}},
			})
			assertContains(t, rendered, `src="https://fortyone.app/email-assets/v1/icons/`+icon+`.png" width="24" height="24" alt=""`)
			assertContains(t, rendered, ">HE</span>")
		})
	}
}

func TestActivityEmphasisEscapesValuesAndPreservesActor(t *testing.T) {
	text := "Sam moved Backlog to To Do; highest is not High. <script>run()</script>"
	rendered := string(activityText(text, EmailActor{Name: "Sam Taylor"}, []string{"Backlog", "To Do", "High", "<script>run()</script>", ""}))
	for _, want := range []string{"<strong>Backlog</strong>", `<strong style="white-space:nowrap;">To Do</strong>`, "highest is not <strong>High</strong>", "<strong>&lt;script&gt;run()&lt;/script&gt;</strong>", ">ST</span>"} {
		assertContains(t, rendered, want)
	}
	assertNotContains(t, rendered, "<script>")
	assertContains(t, string(EmphasizeValues("8 Sep 2026 at 14:40–15:40", []string{"8 Sep", "8 Sep 2026 at 14:40–15:40"})), "<strong>8 Sep 2026 at 14:40–15:40</strong>")
	assertContains(t, string(EmphasizeValues("Résolu Résolution", []string{"Résolu"})), "<strong>Résolu</strong> Résolution")
}

func TestNotificationRendersChangeAndReasonSeparately(t *testing.T) {
	rendered := renderTemplateForTest(t, "notifications/notification", map[string]any{
		"NotificationTitle": "Task updated", "NotificationDigest": Digest{Rows: []DigestRow{{
			Label: "Scraping Segments Updates", Text: "Maya moved this task to 8 Sep 2026 at 14:40–15:40",
			Detail: "Your availability changed. <img src=x>", Highlights: []string{"8 Sep 2026 at 14:40–15:40"},
		}}},
	})
	assertContains(t, rendered, "<strong>8 Sep 2026 at 14:40–15:40</strong></p><p")
	assertContains(t, rendered, "&lt;img src=x&gt;")
	assertNotContains(t, rendered, "<img src=x>")
}
