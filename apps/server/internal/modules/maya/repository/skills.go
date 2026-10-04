package mayarepository

import (
	"context"
	"errors"
	"fmt"
	"time"

	mayadomain "github.com/complexus-tech/projects-api/internal/modules/maya/domain"
	mayasql "github.com/complexus-tech/projects-api/internal/modules/maya/repository/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func (r *Repo) ListSkills(ctx context.Context, scope mayadomain.SkillScope) ([]mayadomain.Skill, error) {
	if err := r.configured(); err != nil {
		return nil, err
	}
	rows, err := r.queries.ListMayaSkills(ctx, mayasql.ListMayaSkillsParams{WorkspaceID: scope.WorkspaceID, UserID: scope.UserID})
	if err != nil {
		return nil, fmt.Errorf("list Maya skills: %w", err)
	}
	skills := make([]mayadomain.Skill, 0, len(rows))
	for _, row := range rows {
		skills = append(skills, mayadomain.Skill(row))
	}
	return skills, nil
}

func (r *Repo) CreateSkill(ctx context.Context, scope mayadomain.SkillScope, content mayadomain.SkillContent) (mayadomain.Skill, error) {
	if err := r.configured(); err != nil {
		return mayadomain.Skill{}, err
	}
	row, err := r.queries.CreateMayaSkill(ctx, mayasql.CreateMayaSkillParams{
		WorkspaceID: scope.WorkspaceID, UserID: scope.UserID,
		Name: content.Name, Description: content.Description, Instructions: content.Instructions,
	})
	if err != nil {
		return mayadomain.Skill{}, skillWriteError("create Maya skill", err)
	}
	return mayadomain.Skill(row), nil
}

func (r *Repo) UpdateSkill(ctx context.Context, scope mayadomain.SkillScope, id uuid.UUID, content mayadomain.SkillContent, updatedAt time.Time) (mayadomain.Skill, error) {
	if err := r.configured(); err != nil {
		return mayadomain.Skill{}, err
	}
	row, err := r.queries.UpdateMayaSkill(ctx, mayasql.UpdateMayaSkillParams{
		WorkspaceID: scope.WorkspaceID, UserID: scope.UserID, ID: id,
		Name: content.Name, Description: content.Description, Instructions: content.Instructions,
		ExpectedUpdatedAt: updatedAt,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		exists, lookupErr := r.queries.MayaSkillExists(ctx, mayasql.MayaSkillExistsParams{
			WorkspaceID: scope.WorkspaceID, UserID: scope.UserID, ID: id,
		})
		if lookupErr != nil {
			return mayadomain.Skill{}, fmt.Errorf("check Maya skill version: %w", lookupErr)
		}
		if exists {
			return mayadomain.Skill{}, mayadomain.ErrSkillChanged
		}
		return mayadomain.Skill{}, mayadomain.ErrSkillNotFound
	}
	if err != nil {
		return mayadomain.Skill{}, skillWriteError("update Maya skill", err)
	}
	return mayadomain.Skill(row), nil
}

func (r *Repo) DeleteSkill(ctx context.Context, scope mayadomain.SkillScope, id uuid.UUID) error {
	if err := r.configured(); err != nil {
		return err
	}
	rows, err := r.queries.DeleteMayaSkill(ctx, mayasql.DeleteMayaSkillParams{WorkspaceID: scope.WorkspaceID, UserID: scope.UserID, ID: id})
	if err != nil {
		return fmt.Errorf("delete Maya skill: %w", err)
	}
	if rows != 1 {
		return mayadomain.ErrSkillNotFound
	}
	return nil
}

func skillWriteError(operation string, err error) error {
	var postgresError *pgconn.PgError
	if errors.As(err, &postgresError) && postgresError.Code == "23505" && postgresError.ConstraintName == "maya_skills_owner_name_idx" {
		return mayadomain.ErrSkillNameTaken
	}
	return fmt.Errorf("%s: %w", operation, err)
}
