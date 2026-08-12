package api

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLiveness(t *testing.T) {
	t.Skip("Not implemented")
}

func TestBrevoEmailReplyIngressUsesProductionDependencies(t *testing.T) {
	t.Parallel()

	servicesSource, err := os.ReadFile("services.go")
	require.NoError(t, err)
	routesSource, err := os.ReadFile("routes.go")
	require.NoError(t, err)

	require.Contains(t, string(servicesSource), "emailreply.New(cfg.SecretKey, messagingRepo, cfg.TasksService)")
	require.Contains(t, string(servicesSource), "emailReply:          emailReplyService")
	require.Contains(t, string(routesSource), "emailreplyhttp.Routes(emailreplyhttp.Config{")
	require.Contains(t, string(routesSource), "Service: svcs.emailReply")
}
