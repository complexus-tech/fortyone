package objectivesrepository

import (
	"errors"
	"testing"

	objectivesdomain "github.com/complexus-tech/projects-api/internal/modules/objectives/domain"
	"github.com/jackc/pgx/v5/pgconn"
)

func TestMapDatabaseErrorUsesStableObjectiveErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		code       string
		constraint string
		want       error
	}{
		{name: "objective name", code: "23505", constraint: "objectives_name_team_unique", want: objectivesdomain.ErrNameExists},
		{name: "pillar name", code: "23505", constraint: "strategic_pillars_workspace_name_unique", want: objectivesdomain.ErrNameExists},
		{name: "foreign key", code: "23503", want: objectivesdomain.ErrInvalidReference},
		{name: "not null", code: "23502", want: objectivesdomain.ErrInvalidReference},
		{name: "check", code: "23514", want: objectivesdomain.ErrInvalidReference},
		{name: "value too long", code: "22001", want: objectivesdomain.ErrInvalidReference},
		{name: "invalid cast", code: "22P02", want: objectivesdomain.ErrInvalidReference},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			err := mapDatabaseError(&pgconn.PgError{Code: test.code, ConstraintName: test.constraint})
			if !errors.Is(err, test.want) {
				t.Fatalf("mapDatabaseError() = %v, want %v", err, test.want)
			}
		})
	}

	unclassified := &pgconn.PgError{Code: "40001"}
	if err := mapDatabaseError(unclassified); !errors.Is(err, unclassified) {
		t.Fatalf("unclassified error = %v, want original", err)
	}
}
