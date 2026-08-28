package labelsdomain

import "errors"

var (
	ErrNotFound          = errors.New("label not found")
	ErrInvalidPagination = errors.New("label pagination is outside the supported range")
)
