package workspacesrepository

import (
	"context"
	"errors"
	"fmt"

	workspacedomain "github.com/complexus-tech/projects-api/internal/modules/workspaces/domain"
	workspacesql "github.com/complexus-tech/projects-api/internal/modules/workspaces/repository/sqlc"
	platformdatabase "github.com/complexus-tech/projects-api/internal/platform/database"
	"github.com/complexus-tech/projects-api/internal/platform/safecast"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (r *repo) Update(ctx context.Context, workspaceID uuid.UUID, updates workspacedomain.Workspace) (workspacedomain.Workspace, error) {
	if updates.Name == "" && updates.AvatarURL == nil {
		return workspacedomain.Workspace{}, errors.New("update workspace: no fields to update")
	}
	avatarURL := ""
	if updates.AvatarURL != nil {
		avatarURL = *updates.AvatarURL
	}
	row, err := r.queries.UpdateWorkspace(ctx, workspacesql.UpdateWorkspaceParams{
		UpdateName:      updates.Name != "",
		Name:            updates.Name,
		UpdateAvatarURL: updates.AvatarURL != nil,
		AvatarURL:       avatarURL,
		WorkspaceID:     workspaceID,
	})
	if err != nil {
		return workspacedomain.Workspace{}, mapWorkspaceNotFound("update workspace", err)
	}
	return workspaceFromRecord(workspaceRecordFromUpdateRow(row)), nil
}

func (r *repo) Delete(ctx context.Context, workspaceID, deletedBy uuid.UUID) error {
	rows, err := r.queries.SoftDeleteWorkspace(ctx, workspacesql.SoftDeleteWorkspaceParams{
		DeletedBy:   &deletedBy,
		WorkspaceID: workspaceID,
	})
	if err != nil {
		return fmt.Errorf("soft delete workspace: %w", err)
	}
	if rows == 0 {
		return workspacedomain.ErrNotFound
	}
	return nil
}

func (r *repo) Restore(ctx context.Context, workspaceID, _ uuid.UUID) error {
	rows, err := r.queries.RestoreWorkspace(ctx, workspacesql.RestoreWorkspaceParams{WorkspaceID: workspaceID})
	if err != nil {
		return fmt.Errorf("restore workspace: %w", err)
	}
	if rows == 0 {
		return workspacedomain.ErrNotFound
	}
	return nil
}

// WorkspaceTransaction is the workspace-owned persistence capability bound to
// a caller-owned transaction. It exposes no pool-backed query methods.
type WorkspaceTransaction interface {
	CreateWorkspace(context.Context, workspacedomain.Workspace, uuid.UUID) (workspacedomain.Workspace, error)
	AddMember(context.Context, uuid.UUID, uuid.UUID, string) error
	InitializeSettings(context.Context, uuid.UUID) error
}

type workspaceTransaction struct {
	queries workspacesql.Querier
}

func (r *repo) BindWorkspaceTransaction(tx pgx.Tx) (WorkspaceTransaction, error) {
	if tx == nil {
		return nil, errors.New("workspace transaction is required")
	}
	if r == nil || r.bindTransaction == nil {
		return nil, errors.New("workspace transaction binding is unavailable")
	}
	return &workspaceTransaction{queries: r.bindTransaction(tx)}, nil
}

func (transaction *workspaceTransaction) CreateWorkspace(
	ctx context.Context,
	workspace workspacedomain.Workspace,
	createdBy uuid.UUID,
) (workspacedomain.Workspace, error) {
	color, err := generateRandomColor()
	if err != nil {
		return workspacedomain.Workspace{}, err
	}
	row, err := transaction.queries.CreateWorkspace(ctx, workspacesql.CreateWorkspaceParams{
		Name:      workspace.Name,
		Slug:      workspace.Slug,
		Color:     color,
		TeamSize:  workspace.TeamSize,
		CreatedBy: &createdBy,
	})
	if err != nil {
		if platformdatabase.Classify(err) == platformdatabase.ErrorClassUniqueViolation {
			return workspacedomain.Workspace{}, workspacedomain.ErrSlugTaken
		}
		return workspacedomain.Workspace{}, fmt.Errorf("create workspace: %w", err)
	}

	for _, status := range workspacedomain.DefaultObjectiveStatuses {
		orderIndex, err := safecast.Int32(status.OrderIndex)
		if err != nil {
			return workspacedomain.Workspace{}, fmt.Errorf("validate default objective status order: %w", err)
		}
		rows, createErr := transaction.queries.CreateDefaultObjectiveStatus(ctx, workspacesql.CreateDefaultObjectiveStatusParams{
			Name:        status.Name,
			Category:    status.Category,
			OrderIndex:  &orderIndex,
			Color:       status.Color,
			WorkspaceID: row.WorkspaceID,
		})
		if createErr != nil {
			return workspacedomain.Workspace{}, fmt.Errorf("create default objective status %q: %w", status.Name, createErr)
		}
		if rows != 1 {
			return workspacedomain.Workspace{}, fmt.Errorf("create default objective status %q: affected %d rows, want 1", status.Name, rows)
		}
	}
	return workspaceFromRecord(workspaceRecordFromCreateRow(row)), nil
}

func (transaction *workspaceTransaction) AddMember(
	ctx context.Context,
	workspaceID uuid.UUID,
	userID uuid.UUID,
	role string,
) error {
	return addMember(ctx, transaction.queries, workspaceID, userID, role)
}

func (transaction *workspaceTransaction) InitializeSettings(ctx context.Context, workspaceID uuid.UUID) error {
	rows, err := transaction.queries.InitializeWorkspaceSettings(
		ctx,
		workspacesql.InitializeWorkspaceSettingsParams{WorkspaceID: workspaceID},
	)
	if err != nil {
		return fmt.Errorf("initialize workspace settings: %w", err)
	}
	if rows != 1 {
		return fmt.Errorf("initialize workspace settings: affected %d rows, want 1", rows)
	}
	return nil
}
