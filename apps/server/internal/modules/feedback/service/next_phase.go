package feedback

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/complexus-tech/projects-api/pkg/events"
	"github.com/complexus-tech/projects-api/pkg/feedbacksecurity"
	"github.com/complexus-tech/projects-api/pkg/tasks"
	"github.com/complexus-tech/projects-api/pkg/validate"
	"github.com/google/uuid"
)

const (
	verificationLifetime       = 10 * time.Minute
	contributorSessionLifetime = 30 * 24 * time.Hour
	preferenceSessionLifetime  = 30 * time.Minute
	widgetSessionLifetime      = 24 * time.Hour
	widgetAssertionMaxLifetime = 5 * time.Minute
	widgetClockSkew            = 30 * time.Second
	widgetRotationGrace        = 24 * time.Hour
	maxDisplayNameRunes        = 100
	maxUpdateTitleRunes        = 200
	maxUpdateSummaryRunes      = 500
	maxUpdateBodyRunes         = 50_000
	defaultUpdatesPageSize     = 20
	maxUpdatesPageSize         = 50
	maxMergeCandidatesLimit    = 50
)

type ContributorTasks interface {
	EnqueueFeedbackContributorDelivery(tasks.FeedbackContributorDeliveryPayload) error
}

type contributorSecurity struct {
	codeKey        [sha256.Size]byte
	unsubscribeKey [sha256.Size]byte
	widgetAEAD     cipher.AEAD
	widgetAADBase  []byte
}

func WithContributorFeatures(authSecret, websiteURL string, taskService ContributorTasks) Option {
	return func(service *Service) {
		service.websiteURL = strings.TrimRight(strings.TrimSpace(websiteURL), "/")
		service.tasks = taskService
		service.security = newContributorSecurity(authSecret)
	}
}

func newContributorSecurity(secret string) *contributorSecurity {
	secret = strings.TrimSpace(secret)
	if secret == "" {
		return nil
	}
	codeKey := sha256.Sum256([]byte("fortyone:feedback-verification-code:v1\x00" + secret))
	unsubscribeKey, err := feedbacksecurity.DeriveUnsubscribeKey(secret)
	if err != nil {
		return nil
	}
	widgetKey := sha256.Sum256([]byte("fortyone:feedback-widget-signing-secret:v1\x00" + secret))
	block, err := aes.NewCipher(widgetKey[:])
	if err != nil {
		return nil
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil
	}
	return &contributorSecurity{
		codeKey:        codeKey,
		unsubscribeKey: unsubscribeKey,
		widgetAEAD:     aead,
		widgetAADBase:  []byte("fortyone:feedback-widget-signing-secret:v1"),
	}
}

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
	challenge, err := s.nextRepo.CreateContributorVerification(ctx, CoreVerificationRequest{
		PortalID:     portal.ID,
		Email:        normalizedEmail,
		DisplayName:  displayName,
		PublicMasked: publicMasked,
		TokenHash:    tokenHash,
		CodeHash:     s.verificationCodeHash(portal.ID, normalizedEmail, code),
		Source:       source,
		ExpiresAt:    expiresAt,
	})
	if err != nil {
		return CoreVerificationChallenge{}, err
	}
	verificationURL := fmt.Sprintf(
		"%s/portal/%s/feedback/verify?token=%s",
		s.websiteURL,
		url.PathEscape(portal.Slug),
		url.QueryEscape(opaqueToken),
	)
	s.publish(ctx, events.Event{
		Type: events.FeedbackContributorVerification,
		Payload: events.FeedbackContributorVerificationPayload{
			Email:           normalizedEmail,
			DisplayName:     displayName,
			PortalName:      portal.Name,
			VerificationURL: verificationURL,
			Code:            code,
			ExpiresAt:       expiresAt,
		},
		Timestamp: s.now().UTC(),
	})
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
	participant, session, err := s.nextRepo.ConfirmContributorVerification(ctx, CoreVerificationConfirmation{
		PortalID:         portal.ID,
		TokenHash:        tokenHash,
		Email:            normalizedEmail,
		CodeHash:         codeHash,
		Source:           source,
		SessionTokenHash: sessionHash,
		SessionExpiresAt: s.now().UTC().Add(contributorSessionLifetime),
	})
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

// ResolveContributorRateLimitIdentity is intentionally narrow: only normal
// portal/widget contributor sessions can receive authenticated rate limits.
// Preference sessions never authorize participation or bypass ingress limits.
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

func (s *Service) FollowItem(ctx context.Context, portalSlug string, itemID uuid.UUID, participant CoreParticipant, following bool) (CoreFollowState, error) {
	portal, item, err := s.publicPortalItem(ctx, portalSlug, itemID)
	if err != nil {
		return CoreFollowState{}, err
	}
	if participant.PortalID != portal.ID || participant.ID == uuid.Nil || participant.Kind == ContributorKindAnonymous {
		return CoreFollowState{}, ErrAuthenticationRequired
	}
	return s.nextRepo.SetItemFollow(ctx, item.ID, participant.ID, following)
}

