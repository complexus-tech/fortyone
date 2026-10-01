package customfieldsrepository

import (
	"context"
	"fmt"
	"time"

	domain "github.com/complexus-tech/projects-api/internal/modules/customfields/domain"
	sql "github.com/complexus-tech/projects-api/internal/modules/customfields/repository/sqlc"
	"github.com/google/uuid"
)

func (r *Repository) Report(ctx context.Context, scope domain.Scope, input domain.ReportInput) (domain.Report, error) {
	var result domain.Report
	err := r.within(ctx, false, func(q *sql.Queries) error {
		row, err := q.FindField(ctx, sql.FindFieldParams{FieldID: input.FieldID, WorkspaceID: scope.WorkspaceID})
		if err != nil {
			return err
		}
		scope.TeamID = row.TeamID
		// The current actor/team boundary is established before any aggregate is
		// evaluated, including when a caller supplies an existing foreign field ID.
		if err := authorize(ctx, q, scope, false, false); err != nil {
			return err
		}
		field := fieldFromRow(sql.ListFieldsRow(row))
		if input.Aggregation != "count" && field.Type != domain.Number && field.Type != domain.Money {
			return fmt.Errorf("%w: numeric aggregation requires a number or money field", domain.ErrInvalid)
		}
		fields, err := listFields(ctx, q, scope)
		if err != nil {
			return err
		}
		for _, candidate := range fields {
			if candidate.ID == field.ID {
				field.Options = candidate.Options
				break
			}
		}
		var dateFieldID *uuid.UUID
		if input.DateBasis != "created" && input.DateBasis != "completed" {
			id, err := uuid.Parse(input.DateBasis)
			if err != nil {
				return domain.ErrInvalid
			}
			valid := false
			for _, candidate := range fields {
				if candidate.ID == id && candidate.Type == domain.Date {
					valid = true
					break
				}
			}
			if !valid {
				return fmt.Errorf("%w: date basis must be a date field in this team", domain.ErrInvalid)
			}
			dateFieldID = &id
		}
		params := sql.AggregateCustomFieldParams{Aggregation: input.Aggregation, DateBasis: input.DateBasis, FieldID: input.FieldID, DateFieldID: dateFieldID, WorkspaceID: scope.WorkspaceID, TeamID: scope.TeamID, StatusIds: input.StatusIDs, AssigneeIds: input.AssigneeIDs, GroupBy: input.GroupBy}
		if params.StatusIds == nil {
			params.StatusIds = []uuid.UUID{}
		}
		if params.AssigneeIds == nil {
			params.AssigneeIds = []uuid.UUID{}
		}
		if input.StartDate != nil {
			date, err := time.Parse(time.DateOnly, *input.StartDate)
			if err != nil {
				return domain.ErrInvalid
			}
			params.StartDate = &date
		}
		if input.EndDate != nil {
			date, err := time.Parse(time.DateOnly, *input.EndDate)
			if err != nil {
				return domain.ErrInvalid
			}
			params.EndDate = &date
		}
		rows, err := q.AggregateCustomField(ctx, params)
		if err != nil {
			return err
		}
		if len(rows) > 500 {
			return fmt.Errorf("%w: narrow the report date range", domain.ErrLimit)
		}
		result = domain.Report{Field: field, Aggregation: input.Aggregation, GroupBy: input.GroupBy, Currency: field.Currency, Rows: []domain.ReportRow{}}
		if input.Aggregation == "count" {
			result.Currency = nil
		}
		for _, row := range rows {
			result.TotalCount += row.TotalCount
			result.ValuedCount += row.ValuedCount
			result.Rows = append(result.Rows, domain.ReportRow{Key: row.GroupKey, Label: row.Label, Value: row.Result, Count: row.ValuedCount})
		}
		result.MissingCount = result.TotalCount - result.ValuedCount
		return nil
	})
	return result, err
}
