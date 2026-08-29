package contract

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestValidateManifest(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		mutate    func(*Manifest)
		wantError string
	}{
		{
			name: "valid",
		},
		{
			name: "fixed baseline",
			mutate: func(manifest *Manifest) {
				manifest.BaselineVersion++
			},
			wantError: "established baseline",
		},
		{
			name: "contiguous versions",
			mutate: func(manifest *Manifest) {
				manifest.Migrations[0].Version = 153
			},
			wantError: "contiguous ordered version",
		},
		{
			name: "exact filenames",
			mutate: func(manifest *Manifest) {
				manifest.Migrations[0].Files.Up = "000152_wrong.up.sql"
			},
			wantError: "up file",
		},
		{
			name: "known classification",
			mutate: func(manifest *Manifest) {
				manifest.Migrations[0].Classification = "sometimes"
			},
			wantError: "classification",
		},
		{
			name: "classification recovery consistency",
			mutate: func(manifest *Manifest) {
				manifest.Migrations[0].Recovery.Strategy = RecoveryDownMigration
			},
			wantError: `forward-only migration must use "forward-fix" recovery`,
		},
		{
			name: "policy vocabulary",
			mutate: func(manifest *Manifest) {
				manifest.Policy.AppliedMigrationMutation = "rewrite-on-demand"
			},
			wantError: "applied_migration_mutation",
		},
		{
			name: "compatibility required",
			mutate: func(manifest *Manifest) {
				manifest.Migrations[0].Compatibility.Worker = ""
			},
			wantError: "worker compatibility is required",
		},
		{
			name: "procedure required",
			mutate: func(manifest *Manifest) {
				manifest.Migrations[0].Recovery.Procedure = nil
			},
			wantError: "recovery procedure must not be empty",
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			manifest := validTestManifest()
			if test.mutate != nil {
				test.mutate(&manifest)
			}

			err := ValidateManifest(manifest)
			if test.wantError == "" {
				if err != nil {
					t.Fatalf("ValidateManifest() error = %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), test.wantError) {
				t.Fatalf("ValidateManifest() error = %v, want substring %q", err, test.wantError)
			}
		})
	}
}

func TestDecodeManifestIsStrict(t *testing.T) {
	t.Parallel()

	for name, input := range map[string]string{
		"unknown field":  `{"format_version":1,"unknown":true}`,
		"trailing value": `{"format_version":1} {"format_version":1}`,
	} {
		name, input := name, input
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if _, err := decodeManifest(strings.NewReader(input)); err == nil {
				t.Fatal("decodeManifest() expected an error")
			}
		})
	}
}

func TestValidateMigrationDirectory(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		files     []string
		wantError string
	}{
		{
			name: "paired and represented",
			files: []string{
				"000152_harden_tokens.up.sql",
				"000152_harden_tokens.down.sql",
			},
		},
		{
			name: "missing direction",
			files: []string{
				"000152_harden_tokens.up.sql",
			},
			wantError: "missing its down file",
		},
		{
			name: "unrepresented migration",
			files: []string{
				"000152_harden_tokens.up.sql",
				"000152_harden_tokens.down.sql",
				"000153_unlisted.up.sql",
				"000153_unlisted.down.sql",
			},
			wantError: "is not represented in the manifest",
		},
		{
			name: "different slugs for pair",
			files: []string{
				"000152_harden_tokens.up.sql",
				"000152_other_name.down.sql",
			},
			wantError: "uses multiple slugs",
		},
		{
			name: "malformed filename",
			files: []string{
				"152_harden_tokens.up.sql",
				"000152_harden_tokens.down.sql",
			},
			wantError: "does not match",
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			directory := t.TempDir()
			for _, name := range test.files {
				if err := os.WriteFile(filepath.Join(directory, name), []byte("-- fixture\n"), 0o600); err != nil {
					t.Fatalf("write migration fixture: %v", err)
				}
			}

			root, err := os.OpenRoot(directory)
			if err != nil {
				t.Fatalf("open migration fixture root: %v", err)
			}
			t.Cleanup(func() { _ = root.Close() })

			err = ValidateMigrationDirectory(root.FS(), ".", validTestManifest())
			if test.wantError == "" {
				if err != nil {
					t.Fatalf("ValidateMigrationDirectory() error = %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), test.wantError) {
				t.Fatalf("ValidateMigrationDirectory() error = %v, want substring %q", err, test.wantError)
			}
		})
	}
}

func TestRepositoryMigrationContract(t *testing.T) {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve contract test source path")
	}
	serverRoot := filepath.Clean(filepath.Join(filepath.Dir(currentFile), "..", "..", ".."))
	root, err := os.OpenRoot(serverRoot)
	if err != nil {
		t.Fatalf("open server root: %v", err)
	}
	t.Cleanup(func() { _ = root.Close() })

	manifest, err := LoadManifest(root.FS(), ManifestFile)
	if err != nil {
		t.Fatalf("load repository manifest: %v", err)
	}
	if err := ValidateRepository(root.FS(), manifest); err != nil {
		t.Fatalf("repository migration contract: %v", err)
	}
}

func validTestManifest() Manifest {
	return Manifest{
		FormatVersion:   FormatVersion,
		BaselineVersion: BaselineVersion,
		Policy: Policy{
			BaselineRule:             baselineRule,
			AppliedMigrationMutation: appliedMigrationRule,
			RequiredDirections:       []string{"up", "down"},
			ProductionRecovery:       productionRecoveryRule,
			ForwardOnlyRecovery:      RecoveryForwardFix,
			ReversibleRecovery:       RecoveryDownMigration,
		},
		Migrations: []Migration{
			{
				Version:        152,
				Slug:           "harden_tokens",
				Classification: ClassificationForwardOnly,
				Files: Files{
					Up:   "000152_harden_tokens.up.sql",
					Down: "000152_harden_tokens.down.sql",
				},
				Compatibility: Compatibility{
					Schema:        "additive",
					API:           "requires schema",
					Worker:        "unaffected",
					MixedVersions: "bounded",
				},
				Rollout: Rollout{
					Sequence: 1,
					Mode:     RolloutSchemaFirst,
					Steps:    []string{"apply schema"},
				},
				Recovery: Recovery{
					Strategy:  RecoveryForwardFix,
					Procedure: []string{"deploy a forward fix"},
				},
				OperationalNotes: []string{"fixture"},
			},
		},
	}
}
