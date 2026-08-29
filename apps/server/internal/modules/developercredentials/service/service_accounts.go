package developercredentials

import (
	"context"
	"errors"
	"time"

	developercredentialsdomain "github.com/complexus-tech/projects-api/internal/modules/developercredentials/domain"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/complexus-tech/projects-api/internal/platform/authorization"
	"github.com/google/uuid"
)

type CreateServiceAccountInput struct {
	Name          string
	WorkspaceRole authorization.WorkspaceRole
	RequestID     string
}

type CreateServiceAccountKeyInput struct {
	Name      string
	Scopes    []platformauth.Scope
	TeamIDs   []uuid.UUID
	ExpiresAt time.Time
	RequestID string
}

type RotateServiceAccountKeyInput struct {
	ExpiresAt time.Time
	Overlap   time.Duration
	RequestID string
}

func (service *Service) CreateServiceAccount(
	ctx context.Context,
	access developercredentialsdomain.Access,
	input CreateServiceAccountInput,
) (developercredentialsdomain.ServiceAccount, error) {
	if err := authorizeServiceAccountManagement(access); err != nil {
		return developercredentialsdomain.ServiceAccount{}, err
	}
	actorUserID, err := access.Actor.UserID()
	if err != nil {
		return developercredentialsdomain.ServiceAccount{}, err
	}
	name, err := normalizeName(input.Name)
	if err != nil {
		return developercredentialsdomain.ServiceAccount{}, err
	}
	if err := validateServiceAccountRole(input.WorkspaceRole); err != nil {
		return developercredentialsdomain.ServiceAccount{}, err
	}
	principalID, err := service.nextID()
	if err != nil {
		return developercredentialsdomain.ServiceAccount{}, err
	}
	auditID, err := service.nextID()
	if err != nil {
		return developercredentialsdomain.ServiceAccount{}, err
	}
	now := service.clock.Now().UTC()
	event := auditEvent(auditID, access, "service_account.created", "principal", principalID, input.RequestID, now, 0, 0)
	event.WorkspaceRole = input.WorkspaceRole
	return service.repository.CreateServiceAccount(ctx, developercredentialsdomain.CreateServiceAccount{
		PrincipalID:   principalID,
		WorkspaceID:   access.WorkspaceID,
		ActorUserID:   actorUserID,
		Name:          name,
		WorkspaceRole: input.WorkspaceRole,
		CreatedAt:     now,
		Audit:         event,
	})
}

func (service *Service) ListServiceAccounts(
	ctx context.Context,
	access developercredentialsdomain.Access,
) ([]developercredentialsdomain.ServiceAccount, error) {
	if err := authorizeServiceAccountManagement(access); err != nil {
		return nil, err
	}
	actorUserID, err := access.Actor.UserID()
	if err != nil {
		return nil, err
	}
	return service.repository.ListServiceAccounts(ctx, access.WorkspaceID, actorUserID)
}

func (service *Service) DisableServiceAccount(
	ctx context.Context,
	access developercredentialsdomain.Access,
	principalID uuid.UUID,
	input RevokeCredentialInput,
) error {
	if err := authorizeServiceAccountManagement(access); err != nil {
		return err
	}
	actorUserID, err := access.Actor.UserID()
	if err != nil {
		return err
	}
	reason, err := normalizeReason(input.Reason, "admin_disabled")
	if err != nil {
		return err
	}
	now := service.clock.Now().UTC()
	auditID, err := service.nextID()
	if err != nil {
		return err
	}
	return service.repository.DisableServiceAccount(ctx, developercredentialsdomain.DisableServiceAccount{
		PrincipalID: principalID,
		WorkspaceID: access.WorkspaceID,
		ActorUserID: actorUserID,
		Reason:      reason,
		DisabledAt:  now,
		Audit:       auditEvent(auditID, access, "service_account.disabled", "principal", principalID, input.RequestID, now, 0, 0),
	})
}

func (service *Service) CreateServiceAccountKey(
	ctx context.Context,
	access developercredentialsdomain.Access,
	principalID uuid.UUID,
	input CreateServiceAccountKeyInput,
) (developercredentialsdomain.IssuedCredential, error) {
	if err := authorizeServiceAccountManagement(access); err != nil {
		return developercredentialsdomain.IssuedCredential{}, err
	}
	actorUserID, err := access.Actor.UserID()
	if err != nil {
		return developercredentialsdomain.IssuedCredential{}, err
	}
	name, scopes, teamIDs, expiresAt, now, err := service.validateNewCredential(input.Name, input.Scopes, input.TeamIDs, input.ExpiresAt)
	if err != nil {
		return developercredentialsdomain.IssuedCredential{}, err
	}
	if err := validateServiceAccountScopes(scopes); err != nil {
		return developercredentialsdomain.IssuedCredential{}, err
	}
	auditID, err := service.nextID()
	if err != nil {
		return developercredentialsdomain.IssuedCredential{}, err
	}
	for attempt := 0; attempt < credentialIssueAttempts; attempt++ {
		credentialID, err := service.nextID()
		if err != nil {
			return developercredentialsdomain.IssuedCredential{}, err
		}
		issued, err := service.tokens.issue(developercredentialsdomain.CredentialServiceAccountKey, credentialID)
		if err != nil {
			return developercredentialsdomain.IssuedCredential{}, err
		}
		credential, err := service.repository.CreateServiceAccountKey(ctx, developercredentialsdomain.CreateServiceAccountKey{
			Credential:  issued.Material,
			WorkspaceID: access.WorkspaceID,
			PrincipalID: principalID,
			ActorUserID: actorUserID,
			Name:        name,
			Scopes:      scopes,
			TeamIDs:     teamIDs,
			ExpiresAt:   expiresAt,
			CreatedAt:   now,
			Audit:       auditEvent(auditID, access, "service_account_key.created", "api_credential", credentialID, input.RequestID, now, len(scopes), len(teamIDs)),
		})
		if errors.Is(err, developercredentialsdomain.ErrTokenPrefixCollision) {
			continue
		}
		if err != nil {
			return developercredentialsdomain.IssuedCredential{}, err
		}
		return developercredentialsdomain.IssuedCredential{Credential: credential, Token: issued.Plaintext}, nil
	}
	return developercredentialsdomain.IssuedCredential{}, developercredentialsdomain.ErrTokenPrefixCollision
}

