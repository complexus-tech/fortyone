package okractivities

import (
	"time"

	"github.com/google/uuid"
)

// OKRUpdateType represents whether the activity is for an objective or key result
type OKRUpdateType string

const (
	UpdateTypeObjective OKRUpdateType = "objective"
	UpdateTypeKeyResult OKRUpdateType = "key_result"
)

// OKRActivityType represents the type of activity performed
type OKRActivityType string

const (
	ActivityTypeCreate OKRActivityType = "create"
	ActivityTypeUpdate OKRActivityType = "update"
	ActivityTypeDelete OKRActivityType = "delete"
)

// CoreActivity represents an OKR activity in the system
type CoreActivity struct {
	ID            uuid.UUID       `json:"id"`
	ObjectiveID   uuid.UUID       `json:"objectiveId"`
	KeyResultID   *uuid.UUID      `json:"keyResultId"`
	UserID        uuid.UUID       `json:"userId"`
	Type          OKRActivityType `json:"type"`
	UpdateType    OKRUpdateType   `json:"updateType"`
	Field         string          `json:"field"`
	CurrentValue  string          `json:"currentValue"`
	PreviousValue string          `json:"previousValue"`
	Comment       string          `json:"comment"`
	CreatedAt     time.Time       `json:"createdAt"`
	WorkspaceID   uuid.UUID       `json:"workspaceId"`
}

// CoreNewActivity represents the data needed to create a new activity
type CoreNewActivity struct {
	ObjectiveID   uuid.UUID
	KeyResultID   *uuid.UUID
	UserID        uuid.UUID
	Type          OKRActivityType
	UpdateType    OKRUpdateType
	Field         string
	CurrentValue  string
	PreviousValue string
	Comment       string
	WorkspaceID   uuid.UUID
}
