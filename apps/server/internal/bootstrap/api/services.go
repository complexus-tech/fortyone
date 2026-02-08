package api

import (
	"fmt"

	activitiesrepository "github.com/complexus-tech/projects-api/internal/modules/activities/repository"
	activities "github.com/complexus-tech/projects-api/internal/modules/activities/service"
	attachmentsrepository "github.com/complexus-tech/projects-api/internal/modules/attachments/repository"
	attachments "github.com/complexus-tech/projects-api/internal/modules/attachments/service"
	chatsessionsrepository "github.com/complexus-tech/projects-api/internal/modules/chatsessions/repository"
	chatsessions "github.com/complexus-tech/projects-api/internal/modules/chatsessions/service"
	commentsrepository "github.com/complexus-tech/projects-api/internal/modules/comments/repository"
	comments "github.com/complexus-tech/projects-api/internal/modules/comments/service"
	documentsrepository "github.com/complexus-tech/projects-api/internal/modules/documents/repository"
	documents "github.com/complexus-tech/projects-api/internal/modules/documents/service"
	epicsrepository "github.com/complexus-tech/projects-api/internal/modules/epics/repository"
	epics "github.com/complexus-tech/projects-api/internal/modules/epics/service"
	invitationsrepository "github.com/complexus-tech/projects-api/internal/modules/invitations/repository"
	invitations "github.com/complexus-tech/projects-api/internal/modules/invitations/service"
	keyresultsrepository "github.com/complexus-tech/projects-api/internal/modules/keyresults/repository"
	keyresults "github.com/complexus-tech/projects-api/internal/modules/keyresults/service"
	labelsrepository "github.com/complexus-tech/projects-api/internal/modules/labels/repository"
	labels "github.com/complexus-tech/projects-api/internal/modules/labels/service"
	linksrepository "github.com/complexus-tech/projects-api/internal/modules/links/repository"
	links "github.com/complexus-tech/projects-api/internal/modules/links/service"
	mentionsrepository "github.com/complexus-tech/projects-api/internal/modules/mentions/repository"
	notificationsrepository "github.com/complexus-tech/projects-api/internal/modules/notifications/repository"
	notifications "github.com/complexus-tech/projects-api/internal/modules/notifications/service"
	objectivesrepository "github.com/complexus-tech/projects-api/internal/modules/objectives/repository"
	objectives "github.com/complexus-tech/projects-api/internal/modules/objectives/service"
	objectivestatusrepository "github.com/complexus-tech/projects-api/internal/modules/objectivestatus/repository"
	objectivestatus "github.com/complexus-tech/projects-api/internal/modules/objectivestatus/service"
	okractivitiesrepository "github.com/complexus-tech/projects-api/internal/modules/okractivities/repository"
	okractivities "github.com/complexus-tech/projects-api/internal/modules/okractivities/service"
	reportsrepository "github.com/complexus-tech/projects-api/internal/modules/reports/repository"
	reports "github.com/complexus-tech/projects-api/internal/modules/reports/service"
	searchrepository "github.com/complexus-tech/projects-api/internal/modules/search/repository"
	search "github.com/complexus-tech/projects-api/internal/modules/search/service"
	sprintsrepository "github.com/complexus-tech/projects-api/internal/modules/sprints/repository"
	sprints "github.com/complexus-tech/projects-api/internal/modules/sprints/service"
	statesrepository "github.com/complexus-tech/projects-api/internal/modules/states/repository"
	states "github.com/complexus-tech/projects-api/internal/modules/states/service"
	storiesrepository "github.com/complexus-tech/projects-api/internal/modules/stories/repository"
	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	subscriptionsrepository "github.com/complexus-tech/projects-api/internal/modules/subscriptions/repository"
	subscriptions "github.com/complexus-tech/projects-api/internal/modules/subscriptions/service"
	teamsrepository "github.com/complexus-tech/projects-api/internal/modules/teams/repository"
	teams "github.com/complexus-tech/projects-api/internal/modules/teams/service"
	teamsettingsrepository "github.com/complexus-tech/projects-api/internal/modules/teamsettings/repository"
	teamsettings "github.com/complexus-tech/projects-api/internal/modules/teamsettings/service"
	usersrepository "github.com/complexus-tech/projects-api/internal/modules/users/repository"
	users "github.com/complexus-tech/projects-api/internal/modules/users/service"
	workspacesrepository "github.com/complexus-tech/projects-api/internal/modules/workspaces/repository"
	workspaces "github.com/complexus-tech/projects-api/internal/modules/workspaces/service"
	"github.com/complexus-tech/projects-api/internal/platform/http/mux"
)

