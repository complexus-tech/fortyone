package workerbootstrap

import (
	"context"
	workspacesrepository "github.com/complexus-tech/projects-api/internal/modules/workspaces/repository"
	"github.com/complexus-tech/projects-api/internal/platform/chatrollout"

	chatrepository "github.com/complexus-tech/projects-api/internal/modules/chat/repository"
	chatservice "github.com/complexus-tech/projects-api/internal/modules/chat/service"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/tasks"
	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"
)

func registerChatTranscription(mux *asynq.ServeMux, scheduler scheduleRegistrar, log *logger.Logger, pool *pgxpool.Pool, cfg Config, storage chatservice.TranscriptionStorage, access chatservice.TranscriptionAccess) error {
	settings := chatrollout.New(workspacesrepository.New(pool))
	processor := &chatservice.TranscriptionProcessor{
		Settings: &settings,
		Store:    chatrepository.New(log, pool), Storage: storage, Container: cfg.Storage.AttachmentsBucket,
		Provider: chatservice.NewOpenAITranscriber(cfg.AIAPIKey, cfg.AudioTranscriptionModel), Access: access,
	}
	mux.HandleFunc(tasks.TypeChatTranscription, func(ctx context.Context, _ *asynq.Task) error {
		return processor.ProcessPending(ctx)
	})
	_, err := scheduler.Register("@every 5s", asynq.NewTask(tasks.TypeChatTranscription, nil), tasks.ChatTranscriptionOptions()...)
	return err
}
