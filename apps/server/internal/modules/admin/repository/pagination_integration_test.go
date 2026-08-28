//go:build integration

package adminrepository

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strconv"
	"strings"
	"testing"
	"time"

	admindomain "github.com/complexus-tech/projects-api/internal/modules/admin/domain"
	"github.com/complexus-tech/projects-api/internal/platform/pagination"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

func TestAdminStablePaginationFiltersAndIndexPlansOnPostgres18(t *testing.T) {
	fixture := newAdminIntegrationFixture(t)
	ctx := t.Context()
	createdAt := fixture.now.Add(-30 * time.Minute)
	ids := []uuid.UUID{
		uuid.MustParse("00000000-0000-0000-0000-000000000101"),
		uuid.MustParse("00000000-0000-0000-0000-000000000102"),
		uuid.MustParse("00000000-0000-0000-0000-000000000103"),
	}
	for index, id := range ids {
		insertAdminTestUser(t, fixture.postgres.Pool, id, "stable-user-"+string(rune('a'+index)), true, false, createdAt)
		insertAdminTestWorkspace(t, fixture.postgres.Pool, id, fixture.actorA, "stable-workspace-"+string(rune('a'+index)), createdAt)
		_, err := fixture.postgres.Pool.Exec(ctx, `
			INSERT INTO admin_audit_logs (
				id, actor_user_id, target_type, target_id, workspace_id,
				action, field_name, metadata, created_at
			) VALUES (
				CAST($1 AS uuid), CAST($2 AS uuid), 'workspace', CAST($1 AS uuid),
				CAST($1 AS uuid), 'workspace.deleted', 'deleted_at', '{}', CAST($3 AS timestamptz)
			)
		`, id, fixture.actorA, createdAt)
		require.NoError(t, err)
		_, err = fixture.postgres.Pool.Exec(ctx, `
			INSERT INTO admin_notes (
				id, target_type, target_id, workspace_id, body, created_by_user_id, created_at
			) VALUES (
				CAST($1 AS uuid), 'workspace', CAST($1 AS uuid), CAST($1 AS uuid),
				CAST($2 AS text), CAST($3 AS uuid), CAST($4 AS timestamptz)
			)
		`, id, "stable note "+string(rune('a'+index)), fixture.actorA, createdAt)
		require.NoError(t, err)
	}

	pageOne := pagination.OffsetParams{Page: 1, PageSize: 2}
	pageTwo := pagination.OffsetParams{Page: 2, PageSize: 2}
	usersOne, err := fixture.repository.ListUsers(ctx, admindomain.ListUsersQuery{
		ActorID: fixture.actorA, Page: pageOne, Search: "stable-user-",
	})
	require.NoError(t, err)
	usersTwo, err := fixture.repository.ListUsers(ctx, admindomain.ListUsersQuery{
		ActorID: fixture.actorA, Page: pageTwo, Search: "stable-user-",
	})
	require.NoError(t, err)
	require.Equal(t, []uuid.UUID{ids[2], ids[1]}, userIDs(usersOne.Items))
	require.Equal(t, []uuid.UUID{ids[0]}, userIDs(usersTwo.Items))
	require.Equal(t, 3, usersOne.Pagination.Total)

	workspacesOne, err := fixture.repository.ListWorkspaces(ctx, admindomain.ListWorkspacesQuery{
		ActorID: fixture.actorA, Page: pageOne, Search: "stable-workspace-", Now: fixture.now,
	})
	require.NoError(t, err)
	workspacesTwo, err := fixture.repository.ListWorkspaces(ctx, admindomain.ListWorkspacesQuery{
		ActorID: fixture.actorA, Page: pageTwo, Search: "stable-workspace-", Now: fixture.now,
	})
	require.NoError(t, err)
	require.Equal(t, []uuid.UUID{ids[2], ids[1]}, workspaceIDs(workspacesOne.Items))
	require.Equal(t, []uuid.UUID{ids[0]}, workspaceIDs(workspacesTwo.Items))

	auditsOne, err := fixture.repository.ListAuditLogs(ctx, admindomain.ListAuditLogsQuery{
		ActorID: fixture.actorA, Page: pageOne, TargetType: admindomain.TargetWorkspace,
		Action: admindomain.AuditWorkspaceDeleted,
	})
	require.NoError(t, err)
	auditsTwo, err := fixture.repository.ListAuditLogs(ctx, admindomain.ListAuditLogsQuery{
		ActorID: fixture.actorA, Page: pageTwo, TargetType: admindomain.TargetWorkspace,
		Action: admindomain.AuditWorkspaceDeleted,
	})
	require.NoError(t, err)
	require.Equal(t, []uuid.UUID{ids[2], ids[1]}, auditIDs(auditsOne.Items))
	require.Equal(t, []uuid.UUID{ids[0]}, auditIDs(auditsTwo.Items))

	notesOne, err := fixture.repository.ListAdminNotes(ctx, admindomain.ListAdminNotesQuery{
		ActorID: fixture.actorA, Page: pageOne, TargetType: admindomain.TargetWorkspace,
	})
	require.NoError(t, err)
	notesTwo, err := fixture.repository.ListAdminNotes(ctx, admindomain.ListAdminNotesQuery{
		ActorID: fixture.actorA, Page: pageTwo, TargetType: admindomain.TargetWorkspace,
	})
	require.NoError(t, err)
	require.Equal(t, []uuid.UUID{ids[2], ids[1]}, noteIDs(notesOne.Items))
	require.Equal(t, []uuid.UUID{ids[0]}, noteIDs(notesTwo.Items))

	assertAdminPaginationIndexPlans(t, fixture.postgres.Pool)
}

