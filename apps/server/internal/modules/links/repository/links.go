package linksrepository

import (
	linksql "github.com/complexus-tech/projects-api/internal/modules/links/repository/sqlc"
	"github.com/complexus-tech/projects-api/pkg/logger"
)

type repo struct {
	log     *logger.Logger
	queries linksql.Querier
}

func New(log *logger.Logger, db linksql.DBTX) *repo {
	return newWithQueries(log, linksql.New(db))
}

func newWithQueries(log *logger.Logger, queries linksql.Querier) *repo {
	return &repo{
		log:     log,
		queries: queries,
	}
}