type services struct {
	activities     *activities.Service
	attachments    *attachments.Service
	chatSessions   *chatsessions.Service
	comments       *comments.Service
	documents      *documents.Service
	epics          *epics.Service
	invitations    *invitations.Service
	keyResults     *keyresults.Service
	labels         *labels.Service
	links          *links.Service
	notifications  *notifications.Service
	objectives     *objectives.Service
	objectiveStats *objectivestatus.Service
	okrActivities  *okractivities.Service
	reports        *reports.Service
	search         *search.Service
	sprints        *sprints.Service
	states         *states.Service
	stories        *stories.Service
	subscriptions  *subscriptions.Service
	teams          *teams.Service
	teamSettings   *teamsettings.Service
	users          *users.Service
	workspaces     *workspaces.Service
}

func buildServices(cfg mux.Config) services {
	mentionsRepo := mentionsrepository.New(cfg.Log, cfg.DB)

	attachmentsService := attachments.New(
		cfg.Log,
		attachmentsrepository.New(cfg.Log, cfg.DB),
		cfg.StorageService,
		cfg.StorageConfig,
	)

	usersService := users.New(cfg.Log, usersrepository.New(cfg.Log, cfg.DB), cfg.TasksService)
	teamsService := teams.New(cfg.Log, teamsrepository.New(cfg.Log, cfg.DB))
	statesService := states.New(cfg.Log, statesrepository.New(cfg.Log, cfg.DB))
	objectiveStatusService := objectivestatus.New(cfg.Log, objectivestatusrepository.New(cfg.Log, cfg.DB))
	subscriptionsService := subscriptions.New(
		cfg.Log,
		subscriptionsrepository.New(cfg.Log, cfg.DB),
		cfg.StripeClient,
		cfg.WebhookSecret,
		cfg.TasksService,
	)
	storiesService := stories.New(cfg.Log, storiesrepository.New(cfg.Log, cfg.DB), mentionsRepo, cfg.Publisher)
	commentsService := comments.New(cfg.Log, commentsrepository.New(cfg.Log, cfg.DB), mentionsRepo)
	linksService := links.New(cfg.Log, linksrepository.New(cfg.Log, cfg.DB))
	workspacesService := workspaces.New(
		cfg.Log,
		workspacesrepository.New(cfg.Log, cfg.DB),
		cfg.DB,
		workspaces.Dependencies{
			Teams:           teamsService,
			Stories:         storiesService,
			Statuses:        statesService,
			Users:           usersService,
			ObjectiveStatus: objectiveStatusService,
			Subscriptions:   subscriptionsService,
			Attachments:     attachmentsService,
			Cache:           cfg.Cache,
			SystemUserID:    cfg.SystemUserID,
			Publisher:       cfg.Publisher,
			TasksService:    cfg.TasksService,
		},
	)

	invitationsService := invitations.New(
		invitationsrepository.New(cfg.Log, cfg.DB),
		cfg.Log,
		cfg.Publisher,
		usersService,
		workspacesService,
		teamsService,
	)

	okrActivitiesService := okractivities.New(cfg.Log, okractivitiesrepository.New(cfg.Log, cfg.DB))
	keyResultsService := keyresults.New(cfg.Log, keyresultsrepository.New(cfg.Log, cfg.DB), okrActivitiesService)
	objectivesService := objectives.New(cfg.Log, objectivesrepository.New(cfg.Log, cfg.DB), okrActivitiesService)

	return services{
		activities:     activities.New(cfg.Log, activitiesrepository.New(cfg.Log, cfg.DB)),
		attachments:    attachmentsService,
		chatSessions:   chatsessions.New(cfg.Log, chatsessionsrepository.New(cfg.Log, cfg.DB)),
		comments:       commentsService,
		documents:      documents.New(cfg.Log, documentsrepository.New(cfg.Log, cfg.DB)),
		epics:          epics.New(cfg.Log, epicsrepository.New(cfg.Log, cfg.DB)),
		invitations:    invitationsService,
		keyResults:     keyResultsService,
		labels:         labels.New(cfg.Log, labelsrepository.New(cfg.Log, cfg.DB)),
		links:          linksService,
		notifications:  notifications.New(cfg.Log, notificationsrepository.New(cfg.Log, cfg.DB), cfg.Redis, cfg.TasksService),
		objectives:     objectivesService,
		objectiveStats: objectiveStatusService,
		okrActivities:  okrActivitiesService,
		reports:        reports.New(cfg.Log, reportsrepository.New(cfg.Log, cfg.DB)),
		search:         search.New(cfg.Log, searchrepository.New(cfg.Log, cfg.DB)),
		sprints:        sprints.New(cfg.Log, sprintsrepository.New(cfg.Log, cfg.DB)),
		states:         statesService,
		stories:        storiesService,
		subscriptions:  subscriptionsService,
		teams:          teamsService,
		teamSettings:   teamsettings.New(cfg.Log, teamsettingsrepository.New(cfg.Log, cfg.DB), cfg.TasksService),
		users:          usersService,
		workspaces:     workspacesService,
	}
}

