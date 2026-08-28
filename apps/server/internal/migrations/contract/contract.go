// Package contract validates the operational contract for database migrations
// created after the established schema baseline.
package contract

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

const (
	// FormatVersion is the only manifest format understood by this package.
	FormatVersion = 1
	// BaselineVersion is the last migration governed by the established legacy
	// policy. Every newer migration must be represented in the manifest.
	BaselineVersion = 151
)

const (
	// ManifestFile is the manifest path relative to the server root.
	ManifestFile = "internal/migrations/manifest.json"
	// MigrationsDirectory is the SQL migration path relative to the server root.
	MigrationsDirectory = "internal/migrations"
	// DocumentationFile is the generated operator-guide path relative to the
	// server root.
	DocumentationFile = "docs/database/migration-operations.md"
)

const (
	ClassificationReversible  = "reversible"
	ClassificationForwardOnly = "forward-only"

	RecoveryDownMigration = "down-migration"
	RecoveryForwardFix    = "forward-fix"

	RolloutSchemaFirst        = "schema-first"
	RolloutCoordinatedCutover = "coordinated-cutover"
)

const (
	baselineRule           = "versions-at-or-below-baseline-are-established"
	appliedMigrationRule   = "immutable"
	productionRecoveryRule = "classification-driven"
)

var (
	migrationFilenamePattern = regexp.MustCompile(`^(\d{6})_([a-z][a-z0-9_]*)\.(up|down)\.sql$`)
	slugPattern              = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)
)

// Manifest is the machine-readable source of truth for post-baseline
// migration rollout and recovery decisions.
type Manifest struct {
	FormatVersion   int         `json:"format_version"`
	BaselineVersion int         `json:"baseline_version"`
	Policy          Policy      `json:"policy"`
	Migrations      []Migration `json:"migrations"`
}

// Policy records repository-wide migration invariants. Values are explicit so
// policy changes require both code review and validator changes.
type Policy struct {
	BaselineRule             string   `json:"baseline_rule"`
	AppliedMigrationMutation string   `json:"applied_migration_mutation"`
	RequiredDirections       []string `json:"required_directions"`
	ProductionRecovery       string   `json:"production_recovery"`
	ForwardOnlyRecovery      string   `json:"forward_only_recovery"`
	ReversibleRecovery       string   `json:"reversible_recovery"`
}

// Migration describes one paired golang-migrate schema change.
type Migration struct {
	Version          int           `json:"version"`
	Slug             string        `json:"slug"`
	Classification   string        `json:"classification"`
	Files            Files         `json:"files"`
	Compatibility    Compatibility `json:"compatibility"`
	Rollout          Rollout       `json:"rollout"`
	Recovery         Recovery      `json:"recovery"`
	OperationalNotes []string      `json:"operational_notes"`
}

// Files names the exact migration pair represented by an entry.
type Files struct {
	Up   string `json:"up"`
	Down string `json:"down"`
}

// Compatibility records how the schema change interacts with each runtime.
type Compatibility struct {
	Schema        string `json:"schema"`
	API           string `json:"api"`
	Worker        string `json:"worker"`
	MixedVersions string `json:"mixed_versions"`
}

// Rollout records the release order and whether mixed binaries are safe.
type Rollout struct {
	Sequence int      `json:"sequence"`
	Mode     string   `json:"mode"`
	Steps    []string `json:"steps"`
}

// Recovery records the only supported production response after a migration
// has been activated.
type Recovery struct {
	Strategy  string   `json:"strategy"`
	Procedure []string `json:"procedure"`
}

// LoadManifest decodes a manifest and rejects unknown or trailing JSON. A
// strict decoder prevents misspelled operational fields from being ignored.
func LoadManifest(filesystem fs.FS, path string) (Manifest, error) {
	file, err := filesystem.Open(path)
	if err != nil {
		return Manifest{}, fmt.Errorf("open migration manifest: %w", err)
	}
	defer file.Close()

	manifest, err := decodeManifest(file)
	if err != nil {
		return Manifest{}, fmt.Errorf("decode migration manifest: %w", err)
	}
	return manifest, nil
}

func decodeManifest(reader io.Reader) (Manifest, error) {
	decoder := json.NewDecoder(reader)
	decoder.DisallowUnknownFields()

	var manifest Manifest
	if err := decoder.Decode(&manifest); err != nil {
		return Manifest{}, err
	}

	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		if err == nil {
			return Manifest{}, errors.New("manifest contains more than one JSON value")
		}
		return Manifest{}, fmt.Errorf("read trailing manifest data: %w", err)
	}
	return manifest, nil
}

