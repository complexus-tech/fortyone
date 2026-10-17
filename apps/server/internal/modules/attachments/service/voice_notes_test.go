package attachments

import (
	"bytes"
	"context"
	"encoding/binary"
	"io"
	"log/slog"
	"mime/multipart"
	"testing"

	attachmentdomain "github.com/complexus-tech/projects-api/internal/modules/attachments/domain"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/storage"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestChatVoiceUploadUsesCanonicalAudioMimeAndExistingPrivateStorageLifecycle(t *testing.T) {
	// A streaming audio-only WebM header, one Opus track and a cluster. The API
	// validates the container; the browser/provider own actual audio decoding.
	data := []byte{
		0x1a, 0x45, 0xdf, 0xa3, 0x87, 0x42, 0x82, 0x84, 'w', 'e', 'b', 'm',
		0x18, 0x53, 0x80, 0x67, 0xff,
		0x16, 0x54, 0xae, 0x6b, 0x8d, 0xae, 0x8b, 0x83, 0x81, 0x02, 0x86, 0x86, 'A', '_', 'O', 'P', 'U', 'S',
		0x1f, 0x43, 0xb6, 0x75, 0xff, 0xa3, 0x81, 0x00,
	}
	repo := &attachmentRepositoryStub{}
	objects := &attachmentStorageStub{}
	service := New(logger.NewWithText(io.Discard, slog.LevelError, "voice-test"), repo, objects, storage.Config{AttachmentsBucket: "attachments"}, nil)
	header := &multipart.FileHeader{Filename: "voice.webm", Size: int64(len(data)), Header: map[string][]string{"Content-Type": {"application/octet-stream"}}}
	file := testMultipartFile{Reader: bytes.NewReader(data)}
	uploaded, err := service.UploadChatVoiceNote(context.Background(), file, header, uuid.New(), uuid.New(), 1000)
	require.NoError(t, err)
	require.Equal(t, "audio/webm", uploaded.MimeType)
	require.Equal(t, "audio/webm", objects.contentType)
	require.Equal(t, 1, objects.uploadCount)
	require.Equal(t, 1, objects.generateCount)
	require.Equal(t, attachmentdomain.ScanStatusUnscanned, repo.attachment.ScanStatus)
	require.Nil(t, repo.attachment.ScanCompletedAt)
	_, err = service.UploadAttachment(context.Background(), testMultipartFile{Reader: bytes.NewReader(data)}, header, uuid.New(), uuid.New())
	require.ErrorIs(t, err, ErrInvalidFileType, "recording support does not broaden generic attachment allowlists")
	require.Equal(t, 1, objects.uploadCount)
	uploaded, err = service.UploadChatAttachment(context.Background(), testMultipartFile{Reader: bytes.NewReader(data)}, header, uuid.New(), uuid.New())
	require.NoError(t, err)
	require.Equal(t, "audio/webm", uploaded.MimeType)
	require.Equal(t, 2, objects.uploadCount)
}

func TestChatAttachmentUploadAcceptsExistingAudioAndPreservesVideo(t *testing.T) {
	mp4box := func(kind string, content []byte) []byte {
		box := make([]byte, 8, 8+len(content))
		binary.BigEndian.PutUint32(box, uint32(8+len(content)))
		copy(box[4:], kind)
		return append(box, content...)
	}
	mp4 := func(handler string) []byte {
		data := mp4box("ftyp", []byte("mp42"+"\x00\x00\x00\x00"+"mp42"))
		hdlr := mp4box("hdlr", append(make([]byte, 8), []byte(handler)...))
		data = append(data, mp4box("moov", mp4box("trak", mp4box("mdia", hdlr)))...)
		return append(data, mp4box("mdat", []byte{1, 2, 3})...)
	}
	for _, tc := range []struct {
		name, filename, mime string
		data                 []byte
	}{
		{"audio file", "interview.m4a", "audio/mp4", mp4("soun")},
		{"audio mp4", "interview.mp4", "audio/mp4", mp4("soun")},
		{"video mp4", "walkthrough.mp4", "video/mp4", mp4("vide")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo := &attachmentRepositoryStub{}
			objects := &attachmentStorageStub{}
			service := New(logger.NewWithText(io.Discard, slog.LevelError, "audio-test"), repo, objects, storage.Config{AttachmentsBucket: "attachments"}, nil)
			header := &multipart.FileHeader{Filename: tc.filename, Size: int64(len(tc.data)), Header: map[string][]string{"Content-Type": {"application/octet-stream"}}}
			uploaded, err := service.UploadChatAttachment(context.Background(), testMultipartFile{Reader: bytes.NewReader(tc.data)}, header, uuid.New(), uuid.New())
			require.NoError(t, err)
			require.Equal(t, tc.mime, uploaded.MimeType)
			require.Equal(t, tc.mime, objects.contentType)
			require.Equal(t, 1, objects.uploadCount)
		})
	}
}

