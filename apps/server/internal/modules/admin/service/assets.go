package admin

import (
	"context"
	"strings"
	"time"
)

const adminAssetURLExpiry = 24 * time.Hour

func (service *Service) resolveUserAvatar(ctx context.Context, user *UserSummary) {
	if service.assetResolver == nil || user == nil || strings.TrimSpace(user.AvatarURL) == "" {
		return
	}
	resolved, err := service.assetResolver.ResolveProfileImageURL(ctx, user.AvatarURL, adminAssetURLExpiry)
	if err != nil {
		user.AvatarURL = ""
		return
	}
	user.AvatarURL = resolved
}

func (service *Service) resolveWorkspaceLogo(ctx context.Context, workspace *WorkspaceSummary) {
	if service.assetResolver == nil || workspace == nil || workspace.AvatarURL == nil || strings.TrimSpace(*workspace.AvatarURL) == "" {
		return
	}
	resolved, err := service.assetResolver.ResolveWorkspaceLogoURL(ctx, *workspace.AvatarURL, adminAssetURLExpiry)
	if err != nil {
		workspace.AvatarURL = nil
		return
	}
	workspace.AvatarURL = &resolved
}
