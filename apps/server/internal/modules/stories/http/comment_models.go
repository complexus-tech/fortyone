package storieshttp

import (
	"fmt"
	"strings"

	attachments "github.com/complexus-tech/projects-api/internal/modules/attachments/service"
	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
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
	var creationKey *string
	if a.IdempotencyKey != nil {
		trimmedKey := strings.TrimSpace(*a.IdempotencyKey)
		if trimmedKey != "" {
			key := fmt.Sprintf("app:%s:%s", userID, trimmedKey)
			creationKey = &key
		}
	}

	return stories.CoreNewStory{
		Title:                    a.Title,
		EstimateValue:            a.EstimateValue,
		EstimatedDurationMinutes: a.EstimatedDurationMinutes,
		MinimumFocusBlockMinutes: a.MinimumFocusBlockMinutes,
		AutoSchedulingEnabled:    a.AutoSchedulingEnabled,
		Description:              a.Description,
		DescriptionHTML:          a.DescriptionHTML,
		Parent:                   a.Parent,
		Objective:                a.Objective,
		Status:                   a.Status,
		Assignee:                 a.Assignee,
		Reporter:                 &userID,
		Priority:                 a.Priority,
		Sprint:                   a.Sprint,
		KeyResult:                a.KeyResult,
		LabelIDs:                 a.LabelIDs,
		StartDate:                a.StartDate.TimePtr(),
		EndDate:                  a.EndDate.TimePtr(),
		Team:                     a.Team,
		CreationKey:              creationKey,
	}
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
