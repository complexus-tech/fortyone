package developercredentialsdomain

import "errors"

var (
	ErrAccessDenied               = errors.New("developer credential access denied")
	ErrAuthenticationFailed       = errors.New("machine credential authentication failed")
	ErrConcurrentUpdate           = errors.New("developer credential state changed concurrently")
	ErrCredentialNotFound         = errors.New("developer credential not found")
	ErrCredentialRotationConflict = errors.New("developer credential has already been rotated")
	ErrExpiryRequired             = errors.New("credential expiry is required")
	ErrExpiryTooSoon              = errors.New("credential expiry must be at least one minute in the future")
	ErrExpiryTooLong              = errors.New("credential expiry cannot exceed 365 days")
	ErrInvalidCredentialKind      = errors.New("invalid developer credential kind")
	ErrInvalidName                = errors.New("name must contain between 1 and 120 characters")
	ErrInvalidReason              = errors.New("reason must contain between 1 and 240 characters")
	ErrInvalidRotationOverlap     = errors.New("invalid credential rotation overlap")
	ErrInvalidScope               = errors.New("credential contains an invalid scope")
	ErrInvalidServiceAccountRole  = errors.New("service-account role must be guest or member")
	ErrInvalidTeamRestriction     = errors.New("credential contains an invalid team restriction")
	ErrNoScopes                   = errors.New("at least one credential scope is required")
	ErrPrincipalNotFound          = errors.New("service account not found")
	ErrServiceAccountDisabled     = errors.New("service account is disabled")
	ErrTeamRestrictionNotAllowed  = errors.New("one or more team restrictions are not accessible")
	ErrTokenKeyUnavailable        = errors.New("credential digest key is unavailable")
	ErrTokenPrefixCollision       = errors.New("credential lookup prefix collision")
)
