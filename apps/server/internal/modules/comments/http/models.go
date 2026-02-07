package commentshttp

import (
	"errors"

	"github.com/google/uuid"
)

type UpdateComment struct {
	Content  string      `json:"content"`
	Mentions []uuid.UUID `json:"mentions"`
}

var (
	ErrInvalidCommentID = errors.New("comment id is not in its proper form")
)
