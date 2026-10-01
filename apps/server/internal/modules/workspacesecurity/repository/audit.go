package workspacesecurityrepository

import (
	"context"

	domain "github.com/complexus-tech/projects-api/internal/modules/workspacesecurity/domain"
	sql "github.com/complexus-tech/projects-api/internal/modules/workspacesecurity/repository/sqlc"
	"github.com/google/uuid"
)

func (r *Repository) Audit(ctx context.Context, scope domain.Scope, filter domain.AuditFilter) ([]domain.AuditEvent, error) {
	result := []domain.AuditEvent{}
	err := r.within(ctx, false, func(q *sql.Queries) error {
		if _, err := admin(ctx, q, scope, false); err != nil {
			return err
		}
		input := sql.ListAuditParams{WorkspaceID: scope.WorkspaceID, ActorID: filter.ActorID, ResourceType: filter.ResourceType, ResourceID: filter.ResourceID, FromTime: filter.From, ToTime: filter.To, MaxRows: filter.Limit}
		if filter.Before != nil {
			input.BeforeTime = &filter.Before.CreatedAt
			input.BeforeID = filter.Before.ID
			input.BeforeSource = filter.Before.Source
		}
		rows, err := q.ListAudit(ctx, input)
		if err != nil {
			return err
		}
		for _, row := range rows {
			event := domain.AuditEvent{ID: row.ID, Source: row.Source, ActorType: row.ActorType, ResourceType: row.ResourceType, Operation: row.Operation, Metadata: row.Metadata, CreatedAt: row.CreatedAt}
			if row.ActorID != uuid.Nil {
				event.ActorID = &row.ActorID
			}
			if row.ResourceID != uuid.Nil {
				event.ResourceID = &row.ResourceID
			}
			result = append(result, event)
		}
		return nil
	})
	return result, err
}
func (r *Repository) recordExport(ctx context.Context, scope domain.Scope, operation string, count int) error {
	return r.within(ctx, true, func(q *sql.Queries) error {
		if _, err := admin(ctx, q, scope, true); err != nil {
			return err
		}
		return appendAudit(ctx, q, scope, "workspace", scope.WorkspaceID, operation, map[string]any{"count": count})
	})
}
func (r *Repository) RecordExport(ctx context.Context, scope domain.Scope, count int) error {
	return r.recordExport(ctx, scope, "workspace.data_exported", count)
}
func (r *Repository) RecordAuditExport(ctx context.Context, scope domain.Scope, count int) error {
	return r.recordExport(ctx, scope, "workspace.audit_exported", count)
}
