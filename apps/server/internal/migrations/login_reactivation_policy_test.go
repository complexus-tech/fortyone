package migrations

import (
	"strings"
	"testing"
)

func TestLoginReactivationPolicyMigrationFailsClosedForLegacyInactiveAccounts(t *testing.T) {
	t.Parallel()

	forward, err := FS.ReadFile("000174_login_reactivation_policy.up.sql")
	if err != nil {
		t.Fatalf("read forward migration: %v", err)
	}
	source := strings.Join(strings.Fields(strings.ToLower(string(forward))), " ")
	for _, contract := range []string{
		"add column login_reactivation_policy text not null default 'verified_sign_in'",
		"audit.action in ('user.activated', 'user.deactivated')",
		"order by audit.created_at desc, audit.id desc limit 1",
		"then 'admin_only' else 'legacy_admin_review'",
		"where account.is_active = false",
		"users_login_reactivation_policy_check",
		"'verified_sign_in', 'admin_only', 'legacy_admin_review'",
		"only verified_sign_in may reactivate through verified authentication",
	} {
		if !strings.Contains(source, contract) {
			t.Errorf("forward migration is missing %q", contract)
		}
	}
}

func TestLoginReactivationPolicyMigrationIsReversible(t *testing.T) {
	t.Parallel()

	rollback, err := FS.ReadFile("000174_login_reactivation_policy.down.sql")
	if err != nil {
		t.Fatalf("read rollback migration: %v", err)
	}
	source := strings.Join(strings.Fields(strings.ToLower(string(rollback))), " ")
	for _, contract := range []string{
		"drop constraint users_login_reactivation_policy_check",
		"drop column login_reactivation_policy",
	} {
		if !strings.Contains(source, contract) {
			t.Errorf("rollback migration is missing %q", contract)
		}
	}
}
