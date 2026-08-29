package migrations

import (
	"strings"
	"testing"
)

const storyTimeContractMigration = "000130_story_time_contract"

func TestStoryTimeContractForwardMigration(t *testing.T) {
	t.Parallel()

	data, err := FS.ReadFile(storyTimeContractMigration + ".up.sql")
	if err != nil {
		t.Fatalf("read forward migration: %v", err)
	}
	migration := string(data)

	for _, contract := range []string{
		"ADD COLUMN estimated_duration_minutes integer",
		"ADD COLUMN minimum_focus_block_minutes integer",
		"estimated_duration_minutes BETWEEN 1 AND 2400",
		"minimum_focus_block_minutes BETWEEN 1 AND 2400",
		"minimum_focus_block_minutes IS NULL OR estimated_duration_minutes IS NOT NULL",
		"minimum_focus_block_minutes <= estimated_duration_minutes",
		"WHEN 1 THEN 30",
		"WHEN 8 THEN 480",
		"WHEN 1 THEN 240",
		"WHEN 8 THEN 2400",
		"estimate_unit = NULL",
		"WHERE scheme IN ('hours', 'ideal_days')",
		"ALTER COLUMN scheme SET DEFAULT 'tshirt'",
		"CHECK (scheme IN ('points', 'tshirt'))",
	} {
		if !strings.Contains(migration, contract) {
			t.Errorf("migration is missing story time contract %q", contract)
		}
	}

	backfillPosition := strings.Index(migration, "UPDATE public.stories AS story")
	schemeConversionPosition := strings.Index(migration, "UPDATE public.team_estimation_settings")
	if backfillPosition < 0 || schemeConversionPosition < 0 || backfillPosition >= schemeConversionPosition {
		t.Fatal("legacy time estimates must be backfilled before team schemes are converted")
	}
}

func TestStoryTimeContractRollbackProtectsMigratedDurationData(t *testing.T) {
	t.Parallel()

	data, err := FS.ReadFile(storyTimeContractMigration + ".down.sql")
	if err != nil {
		t.Fatalf("read rollback migration: %v", err)
	}
	migration := string(data)

	for _, contract := range []string{
		"migration 000130 is forward-only",
		"legacy time estimates were moved out of estimate_unit",
		"time-based team schemes were collapsed to tshirt",
	} {
		if !strings.Contains(migration, contract) {
			t.Errorf("rollback is missing duration-data protection %q", contract)
		}
	}
	if strings.Contains(migration, "DROP COLUMN") {
		t.Fatal("rollback must not drop the only persisted copy of migrated duration data")
	}
}
