package migrations

import (
	"testing"
	"testing/fstest"
)

func TestSchemaPreflightRequiresExactCleanCandidateVersion(t *testing.T) {
	for _, tc := range []struct {
		name      string
		actual    uint64
		dirty     bool
		wantError bool
	}{
		{"ready", 233, false, false},
		{"pending migration", 232, false, true},
		{"newer schema requires rollback review", 234, false, true},
		{"dirty matching version", 233, true, true},
		{"dirty older version", 232, true, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := validateSchemaVersion(tc.actual, 233, tc.dirty); (err != nil) != tc.wantError {
				t.Fatalf("validateSchemaVersion() = %v, want error %v", err, tc.wantError)
			}
		})
	}
}

func TestLatestSchemaVersionUsesEmbeddedUpMigrations(t *testing.T) {
	source := fstest.MapFS{
		"000229_chat.up.sql":        &fstest.MapFile{},
		"000233_thumbnails.up.sql":  &fstest.MapFile{},
		"000234_unrelated.down.sql": &fstest.MapFile{},
	}
	if got, err := latestSchemaVersion(source); err != nil || got != 233 {
		t.Fatalf("latest schema = %d, %v", got, err)
	}
	for _, source := range []fstest.MapFS{{}, {"broken.up.sql": &fstest.MapFile{}}} {
		if _, err := latestSchemaVersion(source); err == nil {
			t.Fatal("expected invalid migration source to fail closed")
		}
	}
}