func (s *Service) GetItemFollow(ctx context.Context, portalSlug string, itemID uuid.UUID, participant CoreParticipant) (CoreFollowState, error) {
	portal, item, err := s.publicPortalItem(ctx, portalSlug, itemID)
	if err != nil {
		return CoreFollowState{}, err
	}
	if participant.PortalID != portal.ID || participant.ID == uuid.Nil {
		return CoreFollowState{}, ErrAuthenticationRequired
	}
	return s.nextRepo.GetItemFollow(ctx, item.ID, participant.ID)
}

func (s *Service) GetContributorPreferences(ctx context.Context, portalSlug string, participant CoreParticipant) (CoreContributorPreferences, error) {
	portal, err := s.repo.GetPortalBySlug(ctx, strings.TrimSpace(portalSlug))
	if err != nil {
		return CoreContributorPreferences{}, err
	}
	if participant.PortalID != portal.ID {
		return CoreContributorPreferences{}, ErrAuthenticationRequired
	}
	return s.nextRepo.GetContributorPreferences(ctx, portal.ID, participant.ID)
}

func (s *Service) SetPortalEmailPreference(ctx context.Context, portalSlug string, participant CoreParticipant, enabled bool) (CoreContributorPreferences, error) {
	portal, err := s.repo.GetPortalBySlug(ctx, strings.TrimSpace(portalSlug))
	if err != nil {
		return CoreContributorPreferences{}, err
	}
	if participant.PortalID != portal.ID {
		return CoreContributorPreferences{}, ErrAuthenticationRequired
	}
	return s.nextRepo.SetPortalEmailPreference(ctx, portal.ID, participant.ID, enabled)
}

func (s *Service) GetUnreadUpdateCount(ctx context.Context, portalSlug string, participant CoreParticipant) (int, error) {
	portal, err := s.repo.GetPortalBySlug(ctx, strings.TrimSpace(portalSlug))
	if err != nil {
		return 0, err
	}
	if participant.PortalID != portal.ID || participant.ID == uuid.Nil || participant.Kind == ContributorKindAnonymous {
		return 0, ErrAuthenticationRequired
	}
	return s.nextRepo.GetUnreadUpdateCount(ctx, portal.ID, participant.ID)
}

func (s *Service) MarkUpdatesSeen(ctx context.Context, portalSlug string, participant CoreParticipant) (time.Time, error) {
	portal, err := s.repo.GetPortalBySlug(ctx, strings.TrimSpace(portalSlug))
	if err != nil {
		return time.Time{}, err
	}
	if participant.PortalID != portal.ID || participant.ID == uuid.Nil || participant.Kind == ContributorKindAnonymous {
		return time.Time{}, ErrAuthenticationRequired
	}
	return s.nextRepo.MarkUpdatesSeen(ctx, portal.ID, participant.ID)
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
	participant, session, err := s.nextRepo.ConsumeUnsubscribeToken(
		ctx,
		portal.ID,
		hashOpaqueToken(strings.TrimSpace(token)),
		preferenceHash,
		s.now().UTC().Add(preferenceSessionLifetime),
	)
	if err != nil {
		return CoreContributorSessionResult{}, normalizeSessionError(err)
	}
	return CoreContributorSessionResult{Participant: maskParticipant(participant, portal.GuestIdentityPolicy), Session: session, Token: preferenceToken}, nil
}

func (s *Service) ListWorkspaceUpdates(ctx context.Context, workspaceID uuid.UUID, page, pageSize int) (CoreUpdatesPage, error) {
	page, pageSize = normalizeUpdatePagination(page, pageSize)
	return s.nextRepo.ListWorkspaceUpdates(ctx, workspaceID, page, pageSize)
}

func (s *Service) GetWorkspaceUpdate(ctx context.Context, workspaceID, updateID uuid.UUID) (CoreFeedbackUpdate, error) {
	if workspaceID == uuid.Nil || updateID == uuid.Nil {
		return CoreFeedbackUpdate{}, invalidInput("workspace and update ids are required")
	}
	return s.nextRepo.GetWorkspaceUpdate(ctx, workspaceID, updateID)
}

func (s *Service) CreateUpdate(ctx context.Context, input CoreUpdateInput) (CoreFeedbackUpdate, error) {
	if err := s.validateUpdateInput(ctx, &input, false); err != nil {
		return CoreFeedbackUpdate{}, err
	}
	return s.nextRepo.CreateUpdate(ctx, input)
}

func (s *Service) UpdateUpdate(ctx context.Context, input CoreUpdateInput) (CoreFeedbackUpdate, error) {
	if err := s.validateUpdateInput(ctx, &input, true); err != nil {
		return CoreFeedbackUpdate{}, err
	}
	return s.nextRepo.UpdateUpdate(ctx, input)
}

func (s *Service) DeleteUpdate(ctx context.Context, workspaceID, updateID uuid.UUID) error {
	if workspaceID == uuid.Nil || updateID == uuid.Nil {
		return invalidInput("workspace and update ids are required")
	}
	return s.nextRepo.DeleteUpdate(ctx, workspaceID, updateID)
}

