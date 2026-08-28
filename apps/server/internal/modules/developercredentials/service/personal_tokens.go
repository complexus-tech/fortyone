package developercredentials

import (
	"context"
	"errors"
	"strings"
	"time"
	"unicode"

	developercredentialsdomain "github.com/complexus-tech/projects-api/internal/modules/developercredentials/domain"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/google/uuid"
)

const credentialIssueAttempts = 3

type CreatePersonalTokenInput struct {
	Name      string
	Scopes    []platformauth.Scope
	TeamIDs   []uuid.UUID
	ExpiresAt time.Time
	RequestID string
}

type RotatePersonalTokenInput struct {
	ExpiresAt time.Time
	RequestID string
}

type RevokeCredentialInput struct {
	Reason    string
	RequestID string
}

func (service *Service) CreatePersonalToken(
	ctx context.Context,
	access developercredentialsdomain.Access,
	input CreatePersonalTokenInput,
) (developercredentialsdomain.IssuedCredential, error) {
	if err := authorizePersonalToken(access); err != nil {
		return developercredentialsdomain.IssuedCredential{}, err
	}
	userID, err := access.Actor.UserID()
	if err != nil {
		return developercredentialsdomain.IssuedCredential{}, err
	}
	name, scopes, teamIDs, expiresAt, now, err := service.validateNewCredential(input.Name, input.Scopes, input.TeamIDs, input.ExpiresAt)
	if err != nil {
		return developercredentialsdomain.IssuedCredential{}, err
	}
	principalCandidateID, err := service.nextID()
	if err != nil {
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
		issued, err := service.tokens.issue(developercredentialsdomain.CredentialPersonalAccessToken, credentialID)
		if err != nil {
			return developercredentialsdomain.IssuedCredential{}, err
		}
		credential, err := service.repository.CreatePersonalAccessToken(ctx, developercredentialsdomain.CreatePersonalToken{
			PrincipalCandidateID: principalCandidateID,
			Credential:           issued.Material,
			WorkspaceID:          access.WorkspaceID,
			UserID:               userID,
			Name:                 name,
			Scopes:               scopes,
			TeamIDs:              teamIDs,
			ExpiresAt:            expiresAt,
			CreatedAt:            now,
			Audit: auditEvent(auditID, access, "personal_token.created", "api_credential", credentialID, input.RequestID, now,
				len(scopes), len(teamIDs)),
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

func (service *Service) ListPersonalTokens(
	ctx context.Context,
	access developercredentialsdomain.Access,
) ([]developercredentialsdomain.Credential, error) {
	if err := authorizePersonalToken(access); err != nil {
		return nil, err
	}
	userID, err := access.Actor.UserID()
	if err != nil {
		return nil, err
	}
	return service.repository.ListPersonalAccessTokens(ctx, access.WorkspaceID, userID)
}

func (service *Service) RotatePersonalToken(
	ctx context.Context,
	access developercredentialsdomain.Access,
	credentialID uuid.UUID,
	input RotatePersonalTokenInput,
) (developercredentialsdomain.IssuedCredential, error) {
	if err := authorizePersonalToken(access); err != nil {
		return developercredentialsdomain.IssuedCredential{}, err
	}
	if credentialID == uuid.Nil {
		return developercredentialsdomain.IssuedCredential{}, developercredentialsdomain.ErrCredentialNotFound
	}
	userID, err := access.Actor.UserID()
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
		issued, err := service.tokens.issue(developercredentialsdomain.CredentialPersonalAccessToken, newCredentialID)
		if err != nil {
			return developercredentialsdomain.IssuedCredential{}, err
		}
		rotatedFrom := credentialID
		credential, err := service.repository.RotatePersonalAccessToken(ctx, developercredentialsdomain.RotateCredential{
			OldCredentialID: credentialID,
			NewCredential:   issued.Material,
			WorkspaceID:     access.WorkspaceID,
			ActorUserID:     userID,
			ExpiresAt:       expiresAt,
			OverlapUntil:    now,
			RotatedAt:       now,
			Audit: auditEvent(auditID, access, "personal_token.rotated", "api_credential", newCredentialID, input.RequestID, now,
				0, 0, withRotatedFrom(&rotatedFrom)),
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

func (service *Service) RevokePersonalToken(
	ctx context.Context,
	access developercredentialsdomain.Access,
	credentialID uuid.UUID,
	input RevokeCredentialInput,
) error {
	if err := authorizePersonalToken(access); err != nil {
		return err
	}
	userID, err := access.Actor.UserID()
	if err != nil {
		return err
	}
	reason, err := normalizeReason(input.Reason, "user_requested")
	if err != nil {
		return err
	}
	now := service.clock.Now().UTC()
	auditID, err := service.nextID()
	if err != nil {
		return err
	}
	return service.repository.RevokePersonalAccessToken(ctx, developercredentialsdomain.RevokeCredential{
		CredentialID: credentialID,
		WorkspaceID:  access.WorkspaceID,
		ActorUserID:  userID,
		Reason:       reason,
		RevokedAt:    now,
		Audit:        auditEvent(auditID, access, "personal_token.revoked", "api_credential", credentialID, input.RequestID, now, 0, 0),
	})
}

func (service *Service) validateNewCredential(
	name string,
	scopes []platformauth.Scope,
	teamIDs []uuid.UUID,
	expiresAt time.Time,
) (string, []platformauth.Scope, []uuid.UUID, time.Time, time.Time, error) {
	now := service.clock.Now().UTC()
	name, err := normalizeName(name)
	if err != nil {
		return "", nil, nil, time.Time{}, time.Time{}, err
	}
	scopes, err = normalizeScopes(scopes)
	if err != nil {
		return "", nil, nil, time.Time{}, time.Time{}, err
	}
	teamIDs, err = normalizeTeamIDs(teamIDs)
	if err != nil {
		return "", nil, nil, time.Time{}, time.Time{}, err
	}
	expiresAt, err = validateExpiry(now, expiresAt)
	if err != nil {
		return "", nil, nil, time.Time{}, time.Time{}, err
	}
	return name, scopes, teamIDs, expiresAt, now, nil
}

type auditOption func(*developercredentialsdomain.AuditEvent)

func withRotatedFrom(credentialID *uuid.UUID) auditOption {
	return func(event *developercredentialsdomain.AuditEvent) { event.RotatedFromID = credentialID }
}

func auditEvent(
	id uuid.UUID,
	access developercredentialsdomain.Access,
	operation string,
	subjectType string,
	subjectID uuid.UUID,
	requestID string,
	createdAt time.Time,
	scopeCount int,
	teamCount int,
	options ...auditOption,
) developercredentialsdomain.AuditEvent {
	event := developercredentialsdomain.AuditEvent{
		ID:            id,
		WorkspaceID:   access.WorkspaceID,
		Actor:         access.Actor,
		Operation:     operation,
		SubjectType:   subjectType,
		SubjectID:     subjectID,
		Result:        "succeeded",
		RequestID:     inputSafeRequestID(requestID),
		ScopeCount:    scopeCount,
		TeamCount:     teamCount,
		WorkspaceRole: access.WorkspaceRole,
		CreatedAt:     createdAt,
	}
	for _, option := range options {
		option(&event)
	}
	return event
}

func inputSafeRequestID(value string) string {
	value = strings.Map(func(character rune) rune {
		if unicode.IsControl(character) {
			return -1
		}
		return character
	}, strings.TrimSpace(value))
	runes := []rune(value)
	if len(runes) > 128 {
		value = string(runes[:128])
	}
	return value
}
