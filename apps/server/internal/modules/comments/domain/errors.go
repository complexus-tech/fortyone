package commentsdomain

import "errors"

var (
	ErrNotFound       = errors.New("comment not found")
	ErrForbidden      = errors.New("comment operation is not permitted")
	ErrInvalidComment = errors.New("comment input is invalid")
	ErrInvalidMention = errors.New("one or more mentioned users are unavailable")
)
