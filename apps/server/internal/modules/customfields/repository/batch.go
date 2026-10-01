package customfieldsrepository

import (
	"context"

	domain "github.com/complexus-tech/projects-api/internal/modules/customfields/domain"
	sql "github.com/complexus-tech/projects-api/internal/modules/customfields/repository/sqlc"
	"github.com/google/uuid"
)

func (r *Repository) Batch(ctx context.Context, scope domain.Scope, storyIDs []uuid.UUID) (domain.BatchSnapshot, error) {
	result := domain.BatchSnapshot{Items: []domain.BatchStory{}}
	err := r.within(ctx, false, func(q *sql.Queries) error {
		rows, err := q.ListBatchStoryValues(ctx, sql.ListBatchStoryValuesParams{ActorID: scope.ActorID, WorkspaceID: scope.WorkspaceID, StoryIds: storyIDs})
		if err != nil {
			return err
		}
		indexes := make(map[uuid.UUID]int, len(storyIDs))
		for _, row := range rows {
			index, ok := indexes[row.StoryID]
			if !ok {
				index = len(result.Items)
				indexes[row.StoryID] = index
				result.Items = append(result.Items, domain.BatchStory{StoryID: row.StoryID, Version: row.CustomFieldsVersion, Values: []domain.Value{}})
			}
			if row.FieldID != nil {
				result.Items[index].Values = append(result.Items[index].Values, domain.Value{FieldID: *row.FieldID, Value: row.Value})
			}
		}
		return nil
	})
	return result, err
}
