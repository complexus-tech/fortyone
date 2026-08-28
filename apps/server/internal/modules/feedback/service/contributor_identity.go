package feedback

import (
	"context"
	"fmt"
	"github.com/complexus-tech/projects-api/pkg/events"
	"github.com/complexus-tech/projects-api/pkg/validate"
	"github.com/google/uuid"
	"net/url"
	"strings"
	"unicode/utf8"
)

func (s *Service) RequestContributorVerification(ctx context.Context, portalSlug, email, displayName string, hideNamePublicly bool, source string) (CoreVerificationChallenge, error) {
	if err := s.requireContributorFeatures(); err != nil {
		return CoreVerificationChallenge{}, err
	}
	portal, err := s.repo.GetPortalBySlug(ctx, strings.TrimSpace(portalSlug))
	if err != nil {
		return CoreVerificationChallenge{}, err
	}
	if portal.ParticipationMode == ParticipationModeAccountRequired {
		return CoreVerificationChallenge{}, ErrParticipationNotAllowed
	}
	normalizedEmail, err := validate.Email(email)
	if err != nil {
		return CoreVerificationChallenge{}, invalidInput("a valid email address is required")
	}
	displayName = strings.TrimSpace(displayName)
	if utf8.RuneCountInString(displayName) > maxDisplayNameRunes {
		return CoreVerificationChallenge{}, invalidInputf("display name must be %d characters or fewer", maxDisplayNameRunes)
	}
	source = normalizeContributorSessionSource(source)
	if source != ContributorSessionSourcePortal && source != ContributorSessionSourceWidget {
		return CoreVerificationChallenge{}, invalidInput("verification source must be portal or widget")
	}
	publicMasked := false
	switch portal.GuestIdentityPolicy {
	case GuestIdentityPolicyAlwaysMaskGuests:
		publicMasked = true
	case GuestIdentityPolicyAllowPublicMasking:
		publicMasked = hideNamePublicly
	}
	opaqueToken, tokenHash, err := s.randomToken(32)
	if err != nil {
		return CoreVerificationChallenge{}, err
	}
	code, err := s.randomCode()
	if err != nil {
		return CoreVerificationChallenge{}, err
	}
	expiresAt := s.now().UTC().Add(verificationLifetime)
	challenge, err := s.nextRepo.CreateContributorVerification(ctx, CoreVerificationRequest{PortalID: portal.ID, Email: normalizedEmail, DisplayName: displayName, PublicMasked: publicMasked, TokenHash: tokenHash, CodeHash: s.verificationCodeHash(portal.ID, normalizedEmail, code), Source: source, ExpiresAt: expiresAt})
	if err != nil {
		return CoreVerificationChallenge{}, err
	}
	verificationURL := fmt.Sprintf("%s/portal/%s/feedback/verify?token=%s", s.websiteURL, url.PathEscape(portal.Slug), url.QueryEscape(opaqueToken))
	s.publish(ctx, events.Event{Type: events.FeedbackContributorVerification, Payload: events.FeedbackContributorVerificationPayload{Email: normalizedEmail, DisplayName: displayName, PortalName: portal.Name, VerificationURL: verificationURL, Code: code, ExpiresAt: expiresAt}, Timestamp: s.now().UTC()})
	return challenge, nil
}
func (s *Service) ConfirmContributorVerification(ctx context.Context, portalSlug, token, email, code, source string) (CoreContributorSessionResult, error) {
	if err := s.requireContributorFeatures(); err != nil {
		return CoreContributorSessionResult{}, err
	}
	portal, err := s.repo.GetPortalBySlug(ctx, strings.TrimSpace(portalSlug))
	if err != nil {
		return CoreContributorSessionResult{}, err
	}
	token = strings.TrimSpace(token)
	code = strings.TrimSpace(code)
	if (token == "") == (code == "") {
		return CoreContributorSessionResult{}, invalidInput("provide exactly one verification token or code")
	}
	var tokenHash []byte
	var codeHash []byte
	normalizedEmail := ""
	if token != "" {
		tokenHash = hashOpaqueToken(token)
	} else {
		normalizedEmail, err = validate.Email(email)
		if err != nil {
			return CoreContributorSessionResult{}, invalidInput("email is required when confirming a code")
		}
		if len(code) != 6 {
			return CoreContributorSessionResult{}, invalidInput("verification code must contain 6 digits")
		}
		codeHash = s.verificationCodeHash(portal.ID, normalizedEmail, code)
	}
	source = normalizeContributorSessionSource(source)
	if source != ContributorSessionSourcePortal && source != ContributorSessionSourceWidget {
		return CoreContributorSessionResult{}, invalidInput("verification source must be portal or widget")
	}
	sessionToken, sessionHash, err := s.randomToken(32)
	if err != nil {
		return CoreContributorSessionResult{}, err
	}
	participant, session, err := s.nextRepo.ConfirmContributorVerification(ctx, CoreVerificationConfirmation{PortalID: portal.ID, TokenHash: tokenHash, Email: normalizedEmail, CodeHash: codeHash, Source: source, SessionTokenHash: sessionHash, SessionExpiresAt: s.now().UTC().Add(contributorSessionLifetime)})
	if err != nil {
		return CoreContributorSessionResult{}, err
	}
	return CoreContributorSessionResult{Participant: maskParticipant(participant, portal.GuestIdentityPolicy), Session: session, Token: sessionToken}, nil
}
func (s *Service) ResolveContributorSession(ctx context.Context, portalSlug, authorization, source string) (CoreContributorSessionResult, error) {
	if err := s.requireContributorFeatures(); err != nil {
		return CoreContributorSessionResult{}, err
	}
	portal, err := s.repo.GetPortalBySlug(ctx, strings.TrimSpace(portalSlug))
	if err != nil {
		return CoreContributorSessionResult{}, err
	}
	token, err := parseAuthorizationToken(authorization, "FeedbackSession")
	if err != nil {
		return CoreContributorSessionResult{}, err
	}
	participant, session, err := s.nextRepo.GetContributorSession(ctx, portal.ID, hashOpaqueToken(token), normalizeContributorSessionSource(source))
	if err != nil {
		return CoreContributorSessionResult{}, normalizeSessionError(err)
	}
	if session.Source == ContributorSessionSourcePreferences || (source != "" && session.Source != normalizeContributorSessionSource(source)) {
		return CoreContributorSessionResult{}, ErrContributorSessionInvalid
	}
	if participant.BlockedAt != nil {
		return CoreContributorSessionResult{}, ErrContributorBlocked
	}
	return CoreContributorSessionResult{Participant: maskParticipant(participant, portal.GuestIdentityPolicy), Session: session}, nil
}

