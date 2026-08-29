package storiesrepository

import (
	"strings"

	"github.com/google/uuid"
)

func isStoryInlineMediaType(mimeType string) bool {
	return strings.HasPrefix(mimeType, "image/") || mimeType == "video/mp4"
}

func storyMediaPath(storyID uuid.UUID) string {
	return "/stories/" + storyID.String() + "/media/"
}

func rewriteStoryMediaHTML(contentHTML *string, originalStoryID, duplicatedStoryID uuid.UUID) *string {
	if contentHTML == nil {
		return nil
	}
	rewritten := strings.ReplaceAll(*contentHTML, storyMediaPath(originalStoryID), storyMediaPath(duplicatedStoryID))
	return &rewritten
}
