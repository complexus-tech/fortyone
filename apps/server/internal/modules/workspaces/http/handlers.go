package workspaceshttp

import (
	"context"
	"time"

	workspaces "github.com/complexus-tech/projects-api/internal/modules/workspaces/service"
	"github.com/complexus-tech/projects-api/pkg/logger"
)

const workspaceLogoAccessURLExpiry = 24 * time.Hour

type Handlers struct {
	workspaces  *workspaces.Service
	log         *logger.Logger
	attachments workspaces.AttachmentsService
}

func New(
	workspacesService *workspaces.Service,
	log *logger.Logger,
	attachments workspaces.AttachmentsService,
) *Handlers {
	return &Handlers{
		workspaces: workspacesService, log: log, attachments: attachments,
	}
}

func (h *Handlers) resolveWorkspaceLogoURL(ctx context.Context, avatarURL *string) *string {
	if avatarURL == nil || *avatarURL == "" || h.attachments == nil {
		return avatarURL
	}
	resolved, err := h.attachments.ResolveWorkspaceLogoURL(ctx, *avatarURL, workspaceLogoAccessURLExpiry)
	if err != nil || resolved == "" {
		return nil
	}
	return &resolved
}

func (h *Handlers) resolveWorkspaceLogos(ctx context.Context, workspacesList []workspaces.CoreWorkspace) {
	for index := range workspacesList {
		workspacesList[index].AvatarURL = h.resolveWorkspaceLogoURL(ctx, workspacesList[index].AvatarURL)
	}
}
