package chatservice

import (
	"context"
	"fmt"

	domain "github.com/complexus-tech/projects-api/internal/modules/chat/domain"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/google/uuid"
)

type TranscriptionRepository interface {
	GetTranscription(context.Context, domain.AttachmentScope, uuid.UUID) (domain.Transcription, error)
	RequestTranscription(context.Context, domain.AttachmentScope, uuid.UUID, bool) (domain.Transcription, error)
}

type TranscriptionAccess interface {
	WorkspaceCanUseMaya(context.Context, uuid.UUID) (bool, error)
}

type Transcriptions struct {
	repo       TranscriptionRepository
	access     TranscriptionAccess
	configured bool
	wake       func(context.Context) error
	log        *logger.Logger
}

func NewTranscriptions(repo TranscriptionRepository, access TranscriptionAccess, configured bool, wake func(context.Context) error, log *logger.Logger) *Transcriptions {
	return &Transcriptions{repo: repo, access: access, configured: configured, wake: wake, log: log}
}

func (s *Transcriptions) Get(ctx context.Context, scope domain.AttachmentScope, id uuid.UUID) (domain.Transcription, error) {
	if err := validateAttachmentScope(scope, id, false); err != nil {
		return domain.Transcription{}, err
	}
	return s.repo.GetTranscription(ctx, scope, id)
}

func (s *Transcriptions) Request(ctx context.Context, scope domain.AttachmentScope, id uuid.UUID, retry bool) (domain.Transcription, error) {
	if err := validateAttachmentScope(scope, id, true); err != nil {
		return domain.Transcription{}, err
	}
	current, err := s.repo.GetTranscription(ctx, scope, id)
	if err != nil {
		return domain.Transcription{}, err
	}
	// Repeated clicks never bill again. Existing results and in-flight requests
	// remain readable even if Maya eligibility has since changed.
	if current.Status != "not_requested" && !(current.Status == "failed" && retry) {
		return current, nil
	}
	if !s.configured || s.access == nil {
		return domain.Transcription{}, domain.ErrTranscriptionUnavailable
	}
	allowed, err := s.access.WorkspaceCanUseMaya(ctx, scope.WorkspaceID)
	if err != nil {
		return domain.Transcription{}, fmt.Errorf("check voice transcription access: %w", err)
	}
	if !allowed {
		return domain.Transcription{}, domain.ErrTranscriptionAccess
	}
	result, err := s.repo.RequestTranscription(ctx, scope, id, retry)
	if err != nil {
		return domain.Transcription{}, err
	}
	if result.Status == "pending" && s.wake != nil {
		if err := s.wake(ctx); err != nil && s.log != nil {
			s.log.Error(ctx, "voice transcription awaits scheduled dispatch", "attachment_id", id, "error", err)
		}
	}
	return result, nil
}