func (s *Service) MergeItems(ctx context.Context, input CoreMergeItemInput) (CoreMergeItemResult, error) {
	if s.nextRepo == nil {
		return CoreMergeItemResult{}, ErrFeatureUnavailable
	}
	if input.WorkspaceID == uuid.Nil || input.SourceItemID == uuid.Nil || input.TargetItemID == uuid.Nil || input.ActorID == uuid.Nil {
		return CoreMergeItemResult{}, invalidInput("workspace, source, target, and actor ids are required")
	}
	if input.SourceItemID == input.TargetItemID {
		return CoreMergeItemResult{}, ErrMergeConflict
	}
	return s.nextRepo.MergeItems(ctx, input)
}

func (s *Service) ListMergeCandidates(ctx context.Context, workspaceID, sourceItemID uuid.UUID, search string, limit int) (CoreMergeCandidatesPage, error) {
	if s.nextRepo == nil {
		return CoreMergeCandidatesPage{}, ErrFeatureUnavailable
	}
	if workspaceID == uuid.Nil || sourceItemID == uuid.Nil {
		return CoreMergeCandidatesPage{}, invalidInput("workspace and source feedback ids are required")
	}
	search = strings.TrimSpace(search)
	if utf8.RuneCountInString(search) > maxPublicFeedbackTitleCharacters {
		return CoreMergeCandidatesPage{}, invalidInput("merge candidate search is too long")
	}
	if limit <= 0 {
		limit = 30
	}
	if limit > maxMergeCandidatesLimit {
		limit = maxMergeCandidatesLimit
	}
	source, err := s.repo.GetItem(ctx, workspaceID, sourceItemID)
	if err != nil {
		return CoreMergeCandidatesPage{}, err
	}
	if source.DeletedAt != nil || source.MergedIntoItemID != nil {
		return CoreMergeCandidatesPage{}, ErrMergeConflict
	}
	return s.nextRepo.ListItemCandidates(ctx, workspaceID, source.PortalID, sourceItemID, search, limit)
}

func (s *Service) ListPortalItemCandidates(ctx context.Context, workspaceID, portalID uuid.UUID, search string, limit int) (CoreMergeCandidatesPage, error) {
	if s.nextRepo == nil {
		return CoreMergeCandidatesPage{}, ErrFeatureUnavailable
	}
	if workspaceID == uuid.Nil || portalID == uuid.Nil {
		return CoreMergeCandidatesPage{}, invalidInput("workspace and portal ids are required")
	}
	search = strings.TrimSpace(search)
	if utf8.RuneCountInString(search) > maxPublicFeedbackTitleCharacters {
		return CoreMergeCandidatesPage{}, invalidInput("item candidate search is too long")
	}
	if limit <= 0 {
		limit = 30
	}
	if limit > maxMergeCandidatesLimit {
		limit = maxMergeCandidatesLimit
	}
	if _, err := s.repo.GetPortal(ctx, workspaceID, portalID); err != nil {
		return CoreMergeCandidatesPage{}, err
	}
	return s.nextRepo.ListItemCandidates(ctx, workspaceID, portalID, uuid.Nil, search, limit)
}

func (s *Service) PublishUpdate(ctx context.Context, workspaceID, updateID, actorID uuid.UUID) (CoreFeedbackUpdate, error) {
	if workspaceID == uuid.Nil || updateID == uuid.Nil || actorID == uuid.Nil {
		return CoreFeedbackUpdate{}, invalidInput("workspace, update, and actor ids are required")
	}
	update, _, err := s.nextRepo.PublishUpdate(ctx, workspaceID, updateID, actorID)
	if err != nil {
		return CoreFeedbackUpdate{}, err
	}
	// A draft-to-published transition and its immutable publication outbox row
	// are committed atomically by the repository. Delivery is deliberately not
	// attempted on the request path: the shared feedback outbox worker claims
	// and dispatches it with durable retries.
	return update, nil
}

func (s *Service) UnpublishUpdate(ctx context.Context, workspaceID, updateID uuid.UUID) (CoreFeedbackUpdate, error) {
	if workspaceID == uuid.Nil || updateID == uuid.Nil {
		return CoreFeedbackUpdate{}, invalidInput("workspace and update ids are required")
	}
	return s.nextRepo.UnpublishUpdate(ctx, workspaceID, updateID)
}

func (s *Service) ListPublicUpdates(ctx context.Context, portalSlug string, page, pageSize int) (CoreUpdatesPage, error) {
	portal, err := s.repo.GetPortalBySlug(ctx, strings.TrimSpace(portalSlug))
	if err != nil {
		return CoreUpdatesPage{}, err
	}
	page, pageSize = normalizeUpdatePagination(page, pageSize)
	return s.nextRepo.ListPublicUpdates(ctx, portal.ID, page, pageSize)
}

func (s *Service) GetPublicUpdate(ctx context.Context, portalSlug, updateSlug string) (CoreFeedbackUpdate, error) {
	portal, err := s.repo.GetPortalBySlug(ctx, strings.TrimSpace(portalSlug))
	if err != nil {
		return CoreFeedbackUpdate{}, err
	}
	return s.nextRepo.GetPublicUpdate(ctx, portal.ID, normalizeSlug(updateSlug))
}

