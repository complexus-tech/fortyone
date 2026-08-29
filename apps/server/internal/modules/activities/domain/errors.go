package activitiesdomain

import "errors"

var (
	ErrScopeMismatch = errors.New("activity story or actor does not belong to the workspace")
	ErrInvalidLimit  = errors.New("activity limit must be between 1 and 100")
)
