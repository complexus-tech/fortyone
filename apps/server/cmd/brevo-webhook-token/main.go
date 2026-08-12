package main

import (
	"fmt"
	"os"
	"strings"

	emailreply "github.com/complexus-tech/projects-api/internal/modules/emailreply/service"
)

func main() {
	secret := strings.TrimSpace(os.Getenv("APP_AUTH_SECRET_KEY"))
	if err := emailreply.ValidateRuntimeSecret(secret, "production"); err != nil {
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
