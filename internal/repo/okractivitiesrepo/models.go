package okractivitiesrepo

import (
	"time"

	"github.com/complexus-tech/projects-api/internal/core/okractivities"
	"github.com/google/uuid"
)

type dbActivity struct {
	ID           uuid.UUID  `db:"activity_id"`
	ObjectiveID  uuid.UUID  `db:"objective_id"`
	KeyResultID  *uuid.UUID `db:"key_result_id"`
	UserID       uuid.UUID  `db:"user_id"`
	Type         string     `db:"activity_type"`
	UpdateType   string     `db:"update_type"`
	Field        string     `db:"field_changed"`
	CurrentValue string     `db:"current_value"`
	Comment      string     `db:"comment"`
	CreatedAt    time.Time  `db:"created_at"`
	WorkspaceID  uuid.UUID  `db:"workspace_id"`
}

func toCoreActivity(a dbActivity) okractivities.CoreActivity {
	return okractivities.CoreActivity{
		ID:           a.ID,
		ObjectiveID:  a.ObjectiveID,
		KeyResultID:  a.KeyResultID,
		UserID:       a.UserID,
		Type:         okractivities.OKRActivityType(a.Type),
		UpdateType:   okractivities.OKRUpdateType(a.UpdateType),
		Field:        a.Field,
		CurrentValue: a.CurrentValue,
		Comment:      a.Comment,
		CreatedAt:    a.CreatedAt,
		WorkspaceID:  a.WorkspaceID,
	}
}

func toCoreActivities(acts []dbActivity) []okractivities.CoreActivity {
	activities := make([]okractivities.CoreActivity, len(acts))
	for i, a := range acts {
		activities[i] = toCoreActivity(a)
	}
	return activities
}
