package api

import (
	"context"
	workspacesrepository "github.com/complexus-tech/projects-api/internal/modules/workspaces/repository"
	"github.com/complexus-tech/projects-api/internal/platform/chatrollout"
	"time"

	"github.com/complexus-tech/projects-api/internal/bootstrap/agentadapter"
	"github.com/complexus-tech/projects-api/internal/bootstrap/chatadapter"
	agentruns "github.com/complexus-tech/projects-api/internal/modules/agentinvocations/service"
	chatrepository "github.com/complexus-tech/projects-api/internal/modules/chat/repository"
	chatservice "github.com/complexus-tech/projects-api/internal/modules/chat/service"
	"github.com/complexus-tech/projects-api/internal/platform/http/mux"
	"github.com/google/uuid"
)

func buildChatService(cfg mux.Config, dependencies Dependencies, runs *agentruns.Service, mayaID uuid.UUID) (*chatservice.Service, *chatadapter.AttachmentRetirer) {
	creator := &agentadapter.ChatMessageCreator{Runs: runs, MayaID: mayaID, Tasks: cfg.TasksService, Log: cfg.Log}
	attachments := chatadapter.NewAttachmentRetirer(dependencies.DatabasePool, cfg.StorageConfig.Provider, cfg.StorageConfig.AttachmentsBucket)
	repository := chatrepository.New(cfg.Log, dependencies.DatabasePool, chatrepository.WithAttachmentRetirement(attachments.Retire))
	service := chatservice.New(repository,
		chatservice.WithWorkspaceSettings(chatrollout.New(workspacesrepository.New(dependencies.DatabasePool))),
		chatservice.WithMayaActorID(mayaID), chatservice.WithMessageCreatedHook(creator.MessageCreated),
		chatservice.WithTypingBroadcaster(chatadapter.NewTypingBroadcaster(cfg.Redis, cfg.Publisher)),
		chatservice.WithAfterCommit(func(ctx context.Context) {
			wakeup, cancel := context.WithTimeout(ctx, 2*time.Second)
			defer cancel()
			if err := cfg.TasksService.EnqueueChatDelivery(wakeup); err != nil {
				cfg.Log.Error(ctx, "chat delivery wake-up failed; durable outbox will retry", "error", err)
			}
		}),
	)
	creator.Chats = service
	return service, attachments
}

func buildChatTranscriptions(cfg mux.Config, dependencies Dependencies, access chatservice.TranscriptionAccess) *chatservice.Transcriptions {
	return chatservice.NewTranscriptions(chatrepository.New(cfg.Log, dependencies.DatabasePool), access, cfg.AIAPIKey != "", func(ctx context.Context) error {
		wakeup, cancel := context.WithTimeout(ctx, 2*time.Second)
		defer cancel()
		return cfg.TasksService.EnqueueChatTranscription(wakeup)
	}, cfg.Log)
}