func (s services) validate() error {
	if s.activities == nil {
		return fmt.Errorf("missing service: activities")
	}
	if s.attachments == nil {
		return fmt.Errorf("missing service: attachments")
	}
	if s.chatSessions == nil {
		return fmt.Errorf("missing service: chatSessions")
	}
	if s.comments == nil {
		return fmt.Errorf("missing service: comments")
	}
	if s.documents == nil {
		return fmt.Errorf("missing service: documents")
	}
	if s.epics == nil {
		return fmt.Errorf("missing service: epics")
	}
	if s.invitations == nil {
		return fmt.Errorf("missing service: invitations")
	}
	if s.keyResults == nil {
		return fmt.Errorf("missing service: keyResults")
	}
	if s.labels == nil {
		return fmt.Errorf("missing service: labels")
	}
	if s.links == nil {
		return fmt.Errorf("missing service: links")
	}
	if s.notifications == nil {
		return fmt.Errorf("missing service: notifications")
	}
	if s.objectives == nil {
		return fmt.Errorf("missing service: objectives")
	}
	if s.objectiveStats == nil {
		return fmt.Errorf("missing service: objectiveStats")
	}
	if s.okrActivities == nil {
		return fmt.Errorf("missing service: okrActivities")
	}
	if s.reports == nil {
		return fmt.Errorf("missing service: reports")
	}
	if s.search == nil {
		return fmt.Errorf("missing service: search")
	}
	if s.sprints == nil {
		return fmt.Errorf("missing service: sprints")
	}
	if s.states == nil {
		return fmt.Errorf("missing service: states")
	}
	if s.stories == nil {
		return fmt.Errorf("missing service: stories")
	}
	if s.subscriptions == nil {
		return fmt.Errorf("missing service: subscriptions")
	}
	if s.teams == nil {
		return fmt.Errorf("missing service: teams")
	}
	if s.teamSettings == nil {
		return fmt.Errorf("missing service: teamSettings")
	}
	if s.users == nil {
		return fmt.Errorf("missing service: users")
	}
	if s.workspaces == nil {
		return fmt.Errorf("missing service: workspaces")
	}

	return nil
}
