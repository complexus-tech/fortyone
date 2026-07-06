package database

import (
	"testing"

	"github.com/jmoiron/sqlx"
)

func TestPGXUsesPostgresBindVars(t *testing.T) {
	if got := sqlx.BindType("pgx"); got != sqlx.DOLLAR {
		t.Fatalf("pgx bind type = %d, want %d", got, sqlx.DOLLAR)
	}
}
