package keyresults

import (
	keyresultsdomain "github.com/complexus-tech/projects-api/internal/modules/keyresults/domain"
	"github.com/google/uuid"
)

type CoreNewKeyResult = keyresultsdomain.NewKeyResult
type CoreKeyResult = keyresultsdomain.KeyResult
type CoreKeyResultFilters = keyresultsdomain.Filters
type KeyResultPatch = keyresultsdomain.Patch

const (
	DefaultPageSize = keyresultsdomain.DefaultPageSize
	MaximumPageSize = keyresultsdomain.MaximumPageSize
)

type CoreKeyResultWithObjective struct {
	CoreKeyResult
	ObjectiveName string    `json:"objectiveName"`
	ObjectiveID   uuid.UUID `json:"objectiveId"`
	TeamID        uuid.UUID `json:"teamId"`
	TeamName      string    `json:"teamName"`
	TeamCode      string    `json:"teamCode"`
	WorkspaceID   uuid.UUID `json:"workspaceId"`
}

type CoreKeyResultListResponse struct {
	KeyResults []CoreKeyResultWithObjective `json:"keyResults"`
	TotalCount int                          `json:"totalCount"`
	Page       int                          `json:"page"`
	PageSize   int                          `json:"pageSize"`
	HasMore    bool                         `json:"hasMore"`
}

func SetField[T any](value T) keyresultsdomain.Field[T] {
	return keyresultsdomain.SetField(value)
}

func ClearField[T any]() keyresultsdomain.Field[*T] {
	return keyresultsdomain.ClearField[T]()
}
