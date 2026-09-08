package notificationsrepository

import (
	"errors"
	"testing"

	notificationsdomain "github.com/complexus-tech/projects-api/internal/modules/notifications/domain"
	"github.com/jackc/pgx/v5/pgconn"
)

func TestMapWriteErrorPreservesDatabaseDiagnostic(t *testing.T) {
	for _, test := range []struct {
		code string
		kind error
	}{
		{"23505", notificationsdomain.ErrConflict},
		{"23503", notificationsdomain.ErrNotFound},
		{"23502", notificationsdomain.ErrInvalid},
		{"23514", notificationsdomain.ErrInvalid},
	} {
		t.Run(test.code, func(t *testing.T) {
			cause := &pgconn.PgError{Code: test.code, ConstraintName: "notification_constraint"}
			got := mapWriteError("create notification", cause)
			var databaseError *pgconn.PgError
			if !errors.Is(got, test.kind) || !errors.As(got, &databaseError) || databaseError != cause {
				t.Fatalf("write error lost its domain classification or database cause: %v", got)
			}
		})
	}
}