func TestChatDocumentUploadNeedsNoScannerAndDoesNotClaimToBeScanned(t *testing.T) {
	repo, objects := &attachmentRepositoryStub{}, &attachmentStorageStub{}
	service := New(logger.NewWithText(io.Discard, slog.LevelError, "chat-upload"), repo, objects, storage.Config{AttachmentsBucket: "attachments"}, nil)
	data := []byte("Product launch checklist")
	header := &multipart.FileHeader{Filename: "notes.txt", Size: int64(len(data)), Header: map[string][]string{"Content-Type": {"image/png"}}}
	result, err := service.UploadChatAttachment(t.Context(), testMultipartFile{Reader: bytes.NewReader(data)}, header, uuid.New(), uuid.New())
	require.NoError(t, err)
	require.Equal(t, "text/plain", result.MimeType, "the request MIME header does not override detected content")
	require.Equal(t, 1, objects.uploadCount)
	require.Equal(t, attachmentdomain.ScanStatusUnscanned, repo.attachment.ScanStatus)
	require.Nil(t, repo.attachment.ScanCompletedAt)
	require.True(t, repo.attachment.AvailableForDownload())
}

func TestChatUploadValidationStillStopsInvalidFilesBeforeStorage(t *testing.T) {
	for _, test := range []struct {
		name, filename string
		data           []byte
		size, duration int64
		voice          bool
		expected       error
	}{
		{name: "oversized document", filename: "notes.txt", size: 26 << 20, expected: ErrFileTooLarge},
		{name: "unsupported binary", filename: "app.exe", data: []byte{0, 1, 2, 3}, size: 4, expected: ErrInvalidFileType},
		{name: "HTML disguised as PDF", filename: "notes.pdf", data: []byte("<!DOCTYPE html><title>example</title>"), size: 36, expected: ErrInvalidFileType},
		{name: "oversized recording", filename: "voice.webm", size: 21 << 20, duration: 1000, voice: true, expected: ErrFileTooLarge},
		{name: "recording over five minutes", filename: "voice.webm", size: 4, duration: 300001, voice: true, expected: ErrInvalidFile},
		{name: "malformed recording", filename: "voice.webm", size: 4, duration: 1000, voice: true, expected: ErrInvalidFileType},
	} {
		t.Run(test.name, func(t *testing.T) {
			repo, objects := &attachmentRepositoryStub{}, &attachmentStorageStub{}
			service := New(logger.NewWithText(io.Discard, slog.LevelError, "chat-upload"), repo, objects, storage.Config{AttachmentsBucket: "attachments"}, nil)
			data := test.data
			if data == nil {
				data = []byte("text")
			}
			file := testMultipartFile{Reader: bytes.NewReader(data)}
			header := &multipart.FileHeader{Filename: test.filename, Size: test.size}
			var err error
			if test.voice {
				_, err = service.UploadChatVoiceNote(t.Context(), file, header, uuid.New(), uuid.New(), test.duration)
			} else {
				_, err = service.UploadChatAttachment(t.Context(), file, header, uuid.New(), uuid.New())
			}
			require.ErrorIs(t, err, test.expected)
			require.Zero(t, objects.uploadCount)
			require.Zero(t, repo.createCount)
		})
	}
}
