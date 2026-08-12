package emailreply_test

import (
	emailreply "github.com/complexus-tech/projects-api/internal/modules/emailreply/service"
	messagingrepository "github.com/complexus-tech/projects-api/internal/modules/messaging/repository"
)

var _ emailreply.Store = (*messagingrepository.Repository)(nil)
var _ emailreply.ProcessorStore = (*messagingrepository.Repository)(nil)
