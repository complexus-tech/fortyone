package slack

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	slackdomain "github.com/complexus-tech/projects-api/internal/modules/slack/domain"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/google/uuid"
)

// SlackFileImportStore persists and authorizes imports independently of the
// Asynq payload. A queued task carries only the row ID.
type SlackFileImportStore interface {
	RegisterSlackFileImport(context.Context, slackdomain.RegisterFileImport) (uuid.UUID, error)
	ClaimSlackFileImport(context.Context, uuid.UUID) (slackdomain.FileImport, bool, error)
	AuthorizeSlackFileImport(context.Context, uuid.UUID) (bool, error)
	GetSlackWorkspaceByTeamID(context.Context, string) (slackdomain.Installation, error)
	CompleteSlackFileImport(context.Context, uuid.UUID, int32, uuid.UUID) error
	FailSlackFileImport(context.Context, uuid.UUID, int32) error
	CancelSlackFileImport(context.Context, uuid.UUID, int32) error
	ListRecoverableSlackFileImports(context.Context, int) ([]uuid.UUID, error)
}

type SlackFileImportTaskEnqueuer interface {
	EnqueueSlackFileImport(context.Context, uuid.UUID) error
}

type SlackFileImportDispatcher struct {
	store SlackFileImportStore
	tasks SlackFileImportTaskEnqueuer
}

func NewSlackFileImportDispatcher(store SlackFileImportStore, tasks SlackFileImportTaskEnqueuer) *SlackFileImportDispatcher {
	return &SlackFileImportDispatcher{store: store, tasks: tasks}
}

func (d *SlackFileImportDispatcher) QueueSlackFileImport(ctx context.Context, intent SlackFileImportIntent) error {
	if d == nil || d.store == nil || d.tasks == nil {
		return errors.New("Slack file import queue is unavailable")
	}
	if intent.WorkspaceID == uuid.Nil || intent.ActorID == uuid.Nil || intent.InstallationID == uuid.Nil ||
		intent.InstallGeneration == uuid.Nil || intent.StoryID == uuid.Nil ||
		strings.TrimSpace(intent.IdempotencyKey) == "" || strings.TrimSpace(intent.SlackTeamID) == "" ||
		strings.TrimSpace(intent.SlackUserID) == "" || !isSlackModalFileID(strings.TrimSpace(intent.FileID)) {
		return ErrInvalidInput
	}
	if (intent.ChannelID == "") != (intent.MessageTS == "") {
		return ErrInvalidInput
	}
	if intent.ChannelID != "" && intent.ThreadTS == "" {
		return ErrInvalidInput
	}
	importID, err := d.store.RegisterSlackFileImport(ctx, slackdomain.RegisterFileImport{
		IdempotencyKey: intent.IdempotencyKey, WorkspaceID: intent.WorkspaceID, ActorID: intent.ActorID,
		InstallationID: intent.InstallationID, InstallGeneration: intent.InstallGeneration,
		StoryID: intent.StoryID, SlackTeamID: intent.SlackTeamID, SlackUserID: intent.SlackUserID,
		ChannelID: intent.ChannelID, ThreadTS: intent.ThreadTS, MessageTS: intent.MessageTS,
		FileID: intent.FileID,
	})
	if err != nil {
		return fmt.Errorf("register Slack file import: %w", err)
	}
	if err := d.tasks.EnqueueSlackFileImport(ctx, importID); err != nil {
		return fmt.Errorf("enqueue Slack file import %s: %w", importID, err)
	}
	return nil
}

type SlackFileImportProcessor struct {
	log      *logger.Logger
	store    SlackFileImportStore
	tasks    SlackFileImportTaskEnqueuer
	importer *SlackFileImporter
	codec    *credentialCodec
	web      *slackWebClient
}

func NewSlackFileImportProcessor(
	log *logger.Logger,
	store SlackFileImportStore,
	tasks SlackFileImportTaskEnqueuer,
	importer *SlackFileImporter,
	vault CredentialVault,
) (*SlackFileImportProcessor, error) {
	if store == nil || tasks == nil || importer == nil {
		return nil, errors.New("Slack file import processor dependencies are required")
	}
	codec, err := newCredentialCodec(vault)
	if err != nil {
		return nil, fmt.Errorf("configure Slack file import credentials: %w", err)
	}
	return &SlackFileImportProcessor{
		log: log, store: store, tasks: tasks, importer: importer, codec: codec,
		web: newSlackWebClient(&http.Client{Timeout: 8 * time.Second}),
	}, nil
}

