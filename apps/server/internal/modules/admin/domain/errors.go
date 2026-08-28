package admindomain

import "errors"

var (
	ErrForbidden              = errors.New("admin access requires an active internal user")
	ErrNotFound               = errors.New("admin resource not found")
	ErrConflict               = errors.New("admin resource changed concurrently")
	ErrInvalidAction          = errors.New("invalid admin action")
	ErrInvalidFilter          = errors.New("invalid admin filter")
	ErrInvalidNote            = errors.New("admin note body is required")
	ErrInvalidPagination      = errors.New("invalid admin pagination")
	ErrReasonRequired         = errors.New("admin action reason is required")
	ErrSelfMutation           = errors.New("admin users cannot change their own access state")
	ErrInvalidTrialEndsOn     = errors.New("trial end must extend access into the future")
	ErrIntegrationUnavailable = errors.New("subscription synchronization is unavailable")
)
