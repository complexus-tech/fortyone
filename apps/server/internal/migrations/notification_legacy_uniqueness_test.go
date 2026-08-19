package migrations

import (
	"strings"
	"testing"
)

const removeLegacyNotificationUniquenessMigration = "000134_remove_legacy_notification_entity_uniqueness"

func TestRemoveLegacyNotificationUniquenessMigrationContracts(t *testing.T) {
	t.Parallel()

	data, err := FS.ReadFile(removeLegacyNotificationUniquenessMigration + ".up.sql")
	if err != nil {
		t.Fatalf("read forward migration: %v", err)
	}
	migration := string(data)

	for _, contract := range []string{
		"table_class.relname = 'notifications'",
		"index_row.indisunique = true",
		"ARRAY['recipient_id', 'workspace_id', 'entity_id']::name[]",
		"ARRAY['recipient_id', 'workspace_id', 'entity_id', 'entity_type']::name[]",
		"ALTER TABLE public.notifications DROP CONSTRAINT",
		"DROP INDEX public.%I",
	} {
		if !strings.Contains(migration, contract) {
			t.Fatalf("migration is missing contract %q", contract)
		}
	}
}

func TestRemoveLegacyNotificationUniquenessMigrationIsForwardOnly(t *testing.T) {
	t.Parallel()

	data, err := FS.ReadFile(removeLegacyNotificationUniquenessMigration + ".down.sql")
	if err != nil {
		t.Fatalf("read rollback migration: %v", err)
	}
	migration := string(data)

	if !strings.Contains(migration, "migration 000134 is forward-only") {
		t.Fatal("migration must document its forward-only rollback")
	}
	if strings.Contains(migration, "CREATE UNIQUE") {
		t.Fatal("rollback must not restore legacy notification uniqueness")
	}
}
