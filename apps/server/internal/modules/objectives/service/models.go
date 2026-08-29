package objectives

import objectivesdomain "github.com/complexus-tech/projects-api/internal/modules/objectives/domain"

type ObjectiveHealth = objectivesdomain.ObjectiveHealth

const (
	HealthAtRisk   = objectivesdomain.HealthAtRisk
	HealthOnTrack  = objectivesdomain.HealthOnTrack
	HealthOffTrack = objectivesdomain.HealthOffTrack
)

const DefaultObjectiveColor = objectivesdomain.DefaultObjectiveColor

type ObjectiveScheduleStatus = objectivesdomain.ObjectiveScheduleStatus

const (
	ScheduleStatusOnTrack    = objectivesdomain.ScheduleStatusOnTrack
	ScheduleStatusAtRisk     = objectivesdomain.ScheduleStatusAtRisk
	ScheduleStatusNoTarget   = objectivesdomain.ScheduleStatusNoTarget
	ScheduleStatusNoSchedule = objectivesdomain.ScheduleStatusNoSchedule
)

// Compatibility aliases keep the established public service API stable while
// the domain package owns transport-neutral objective types.
type CoreForecastCauseStory = objectivesdomain.ForecastCauseStory
type CoreObjective = objectivesdomain.Objective
type CoreNewObjective = objectivesdomain.NewObjective
type CoreNewKeyResult = objectivesdomain.NewKeyResult
type CoreKeyResult = objectivesdomain.KeyResult
type CoreStrategyMap = objectivesdomain.StrategyMap
type CoreStrategicPillar = objectivesdomain.StrategicPillar
type CoreStrategyUpdate = objectivesdomain.StrategyUpdate
type CoreNewStrategicPillar = objectivesdomain.NewStrategicPillar

// CoreUpdateStrategicPillar is retained for existing callers. The service
// converts it to a tri-state domain patch before persistence.
type CoreUpdateStrategicPillar struct {
	Name        *string
	Description *string
	OrderIndex  *int
}

type CoreObjectiveAnalytics = objectivesdomain.ObjectiveAnalytics
type CorePriorityBreakdown = objectivesdomain.PriorityBreakdown
type CoreProgressBreakdown = objectivesdomain.ProgressBreakdown
type CoreTeamMemberAllocation = objectivesdomain.TeamMemberAllocation
type CoreObjectiveProgressDataPoint = objectivesdomain.ObjectiveProgressDataPoint
