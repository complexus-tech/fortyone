package chatservice

import (
	"context"
	"errors"
	"github.com/complexus-tech/projects-api/internal/platform/chatrollout"
	"testing"

	domain "github.com/complexus-tech/projects-api/internal/modules/chat/domain"
	"github.com/complexus-tech/projects-api/pkg/validate"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type transcriptionJobStub struct {
	requeued         int
	jobs             []domain.TranscriptionJob
	reads            int
	readErrors       map[int]error
	completed        int
	completionErrors int
	failed           string
	text             string
}

func (s *transcriptionJobStub) ClaimTranscriptions(context.Context, int) ([]domain.TranscriptionJob, error) {
	jobs := s.jobs
	s.jobs = nil
	return jobs, nil
}
func (s *transcriptionJobStub) ReadTranscriptionAudio(context.Context, domain.TranscriptionJob) (domain.TranscriptionAudio, error) {
	s.reads++
	return domain.TranscriptionAudio{BlobName: "private-key", FileName: "voice.webm", MediaType: "audio/webm", SizeBytes: 4}, s.readErrors[s.reads]
}
func (s *transcriptionJobStub) CompleteTranscription(_ context.Context, _ domain.TranscriptionJob, text, _ string) error {
	s.completed++
	if s.completed <= s.completionErrors {
		return errors.New("database temporarily unavailable")
	}
	s.text = text
	return nil
}
func (s *transcriptionJobStub) FailTranscription(_ context.Context, _ domain.TranscriptionJob, code string) error {
	s.failed = code
	return nil
}

type transcriptionStorageStub struct {
	calls int
	limit int64
	err   error
	data  []byte
}

func (s *transcriptionStorageStub) DownloadFile(_ context.Context, container, blob string, limit int64) ([]byte, string, error) {
	s.calls++
	s.limit = limit
	return s.data, "audio/webm", s.err
}

type transcriberStub struct {
	calls int
	err   error
	text  string
}

func (p *transcriberStub) Transcribe(context.Context, domain.TranscriptionAudio, []byte) (string, string, error) {
	p.calls++
	return p.text, DefaultTranscriptionModel, p.err
}

func voiceProcessorFixture() (*TranscriptionProcessor, *transcriptionJobStub, *transcriptionStorageStub, *transcriberStub) {
	store := &transcriptionJobStub{jobs: []domain.TranscriptionJob{{AttachmentID: uuid.New(), WorkspaceID: uuid.New(), ConversationID: uuid.New(), RequesterID: uuid.New(), LeaseToken: uuid.New()}}, readErrors: map[int]error{}}
	storage := &transcriptionStorageStub{data: []byte("data")}
	provider := &transcriberStub{text: "The transcript"}
	return &TranscriptionProcessor{Store: store, Storage: storage, Container: "attachments", Provider: provider, Access: &transcriptionAccessStub{allowed: true}}, store, storage, provider
}

func TestTranscriptionWorkerPersistsResultWithoutRepeatingInferenceOnDBFailure(t *testing.T) {
	processor, store, storage, provider := voiceProcessorFixture()
	store.completionErrors = 2
	require.NoError(t, processor.ProcessPending(t.Context()))
	require.Equal(t, 3, store.completed)
	require.Equal(t, 1, provider.calls)
	require.Equal(t, "The transcript", store.text)
	require.EqualValues(t, validate.MaxVoiceNoteSize, storage.limit)
	require.Equal(t, 2, store.reads, "authorization is rechecked after storage IO")
	require.NoError(t, processor.ProcessPending(t.Context()))
	require.Equal(t, 1, provider.calls, "redelivered wake-up has no paid work to replay")
}

func TestTranscriptionWorkerDoesNotSendUnauthorizedOrInvalidAudio(t *testing.T) {
	for _, tc := range []struct {
		name, code string
		change     func(*TranscriptionProcessor, *transcriptionJobStub, *transcriptionStorageStub)
	}{
		{"deleted before download", "unavailable", func(_ *TranscriptionProcessor, s *transcriptionJobStub, _ *transcriptionStorageStub) {
			s.readErrors[1] = domain.ErrNotFound
		}},
		{"revoked during download", "unavailable", func(_ *TranscriptionProcessor, s *transcriptionJobStub, _ *transcriptionStorageStub) {
			s.readErrors[2] = domain.ErrNotFound
		}},
		{"no entitlement", "access_unavailable", func(p *TranscriptionProcessor, _ *transcriptionJobStub, _ *transcriptionStorageStub) {
			p.Access = &transcriptionAccessStub{}
		}},
		{"storage failure", "download_failed", func(_ *TranscriptionProcessor, _ *transcriptionJobStub, s *transcriptionStorageStub) {
			s.err = errors.New("storage failure")
		}},
		{"wrong size", "invalid_audio", func(_ *TranscriptionProcessor, _ *transcriptionJobStub, s *transcriptionStorageStub) {
			s.data = []byte("incorrect size")
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			processor, store, storage, provider := voiceProcessorFixture()
			tc.change(processor, store, storage)
			require.NoError(t, processor.ProcessPending(t.Context()))
			require.Zero(t, provider.calls)
			require.Equal(t, tc.code, store.failed)
			require.Zero(t, store.completed)
		})
	}
}

func TestTranscriptionWorkerMakesProviderFailureExplicitWithoutAutomaticRetry(t *testing.T) {
	processor, store, _, provider := voiceProcessorFixture()
	provider.err = context.DeadlineExceeded
	require.NoError(t, processor.ProcessPending(t.Context()))
	require.Equal(t, "transcription_failed", store.failed)
	require.Equal(t, 1, provider.calls)
	require.Zero(t, store.completed)
	require.NoError(t, processor.ProcessPending(t.Context()))
	require.Equal(t, 1, provider.calls)
}

func (s *transcriptionJobStub) RequeueTranscription(context.Context, domain.TranscriptionJob) error {
	s.requeued++
	return nil
}
func TestTranscriptionPausedBeforeProviderWorkPreservesPendingJob(t *testing.T) {
	for _, disableAt := range []int{1, 2} {
		processor, store, storage, provider := voiceProcessorFixture()
		checks := 0
		policy := chatrollout.New(chatrollout.ReaderFunc(func(context.Context, uuid.UUID) (bool, error) { checks++; return checks < disableAt, nil }))
		processor.Settings = &policy
		require.NoError(t, processor.ProcessPending(t.Context()))
		require.Equal(t, 1, store.requeued)
		require.Zero(t, provider.calls)
		require.Zero(t, store.completed)
		require.Empty(t, store.failed)
		require.Equal(t, disableAt-1, storage.calls)
	}
}
