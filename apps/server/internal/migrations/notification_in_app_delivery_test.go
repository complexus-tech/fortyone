package migrations

import (
	"strings"
	"testing"
)

const notificationInAppDeliveryMigration = "000136_notification_in_app_delivery"

func TestNotificationInAppDeliveryMigrationContracts(t *testing.T) {
	t.Parallel()

	data, err := FS.ReadFile(notificationInAppDeliveryMigration + ".up.sql")
	if err != nil {
		t.Fatalf("read forward migration: %v", err)
	}
	migration := string(data)

	for _, contract := range []string{
		"ADD COLUMN IF NOT EXISTS in_app_enabled boolean NOT NULL DEFAULT true",
		"CAST(type AS text) = 'strategy_update'",
		"SET in_app_enabled = false",
		"WHERE in_app_enabled = true",
	} {
		if !strings.Contains(migration, contract) {
			t.Fatalf("migration is missing contract %q", contract)
		}
	}
}

func TestNotificationInAppDeliveryMigrationIsForwardOnly(t *testing.T) {
	t.Parallel()

	data, err := FS.ReadFile(notificationInAppDeliveryMigration + ".down.sql")
	if err != nil {
		t.Fatalf("read rollback migration: %v", err)
	}
	if !strings.Contains(string(data), "migration 000136 is forward-only") {
		t.Fatal("migration must document its forward-only rollback")
	}
}