func (s *Service) GetWidgetSettings(ctx context.Context, workspaceID, portalID uuid.UUID) (CoreWidgetSettings, error) {
	if workspaceID == uuid.Nil || portalID == uuid.Nil {
		return CoreWidgetSettings{}, invalidInput("workspace and portal ids are required")
	}
	return s.nextRepo.GetWidgetSettings(ctx, workspaceID, portalID)
}

// GetPublicWidgetSettings returns only the non-secret configuration that an
// embed needs before it renders. Portals without settings are safely disabled.
func (s *Service) GetPublicWidgetSettings(ctx context.Context, portalSlug string) (CoreWidgetSettings, error) {
	portal, err := s.repo.GetPortalBySlug(ctx, strings.TrimSpace(portalSlug))
	if err != nil {
		return CoreWidgetSettings{}, err
	}
	settings, err := s.nextRepo.GetPublicWidgetSettings(ctx, portal.ID)
	if errors.Is(err, ErrNotFound) {
		return CoreWidgetSettings{PortalID: portal.ID, AllowedOrigins: []string{}}, nil
	}
	return settings, err
}

func (s *Service) UpdateWidgetSettings(ctx context.Context, input CoreWidgetSettingsInput) (CoreWidgetSettings, error) {
	if input.WorkspaceID == uuid.Nil || input.PortalID == uuid.Nil {
		return CoreWidgetSettings{}, invalidInput("workspace and portal ids are required")
	}
	origins, err := normalizeAllowedOrigins(input.AllowedOrigins)
	if err != nil {
		return CoreWidgetSettings{}, err
	}
	input.AllowedOrigins = origins
	if input.Enabled {
		if len(origins) == 0 {
			return CoreWidgetSettings{}, invalidInput("at least one allowed origin is required before enabling the widget")
		}
	}
	return s.nextRepo.UpsertWidgetSettings(ctx, input)
}

func (s *Service) CreateWidgetSigningSecret(ctx context.Context, workspaceID, portalID uuid.UUID) (CoreWidgetSecretResult, error) {
	if err := s.requireContributorFeatures(); err != nil {
		return CoreWidgetSecretResult{}, err
	}
	secret, _, err := s.randomToken(32)
	if err != nil {
		return CoreWidgetSecretResult{}, err
	}
	keyID := uuid.New()
	encrypted, err := s.sealWidgetSecret(portalID, 1, secret)
	if err != nil {
		return CoreWidgetSecretResult{}, err
	}
	settings, err := s.nextRepo.SetInitialWidgetSecret(ctx, workspaceID, portalID, keyID, encrypted, 1)
	if err != nil {
		return CoreWidgetSecretResult{}, err
	}
	return CoreWidgetSecretResult{Settings: settings, SigningSecret: secret}, nil
}

func (s *Service) RotateWidgetSigningSecret(ctx context.Context, workspaceID, portalID uuid.UUID) (CoreWidgetSecretResult, error) {
	if err := s.requireContributorFeatures(); err != nil {
		return CoreWidgetSecretResult{}, err
	}
	current, err := s.nextRepo.GetWidgetSettings(ctx, workspaceID, portalID)
	if err != nil {
		return CoreWidgetSecretResult{}, err
	}
	if strings.TrimSpace(current.SigningSecretEncrypted) == "" {
		return CoreWidgetSecretResult{}, invalidInput("create a widget signing secret before rotating it")
	}
	secret, _, err := s.randomToken(32)
	if err != nil {
		return CoreWidgetSecretResult{}, err
	}
	version := current.SigningSecretVersion + 1
	encrypted, err := s.sealWidgetSecret(portalID, version, secret)
	if err != nil {
		return CoreWidgetSecretResult{}, err
	}
	settings, err := s.nextRepo.RotateWidgetSecret(ctx, workspaceID, portalID, encrypted, version, s.now().UTC().Add(widgetRotationGrace))
	if err != nil {
		return CoreWidgetSecretResult{}, err
	}
	return CoreWidgetSecretResult{Settings: settings, SigningSecret: secret}, nil
}

