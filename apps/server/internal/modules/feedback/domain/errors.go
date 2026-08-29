package feedbackdomain

import "errors"

var (
	ErrInvalidInput              = errors.New("invalid feedback input")
	ErrNotFound                  = errors.New("feedback resource was not found")
	ErrBoardExists               = errors.New("a feedback board already exists for this team")
	ErrAlreadyPlanned            = errors.New("feedback is already linked to a primary story")
	ErrTeamMismatch              = errors.New("feedback and story must belong to the same team")
	ErrStoryManaged              = errors.New("feedback status is managed by its linked story")
	ErrDuplicateItem             = errors.New("similar feedback has already been reported")
	ErrVersionConflict           = errors.New("feedback changed since it was reviewed")
	ErrParticipationNotAllowed   = errors.New("anonymous feedback is not allowed for this portal")
	ErrAuthenticationRequired    = errors.New("feedback account authentication is required")
	ErrContributorSessionInvalid = errors.New("feedback contributor session is invalid")
	ErrContributorBlocked        = errors.New("feedback contributor is blocked")
	ErrVerificationExpired       = errors.New("feedback verification has expired")
	ErrVerificationConsumed      = errors.New("feedback verification has already been used")
	ErrVerificationAttempts      = errors.New("feedback verification attempt limit reached")
	ErrWidgetOriginNotAllowed    = errors.New("feedback widget origin is not allowed")
	ErrWidgetAssertionInvalid    = errors.New("feedback widget identity assertion is invalid")
	ErrWidgetAssertionReplayed   = errors.New("feedback widget identity assertion was already used")
	ErrFeatureUnavailable        = errors.New("feedback contributor feature is unavailable")
	ErrMergeConflict             = errors.New("feedback cannot be merged into the requested target")
)
