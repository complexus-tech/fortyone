package notificationshttp

import (
	"context"
	"strings"

	notifications "github.com/complexus-tech/projects-api/internal/modules/notifications/service"
	users "github.com/complexus-tech/projects-api/internal/modules/users/service"
	"github.com/google/uuid"
)

func (handlers *Handlers) resolveNotificationActors(ctx context.Context, items []notifications.CoreNotification) {
	if handlers.users == nil || len(items) == 0 {
		return
	}
	actorIDs := make([]uuid.UUID, 0, len(items))
	seen := make(map[uuid.UUID]struct{}, len(items))
	for _, notification := range items {
		if notification.ActorID == uuid.Nil {
			continue
		}
		if _, exists := seen[notification.ActorID]; exists {
			continue
		}
		seen[notification.ActorID] = struct{}{}
		actorIDs = append(actorIDs, notification.ActorID)
	}
	actors, err := handlers.users.GetUsersByIDs(ctx, actorIDs)
	if err != nil {
		if handlers.log != nil {
			handlers.log.Warn(ctx, "failed to resolve notification actors", "error", err)
		}
		return
	}
	actorsByID := make(map[uuid.UUID]users.CoreUser, len(actors))
	for _, actor := range actors {
		actor.AvatarURL = handlers.resolveNotificationAvatarURL(ctx, actor.AvatarURL)
		actorsByID[actor.ID] = actor
	}
	for index := range items {
		actor, exists := actorsByID[items[index].ActorID]
		if !exists {
			continue
		}
		items[index].Actor = &notifications.CoreNotificationActor{
			ID: actor.ID, Username: actor.Username, FullName: actor.FullName,
			AvatarURL: actor.AvatarURL, IsActive: actor.IsActive, IsSystem: actor.IsSystem,
		}
	}
}

func (handlers *Handlers) resolveNotificationAvatarURL(ctx context.Context, avatar string) string {
	if strings.TrimSpace(avatar) == "" || handlers.profileImages == nil {
		return avatar
	}
	resolved, err := handlers.profileImages.ResolveProfileImageURL(ctx, avatar, notificationAvatarExpiry)
	if err != nil {
		if handlers.log != nil {
			handlers.log.Warn(ctx, "failed to resolve notification actor avatar", "error", err)
		}
		return ""
	}
	return resolved
}

func (handlers *Handlers) resolvePortalNotificationAvatars(ctx context.Context, items []notifications.CorePortalNotification) {
	resolved := make(map[string]*string)
	for index := range items {
		avatar := items[index].ActorAvatar
		if avatar == nil || strings.TrimSpace(*avatar) == "" {
			items[index].ActorAvatar = nil
			continue
		}
		key := strings.TrimSpace(*avatar)
		if cached, ok := resolved[key]; ok {
			items[index].ActorAvatar = cached
			continue
		}
		if handlers.profileImages == nil {
			resolved[key] = nil
			items[index].ActorAvatar = nil
			continue
		}
		avatarURL, err := handlers.profileImages.ResolveProfileImageURL(ctx, key, notificationAvatarExpiry)
		if err != nil || strings.TrimSpace(avatarURL) == "" {
			if err != nil && handlers.log != nil {
				handlers.log.Warn(ctx, "failed to resolve portal notification actor avatar", "error", err)
			}
			resolved[key] = nil
			items[index].ActorAvatar = nil
			continue
		}
		resolved[key] = &avatarURL
		items[index].ActorAvatar = &avatarURL
	}
}
