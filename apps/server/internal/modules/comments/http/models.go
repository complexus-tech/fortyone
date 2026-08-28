package commentshttp

import (
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

const maxCommentMentions = 100

type UpdateComment struct {
	Content  string      `json:"content"`
	Mentions []uuid.UUID `json:"mentions"`
}

func (input UpdateComment) Validate() error {
	if strings.TrimSpace(input.Content) == "" {
		return errors.New("comment content is required")
	}
	if len(input.Mentions) > maxCommentMentions {
		return fmt.Errorf("a comment may mention at most %d users", maxCommentMentions)
	}
	for _, userID := range input.Mentions {
		if userID == uuid.Nil {
			return errors.New("mentions must contain valid user ids")
		}
	}
	return nil
}

var (
	ErrInvalidCommentID = errors.New("comment id is not in its proper form")
)
