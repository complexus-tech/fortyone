package slackrepository

import (
	"context"
	"database/sql"

	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type Repo struct {
	log *logger.Logger
	db  *sqlx.DB
}

func New(log *logger.Logger, db *sqlx.DB) *Repo {
	return &Repo{log: log, db: db}
}

type WorkspaceRecord struct {
	ID   uuid.UUID `db:"workspace_id"`
	Slug string    `db:"slug"`
	Name string    `db:"name"`
}

type TeamRecord struct {
	ID   uuid.UUID `db:"team_id"`
	Code string    `db:"code"`
	Name string    `db:"name"`
}

func (r *Repo) FindWorkspaceBySlug(ctx context.Context, slug string) (WorkspaceRecord, error) {
	var row WorkspaceRecord
	err := r.db.GetContext(ctx, &row, `
		SELECT workspace_id, slug, name
		FROM workspaces
		WHERE slug = $1 AND deleted_at IS NULL
	`, slug)
	if err != nil {
		return WorkspaceRecord{}, err
	}
	return row, nil
}

func (r *Repo) FindTeamByCode(ctx context.Context, workspaceID uuid.UUID, code string) (TeamRecord, error) {
	var row TeamRecord
	err := r.db.GetContext(ctx, &row, `
		SELECT team_id, code, name
		FROM teams
		WHERE workspace_id = $1 AND LOWER(code) = LOWER($2)
		LIMIT 1
	`, workspaceID, code)
	if err != nil {
		return TeamRecord{}, err
	}
	return row, nil
}

func IsNotFound(err error) bool {
	return err == sql.ErrNoRows
}
