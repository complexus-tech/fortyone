package integrationrequestsrepository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	integrationrequestdomain "github.com/complexus-tech/projects-api/internal/modules/integrationrequests/domain"
	integrationrequestssql "github.com/complexus-tech/projects-api/internal/modules/integrationrequests/repository/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (r *Repo) CreateOutboundComment(ctx context.Context, input integrationrequestdomain.CreateCommentInput, prepared integrationrequestdomain.PreparedProviderComment) (integrationrequestdomain.ProviderThread, integrationrequestdomain.Comment, error) {
	if err := r.configured(); err != nil {
		return integrationrequestdomain.ProviderThread{}, integrationrequestdomain.Comment{}, err
	}
	providerPayload := strings.TrimSpace(string(prepared.ProviderPayload))
	if providerPayload != "" && !json.Valid([]byte(providerPayload)) {
		return integrationrequestdomain.ProviderThread{}, integrationrequestdomain.Comment{}, errors.New("provider comment payload must be valid JSON")
	}
	body := strings.TrimSpace(input.Body)
	var resultThread integrationrequestdomain.ProviderThread
	var resultComment integrationrequestdomain.Comment
	err := r.withinTransaction(ctx, func(queries integrationrequestssql.Querier) error {
		existingRow, err := queries.FindIntegrationRequestCommentByClientKey(ctx, integrationrequestssql.FindIntegrationRequestCommentByClientKeyParams{
			WorkspaceID: input.WorkspaceID, ClientIdempotencyKey: uuidPointer(input.ClientIdempotencyKey),
		})
		if err == nil {
			existing := commentFromRecord(commentRecord(existingRow))
			threadRow, threadErr := queries.FindAuthorizedProviderThreadForComment(ctx, integrationrequestssql.FindAuthorizedProviderThreadForCommentParams{
				CommentID: existing.ID, RequestID: input.RequestID, ActorID: input.AuthorID,
			})
			if threadErr != nil || existing.AuthorUserID == nil || *existing.AuthorUserID != input.AuthorID || existing.Body != body {
				return integrationrequestdomain.ErrIdempotencyConflict
			}
			resultThread = providerThreadFromRecord(providerThreadRecord(threadRow))
			resultComment = existing
			return nil
		}
		if !errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("find outbound integration request comment by client key: %w", err)
		}

		threadRow, err := queries.LockAuthorizedIntegrationRequestProviderThread(ctx, integrationrequestssql.LockAuthorizedIntegrationRequestProviderThreadParams{
			WorkspaceID: input.WorkspaceID, RequestID: input.RequestID, ActorID: input.AuthorID,
		})
		if err != nil {
			return mapProviderThreadNotFound("lock authorized integration request provider thread", err)
		}
		thread := providerThreadFromRecord(providerThreadRecord(threadRow))
		idempotencyKey := "integration-request-comment:" + input.ClientIdempotencyKey.String()
		commentRow, err := queries.CreateOutboundIntegrationRequestComment(ctx, integrationrequestssql.CreateOutboundIntegrationRequestCommentParams{
			CommentID: uuid.New(), WorkspaceID: input.WorkspaceID, ThreadID: thread.ID,
			ActorID: uuidPointer(input.AuthorID), ClientIdempotencyKey: uuidPointer(input.ClientIdempotencyKey),
			OutboundIdempotencyKey: stringPointer(idempotencyKey), Body: body,
		})
		if errors.Is(err, pgx.ErrNoRows) {
			existingThread, existingComment, conflictErr := resolveOutboundCommentConflict(ctx, queries, input, body)
			if conflictErr != nil {
				return integrationrequestdomain.ErrIdempotencyConflict
			}
			resultThread, resultComment = existingThread, existingComment
			return nil
		}
		if err != nil {
			return fmt.Errorf("create outbound integration request comment: %w", err)
		}
		comment := commentFromRecord(commentRecord(commentRow))
		if err := queries.CreateIntegrationRequestCommentDelivery(ctx, integrationrequestssql.CreateIntegrationRequestCommentDeliveryParams{
			Provider: thread.Provider, WorkspaceID: input.WorkspaceID, ActorID: uuidPointer(input.AuthorID),
			InstallationGeneration: thread.InstallationGeneration, ExternalWorkspaceID: thread.ExternalWorkspaceID,
			ExternalRecipientUserID: strings.TrimSpace(prepared.ExternalRecipientUserID), IdempotencyKey: idempotencyKey,
			ExternalChannelID: thread.ExternalChannelID, ExternalThreadID: stringPointer(thread.ExternalThreadID),
			Content: stringPointer(body), ProviderPayload: providerPayload,
		}); err != nil {
			return fmt.Errorf("create integration request comment delivery: %w", err)
		}
		resultThread, resultComment = thread, comment
		return nil
	})
	if err != nil {
		return integrationrequestdomain.ProviderThread{}, integrationrequestdomain.Comment{}, fmt.Errorf("create outbound integration request comment transaction: %w", err)
	}
	return resultThread, resultComment, nil
}

