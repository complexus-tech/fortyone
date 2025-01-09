package commentsgrp

import (
	"errors"
)

type UpdateComment struct {
	Content string `json:"content"`
}

var (
	ErrInvalidCommentID = errors.New("comment id is not in its proper form")
)
