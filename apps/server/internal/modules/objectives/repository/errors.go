package objectivesrepository

import (
	"errors"

	objectivesdomain "github.com/complexus-tech/projects-api/internal/modules/objectives/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func mapDatabaseError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return objectivesdomain.ErrNotFound
	}

	var postgresError *pgconn.PgError
	if errors.As(err, &postgresError) {
		switch postgresError.Code {
		case "23505":
			if postgresError.ConstraintName == "objectives_name_team_unique" ||
				postgresError.ConstraintName == "strategic_pillars_workspace_name_unique" {
				return objectivesdomain.ErrNameExists
			}
		case "23502", "23503", "23514", "22001", "22P02":
			return objectivesdomain.ErrInvalidReference
		}
	}
	return err
}
