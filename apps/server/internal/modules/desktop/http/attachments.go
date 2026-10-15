package desktophttp

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"mime"
	"mime/multipart"
	"net/textproto"
	"strings"
	"time"

	attachments "github.com/complexus-tech/projects-api/internal/modules/attachments/service"
	documents "github.com/complexus-tech/projects-api/internal/modules/documents/service"
	"github.com/google/uuid"
)

const maximumAttachmentBytes = 10 << 20

type attachmentUpload struct {
	FileName    string `json:"fileName"`
	ContentType string `json:"contentType"`
	Data        string `json:"data"`
	TargetKind  string `json:"targetKind"`
}

type memoryUpload struct{ *bytes.Reader }

func (memoryUpload) Close() error { return nil }

func decodeAttachment(raw json.RawMessage) (attachmentUpload, []byte, error) {
	var input attachmentUpload
	if err := decodePayload(raw, &input); err != nil {
		return input, nil, err
	}
	if strings.TrimSpace(input.FileName) == "" || len(input.FileName) > 255 ||
		strings.ContainsAny(input.FileName, "/\\\x00\r\n") ||
		(input.TargetKind != "story" && input.TargetKind != "document") {
		return input, nil, errInvalidOperation
	}
	if _, _, err := mime.ParseMediaType(input.ContentType); err != nil {
		return input, nil, errInvalidOperation
	}
	if len(input.Data) > base64.StdEncoding.EncodedLen(maximumAttachmentBytes) {
		return input, nil, attachments.ErrFileTooLarge
	}
	data, err := base64.StdEncoding.Strict().DecodeString(input.Data)
	if err != nil || len(data) == 0 {
		return input, nil, attachments.ErrInvalidFile
	}
	if len(data) > maximumAttachmentBytes {
		return input, nil, attachments.ErrFileTooLarge
	}
	return input, data, nil
}

func (h *Handlers) prepareAttachment(operation Operation, workspaceID, userID uuid.UUID) (mutation, error) {
	input, data, err := decodeAttachment(operation.Payload)
	if err != nil {
		return nil, err
	}
	return func(ctx context.Context) (uuid.UUID, error) {
		// Resolve permissions before touching object storage. Domain link queries
		// recheck target/workspace/actor permissions in the receipt transaction.
		if input.TargetKind == "story" {
			if err := h.Stories.AuthorizeMutation(ctx, operation.ResourceID, workspaceID); err != nil {
				return uuid.Nil, err
			}
		} else {
			document, err := h.Documents.Get(ctx, workspaceID, userID, operation.ResourceID)
			if err != nil {
				return uuid.Nil, err
			}
			if !document.CanEdit {
				return uuid.Nil, documents.ErrForbidden
			}
		}
		file := memoryUpload{bytes.NewReader(data)}
		header := &multipart.FileHeader{Filename: input.FileName, Size: int64(len(data)), Header: textproto.MIMEHeader{"Content-Type": []string{input.ContentType}}}
		if input.TargetKind == "story" {
			attachment, err := h.Attachments.UploadAndLinkToStory(ctx, file, header, userID, operation.ResourceID, workspaceID)
			return attachment.ID, err
		}
		attachment, err := h.Attachments.UploadDocumentMedia(ctx, file, header, userID, workspaceID)
		if err != nil {
			return uuid.Nil, err
		}
		err = h.Documents.LinkMedia(ctx, documents.CoreMediaInput{WorkspaceID: workspaceID, UserID: userID, DocumentID: operation.ResourceID, AttachmentID: attachment.ID})
		return attachment.ID, err
	}, nil
}

func (h *Handlers) attachmentResource(ctx context.Context, workspaceID, userID uuid.UUID, operation Operation, attachmentID uuid.UUID) (any, error) {
	var input attachmentUpload
	if err := json.Unmarshal(operation.Payload, &input); err != nil {
		return nil, errInvalidOperation
	}
	if input.TargetKind == "story" {
		if _, err := h.Stories.Get(ctx, operation.ResourceID, workspaceID); err != nil {
			return nil, err
		}
		files, err := h.Attachments.GetAttachmentsForStory(ctx, operation.ResourceID, workspaceID)
		if err != nil {
			return nil, err
		}
		for _, file := range files {
			if file.ID == attachmentID {
				return file, nil
			}
		}
		return nil, attachments.ErrNotFound
	}
	if err := h.Documents.AuthorizeMedia(ctx, documents.CoreMediaInput{WorkspaceID: workspaceID, UserID: userID, DocumentID: operation.ResourceID, AttachmentID: attachmentID}); err != nil {
		return nil, err
	}
	return h.Attachments.ResolveAttachmentAccessURL(ctx, attachmentID, workspaceID, 30*time.Minute)
}