func (s *Service) CreateWidgetContributorSession(ctx context.Context, input CoreWidgetSessionInput) (CoreContributorSessionResult, error) {
	if err := s.requireContributorFeatures(); err != nil {
		return CoreContributorSessionResult{}, err
	}
	portal, err := s.repo.GetPortalBySlug(ctx, strings.TrimSpace(input.PortalSlug))
	if err != nil {
		return CoreContributorSessionResult{}, err
	}
	settings, err := s.nextRepo.GetPublicWidgetSettings(ctx, portal.ID)
	if err != nil || !settings.Enabled {
		return CoreContributorSessionResult{}, ErrWidgetAssertionInvalid
	}
	parentOrigin, err := normalizeOrigin(input.ParentOrigin)
	if err != nil || !slices.Contains(settings.AllowedOrigins, parentOrigin) {
		return CoreContributorSessionResult{}, ErrWidgetOriginNotAllowed
	}
	assertion, signedPart, signature, err := parseWidgetAssertion(input.Assertion)
	if err != nil {
		return CoreContributorSessionResult{}, ErrWidgetAssertionInvalid
	}
	now := s.now().UTC()
	issuedAt := time.Unix(assertion.IssuedAt, 0).UTC()
	expiresAt := time.Unix(assertion.ExpiresAt, 0).UTC()
	assertionKeyID, keyErr := uuid.Parse(assertion.KeyID)
	if keyErr != nil || assertion.Version < 1 || assertionKeyID != settings.WidgetKeyID || assertion.Origin != parentOrigin || assertion.ExternalID == "" || assertion.Nonce == "" {
		return CoreContributorSessionResult{}, ErrWidgetAssertionInvalid
	}
	if issuedAt.After(now.Add(widgetClockSkew)) || !expiresAt.After(now) || !expiresAt.After(issuedAt) || expiresAt.Sub(issuedAt) > widgetAssertionMaxLifetime {
		return CoreContributorSessionResult{}, ErrWidgetAssertionInvalid
	}
	normalizedEmail, err := validate.Email(assertion.Email)
	if err != nil {
		return CoreContributorSessionResult{}, ErrWidgetAssertionInvalid
	}
	assertion.ExternalID = strings.TrimSpace(assertion.ExternalID)
	assertion.DisplayName = strings.TrimSpace(assertion.DisplayName)
	if assertion.ExternalID == "" || assertion.DisplayName == "" || utf8.RuneCountInString(assertion.DisplayName) > maxDisplayNameRunes || len(assertion.ExternalID) > 255 || len(assertion.Nonce) > 255 {
		return CoreContributorSessionResult{}, ErrWidgetAssertionInvalid
	}
	var avatarURL *string
	if strings.TrimSpace(assertion.AvatarURL) != "" {
		parsed, parseErr := url.ParseRequestURI(strings.TrimSpace(assertion.AvatarURL))
		if parseErr != nil || parsed.Scheme != "https" || parsed.Host == "" {
			return CoreContributorSessionResult{}, ErrWidgetAssertionInvalid
		}
		value := parsed.String()
		avatarURL = &value
	}
	encryptedSecret, err := s.nextRepo.GetWidgetSigningSecret(ctx, portal.ID, assertionKeyID, assertion.Version)
	if err != nil {
		return CoreContributorSessionResult{}, ErrWidgetAssertionInvalid
	}
	secret, err := s.openWidgetSecret(portal.ID, assertion.Version, encryptedSecret)
	if err != nil {
		return CoreContributorSessionResult{}, ErrWidgetAssertionInvalid
	}
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(signedPart))
	if !hmac.Equal(signature, mac.Sum(nil)) {
		return CoreContributorSessionResult{}, ErrWidgetAssertionInvalid
	}
	token, tokenHash, err := s.randomToken(32)
	if err != nil {
		return CoreContributorSessionResult{}, err
	}
	if err := s.nextRepo.ConsumeWidgetAssertionNonce(ctx, portal.ID, assertionKeyID, assertion.Version, assertion.Nonce, parentOrigin, expiresAt); err != nil {
		return CoreContributorSessionResult{}, err
	}
	participant, session, err := s.nextRepo.CreateExternalContributorSession(
		ctx,
		portal.ID,
		assertion.ExternalID,
		normalizedEmail,
		assertion.DisplayName,
		avatarURL,
		tokenHash,
		now.Add(widgetSessionLifetime),
	)
	if err != nil {
		return CoreContributorSessionResult{}, err
	}
	return CoreContributorSessionResult{Participant: maskParticipant(participant, portal.GuestIdentityPolicy), Session: session, Token: token}, nil
}

func (s *Service) enqueueDeliveries(ctx context.Context, portalID uuid.UUID, itemID, updateID *uuid.UUID, actorContributorID uuid.UUID, eventType, eventKey, subject, message, destinationURL string) {
	if s.nextRepo == nil || s.tasks == nil {
		return
	}
	recipients, err := s.nextRepo.ListDeliveryRecipients(ctx, portalID, itemID, updateID, actorContributorID)
	if err != nil {
		s.logNextPhaseError(ctx, "list feedback delivery recipients", err)
		return
	}
	for _, recipient := range recipients {
		if recipient.Kind != ContributorKindVerifiedGuest && recipient.Kind != ContributorKindExternal {
			continue
		}
		deliveryID := uuid.New()
		unsubscribeToken, unsubscribeHash, err := feedbacksecurity.DeriveUnsubscribeTokenWithKey(s.security.unsubscribeKey[:], deliveryID)
		if err != nil {
			s.logNextPhaseError(ctx, "generate feedback unsubscribe token", err)
			continue
		}
		dedupeKey := eventKey + ":" + recipient.ContributorID.String()
		delivery, created, err := s.nextRepo.CreateContributorDelivery(ctx, CoreCreateDeliveryInput{
			DeliveryID:     deliveryID,
			PortalID:       portalID,
			ContributorID:  recipient.ContributorID,
			ItemID:         itemID,
			UpdateID:       updateID,
			EventType:      eventType,
			DedupeKey:      dedupeKey,
			Subject:        subject,
			Message:        message,
			DestinationURL: destinationURL,
			TokenHash:      unsubscribeHash,
		})
		if err != nil {
			s.logNextPhaseError(ctx, "create feedback contributor delivery", err)
			continue
		}
		if !created {
			continue
		}
		if err := s.tasks.EnqueueFeedbackContributorDelivery(tasks.FeedbackContributorDeliveryPayload{
			DeliveryID:       delivery.ID,
			UnsubscribeToken: unsubscribeToken,
		}); err != nil {
			s.logNextPhaseError(ctx, "enqueue feedback contributor delivery", err)
		}
	}
}

