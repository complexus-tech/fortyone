package storieshttp

import (
	attachments "github.com/complexus-tech/projects-api/internal/modules/attachments/service"
	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	storytransport "github.com/complexus-tech/projects-api/internal/modules/stories/transport"
	"github.com/google/uuid"
)

func toAppComment(i stories.CoreComment, usersByID map[uuid.UUID]AppUserSummary) AppComment {
	return AppComment{
		ID:          i.ID,
		StoryID:     i.StoryID,
		Parent:      i.Parent,
		UserID:      i.UserID,
		User:        usersByID[i.UserID],
		Comment:     i.Comment,
		CreatedAt:   i.CreatedAt,
		UpdatedAt:   i.UpdatedAt,
		SubComments: toAppComments(i.SubComments, usersByID),
	}
}

func toAppComments(i []stories.CoreComment, usersByID map[uuid.UUID]AppUserSummary) []AppComment {
	appComments := make([]AppComment, len(i))
	for i, comment := range i {
		appComments[i] = toAppComment(comment, usersByID)
	}
	return appComments
}

func toCoreNewStory(a AppNewStory, userID uuid.UUID) stories.CoreNewStory {
	return storytransport.ToNewStory(a, userID)
}

func toAppStoryMedia(file attachments.FileInfo, stableURL string) AppStoryMedia {
	return AppStoryMedia{
		ID:         file.ID,
		Filename:   file.Filename,
		Size:       file.Size,
		MimeType:   file.MimeType,
		URL:        stableURL,
		CreatedAt:  file.CreatedAt,
		UploadedBy: file.UploadedBy,
	}
}
