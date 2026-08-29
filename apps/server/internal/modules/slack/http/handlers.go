package slackhttp

import (
	slack "github.com/complexus-tech/projects-api/internal/modules/slack/service"
	"github.com/complexus-tech/projects-api/pkg/logger"
)

const maxSlackRequestBodyBytes = 1 << 20
const maxConcurrentSlackRequestLogs = 32

type Handlers struct {
	log             *logger.Logger
	service         *slack.Service
	requestLogSlots chan struct{}
}

func New(log *logger.Logger, service *slack.Service) *Handlers {
	return &Handlers{
		log:             log,
		service:         service,
		requestLogSlots: make(chan struct{}, maxConcurrentSlackRequestLogs),
	}
}
