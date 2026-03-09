package usershttp

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	users "github.com/complexus-tech/projects-api/internal/modules/users/service"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/cache"
	"github.com/complexus-tech/projects-api/pkg/events"
	"github.com/complexus-tech/projects-api/pkg/google"
	"github.com/complexus-tech/projects-api/pkg/publisher"
	"github.com/complexus-tech/projects-api/pkg/validate"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

var (
	ErrInvalidWorkspaceID = errors.New("workspace id is not in its proper form")
	SessionDuration       = time.Hour * 24 * 30
	avatarAccessURLExpiry = 24 * time.Hour
)

const sessionCookieName = "fortyone_session"

const googleAuthStateTTL = 10 * time.Minute

type Handlers struct {
	users         *users.Service
	attachments   users.AttachmentsService
	secretKey     string
	cookieDomain  string
	cache         *cache.Service
	googleService *google.Service
	publisher     *publisher.Publisher
}

func New(users *users.Service, attachments users.AttachmentsService, secretKey string, cookieDomain string, cacheService *cache.Service, googleService *google.Service, publisher *publisher.Publisher) *Handlers {
	return &Handlers{
		users:         users,
		attachments:   attachments,
		secretKey:     secretKey,
		cookieDomain:  cookieDomain,
		cache:         cacheService,
		googleService: googleService,
		publisher:     publisher,
	}
}

func (h *Handlers) resolveUserAvatarURL(ctx context.Context, avatar string) string {
	if h.attachments == nil {
		return avatar
	}
	resolved, err := h.attachments.ResolveProfileImageURL(ctx, avatar, avatarAccessURLExpiry)
	if err != nil {
		return ""
	}
	return resolved
}

func (h *Handlers) resolveUserAvatar(ctx context.Context, user *users.CoreUser) {
	if user == nil {
		return
	}
	user.AvatarURL = h.resolveUserAvatarURL(ctx, user.AvatarURL)
}

func (h *Handlers) resolveUserAvatars(ctx context.Context, usersList []users.CoreUser) {
	for i := range usersList {
		usersList[i].AvatarURL = h.resolveUserAvatarURL(ctx, usersList[i].AvatarURL)
	}
}

func (h *Handlers) createSessionToken() (string, error) {
	randomBytes := make([]byte, 32)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", fmt.Errorf("generate session token: %w", err)
	}

	return base64.RawURLEncoding.EncodeToString(randomBytes), nil
}

func (h *Handlers) persistSession(
	ctx context.Context,
	userID uuid.UUID,
	token string,
	expires time.Time,
) error {
	if h.cache == nil {
		return errors.New("auth session cache is not configured")
	}

	ttl := time.Until(expires)
	if ttl <= 0 {
		return errors.New("auth session expiry must be in the future")
	}

	return h.cache.Set(ctx, cache.AuthSessionCacheKey(token), userID.String(), ttl)
}

func (h *Handlers) setSessionCookie(w http.ResponseWriter, r *http.Request, value string, expires time.Time) {
	secure := isSecureRequest(r)
	cookie := http.Cookie{
		Name:     sessionCookieName,
		Value:    value,
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
		Expires:  expires,
		MaxAge:   int(time.Until(expires).Seconds()),
	}

	if domain := cookieDomainForRequest(r, h.cookieDomain); domain != "" {
		cookie.Domain = domain
	}

	http.SetCookie(w, &cookie)
}

func (h *Handlers) clearSessionCookie(w http.ResponseWriter, r *http.Request) {
	expires := time.Unix(0, 0)
	cookie := http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   isSecureRequest(r),
		SameSite: http.SameSiteLaxMode,
		Expires:  expires,
		MaxAge:   -1,
	}

	if domain := cookieDomainForRequest(r, h.cookieDomain); domain != "" {
		cookie.Domain = domain
	}

	http.SetCookie(w, &cookie)
}

