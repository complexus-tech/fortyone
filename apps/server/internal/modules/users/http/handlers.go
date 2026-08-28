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
	"unicode"

	users "github.com/complexus-tech/projects-api/internal/modules/users/service"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	"github.com/complexus-tech/projects-api/internal/platform/deployment"
	"github.com/complexus-tech/projects-api/pkg/cache"
	"github.com/complexus-tech/projects-api/pkg/google"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/microsoft"
	"github.com/complexus-tech/projects-api/pkg/publisher"
	"github.com/google/uuid"
)

var (
	ErrInvalidWorkspaceID = errors.New("workspace id is not in its proper form")
	SessionDuration       = time.Hour * 24 * 30
	avatarAccessURLExpiry = 24 * time.Hour
)

const sessionCookieName = "fortyone_session"

const oauthStateTTL = 10 * time.Minute

type Handlers struct {
	users                  *users.Service
	attachments            users.AttachmentsService
	secretKey              string
	cookieDomain           string
	websiteURL             string
	cache                  *cache.Service
	log                    *logger.Logger
	deploymentMode         deployment.Mode
	verificationRateLimits verificationRateLimitStore
	googleService          *google.Service
	microsoftService       *microsoft.Service
	publisher              *publisher.Publisher
}

func New(users *users.Service, attachments users.AttachmentsService, secretKey, cookieDomain, websiteURL string, cacheService *cache.Service, log *logger.Logger, deploymentMode deployment.Mode, googleService *google.Service, microsoftService *microsoft.Service, publisher *publisher.Publisher) *Handlers {
	return &Handlers{
		users:                  users,
		attachments:            attachments,
		secretKey:              secretKey,
		cookieDomain:           cookieDomain,
		websiteURL:             websiteURL,
		cache:                  cacheService,
		log:                    log,
		deploymentMode:         deploymentMode,
		verificationRateLimits: cacheService,
		googleService:          googleService,
		microsoftService:       microsoftService,
		publisher:              publisher,
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
	if h.users == nil {
		return errors.New("user service is not configured")
	}

	version, active, err := h.users.ResolveActiveBrowserSessionVersion(ctx, userID)
	if err != nil {
		return fmt.Errorf("resolve browser session account: %w", err)
	}
	if !active {
		return platformauth.ErrInvalidBrowserSession
	}
	session, err := platformauth.NewBrowserSession(userID, version)
	if err != nil {
		return err
	}

	return h.cache.Set(ctx, cache.AuthSessionCacheKey(token), session, ttl)
}

func (h *Handlers) setSessionCookie(w http.ResponseWriter, r *http.Request, value string, expires time.Time) {
	cookie := http.Cookie{ // #nosec G124 -- Secure is required for HTTPS and deliberately disabled only for local HTTP development.
		Name:     sessionCookieName,
		Value:    value,
		Path:     "/",
		HttpOnly: true,
		Secure:   h.secureSessionCookie(r),
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
	cookie := http.Cookie{ // #nosec G124 -- The deletion cookie must exactly match the request-scoped Secure policy of the session cookie.
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   h.secureSessionCookie(r),
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

func (h *Handlers) secureSessionCookie(r *http.Request) bool {
	if h.deploymentMode.IsProduction() {
		return true
	}
	if r.TLS != nil {
		return true
	}
	return strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https")
}

const maxCallbackURLLength = 2048

func isLocalCallbackHostname(hostname string) bool {
	return hostname == "localhost" ||
		hostname == "127.0.0.1" ||
		strings.HasSuffix(hostname, ".localhost")
}

func configuredWebsiteHostname(websiteURL string) string {
	parsed, err := url.Parse(strings.TrimSpace(websiteURL))
	if err != nil {
		return ""
	}
	return strings.ToLower(strings.TrimSuffix(parsed.Hostname(), "."))
}

func isAllowedCallbackHostname(hostname, configuredDomain, websiteURL string) bool {
	hostname = strings.ToLower(strings.TrimSuffix(hostname, "."))
	configuredDomain = strings.ToLower(strings.Trim(strings.TrimSpace(configuredDomain), "."))
	websiteHostname := configuredWebsiteHostname(websiteURL)

	if hostname == websiteHostname && websiteHostname != "" {
		return true
	}
	if isLocalCallbackHostname(hostname) && isLocalCallbackHostname(websiteHostname) {
		return true
	}
	if configuredDomain != "" &&
		(hostname == configuredDomain || strings.HasSuffix(hostname, "."+configuredDomain)) {
		return true
	}

	return hostname == "fortyone.app" || strings.HasSuffix(hostname, ".fortyone.app")
}

func sanitizeCallbackURL(raw, configuredDomain, websiteURL string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", nil
	}
	if len(raw) > maxCallbackURLLength {
		return "", errors.New("invalid callbackURL")
	}
	if strings.IndexFunc(raw, func(character rune) bool {
		return character == '\\' || unicode.IsControl(character)
	}) >= 0 {
		return "", errors.New("invalid callbackURL")
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
	if parsed.User != nil {
		return "", errors.New("invalid callbackURL")
	}
	if !isAllowedCallbackHostname(parsed.Hostname(), configuredDomain, websiteURL) {
		return "", errors.New("invalid callbackURL")
	}
	if parsed.Scheme == "http" && !isLocalCallbackHostname(parsed.Hostname()) {
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

func (h *Handlers) reactivateUserForSignIn(ctx context.Context, user users.CoreUser) (users.CoreUser, error) {
	if user.IsActive {
		return user, nil
	}

	return h.users.ReactivateUserForVerifiedSignIn(ctx, user.ID)
}

func publicSignInError(err error) (int, error) {
	if errors.Is(err, users.ErrInvalidCredentials) || errors.Is(err, platformauth.ErrInvalidBrowserSession) {
		return http.StatusUnauthorized, users.ErrInvalidCredentials
	}
	return http.StatusInternalServerError, err
}
