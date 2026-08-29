package maya

import (
	"github.com/google/uuid"
)

type Dependencies struct {
	Repository        Repository
	Realtime          RealtimeRepository
	Stories           StoriesService
	Reports           ReportsService
	Calendar          CalendarService
	Users             UsersService
	WorkspaceSettings WorkspaceSettingsService
	Planner           Planner
	Clock             Clock
	MayaActorID       uuid.UUID
}

type Service struct {
	repo              Repository
	realtime          RealtimeRepository
	stories           StoriesService
	reports           ReportsService
	calendar          CalendarService
	users             UsersService
	workspaceSettings WorkspaceSettingsService
	planner           Planner
	clock             Clock
	mayaActorID       uuid.UUID
}

func New(dependencies Dependencies) *Service {
	clock := dependencies.Clock
	if clock == nil {
		clock = systemClock{}
	}
	return &Service{
		repo:              dependencies.Repository,
		realtime:          dependencies.Realtime,
		stories:           dependencies.Stories,
		reports:           dependencies.Reports,
		calendar:          dependencies.Calendar,
		users:             dependencies.Users,
		workspaceSettings: dependencies.WorkspaceSettings,
		planner:           dependencies.Planner,
		clock:             clock,
		mayaActorID:       dependencies.MayaActorID,
	}
}
