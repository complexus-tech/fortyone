package okractivities

import okractivitiesdomain "github.com/complexus-tech/projects-api/internal/modules/okractivities/domain"

type OKRUpdateType = okractivitiesdomain.UpdateType

const (
	UpdateTypeObjective = okractivitiesdomain.UpdateTypeObjective
	UpdateTypeKeyResult = okractivitiesdomain.UpdateTypeKeyResult
)

type OKRActivityType = okractivitiesdomain.ActivityType

const (
	ActivityTypeCreate = okractivitiesdomain.ActivityTypeCreate
	ActivityTypeUpdate = okractivitiesdomain.ActivityTypeUpdate
	ActivityTypeDelete = okractivitiesdomain.ActivityTypeDelete
)

type CoreActivity = okractivitiesdomain.Activity
type UserDetails = okractivitiesdomain.UserDetails
type CoreNewActivity = okractivitiesdomain.NewActivity

const (
	DefaultPageSize = okractivitiesdomain.DefaultPageSize
	MaximumPageSize = okractivitiesdomain.MaximumPageSize
)
