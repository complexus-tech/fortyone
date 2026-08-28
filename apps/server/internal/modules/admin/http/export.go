package adminhttp

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"net/http"
	"time"

	admin "github.com/complexus-tech/projects-api/internal/modules/admin/service"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

func (h *Handlers) ExportAuditLogs(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	filters, err := parseAuditLogQuery(r)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	result, err := h.admin.ListAuditLogs(ctx, userID, admin.ListAuditLogsInput{
		Pagination:  admin.PaginationInput{Page: 1, Limit: maxPageSize},
		WorkspaceID: filters.WorkspaceID,
		TargetType:  filters.TargetType,
		Query:       filters.Query,
		Action:      filters.Action,
		ActorQuery:  filters.ActorQuery,
		From:        filters.From,
		To:          filters.To,
	})
	if err != nil {
		return h.respondAdminError(ctx, w, "export_audit_logs", err)
	}

	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="admin-audit-logs.csv"`)
	writer := csv.NewWriter(w)
	if err := writer.Write([]string{"actor", "target_type", "target_id", "workspace", "action", "field", "old_value", "new_value", "reason", "created_at"}); err != nil {
		return err
	}
	for _, entry := range result.Items {
		if err := writer.Write(auditCSVRow(entry)); err != nil {
			return err
		}
	}
	writer.Flush()
	return writer.Error()
}

func auditCSVRow(entry admin.AuditLog) []string {
	return []string{
		entry.ActorName, entry.TargetType, uuidString(entry.TargetID), stringValue(entry.WorkspaceName),
		entry.Action, entry.FieldName, csvValue(entry.OldValue), csvValue(entry.NewValue),
		entry.Reason, entry.CreatedAt.Format(time.RFC3339),
	}
}

func uuidString(value *uuid.UUID) string {
	if value == nil {
		return ""
	}
	return value.String()
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func csvValue(value any) string {
	if value == nil {
		return ""
	}
	raw, err := json.Marshal(value)
	if err != nil {
		return ""
	}
	return string(raw)
}