func (service *Service) ListServiceAccountKeys(
	ctx context.Context,
	access developercredentialsdomain.Access,
	principalID uuid.UUID,
) ([]developercredentialsdomain.Credential, error) {
	if err := authorizeServiceAccountManagement(access); err != nil {
		return nil, err
	}
	actorUserID, err := access.Actor.UserID()
	if err != nil {
		return nil, err
	}
	return service.repository.ListServiceAccountKeys(ctx, access.WorkspaceID, principalID, actorUserID)
}

func (service *Service) RotateServiceAccountKey(
	ctx context.Context,
	access developercredentialsdomain.Access,
	principalID uuid.UUID,
	credentialID uuid.UUID,
	input RotateServiceAccountKeyInput,
) (developercredentialsdomain.IssuedCredential, error) {
	if err := authorizeServiceAccountManagement(access); err != nil {
		return developercredentialsdomain.IssuedCredential{}, err
	}
	if err := validateRotationOverlap(input.Overlap, true); err != nil {
		return developercredentialsdomain.IssuedCredential{}, err
	}
	actorUserID, err := access.Actor.UserID()
	if err != nil {
		return developercredentialsdomain.IssuedCredential{}, err
	}
	now := service.clock.Now().UTC()
	expiresAt, err := validateExpiry(now, input.ExpiresAt)
	if err != nil {
		return developercredentialsdomain.IssuedCredential{}, err
	}
	auditID, err := service.nextID()
	if err != nil {
		return developercredentialsdomain.IssuedCredential{}, err
	}
	for attempt := 0; attempt < credentialIssueAttempts; attempt++ {
		newCredentialID, err := service.nextID()
		if err != nil {
			return developercredentialsdomain.IssuedCredential{}, err
		}
		issued, err := service.tokens.issue(developercredentialsdomain.CredentialServiceAccountKey, newCredentialID)
		if err != nil {
			return developercredentialsdomain.IssuedCredential{}, err
		}
		rotatedFrom := credentialID
		event := auditEvent(auditID, access, "service_account_key.rotated", "api_credential", newCredentialID, input.RequestID, now, 0, 0, withRotatedFrom(&rotatedFrom))
		event.RotationOverlap = input.Overlap
		credential, err := service.repository.RotateServiceAccountKey(ctx, developercredentialsdomain.RotateCredential{
			OldCredentialID: credentialID,
			NewCredential:   issued.Material,
			WorkspaceID:     access.WorkspaceID,
			PrincipalID:     principalID,
			ActorUserID:     actorUserID,
			ExpiresAt:       expiresAt,
			OverlapUntil:    now.Add(input.Overlap),
			RotatedAt:       now,
			Audit:           event,
		})
		if errors.Is(err, developercredentialsdomain.ErrTokenPrefixCollision) {
			continue
		}
		if err != nil {
			return developercredentialsdomain.IssuedCredential{}, err
		}
		return developercredentialsdomain.IssuedCredential{Credential: credential, Token: issued.Plaintext}, nil
	}
	return developercredentialsdomain.IssuedCredential{}, developercredentialsdomain.ErrTokenPrefixCollision
}

func (service *Service) RevokeServiceAccountKey(
	ctx context.Context,
	access developercredentialsdomain.Access,
	principalID uuid.UUID,
	credentialID uuid.UUID,
	input RevokeCredentialInput,
) error {
	if err := authorizeServiceAccountManagement(access); err != nil {
		return err
	}
	actorUserID, err := access.Actor.UserID()
	if err != nil {
		return err
	}
	reason, err := normalizeReason(input.Reason, "admin_revoked")
	if err != nil {
		return err
	}
	now := service.clock.Now().UTC()
	auditID, err := service.nextID()
	if err != nil {
		return err
	}
	return service.repository.RevokeServiceAccountKey(ctx, developercredentialsdomain.RevokeCredential{
		CredentialID: credentialID,
		WorkspaceID:  access.WorkspaceID,
		PrincipalID:  principalID,
		ActorUserID:  actorUserID,
		Reason:       reason,
		RevokedAt:    now,
		Audit:        auditEvent(auditID, access, "service_account_key.revoked", "api_credential", credentialID, input.RequestID, now, 0, 0),
	})
}