func assertAdminPaginationIndexPlans(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	ctx := t.Context()
	connection, err := pool.Acquire(ctx)
	require.NoError(t, err)
	defer connection.Release()
	_, err = connection.Exec(ctx, "ANALYZE users, workspaces, admin_audit_logs, admin_notes")
	require.NoError(t, err)
	_, err = connection.Exec(ctx, "SET enable_seqscan = off")
	require.NoError(t, err)

	type planCase struct {
		indexName string
		file      string
		queryName string
		args      []any
	}
	plans := []planCase{
		{
			indexName: "idx_users_admin_created_id", file: "sqlc/reads.sql.go",
			queryName: "listAdminUsers", args: []any{"", int32(0), int32(20)},
		},
		{
			indexName: "idx_workspaces_admin_created_id", file: "sqlc/reads.sql.go",
			queryName: "listAdminWorkspaces", args: []any{"", "", time.Now().UTC(), int32(0), int32(20)},
		},
		{
			indexName: "idx_admin_audit_logs_created_id", file: "sqlc/audit.sql.go",
			queryName: "listAdminAuditLogs", args: []any{
				false, uuid.Nil, "", "", "", "", false, time.Time{}, false, time.Time{}, int32(0), int32(20),
			},
		},
		{
			indexName: "idx_admin_notes_created_id", file: "sqlc/audit.sql.go",
			queryName: "listAdminNotes", args: []any{
				"", false, uuid.Nil, false, uuid.Nil, int32(0), int32(20),
			},
		},
	}
	for _, plan := range plans {
		statement := explainGeneratedQuery(t, plan.file, plan.queryName)
		rows, err := connection.Query(ctx, statement, plan.args...)
		require.NoError(t, err)
		var lines []string
		for rows.Next() {
			var line string
			require.NoError(t, rows.Scan(&line))
			lines = append(lines, line)
		}
		require.NoError(t, rows.Err())
		rows.Close()
		require.Contains(t, strings.Join(lines, "\n"), plan.indexName)

		var definition string
		require.NoError(t, connection.QueryRow(ctx, `
			SELECT indexdef FROM pg_indexes WHERE schemaname = 'public' AND indexname = CAST($1 AS text)
		`, plan.indexName).Scan(&definition))
		require.Contains(t, definition, "created_at DESC")
	}
}

func explainGeneratedQuery(t *testing.T, path, name string) string {
	t.Helper()
	parsed, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
	require.NoError(t, err)
	for _, declaration := range parsed.Decls {
		generic, ok := declaration.(*ast.GenDecl)
		if !ok || generic.Tok != token.CONST {
			continue
		}
		for _, specification := range generic.Specs {
			valueSpec, ok := specification.(*ast.ValueSpec)
			if !ok {
				continue
			}
			for index, identifier := range valueSpec.Names {
				if identifier.Name != name || index >= len(valueSpec.Values) {
					continue
				}
				literal, ok := valueSpec.Values[index].(*ast.BasicLit)
				require.True(t, ok, "generated query %s must be a string literal", name)
				query, err := strconv.Unquote(literal.Value)
				require.NoError(t, err)
				firstLine := strings.IndexByte(query, '\n')
				require.NotEqual(t, -1, firstLine)
				return "EXPLAIN (COSTS OFF) " + query[firstLine+1:]
			}
		}
	}
	t.Fatalf("generated query constant %s was not found in %s", name, path)
	return ""
}

func userIDs(items []admindomain.UserSummary) []uuid.UUID {
	result := make([]uuid.UUID, 0, len(items))
	for _, item := range items {
		result = append(result, item.ID)
	}
	return result
}

func workspaceIDs(items []admindomain.WorkspaceSummary) []uuid.UUID {
	result := make([]uuid.UUID, 0, len(items))
	for _, item := range items {
		result = append(result, item.ID)
	}
	return result
}

func auditIDs(items []admindomain.AuditLog) []uuid.UUID {
	result := make([]uuid.UUID, 0, len(items))
	for _, item := range items {
		result = append(result, item.ID)
	}
	return result
}

func noteIDs(items []admindomain.AdminNote) []uuid.UUID {
	result := make([]uuid.UUID, 0, len(items))
	for _, item := range items {
		result = append(result, item.ID)
	}
	return result
}
