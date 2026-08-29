package api

import (
	developercredentials "github.com/complexus-tech/projects-api/internal/modules/developercredentials/service"
	developeroauth "github.com/complexus-tech/projects-api/internal/modules/developeroauth/service"
	invitations "github.com/complexus-tech/projects-api/internal/modules/invitations/service"
	users "github.com/complexus-tech/projects-api/internal/modules/users/service"
	"github.com/complexus-tech/projects-api/internal/platform/credentialvault"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Dependencies contains process-level infrastructure used to construct API
// services. These dependencies are intentionally kept out of HTTP route
// configuration so handlers cannot grow direct database access.
type Dependencies struct {
	DatabasePool               *pgxpool.Pool
	VerificationTokens         *users.VerificationTokenManager
	InvitationTokens           *invitations.InvitationTokenManager
	CredentialVault            *credentialvault.Vault
	DeveloperCredentialTokens  *developercredentials.TokenManager
	DeveloperOAuthPlatform     *developeroauth.Platform
	DeveloperAPIOAuth          *developeroauth.Service
	DeveloperOAuthApplications *developeroauth.ApplicationManager
}
