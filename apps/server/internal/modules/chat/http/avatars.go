package chathttp

import (
	"context"
	"slices"
	"strings"
	"time"

	domain "github.com/complexus-tech/projects-api/internal/modules/chat/domain"
)

const avatarAccessURLExpiry = 24 * time.Hour

type ProfileImageResolver interface {
	ResolveProfileImageURL(context.Context, string, time.Duration) (string, error)
}

// Resolve only authorized response data, keeping expiring access URLs out of
// repository values and persisted events. Repeated authors are signed once per
// response, including avatars repeated between a message and its reply summary.
func (h *handlers) resolveResponseAvatars(ctx context.Context, value any) any {
	resolvedByAvatar := make(map[string]*string)
	person := func(person domain.Person) domain.Person {
		if person.AvatarURL == nil {
			return person
		}
		key := strings.TrimSpace(*person.AvatarURL)
		person.AvatarURL = nil
		if key == "" {
			return person
		}
		if resolved, ok := resolvedByAvatar[key]; ok {
			person.AvatarURL = resolved
			return person
		}
		resolvedByAvatar[key] = nil
		if h.profileImages == nil {
			return person
		}
		resolved, err := h.profileImages.ResolveProfileImageURL(ctx, key, avatarAccessURLExpiry)
		if err != nil {
			if h.log != nil {
				h.log.Warn(ctx, "failed to resolve chat profile image", "user_id", person.ID)
			}
			return person
		}
		if strings.TrimSpace(resolved) != "" {
			resolvedByAvatar[key] = &resolved
			person.AvatarURL = &resolved
		}
		return person
	}
	people := func(values []domain.Person) []domain.Person {
		result := slices.Clone(values)
		for i := range result {
			result[i] = person(result[i])
		}
		return result
	}
	message := func(value domain.Message) domain.Message {
		value.Author = person(value.Author)
		value.ReplyAuthors = people(value.ReplyAuthors)
		return value
	}
	messages := func(values []domain.Message) []domain.Message {
		result := slices.Clone(values)
		for i := range result {
			result[i] = message(result[i])
		}
		return result
	}

	switch value := value.(type) {
	case domain.Conversation:
		value.Members = people(value.Members)
		return value
	case domain.ConversationPage:
		value.Items = slices.Clone(value.Items)
		for i := range value.Items {
			value.Items[i].Members = people(value.Items[i].Members)
		}
		return value
	case domain.MemberPage:
		value.Items = people(value.Items)
		return value
	case domain.Message:
		return message(value)
	case domain.MessagePage:
		value.Items = messages(value.Items)
		return value
	case domain.SearchPage:
		value.Items = messages(value.Items)
		return value
	case domain.UnreadThreadPage:
		value.Items = slices.Clone(value.Items)
		for i := range value.Items {
			value.Items[i].Message = message(value.Items[i].Message)
		}
		return value
	case domain.ResourcePage:
		value.Items = slices.Clone(value.Items)
		for i := range value.Items {
			value.Items[i].Author = person(value.Items[i].Author)
		}
		return value
	default:
		return value
	}
}
