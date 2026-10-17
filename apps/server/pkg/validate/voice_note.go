package validate

import (
	"errors"
	"io"
	"math"
	"mime/multipart"
	"path/filepath"
	"strings"
)

const (
	MaxVoiceNoteSize       = 20 * 1024 * 1024
	MaxVoiceNoteDurationMS = 5 * 60 * 1000
)

var ErrVoiceNoteDuration = errors.New("voice notes must be five minutes or shorter")

// VoiceNote validates the browser recording container without trusting a MIME
// header. It deliberately supports only the audio-only formats our recorder
// produces. Duration from a streaming WebM may be absent; the declared duration
// is advisory and the byte ceiling is always enforced server-side.
func VoiceNote(file multipart.File, header *multipart.FileHeader, durationMS int64) (string, error) {
	if durationMS <= 0 || durationMS > MaxVoiceNoteDurationMS {
		return "", ErrVoiceNoteDuration
	}
	return validateAudioContainer(file, header, MaxVoiceNoteSize, MaxVoiceNoteDurationMS)
}

// ChatAudio accepts existing audio files without claiming they were recorded in
// Chat or requiring a client-supplied duration. MP4s are classified before the
// smaller audio size ceiling so the caller can preserve ordinary video uploads.
func ChatAudio(file multipart.File, header *multipart.FileHeader) (string, error) {
	contentType, err := validateAudioContainer(file, header, MaxAttachmentSize, 0)
	if err != nil {
		return "", err
	}
	if header.Size > MaxVoiceNoteSize {
		return "", ErrFileTooLarge
	}
	return contentType, nil
}

func validateAudioContainer(file multipart.File, header *multipart.FileHeader, maxBytes int64, maxDurationMS float64) (string, error) {
	if file == nil || header == nil || header.Size <= 0 {
		return "", ErrEmptyFile
	}
	if header.Size > maxBytes {
		return "", ErrFileTooLarge
	}
	if len(header.Filename) > 255 {
		return "", ErrFileNameTooLong
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return "", err
	}
	// Structural parsing is bounded by the upload limit and a small metadata
	// budget. Media payloads are skipped with ReaderAt, never decoded in the API.
	p := audioContainer{source: file, size: header.Size, budget: 4096, maxDurationMS: maxDurationMS}
	switch strings.ToLower(filepath.Ext(header.Filename)) {
	case ".webm":
		if err := p.webm(); err != nil {
			return "", err
		}
		return "audio/webm", nil
	case ".m4a", ".mp4":
		if err := p.mp4(); err != nil {
			return "", err
		}
		return "audio/mp4", nil
	default:
		return "", ErrInvalidFileType
	}
}

type audioContainer struct {
	source        io.ReaderAt
	size          int64
	budget        int
	maxDurationMS float64
}

func (p *audioContainer) read(offset int64, count int) ([]byte, error) {
	if offset < 0 || count < 0 || offset > p.size-int64(count) {
		return nil, ErrInvalidFileType
	}
	data := make([]byte, count)
	if _, err := p.source.ReadAt(data, offset); err != nil {
		return nil, ErrInvalidFileType
	}
	return data, nil
}

func (p *audioContainer) step() error {
	p.budget--
	if p.budget < 0 {
		return ErrInvalidFileType
	}
	return nil
}

func (p *audioContainer) checkDuration(durationMS float64) error {
	if math.IsNaN(durationMS) || math.IsInf(durationMS, 0) || durationMS < 0 {
		return ErrInvalidFileType
	}
	// Container clocks can include encoder padding at the five-minute boundary.
	if p.maxDurationMS > 0 && durationMS > p.maxDurationMS+1000 {
		return ErrVoiceNoteDuration
	}
	return nil
}