func (p *SlackFileImportProcessor) ProcessSlackFileImport(ctx context.Context, importID uuid.UUID) error {
	if p == nil || importID == uuid.Nil {
		return errors.New("Slack file import processor or ID is invalid")
	}
	record, claimed, err := p.store.ClaimSlackFileImport(ctx, importID)
	if err != nil {
		return fmt.Errorf("claim Slack file import: %w", err)
	}
	if !claimed {
		return nil
	}
	permitted, err := p.store.AuthorizeSlackFileImport(ctx, importID)
	if err != nil {
		return p.fail(ctx, record, "")
	}
	if !permitted {
		return p.cancel(ctx, record)
	}
	installation, err := p.store.GetSlackWorkspaceByTeamID(ctx, record.SlackTeamID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return p.cancel(ctx, record)
		}
		return p.fail(ctx, record, "")
	}
	if installation.ID != record.SlackWorkspaceID ||
		installation.WorkspaceID != record.WorkspaceID ||
		installation.InstallGeneration != record.InstallGeneration {
		return p.cancel(ctx, record)
	}
	credential, _, err := p.codec.open(slackCredentialBinding{
		WorkspaceID: record.WorkspaceID, SlackTeamID: record.SlackTeamID,
		InstallGeneration: record.InstallGeneration,
	}, installation.BotAccessToken)
	if err != nil {
		return p.fail(ctx, record, "")
	}
	source := SlackFileSource{Kind: SlackFileSourceModal, SlackUserID: record.SlackUserID}
	if record.ChannelID != "" && record.MessageTS != "" {
		source = SlackFileSource{
			Kind: SlackFileSourceMessage, ChannelID: record.ChannelID,
			ThreadTS: record.ThreadTS, MessageTS: record.MessageTS,
		}
	}
	attachmentID, err := p.importer.ImportWithAccessCheck(
		ctx, credential.AccessToken, record.FileID,
		record.WorkspaceID, record.StoryID, record.ActorID, record.ID, source,
		func(checkCtx context.Context) error {
			current, checkErr := p.store.AuthorizeSlackFileImport(checkCtx, record.ID)
			if checkErr != nil {
				return checkErr
			}
			if !current {
				return ErrForbidden
			}
			installation, checkErr := p.store.GetSlackWorkspaceByTeamID(checkCtx, record.SlackTeamID)
			if checkErr != nil {
				return checkErr
			}
			if installation.ID != record.SlackWorkspaceID ||
				installation.WorkspaceID != record.WorkspaceID ||
				installation.InstallGeneration != record.InstallGeneration {
				return ErrForbidden
			}
			return nil
		},
	)
	if err != nil {
		if errors.Is(err, ErrForbidden) {
			return p.cancel(ctx, record)
		}
		return p.fail(ctx, record, credential.AccessToken)
	}
	transitionCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer cancel()
	if err := p.store.CompleteSlackFileImport(transitionCtx, record.ID, record.AttemptCount, attachmentID); err != nil {
		return fmt.Errorf("complete Slack file import %s: %w", record.ID, err)
	}
	if p.log != nil {
		p.log.Info(ctx, "Slack file attached", "import_id", record.ID, "story_id", record.StoryID, "attachment_id", attachmentID)
	}
	return nil
}

func (p *SlackFileImportProcessor) fail(ctx context.Context, record slackdomain.FileImport, botToken string) error {
	transitionCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer cancel()
	if err := p.store.FailSlackFileImport(transitionCtx, record.ID, record.AttemptCount); err != nil {
		return fmt.Errorf("record Slack file import failure %s: %w", record.ID, err)
	}
	if record.AttemptCount >= 8 {
		if p.log != nil {
			p.log.Error(ctx, "Slack file import exhausted retries", "import_id", record.ID)
		}
		p.notifyTerminalFailure(ctx, record, botToken)
	} else if p.log != nil {
		// Provider transport errors can contain a private URL; never log cause.
		p.log.Warn(ctx, "Slack file import will retry", "import_id", record.ID, "attempt", record.AttemptCount)
	}
	return fmt.Errorf("Slack file import %s failed", record.ID)
}

func (p *SlackFileImportProcessor) notifyTerminalFailure(ctx context.Context, record slackdomain.FileImport, botToken string) {
	if p.web == nil || botToken == "" || record.SlackUserID == "" {
		return
	}
	notifyCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 8*time.Second)
	defer cancel()
	err := p.web.callJSON(notifyCtx, botToken, "chat.postMessage", map[string]any{
		"channel":      record.SlackUserID,
		"text":         "A file selected for your FortyOne story could not be attached. Please open the story to check its attachments, then try again.",
		"unfurl_links": false,
	}, nil)
	if err != nil && p.log != nil {
		p.log.Warn(ctx, "Slack file import failure notification could not be sent", "import_id", record.ID)
	}
}

func (p *SlackFileImportProcessor) cancel(ctx context.Context, record slackdomain.FileImport) error {
	transitionCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer cancel()
	if err := p.store.CancelSlackFileImport(transitionCtx, record.ID, record.AttemptCount); err != nil {
		return fmt.Errorf("cancel Slack file import %s: %w", record.ID, err)
	}
	if p.log != nil {
		p.log.Warn(ctx, "Slack file import cancelled after access changed", "import_id", record.ID)
	}
	return nil
}

func (p *SlackFileImportProcessor) RecoverSlackFileImports(ctx context.Context) (int, error) {
	if p == nil {
		return 0, errors.New("Slack file import processor is unavailable")
	}
	ids, err := p.store.ListRecoverableSlackFileImports(ctx, 100)
	if err != nil {
		return 0, fmt.Errorf("list recoverable Slack file imports: %w", err)
	}
	recovered := 0
	for _, id := range ids {
		if err := p.tasks.EnqueueSlackFileImport(ctx, id); err != nil {
			return recovered, fmt.Errorf("enqueue recoverable Slack file import %s: %w", id, err)
		}
		recovered++
	}
	return recovered, nil
}
