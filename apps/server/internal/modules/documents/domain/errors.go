package documentdomain

import "errors"

var (
	ErrConflict      = errors.New("document changed; reload the latest version before saving")
	ErrInvalidInput  = errors.New("invalid document input")
	ErrForbidden     = errors.New("document access denied")
	ErrNotFound      = errors.New("document not found")
	ErrNotConfigured = errors.New("document repository is not configured")
)
