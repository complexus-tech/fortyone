package slack

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

// SlackAttachmentUploader is the attachment service operation needed to
// transfer one Slack-hosted file into a story.
type SlackAttachmentUploader interface {
	UploadSlackAttachmentAndLinkToStory(context.Context, multipart.File, *multipart.FileHeader, uuid.UUID, uuid.UUID, uuid.UUID, uuid.UUID) (uuid.UUID, error)
}

type SlackFileSourceKind string

const (
	SlackFileSourceMessage SlackFileSourceKind = "message"
	SlackFileSourceModal   SlackFileSourceKind = "modal"
)

// SlackFileSource records the provider interaction that authorized this file
// selection. A message source is checked against the exact Slack message;
// a modal upload is checked against the submitting Slack user.
type SlackFileSource struct {
	Kind        SlackFileSourceKind
	ChannelID   string
	ThreadTS    string
	MessageTS   string
	SlackUserID string
}

// SlackFileImporter downloads Slack-hosted file bytes using the installation's
// bot token, then sends them through the regular attachment storage path.
type SlackFileImporter struct {
	attachments SlackAttachmentUploader
	webClient   *slackWebClient
	client      *http.Client
}

// NewSlackFileImporter accepts an HTTP client for provider transport injection.
// Production URLs remain fixed to Slack and redirects are never followed with
// the bot token, including when a custom client is supplied.
func NewSlackFileImporter(uploader SlackAttachmentUploader, client *http.Client) *SlackFileImporter {
	if client == nil {
		client = &http.Client{}
	}
	providerClient := *client
	providerClient.CheckRedirect = func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	}
	apiClient := providerClient
	if apiClient.Timeout == 0 {
		apiClient.Timeout = 12 * time.Second
	}
	return &SlackFileImporter{
		attachments: uploader,
		webClient:   newSlackWebClient(&apiClient),
		client:      &providerClient,
	}
}

// Import attaches one Slack file to one existing story. The caller is
// responsible for authorizing the actor's access to the story and Slack thread.
func (i *SlackFileImporter) Import(ctx context.Context, botToken, slackFileID string, workspaceID, storyID, actorID, importID uuid.UUID, source SlackFileSource) (uuid.UUID, error) {
	return i.importFile(ctx, botToken, slackFileID, workspaceID, storyID, actorID, importID, source, nil)
}

// ImportWithAccessCheck rechecks mutable workspace authorization after the
// download and immediately before storing the attachment.
func (i *SlackFileImporter) ImportWithAccessCheck(
	ctx context.Context,
	botToken, slackFileID string,
	workspaceID, storyID, actorID, importID uuid.UUID,
	source SlackFileSource,
	checkAccess func(context.Context) error,
) (uuid.UUID, error) {
	return i.importFile(ctx, botToken, slackFileID, workspaceID, storyID, actorID, importID, source, checkAccess)
}

func (i *SlackFileImporter) importFile(
	ctx context.Context,
	botToken, slackFileID string,
	workspaceID, storyID, actorID, importID uuid.UUID,
	source SlackFileSource,
	checkAccess func(context.Context) error,
) (uuid.UUID, error) {
	if i == nil || i.attachments == nil {
		return uuid.Nil, errors.New("Slack file importer is not configured")
	}
	if workspaceID == uuid.Nil || storyID == uuid.Nil || actorID == uuid.Nil || importID == uuid.Nil {
		return uuid.Nil, errors.New("workspace, story, actor, and import ID are required to import a Slack file")
	}

	file, header, cleanup, err := i.Download(ctx, botToken, slackFileID, source)
	if err != nil {
		return uuid.Nil, err
	}
	defer cleanup()

	if checkAccess != nil {
		if err := checkAccess(ctx); err != nil {
			return uuid.Nil, fmt.Errorf("recheck Slack file import access: %w", err)
		}
	}

	attachmentID, err := i.attachments.UploadSlackAttachmentAndLinkToStory(ctx, file, header, actorID, storyID, workspaceID, importID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("attach Slack file to story: %w", err)
	}
	return attachmentID, nil
}