func (s *Service) enqueueStatusDeliveries(ctx context.Context, item CoreItem, actorID uuid.UUID, transitionKey string) {
	if s.nextRepo == nil {
		return
	}
	actorContributorID := uuid.Nil
	if actorID != uuid.Nil {
		if participant, err := s.nextRepo.GetParticipantByUser(ctx, item.PortalID, actorID); err == nil {
			actorContributorID = participant.ID
		}
	}
	portal, err := s.repo.GetPortal(ctx, item.WorkspaceID, item.PortalID)
	if err != nil {
		s.logNextPhaseError(ctx, "load feedback portal for status delivery", err)
		return
	}
	if transitionKey == "" {
		transitionKey = item.UpdatedAt.UTC().Format(time.RFC3339Nano)
	}
	s.enqueueDeliveries(
		ctx,
		item.PortalID,
		&item.ID,
		nil,
		actorContributorID,
		"feedback.status.updated",
		"status:"+item.ID.String()+":"+item.Status+":"+transitionKey,
		item.Title+" is now "+strings.ReplaceAll(item.Status, "_", " "),
		item.RoadmapSummaryText(),
		s.publicItemURL(portal.Slug, item.Slug),
	)
}

// NotifyLinkedStoryStatusTransition bridges projected story workflow changes
// into the same account and contributor delivery paths as direct status edits.
// The event id and delivery key are stable across stream redelivery.
func (s *Service) NotifyLinkedStoryStatusTransition(ctx context.Context, workspaceID, storyID, actorID uuid.UUID, occurredAt time.Time) error {
	if s.nextRepo == nil {
		return ErrFeatureUnavailable
	}
	if workspaceID == uuid.Nil || storyID == uuid.Nil || actorID == uuid.Nil || occurredAt.IsZero() {
		return invalidInput("workspace, story, actor, and transition time are required")
	}
	items, err := s.nextRepo.ListPrimaryStoryItems(ctx, workspaceID, storyID)
	if err != nil {
		return err
	}
	transitionKey := storyID.String() + ":" + occurredAt.UTC().Format(time.RFC3339Nano)
	for _, item := range items {
		if !shouldNotifyFeedbackStatusTransition(item.Status, item.RoadmapSummary) {
			continue
		}
		eventID := uuid.NewSHA1(uuid.NameSpaceOID, []byte(transitionKey+":"+item.ID.String()+":"+item.Status))
		s.publishAccountStatusNotifications(ctx, item, actorID, eventID, occurredAt.UTC())
		s.enqueueStatusDeliveries(ctx, item, actorID, eventID.String())
	}
	return nil
}

// shouldNotifyFeedbackStatusTransition is the single public-notification
// policy for feedback workflow changes. Internal triage is silent. Closing an
// item is public only when the transition supplies an explanation; callers
// without transition input (for example linked-story projection) pass the
// item's current public roadmap summary.
func shouldNotifyFeedbackStatusTransition(status string, publicExplanation *string) bool {
	switch status {
	case StatusPlanned, StatusInProgress, StatusCompleted:
		return true
	case StatusClosed:
		return publicExplanation != nil && strings.TrimSpace(*publicExplanation) != ""
	default:
		return false
	}
}

func (s *Service) requireContributorFeatures() error {
	if s.nextRepo == nil || s.security == nil || s.random == nil {
		return ErrFeatureUnavailable
	}
	return nil
}

func (s *Service) randomToken(size int) (string, []byte, error) {
	buffer := make([]byte, size)
	if _, err := io.ReadFull(s.random, buffer); err != nil {
		return "", nil, fmt.Errorf("generate feedback token: %w", err)
	}
	token := base64.RawURLEncoding.EncodeToString(buffer)
	return token, hashOpaqueToken(token), nil
}

func (s *Service) randomCode() (string, error) {
	value, err := rand.Int(s.random, big.NewInt(1_000_000))
	if err != nil {
		return "", fmt.Errorf("generate feedback verification code: %w", err)
	}
	return fmt.Sprintf("%06d", value.Int64()), nil
}

func (s *Service) verificationCodeHash(portalID uuid.UUID, email, code string) []byte {
	mac := hmac.New(sha256.New, s.security.codeKey[:])
	_, _ = mac.Write([]byte("fortyone:feedback-verification-code:v1\n"))
	_, _ = mac.Write([]byte(portalID.String()))
	_, _ = mac.Write([]byte("\n" + strings.ToLower(strings.TrimSpace(email)) + "\n" + strings.TrimSpace(code)))
	return mac.Sum(nil)
}

