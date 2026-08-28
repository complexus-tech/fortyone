package workerbootstrap

import (
	bootstrapproviders "github.com/complexus-tech/projects-api/internal/bootstrap/providers"
	commentsrepository "github.com/complexus-tech/projects-api/internal/modules/comments/repository"
	comments "github.com/complexus-tech/projects-api/internal/modules/comments/service"
	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/jackc/pgx/v5/pgxpool"
)

func buildStoryCommentCreator(log *logger.Logger, pool *pgxpool.Pool) stories.CommentCreator {
	service := comments.New(commentsrepository.New(log, pool))
	adapter, err := bootstrapproviders.NewStoryCommentCreator(service)
	if err != nil {
		panic("failed to initialize story comment adapter: " + err.Error())
	}
	return adapter
}
