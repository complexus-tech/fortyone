//go:build integration

package feedbackrepository

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	feedback "github.com/complexus-tech/projects-api/internal/modules/feedback/service"
	"github.com/complexus-tech/projects-api/internal/testkit"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestFeedbackRepositoryPostgres18TenantRollbackConcurrencyAndPlan(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 2*time.Minute)
	defer cancel()
	postgres := testkit.NewPostgres(t)
	fixture := newFeedbackIntegrationFixture(t, ctx, postgres.Pool)
	repo := New(nil, postgres.Pool)

	var version int
	if err := postgres.Pool.QueryRow(ctx, `SELECT current_setting('server_version_num')::integer`).Scan(&version); err != nil {
		t.Fatalf("read PostgreSQL version: %v", err)
	}
	if version < 180000 || version >= 190000 {
		t.Fatalf("PostgreSQL version = %d, want 18.x", version)
	}

	access := feedback.CoreAccessScope{WorkspaceID: fixture.workspaceA, ActorID: fixture.actorA, AllTeams: true}
	page, err := repo.ListItemsScoped(ctx, access, feedback.CoreListItemsInput{
		WorkspaceID: fixture.workspaceA, Status: "all", Sort: "newest", Page: 1, PageSize: 20,
	})
	if err != nil || len(page.Items) != 1 || page.Items[0].ID != fixture.itemA {
		t.Fatalf("tenant-scoped list = %#v, %v", page, err)
	}
	crossTenant, err := repo.GetItemScoped(ctx, access, fixture.itemB)
	if err == nil || crossTenant.ID != uuid.Nil {
		t.Fatalf("cross-tenant item = %#v, %v; want hidden", crossTenant, err)
	}

	_, err = repo.CreateBoard(ctx, feedback.CoreBoardInput{
		WorkspaceID: fixture.workspaceA, PortalID: fixture.portalA, TeamID: fixture.teamA,
		CreatorID: uuid.New(), Name: "Must roll back", Slug: "must-roll-back", Color: "green",
	})
	if err == nil {
		t.Fatal("create board with ineligible reviewer succeeded")
	}
	var rolledBack int
	if err := postgres.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM feedback_boards WHERE portal_id = $1 AND slug = 'must-roll-back'`, fixture.portalA).Scan(&rolledBack); err != nil || rolledBack != 0 {
		t.Fatalf("rolled-back board count = %d, %v", rolledBack, err)
	}

	const writers = 8
	var wait sync.WaitGroup
	type claimResult struct {
		claimed bool
		err     error
	}
	claims := make(chan claimResult, writers)
	claim := feedback.CoreDigestDeliveryClaim{
		WorkspaceID: fixture.workspaceA,
		RecipientID: fixture.actorA,
		LocalDate:   time.Date(2026, time.August, 28, 0, 0, 0, 0, time.UTC),
		WindowStart: time.Date(2026, time.August, 27, 0, 0, 0, 0, time.UTC),
		WindowEnd:   time.Date(2026, time.August, 28, 0, 0, 0, 0, time.UTC),
		StaleBefore: time.Date(2026, time.August, 28, 0, 0, 0, 0, time.UTC).Add(-2 * time.Hour),
	}
	for range writers {
		wait.Add(1)
		go func() {
			defer wait.Done()
			_, claimed, claimErr := repo.ClaimDigestDelivery(ctx, claim)
			claims <- claimResult{claimed: claimed, err: claimErr}
		}()
	}
	wait.Wait()
	close(claims)
	claimedCount := 0
	for result := range claims {
		if result.err != nil {
			t.Fatalf("concurrent digest claim: %v", result.err)
		}
		if result.claimed {
			claimedCount++
		}
	}
	if claimedCount != 1 {
		t.Fatalf("successful digest claims = %d, want 1", claimedCount)
	}

	var planLines []string
	tx, err := postgres.Pool.Begin(ctx)
	if err != nil {
		t.Fatalf("begin plan transaction: %v", err)
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, `SET LOCAL enable_seqscan = off`); err != nil {
		t.Fatalf("prefer index plans: %v", err)
	}
	rows, err := tx.Query(ctx, `
		EXPLAIN (COSTS OFF)
		SELECT id FROM feedback_items
		WHERE board_id = $1 AND status = 'pending' AND deleted_at IS NULL
		ORDER BY created_at DESC LIMIT 20
	`, fixture.boardA)
	if err != nil {
		t.Fatalf("explain feedback list: %v", err)
	}
	for rows.Next() {
		var line string
		if scanErr := rows.Scan(&line); scanErr != nil {
			rows.Close()
			t.Fatalf("scan plan: %v", scanErr)
		}
		planLines = append(planLines, line)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		t.Fatalf("read plan: %v", err)
	}
	if plan := strings.Join(planLines, "\n"); !strings.Contains(plan, "idx_feedback_items_board_status") {
		t.Fatalf("feedback list plan does not use board/status index:\n%s", plan)
	}
	if err := tx.Commit(ctx); err != nil {
		t.Fatalf("commit plan transaction: %v", err)
	}

	// Board deletion locks the contributor rows themselves (not a DISTINCT
	// projection, which PostgreSQL rejects with FOR UPDATE), then removes only
	// anonymous contributors that became orphaned in the same transaction.
	anonymousContributorID := uuid.New()
	feedbackIntegrationExec(t, ctx, postgres.Pool,
		`INSERT INTO feedback_contributors (id, portal_id, kind) VALUES ($1, $2, 'anonymous')`,
		anonymousContributorID, fixture.portalA,
	)
	feedbackIntegrationExec(t, ctx, postgres.Pool,
		`INSERT INTO feedback_items (id, workspace_id, portal_id, board_id, contributor_id, title, slug) VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		uuid.New(), fixture.workspaceA, fixture.portalA, fixture.boardA, anonymousContributorID,
		"Anonymous board item", "anonymous-"+uuid.NewString(),
	)
	if err := repo.DeleteBoard(ctx, fixture.workspaceB, fixture.boardA); err == nil {
		t.Fatal("cross-tenant board deletion succeeded")
	}
	var retained int
	if err := postgres.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM feedback_contributors WHERE id = $1`, anonymousContributorID).Scan(&retained); err != nil || retained != 1 {
		t.Fatalf("anonymous contributor after denied delete = %d, %v; want 1", retained, err)
	}
	if err := repo.DeleteBoard(ctx, fixture.workspaceA, fixture.boardA); err != nil {
		t.Fatalf("delete authorized board: %v", err)
	}
	if err := postgres.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM feedback_contributors WHERE id = $1`, anonymousContributorID).Scan(&retained); err != nil || retained != 0 {
		t.Fatalf("orphan anonymous contributor after delete = %d, %v; want 0", retained, err)
	}
	var otherTenantBoard int
	if err := postgres.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM feedback_boards WHERE id = $1`, fixture.boardB).Scan(&otherTenantBoard); err != nil || otherTenantBoard != 1 {
		t.Fatalf("other-tenant board after delete = %d, %v; want 1", otherTenantBoard, err)
	}
}

type feedbackIntegrationFixture struct {
	workspaceA uuid.UUID
	workspaceB uuid.UUID
	teamA      uuid.UUID
	teamB      uuid.UUID
	actorA     uuid.UUID
	actorB     uuid.UUID
	portalA    uuid.UUID
	portalB    uuid.UUID
	boardA     uuid.UUID
	boardB     uuid.UUID
	itemA      uuid.UUID
	itemB      uuid.UUID
}

func newFeedbackIntegrationFixture(t *testing.T, ctx context.Context, pool *pgxpool.Pool) feedbackIntegrationFixture {
	t.Helper()
	fixture := feedbackIntegrationFixture{
		workspaceA: uuid.New(), workspaceB: uuid.New(), teamA: uuid.New(), teamB: uuid.New(),
		actorA: uuid.New(), actorB: uuid.New(), portalA: uuid.New(), portalB: uuid.New(),
		boardA: uuid.New(), boardB: uuid.New(), itemA: uuid.New(), itemB: uuid.New(),
	}
	for label, userID := range map[string]uuid.UUID{"a": fixture.actorA, "b": fixture.actorB} {
		feedbackIntegrationExec(t, ctx, pool, `INSERT INTO users (user_id, username, email, full_name) VALUES ($1, $2, $3, $4)`,
			userID, "feedback-"+label+"-"+uuid.NewString(), fmt.Sprintf("feedback-%s-%s@example.com", label, uuid.NewString()), "Feedback actor "+label)
	}
	for _, tenant := range []struct {
		workspace, team, actor, portal, board, item uuid.UUID
		suffix                                      string
	}{
		{fixture.workspaceA, fixture.teamA, fixture.actorA, fixture.portalA, fixture.boardA, fixture.itemA, "a"},
		{fixture.workspaceB, fixture.teamB, fixture.actorB, fixture.portalB, fixture.boardB, fixture.itemB, "b"},
	} {
		feedbackIntegrationExec(t, ctx, pool, `INSERT INTO workspaces (workspace_id, name, slug, created_by) VALUES ($1, $2, $3, $4)`, tenant.workspace, "Feedback "+tenant.suffix, "feedback-"+tenant.suffix+"-"+uuid.NewString(), tenant.actor)
		feedbackIntegrationExec(t, ctx, pool, `INSERT INTO workspace_members (workspace_id, user_id, role) VALUES ($1, $2, 'member')`, tenant.workspace, tenant.actor)
		feedbackIntegrationExec(t, ctx, pool, `INSERT INTO teams (team_id, name, workspace_id, code, color) VALUES ($1, $2, $3, $4, '#000000')`, tenant.team, "Feedback team "+tenant.suffix, tenant.workspace, "F"+strings.ToUpper(tenant.suffix))
		feedbackIntegrationExec(t, ctx, pool, `INSERT INTO team_members (team_id, user_id) VALUES ($1, $2)`, tenant.team, tenant.actor)
		feedbackIntegrationExec(t, ctx, pool, `INSERT INTO feedback_portals (id, workspace_id, is_public) VALUES ($1, $2, TRUE)`, tenant.portal, tenant.workspace)
		feedbackIntegrationExec(t, ctx, pool, `INSERT INTO feedback_boards (id, workspace_id, portal_id, team_id, name, slug) VALUES ($1, $2, $3, $4, $5, $6)`, tenant.board, tenant.workspace, tenant.portal, tenant.team, "Feedback board "+tenant.suffix, "board-"+tenant.suffix)
		contributorID := uuid.New()
		feedbackIntegrationExec(t, ctx, pool, `INSERT INTO feedback_contributors (id, portal_id, user_id, kind) VALUES ($1, $2, $3, 'account')`, contributorID, tenant.portal, tenant.actor)
		feedbackIntegrationExec(t, ctx, pool, `INSERT INTO feedback_items (id, workspace_id, portal_id, board_id, contributor_id, author_id, title, slug) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`, tenant.item, tenant.workspace, tenant.portal, tenant.board, contributorID, tenant.actor, "Feedback item "+tenant.suffix, "item-"+tenant.suffix)
	}
	return fixture
}

func feedbackIntegrationExec(t *testing.T, ctx context.Context, pool *pgxpool.Pool, statement string, arguments ...any) {
	t.Helper()
	if _, err := pool.Exec(ctx, statement, arguments...); err != nil {
		t.Fatalf("execute feedback fixture SQL: %v", err)
	}
}
