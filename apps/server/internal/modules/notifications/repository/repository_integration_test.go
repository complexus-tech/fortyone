//go:build integration

package notificationsrepository

import (
	"context"
	"errors"
	"sort"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	notificationsdomain "github.com/complexus-tech/projects-api/internal/modules/notifications/domain"
	platformpatch "github.com/complexus-tech/projects-api/internal/platform/patch"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func TestNotificationRepositoryPostgres18SecurityConcurrencyAndPlans(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(t.Context(), 2*time.Minute)
	defer cancel()
	fixture := newNotificationIntegrationFixture(t, ctx)
	assertNotificationPostgres18(t, ctx, fixture)

	testNotificationCreateIdempotencyAndSecurity(t, ctx, fixture)
	testNotificationInboxScopePaginationAndConcurrency(t, ctx, fixture)
	testNotificationPreferencesConcurrencyAndRevocation(t, ctx, fixture)
	testNotificationPortalScope(t, ctx, fixture)
	testNotificationDeliveryScopeAndAudience(t, ctx, fixture)
	testNotificationCreateRollback(t, ctx, fixture)
	assertNotificationQueryPlans(t, ctx, fixture)
}

func testNotificationCreateIdempotencyAndSecurity(
	t *testing.T,
	ctx context.Context,
	fixture notificationIntegrationFixture,
) {
	t.Helper()

	dedupeKey := notificationDedupeKey("exact-replay")
	input := fixture.storyNotification(fixture.recipientA, dedupeKey)
	created, inserted, err := fixture.repo.Create(ctx, input)
	if err != nil || !inserted {
		t.Fatalf("create notification = %#v/%t, %v", created, inserted, err)
	}
	replayed, inserted, err := fixture.repo.Create(ctx, input)
	if err != nil || inserted || replayed.ID != created.ID || replayed.CreatedAt != created.CreatedAt {
		t.Fatalf("exact replay = %#v/%t, %v; want original %s", replayed, inserted, err, created.ID)
	}

	conflicting := input
	conflicting.Title = "Conflicting replay"
	if _, _, err := fixture.repo.Create(ctx, conflicting); !errors.Is(err, notificationsdomain.ErrConflict) {
		t.Fatalf("conflicting replay error = %v, want ErrConflict", err)
	}
	crossTenantCollision := input
	crossTenantCollision.RecipientID = fixture.recipientB
	crossTenantCollision.WorkspaceID = fixture.workspaceB
	crossTenantCollision.EntityID = fixture.storyB
	crossTenantCollision.ActorID = fixture.actorB
	if _, _, err := fixture.repo.Create(ctx, crossTenantCollision); !errors.Is(err, notificationsdomain.ErrForbidden) {
		t.Fatalf("cross-tenant dedupe collision error = %v, want ErrForbidden", err)
	}

	const writers = 8
	concurrentInput := fixture.storyNotification(fixture.recipientA, notificationDedupeKey("concurrent-replay"))
	type outcome struct {
		id       uuid.UUID
		inserted bool
		err      error
	}
	outcomes := make(chan outcome, writers)
	var wait sync.WaitGroup
	for range writers {
		wait.Add(1)
		go func() {
			defer wait.Done()
			notification, wasInserted, createErr := fixture.repo.Create(ctx, concurrentInput)
			outcomes <- outcome{id: notification.ID, inserted: wasInserted, err: createErr}
		}()
	}
	wait.Wait()
	close(outcomes)
	insertions := 0
	var notificationID uuid.UUID
	for result := range outcomes {
		if result.err != nil {
			t.Fatalf("concurrent exact replay: %v", result.err)
		}
		if notificationID == uuid.Nil {
			notificationID = result.id
		}
		if result.id != notificationID {
			t.Fatalf("concurrent replay IDs differ: got %s, want %s", result.id, notificationID)
		}
		if result.inserted {
			insertions++
		}
	}
	if insertions != 1 {
		t.Fatalf("concurrent replay insertions = %d, want 1", insertions)
	}

	guest, inserted, err := fixture.repo.Create(ctx, fixture.storyNotification(fixture.guestA, notificationDedupeKey("guest")))
	if err != nil || !inserted || guest.RecipientID != fixture.guestA {
		t.Fatalf("authorized guest create = %#v/%t, %v", guest, inserted, err)
	}
	systemInput := withNotificationActor(input, fixture.system, notificationDedupeKey("system-actor"))
	if _, inserted, err := fixture.repo.Create(ctx, systemInput); err != nil || !inserted {
		t.Fatalf("authorized system actor create = %t, %v", inserted, err)
	}

	denied := []struct {
		name  string
		input notificationsdomain.NewNotification
	}{
		{name: "inactive recipient", input: fixture.storyNotification(fixture.inactiveA, notificationDedupeKey("inactive-recipient"))},
		{name: "recipient without workspace membership", input: fixture.storyNotification(fixture.outsiderA, notificationDedupeKey("outsider-recipient"))},
		{name: "actor from another tenant", input: withNotificationActor(input, fixture.actorB, notificationDedupeKey("foreign-actor"))},
		{name: "inactive actor", input: withNotificationActor(input, fixture.inactiveA, notificationDedupeKey("inactive-actor"))},
		{name: "resource from another tenant", input: withNotificationEntity(input, fixture.storyB, notificationDedupeKey("foreign-story"))},
	}
	for _, test := range denied {
		if _, _, err := fixture.repo.Create(ctx, test.input); !errors.Is(err, notificationsdomain.ErrForbidden) {
			t.Errorf("%s error = %v, want ErrForbidden", test.name, err)
		}
	}
	mustNotificationExec(t, ctx, fixture.postgres.Pool, `DELETE FROM team_members WHERE team_id = $1 AND user_id = $2`, fixture.teamA, fixture.revocableA)
	teamRevokedActor := withNotificationActor(input, fixture.revocableA, notificationDedupeKey("team-revoked-actor"))
	if _, _, err := fixture.repo.Create(ctx, teamRevokedActor); !errors.Is(err, notificationsdomain.ErrForbidden) {
		t.Errorf("team-revoked actor error = %v, want ErrForbidden", err)
	}
	insertNotificationTeamMember(t, ctx, fixture.postgres.Pool, fixture.teamA, fixture.revocableA)
}

func testNotificationInboxScopePaginationAndConcurrency(
	t *testing.T,
	ctx context.Context,
	fixture notificationIntegrationFixture,
) {
	t.Helper()

	for index := range 5 {
		input := fixture.storyNotification(fixture.recipientA, notificationDedupeKey("pagination"))
		input.Title = "Stable notification " + strconv.Itoa(index)
		if _, inserted, err := fixture.repo.Create(ctx, input); err != nil || !inserted {
			t.Fatalf("create pagination notification %d = %t, %v", index, inserted, err)
		}
	}
	fixedCreatedAt := time.Date(2026, time.August, 28, 12, 0, 0, 0, time.UTC)
	mustNotificationExec(t, ctx, fixture.postgres.Pool, `
		UPDATE notifications
		SET created_at = $1
		WHERE recipient_id = $2 AND workspace_id = $3 AND entity_type = 'story'
	`, fixedCreatedAt, fixture.recipientA, fixture.workspaceA)

	expected := notificationIDs(t, ctx, fixture, `
		SELECT notification_id
		FROM notifications
		WHERE recipient_id = $1 AND workspace_id = $2 AND entity_type = 'story'
		ORDER BY created_at DESC NULLS LAST, notification_id DESC
	`, fixture.recipientA, fixture.workspaceA)
	if len(expected) < 7 {
		t.Fatalf("pagination fixture rows = %d, want at least 7", len(expected))
	}
	first, err := fixture.repo.List(ctx, notificationsdomain.ListQuery{
		Access: fixture.accessA(fixture.recipientA), Limit: 3, Offset: 0,
	})
	if err != nil {
		t.Fatalf("list first stable page: %v", err)
	}
	repeated, err := fixture.repo.List(ctx, notificationsdomain.ListQuery{
		Access: fixture.accessA(fixture.recipientA), Limit: 3, Offset: 0,
	})
	if err != nil {
		t.Fatalf("repeat first stable page: %v", err)
	}
	second, err := fixture.repo.List(ctx, notificationsdomain.ListQuery{
		Access: fixture.accessA(fixture.recipientA), Limit: 3, Offset: 3,
	})
	if err != nil {
		t.Fatalf("list second stable page: %v", err)
	}
	if got := domainNotificationIDs(first); !equalNotificationIDs(got, expected[:3]) {
		t.Fatalf("first page = %v, want %v", got, expected[:3])
	}
	if got := domainNotificationIDs(repeated); !equalNotificationIDs(got, expected[:3]) {
		t.Fatalf("repeated first page = %v, want %v", got, expected[:3])
	}
	if got := domainNotificationIDs(second); !equalNotificationIDs(got, expected[3:6]) {
		t.Fatalf("second page = %v, want %v", got, expected[3:6])
	}

	unread, err := fixture.repo.CountUnread(ctx, fixture.accessA(fixture.recipientA))
	if err != nil || unread != len(expected) {
		t.Fatalf("unread count = %d, %v; want %d", unread, err, len(expected))
	}

	notificationID := expected[0]
	runConcurrentNotificationMutations(t, ctx, fixture, notificationID, notificationsdomain.NotificationMutationRead)
	var readAt *time.Time
	if err := fixture.postgres.Pool.QueryRow(ctx, `SELECT read_at FROM notifications WHERE notification_id = $1`, notificationID).Scan(&readAt); err != nil || readAt == nil {
		t.Fatalf("concurrent mark-read persisted read_at = %v, %v", readAt, err)
	}
	runConcurrentNotificationMutations(t, ctx, fixture, notificationID, notificationsdomain.NotificationMutationUnread)
	if err := fixture.postgres.Pool.QueryRow(ctx, `SELECT read_at FROM notifications WHERE notification_id = $1`, notificationID).Scan(&readAt); err != nil || readAt != nil {
		t.Fatalf("concurrent mark-unread persisted read_at = %v, %v", readAt, err)
	}

	if err := fixture.repo.Mutate(ctx, notificationsdomain.NotificationMutation{
		Access: fixture.accessA(fixture.actorA), NotificationID: notificationID,
		Kind: notificationsdomain.NotificationMutationRead, At: time.Now().UTC(),
	}); !errors.Is(err, notificationsdomain.ErrNotFound) {
		t.Fatalf("same-workspace non-recipient mutation error = %v, want ErrNotFound", err)
	}
	if err := fixture.repo.Mutate(ctx, notificationsdomain.NotificationMutation{
		Access:         notificationsdomain.WorkspaceAccess{ActorID: fixture.actorB, WorkspaceID: fixture.workspaceB},
		NotificationID: notificationID, Kind: notificationsdomain.NotificationMutationDelete, At: time.Now().UTC(),
	}); !errors.Is(err, notificationsdomain.ErrNotFound) {
		t.Fatalf("cross-tenant mutation error = %v, want ErrNotFound", err)
	}

	mustNotificationExec(t, ctx, fixture.postgres.Pool, `DELETE FROM team_members WHERE team_id = $1 AND user_id = $2`, fixture.teamA, fixture.recipientA)
	items, err := fixture.repo.List(ctx, notificationsdomain.ListQuery{Access: fixture.accessA(fixture.recipientA), Limit: 10})
	if err != nil || len(items) != 0 {
		t.Fatalf("team-revoked inbox = %#v, %v; want empty", items, err)
	}
	if err := fixture.repo.Mutate(ctx, notificationsdomain.NotificationMutation{
		Access: fixture.accessA(fixture.recipientA), NotificationID: notificationID,
		Kind: notificationsdomain.NotificationMutationRead, At: time.Now().UTC(),
	}); !errors.Is(err, notificationsdomain.ErrNotFound) {
		t.Fatalf("team-revoked mutation error = %v, want ErrNotFound", err)
	}
	insertNotificationTeamMember(t, ctx, fixture.postgres.Pool, fixture.teamA, fixture.recipientA)

	mustNotificationExec(t, ctx, fixture.postgres.Pool, `DELETE FROM team_members WHERE team_id = $1 AND user_id = $2`, fixture.teamA, fixture.guestA)
	guestItems, err := fixture.repo.List(ctx, notificationsdomain.ListQuery{Access: fixture.accessA(fixture.guestA), Limit: 10})
	if err != nil || len(guestItems) != 0 {
		t.Fatalf("guest without current team scope = %#v, %v; want empty", guestItems, err)
	}
	insertNotificationTeamMember(t, ctx, fixture.postgres.Pool, fixture.teamA, fixture.guestA)
}

func testNotificationPreferencesConcurrencyAndRevocation(
	t *testing.T,
	ctx context.Context,
	fixture notificationIntegrationFixture,
) {
	t.Helper()

	preferences, err := fixture.repo.GetPreferences(ctx, fixture.accessA(fixture.recipientA))
	if err != nil {
		t.Fatalf("get default preferences: %v", err)
	}
	if !preferences.Preferences[notificationsdomain.PreferenceTypeMention].Email ||
		!preferences.Preferences[notificationsdomain.PreferenceTypeMention].InApp ||
		preferences.Preferences[notificationsdomain.PreferenceTypeStrategyUpdate].InApp {
		t.Fatalf("default preference set = %#v", preferences.Preferences)
	}

	now := time.Date(2026, time.August, 28, 13, 0, 0, 0, time.UTC)
	commands := []notificationsdomain.UpdatePreference{
		{
			Access: fixture.accessA(fixture.recipientA), Type: notificationsdomain.PreferenceTypeMention,
			Patch: notificationsdomain.ChannelPatch{Email: platformpatch.Set(false)}, At: now,
		},
		{
			Access: fixture.accessA(fixture.recipientA), Type: notificationsdomain.PreferenceTypeMention,
			Patch: notificationsdomain.ChannelPatch{InApp: platformpatch.Set(false)}, At: now.Add(time.Second),
		},
	}
	errorsChannel := make(chan error, len(commands))
	var wait sync.WaitGroup
	for _, command := range commands {
		command := command
		wait.Add(1)
		go func() {
			defer wait.Done()
			_, updateErr := fixture.repo.UpdatePreference(ctx, command)
			errorsChannel <- updateErr
		}()
	}
	wait.Wait()
	close(errorsChannel)
	for updateErr := range errorsChannel {
		if updateErr != nil {
			t.Fatalf("concurrent preference update: %v", updateErr)
		}
	}
	preferences, err = fixture.repo.GetPreferences(ctx, fixture.accessA(fixture.recipientA))
	if err != nil {
		t.Fatalf("get merged preferences: %v", err)
	}
	if mention := preferences.Preferences[notificationsdomain.PreferenceTypeMention]; mention.Email || mention.InApp {
		t.Fatalf("concurrent channel patches lost an update: %#v", mention)
	}

	if _, err := fixture.repo.UpdatePreference(ctx, notificationsdomain.UpdatePreference{
		Access: notificationsdomain.WorkspaceAccess{ActorID: fixture.actorA, WorkspaceID: fixture.workspaceB},
		Type:   notificationsdomain.PreferenceTypeMention,
		Patch:  notificationsdomain.ChannelPatch{Email: platformpatch.Set(false)}, At: now,
	}); !errors.Is(err, notificationsdomain.ErrForbidden) {
		t.Fatalf("cross-tenant preference update error = %v, want ErrForbidden", err)
	}

	mustNotificationExec(t, ctx, fixture.postgres.Pool, `DELETE FROM workspace_members WHERE workspace_id = $1 AND user_id = $2`, fixture.workspaceA, fixture.revocableA)
	if _, err := fixture.repo.GetPreferences(ctx, fixture.accessA(fixture.revocableA)); !errors.Is(err, notificationsdomain.ErrForbidden) {
		t.Fatalf("revoked workspace preference read error = %v, want ErrForbidden", err)
	}
	mustNotificationExec(t, ctx, fixture.postgres.Pool, `UPDATE users SET is_active = FALSE WHERE user_id = $1`, fixture.recipientA)
	if _, err := fixture.repo.GetPreferences(ctx, fixture.accessA(fixture.recipientA)); !errors.Is(err, notificationsdomain.ErrForbidden) {
		t.Fatalf("inactive preference read error = %v, want ErrForbidden", err)
	}
	mustNotificationExec(t, ctx, fixture.postgres.Pool, `UPDATE users SET is_active = TRUE WHERE user_id = $1`, fixture.recipientA)
}

func testNotificationDeliveryScopeAndAudience(
	t *testing.T,
	ctx context.Context,
	fixture notificationIntegrationFixture,
) {
	t.Helper()

	input := fixture.storyNotification(fixture.recipientA, notificationDedupeKey("delivery"))
	created, inserted, err := fixture.repo.Create(ctx, input)
	if err != nil || !inserted {
		t.Fatalf("create delivery notification = %#v/%t, %v", created, inserted, err)
	}
	delivery, err := fixture.repo.GetEmailDelivery(ctx, notificationsdomain.EmailNotificationQuery{
		Scope:          notificationsdomain.DeliveryScope{RecipientID: fixture.recipientA, WorkspaceID: fixture.workspaceA},
		NotificationID: created.ID,
	})
	if err != nil || delivery == nil || delivery.NotificationID != created.ID || !delivery.EmailEnabled {
		t.Fatalf("authorized email delivery = %#v, %v", delivery, err)
	}
	wrongScope, err := fixture.repo.GetEmailDelivery(ctx, notificationsdomain.EmailNotificationQuery{
		Scope:          notificationsdomain.DeliveryScope{RecipientID: fixture.recipientB, WorkspaceID: fixture.workspaceB},
		NotificationID: created.ID,
	})
	if err != nil || wrongScope != nil {
		t.Fatalf("cross-tenant email delivery = %#v, %v; want nil", wrongScope, err)
	}

	teamIDs, err := fixture.repo.ListDeliveryTeamIDs(ctx, notificationsdomain.DeliveryScope{
		RecipientID: fixture.recipientA, WorkspaceID: fixture.workspaceA,
	})
	if err != nil || len(teamIDs) != 1 || teamIDs[0] != fixture.teamA {
		t.Fatalf("delivery team scope = %v, %v", teamIDs, err)
	}

	objectiveID, keyResultID := uuid.New(), uuid.New()
	mustNotificationExec(t, ctx, fixture.postgres.Pool, `
		INSERT INTO objectives (
			objective_id, name, lead_user_id, team_id, workspace_id, created_by, sequence_id, color
		) VALUES ($1, $2, $3, $4, $5, $6, 1, '#686DE0')
	`, objectiveID, "Notification objective", fixture.recipientA, fixture.teamA, fixture.workspaceA, fixture.actorA)
	mustNotificationExec(t, ctx, fixture.postgres.Pool, `
		INSERT INTO key_results (
			id, objective_id, name, measurement_type, team_id, sequence_id,
			lead, start_date, end_date, created_by
		) VALUES ($1, $2, $3, 'percentage', $4, 1, $5, CURRENT_DATE, CURRENT_DATE + 30, $6)
	`, keyResultID, objectiveID, "Notification key result", fixture.teamA, fixture.recipientA, fixture.actorA)
	mustNotificationExec(t, ctx, fixture.postgres.Pool, `
		INSERT INTO key_result_contributors (key_result_id, user_id) VALUES ($1, $2)
	`, keyResultID, fixture.guestA)
	audience, err := fixture.repo.ListKeyResultAudience(ctx, notificationsdomain.KeyResultAudienceQuery{
		ActorID: fixture.actorA, WorkspaceID: fixture.workspaceA, KeyResultID: keyResultID,
	})
	if err != nil {
		t.Fatalf("list key-result notification audience: %v", err)
	}
	gotRecipients := make([]uuid.UUID, 0, len(audience))
	for _, member := range audience {
		gotRecipients = append(gotRecipients, member.RecipientID)
		if member.KeyResultID != keyResultID || member.ObjectiveID != objectiveID ||
			member.KeyResultName != "Notification key result" || member.ObjectiveName != "Notification objective" {
			t.Fatalf("key-result audience context = %#v", member)
		}
	}
	sort.Slice(gotRecipients, func(i, j int) bool { return gotRecipients[i].String() < gotRecipients[j].String() })
	wantRecipients := []uuid.UUID{fixture.recipientA, fixture.guestA}
	sort.Slice(wantRecipients, func(i, j int) bool { return wantRecipients[i].String() < wantRecipients[j].String() })
	if !equalNotificationIDs(gotRecipients, wantRecipients) {
		t.Fatalf("key-result audience recipients = %v, want %v", gotRecipients, wantRecipients)
	}
	foreignAudience, err := fixture.repo.ListKeyResultAudience(ctx, notificationsdomain.KeyResultAudienceQuery{
		ActorID: fixture.actorB, WorkspaceID: fixture.workspaceA, KeyResultID: keyResultID,
	})
	if err != nil || len(foreignAudience) != 0 {
		t.Fatalf("cross-tenant key-result audience = %#v, %v; want empty", foreignAudience, err)
	}
	mustNotificationExec(t, ctx, fixture.postgres.Pool, `DELETE FROM team_members WHERE team_id = $1 AND user_id = $2`, fixture.teamA, fixture.guestA)
	audience, err = fixture.repo.ListKeyResultAudience(ctx, notificationsdomain.KeyResultAudienceQuery{
		ActorID: fixture.actorA, WorkspaceID: fixture.workspaceA, KeyResultID: keyResultID,
	})
	if err != nil || len(audience) != 1 || audience[0].RecipientID != fixture.recipientA {
		t.Fatalf("team-revoked key-result audience = %#v, %v", audience, err)
	}
	insertNotificationTeamMember(t, ctx, fixture.postgres.Pool, fixture.teamA, fixture.guestA)

	if err := fixture.repo.MarkEmailSent(ctx, notificationsdomain.MarkEmailSent{
		Scope:           notificationsdomain.DeliveryScope{RecipientID: fixture.recipientA, WorkspaceID: fixture.workspaceA},
		NotificationIDs: []uuid.UUID{created.ID}, At: time.Now().UTC(),
	}); err != nil {
		t.Fatalf("mark email sent: %v", err)
	}
	delivery, err = fixture.repo.GetEmailDelivery(ctx, notificationsdomain.EmailNotificationQuery{
		Scope:          notificationsdomain.DeliveryScope{RecipientID: fixture.recipientA, WorkspaceID: fixture.workspaceA},
		NotificationID: created.ID,
	})
	if err != nil || delivery != nil {
		t.Fatalf("sent notification delivery = %#v, %v; want nil", delivery, err)
	}
}

func testNotificationCreateRollback(t *testing.T, ctx context.Context, fixture notificationIntegrationFixture) {
	t.Helper()

	mustNotificationExec(t, ctx, fixture.postgres.Pool, `
		CREATE FUNCTION reject_notification_insert() RETURNS trigger AS $$
		BEGIN
			RAISE EXCEPTION 'forced notification insert failure';
		END;
		$$ LANGUAGE plpgsql
	`)
	mustNotificationExec(t, ctx, fixture.postgres.Pool, `
		CREATE TRIGGER reject_notification_insert
		BEFORE INSERT ON notifications
		FOR EACH ROW EXECUTE FUNCTION reject_notification_insert()
	`)
	dedupeKey := notificationDedupeKey("rollback")
	if _, _, err := fixture.repo.Create(ctx, fixture.storyNotification(fixture.recipientA, dedupeKey)); err == nil {
		t.Fatal("forced notification insert error = nil")
	}
	mustNotificationExec(t, ctx, fixture.postgres.Pool, `DROP TRIGGER reject_notification_insert ON notifications`)
	mustNotificationExec(t, ctx, fixture.postgres.Pool, `DROP FUNCTION reject_notification_insert()`)
	var count int
	if err := fixture.postgres.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM notifications WHERE dedupe_key = $1`, dedupeKey).Scan(&count); err != nil || count != 0 {
		t.Fatalf("rolled-back notification count = %d, %v; want 0", count, err)
	}
}

func runConcurrentNotificationMutations(
	t *testing.T,
	ctx context.Context,
	fixture notificationIntegrationFixture,
	notificationID uuid.UUID,
	kind notificationsdomain.NotificationMutationKind,
) {
	t.Helper()
	const writers = 8
	errorsChannel := make(chan error, writers)
	var wait sync.WaitGroup
	for index := range writers {
		wait.Add(1)
		go func() {
			defer wait.Done()
			errorsChannel <- fixture.repo.Mutate(ctx, notificationsdomain.NotificationMutation{
				Access: fixture.accessA(fixture.recipientA), NotificationID: notificationID,
				Kind: kind, At: time.Date(2026, time.August, 28, 12, index, 0, 0, time.UTC),
			})
		}()
	}
	wait.Wait()
	close(errorsChannel)
	for mutationErr := range errorsChannel {
		if mutationErr != nil {
			t.Fatalf("concurrent %s mutation: %v", kind, mutationErr)
		}
	}
}

func assertNotificationPostgres18(t *testing.T, ctx context.Context, fixture notificationIntegrationFixture) {
	t.Helper()
	var raw string
	if err := fixture.postgres.Pool.QueryRow(ctx, "SHOW server_version_num").Scan(&raw); err != nil {
		t.Fatalf("read PostgreSQL version: %v", err)
	}
	version, err := strconv.Atoi(raw)
	if err != nil || version < 180000 || version >= 190000 {
		t.Fatalf("PostgreSQL version = %q, want 18.x", raw)
	}
}

func assertNotificationQueryPlans(t *testing.T, ctx context.Context, fixture notificationIntegrationFixture) {
	t.Helper()
	connection, err := fixture.postgres.Pool.Acquire(ctx)
	if err != nil {
		t.Fatalf("acquire notification query-plan connection: %v", err)
	}
	defer connection.Release()
	tx, err := connection.BeginTx(ctx, pgx.TxOptions{AccessMode: pgx.ReadOnly})
	if err != nil {
		t.Fatalf("begin notification query-plan transaction: %v", err)
	}
	defer func() { _ = tx.Rollback(context.Background()) }()
	if _, err := tx.Exec(ctx, "SET LOCAL enable_seqscan = off"); err != nil {
		t.Fatalf("disable sequential scans for notification plans: %v", err)
	}
	assertNotificationPlanUsesIndex(t, ctx, tx, "idx_notifications_in_app_recipient_workspace_created", `
		EXPLAIN (COSTS OFF)
		SELECT notification_id
		FROM notifications
		WHERE recipient_id = $1
		  AND workspace_id = $2
		  AND in_app_enabled = TRUE
		ORDER BY created_at DESC NULLS LAST, notification_id DESC
		LIMIT 25
	`, fixture.recipientA, fixture.workspaceA)
	assertNotificationPlanUsesIndex(t, ctx, tx, "idx_notifications_dedupe_key", `
		EXPLAIN (COSTS OFF)
		SELECT notification_id FROM notifications WHERE dedupe_key = $1
	`, notificationDedupeKey("missing"))
	assertNotificationPlanUsesIndex(t, ctx, tx, "notification_preferences_user_id_workspace_id_key", `
		EXPLAIN (COSTS OFF)
		SELECT preference_id
		FROM notification_preferences
		WHERE user_id = $1 AND workspace_id = $2
	`, fixture.recipientA, fixture.workspaceA)
}

func assertNotificationPlanUsesIndex(
	t *testing.T,
	ctx context.Context,
	tx pgx.Tx,
	indexName, query string,
	arguments ...any,
) {
	t.Helper()
	rows, err := tx.Query(ctx, query, arguments...)
	if err != nil {
		t.Fatalf("explain notification query for %s: %v", indexName, err)
	}
	defer rows.Close()
	var plan strings.Builder
	for rows.Next() {
		var line string
		if err := rows.Scan(&line); err != nil {
			t.Fatalf("scan notification query plan for %s: %v", indexName, err)
		}
		plan.WriteString(line)
		plan.WriteByte('\n')
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("read notification query plan for %s: %v", indexName, err)
	}
	if !strings.Contains(plan.String(), indexName) {
		t.Fatalf("notification query plan did not use %s:\n%s", indexName, plan.String())
	}
}

func notificationIDs(
	t *testing.T,
	ctx context.Context,
	fixture notificationIntegrationFixture,
	query string,
	arguments ...any,
) []uuid.UUID {
	t.Helper()
	rows, err := fixture.postgres.Pool.Query(ctx, query, arguments...)
	if err != nil {
		t.Fatalf("list notification IDs: %v", err)
	}
	defer rows.Close()
	result := make([]uuid.UUID, 0)
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			t.Fatalf("scan notification ID: %v", err)
		}
		result = append(result, id)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("read notification IDs: %v", err)
	}
	return result
}

func domainNotificationIDs(notifications []notificationsdomain.Notification) []uuid.UUID {
	result := make([]uuid.UUID, 0, len(notifications))
	for _, notification := range notifications {
		result = append(result, notification.ID)
	}
	return result
}

func equalNotificationIDs(left, right []uuid.UUID) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func withNotificationActor(
	notification notificationsdomain.NewNotification,
	actorID uuid.UUID,
	dedupeKey string,
) notificationsdomain.NewNotification {
	notification.ActorID = actorID
	notification.DedupeKey = dedupeKey
	return notification
}

func withNotificationEntity(
	notification notificationsdomain.NewNotification,
	entityID uuid.UUID,
	dedupeKey string,
) notificationsdomain.NewNotification {
	notification.EntityID = entityID
	notification.DedupeKey = dedupeKey
	return notification
}
