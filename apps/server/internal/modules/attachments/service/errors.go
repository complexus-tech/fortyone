package attachments

import "errors"

var (
	ErrNotFound        = errors.New("attachment not found")
	ErrInvalidFile     = errors.New("invalid file")
	ErrFileTooLarge    = errors.New("file too large")
	ErrInvalidFileType = errors.New("invalid file type")
	ErrUnauthorized    = errors.New("unauthorized")
)
