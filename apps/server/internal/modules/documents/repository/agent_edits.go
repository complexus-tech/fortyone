package documentsrepository

import (
	"context"
	"encoding/json"
	"errors"

	domain "github.com/complexus-tech/projects-api/internal/modules/documents/domain"
	queries "github.com/complexus-tech/projects-api/internal/modules/documents/repository/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (r *Repository) EnqueueAgentEdit(ctx context.Context, input domain.AgentEditInput) (domain.AgentEdit, error) {
	doc, err := r.Get(ctx, input.WorkspaceID, input.RequesterID, input.DocumentID)
	if err != nil {
		return domain.AgentEdit{}, err
	}
	if !doc.CanEdit {
		return domain.AgentEdit{}, domain.ErrForbidden
	}
	q := r.commentQueries(ctx)
	row, err := q.EnqueueDocumentAgentEdit(ctx, queries.EnqueueDocumentAgentEditParams{InvocationID: input.InvocationID, WorkspaceID: input.WorkspaceID, DocumentID: input.DocumentID, RequesterID: input.RequesterID, ExpectedRevision: input.ExpectedRevision, ExpectedEpoch: input.ExpectedEpoch, Operations: input.Operations})
	if errors.Is(err, pgx.ErrNoRows) {
		row, err = q.GetDocumentAgentEdit(ctx, queries.GetDocumentAgentEditParams{WorkspaceID: input.WorkspaceID, InvocationID: input.InvocationID})
		if err == nil {
			var saved, requested any
			_ = json.Unmarshal(row.Operations, &saved)
			_ = json.Unmarshal(input.Operations, &requested)
			a, _ := json.Marshal(saved)
			b, _ := json.Marshal(requested)
			if row.DocumentID != input.DocumentID || row.RequesterID != input.RequesterID || row.ExpectedRevision != input.ExpectedRevision || row.ExpectedEpoch != input.ExpectedEpoch || string(a) != string(b) {
				return domain.AgentEdit{}, domain.ErrConflict
			}
		}
	}
	if err != nil {
		return domain.AgentEdit{}, mapNotFound("queue document edit", err)
	}
	return agentEditFromRow(row), nil
}
func (r *Repository) GetAgentEdit(ctx context.Context, w, id uuid.UUID) (domain.AgentEdit, error) {
	if err := r.configured(); err != nil {
		return domain.AgentEdit{}, err
	}
	if err := r.commentQueries(ctx).ExpireDocumentAgentEdit(ctx, queries.ExpireDocumentAgentEditParams{WorkspaceID: w, InvocationID: id}); err != nil {
		return domain.AgentEdit{}, err
	}
	row, err := r.commentQueries(ctx).GetDocumentAgentEdit(ctx, queries.GetDocumentAgentEditParams{WorkspaceID: w, InvocationID: id})
	if err != nil {
		return domain.AgentEdit{}, mapNotFound("get document edit", err)
	}
	return agentEditFromRow(row), nil
}
func (r *Repository) AgentEditsAvailable(ctx context.Context) (bool, error) {
	if err := r.configured(); err != nil {
		return false, err
	}
	return r.commentQueries(ctx).DocumentAgentEditsAvailable(ctx)
}
func agentEditFromRow(row queries.DocumentAgentEdit) domain.AgentEdit {
	return domain.AgentEdit{ID: row.ID, InvocationID: row.InvocationID, WorkspaceID: row.WorkspaceID, DocumentID: row.DocumentID, RequesterID: row.RequesterID, ExpectedRevision: row.ExpectedRevision, ExpectedEpoch: row.ExpectedEpoch, Operations: row.Operations, Status: row.Status, ResultingRevision: row.ResultingRevision, Receipt: row.Receipt, ErrorCode: row.ErrorCode, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt, AppliedAt: row.AppliedAt}
}
