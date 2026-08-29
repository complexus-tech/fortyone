package keyresultsrepository

import (
	"errors"
	"math"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	keyresultsdomain "github.com/complexus-tech/projects-api/internal/modules/keyresults/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func TestKeyResultQueriesEnforceTenantActorAndTeamScope(t *testing.T) {
	t.Parallel()

	reads := readQueryFile(t, "queries/reads.sql")
	mutations := readQueryFile(t, "queries/mutations.sql")
	for name, query := range map[string]string{"reads": reads, "mutations": mutations} {
		for _, contract := range []string{
			"objective.workspace_id = sqlc.arg(workspace_id)",
			"public.workspace_members",
			"public.team_members",
			"actor.is_active = TRUE",
			"membership.role IN ('member', 'admin')",
		} {
			if !strings.Contains(query, contract) {
				t.Errorf("%s query set is missing %q", name, contract)
			}
		}
	}
	for _, contract := range []string{
		"sqlc.arg(all_teams)",
		"sqlc.arg(allowed_team_ids)",
		"objective.team_id = ANY",
	} {
		if !strings.Contains(reads, contract) || !strings.Contains(mutations, contract) {
			t.Errorf("credential team restriction is missing %q", contract)
		}
	}
}

func TestKeyResultQueriesAreTypedStaticAndDeterministicallyOrdered(t *testing.T) {
	t.Parallel()

	queries := readQueryFile(t, "queries/reads.sql") + "\n" + readQueryFile(t, "queries/mutations.sql")
	wildcard := regexp.MustCompile(`(?i)select\s+(?:[a-z_][a-z0-9_]*\.)?\*`)
	if wildcard.MatchString(queries) {
		t.Fatal("key-result SQL contains a wildcard projection")
	}
	for _, contract := range []string{
		"CASE WHEN CAST(sqlc.arg(sort_key) AS text)",
		"key_result.created_at DESC,",
		"key_result.id DESC",
		"CAST(sqlc.arg(start_value) AS double precision)",
		"CAST(sqlc.arg(current_value) AS double precision)",
		"CAST(sqlc.arg(target_value) AS double precision)",
		"FOR UPDATE OF key_result",
	} {
		if !strings.Contains(queries, contract) {
			t.Errorf("typed/deterministic query contract is missing %q", contract)
		}
	}
	if strings.Contains(queries, "ORDER BY "+"sqlc.arg") {
		t.Fatal("untrusted order expression is interpolated into key-result SQL")
	}
}

func TestKeyResultMutationQueriesKeepAggregateAndActivityWritesAtomicCapable(t *testing.T) {
	t.Parallel()

	queries := readQueryFile(t, "queries/mutations.sql")
	for _, queryName := range []string{
		"CreateKeyResult", "CreateKeyResultActivity", "GetKeyResultForMutation",
		"UpdateKeyResult", "DeleteKeyResult", "CreateDeletedKeyResultActivity",
	} {
		if !strings.Contains(queries, "-- name: "+queryName+" ") {
			t.Errorf("mutation query set is missing %s", queryName)
		}
	}
	if !strings.Contains(queries, "GREATEST(clock_timestamp(), key_result.updated_at + INTERVAL '1 microsecond')") {
		t.Fatal("key-result updates do not guarantee a monotonic CAS timestamp")
	}
}

func TestKeyResultRepositoryIntegerAndPatchHelpers(t *testing.T) {
	t.Parallel()

	if first, err := firstSequenceID(10, 3); err != nil || first != 8 {
		t.Fatalf("firstSequenceID(10,3) = %d, %v; want 8", first, err)
	}
	if _, err := firstSequenceID(1, 3); !errors.Is(err, keyresultsdomain.ErrInvalid) {
		t.Fatalf("invalid sequence range error = %v, want ErrInvalid", err)
	}
	if offset, err := checkedOffset(3, 20); err != nil || offset != 40 {
		t.Fatalf("checkedOffset(3,20) = %d, %v; want 40", offset, err)
	}
	if _, err := checkedOffset(math.MaxInt, 100); !errors.Is(err, keyresultsdomain.ErrInvalid) {
		t.Fatalf("overflow offset error = %v, want ErrInvalid", err)
	}

	start := time.Date(2026, time.August, 28, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 0, 1)
	current := keyresultsdomain.KeyResult{Name: "Old", CurrentValue: 5, StartDate: &start, EndDate: &end}
	fields := changedFields(current, keyresultsdomain.Patch{
		Name: keyresultsdomain.SetField("New"), CurrentValue: keyresultsdomain.SetField(0.0),
	})
	if strings.Join(fields, ",") != "name,current_value" {
		t.Fatalf("changedFields() = %#v", fields)
	}
	reversed := start.AddDate(0, 0, -1)
	if err := validateEffectiveDates(current, keyresultsdomain.Patch{EndDate: keyresultsdomain.SetField(&reversed)}); !errors.Is(err, keyresultsdomain.ErrInvalid) {
		t.Fatalf("reversed date error = %v, want ErrInvalid", err)
	}
}

func TestMapDatabaseErrorPreservesStableDomainErrors(t *testing.T) {
	t.Parallel()

	if !errors.Is(mapDatabaseError(pgx.ErrNoRows), keyresultsdomain.ErrNotFound) {
		t.Fatal("pgx.ErrNoRows did not map to ErrNotFound")
	}
	if !errors.Is(mapDatabaseError(&pgconn.PgError{Code: "23503"}), keyresultsdomain.ErrInvalidReference) {
		t.Fatal("foreign-key violation did not map to ErrInvalidReference")
	}
	want := errors.New("network unavailable")
	if !errors.Is(mapDatabaseError(want), want) {
		t.Fatal("unknown database error was not preserved")
	}
}

func readQueryFile(t *testing.T, path string) string {
	t.Helper()
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(contents)
}

func TestEqualUUIDPointer(t *testing.T) {
	t.Parallel()
	id := uuid.New()
	copyID := id
	if !equalUUIDPointer(&id, &copyID) || equalUUIDPointer(&id, nil) {
		t.Fatal("equalUUIDPointer() did not preserve nil/value semantics")
	}
}
