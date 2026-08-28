//go:build integration

package notificationsrepository

import (
	"context"
	"errors"
	"testing"
	"time"

	notificationsdomain "github.com/complexus-tech/projects-api/internal/modules/notifications/domain"
)

func testNotificationPortalScope(t *testing.T, ctx context.Context, fixture notificationIntegrationFixture) {
	t.Helper()

	staffActorInput := fixture.feedbackNotification(fixture.recipientA, notificationDedupeKey("portal-staff-actor"))
	staffActorInput.ActorID = fixture.actorA
	staffActorNotification, inserted, err := fixture.repo.Create(ctx, staffActorInput)
	if err != nil || !inserted {
		t.Fatalf("authorized feedback staff actor create = %t, %v", inserted, err)
	}
	mustNotificationExec(t, ctx, fixture.postgres.Pool, `DELETE FROM notifications WHERE notification_id = $1`, staffActorNotification.ID)

	insertNotificationContributor(t, ctx, fixture.postgres.Pool, fixture.portalA, fixture.outsiderA)
	contributorActorInput := fixture.feedbackNotification(fixture.recipientA, notificationDedupeKey("portal-contributor-actor"))
	contributorActorInput.ActorID = fixture.outsiderA
	contributorActorNotification, inserted, err := fixture.repo.Create(ctx, contributorActorInput)
	if err != nil || !inserted {
		t.Fatalf("authorized feedback contributor actor create = %t, %v", inserted, err)
	}
	mustNotificationExec(t, ctx, fixture.postgres.Pool, `DELETE FROM notifications WHERE notification_id = $1`, contributorActorNotification.ID)
	mustNotificationExec(t, ctx, fixture.postgres.Pool, `
		UPDATE feedback_contributors SET blocked_at = now()
		WHERE portal_id = $1 AND user_id = $2
	`, fixture.portalA, fixture.outsiderA)
	blockedActorInput := contributorActorInput
	blockedActorInput.DedupeKey = notificationDedupeKey("portal-blocked-actor")
	if _, _, err := fixture.repo.Create(ctx, blockedActorInput); !errors.Is(err, notificationsdomain.ErrForbidden) {
		t.Fatalf("blocked feedback contributor actor error = %v, want ErrForbidden", err)
	}

	mustNotificationExec(t, ctx, fixture.postgres.Pool, `DELETE FROM team_members WHERE team_id = $1 AND user_id = $2`, fixture.teamA, fixture.revocableA)
	revokedStaffInput := fixture.feedbackNotification(fixture.recipientA, notificationDedupeKey("portal-revoked-staff-actor"))
	revokedStaffInput.ActorID = fixture.revocableA
	if _, _, err := fixture.repo.Create(ctx, revokedStaffInput); !errors.Is(err, notificationsdomain.ErrForbidden) {
		t.Fatalf("team-revoked feedback staff actor error = %v, want ErrForbidden", err)
	}
	insertNotificationTeamMember(t, ctx, fixture.postgres.Pool, fixture.teamA, fixture.revocableA)

	created, inserted, err := fixture.repo.Create(ctx, fixture.feedbackNotification(fixture.recipientA, notificationDedupeKey("portal")))
	if err != nil || !inserted {
		t.Fatalf("create portal feedback notification = %#v/%t, %v", created, inserted, err)
	}
	access := notificationsdomain.PortalAccess{ActorID: fixture.recipientA, PortalSlug: fixture.slugA}
	items, err := fixture.repo.ListPortalFeedback(ctx, notificationsdomain.PortalListQuery{Access: access, Limit: 10})
	if err != nil || len(items) != 1 || items[0].Notification.ID != created.ID || items[0].FeedbackSlug != "feedback-a" {
		t.Fatalf("portal feedback inbox = %#v, %v", items, err)
	}
	count, err := fixture.repo.CountUnreadPortalFeedback(ctx, access)
	if err != nil || count != 1 {
		t.Fatalf("portal unread count = %d, %v; want 1", count, err)
	}
	if err := fixture.repo.MarkPortalFeedbackRead(ctx, notificationsdomain.PortalNotificationMutation{
		Access: access, NotificationID: created.ID, At: time.Now().UTC(),
	}); err != nil {
		t.Fatalf("mark portal feedback read: %v", err)
	}
	// Restore pending state to exercise the email-delivery authorization path.
	mustNotificationExec(t, ctx, fixture.postgres.Pool, `UPDATE notifications SET read_at = NULL WHERE notification_id = $1`, created.ID)
	deliveryQuery := notificationsdomain.EmailNotificationQuery{
		Scope:          notificationsdomain.DeliveryScope{RecipientID: fixture.recipientA, WorkspaceID: fixture.workspaceA},
		NotificationID: created.ID,
	}
	if delivery, err := fixture.repo.GetEmailDelivery(ctx, deliveryQuery); err != nil || delivery == nil {
		t.Fatalf("authorized portal email delivery = %#v, %v", delivery, err)
	}

	deniedAccesses := []notificationsdomain.PortalAccess{
		{ActorID: fixture.recipientA, PortalSlug: fixture.slugB},
		{ActorID: fixture.outsiderA, PortalSlug: fixture.slugA},
		{ActorID: fixture.inactiveA, PortalSlug: fixture.slugA},
	}
	for _, denied := range deniedAccesses {
		if _, err := fixture.repo.ListPortalFeedback(ctx, notificationsdomain.PortalListQuery{Access: denied, Limit: 10}); !errors.Is(err, notificationsdomain.ErrForbidden) {
			t.Fatalf("denied portal access %#v error = %v, want ErrForbidden", denied, err)
		}
	}

	mustNotificationExec(t, ctx, fixture.postgres.Pool, `
		UPDATE feedback_contributors SET blocked_at = now()
		WHERE portal_id = $1 AND user_id = $2
	`, fixture.portalA, fixture.recipientA)
	if _, err := fixture.repo.CountUnreadPortalFeedback(ctx, access); !errors.Is(err, notificationsdomain.ErrForbidden) {
		t.Fatalf("blocked portal contributor error = %v, want ErrForbidden", err)
	}
	if delivery, err := fixture.repo.GetEmailDelivery(ctx, deliveryQuery); err != nil || delivery != nil {
		t.Fatalf("blocked contributor email delivery = %#v, %v; want nil", delivery, err)
	}
	mustNotificationExec(t, ctx, fixture.postgres.Pool, `
		UPDATE feedback_contributors SET blocked_at = NULL
		WHERE portal_id = $1 AND user_id = $2
	`, fixture.portalA, fixture.recipientA)
	mustNotificationExec(t, ctx, fixture.postgres.Pool, `UPDATE feedback_portals SET is_public = FALSE WHERE id = $1`, fixture.portalA)
	if _, err := fixture.repo.ListPortalFeedback(ctx, notificationsdomain.PortalListQuery{Access: access, Limit: 10}); !errors.Is(err, notificationsdomain.ErrForbidden) {
		t.Fatalf("private portal access error = %v, want ErrForbidden", err)
	}
	if delivery, err := fixture.repo.GetEmailDelivery(ctx, deliveryQuery); err != nil || delivery != nil {
		t.Fatalf("private portal email delivery = %#v, %v; want nil", delivery, err)
	}
	mustNotificationExec(t, ctx, fixture.postgres.Pool, `UPDATE feedback_portals SET is_public = TRUE WHERE id = $1`, fixture.portalA)
}
