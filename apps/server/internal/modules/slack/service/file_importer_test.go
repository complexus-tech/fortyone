package slack

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"
)

type slackAttachmentUploaderStub struct {
	called      bool
	bytes       []byte
	filename    string
	tempPath    string
	actorID     uuid.UUID
	storyID     uuid.UUID
	workspaceID uuid.UUID
	importID    uuid.UUID
	attachment  uuid.UUID
}

func (s *slackAttachmentUploaderStub) UploadSlackAttachmentAndLinkToStory(_ context.Context, file multipart.File, header *multipart.FileHeader, actorID, storyID, workspaceID, importID uuid.UUID) (uuid.UUID, error) {
	s.called = true
	s.filename = header.Filename
	s.actorID = actorID
	s.storyID = storyID
	s.workspaceID = workspaceID
	s.importID = importID
	if tempFile, ok := file.(*os.File); ok {
		s.tempPath = tempFile.Name()
	}
	data, err := io.ReadAll(file)
	if err != nil {
		return uuid.Nil, err
	}
	s.bytes = data
	return s.attachment, nil
}

func TestSlackFileImporterImport(t *testing.T) {
	const fileID = "F123ABC"
	const downloadURL = "https://files.slack.com/files-pri/T123-F123ABC/download/notes.txt"
	const content = "meeting notes\n"
	uploader := &slackAttachmentUploaderStub{attachment: uuid.New()}
	client := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.Header.Get("Authorization") != "Bearer xoxb-test" {
			t.Errorf("missing provider authorization on %s", req.URL)
		}
		switch req.URL.Host {
		case "slack.com":
			switch req.URL.Path {
			case "/api/files.info":
				if req.URL.Query().Get("file") != fileID {
					t.Errorf("unexpected files.info URL: %s", req.URL)
				}
				return slackFileInfoResponse(t, fileID, "../notes.txt", "hosted", false, int64(len(content)), downloadURL), nil
			case "/api/conversations.replies":
				query := req.URL.Query()
				if query.Get("channel") != "C123" || query.Get("ts") != "111.000001" || query.Get("oldest") != "111.000002" || query.Get("inclusive") != "true" || query.Get("limit") != "1" {
					t.Errorf("unexpected source lookup: %s", req.URL)
				}
				return slackThreadResponse("111.000002", fileID), nil
			default:
				t.Errorf("unexpected Slack API request: %s", req.URL)
				return nil, nil
			}
		case "files.slack.com":
			if req.URL.String() != downloadURL || req.Header.Get("Accept-Encoding") != "identity" {
				t.Errorf("unexpected file download request: %s", req.URL)
			}
			return slackDownloadResponse(content, int64(len(content))), nil
		default:
			t.Errorf("unexpected provider request: %s", req.URL)
			return nil, nil
		}
	})}
	importer := NewSlackFileImporter(uploader, client)
	workspaceID, storyID, actorID, importID := uuid.New(), uuid.New(), uuid.New(), uuid.New()

	attachmentID, err := importer.Import(context.Background(), "xoxb-test", fileID, workspaceID, storyID, actorID, importID, SlackFileSource{
		Kind: SlackFileSourceMessage, ChannelID: "C123", ThreadTS: "111.000001", MessageTS: "111.000002",
	})
	if err != nil {
		t.Fatalf("Import() error = %v", err)
	}
	if attachmentID != uploader.attachment || !uploader.called {
		t.Fatalf("attachment ID = %s, uploader called = %t", attachmentID, uploader.called)
	}
	if string(uploader.bytes) != content || uploader.filename != "notes.txt" {
		t.Fatalf("uploaded bytes = %q, filename = %q", uploader.bytes, uploader.filename)
	}
	if uploader.workspaceID != workspaceID || uploader.storyID != storyID || uploader.actorID != actorID || uploader.importID != importID {
		t.Fatalf("uploader received wrong actor or story scope")
	}
	if _, err := os.Stat(uploader.tempPath); !os.IsNotExist(err) {
		t.Fatalf("temporary file still exists after Import(): %s, stat error: %v", uploader.tempPath, err)
	}
}

