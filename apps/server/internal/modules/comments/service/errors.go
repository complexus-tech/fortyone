package comments

import commentsdomain "github.com/complexus-tech/projects-api/internal/modules/comments/domain"

var (
	ErrNotFound       = commentsdomain.ErrNotFound
	ErrForbidden      = commentsdomain.ErrForbidden
	ErrInvalidComment = commentsdomain.ErrInvalidComment
	ErrInvalidMention = commentsdomain.ErrInvalidMention
)
