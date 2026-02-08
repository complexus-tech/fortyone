package commentsrepository

import (
	comments "github.com/complexus-tech/projects-api/internal/modules/comments/service"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/jmoiron/sqlx"
)

type repo struct {
	log *logger.Logger
	db  *sqlx.DB
}

func New(log *logger.Logger, db *sqlx.DB) *repo {
	return &repo{
		log: log,
		db:  db,
	}
}

// toCoreComment converts a DbComment to a CoreComment
func toCoreComment(dbComment DbComment) comments.CoreComment {
	return comments.CoreComment{
		ID:          dbComment.ID,
		StoryID:     dbComment.StoryID,
		Parent:      dbComment.Parent,
		UserID:      dbComment.UserID,
		Comment:     dbComment.Comment,
		CreatedAt:   dbComment.CreatedAt,
		UpdatedAt:   dbComment.UpdatedAt,
		SubComments: []comments.CoreComment{}, // Empty for single comment fetch
	}
}