// ValidateManifest verifies the fixed baseline, policy vocabulary, entry
// ordering, filenames, classifications, compatibility, and recovery contract.
func ValidateManifest(manifest Manifest) error {
	var problems []error
	if manifest.FormatVersion != FormatVersion {
		problems = append(problems, fmt.Errorf("format_version = %d, want %d", manifest.FormatVersion, FormatVersion))
	}
	if manifest.BaselineVersion != BaselineVersion {
		problems = append(problems, fmt.Errorf("baseline_version = %d, want established baseline %d", manifest.BaselineVersion, BaselineVersion))
	}
	problems = append(problems, validatePolicy(manifest.Policy)...)

	seenVersions := make(map[int]struct{}, len(manifest.Migrations))
	for index, migration := range manifest.Migrations {
		label := fmt.Sprintf("migration[%d]", index)
		if migration.Version <= manifest.BaselineVersion {
			problems = append(problems, fmt.Errorf("%s version %d is not above baseline %d", label, migration.Version, manifest.BaselineVersion))
		}
		expectedVersion := manifest.BaselineVersion + index + 1
		if migration.Version != expectedVersion {
			problems = append(problems, fmt.Errorf("%s version = %d, want contiguous ordered version %d", label, migration.Version, expectedVersion))
		}
		if _, exists := seenVersions[migration.Version]; exists {
			problems = append(problems, fmt.Errorf("migration version %d is represented more than once", migration.Version))
		}
		seenVersions[migration.Version] = struct{}{}

		if !slugPattern.MatchString(migration.Slug) {
			problems = append(problems, fmt.Errorf("%s slug %q must be lower snake case", label, migration.Slug))
		}
		expectedPrefix := fmt.Sprintf("%06d_%s", migration.Version, migration.Slug)
		if migration.Files.Up != expectedPrefix+".up.sql" {
			problems = append(problems, fmt.Errorf("%s up file = %q, want %q", label, migration.Files.Up, expectedPrefix+".up.sql"))
		}
		if migration.Files.Down != expectedPrefix+".down.sql" {
			problems = append(problems, fmt.Errorf("%s down file = %q, want %q", label, migration.Files.Down, expectedPrefix+".down.sql"))
		}

		switch migration.Classification {
		case ClassificationReversible:
			if migration.Recovery.Strategy != RecoveryDownMigration {
				problems = append(problems, fmt.Errorf("%s reversible migration must use %q recovery", label, RecoveryDownMigration))
			}
		case ClassificationForwardOnly:
			if migration.Recovery.Strategy != RecoveryForwardFix {
				problems = append(problems, fmt.Errorf("%s forward-only migration must use %q recovery", label, RecoveryForwardFix))
			}
		default:
			problems = append(problems, fmt.Errorf("%s classification %q is not supported", label, migration.Classification))
		}

		if strings.TrimSpace(migration.Compatibility.Schema) == "" {
			problems = append(problems, fmt.Errorf("%s schema compatibility is required", label))
		}
		if strings.TrimSpace(migration.Compatibility.API) == "" {
			problems = append(problems, fmt.Errorf("%s API compatibility is required", label))
		}
		if strings.TrimSpace(migration.Compatibility.Worker) == "" {
			problems = append(problems, fmt.Errorf("%s worker compatibility is required", label))
		}
		if strings.TrimSpace(migration.Compatibility.MixedVersions) == "" {
			problems = append(problems, fmt.Errorf("%s mixed-version compatibility is required", label))
		}

		if migration.Rollout.Sequence != index+1 {
			problems = append(problems, fmt.Errorf("%s rollout sequence = %d, want %d", label, migration.Rollout.Sequence, index+1))
		}
		switch migration.Rollout.Mode {
		case RolloutSchemaFirst, RolloutCoordinatedCutover:
		default:
			problems = append(problems, fmt.Errorf("%s rollout mode %q is not supported", label, migration.Rollout.Mode))
		}
		problems = append(problems, validateStringList(label+" rollout steps", migration.Rollout.Steps)...)
		problems = append(problems, validateStringList(label+" recovery procedure", migration.Recovery.Procedure)...)
		problems = append(problems, validateStringList(label+" operational notes", migration.OperationalNotes)...)
	}

	return errors.Join(problems...)
}

func validatePolicy(policy Policy) []error {
	var problems []error
	if policy.BaselineRule != baselineRule {
		problems = append(problems, fmt.Errorf("policy baseline_rule = %q, want %q", policy.BaselineRule, baselineRule))
	}
	if policy.AppliedMigrationMutation != appliedMigrationRule {
		problems = append(problems, fmt.Errorf("policy applied_migration_mutation = %q, want %q", policy.AppliedMigrationMutation, appliedMigrationRule))
	}
	if len(policy.RequiredDirections) != 2 || policy.RequiredDirections[0] != "up" || policy.RequiredDirections[1] != "down" {
		problems = append(problems, errors.New(`policy required_directions must be exactly ["up", "down"]`))
	}
	if policy.ProductionRecovery != productionRecoveryRule {
		problems = append(problems, fmt.Errorf("policy production_recovery = %q, want %q", policy.ProductionRecovery, productionRecoveryRule))
	}
	if policy.ForwardOnlyRecovery != RecoveryForwardFix {
		problems = append(problems, fmt.Errorf("policy forward_only_recovery = %q, want %q", policy.ForwardOnlyRecovery, RecoveryForwardFix))
	}
	if policy.ReversibleRecovery != RecoveryDownMigration {
		problems = append(problems, fmt.Errorf("policy reversible_recovery = %q, want %q", policy.ReversibleRecovery, RecoveryDownMigration))
	}
	return problems
}

