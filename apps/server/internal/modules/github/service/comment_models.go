package github

import (
	"context"
	"regexp"
	"time"
)

const avatarURLExpiry = 24 * time.Hour

// resolveAvatarURL converts a stored avatar blob name to a signed URL.
// If the avatar is already a full URL or the resolver is not configured, it returns as-is.
func (s *Service) resolveAvatarURL(ctx context.Context, avatar *string) string {
	if avatar == nil || *avatar == "" {
		return ""
	}
	if s.avatars == nil {
		return *avatar
	}
	resolved, err := s.avatars.ResolveProfileImageURL(ctx, *avatar, avatarURLExpiry)
	if err != nil {
		return ""
	}
	return resolved
}

// GitHubComment represents a single comment fetched from the GitHub API.
type GitHubComment struct {
	ID         int64  `json:"id"`
	Body       string `json:"body"`
	UserLogin  string `json:"userLogin"`
	UserAvatar string `json:"userAvatar"`
	CreatedAt  string `json:"createdAt"`
	UpdatedAt  string `json:"updatedAt"`
	HTMLURL    string `json:"htmlUrl"`
}

// Regex to parse bot attribution: **Name** commented via/on FortyOne:\n\nbody
var fortyOneCommentPattern = regexp.MustCompile(`(?s)\A\*\*(.+?)\*\* commented (?:via|on) FortyOne:\s*\n\n?(.*)`)
var fortyOneCommentMarkerPattern = regexp.MustCompile(`(?m)\n*\s*<!--\s*fortyone:comment:[0-9a-fA-F-]{36}\s*-->\s*`)