func cookieDomainForRequest(r *http.Request, configuredDomain string) string {
	if configuredDomain != "" {
		return configuredDomain
	}

	host := r.Host
	if host == "" {
		return ""
	}
	if strings.Contains(host, ":") {
		host = strings.Split(host, ":")[0]
	}
	if strings.HasSuffix(host, ".fortyone.app") {
		return ".fortyone.app"
	}
	return ""
}

func isSecureRequest(r *http.Request) bool {
	if r.TLS != nil {
		return true
	}
	return strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https")
}

func sanitizeCallbackURL(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", nil
	}

	if strings.HasPrefix(raw, "/") && !strings.HasPrefix(raw, "//") {
		return raw, nil
	}

	parsed, err := url.Parse(raw)
	if err != nil {
		return "", errors.New("invalid callbackURL")
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", errors.New("invalid callbackURL")
	}
	if parsed.Host == "" {
		return "", errors.New("invalid callbackURL")
	}

	return parsed.String(), nil
}

func buildGoogleFullName(identity google.Identity) string {
	if identity.FullName != "" {
		return identity.FullName
	}

	fullName := strings.TrimSpace(strings.TrimSpace(identity.FirstName) + " " + strings.TrimSpace(identity.LastName))
	if fullName != "" {
		return fullName
	}

	localPart := strings.TrimSpace(strings.Split(identity.Email, "@")[0])
	if localPart != "" {
		return localPart
	}

	return "User"
}

