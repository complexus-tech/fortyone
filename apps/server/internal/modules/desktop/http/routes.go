package desktophttp

import (
	"context"
	"encoding/json"
	"mime/multipart"
	"time"

	attachments "github.com/complexus-tech/projects-api/internal/modules/attachments/service"
	desktoprepository "github.com/complexus-tech/projects-api/internal/modules/desktop/repository"
	documents "github.com/complexus-tech/projects-api/internal/modules/documents/service"
	states "github.com/complexus-tech/projects-api/internal/modules/states/service"
	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	teams "github.com/complexus-tech/projects-api/internal/modules/teams/service"
	users "github.com/complexus-tech/projects-api/internal/modules/users/service"
	workspaces "github.com/complexus-tech/projects-api/internal/modules/workspaces/service"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/cache"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

type StoryService interface {
	AuthorizeMutation(context.Context, uuid.UUID, uuid.UUID) error
	List(context.Context, uuid.UUID, stories.CoreStoryFilters) ([]stories.CoreStoryList, error)
	Get(context.Context, uuid.UUID, uuid.UUID) (stories.CoreSingleStory, error)
	Create(context.Context, stories.CoreNewStory, uuid.UUID) (stories.CoreSingleStory, error)
	UpdatePatchIfUnchanged(context.Context, uuid.UUID, uuid.UUID, time.Time, stories.StoryPatch) error
	UpdatePatchWithMediaReconciliationIfUnchanged(context.Context, uuid.UUID, uuid.UUID, time.Time, stories.StoryPatch, []uuid.UUID) error
	GetComments(context.Context, uuid.UUID, uuid.UUID, int, int) ([]stories.CoreComment, bool, error)
	CreateComment(context.Context, uuid.UUID, stories.CoreNewComment) (stories.CoreComment, error)
	GetComment(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) (stories.CoreComment, error)
}
type DocumentService interface {
	List(context.Context, documents.CoreListInput) ([]documents.CoreDocumentSummary, error)
	Get(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) (documents.CoreDocument, error)
	GetForOffline(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) (documents.CoreDocument, error)
	Create(context.Context, documents.CoreCreateInput) (documents.CoreDocument, error)
	Update(context.Context, documents.CoreUpdateInput) (documents.CoreDocument, error)
	LinkMedia(context.Context, documents.CoreMediaInput) error
	AuthorizeMedia(context.Context, documents.CoreMediaInput) error
}
type AttachmentService interface {
	UploadAndLinkToStory(context.Context, multipart.File, *multipart.FileHeader, uuid.UUID, uuid.UUID, uuid.UUID) (attachments.FileInfo, error)
	UploadDocumentMedia(context.Context, multipart.File, *multipart.FileHeader, uuid.UUID, uuid.UUID) (attachments.FileInfo, error)
	ResolveAttachmentAccessURL(context.Context, uuid.UUID, uuid.UUID, time.Duration) (attachments.FileInfo, error)
	GetAttachmentsForStory(context.Context, uuid.UUID, uuid.UUID) ([]attachments.FileInfo, error)
}
type UserSettingsReader interface {
	GetUser(context.Context, uuid.UUID) (users.CoreUser, error)
	List(context.Context, uuid.UUID, users.CoreListUsersFilter) ([]users.CoreUser, error)
	GetAutomationPreferences(context.Context, uuid.UUID, uuid.UUID) (users.CoreAutomationPreferences, error)
}
type WorkspaceSettingsReader interface {
	Get(context.Context, uuid.UUID, uuid.UUID) (workspaces.CoreWorkspace, error)
	GetOrCreateWorkspaceSettings(context.Context, uuid.UUID) (workspaces.CoreWorkspaceSettings, error)
}
type TeamReader interface {
	List(context.Context, uuid.UUID, uuid.UUID, ...teams.CoreListTeamsFilter) ([]teams.CoreTeam, error)
}
type StatusReader interface {
	List(context.Context, uuid.UUID, uuid.UUID) ([]states.CoreState, error)
}
type OperationStore interface {
	Execute(context.Context, uuid.UUID, uuid.UUID, uuid.UUID, []byte, func(context.Context) (desktoprepository.Result, error)) (desktoprepository.Result, bool, error)
}
type Config struct {
	Log               *logger.Logger
	SecretKey         string
	BrowserSessions   mid.SessionResolver
	WorkspaceResolver mid.WorkspaceResolver
	Stories           StoryService
	Documents         DocumentService
	Attachments       AttachmentService
	Teams             TeamReader
	Users             UserSettingsReader
	Workspaces        WorkspaceSettingsReader
	Statuses          StatusReader
	Operations        OperationStore
	Cache             *cache.Service
}
type Handlers struct{ Config }

func Routes(cfg Config, app *web.App) {
	h := &Handlers{Config: cfg}
	auth := mid.Auth(cfg.Log, cfg.SecretKey, cfg.BrowserSessions)
	workspace := mid.Workspace(cfg.Log, cfg.WorkspaceResolver)
	member := mid.RequireMinimumRole(cfg.Log, mid.RoleMember)
	app.Get("/workspaces/{workspaceSlug}/desktop/sync", h.Sync, auth, workspace)
	app.Get("/workspaces/{workspaceSlug}/desktop/settings", h.Settings, auth, workspace)
	app.Get("/workspaces/{workspaceSlug}/desktop/stories/{id}/comments", h.Comments, auth, workspace)
	app.Post("/workspaces/{workspaceSlug}/desktop/operations", h.Apply, auth, workspace, member)
}

type Operation struct {
	ID          uuid.UUID       `json:"id"`
	Kind        string          `json:"kind"`
	ResourceID  uuid.UUID       `json:"resourceId"`
	Payload     json.RawMessage `json:"payload"`
	BaseVersion *string         `json:"baseVersion"`
}
