package mayahttp

import (
	"errors"
	"net/http"
	"strings"
	"time"

	activities "github.com/complexus-tech/projects-api/internal/modules/activities/service"
	feedback "github.com/complexus-tech/projects-api/internal/modules/feedback/service"
	keyresults "github.com/complexus-tech/projects-api/internal/modules/keyresults/service"
	maya "github.com/complexus-tech/projects-api/internal/modules/maya/service"
	notifications "github.com/complexus-tech/projects-api/internal/modules/notifications/service"
	objectives "github.com/complexus-tech/projects-api/internal/modules/objectives/service"
	reports "github.com/complexus-tech/projects-api/internal/modules/reports/service"
	search "github.com/complexus-tech/projects-api/internal/modules/search/service"
	sprints "github.com/complexus-tech/projects-api/internal/modules/sprints/service"
	states "github.com/complexus-tech/projects-api/internal/modules/states/service"
	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	teams "github.com/complexus-tech/projects-api/internal/modules/teams/service"
	users "github.com/complexus-tech/projects-api/internal/modules/users/service"
	workspaces "github.com/complexus-tech/projects-api/internal/modules/workspaces/service"
	"github.com/complexus-tech/projects-api/pkg/cache"
	"github.com/complexus-tech/projects-api/pkg/logger"
)

const defaultPlanningWindow = 14 * 24 * time.Hour
const defaultRealtimeBaseURL = "https://api.openai.com/v1"
const defaultRealtimeModel = "gpt-realtime-2.1-mini"
const defaultRealtimeTranscriptionModel = "gpt-4o-mini-transcribe"
const defaultRealtimeVoice = "marin"
const realtimeMonthlyVoiceLimit = maya.RealtimeMonthlyVoiceLimit
const realtimeMaxSessionDuration = maya.RealtimeMaxSessionDuration

var ErrMayaAccessRequired = errors.New("maya agent is available on paid plans and active trials")
var ErrMayaRealtimeNotConfigured = errors.New("maya realtime voice is not configured")
var ErrMayaRealtimeToolNotConfigured = errors.New("maya realtime tools are not configured")
var ErrMayaRealtimeMonthlyLimitExceeded = maya.ErrRealtimeMonthlyLimitExceeded
var ErrMayaRealtimeSessionInactive = maya.ErrRealtimeSessionInactive
var ErrMayaRealtimeToolCallConflict = maya.ErrRealtimeToolCallConflict
var ErrMayaRealtimeToolCallInProgress = maya.ErrRealtimeToolCallInProgress

var realtimeStoryPriorities = map[string]struct{}{
	"No Priority": {},
	"Low":         {},
	"Medium":      {},
	"High":        {},
	"Urgent":      {},
}

type Handlers struct {
	log           *logger.Logger
	cache         *cache.Service
	service       *maya.Service
	workspaces    *workspaces.Service
	stories       *stories.Service
	states        *states.Service
	teams         *teams.Service
	users         *users.Service
	objectives    *objectives.Service
	keyResults    *keyresults.Service
	search        *search.Service
	activities    *activities.Service
	feedback      *feedback.Service
	notifications *notifications.Service
	reports       *reports.Service
	sprints       *sprints.Service
	secretKey     string
	aiAPIKey      string
	baseURL       string
	client        *http.Client
	now           func() time.Time
}

func New(cfg Config) *Handlers {
	return &Handlers{
		log:           cfg.Log,
		cache:         cfg.Cache,
		service:       cfg.Service,
		workspaces:    cfg.Workspaces,
		stories:       cfg.Stories,
		states:        cfg.States,
		teams:         cfg.Teams,
		users:         cfg.Users,
		objectives:    cfg.Objectives,
		keyResults:    cfg.KeyResults,
		search:        cfg.Search,
		activities:    cfg.Activities,
		feedback:      cfg.Feedback,
		notifications: cfg.Notifications,
		reports:       cfg.Reports,
		sprints:       cfg.Sprints,
		secretKey:     cfg.SecretKey,
		aiAPIKey:      strings.TrimSpace(cfg.AIAPIKey),
		baseURL:       defaultRealtimeBaseURL,
		client:        &http.Client{Timeout: 20 * time.Second},
		now:           time.Now,
	}
}
