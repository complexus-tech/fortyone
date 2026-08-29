package database

import "testing"

func TestOpenMigrationConnectionRejectsInvalidBudgetsBeforeOpening(t *testing.T) {
	t.Parallel()

	tests := map[string]MigrationConfig{
		"negative maximum": {
			Config: Config{MaxOpenConns: -1},
		},
		"negative idle": {
			MaxIdleConns: -1,
		},
		"idle exceeds maximum": {
			Config:       Config{MaxOpenConns: 2},
			MaxIdleConns: 3,
		},
	}

	for name, cfg := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if _, err := OpenMigrationConnection(cfg); err == nil {
				t.Fatal("expected invalid migration connection budget to fail")
			}
		})
	}
}
