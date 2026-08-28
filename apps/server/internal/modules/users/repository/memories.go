package usersrepository

import (
	"context"
	"errors"
	"fmt"

	usersdomain "github.com/complexus-tech/projects-api/internal/modules/users/domain"
	usersql "github.com/complexus-tech/projects-api/internal/modules/users/repository/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (r *repo) AddUserMemory(
	ctx context.Context,
	memory usersdomain.NewUserMemoryItem,
) (usersdomain.UserMemoryItem, error) {
	row, err := r.queries.CreateUserMemoryForMember(ctx, usersql.CreateUserMemoryForMemberParams{
		Content:     memory.Content,
		WorkspaceID: memory.WorkspaceID,
		UserID:      memory.UserID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return usersdomain.UserMemoryItem{}, usersdomain.ErrMemoryNotFound
		}
		return usersdomain.UserMemoryItem{}, fmt.Errorf("create scoped user memory: %w", err)
	}
	return mapCreatedMemory(row), nil
}

func (r *repo) UpdateUserMemory(
	ctx context.Context,
	id uuid.UUID,
	scope usersdomain.UserMemoryScope,
	update usersdomain.UpdateUserMemoryItem,
) error {
	if update.Content == nil {
		return errors.New("user memory content is required")
	}
	rows, err := r.queries.UpdateUserMemoryForOwner(ctx, usersql.UpdateUserMemoryForOwnerParams{
		Content:     *update.Content,
		MemoryID:    id,
		UserID:      scope.UserID,
		WorkspaceID: scope.WorkspaceID,
	})
	if err != nil {
		return fmt.Errorf("update scoped user memory: %w", err)
	}
	return validateUserMemoryMutation(rows)
}

func (r *repo) DeleteUserMemory(
	ctx context.Context,
	id uuid.UUID,
	scope usersdomain.UserMemoryScope,
) error {
	rows, err := r.queries.DeleteUserMemoryForOwner(ctx, usersql.DeleteUserMemoryForOwnerParams{
		MemoryID:    id,
		UserID:      scope.UserID,
		WorkspaceID: scope.WorkspaceID,
	})
	if err != nil {
		return fmt.Errorf("delete scoped user memory: %w", err)
	}
	return validateUserMemoryMutation(rows)
}

func (r *repo) ListUserMemories(
	ctx context.Context,
	userID uuid.UUID,
	workspaceID uuid.UUID,
) ([]usersdomain.UserMemoryItem, error) {
	rows, err := r.queries.ListUserMemoriesForOwner(ctx, usersql.ListUserMemoriesForOwnerParams{
		UserID:      userID,
		WorkspaceID: workspaceID,
	})
	if err != nil {
		return nil, fmt.Errorf("list scoped user memories: %w", err)
	}
	return mapUserMemories(rows), nil
}

func validateUserMemoryMutation(rowsAffected int64) error {
	if rowsAffected == 0 {
		return usersdomain.ErrMemoryNotFound
	}
	return nil
}
