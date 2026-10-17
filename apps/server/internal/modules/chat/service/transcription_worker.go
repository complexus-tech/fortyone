package chatservice

import (
	"context"
	"errors"
	"fmt"
	"time"
	"unicode/utf8"

	domain "github.com/complexus-tech/projects-api/internal/modules/chat/domain"
	"github.com/complexus-tech/projects-api/internal/platform/chatrollout"
	"github.com/complexus-tech/projects-api/pkg/validate"
)

type TranscriptionJobStore interface {
	ClaimTranscriptions(context.Context, int) ([]domain.TranscriptionJob, error)
	ReadTranscriptionAudio(context.Context, domain.TranscriptionJob) (domain.TranscriptionAudio, error)
	CompleteTranscription(context.Context, domain.TranscriptionJob, string, string) error
	FailTranscription(context.Context, domain.TranscriptionJob, string) error
}

type TranscriptionStorage interface {
	DownloadFile(context.Context, string, string, int64) ([]byte, string, error)
}

type AudioTranscriber interface {
	Transcribe(context.Context, domain.TranscriptionAudio, []byte) (text, model string, err error)
}

type TranscriptionProcessor struct {
	Settings  *chatrollout.Policy
	Store     TranscriptionJobStore
	Storage   TranscriptionStorage
	Container string
	Provider  AudioTranscriber
	Access    TranscriptionAccess
}

func (p *TranscriptionProcessor) ProcessPending(ctx context.Context) error {
	jobs, err := p.Store.ClaimTranscriptions(ctx, 2)
	if err != nil {
		return err
	}
	var failures []error
	for _, job := range jobs {
		if err := p.process(ctx, job); err != nil {
			failures = append(failures, err)
		}
	}
	return errors.Join(failures...)
}

func (p *TranscriptionProcessor) process(ctx context.Context, job domain.TranscriptionJob) error {
	operation, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()
	if p.Settings != nil && !p.Settings.Allows(operation, job.WorkspaceID) {
		return p.requeue(ctx, job)
	}
	audio, err := p.Store.ReadTranscriptionAudio(operation, job)
	if err != nil {
		return p.fail(ctx, job, "unavailable")
	}
	allowed, err := p.Access.WorkspaceCanUseMaya(operation, job.WorkspaceID)
	if err != nil || !allowed {
		return p.fail(ctx, job, "access_unavailable")
	}
	data, _, err := p.Storage.DownloadFile(operation, p.Container, audio.BlobName, validate.MaxVoiceNoteSize)
	if err != nil {
		return p.fail(ctx, job, "download_failed")
	}
	if len(data) == 0 || int64(len(data)) != audio.SizeBytes || len(data) > validate.MaxVoiceNoteSize {
		return p.fail(ctx, job, "invalid_audio")
	}
	// Check the attachment and live requester again after storage IO. A pending
	// upload, removed message, revoked membership or scanner block cannot become
	// authority for a provider request just because a job was previously claimed.
	if p.Settings != nil && !p.Settings.Allows(operation, job.WorkspaceID) {
		return p.requeue(ctx, job)
	}
	if _, err := p.Store.ReadTranscriptionAudio(operation, job); err != nil {
		return p.fail(ctx, job, "unavailable")
	}
	allowed, err = p.Access.WorkspaceCanUseMaya(operation, job.WorkspaceID)
	if err != nil || !allowed {
		return p.fail(ctx, job, "access_unavailable")
	}
	text, model, err := p.Provider.Transcribe(operation, audio, data)
	if err != nil {
		return p.fail(ctx, job, "transcription_failed")
	}
	if len(text) > 1<<20 || !utf8.ValidString(text) {
		return p.fail(ctx, job, "invalid_response")
	}
	// A provider result is held in memory while retrying only its DB write.
	// Never repeat the paid call when a completion write or queue delivery fails.
	completion, completeCancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
	defer completeCancel()
	for attempt := 0; attempt < 3; attempt++ {
		err = p.Store.CompleteTranscription(completion, job, text, model)
		if err == nil || errors.Is(err, domain.ErrConflict) {
			return nil
		}
		if completion.Err() != nil {
			break
		}
	}
	return fmt.Errorf("persist voice transcription result: %w", err)
}

func (p *TranscriptionProcessor) fail(ctx context.Context, job domain.TranscriptionJob, code string) error {
	completion, cancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
	defer cancel()
	return p.Store.FailTranscription(completion, job, code)
}

func (p *TranscriptionProcessor) requeue(ctx context.Context, job domain.TranscriptionJob) error {
	store, ok := p.Store.(interface {
		RequeueTranscription(context.Context, domain.TranscriptionJob) error
	})
	if !ok {
		return errors.New("transcription pause support unavailable")
	}
	completion, cancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
	defer cancel()
	return store.RequeueTranscription(completion, job)
}