func TestSlackFileImporterRejectsUnsafeOrUnusableFiles(t *testing.T) {
	const fileID = "F123ABC"
	const validURL = "https://files.slack.com/files-pri/T123-F123ABC/download/notes.txt"
	tests := []struct {
		name          string
		mode          string
		external      bool
		size          int64
		url           string
		body          string
		contentLength int64
		wantError     string
		wantDownloads int
	}{
		{name: "external file", mode: "external", external: true, size: 5, url: validURL, wantError: "only Slack-hosted files"},
		{name: "missing size", mode: "hosted", size: 0, url: validURL, wantError: "no declared size"},
		{name: "untrusted host", mode: "hosted", size: 5, url: "https://files.slack.com.evil.test/files-pri/file", wantError: "invalid private download URL"},
		{name: "wrong scheme", mode: "hosted", size: 5, url: "http://files.slack.com/files-pri/file", wantError: "invalid private download URL"},
		{name: "short body", mode: "hosted", size: 5, url: validURL, body: "abc", contentLength: -1, wantError: "incomplete body", wantDownloads: 1},
		{name: "extra body", mode: "hosted", size: 3, url: validURL, body: "abcde", contentLength: -1, wantError: "exceeds declared size", wantDownloads: 1},
		{name: "mismatched content length", mode: "hosted", size: 5, url: validURL, body: "abc", contentLength: 3, wantError: "does not match download metadata", wantDownloads: 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			downloads := 0
			client := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				if req.URL.Host == "slack.com" {
					return slackFileInfoResponse(t, fileID, "notes.txt", tt.mode, tt.external, tt.size, tt.url), nil
				}
				downloads++
				return slackDownloadResponse(tt.body, tt.contentLength), nil
			})}
			uploader := &slackAttachmentUploaderStub{attachment: uuid.New()}
			importer := NewSlackFileImporter(uploader, client)
			_, err := importer.Import(context.Background(), "xoxb-test", fileID, uuid.New(), uuid.New(), uuid.New(), uuid.New(), SlackFileSource{
				Kind: SlackFileSourceModal, SlackUserID: "U123",
			})
			if err == nil || !strings.Contains(err.Error(), tt.wantError) {
				t.Fatalf("Import() error = %v, want %q", err, tt.wantError)
			}
			if downloads != tt.wantDownloads || uploader.called {
				t.Fatalf("downloads = %d, uploader called = %t", downloads, uploader.called)
			}
		})
	}
}

func TestSlackFileImporterNeverFollowsDownloadRedirect(t *testing.T) {
	const fileID = "F123ABC"
	requests := 0
	client := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		requests++
		if req.URL.Host == "slack.com" {
			return slackFileInfoResponse(t, fileID, "notes.txt", "hosted", false, 5, "https://files.slack.com/files-pri/T123-F123ABC/download/notes.txt"), nil
		}
		if req.URL.Host != "files.slack.com" {
			t.Fatalf("bot token sent to redirected host: %s", req.URL.Host)
		}
		return &http.Response{
			StatusCode: http.StatusFound,
			Header:     http.Header{"Location": {"https://example.test/steal"}},
			Body:       io.NopCloser(strings.NewReader("")),
		}, nil
	})}
	importer := NewSlackFileImporter(&slackAttachmentUploaderStub{}, client)
	_, err := importer.Import(context.Background(), "xoxb-test", fileID, uuid.New(), uuid.New(), uuid.New(), uuid.New(), SlackFileSource{
		Kind: SlackFileSourceModal, SlackUserID: "U123",
	})
	if err == nil || !strings.Contains(err.Error(), "HTTP 302") || requests != 2 {
		t.Fatalf("Import() error = %v, request count = %d", err, requests)
	}
}