func (s *Service) sealWidgetSecret(portalID uuid.UUID, version int, plaintext string) (string, error) {
	nonce := make([]byte, s.security.widgetAEAD.NonceSize())
	if _, err := io.ReadFull(s.random, nonce); err != nil {
		return "", fmt.Errorf("generate feedback widget encryption nonce: %w", err)
	}
	aad := s.widgetSecretAAD(portalID, version)
	ciphertext := s.security.widgetAEAD.Seal(nil, nonce, []byte(plaintext), aad)
	return "v1." + base64.RawURLEncoding.EncodeToString(append(nonce, ciphertext...)), nil
}

func (s *Service) openWidgetSecret(portalID uuid.UUID, version int, envelope string) (string, error) {
	if !strings.HasPrefix(envelope, "v1.") {
		return "", errors.New("unsupported feedback widget signing envelope")
	}
	raw, err := base64.RawURLEncoding.DecodeString(strings.TrimPrefix(envelope, "v1."))
	if err != nil || len(raw) <= s.security.widgetAEAD.NonceSize() {
		return "", errors.New("invalid feedback widget signing envelope")
	}
	nonceSize := s.security.widgetAEAD.NonceSize()
	plaintext, err := s.security.widgetAEAD.Open(nil, raw[:nonceSize], raw[nonceSize:], s.widgetSecretAAD(portalID, version))
	if err != nil {
		return "", fmt.Errorf("decrypt feedback widget signing secret: %w", err)
	}
	return string(plaintext), nil
}

func (s *Service) widgetSecretAAD(portalID uuid.UUID, version int) []byte {
	return []byte(string(s.security.widgetAADBase) + "\n" + portalID.String() + "\n" + strconv.Itoa(version))
}

func (s *Service) validateUpdateInput(ctx context.Context, input *CoreUpdateInput, requireID bool) error {
	if s.nextRepo == nil {
		return ErrFeatureUnavailable
	}
	if input.WorkspaceID == uuid.Nil || input.PortalID == uuid.Nil || input.ActorID == uuid.Nil || (requireID && input.UpdateID == uuid.Nil) {
		return invalidInput("workspace, portal, update, and actor ids are required")
	}
	portal, err := s.repo.GetPortal(ctx, input.WorkspaceID, input.PortalID)
	if err != nil {
		return err
	}
	if portal.WorkspaceID != input.WorkspaceID {
		return ErrNotFound
	}
	input.Title = strings.TrimSpace(input.Title)
	input.Body = strings.TrimSpace(input.Body)
	if input.Title == "" || input.Body == "" {
		return invalidInput("update title and body are required")
	}
	if utf8.RuneCountInString(input.Title) > maxUpdateTitleRunes || utf8.RuneCountInString(input.Body) > maxUpdateBodyRunes {
		return invalidInput("feedback update content is too long")
	}
	if input.Summary != nil {
		value := strings.TrimSpace(*input.Summary)
		if value == "" {
			input.Summary = nil
		} else {
			if utf8.RuneCountInString(value) > maxUpdateSummaryRunes {
				return invalidInputf("update summary must be %d characters or fewer", maxUpdateSummaryRunes)
			}
			input.Summary = &value
		}
	}
	if input.CoverImageURL != nil {
		value := strings.TrimSpace(*input.CoverImageURL)
		if value == "" {
			input.CoverImageURL = nil
		} else {
			if value != "" {
				parsed, parseErr := url.ParseRequestURI(value)
				if parseErr != nil || parsed.Scheme != "https" || parsed.Host == "" {
					return invalidInput("cover image URL must be an absolute HTTPS URL")
				}
			}
			input.CoverImageURL = &value
		}
	}
	input.Slug = normalizeSlug(input.Slug)
	if input.Slug == "" {
		input.Slug = normalizeSlug(input.Title) + "-" + uuid.NewString()[:8]
	}
	input.ItemIDs = uniqueNonNilUUIDs(input.ItemIDs)
	return nil
}

func (s *Service) publicPortalItem(ctx context.Context, portalSlug string, itemID uuid.UUID) (CorePortal, CoreItem, error) {
	if itemID == uuid.Nil {
		return CorePortal{}, CoreItem{}, invalidInput("feedback id is required")
	}
	portal, err := s.repo.GetPortalBySlug(ctx, strings.TrimSpace(portalSlug))
	if err != nil {
		return CorePortal{}, CoreItem{}, err
	}
	item, err := s.repo.GetItemByPortal(ctx, portal.ID, itemID)
	if err != nil {
		return CorePortal{}, CoreItem{}, err
	}
	if item.WorkspaceID != portal.WorkspaceID {
		return CorePortal{}, CoreItem{}, ErrNotFound
	}
	if item.MergedIntoItemID != nil {
		return CorePortal{}, CoreItem{}, ErrMergeConflict
	}
	return portal, item, nil
}

func (s *Service) publicItemURL(portalSlug, itemSlug string) string {
	return fmt.Sprintf("%s/portal/%s/feedback/%s", s.websiteURL, url.PathEscape(portalSlug), url.PathEscape(itemSlug))
}

func (s *Service) logNextPhaseError(ctx context.Context, operation string, err error) {
	if s.log != nil {
		s.log.Error(ctx, operation, "error", err)
	}
}

