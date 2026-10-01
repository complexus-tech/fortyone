package storiesrepository

import (
	"context"
	"fmt"
	storydomain "github.com/complexus-tech/projects-api/internal/modules/stories/domain"
	storyreadsql "github.com/complexus-tech/projects-api/internal/modules/stories/repository/sqlc"
	"github.com/complexus-tech/projects-api/internal/platform/auth"
)

func (r *repo) RecordImportReceipt(ctx context.Context, receipt storydomain.ImportReceipt) error {
	if r.reads == nil {
		return fmt.Errorf("import receipt repository unavailable")
	}
	return r.reads.RecordImportReceipt(ctx, storyreadsql.RecordImportReceiptParams{
		WorkspaceID: receipt.WorkspaceID, TeamID: receipt.TeamID, CreationKey: receipt.CreationKey,
		Provider: receipt.Provider, SourceDigest: receipt.SourceDigest, SourceNamespace: receipt.SourceNamespace,
		SourceKey: receipt.SourceKey, StoryID: receipt.StoryID, Created: receipt.Created,
		ErrorCode: receipt.ErrorCode, ErrorMessage: receipt.ErrorMessage, SourceMetadata: receipt.SourceMetadata,
	})
}
func (r *repo) ListImportReceipts(ctx context.Context, actor auth.Actor, provider, digest string, namespace *string, limit, offset int) ([]storydomain.ImportReceipt, error) {
	if r.reads == nil {
		return nil, fmt.Errorf("import receipt repository unavailable")
	}
	rows, err := r.reads.ListImportReceipts(ctx, storyreadsql.ListImportReceiptsParams{
		WorkspaceID: actor.WorkspaceID, ActorID: actor.PrincipalID, Provider: provider, SourceDigest: digest, SourceNamespace: namespace,
		TeamAccessUnrestricted: actor.TeamAccess.IsUnrestricted(), AllowedTeamIds: actor.TeamAccess.RestrictedTeamIDs(), PageLimit: int32(limit), PageOffset: int32(offset),
	})
	if err != nil {
		return nil, err
	}
	result := make([]storydomain.ImportReceipt, 0, len(rows))
	for _, row := range rows {
		result = append(result, storydomain.ImportReceipt{WorkspaceID: row.WorkspaceID, TeamID: row.TeamID, CreationKey: row.CreationKey, Provider: row.Provider, SourceDigest: row.SourceDigest, SourceNamespace: row.SourceNamespace, SourceKey: row.SourceKey, StoryID: row.StoryID, Created: row.Created, ErrorCode: row.ErrorCode, ErrorMessage: row.ErrorMessage, SourceMetadata: row.SourceMetadata, UpdatedAt: row.UpdatedAt})
	}
	return result, nil
}
