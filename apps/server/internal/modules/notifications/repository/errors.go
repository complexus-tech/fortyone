package notificationsrepository

import (
	"errors"
	"fmt"

	notificationsdomain "github.com/complexus-tech/projects-api/internal/modules/notifications/domain"
	platformdatabase "github.com/complexus-tech/projects-api/internal/platform/database"
	"github.com/jackc/pgx/v5"
)

func mapWriteError(operation string, err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("%s: %w", operation, notificationsdomain.ErrNotFound)
	}
	switch platformdatabase.Classify(err) {
	case platformdatabase.ErrorClassUniqueViolation:
		return fmt.Errorf("%s: %w", operation, notificationsdomain.ErrConflict)
	case platformdatabase.ErrorClassForeignKeyViolation:
		return fmt.Errorf("%s: %w", operation, notificationsdomain.ErrNotFound)
	case platformdatabase.ErrorClassNotNullViolation, platformdatabase.ErrorClassCheckViolation:
		return fmt.Errorf("%s: %w", operation, notificationsdomain.ErrInvalid)
	default:
		return fmt.Errorf("%s: %w", operation, err)
	}
}
