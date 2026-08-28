package adminrepository

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	admindomain "github.com/complexus-tech/projects-api/internal/modules/admin/domain"
	adminsql "github.com/complexus-tech/projects-api/internal/modules/admin/repository/sqlc"
	"github.com/google/uuid"
)

type auditEntry struct {
	ActorID     uuid.UUID
	TargetType  admindomain.TargetType
	TargetID    *uuid.UUID
	WorkspaceID *uuid.UUID
	Action      admindomain.AuditAction
	FieldName   string
	OldValue    any
	NewValue    any
	Reason      string
	Metadata    map[string]any
}

func (repository *Repository) ListAuditLogs(
	ctx context.Context,
	query admindomain.ListAuditLogsQuery,
) (admindomain.ListResult[admindomain.AuditLog], error) {
	if _, err := admindomain.ParseTargetType(string(query.TargetType)); err != nil {
		return admindomain.ListResult[admindomain.AuditLog]{}, err
	}
	if _, err := admindomain.ParseAuditAction(string(query.Action)); err != nil {
		return admindomain.ListResult[admindomain.AuditLog]{}, err
	}
	if query.From != nil && query.To != nil && query.From.After(*query.To) {
		return admindomain.ListResult[admindomain.AuditLog]{}, admindomain.ErrInvalidFilter
	}
	page, err := newSQLPage(query.Page)
	if err != nil {
		return admindomain.ListResult[admindomain.AuditLog]{}, err
	}

	params := adminsql.ListAdminAuditLogsParams{
		WorkspaceIDSet:   query.WorkspaceID != nil,
		TargetTypeFilter: string(query.TargetType), ActionFilter: string(query.Action),
		SearchText: query.Search, ActorSearch: query.ActorSearch,
		FromSet: query.From != nil, ToSet: query.To != nil,
		RowLimit: page.limit, RowOffset: page.offset,
	}
	if query.WorkspaceID != nil {
		params.WorkspaceID = *query.WorkspaceID
	}
	if query.From != nil {
		params.FromAt = *query.From
	}
	if query.To != nil {
		params.ToAt = *query.To
	}

	var result admindomain.ListResult[admindomain.AuditLog]
	err = repository.withActiveInternalAdmin(ctx, query.ActorID, func(queries adminsql.Querier) error {
		rows, err := queries.ListAdminAuditLogs(ctx, params)
		if err != nil {
			return fmt.Errorf("list admin audit logs: %w", err)
		}
		items := make([]admindomain.AuditLog, 0, len(rows))
		for _, row := range rows {
			item, err := auditLogFromRow(row)
			if err != nil {
				return err
			}
			items = append(items, item)
		}
		total := int64(0)
		if len(rows) > 0 {
			total = rows[0].TotalCount
		}
		pagination, err := paginationResult(query.Page, total)
		if err != nil {
			return err
		}
		result = admindomain.ListResult[admindomain.AuditLog]{Items: items, Pagination: pagination}
		return nil
	})
	return result, err
}

func (repository *Repository) ListAdminNotes(
	ctx context.Context,
	query admindomain.ListAdminNotesQuery,
) (admindomain.ListResult[admindomain.AdminNote], error) {
	if query.TargetType != admindomain.TargetAny && query.TargetType != admindomain.TargetWorkspace &&
		query.TargetType != admindomain.TargetUser {
		return admindomain.ListResult[admindomain.AdminNote]{}, admindomain.ErrInvalidFilter
	}
	page, err := newSQLPage(query.Page)
	if err != nil {
		return admindomain.ListResult[admindomain.AdminNote]{}, err
	}
	params := adminsql.ListAdminNotesParams{
		TargetTypeFilter: string(query.TargetType), TargetIDSet: query.TargetID != nil,
		WorkspaceIDSet: query.WorkspaceID != nil, RowLimit: page.limit, RowOffset: page.offset,
	}
	if query.TargetID != nil {
		params.TargetID = *query.TargetID
	}
	if query.WorkspaceID != nil {
		params.WorkspaceID = *query.WorkspaceID
	}

	var result admindomain.ListResult[admindomain.AdminNote]
	err = repository.withActiveInternalAdmin(ctx, query.ActorID, func(queries adminsql.Querier) error {
		rows, err := queries.ListAdminNotes(ctx, params)
		if err != nil {
			return fmt.Errorf("list admin notes: %w", err)
		}
		items := make([]admindomain.AdminNote, 0, len(rows))
		for _, row := range rows {
			items = append(items, adminNoteFromListRow(row))
		}
		total := int64(0)
		if len(rows) > 0 {
			total = rows[0].TotalCount
		}
		pagination, err := paginationResult(query.Page, total)
		if err != nil {
			return err
		}
		result = admindomain.ListResult[admindomain.AdminNote]{Items: items, Pagination: pagination}
		return nil
	})
	return result, err
}