func (s *Service) ResolveContributorRateLimitIdentity(ctx context.Context, portalSlug, authorization string) (string, error) {
	result, err := s.ResolveContributorSession(ctx, portalSlug, authorization, "")
	if err != nil {
		return "", err
	}
	if result.Participant.ID == uuid.Nil {
		return "", ErrContributorSessionInvalid
	}
	return result.Participant.ID.String(), nil
}
func (s *Service) ResolveContributorAuthorization(ctx context.Context, portalSlug, authorization string) (CoreContributorSessionResult, error) {
	parts := strings.Fields(strings.TrimSpace(authorization))
	if len(parts) != 2 {
		return CoreContributorSessionResult{}, ErrContributorSessionInvalid
	}
	if strings.EqualFold(parts[0], "FeedbackSession") {
		return s.ResolveContributorSession(ctx, portalSlug, authorization, "")
	}
	if !strings.EqualFold(parts[0], "PreferenceSession") {
		return CoreContributorSessionResult{}, ErrContributorSessionInvalid
	}
	portal, err := s.repo.GetPortalBySlug(ctx, strings.TrimSpace(portalSlug))
	if err != nil {
		return CoreContributorSessionResult{}, err
	}
	participant, session, err := s.nextRepo.GetContributorSession(ctx, portal.ID, hashOpaqueToken(parts[1]), ContributorSessionSourcePreferences)
	if err != nil {
		return CoreContributorSessionResult{}, normalizeSessionError(err)
	}
	if participant.BlockedAt != nil {
		return CoreContributorSessionResult{}, ErrContributorBlocked
	}
	return CoreContributorSessionResult{Participant: maskParticipant(participant, portal.GuestIdentityPolicy), Session: session}, nil
}
func (s *Service) RevokeContributorSession(ctx context.Context, portalSlug, authorization string) error {
	if err := s.requireContributorFeatures(); err != nil {
		return err
	}
	portal, err := s.repo.GetPortalBySlug(ctx, strings.TrimSpace(portalSlug))
	if err != nil {
		return err
	}
	token, err := parseAuthorizationToken(authorization, "FeedbackSession")
	if err != nil {
		return err
	}
	return s.nextRepo.RevokeContributorSession(ctx, portal.ID, hashOpaqueToken(token))
}
func (s *Service) ResolvePublicParticipant(ctx context.Context, portalSlug string, accountID uuid.UUID, authorization, expectedKind string) (CoreResolvedParticipant, error) {
	portal, err := s.repo.GetPortalBySlug(ctx, strings.TrimSpace(portalSlug))
	if err != nil {
		return CoreResolvedParticipant{}, err
	}
	if expectedKind == ContributorKindAccount || (expectedKind == "" && accountID != uuid.Nil) {
		if accountID == uuid.Nil {
			return CoreResolvedParticipant{}, ErrAuthenticationRequired
		}
		if s.nextRepo == nil {
			return CoreResolvedParticipant{}, ErrFeatureUnavailable
		}
		participant, err := s.nextRepo.GetParticipantByUser(ctx, portal.ID, accountID)
		if err != nil {
			return CoreResolvedParticipant{}, err
		}
		return CoreResolvedParticipant{Participant: participant, AccountID: accountID}, nil
	}
	result, err := s.ResolveContributorSession(ctx, portalSlug, authorization, "")
	if err != nil {
		return CoreResolvedParticipant{}, err
	}
	if expectedKind != "" && result.Participant.Kind != expectedKind {
		return CoreResolvedParticipant{}, ErrAuthenticationRequired
	}
	if result.Participant.Kind != ContributorKindVerifiedGuest && result.Participant.Kind != ContributorKindExternal {
		return CoreResolvedParticipant{}, ErrAuthenticationRequired
	}
	return CoreResolvedParticipant{Participant: result.Participant, SessionID: result.Session.ID}, nil
}
func (s *Service) ExchangeUnsubscribeToken(ctx context.Context, portalSlug, token string) (CoreContributorSessionResult, error) {
	if err := s.requireContributorFeatures(); err != nil {
		return CoreContributorSessionResult{}, err
	}
	portal, err := s.repo.GetPortalBySlug(ctx, strings.TrimSpace(portalSlug))
	if err != nil {
		return CoreContributorSessionResult{}, err
	}
	preferenceToken, preferenceHash, err := s.randomToken(32)
	if err != nil {
		return CoreContributorSessionResult{}, err
	}
	participant, session, err := s.nextRepo.ConsumeUnsubscribeToken(ctx, portal.ID, hashOpaqueToken(strings.TrimSpace(token)), preferenceHash, s.now().UTC().Add(preferenceSessionLifetime))
	if err != nil {
		return CoreContributorSessionResult{}, normalizeSessionError(err)
	}
	return CoreContributorSessionResult{Participant: maskParticipant(participant, portal.GuestIdentityPolicy), Session: session, Token: preferenceToken}, nil
}

// ResolveContributorRateLimitIdentity is intentionally narrow: only normal
// portal/widget contributor sessions can receive authenticated rate limits.
// Preference sessions never authorize participation or bypass ingress limits.
