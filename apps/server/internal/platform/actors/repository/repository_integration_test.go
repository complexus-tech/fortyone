//go:build integration

package actorsrepository

import (
	"context"
	"errors"
	"strconv"
	"testing"
	"time"

	"github.com/complexus-tech/projects-api/internal/testkit"
	"github.com/google/uuid"
)

func TestRepositoryFindsOnlyActiveSystemActors(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(t.Context(), 60*time.Second)
	defer cancel()
	postgres := testkit.NewPostgres(t)
	repository := New(postgres.Pool)
	var rawVersion string
	if err := postgres.Pool.QueryRow(ctx, "SHOW server_version_num").Scan(&rawVersion); err != nil {
		t.Fatalf("read PostgreSQL version: %v", err)
	}
	version, err := strconv.Atoi(rawVersion)
	if err != nil || version < 180000 || version >= 190000 {
		t.Fatalf("PostgreSQL version = %q, want 18.x", rawVersion)
	}

	activeID := uuid.New()
	for _, row := range []struct {
		id     uuid.UUID
		email  string
		active bool
		system bool
	}{
		{id: activeID, email: "active-system@example.com", active: true, system: true},
		{id: uuid.New(), email: "inactive-system@example.com", active: false, system: true},
		{id: uuid.New(), email: "active-human@example.com", active: true, system: false},
	} {
		if _, err := postgres.Pool.Exec(ctx, `
			INSERT INTO users (user_id, username, email, full_name, is_active, is_system)
			VALUES ($1, $2, $3, 'Actor test', $4, $5)
		`, row.id, "actor-"+uuid.NewString(), row.email, row.active, row.system); err != nil {
			t.Fatalf("insert actor fixture: %v", err)
		}
	}

	resolved, err := repository.FindActiveSystemActorByEmail(ctx, "active-system@example.com")
	if err != nil || resolved != activeID {
		t.Fatalf("resolved actor = %s, error=%v", resolved, err)
	}
	for _, email := range []string{"inactive-system@example.com", "active-human@example.com", "missing@example.com"} {
		if _, err := repository.FindActiveSystemActorByEmail(ctx, email); !errors.Is(err, ErrSystemActorNotFound) {
			t.Fatalf("lookup %q error = %v, want ErrSystemActorNotFound", email, err)
		}
	}
}