func insertAuditLog(
	ctx context.Context,
	queries adminsql.Querier,
	entry auditEntry,
) (adminsql.InsertAdminAuditLogRow, error) {
	oldValue, err := json.Marshal(entry.OldValue)
	if err != nil {
		return adminsql.InsertAdminAuditLogRow{}, fmt.Errorf("marshal admin audit old value: %w", err)
	}
	newValue, err := json.Marshal(entry.NewValue)
	if err != nil {
		return adminsql.InsertAdminAuditLogRow{}, fmt.Errorf("marshal admin audit new value: %w", err)
	}
	metadata := entry.Metadata
	if metadata == nil {
		metadata = map[string]any{}
	}
	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return adminsql.InsertAdminAuditLogRow{}, fmt.Errorf("marshal admin audit metadata: %w", err)
	}
	row, err := queries.InsertAdminAuditLog(ctx, adminsql.InsertAdminAuditLogParams{
		ActorUserID: entry.ActorID, TargetType: string(entry.TargetType),
		TargetID: entry.TargetID, WorkspaceID: entry.WorkspaceID,
		Action: string(entry.Action), FieldName: optionalString(entry.FieldName),
		OldValue: oldValue, NewValue: newValue, Reason: optionalString(entry.Reason),
		Metadata: metadataJSON,
	})
	if err != nil {
		return adminsql.InsertAdminAuditLogRow{}, fmt.Errorf("insert admin audit log: %w", err)
	}
	return row, nil
}

func auditLogFromRow(row adminsql.ListAdminAuditLogsRow) (admindomain.AuditLog, error) {
	oldValue, err := decodeJSON(row.OldValue)
	if err != nil {
		return admindomain.AuditLog{}, fmt.Errorf("decode admin audit %s old value: %w", row.ID, err)
	}
	newValue, err := decodeJSON(row.NewValue)
	if err != nil {
		return admindomain.AuditLog{}, fmt.Errorf("decode admin audit %s new value: %w", row.ID, err)
	}
	metadata, err := decodeJSON(row.Metadata)
	if err != nil {
		return admindomain.AuditLog{}, fmt.Errorf("decode admin audit %s metadata: %w", row.ID, err)
	}
	return admindomain.AuditLog{
		ID: row.ID, ActorUserID: row.ActorUserID, ActorEmail: row.ActorEmail,
		ActorName: stringValue(row.ActorName), TargetType: row.TargetType,
		TargetID: row.TargetID, WorkspaceID: row.WorkspaceID,
		WorkspaceName: row.WorkspaceName, WorkspaceSlug: row.WorkspaceSlug,
		Action: row.Action, FieldName: stringValue(row.FieldName), OldValue: oldValue,
		NewValue: newValue, Reason: stringValue(row.Reason), Metadata: metadata,
		CreatedAt: row.CreatedAt,
	}, nil
}

func adminNoteFromListRow(row adminsql.ListAdminNotesRow) admindomain.AdminNote {
	return admindomain.AdminNote{
		ID: row.ID, TargetType: row.TargetType, TargetID: row.TargetID,
		WorkspaceID: row.WorkspaceID, Body: row.Body, CreatedByUserID: row.CreatedByUserID,
		CreatedByName: stringValue(row.CreatedByName), CreatedByEmail: row.CreatedByEmail,
		CreatedAt: row.CreatedAt,
	}
}

func decodeJSON(raw []byte) (any, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	var value any
	if err := json.Unmarshal(raw, &value); err != nil {
		return nil, err
	}
	return value, nil
}

func optionalString(value string) *string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}