func validateStringList(label string, values []string) []error {
	if len(values) == 0 {
		return []error{fmt.Errorf("%s must not be empty", label)}
	}

	var problems []error
	for index, value := range values {
		if strings.TrimSpace(value) == "" {
			problems = append(problems, fmt.Errorf("%s[%d] must not be blank", label, index))
		}
	}
	return problems
}

type migrationPair struct {
	slug string
	up   string
	down string
}

// ValidateMigrationDirectory proves that every migration above the baseline
// has one up file, one down file, and one matching manifest entry.
func ValidateMigrationDirectory(filesystem fs.FS, directory string, manifest Manifest) error {
	entries, err := fs.ReadDir(filesystem, directory)
	if err != nil {
		return fmt.Errorf("read migration directory: %w", err)
	}

	pairs := make(map[int]migrationPair)
	var problems []error
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".up.sql") && !strings.HasSuffix(name, ".down.sql") {
			continue
		}

		matches := migrationFilenamePattern.FindStringSubmatch(name)
		if matches == nil {
			problems = append(problems, fmt.Errorf("migration filename %q does not match NNNNNN_lower_snake_case.(up|down).sql", name))
			continue
		}
		version, parseErr := strconv.Atoi(matches[1])
		if parseErr != nil {
			problems = append(problems, fmt.Errorf("parse migration version in %q: %w", name, parseErr))
			continue
		}
		if version <= manifest.BaselineVersion {
			continue
		}

		pair := pairs[version]
		if pair.slug != "" && pair.slug != matches[2] {
			problems = append(problems, fmt.Errorf("migration version %06d uses multiple slugs: %q and %q", version, pair.slug, matches[2]))
		}
		pair.slug = matches[2]
		switch matches[3] {
		case "up":
			if pair.up != "" {
				problems = append(problems, fmt.Errorf("migration version %06d has multiple up files", version))
			}
			pair.up = name
		case "down":
			if pair.down != "" {
				problems = append(problems, fmt.Errorf("migration version %06d has multiple down files", version))
			}
			pair.down = name
		}
		pairs[version] = pair
	}

	manifestByVersion := make(map[int]Migration, len(manifest.Migrations))
	for _, migration := range manifest.Migrations {
		manifestByVersion[migration.Version] = migration
	}

	versions := make([]int, 0, len(pairs))
	for version := range pairs {
		versions = append(versions, version)
	}
	sort.Ints(versions)
	for _, version := range versions {
		pair := pairs[version]
		if pair.up == "" {
			problems = append(problems, fmt.Errorf("migration %06d_%s is missing its up file", version, pair.slug))
		}
		if pair.down == "" {
			problems = append(problems, fmt.Errorf("migration %06d_%s is missing its down file", version, pair.slug))
		}
		migration, represented := manifestByVersion[version]
		if !represented {
			problems = append(problems, fmt.Errorf("migration %06d_%s is not represented in the manifest", version, pair.slug))
			continue
		}
		if migration.Files.Up != pair.up || migration.Files.Down != pair.down {
			problems = append(problems, fmt.Errorf("migration %06d manifest files do not match directory pair %q and %q", version, pair.up, pair.down))
		}
	}

	for _, migration := range manifest.Migrations {
		if _, exists := pairs[migration.Version]; !exists {
			problems = append(problems, fmt.Errorf("manifest migration %06d_%s has no migration files", migration.Version, migration.Slug))
		}
	}

	return errors.Join(problems...)
}

// ValidateDocumentation verifies that the checked-in operator guide is an
// exact rendering of the manifest and fixed policy text.
func ValidateDocumentation(filesystem fs.FS, path string, manifest Manifest) error {
	current, err := fs.ReadFile(filesystem, path)
	if err != nil {
		return fmt.Errorf("read migration operator guide: %w", err)
	}
	expected := RenderDocumentation(manifest)
	if !bytes.Equal(current, expected) {
		return errors.New("migration operator guide is stale; run make migration-docs")
	}
	return nil
}

// ValidateRepository runs the complete migration contract check.
func ValidateRepository(filesystem fs.FS, manifest Manifest) error {
	return errors.Join(
		ValidateManifest(manifest),
		ValidateMigrationDirectory(filesystem, MigrationsDirectory, manifest),
		ValidateDocumentation(filesystem, DocumentationFile, manifest),
	)
}