func (h *Handlers) GetProfile(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	user, err := h.users.GetUser(ctx, userID)
	if err != nil {
		if errors.Is(err, users.ErrNotFound) {
			return web.RespondError(ctx, w, err, http.StatusNotFound)
		}
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	h.resolveUserAvatar(ctx, &user)
	return web.Respond(ctx, w, toAppUser(user), http.StatusOK)
}

func (h *Handlers) Me(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	return h.GetProfile(ctx, w, r)
}

func (h *Handlers) UpdateProfile(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	var req UpdateProfileRequest
	if err := web.Decode(r, &req); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	updates := users.CoreUpdateUser{}

	if req.Username != "" {
		updates.Username = &req.Username
	}
	if req.FullName != nil {
		updates.FullName = req.FullName
	}
	if req.AvatarURL != nil {
		updates.AvatarURL = req.AvatarURL
	}
	if req.HasSeenWalkthrough != nil {
		updates.HasSeenWalkthrough = req.HasSeenWalkthrough
	}
	if req.Timezone != nil {
		updates.Timezone = req.Timezone
	}

	if err := h.users.UpdateUser(ctx, userID, updates); err != nil {
		if errors.Is(err, users.ErrNotFound) {
			return web.RespondError(ctx, w, err, http.StatusNotFound)
		}
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	user, err := h.users.GetUser(ctx, userID)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	h.resolveUserAvatar(ctx, &user)
	return web.Respond(ctx, w, toAppUser(user), http.StatusOK)
}

func (h *Handlers) DeleteProfile(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	if err := h.users.DeleteUser(ctx, userID); err != nil {
		if errors.Is(err, users.ErrNotFound) {
			return web.RespondError(ctx, w, err, http.StatusNotFound)
		}
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	return web.Respond(ctx, w, nil, http.StatusNoContent)
}

func (h *Handlers) SwitchWorkspace(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	var req SwitchWorkspaceRequest
	if err := web.Decode(r, &req); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	if err := h.users.UpdateUserWorkspace(ctx, userID, req.WorkspaceID); err != nil {
		if errors.Is(err, users.ErrNotFound) {
			return web.RespondError(ctx, w, err, http.StatusNotFound)
		}
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	user, err := h.users.GetUser(ctx, userID)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	h.resolveUserAvatar(ctx, &user)
	return web.Respond(ctx, w, toAppUser(user), http.StatusOK)
}

func (h *Handlers) List(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	var teamID *uuid.UUID
	teamIDParam := r.URL.Query().Get("teamId")
	if teamIDParam != "" {
		parsedTeamID, err := uuid.Parse(teamIDParam)
		if err != nil {
			return web.RespondError(ctx, w, errors.New("invalid team ID format"), http.StatusBadRequest)
		}
		teamID = &parsedTeamID
	}

	users, err := h.users.List(ctx, workspace.ID, teamID)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	h.resolveUserAvatars(ctx, users)
	web.Respond(ctx, w, toAppUsers(users), http.StatusOK)
	return nil
}

func (h *Handlers) GoogleAuth(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	var req GoogleAuthRequest
	if err := web.Decode(r, &req); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	identity, err := h.googleService.VerifyToken(ctx, req.Token)
	if err != nil {
		switch {
		case errors.Is(err, google.ErrInvalidToken):
			return web.RespondError(ctx, w, err, http.StatusUnauthorized)
		case errors.Is(err, google.ErrNotConfigured):
			return web.RespondError(ctx, w, err, http.StatusServiceUnavailable)
		default:
			return web.RespondError(ctx, w, err, http.StatusInternalServerError)
		}
	}

	if !identity.EmailVerified {
		return web.RespondError(ctx, w, errors.New("email not verified by google"), http.StatusUnauthorized)
	}

	user, err := h.authenticateWithGoogleIdentity(ctx, w, r, identity)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	return web.Respond(ctx, w, toAppUser(user), http.StatusOK)
}

func (h *Handlers) StartGoogleAuth(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	if h.cache == nil {
		return web.RespondError(ctx, w, errors.New("auth session cache is not configured"), http.StatusServiceUnavailable)
	}

	callbackURL, err := sanitizeCallbackURL(r.URL.Query().Get("callbackURL"))
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	state, err := h.createSessionToken()
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	if err := h.cache.Set(ctx, cache.AuthGoogleStateCacheKey(state), googleAuthState{CallbackURL: callbackURL}, googleAuthStateTTL); err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	authURL, err := h.googleService.AuthCodeURL(state)
	if err != nil {
		switch {
		case errors.Is(err, google.ErrNotConfigured):
			return web.RespondError(ctx, w, err, http.StatusServiceUnavailable)
		default:
			return web.RespondError(ctx, w, err, http.StatusInternalServerError)
		}
	}

	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
	return nil
}

func (h *Handlers) CompleteGoogleAuth(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	if h.cache == nil {
		return web.RespondError(ctx, w, errors.New("auth session cache is not configured"), http.StatusServiceUnavailable)
	}

	state := strings.TrimSpace(r.URL.Query().Get("state"))
	code := strings.TrimSpace(r.URL.Query().Get("code"))
	if state == "" || code == "" {
		return web.RespondError(ctx, w, errors.New("missing oauth callback parameters"), http.StatusBadRequest)
	}

	var authState googleAuthState
	if err := h.cache.Get(ctx, cache.AuthGoogleStateCacheKey(state), &authState); err != nil {
		return web.RespondError(ctx, w, errors.New("invalid oauth state"), http.StatusUnauthorized)
	}
	_ = h.cache.Delete(ctx, cache.AuthGoogleStateCacheKey(state))

	identity, err := h.googleService.ExchangeCode(ctx, code)
	if err != nil {
		switch {
		case errors.Is(err, google.ErrInvalidToken):
			return web.RespondError(ctx, w, err, http.StatusUnauthorized)
		case errors.Is(err, google.ErrNotConfigured):
			return web.RespondError(ctx, w, err, http.StatusServiceUnavailable)
		default:
			return web.RespondError(ctx, w, err, http.StatusInternalServerError)
		}
	}

	if !identity.EmailVerified {
		return web.RespondError(ctx, w, errors.New("email not verified by google"), http.StatusUnauthorized)
	}

	user, err := h.authenticateWithGoogleIdentity(ctx, w, r, identity)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	if authState.CallbackURL != "" {
		http.Redirect(w, r, authState.CallbackURL, http.StatusTemporaryRedirect)
		return nil
	}

	return web.Respond(ctx, w, toAppUser(user), http.StatusOK)
}

func (h *Handlers) authenticateWithGoogleIdentity(
	ctx context.Context,
	w http.ResponseWriter,
	r *http.Request,
	identity google.Identity,
) (users.CoreUser, error) {
	if strings.TrimSpace(identity.Email) == "" {
		return users.CoreUser{}, errors.New("google account email is missing")
	}

	user, err := h.users.GetUserByEmail(ctx, identity.Email)
	if err != nil && !errors.Is(err, users.ErrNotFound) {
		return users.CoreUser{}, err
	}

	if errors.Is(err, users.ErrNotFound) {
		newUser := users.CoreNewUser{
			Email:    identity.Email,
			FullName: buildGoogleFullName(identity),
			Timezone: "Antarctica/Troll", // Default timezone for new users
		}
		user, err = h.users.Register(ctx, newUser)
		if err != nil {
			return users.CoreUser{}, err
		}
	}

	if identity.Picture != "" && user.AvatarURL == "" {
		if blobName, uploadErr := h.attachments.UploadProfileImageFromURL(ctx, identity.Picture, user.ID); uploadErr == nil {
			updates := users.CoreUpdateUser{AvatarURL: &blobName}
			if updateErr := h.users.UpdateUser(ctx, user.ID, updates); updateErr == nil {
				user.AvatarURL = blobName
			}
		}
	}

	tokenString, err := h.createSessionToken()
	if err != nil {
		return users.CoreUser{}, err
	}

	expiresAt := time.Now().Add(SessionDuration)
	if err := h.persistSession(ctx, user.ID, tokenString, expiresAt); err != nil {
		return users.CoreUser{}, err
	}

	h.setSessionCookie(w, r, tokenString, expiresAt)
	h.resolveUserAvatar(ctx, &user)
	return user, nil
}

func (h *Handlers) SendEmailVerification(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	var req EmailVerificationRequest
	if err := web.Decode(r, &req); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	normalizedEmail, err := validate.Email(req.Email)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	req.Email = normalizedEmail

	count, err := h.users.GetValidTokenCount(ctx, req.Email, 10*time.Minute)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	if count >= 3 {
		return web.RespondError(ctx, w, users.ErrTooManyAttempts, http.StatusTooManyRequests)
	}

	_, err = h.users.GetUserByEmail(ctx, req.Email)
	tokenType := users.TokenTypeRegistration
	if err == nil {
		tokenType = users.TokenTypeLogin
	} else if !errors.Is(err, users.ErrNotFound) {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	token, err := h.users.CreateVerificationToken(ctx, req.Email, tokenType, time.Now().Add(10*time.Minute))
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	event := events.Event{
		Type: events.EmailVerification,
		Payload: events.EmailVerificationPayload{
			Email:     req.Email,
			IsMobile:  req.IsMobile,
			Token:     token.Token,
			TokenType: string(tokenType),
		},
		Timestamp: time.Now(),
		ActorID:   uuid.Nil,
	}

	if err := h.publisher.Publish(ctx, event); err != nil {
		return web.RespondError(ctx, w, fmt.Errorf("failed to publish email verification event: %w", err), http.StatusInternalServerError)
	}

	return web.Respond(ctx, w, nil, http.StatusNoContent)
}

func (h *Handlers) VerifyEmail(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	var req VerifyEmailRequest
	if err := web.Decode(r, &req); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	normalizedEmail, err := validate.Email(req.Email)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	req.Email = normalizedEmail

	verificationToken, err := h.users.GetVerificationToken(ctx, req.Token)
	if err != nil {
		switch {
		case errors.Is(err, users.ErrTokenExpired):
			return web.RespondError(ctx, w, err, http.StatusGone)
		case errors.Is(err, users.ErrTokenUsed):
			return web.RespondError(ctx, w, err, http.StatusGone)
		case errors.Is(err, users.ErrInvalidToken):
			return web.RespondError(ctx, w, err, http.StatusNotFound)
		default:
			return web.RespondError(ctx, w, err, http.StatusInternalServerError)
		}
	}

	if verificationToken.Email != req.Email {
		return web.RespondError(ctx, w, errors.New("email mismatch"), http.StatusBadRequest)
	}

	if err := h.users.MarkTokenUsed(ctx, verificationToken.ID); err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	user, err := h.users.GetUserByEmail(ctx, req.Email)
	if err != nil && !errors.Is(err, users.ErrNotFound) {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	if errors.Is(err, users.ErrNotFound) {
		newUser := users.CoreNewUser{
			Email:    req.Email,
			Timezone: "Antarctica/Troll", // Default timezone for new users
		}
		user, err = h.users.Register(ctx, newUser)
		if err != nil {
			return web.RespondError(ctx, w, err, http.StatusInternalServerError)
		}
	}

	tokenString, err := h.createSessionToken()
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	expiresAt := time.Now().Add(SessionDuration)
	if err := h.persistSession(ctx, user.ID, tokenString, expiresAt); err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}
	h.setSessionCookie(w, r, tokenString, expiresAt)
	h.resolveUserAvatar(ctx, &user)
	return web.Respond(ctx, w, toAppUser(user), http.StatusOK)
}

func (h *Handlers) GetAutomationPreferences(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "handlers.users.GetAutomationPreferences")
	defer span.End()

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	preferences, err := h.users.GetAutomationPreferences(ctx, userID, workspace.ID)
	if err != nil {
		return web.RespondError(ctx, w, fmt.Errorf("failed to get automation preferences: %w", err), http.StatusInternalServerError)
	}

	span.AddEvent("automation preferences retrieved")
	return web.Respond(ctx, w, toAppAutomationPreferences(preferences), http.StatusOK)
}

func (h *Handlers) UpdateAutomationPreferences(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "handlers.users.UpdateAutomationPreferences")
	defer span.End()

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	var req UpdateAutomationPreferencesRequest
	if err := web.Decode(r, &req); err != nil {
		return web.RespondError(ctx, w, fmt.Errorf("invalid request: %w", err), http.StatusBadRequest)
	}

	updates := toCoreUpdateAutomationPreferences(req)
	if err := h.users.UpdateAutomationPreferences(ctx, userID, workspace.ID, updates); err != nil {
		return web.RespondError(ctx, w, fmt.Errorf("failed to update automation preferences: %w", err), http.StatusInternalServerError)
	}

	preferences, err := h.users.GetAutomationPreferences(ctx, userID, workspace.ID)
	if err != nil {
		return web.RespondError(ctx, w, fmt.Errorf("failed to get updated automation preferences: %w", err), http.StatusInternalServerError)
	}

	span.AddEvent("automation preferences updated")
	return web.Respond(ctx, w, toAppAutomationPreferences(preferences), http.StatusOK)
}

func (h *Handlers) UploadProfileImage(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	if err := r.ParseMultipartForm(6 << 20); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	file, header, err := r.FormFile("image")
	if err != nil {
		return web.RespondError(ctx, w, fmt.Errorf("error getting image file: %w", err), http.StatusBadRequest)
	}
	defer file.Close()

	err = h.users.UploadProfileImage(ctx, userID, file, header, h.attachments)
	if err != nil {
		switch {
		case errors.Is(err, validate.ErrFileTooLarge), errors.Is(err, validate.ErrInvalidFileType):
			return web.RespondError(ctx, w, err, http.StatusBadRequest)
		default:
			return fmt.Errorf("error uploading profile image: %w", err)
		}
	}

	user, err := h.users.GetUser(ctx, userID)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	h.resolveUserAvatar(ctx, &user)
	return web.Respond(ctx, w, toAppUser(user), http.StatusOK)
}

func (h *Handlers) DeleteProfileImage(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	err = h.users.DeleteProfileImage(ctx, userID, h.attachments)
	if err != nil {
		if errors.Is(err, users.ErrNotFound) {
			return web.RespondError(ctx, w, err, http.StatusNotFound)
		}
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	user, err := h.users.GetUser(ctx, userID)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	h.resolveUserAvatar(ctx, &user)
	return web.Respond(ctx, w, toAppUser(user), http.StatusOK)
}
func (h *Handlers) GenerateSessionCode(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	// Get user details
	user, err := h.users.GetUser(ctx, userID)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	// Check rate limiting (5-minute lookback for mobile codes)
	count, err := h.users.GetValidTokenCount(ctx, user.Email, 5*time.Minute)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	if count >= 3 {
		return web.RespondError(ctx, w, users.ErrTooManyAttempts, http.StatusTooManyRequests)
	}

	// Create verification token with 5-minute expiry
	token, err := h.users.CreateVerificationToken(ctx, user.Email, users.TokenTypeLogin, time.Now().Add(5*time.Minute))
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	response := GenerateSessionCodeResponse{
		Code:  token.Token,
		Email: user.Email,
	}

	return web.Respond(ctx, w, response, http.StatusOK)
}

// CreateSession exchanges a valid auth token for a secure session cookie.
func (h *Handlers) CreateSession(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	tokenString, err := h.createSessionToken()
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	expires := time.Now().Add(SessionDuration)
	if err := h.persistSession(ctx, userID, tokenString, expires); err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}
	h.setSessionCookie(w, r, tokenString, expires)

	return web.Respond(ctx, w, map[string]bool{"ok": true}, http.StatusOK)
}

// ClearSession clears the auth session cookie.
func (h *Handlers) ClearSession(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	if h.cache != nil {
		if cookie, err := r.Cookie(sessionCookieName); err == nil && cookie.Value != "" {
			_ = h.cache.Delete(ctx, cache.AuthSessionCacheKey(cookie.Value))
		}
	}
	h.clearSessionCookie(w, r)
	return web.Respond(ctx, w, nil, http.StatusNoContent)
}

// AddUserMemory adds a new memory item for the user.
func (h *Handlers) AddUserMemory(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "handlers.users.AddUserMemory")
	defer span.End()

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	var req AddUserMemoryRequest
	if err := web.Decode(r, &req); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	newItem := users.NewUserMemoryItem{
		UserID:      userID,
		WorkspaceID: workspace.ID,
		Content:     req.Content,
	}

	item, err := h.users.AddUserMemory(ctx, newItem)
	if err != nil {
		return web.RespondError(ctx, w, fmt.Errorf("adding user memory: %w", err), http.StatusInternalServerError)
	}

	return web.Respond(ctx, w, toAppUserMemoryItem(item), http.StatusCreated)
}

// UpdateUserMemory updates a memory item.
func (h *Handlers) UpdateUserMemory(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "handlers.users.UpdateUserMemory")
	defer span.End()

	idParam := web.Params(r, "id")
	memoryID, err := uuid.Parse(idParam)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	var req UpdateUserMemoryRequest
	if err := web.Decode(r, &req); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	update := users.UpdateUserMemoryItem{
		Content: &req.Content,
	}

	if err := h.users.UpdateUserMemory(ctx, memoryID, update); err != nil {
		return web.RespondError(ctx, w, fmt.Errorf("updating user memory: %w", err), http.StatusInternalServerError)
	}

	return web.Respond(ctx, w, nil, http.StatusNoContent)
}

// DeleteUserMemory deletes a memory item.
func (h *Handlers) DeleteUserMemory(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "handlers.users.DeleteUserMemory")
	defer span.End()

	idParam := web.Params(r, "id")
	memoryID, err := uuid.Parse(idParam)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	if err := h.users.DeleteUserMemory(ctx, memoryID); err != nil {
		return web.RespondError(ctx, w, fmt.Errorf("deleting user memory: %w", err), http.StatusInternalServerError)
	}

	return web.Respond(ctx, w, nil, http.StatusNoContent)
}

// ListUserMemories retrieves all memory items for the user in the workspace.
func (h *Handlers) ListUserMemories(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "handlers.users.ListUserMemories")
	defer span.End()

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	items, err := h.users.ListUserMemories(ctx, userID, workspace.ID)
	if err != nil {
		return web.RespondError(ctx, w, fmt.Errorf("listing user memories: %w", err), http.StatusInternalServerError)
	}

	return web.Respond(ctx, w, toAppUserMemoryItems(items), http.StatusOK)
}
