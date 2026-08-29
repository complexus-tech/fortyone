package attachments

import (
	"errors"

	attachmentdomain "github.com/complexus-tech/projects-api/internal/modules/attachments/domain"
)

var (
	ErrNotFound                       = attachmentdomain.ErrNotFound
	ErrInvalidFile                    = errors.New("invalid file")
	ErrFileTooLarge                   = errors.New("file too large")
	ErrInvalidFileType                = errors.New("invalid file type")
	ErrUnauthorized                   = errors.New("unauthorized")
	ErrImageOptimizationNotApplicable = errors.New("image optimization not applicable")
	ErrImageOptimizationSkipped       = errors.New("image optimization skipped")
	ErrRetainedObjectStorageRoute     = errors.New("retained attachment object storage route is invalid")
	ErrRetainedObjectDeletion         = errors.New("retained attachment object deletion failed")
)
