package main

import (
	"fmt"
	"os"
	"strings"

	emailreply "github.com/complexus-tech/projects-api/internal/modules/emailreply/service"
	"github.com/complexus-tech/projects-api/internal/platform/deployment"
)

func main() {
	secret := strings.TrimSpace(os.Getenv("APP_EMAIL_REPLY_SECURITY_KEY"))
	if err := deployment.ValidateProductionSecrets(deployment.Production, deployment.SecretRequirement{
		Name:            "APP_EMAIL_REPLY_SECURITY_KEY",
		Value:           secret,
		ForbiddenValues: []string{"development-only-email-reply-security-key"},
	}); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "brevo-webhook-token:", err)
		os.Exit(1)
	}
	token, err := emailreply.DeriveWebhookToken(secret)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "brevo-webhook-token:", err)
		os.Exit(1)
	}
	_, _ = fmt.Fprintf(os.Stdout, "%s: %s\n", emailreply.WebhookTokenHeader, token)
}
