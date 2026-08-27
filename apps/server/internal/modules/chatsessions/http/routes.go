package chatsessionshttp

import (
	chatsessions "github.com/complexus-tech/projects-api/internal/modules/chatsessions/service"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/cache"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
)

type Config struct {
	DB        *sqlx.DB
	Log       *logger.Logger
	SecretKey string
	Validate  *validator.Validate
	Cache     *cache.Service
	Service   *chatsessions.Service
}

func Routes(cfg Config, app *web.App) {
	chatsessionsService := cfg.Service

	auth := mid.Auth(cfg.Log, cfg.SecretKey)
	gzip := mid.Gzip(cfg.Log)
	workspace := mid.Workspace(cfg.Log, cfg.DB, cfg.Cache)

	h := New(chatsessionsService, cfg.Log)

	// Chat sessions
	app.Post("/workspaces/{workspaceSlug}/chat-sessions", h.CreateSession, auth, workspace)
	app.Get("/workspaces/{workspaceSlug}/chat-sessions", h.ListSessions, auth, workspace, gzip)
	app.Get("/workspaces/{workspaceSlug}/chat-sessions/{sessionId}", h.GetSession, auth, workspace, gzip)
	app.Put("/workspaces/{workspaceSlug}/chat-sessions/{sessionId}", h.UpdateSession, auth, workspace)
	app.Delete("/workspaces/{workspaceSlug}/chat-sessions/{sessionId}", h.DeleteSession, auth, workspace)
	app.Get("/workspaces/{workspaceSlug}/chat-sessions/messages/count", h.GetUserMessageCount, auth, workspace)

	// Chat messages
	app.Post("/workspaces/{workspaceSlug}/chat-sessions/{sessionId}/messages", h.SaveMessages, auth, workspace)
	app.Post("/workspaces/{workspaceSlug}/chat-sessions/{sessionId}/message-writes/begin", h.BeginMessageWrite, auth, workspace)
	app.Post("/workspaces/{workspaceSlug}/chat-sessions/{sessionId}/message-writes/finalize", h.FinalizeMessageWrite, auth, workspace)
	app.Get("/workspaces/{workspaceSlug}/chat-sessions/{sessionId}/messages", h.GetMessages, auth, workspace, gzip)
	app.Get("/workspaces/{workspaceSlug}/chat-sessions/{sessionId}/messages/latest-assistant", h.GetLatestAssistantMessage, auth, workspace, gzip)

	// Durable Maya mutation approval execution
	app.Post("/workspaces/{workspaceSlug}/chat-sessions/{sessionId}/mutation-approvals/{toolCallId}/claim", h.ClaimMutationApproval, auth, workspace)
	app.Post("/workspaces/{workspaceSlug}/chat-sessions/{sessionId}/mutation-approvals/{toolCallId}/start", h.StartMutationApproval, auth, workspace)
	app.Post("/workspaces/{workspaceSlug}/chat-sessions/{sessionId}/mutation-approvals/{toolCallId}/complete", h.CompleteMutationApproval, auth, workspace)
	app.Post("/workspaces/{workspaceSlug}/chat-sessions/{sessionId}/mutation-approvals/{toolCallId}/fail", h.FailMutationApproval, auth, workspace)
	app.Post("/workspaces/{workspaceSlug}/chat-sessions/{sessionId}/mutation-approvals/{toolCallId}/recover-output", h.RecoverMutationApprovalOutput, auth, workspace)
}
