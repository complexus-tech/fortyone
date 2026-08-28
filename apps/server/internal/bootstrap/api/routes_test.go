package api

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBrevoEmailReplyIngressUsesProductionDependencies(t *testing.T) {
	t.Parallel()

	servicesSource, err := os.ReadFile("services.go")
	require.NoError(t, err)
	routesSource, err := os.ReadFile("routes.go")
	require.NoError(t, err)

	servicesText := string(servicesSource)
	require.Contains(t, servicesText, "emailreply.New(cfg.EmailReplySecurityKey, messagingRepo, cfg.TasksService)")
	require.Contains(t, strings.Join(strings.Fields(servicesText), " "), "emailReply: emailReplyService")
	require.Contains(t, string(routesSource), "emailreplyhttp.Routes(emailreplyhttp.Config{")
	require.Contains(t, string(routesSource), "Service: svcs.emailReply")
}

func TestInteractiveStoryHardDeleteReceivesCredentialFreeStorageRoute(t *testing.T) {
	t.Parallel()

	servicesSource, err := os.ReadFile("services.go")
	require.NoError(t, err)
	servicesText := strings.Join(strings.Fields(string(servicesSource)), " ")
	require.Contains(t, servicesText, "storiesrepository.WithAttachmentObjectStorage( cfg.StorageConfig.Provider, cfg.StorageConfig.AttachmentsBucket,")
	require.NotContains(t, servicesText, "storiesrepository.WithAttachmentObjectStorage( cfg.StorageConfig.Azure")
	require.NotContains(t, servicesText, "storiesrepository.WithAttachmentObjectStorage( cfg.StorageConfig.AWS")
}
