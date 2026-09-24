package attachmentsrepository

import (
	attachmentdomain "github.com/complexus-tech/projects-api/internal/modules/attachments/domain"
	attachmentssql "github.com/complexus-tech/projects-api/internal/modules/attachments/repository/sqlc"
	"github.com/google/uuid"
)

func toDomain(row attachmentssql.Attachment) attachmentdomain.Attachment {
	uploadedBy := uuid.Nil
	if row.UploadedBy != nil {
		uploadedBy = *row.UploadedBy
	}
	return attachmentdomain.Attachment{
		ID:                       row.AttachmentID,
		SlackFileImportID:        uuidValue(row.SlackFileImportID),
		Filename:                 row.Filename,
		BlobName:                 row.BlobName,
		Size:                     row.Size,
		MimeType:                 row.MimeType,
		UploadedBy:               uploadedBy,
		WorkspaceID:              row.WorkspaceID,
		CreatedAt:                row.CreatedAt,
		UpdatedAt:                row.UpdatedAt,
		ScanStatus:               attachmentdomain.ScanStatus(row.ScanStatus),
		ScanCompletedAt:          row.ScanCompletedAt,
		ScanFailureReason:        stringValue(row.ScanFailureReason),
		OptimizationStatus:       attachmentdomain.OptimizationStatus(row.OptimizationStatus),
		OptimizationAttempts:     row.OptimizationAttempts,
		OptimizationStartedAt:    row.OptimizationStartedAt,
		OptimizationCompletedAt:  row.OptimizationCompletedAt,
		OptimizationLeaseExpires: row.OptimizationLeaseExpiresAt,
		OptimizationLastError:    stringValue(row.OptimizationLastError),
	}
}

func uuidValue(value *uuid.UUID) uuid.UUID {
	if value == nil {
		return uuid.Nil
	}
	return *value
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
