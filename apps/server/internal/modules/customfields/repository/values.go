package customfieldsrepository

import (
	"context"
	"errors"
	"fmt"

	domain "github.com/complexus-tech/projects-api/internal/modules/customfields/domain"
	sql "github.com/complexus-tech/projects-api/internal/modules/customfields/repository/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func snapshot(ctx context.Context, q *sql.Queries, scope domain.Scope, storyID uuid.UUID, version int64) (domain.Snapshot, error) {
	fields, err := listFields(ctx, q, scope)
	if err != nil {
		return domain.Snapshot{}, err
	}
	rows, err := q.ListStoryValues(ctx, sql.ListStoryValuesParams{StoryID: storyID, WorkspaceID: scope.WorkspaceID, TeamID: scope.TeamID})
	if err != nil {
		return domain.Snapshot{}, err
	}
	result := domain.Snapshot{Fields: []domain.Field{}, Values: []domain.Value{}, Version: version}
	retained := map[uuid.UUID]bool{}
	for _, row := range rows {
		result.Values = append(result.Values, domain.Value{FieldID: row.FieldID, Value: row.Value})
		retained[row.FieldID] = row.Value != nil
	}
	for _, field := range fields {
		if field.ArchivedAt == nil || retained[field.ID] {
			result.Fields = append(result.Fields, field)
		}
	}
	return result, nil
}

func (r *Repository) Snapshot(ctx context.Context, scope domain.Scope, storyID uuid.UUID) (domain.Snapshot, error) {
	var result domain.Snapshot
	err := r.within(ctx, false, func(q *sql.Queries) error {
		story, err := q.GetStoryScope(ctx, sql.GetStoryScopeParams{StoryID: storyID, WorkspaceID: scope.WorkspaceID})
		if err != nil {
			return err
		}
		scope.TeamID = story.TeamID
		if err := authorize(ctx, q, scope, false, false); err != nil {
			return err
		}
		result, err = snapshot(ctx, q, scope, storyID, story.CustomFieldsVersion)
		return err
	})
	return result, err
}

func (r *Repository) Patch(ctx context.Context, scope domain.Scope, storyID uuid.UUID, patch domain.ValuePatch) (domain.Snapshot, error) {
	var result domain.Snapshot
	err := r.within(ctx, true, func(q *sql.Queries) error {
		story, err := q.GetStoryScope(ctx, sql.GetStoryScopeParams{StoryID: storyID, WorkspaceID: scope.WorkspaceID})
		if err != nil {
			return err
		}
		scope.TeamID = story.TeamID
		if err := authorize(ctx, q, scope, true, false); err != nil {
			return err
		}
		locked, err := q.LockStoryScope(ctx, sql.LockStoryScopeParams{StoryID: storyID, WorkspaceID: scope.WorkspaceID})
		if err != nil {
			return err
		}
		if locked.TeamID != scope.TeamID || locked.ArchivedAt != nil {
			return domain.ErrConflict
		}
		if patch.ExpectedVersion != nil && *patch.ExpectedVersion != locked.CustomFieldsVersion {
			return domain.ErrConflict
		}
		version, err := applyValues(ctx, q, scope, storyID, patch.Values, locked.CustomFieldsVersion)
		if err != nil {
			return err
		}
		result, err = snapshot(ctx, q, scope, storyID, version)
		return err
	})
	return result, err
}

func applyValues(ctx context.Context, q *sql.Queries, scope domain.Scope, storyID uuid.UUID, values []domain.Value, version int64) (int64, error) {
	if err := domain.ValidatePatch(domain.ValuePatch{Values: values}); err != nil {
		return 0, err
	}
	fields, err := listFields(ctx, q, scope)
	if err != nil {
		return 0, err
	}
	byID := map[uuid.UUID]domain.Field{}
	for _, field := range fields {
		byID[field.ID] = field
	}
	changed := false
	for _, patch := range values {
		field, ok := byID[patch.FieldID]
		if !ok || field.ArchivedAt != nil {
			return 0, fmt.Errorf("%w: field is unavailable in this team", domain.ErrInvalid)
		}
		old, err := q.GetStoryFieldValue(ctx, sql.GetStoryFieldValueParams{StoryID: storyID, FieldID: field.ID, WorkspaceID: scope.WorkspaceID, TeamID: scope.TeamID})
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return 0, err
		}
		if sameValue(old, patch.Value) {
			continue
		}
		normalized, err := domain.NormalizeValue(field, patch.Value)
		if err != nil {
			return 0, err
		}
		if field.Type == domain.Person && normalized != nil {
			personID, parseErr := uuid.Parse(*normalized)
			if parseErr != nil {
				return 0, domain.ErrInvalid
			}
			if _, err := q.ValidatePerson(ctx, sql.ValidatePersonParams{WorkspaceID: scope.WorkspaceID, TeamID: scope.TeamID, PersonID: personID}); err != nil {
				if errors.Is(err, pgx.ErrNoRows) {
					return 0, fmt.Errorf("%w: person is not a current team member", domain.ErrInvalid)
				}
				return 0, err
			}
		}
		if !changed {
			version, err = q.AdvanceStoryCustomFieldVersion(ctx, sql.AdvanceStoryCustomFieldVersionParams{StoryID: storyID, WorkspaceID: scope.WorkspaceID, TeamID: scope.TeamID})
			if err != nil {
				return 0, err
			}
			changed = true
		}
		if err := q.UpsertStoryValue(ctx, sql.UpsertStoryValueParams{StoryID: storyID, FieldID: field.ID, WorkspaceID: scope.WorkspaceID, TeamID: scope.TeamID, FieldType: string(field.Type), Value: normalized}); err != nil {
			return 0, err
		}
		if err := q.AppendValueAudit(ctx, sql.AppendValueAuditParams{StoryID: storyID, FieldID: field.ID, WorkspaceID: scope.WorkspaceID, ActorID: &scope.ActorID, OldValue: old, NewValue: normalized, Version: version}); err != nil {
			return 0, err
		}
	}
	return version, nil
}

func sameValue(a, b *string) bool {
	return (a == nil && b == nil) || (a != nil && b != nil && *a == *b)
}