func TestSlackFileImporterRejectsUnverifiedSource(t *testing.T) {
	const fileID = "F123ABC"
	const downloadURL = "https://files.slack.com/files-pri/T123-F123ABC/download/notes.txt"
	tests := []struct {
		name        string
		source      SlackFileSource
		messageTS   string
		messageFile string
		wantError   string
	}{
		{name: "modal uploader mismatch", source: SlackFileSource{Kind: SlackFileSourceModal, SlackUserID: "UOTHER"}, wantError: "not uploaded by the submitting user"},
		{name: "message file absent", source: SlackFileSource{Kind: SlackFileSourceMessage, ChannelID: "C123", MessageTS: "111.000002"}, messageTS: "111.000002", messageFile: "FOTHER", wantError: "no longer present in the source message"},
		{name: "message timestamp mismatch", source: SlackFileSource{Kind: SlackFileSourceMessage, ChannelID: "C123", MessageTS: "111.000002"}, messageTS: "111.000003", messageFile: fileID, wantError: "no longer present in the source message"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			downloaded := false
			client := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				switch req.URL.Path {
				case "/api/files.info":
					return slackFileInfoResponse(t, fileID, "notes.txt", "hosted", false, 5, downloadURL), nil
				case "/api/conversations.replies":
					return slackThreadResponse(tt.messageTS, tt.messageFile), nil
				default:
					downloaded = true
					return slackDownloadResponse("notes", 5), nil
				}
			})}
			uploader := &slackAttachmentUploaderStub{}
			_, err := NewSlackFileImporter(uploader, client).Import(context.Background(), "xoxb-test", fileID, uuid.New(), uuid.New(), uuid.New(), uuid.New(), tt.source)
			if err == nil || !strings.Contains(err.Error(), tt.wantError) || downloaded || uploader.called {
				t.Fatalf("Import() error = %v, downloaded = %t, uploader called = %t", err, downloaded, uploader.called)
			}
		})
	}
}

func TestSlackFileImporterRechecksAccessBeforeStorage(t *testing.T) {
	const fileID = "F123ABC"
	const downloadURL = "https://files.slack.com/files-pri/T123-F123ABC/download/notes.txt"
	uploader := &slackAttachmentUploaderStub{attachment: uuid.New()}
	client := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.URL.Path == "/api/files.info" {
			return slackFileInfoResponse(t, fileID, "notes.txt", "hosted", false, 5, downloadURL), nil
		}
		return slackDownloadResponse("notes", 5), nil
	})}
	checked := false
	_, err := NewSlackFileImporter(uploader, client).ImportWithAccessCheck(
		context.Background(), "xoxb-test", fileID,
		uuid.New(), uuid.New(), uuid.New(), uuid.New(),
		SlackFileSource{Kind: SlackFileSourceModal, SlackUserID: "U123"},
		func(context.Context) error {
			checked = true
			return ErrForbidden
		},
	)
	if !checked || !errors.Is(err, ErrForbidden) || uploader.called {
		t.Fatalf("access checked = %t, import error = %v, uploader called = %t", checked, err, uploader.called)
	}
}

func slackFileInfoResponse(t *testing.T, fileID, filename, mode string, external bool, size int64, downloadURL string) *http.Response {
	t.Helper()
	body, err := json.Marshal(map[string]any{
		"ok": true,
		"file": map[string]any{
			"id":                   fileID,
			"name":                 filename,
			"user":                 "U123",
			"mode":                 mode,
			"is_external":          external,
			"size":                 size,
			"url_private_download": downloadURL,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(string(body)))}
}

func slackDownloadResponse(content string, contentLength int64) *http.Response {
	return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(content)), ContentLength: contentLength}
}

func slackThreadResponse(messageTS, fileID string) *http.Response {
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(`{"ok":true,"messages":[{"ts":"` + messageTS + `","files":[{"id":"` + fileID + `"}]}]}`)),
	}
}