func (update CoreFeedbackUpdate) SummaryText() string {
	if update.Summary != nil && strings.TrimSpace(*update.Summary) != "" {
		return strings.TrimSpace(*update.Summary)
	}
	return update.Title
}

func (item CoreItem) RoadmapSummaryText() string {
	if item.RoadmapSummary != nil && strings.TrimSpace(*item.RoadmapSummary) != "" {
		return strings.TrimSpace(*item.RoadmapSummary)
	}
	return "The status of this feedback changed to " + strings.ReplaceAll(item.Status, "_", " ") + "."
}

func hashOpaqueToken(token string) []byte {
	hash := sha256.Sum256([]byte(strings.TrimSpace(token)))
	return hash[:]
}

func parseAuthorizationToken(value, scheme string) (string, error) {
	parts := strings.Fields(strings.TrimSpace(value))
	if len(parts) != 2 || !strings.EqualFold(parts[0], scheme) || parts[1] == "" {
		return "", ErrContributorSessionInvalid
	}
	return parts[1], nil
}

func normalizeSessionError(err error) error {
	if errors.Is(err, ErrVerificationExpired) || errors.Is(err, ErrVerificationConsumed) || errors.Is(err, ErrVerificationAttempts) || errors.Is(err, ErrContributorBlocked) {
		return err
	}
	if errors.Is(err, ErrNotFound) {
		return ErrContributorSessionInvalid
	}
	return err
}

func normalizeContributorSessionSource(source string) string {
	return strings.ToLower(strings.TrimSpace(source))
}

func maskParticipant(participant CoreParticipant, policy string) CoreParticipant {
	if participant.Kind != ContributorKindVerifiedGuest && participant.Kind != ContributorKindExternal {
		return participant
	}
	if policy == GuestIdentityPolicyAlwaysMaskGuests {
		participant.PublicMasked = true
	}
	return participant
}

func normalizeAllowedOrigins(values []string) ([]string, error) {
	result := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		origin, err := normalizeOrigin(value)
		if err != nil {
			return nil, err
		}
		if _, exists := seen[origin]; exists {
			continue
		}
		seen[origin] = struct{}{}
		result = append(result, origin)
	}
	slices.Sort(result)
	return result, nil
}

func normalizeOrigin(value string) (string, error) {
	parsed, err := url.Parse(strings.TrimSpace(value))
	if err != nil || parsed.Host == "" || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" || (parsed.Path != "" && parsed.Path != "/") {
		return "", invalidInput("widget origins must be exact origins without a path, query, or fragment")
	}
	hostname := strings.ToLower(parsed.Hostname())
	local := hostname == "localhost" || hostname == "127.0.0.1" || hostname == "::1"
	if parsed.Scheme != "https" && !(parsed.Scheme == "http" && local) {
		return "", invalidInput("widget origins must use HTTPS except for localhost development")
	}
	if strings.Contains(parsed.Host, "*") {
		return "", invalidInput("widget origin wildcards are not supported")
	}
	parsed.Scheme = strings.ToLower(parsed.Scheme)
	parsed.Host = strings.ToLower(parsed.Host)
	parsed.Path = ""
	return parsed.String(), nil
}

func parseWidgetAssertion(value string) (CoreWidgetIdentityAssertion, string, []byte, error) {
	parts := strings.Split(strings.TrimSpace(value), ".")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return CoreWidgetIdentityAssertion{}, "", nil, ErrWidgetAssertionInvalid
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil || len(payload) > 8<<10 {
		return CoreWidgetIdentityAssertion{}, "", nil, ErrWidgetAssertionInvalid
	}
	signature, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil || len(signature) != sha256.Size {
		return CoreWidgetIdentityAssertion{}, "", nil, ErrWidgetAssertionInvalid
	}
	var assertion CoreWidgetIdentityAssertion
	decoder := json.NewDecoder(strings.NewReader(string(payload)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&assertion); err != nil {
		return CoreWidgetIdentityAssertion{}, "", nil, ErrWidgetAssertionInvalid
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		return CoreWidgetIdentityAssertion{}, "", nil, ErrWidgetAssertionInvalid
	}
	return assertion, parts[0], signature, nil
}

func normalizeUpdatePagination(page, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = defaultUpdatesPageSize
	}
	if pageSize > maxUpdatesPageSize {
		pageSize = maxUpdatesPageSize
	}
	return page, pageSize
}

func uniqueNonNilUUIDs(values []uuid.UUID) []uuid.UUID {
	result := make([]uuid.UUID, 0, len(values))
	seen := make(map[uuid.UUID]struct{}, len(values))
	for _, value := range values {
		if value == uuid.Nil {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}

func encodeWidgetAssertionForTest(assertion CoreWidgetIdentityAssertion, secret string) (string, error) {
	payload, err := json.Marshal(assertion)
	if err != nil {
		return "", err
	}
	encoded := base64.RawURLEncoding.EncodeToString(payload)
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(encoded))
	return encoded + "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil)), nil
}

func tokenHashHex(token string) string {
	return hex.EncodeToString(hashOpaqueToken(token))
}
