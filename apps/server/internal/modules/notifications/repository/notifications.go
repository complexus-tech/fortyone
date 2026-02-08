package notificationsrepository

import (
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

// Helper function for default preferences
func getDefaultPreferences() map[string]map[string]bool {
	return map[string]map[string]bool{
		"story_update": {
			"email":  true,
			"in_app": true,
		},
		"story_comment": {
			"email":  true,
			"in_app": true,
		},
		"comment_reply": {
			"email":  true,
			"in_app": true,
		},
		"objective_update": {
			"email":  true,
			"in_app": true,
		},
		"key_result_update": {
			"email":  true,
			"in_app": true,
		},
		"mention": {
			"email":  true,
			"in_app": true,
		},
		"reminders": {
			"email":  true,
			"in_app": true,
		},
	}
}