// Download resolves a file ID against Slack and streams exactly its declared
// size into a seekable temporary file. The caller must invoke cleanup after
// uploading it, including when the upload fails.
func (i *SlackFileImporter) Download(ctx context.Context, botToken, slackFileID string, source SlackFileSource) (multipart.File, *multipart.FileHeader, func(), error) {
	if i == nil || i.webClient == nil || i.client == nil {
		return nil, nil, nil, errors.New("Slack file importer is not configured")
	}
	botToken = strings.TrimSpace(botToken)
	slackFileID = strings.TrimSpace(slackFileID)
	if botToken == "" || slackFileID == "" {
		return nil, nil, nil, errors.New("Slack bot token and file ID are required")
	}

	var info struct {
		File struct {
			ID                 string `json:"id"`
			Name               string `json:"name"`
			User               string `json:"user"`
			Mode               string `json:"mode"`
			IsExternal         bool   `json:"is_external"`
			Size               int64  `json:"size"`
			URLPrivate         string `json:"url_private"`
			URLPrivateDownload string `json:"url_private_download"`
		} `json:"file"`
	}
	if err := i.webClient.callJSON(ctx, botToken, "files.info?file="+url.QueryEscape(slackFileID), nil, &info); err != nil {
		return nil, nil, nil, fmt.Errorf("get Slack file information: %w", err)
	}
	if info.File.ID != slackFileID {
		return nil, nil, nil, errors.New("Slack returned information for a different file")
	}
	if info.File.Mode != "hosted" || info.File.IsExternal {
		return nil, nil, nil, errors.New("only Slack-hosted files can be imported")
	}
	if info.File.Size <= 0 {
		return nil, nil, nil, errors.New("Slack file has no declared size")
	}
	filename := safeSlackFilename(info.File.Name)
	if filename == "" {
		return nil, nil, nil, errors.New("Slack file has no valid filename")
	}

	downloadURL := info.File.URLPrivateDownload
	if downloadURL == "" {
		downloadURL = info.File.URLPrivate
	}
	if !isSlackPrivateFileURL(downloadURL) {
		return nil, nil, nil, errors.New("Slack file has an invalid private download URL")
	}
	if err := i.verifySource(ctx, botToken, slackFileID, info.File.User, source); err != nil {
		return nil, nil, nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, downloadURL, nil)
	if err != nil {
		return nil, nil, nil, errors.New("create Slack file download request failed")
	}
	req.Header.Set("Authorization", "Bearer "+botToken)
	req.Header.Set("Accept-Encoding", "identity")
	response, err := i.client.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			return nil, nil, nil, fmt.Errorf("download Slack file: %w", ctx.Err())
		}
		return nil, nil, nil, errors.New("download Slack file request failed")
	}
	defer response.Body.Close()
	if response.StatusCode == http.StatusTooManyRequests {
		return nil, nil, nil, &RateLimitError{Method: "files.download", RetryAfter: parseRetryAfter(response.Header.Get("Retry-After"))}
	}
	if response.StatusCode != http.StatusOK {
		return nil, nil, nil, fmt.Errorf("download Slack file: HTTP %d", response.StatusCode)
	}
	if response.ContentLength >= 0 && response.ContentLength != info.File.Size {
		return nil, nil, nil, errors.New("Slack file size does not match download metadata")
	}

	file, err := os.CreateTemp("", "fortyone-slack-file-*")
	if err != nil {
		return nil, nil, nil, fmt.Errorf("create temporary Slack file: %w", err)
	}
	var cleanupOnce sync.Once
	cleanup := func() {
		cleanupOnce.Do(func() {
			_ = file.Close()
			_ = os.Remove(file.Name())
		})
	}
	success := false
	defer func() {
		if !success {
			cleanup()
		}
	}()

	written, err := io.CopyN(file, response.Body, info.File.Size)
	if err != nil || written != info.File.Size {
		return nil, nil, nil, fmt.Errorf("download Slack file: incomplete body (%d of %d bytes): %w", written, info.File.Size, err)
	}
	extraCount, err := io.CopyN(io.Discard, response.Body, 1)
	if extraCount != 0 || (err != nil && !errors.Is(err, io.EOF)) {
		return nil, nil, nil, errors.New("download Slack file: body exceeds declared size")
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return nil, nil, nil, fmt.Errorf("rewind downloaded Slack file: %w", err)
	}

	success = true
	return file, &multipart.FileHeader{Filename: filename, Size: written}, cleanup, nil
}

func (i *SlackFileImporter) verifySource(ctx context.Context, botToken, fileID, fileUser string, source SlackFileSource) error {
	switch source.Kind {
	case SlackFileSourceModal:
		if source.SlackUserID == "" || fileUser == "" || fileUser != source.SlackUserID {
			return errors.New("Slack file was not uploaded by the submitting user")
		}
		return nil
	case SlackFileSourceMessage:
		if source.ChannelID == "" || source.MessageTS == "" {
			return errors.New("Slack file source message is incomplete")
		}
		threadTS := source.ThreadTS
		if threadTS == "" {
			threadTS = source.MessageTS
		}
		query := url.Values{}
		query.Set("channel", source.ChannelID)
		query.Set("ts", threadTS)
		query.Set("oldest", source.MessageTS)
		query.Set("inclusive", "true")
		query.Set("limit", "1")
		var thread struct {
			Messages []struct {
				TS    string `json:"ts"`
				Files []struct {
					ID string `json:"id"`
				} `json:"files"`
			} `json:"messages"`
		}
		if err := i.webClient.callJSON(ctx, botToken, "conversations.replies?"+query.Encode(), nil, &thread); err != nil {
			return fmt.Errorf("verify Slack file source message: %w", err)
		}
		for _, message := range thread.Messages {
			if message.TS != source.MessageTS {
				continue
			}
			for _, file := range message.Files {
				if file.ID == fileID {
					return nil
				}
			}
		}
		return errors.New("Slack file is no longer present in the source message")
	default:
		return errors.New("Slack file source is required")
	}
}

func safeSlackFilename(name string) string {
	name = path.Base(strings.ReplaceAll(name, "\\", "/"))
	name = strings.Map(func(r rune) rune {
		if r < 0x20 || r == 0x7f {
			return -1
		}
		return r
	}, name)
	if name == "." || name == "/" {
		return ""
	}
	return strings.TrimSpace(name)
}

func isSlackPrivateFileURL(rawURL string) bool {
	parsed, err := url.Parse(rawURL)
	return err == nil && parsed.Scheme == "https" && parsed.Host == "files.slack.com" &&
		parsed.User == nil && parsed.Fragment == "" && strings.HasPrefix(parsed.Path, "/files-pri/")
}
