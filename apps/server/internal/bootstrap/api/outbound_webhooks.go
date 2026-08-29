package api

import (
	"context"
	"fmt"
	"time"

	developercredentialsdomain "github.com/complexus-tech/projects-api/internal/modules/developercredentials/domain"
	developercredentials "github.com/complexus-tech/projects-api/internal/modules/developercredentials/service"
	outboundwebhooksrepository "github.com/complexus-tech/projects-api/internal/modules/outboundwebhooks/repository"
	outboundwebhooksservice "github.com/complexus-tech/projects-api/internal/modules/outboundwebhooks/service"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/complexus-tech/projects-api/internal/platform/authorization"
	"github.com/complexus-tech/projects-api/internal/platform/safehttp"
	"github.com/google/uuid"
)

type outboundPrincipalResolver struct {
	credentials *developercredentials.Service
}

func (resolver outboundPrincipalResolver) ResolveHumanPrincipal(
	ctx context.Context,
	actor platformauth.Actor,
	workspaceID uuid.UUID,
	role authorization.WorkspaceRole,
	requestID string,
) (uuid.UUID, error) {
	if resolver.credentials == nil {
		return uuid.Nil, fmt.Errorf("developer credential principal resolver is unavailable")
	}
	return resolver.credentials.EnsureHumanPrincipal(
		ctx,
		developercredentialsdomain.Access{
			Actor: actor, WorkspaceID: workspaceID, WorkspaceRole: role,
		},
		developercredentials.EnsureHumanPrincipalInput{RequestID: requestID},
	)
}

func buildOutboundWebhookManager(
	dependencies Dependencies,
	credentials *developercredentials.Service,
) (*outboundwebhooksservice.Manager, error) {
	if dependencies.DatabasePool == nil || dependencies.CredentialVault == nil || credentials == nil {
		return nil, fmt.Errorf("outbound webhook API dependencies are required")
	}
	secrets, err := outboundwebhooksservice.NewSecretManager(dependencies.CredentialVault)
	if err != nil {
		return nil, fmt.Errorf("initialize outbound webhook secrets: %w", err)
	}
	validator, err := safehttp.New(safehttp.Config{
		Timeout:               10 * time.Second,
		TLSHandshakeTimeout:   5 * time.Second,
		ResponseHeaderTimeout: 5 * time.Second,
		MaxResponseBytes:      64 << 10,
	})
	if err != nil {
		return nil, fmt.Errorf("initialize outbound webhook endpoint validator: %w", err)
	}
	manager, err := outboundwebhooksservice.NewManager(
		outboundwebhooksrepository.New(dependencies.DatabasePool),
		outboundPrincipalResolver{credentials: credentials},
		secrets,
		validator,
	)
	if err != nil {
		return nil, fmt.Errorf("initialize outbound webhook manager: %w", err)
	}
	return manager, nil
}
