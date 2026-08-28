package figmarepository

import (
	"errors"
	"fmt"

	figmadomain "github.com/complexus-tech/projects-api/internal/modules/figma/domain"
	platformdatabase "github.com/complexus-tech/projects-api/internal/platform/database"
	"github.com/jackc/pgx/v5"
)

func mapDatabaseError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return figmadomain.ErrNotFound
	}
	for _, domainError := range []error{
		figmadomain.ErrNotFound,
		figmadomain.ErrConflict,
		figmadomain.ErrForbidden,
	} {
		if errors.Is(err, domainError) {
			return err
		}
	}
	switch platformdatabase.Classify(err) {
	case platformdatabase.ErrorClassUniqueViolation,
		platformdatabase.ErrorClassSerializationFailure,
		platformdatabase.ErrorClassDeadlock:
		return fmt.Errorf("%w: %v", figmadomain.ErrConflict, err)
	case platformdatabase.ErrorClassForeignKeyViolation,
		platformdatabase.ErrorClassNotNullViolation,
		platformdatabase.ErrorClassCheckViolation:
		return fmt.Errorf("%w: %v", figmadomain.ErrForbidden, err)
	default:
		return err
	}
}

func requireAffected(rows int64, err error, zeroError error) error {
	if err != nil {
		return mapDatabaseError(err)
	}
	if rows != 1 {
		return zeroError
	}
	return nil
}
