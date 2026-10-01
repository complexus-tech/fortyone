package workspacesecurityhttp

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	domain "github.com/complexus-tech/projects-api/internal/modules/workspacesecurity/domain"
	"github.com/google/uuid"
)

type auditCursor struct {
	WorkspaceID, ActorID uuid.UUID
	Filter               string
	Position             domain.Position
	ExpiresAt            time.Time
}

func auditFilter(r *http.Request, export bool) (domain.AuditFilter, string, error) {
	filter := domain.AuditFilter{Limit: 51}
	query := r.URL.Query()
	for key, values := range query {
		valid := key == "actorId" || key == "resourceType" || key == "resourceId" || key == "from" || key == "to" || (!export && (key == "limit" || key == "cursor"))
		if !valid || len(values) != 1 {
			return filter, "", domain.ErrInvalid
		}
	}
	for key, target := range map[string]**uuid.UUID{"actorId": &filter.ActorID, "resourceId": &filter.ResourceID} {
		if raw := query.Get(key); raw != "" {
			id, err := uuid.Parse(raw)
			if err != nil || id == uuid.Nil {
				return filter, "", domain.ErrInvalid
			}
			*target = &id
		}
	}
	for key, target := range map[string]**time.Time{"from": &filter.From, "to": &filter.To} {
		if raw := query.Get(key); raw != "" {
			date, err := time.Parse(time.RFC3339Nano, raw)
			if err != nil {
				return filter, "", domain.ErrInvalid
			}
			date = date.UTC()
			*target = &date
		}
	}
	filter.ResourceType = query.Get("resourceType")
	if export {
		filter.Limit = 10001
	} else if query.Has("limit") {
		limit, err := strconv.Atoi(query.Get("limit"))
		if err != nil || limit < 1 || limit > 100 {
			return filter, "", domain.ErrInvalid
		}
		filter.Limit = int32(limit + 1)
	}
	fingerprint, err := json.Marshal(struct {
		ActorID, ResourceID *uuid.UUID
		ResourceType        string
		From, To            *time.Time
	}{filter.ActorID, filter.ResourceID, filter.ResourceType, filter.From, filter.To})
	if err != nil {
		return filter, "", err
	}
	digest := sha256.Sum256(fingerprint)
	return filter, hex.EncodeToString(digest[:]), nil
}
func (h *Handlers) Audit(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	scope, err := identities(ctx)
	if err != nil {
		return respond(ctx, w, nil, err)
	}
	filter, fingerprint, err := auditFilter(r, false)
	if err != nil {
		return respond(ctx, w, nil, err)
	}
	if raw := r.URL.Query().Get("cursor"); raw != "" {
		cursor, err := h.cursors.Decode(raw)
		if err != nil || cursor.WorkspaceID != scope.WorkspaceID || cursor.ActorID != scope.ActorID || cursor.Filter != fingerprint || !cursor.ExpiresAt.After(time.Now()) {
			return respond(ctx, w, nil, domain.ErrInvalid)
		}
		filter.Before = &cursor.Position
	}
	events, err := h.service.Audit(ctx, scope, filter)
	if err != nil {
		return respond(ctx, w, nil, err)
	}
	next := ""
	limit := int(filter.Limit) - 1
	if len(events) > limit {
		events = events[:limit]
		last := events[len(events)-1]
		next, err = h.cursors.Encode(auditCursor{WorkspaceID: scope.WorkspaceID, ActorID: scope.ActorID, Filter: fingerprint, Position: domain.Position{ID: last.ID, Source: last.Source, CreatedAt: last.CreatedAt}, ExpiresAt: time.Now().Add(24 * time.Hour)})
		if err != nil {
			return respond(ctx, w, nil, err)
		}
	}
	return respond(ctx, w, struct {
		Items      []domain.AuditEvent `json:"items"`
		NextCursor string              `json:"nextCursor"`
	}{events, next}, nil)
}
func (h *Handlers) ExportAudit(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	scope, err := identities(ctx)
	if err != nil {
		return respond(ctx, w, nil, err)
	}
	filter, _, err := auditFilter(r, true)
	if err != nil {
		return respond(ctx, w, nil, err)
	}
	events, err := h.service.Audit(ctx, scope, filter)
	if err != nil {
		return respond(ctx, w, nil, err)
	}
	if len(events) > 10000 {
		return respond(ctx, w, nil, domain.ErrLimit)
	}
	var body bytes.Buffer
	writer := csv.NewWriter(&body)
	if err := writer.Write([]string{"event_id", "created_at", "source", "actor_type", "actor_id", "resource_type", "resource_id", "operation", "metadata"}); err != nil {
		return respond(ctx, w, nil, err)
	}
	for _, event := range events {
		actor, resource := "", ""
		if event.ActorID != nil {
			actor = event.ActorID.String()
		}
		if event.ResourceID != nil {
			resource = event.ResourceID.String()
		}
		if err := writer.Write([]string{event.ID.String(), event.CreatedAt.UTC().Format(time.RFC3339Nano), event.Source, event.ActorType, actor, event.ResourceType, resource, event.Operation, string(event.Metadata)}); err != nil {
			return respond(ctx, w, nil, err)
		}
		if body.Len() > 20<<20 {
			return respond(ctx, w, nil, domain.ErrLimit)
		}
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		return respond(ctx, w, nil, err)
	}
	if body.Len() > 20<<20 {
		return respond(ctx, w, nil, domain.ErrLimit)
	}
	if err := h.service.RecordAuditExport(ctx, scope, len(events)); err != nil {
		return respond(ctx, w, nil, err)
	}
	w.Header().Set("Cache-Control", "private, no-store")
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="workspace-audit.csv"`)
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(body.Bytes())
	return err
}