func resolveOutboundCommentConflict(ctx context.Context, queries integrationrequestssql.Querier, input integrationrequestdomain.CreateCommentInput, body string) (integrationrequestdomain.ProviderThread, integrationrequestdomain.Comment, error) {
	row, err := queries.FindIntegrationRequestCommentByClientKey(ctx, integrationrequestssql.FindIntegrationRequestCommentByClientKeyParams{
		WorkspaceID: input.WorkspaceID, ClientIdempotencyKey: uuidPointer(input.ClientIdempotencyKey),
	})
	if err != nil {
		return integrationrequestdomain.ProviderThread{}, integrationrequestdomain.Comment{}, integrationrequestdomain.ErrIdempotencyConflict
	}
	comment := commentFromRecord(commentRecord(row))
	if comment.AuthorUserID == nil || *comment.AuthorUserID != input.AuthorID || comment.Body != body {
		return integrationrequestdomain.ProviderThread{}, integrationrequestdomain.Comment{}, integrationrequestdomain.ErrIdempotencyConflict
	}
	threadRow, err := queries.FindAuthorizedProviderThreadForComment(ctx, integrationrequestssql.FindAuthorizedProviderThreadForCommentParams{
		CommentID: comment.ID, RequestID: input.RequestID, ActorID: input.AuthorID,
	})
	if err != nil {
		return integrationrequestdomain.ProviderThread{}, integrationrequestdomain.Comment{}, integrationrequestdomain.ErrIdempotencyConflict
	}
	return providerThreadFromRecord(providerThreadRecord(threadRow)), comment, nil
}

func (r *Repo) IngestInboundProviderComment(ctx context.Context, input integrationrequestdomain.InboundProviderCommentInput) (bool, error) {
	if err := r.configured(); err != nil {
		return false, err
	}
	provider := strings.TrimSpace(input.Provider)
	externalWorkspaceID := strings.TrimSpace(input.ExternalWorkspaceID)
	externalChannelID := strings.TrimSpace(input.ExternalChannelID)
	externalThreadID := strings.TrimSpace(input.ExternalThreadID)
	externalMessageID := strings.TrimSpace(input.ExternalMessageID)
	externalAuthorID := strings.TrimSpace(input.ExternalAuthorID)
	body := strings.TrimSpace(input.Body)
	if provider == "" || externalWorkspaceID == "" || externalChannelID == "" || externalThreadID == "" || externalMessageID == "" || externalAuthorID == "" || body == "" {
		return false, nil
	}
	createdAt := input.CreatedAt.UTC()
	if createdAt.IsZero() {
		createdAt = time.Now().UTC()
	}
	affected, err := r.queries.IngestInboundIntegrationRequestComment(ctx, integrationrequestssql.IngestInboundIntegrationRequestCommentParams{
		AuthorUserID: input.AuthorUserID, ExternalAuthorID: externalAuthorID, ExternalMessageID: stringPointer(externalMessageID),
		Body: body, CreatedAt: createdAt, Provider: provider, ExternalWorkspaceID: externalWorkspaceID,
		ExternalChannelID: externalChannelID, ExternalThreadID: externalThreadID,
		InstallationGeneration: uuidPointer(input.InstallationGeneration),
	})
	if err != nil {
		return false, fmt.Errorf("ingest inbound integration request comment: %w", err)
	}
	if affected > 0 {
		return true, nil
	}
	known, err := r.queries.IsCurrentIntegrationRequestProviderThreadKnown(ctx, integrationrequestssql.IsCurrentIntegrationRequestProviderThreadKnownParams{
		Provider: provider, ExternalWorkspaceID: externalWorkspaceID, ExternalChannelID: externalChannelID,
		ExternalThreadID: externalThreadID, InstallationGeneration: uuidPointer(input.InstallationGeneration),
	})
	if err != nil {
		return false, fmt.Errorf("check current integration request provider thread after inbound conflict: %w", err)
	}
	return known, nil
}

func (r *Repo) GetCommentForUser(ctx context.Context, workspaceID, commentID, userID uuid.UUID) (integrationrequestdomain.Comment, error) {
	if err := r.configured(); err != nil {
		return integrationrequestdomain.Comment{}, err
	}
	row, err := r.queries.GetAuthorizedIntegrationRequestComment(ctx, integrationrequestssql.GetAuthorizedIntegrationRequestCommentParams{
		WorkspaceID: workspaceID, CommentID: commentID, ActorID: userID,
	})
	if err != nil {
		return integrationrequestdomain.Comment{}, mapNotFound("get authorized integration request comment", err)
	}
	return commentFromRecord(commentRecord(row)), nil
}

func (r *Repo) listThreadComments(ctx context.Context, threadID uuid.UUID) ([]integrationrequestdomain.Comment, error) {
	rows, err := r.queries.ListIntegrationRequestThreadComments(ctx, integrationrequestssql.ListIntegrationRequestThreadCommentsParams{ThreadID: threadID})
	if err != nil {
		return nil, fmt.Errorf("list integration request thread comments: %w", err)
	}
	result := make([]integrationrequestdomain.Comment, 0, len(rows))
	for _, row := range rows {
		result = append(result, commentFromRecord(commentRecord(row)))
	}
	return result, nil
}
