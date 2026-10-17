package chatrepository

import (
	"context"

	domain "github.com/complexus-tech/projects-api/internal/modules/chat/domain"
	chatsql "github.com/complexus-tech/projects-api/internal/modules/chat/repository/sqlc"
	"github.com/google/uuid"
)

func (r *Repository) GetTranscription(ctx context.Context, scope domain.AttachmentScope, id uuid.UUID) (domain.Transcription, error) {
	if _, err := r.authorize(ctx, scope.ActorScope, scope.ConversationID, false); err != nil {
		return domain.Transcription{}, err
	}
	row, err := r.query(ctx).GetAudioTranscription(ctx, chatsql.GetAudioTranscriptionParams{AttachmentID: id, WorkspaceID: scope.WorkspaceID, ConversationID: scope.ConversationID})
	if err != nil {
		return domain.Transcription{}, scopedError(err)
	}
	return domain.Transcription{Status: row.Status, Text: row.Transcript, Error: row.ErrorCode, UpdatedAt: row.UpdatedAt}, nil
}

func (r *Repository) RequestTranscription(ctx context.Context, scope domain.AttachmentScope, id uuid.UUID, retry bool) (result domain.Transcription, err error) {
	err = r.WithinUnit(ctx, func(unit context.Context) error {
		if _, err := r.authorize(unit, scope.ActorScope, scope.ConversationID, true); err != nil {
			return err
		}
		q := r.query(unit)
		if _, err := q.LockTranscriptionAudio(unit, chatsql.LockTranscriptionAudioParams{AttachmentID: id, WorkspaceID: scope.WorkspaceID, ConversationID: scope.ConversationID}); err != nil {
			return scopedError(err)
		}
		if err := q.RequestAudioTranscription(unit, chatsql.RequestAudioTranscriptionParams{AttachmentID: id, WorkspaceID: scope.WorkspaceID, ConversationID: scope.ConversationID, UserID: scope.Actor.PrincipalID, Retry: retry}); err != nil {
			return err
		}
		result, err = r.GetTranscription(unit, scope, id)
		return err
	})
	return
}

func (r *Repository) ClaimTranscriptions(ctx context.Context, limit int) ([]domain.TranscriptionJob, error) {
	if limit < 1 || limit > 2 {
		return nil, domain.ErrInvalid
	}
	// Never automatically replay a provider call with an ambiguous result.
	if err := r.query(ctx).ExpireAudioTranscriptions(ctx); err != nil {
		return nil, err
	}
	rows, err := r.query(ctx).ClaimAudioTranscriptions(ctx, chatsql.ClaimAudioTranscriptionsParams{PageLimit: int32(limit)})
	if err != nil {
		return nil, err
	}
	jobs := make([]domain.TranscriptionJob, 0, len(rows))
	for _, row := range rows {
		if row.LeaseToken != nil {
			jobs = append(jobs, domain.TranscriptionJob{AttachmentID: row.AttachmentID, WorkspaceID: row.WorkspaceID, ConversationID: row.ConversationID, RequesterID: row.RequestedBy, LeaseToken: *row.LeaseToken})
		}
	}
	return jobs, nil
}

func (r *Repository) ReadTranscriptionAudio(ctx context.Context, job domain.TranscriptionJob) (domain.TranscriptionAudio, error) {
	row, err := r.query(ctx).ReadTranscriptionAudio(ctx, chatsql.ReadTranscriptionAudioParams{AttachmentID: job.AttachmentID, LeaseToken: &job.LeaseToken})
	if err != nil {
		return domain.TranscriptionAudio{}, scopedError(err)
	}
	return domain.TranscriptionAudio{BlobName: row.BlobName, FileName: row.Filename, MediaType: row.MimeType, SizeBytes: row.Size}, nil
}

func (r *Repository) CompleteTranscription(ctx context.Context, job domain.TranscriptionJob, text, model string) error {
	count, err := r.query(ctx).CompleteAudioTranscription(ctx, chatsql.CompleteAudioTranscriptionParams{AttachmentID: job.AttachmentID, LeaseToken: &job.LeaseToken, Transcript: &text, Model: &model})
	if err != nil {
		return err
	}
	if count != 1 {
		return domain.ErrConflict
	}
	return nil
}

func (r *Repository) FailTranscription(ctx context.Context, job domain.TranscriptionJob, code string) error {
	_, err := r.query(ctx).FailAudioTranscription(ctx, chatsql.FailAudioTranscriptionParams{AttachmentID: job.AttachmentID, LeaseToken: &job.LeaseToken, ErrorCode: &code})
	return err
}

// RequeueTranscription is only used before any provider request has started.
func (r *Repository) RequeueTranscription(ctx context.Context, job domain.TranscriptionJob) error {
	_, err := r.query(ctx).RequeueAudioTranscription(ctx, chatsql.RequeueAudioTranscriptionParams{AttachmentID: job.AttachmentID, LeaseToken: &job.LeaseToken})
	return err
}
