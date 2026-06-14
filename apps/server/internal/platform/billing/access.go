package billing

import "fmt"

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
