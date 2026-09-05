package notificationsrepository

import (
	notificationssql "github.com/complexus-tech/projects-api/internal/modules/notifications/repository/sqlc"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository is the native pgx persistence adapter for notifications. SQLC's
// generated package remains private to this adapter so database types do not
// leak into services or transports.
type Repository struct {
	queries notificationssql.Querier
	pool    *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Repository {
	return &Repository{queries: notificationssql.New(pool), pool: pool}
}
