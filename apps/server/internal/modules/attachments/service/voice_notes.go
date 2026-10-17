package attachments

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strings"

	"github.com/complexus-tech/projects-api/pkg/validate"
	"github.com/google/uuid"
)

// UploadChatVoiceNote keeps recorded audio separate from the generic story and
// document upload allowlists while reusing private storage and orphan cleanup.
func (s *Service) UploadChatVoiceNote(ctx context.Context, file multipart.File, header *multipart.FileHeader, userID, workspaceID uuid.UUID, durationMS int64) (FileInfo, error) {
	contentType, err := validate.VoiceNote(file, header, durationMS)
	if err != nil {
		switch {
		case errors.Is(err, validate.ErrFileTooLarge):
			return FileInfo{}, fmt.Errorf("%w: maximum voice note size is 20 MB", ErrFileTooLarge)
		case errors.Is(err, validate.ErrInvalidFileType):
			return FileInfo{}, fmt.Errorf("%w: record an audio-only WebM or MP4 voice note", ErrInvalidFileType)
		case errors.Is(err, validate.ErrVoiceNoteDuration):
			return FileInfo{}, fmt.Errorf("%w: %v", ErrInvalidFile, err)
		default:
			return FileInfo{}, attachmentValidationError(err)
		}
	}
	return s.uploadAttachmentWithDetector(ctx, file, header, userID, workspaceID, uuid.Nil,
		func(multipart.File, *multipart.FileHeader) error { return nil }, nil,
		func(multipart.File, *multipart.FileHeader) (string, error) { return contentType, nil },
	)
}

// UploadChatAttachment extends only Chat's upload surface with audio files. The
// absence of voice-note metadata is preserved; a selected file is not a recording.
func (s *Service) UploadChatAttachment(ctx context.Context, file multipart.File, header *multipart.FileHeader, userID, workspaceID uuid.UUID) (FileInfo, error) {
	if header == nil {
		return FileInfo{}, attachmentValidationError(validate.ErrEmptyFile)
	}
	extension := strings.ToLower(filepath.Ext(header.Filename))
	if extension != ".webm" && extension != ".m4a" && extension != ".mp4" {
		return s.uploadAttachmentWithDetector(ctx, file, header, userID, workspaceID, uuid.Nil, validate.Attachment, nil, detectFileContentType)
	}
	contentType, err := validate.ChatAudio(file, header)
	if errors.Is(err, validate.ErrInvalidFileType) && extension == ".mp4" {
		// An MP4 may contain video. Retain the existing generic validation path
		// without allowing video to acquire an audio MIME type or transcription.
		return s.uploadAttachmentWithDetector(ctx, file, header, userID, workspaceID, uuid.Nil, validate.Attachment, nil, detectFileContentType)
	}
	if errors.Is(err, validate.ErrFileTooLarge) {
		return FileInfo{}, fmt.Errorf("%w: maximum audio attachment size is 20 MB", ErrFileTooLarge)
	}
	if errors.Is(err, validate.ErrInvalidFileType) {
		return FileInfo{}, fmt.Errorf("%w: upload an audio-only WebM or MP4 audio file", ErrInvalidFileType)
	}
	if err != nil {
		return FileInfo{}, attachmentValidationError(err)
	}
	return s.uploadAttachmentWithDetector(ctx, file, header, userID, workspaceID, uuid.Nil,
		func(multipart.File, *multipart.FileHeader) error { return nil }, nil,
		func(multipart.File, *multipart.FileHeader) (string, error) { return contentType, nil },
	)
}
