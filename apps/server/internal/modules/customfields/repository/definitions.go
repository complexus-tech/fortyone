package customfieldsrepository

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	domain "github.com/complexus-tech/projects-api/internal/modules/customfields/domain"
	sql "github.com/complexus-tech/projects-api/internal/modules/customfields/repository/sqlc"
	"github.com/google/uuid"
)

func (r *Repository) Create(ctx context.Context, scope domain.Scope, input domain.Definition) (domain.Field, error) {
	var result domain.Field
	err := r.within(ctx, true, func(q *sql.Queries) error {
		if err := authorize(ctx, q, scope, true, true); err != nil {
			return err
		}
		if err := domain.ValidateDefinition(input, nil); err != nil {
			return err
		}
		counts, err := q.CountFields(ctx, sql.CountFieldsParams{WorkspaceID: scope.WorkspaceID, TeamID: scope.TeamID})
		if err != nil {
			return err
		}
		if counts.Total >= domain.MaxFields || counts.Active >= domain.MaxActiveFields {
			return domain.ErrLimit
		}
		row, err := q.InsertField(ctx, sql.InsertFieldParams{ID: uuid.New(), WorkspaceID: scope.WorkspaceID, TeamID: scope.TeamID, Name: input.Name, FieldType: string(input.Type), Currency: input.Currency, Icon: input.Icon, ShowOnCreate: input.ShowOnCreate})
		if err != nil {
			return err
		}
		result = fieldFromRow(sql.ListFieldsRow(row))
		if err := replaceOptions(ctx, q, &result, input.Options); err != nil {
			return err
		}
		return definitionAudit(ctx, q, scope, result, "custom_field.created", nil)
	})
	return result, err
}

func (r *Repository) Update(ctx context.Context, scope domain.Scope, id uuid.UUID, input domain.Definition) (domain.Field, error) {
	var result domain.Field
	err := r.within(ctx, true, func(q *sql.Queries) error {
		if err := authorize(ctx, q, scope, true, true); err != nil {
			return err
		}
		row, err := q.LockField(ctx, sql.LockFieldParams{FieldID: id, WorkspaceID: scope.WorkspaceID, TeamID: scope.TeamID})
		if err != nil {
			return err
		}
		result = fieldFromRow(sql.ListFieldsRow(row))
		fields, err := listFields(ctx, q, scope)
		if err != nil {
			return err
		}
		for _, field := range fields {
			if field.ID == id {
				result.Options = field.Options
				break
			}
		}
		if err := domain.ValidateDefinition(input, &result); err != nil {
			return err
		}
		if input.ExpectedUpdatedAt != nil && !input.ExpectedUpdatedAt.Equal(result.UpdatedAt) {
			return domain.ErrConflict
		}
		oldIcon := result.Icon
		updated, err := q.UpdateField(ctx, sql.UpdateFieldParams{FieldID: id, WorkspaceID: scope.WorkspaceID, TeamID: scope.TeamID, Name: input.Name, ShowOnCreate: input.ShowOnCreate, Icon: input.Icon, IconSet: input.IconSet})
		if err != nil {
			return err
		}
		if err := replaceOptions(ctx, q, &result, input.Options); err != nil {
			return err
		}
		result.Name, result.ShowOnCreate, result.UpdatedAt = updated.Name, updated.ShowOnCreate, updated.UpdatedAt
		result.Icon = updated.Icon
		return definitionAudit(ctx, q, scope, result, "custom_field.updated", oldIcon)
	})
	return result, err
}

func replaceOptions(ctx context.Context, q *sql.Queries, field *domain.Field, inputs []domain.OptionInput) error {
	known := map[uuid.UUID]domain.Option{}
	for _, option := range field.Options {
		known[option.ID] = option
	}
	newCount := 0
	for _, input := range inputs {
		if input.ID == uuid.Nil {
			newCount++
			continue
		}
		if _, ok := known[input.ID]; !ok {
			return fmt.Errorf("%w: option belongs to another field", domain.ErrInvalid)
		}
	}
	// Retained identities are bounded independently from active choices.
	if len(known)+newCount > 500 {
		return domain.ErrLimit
	}
	if err := q.ArchiveOptions(ctx, sql.ArchiveOptionsParams{FieldID: field.ID}); err != nil {
		return err
	}
	for position, input := range inputs {
		id := input.ID
		if id == uuid.Nil {
			id = uuid.New()
		}
		name := strings.TrimSpace(input.Name)
		if err := q.UpsertOption(ctx, sql.UpsertOptionParams{ID: id, FieldID: field.ID, Name: name, Position: int32(position)}); err != nil {
			return err
		}
	}
	// Read committed timestamps from the database so archived identities remain
	// available to render old values without becoming selectable again.
	rows, err := q.ListOptionsForField(ctx, sql.ListOptionsForFieldParams{FieldID: field.ID})
	if err != nil {
		return err
	}
	field.Options = make([]domain.Option, 0, len(rows))
	for _, row := range rows {
		field.Options = append(field.Options, domain.Option{ID: row.ID, Name: row.Name, ArchivedAt: row.ArchivedAt})
	}
	return nil
}

func (r *Repository) Archive(ctx context.Context, scope domain.Scope, id uuid.UUID) error {
	return r.within(ctx, true, func(q *sql.Queries) error {
		if err := authorize(ctx, q, scope, true, true); err != nil {
			return err
		}
		row, err := q.LockField(ctx, sql.LockFieldParams{FieldID: id, WorkspaceID: scope.WorkspaceID, TeamID: scope.TeamID})
		if err != nil {
			return err
		}
		if row.ArchivedAt != nil {
			return nil
		}
		count, err := q.ArchiveField(ctx, sql.ArchiveFieldParams{FieldID: id, WorkspaceID: scope.WorkspaceID, TeamID: scope.TeamID})
		if err != nil {
			return err
		}
		if count == 0 {
			return domain.ErrNotFound
		}
		field := fieldFromRow(sql.ListFieldsRow(row))
		return definitionAudit(ctx, q, scope, field, "custom_field.archived", field.Icon)
	})
}

func definitionAudit(ctx context.Context, q *sql.Queries, scope domain.Scope, field domain.Field, operation string, oldIcon *string) error {
	metadata, err := json.Marshal(map[string]any{"teamId": scope.TeamID, "oldIcon": oldIcon, "newIcon": field.Icon})
	if err != nil {
		return err
	}
	return q.AppendDefinitionAudit(ctx, sql.AppendDefinitionAuditParams{ID: uuid.New(), WorkspaceID: scope.WorkspaceID, FieldID: field.ID, ActorID: scope.ActorID, Operation: operation, Metadata: metadata})
}
