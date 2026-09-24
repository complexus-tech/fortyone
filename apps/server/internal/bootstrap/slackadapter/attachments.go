package slackadapter

import (
	"context"
	"mime/multipart"

	attachments "github.com/complexus-tech/projects-api/internal/modules/attachments/service"
	"github.com/google/uuid"
)

// AttachmentUploader keeps Slack's import boundary independent of the
// attachments module's concrete service and response shape.
type AttachmentUploader struct {
	service *attachments.Service
}

func NewAttachmentUploader(service *attachments.Service) *AttachmentUploader {
	return &AttachmentUploader{service: service}
}

func (a *AttachmentUploader) UploadSlackAttachmentAndLinkToStory(
	ctx context.Context,
	file multipart.File,
	header *multipart.FileHeader,
	actorID, storyID, workspaceID, importID uuid.UUID,
) (uuid.UUID, error) {
	attachment, err := a.service.UploadSlackAttachmentAndLinkToStory(
		ctx, file, header, actorID, storyID, workspaceID, importID,
	)
	if err != nil {
		return uuid.Nil, err
	}
	return attachment.ID, nil
}
