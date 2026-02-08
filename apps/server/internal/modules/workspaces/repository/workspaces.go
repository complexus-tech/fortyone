package workspacesrepository

import (
	"math/rand"

	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/jmoiron/sqlx"
)

type repo struct {
	db  *sqlx.DB
	log *logger.Logger
}

func New(log *logger.Logger, db *sqlx.DB) *repo {
	return &repo{
		db:  db,
		log: log,
	}
}

var colors = []string{
	"#FFE066", "#FF6B6B", "#C0392B", "#FFA07A", "#FFB6C1",
	"#E056FD", "#686DE0", "#E67E22", "#A8E6CF", "#9B59B6", "#8E44AD",
	"#6BCB77", "#4ECDC4", "#4A90E2", "#95A5A6", "#27AE60", "#2ECC71",
	"#30336B", "#B4A6AB", "#636E72", "#34495E", "#2C3E50",
}

// generateRandomColor returns a random color from the Colors slice
func generateRandomColor() string {
	return colors[rand.Intn(len(colors))]
}
