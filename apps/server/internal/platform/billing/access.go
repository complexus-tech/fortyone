package billing

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// WorkspaceMayaAccessSQL returns a SQL predicate for workspace-level Maya access.
// The alias must be a trusted workspace table alias from the static query.
func WorkspaceMayaAccessSQL(workspaceAlias string) string {
	return fmt.Sprintf(`
		(
			%s.trial_ends_on > NOW()
			OR EXISTS (
				SELECT 1
				FROM workspace_subscriptions ws
				WHERE ws.workspace_id = %[1]s.workspace_id
					AND ws.subscription_tier <> 'free'
					AND ws.subscription_status IN ('active', 'trialing', 'past_due')
				LIMIT 1
			)
		)
	`, workspaceAlias)
}

func WorkspaceCanUseMaya(ctx context.Context, db *sqlx.DB, workspaceID uuid.UUID) (bool, error) {
	query := `
		SELECT EXISTS (
			SELECT 1
			FROM workspaces w
			WHERE w.workspace_id = :workspace_id
				AND w.deleted_at IS NULL
				AND ` + WorkspaceMayaAccessSQL("w") + `
		)
	`
	params := map[string]any{
		"workspace_id": workspaceID,
	}

	stmt, err := db.PrepareNamedContext(ctx, query)
	if err != nil {
		return false, fmt.Errorf("prepare Maya access query: %w", err)
	}
	defer stmt.Close()

	var hasAccess bool
	if err := stmt.GetContext(ctx, &hasAccess, params); err != nil {
		return false, fmt.Errorf("execute Maya access query: %w", err)
	}
	return hasAccess, nil
}
