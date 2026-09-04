package domain

import "errors"

var (
	ErrNotFound                = errors.New("Google Drive resource was not found")
	ErrNotConnected            = errors.New("connect Google Drive to continue")
	ErrReauthorizationRequired = errors.New("Google Drive must be reconnected")
	ErrForbidden               = errors.New("Google Drive resource is outside the authorized workspace")
	ErrConflict                = errors.New("Google Drive resource changed concurrently")
	ErrAccountOwned            = errors.New("this Google account is already connected to another FortyOne user")
	ErrInvalidInput            = errors.New("Google Drive request is invalid")
	ErrOperationInProgress     = errors.New("Google Drive file creation is already in progress")
	ErrNotConfigured           = errors.New("Google Drive integration is not configured")
	ErrContentTooLarge         = errors.New("Google Drive content exceeds the supported size limit")
)
