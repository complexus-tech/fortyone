package usershttp

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"

	users "github.com/complexus-tech/projects-api/internal/modules/users/service"
	"github.com/complexus-tech/projects-api/pkg/cache"
	"github.com/complexus-tech/projects-api/pkg/google"
	"github.com/complexus-tech/projects-api/pkg/microsoft"
	"github.com/complexus-tech/projects-api/pkg/validate"
	"github.com/complexus-tech/projects-api/pkg/web"
	"golang.org/x/oauth2"
)

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
		status, publicError := publicSignInError(err)
		return web.RespondError(ctx, w, publicError, status)
	}

	return web.Respond(ctx, w, toAppUser(user), http.StatusOK)
}

func (h *Handlers) StartGoogleAuth(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	if h.cache == nil {
		return web.RespondError(ctx, w, errors.New("auth session cache is not configured"), http.StatusServiceUnavailable)
	}

	callbackURL, err := oauthCallbackURLQuery(r)
	if err == nil {
		callbackURL, err = sanitizeCallbackURL(callbackURL, h.cookieDomain, h.websiteURL)
	}
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	state, err := h.createSessionToken()
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	if err := h.cache.Set(ctx, cache.AuthGoogleStateCacheKey(state), googleAuthState{CallbackURL: callbackURL}, oauthStateTTL); err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	authURL, err := h.googleService.AuthCodeURL(state)
	if err != nil {
		_ = h.cache.Delete(ctx, cache.AuthGoogleStateCacheKey(state))
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

func (h *Handlers) StartMicrosoftAuth(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	if h.cache == nil {
		return web.RespondError(ctx, w, errors.New("auth session cache is not configured"), http.StatusServiceUnavailable)
	}
	if h.microsoftService == nil {
		return web.RespondError(ctx, w, microsoft.ErrNotConfigured, http.StatusServiceUnavailable)
	}

	callbackURL, err := oauthCallbackURLQuery(r)
	if err == nil {
		callbackURL, err = sanitizeCallbackURL(callbackURL, h.cookieDomain, h.websiteURL)
	}
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	state, err := h.createSessionToken()
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}
	nonce, err := h.createSessionToken()
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}
	verifier := oauth2.GenerateVerifier()
	if err := h.cache.Set(ctx, cache.AuthMicrosoftStateCacheKey(state), microsoftAuthState{
		CallbackURL: callbackURL,
		Verifier:    verifier,
		Nonce:       nonce,
	}, oauthStateTTL); err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	authURL, err := h.microsoftService.AuthCodeURL(state, nonce, verifier)
	if err != nil {
		_ = h.cache.Delete(ctx, cache.AuthMicrosoftStateCacheKey(state))
		status := http.StatusInternalServerError
		if errors.Is(err, microsoft.ErrNotConfigured) {
			status = http.StatusServiceUnavailable
		}
		return web.RespondError(ctx, w, err, status)
	}
	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
	return nil
}

func (h *Handlers) CompleteMicrosoftAuth(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	if h.cache == nil {
		return web.RespondError(ctx, w, errors.New("auth session cache is not configured"), http.StatusServiceUnavailable)
	}
	if h.microsoftService == nil {
		return web.RespondError(ctx, w, microsoft.ErrNotConfigured, http.StatusServiceUnavailable)
	}

	callback, err := web.ParseOAuthCallbackQuery(r.URL.Query())
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	state := callback.State
	var authState microsoftAuthState
	if err := takeOAuthState(
		ctx,
		h.cache,
		cache.AuthMicrosoftStateCacheKey(state),
		cache.LegacyAuthMicrosoftStateCacheKey(state),
		&authState,
	); err != nil {
		return web.RespondError(ctx, w, errors.New("invalid oauth state"), http.StatusUnauthorized)
	}

	if providerError := callback.ProviderError; providerError != "" {
		if failureURL := microsoftFailureURL(authState.CallbackURL, providerError); failureURL != "" {
			http.Redirect(w, r, failureURL, http.StatusTemporaryRedirect)
			return nil
		}
		return web.RespondError(ctx, w, errors.New("microsoft sign-in was not completed"), http.StatusUnauthorized)
	}

	identity, err := h.microsoftService.ExchangeCode(ctx, callback.Code, authState.Verifier, authState.Nonce)
	if err != nil {
		status := http.StatusInternalServerError
		switch {
		case errors.Is(err, microsoft.ErrInvalidToken):
			status = http.StatusUnauthorized
		case errors.Is(err, microsoft.ErrNotConfigured):
			status = http.StatusServiceUnavailable
		}
		return web.RespondError(ctx, w, err, status)
	}

	email, err := validate.Email(identity.Email)
	if err != nil {
		return web.RespondError(ctx, w, errors.New("microsoft account did not provide a valid email address"), http.StatusUnauthorized)
	}
	user, err := h.users.AuthenticateExternalIdentity(ctx, users.CoreExternalIdentityInput{
		Provider: "microsoft",
		Issuer:   identity.Issuer,
		Subject:  identity.ObjectID,
		Email:    email,
		FullName: buildMicrosoftFullName(identity, email),
		Timezone: "Antarctica/Troll",
	})
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}
	user, err = h.reactivateUserForSignIn(ctx, user)
	if err != nil {
		status, publicError := publicSignInError(err)
		return web.RespondError(ctx, w, publicError, status)
	}

	tokenString, err := h.createSessionToken()
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}
	expiresAt := time.Now().Add(SessionDuration)
	if err := h.persistSession(ctx, user.ID, tokenString, expiresAt); err != nil {
		status, publicError := publicSignInError(err)
		return web.RespondError(ctx, w, publicError, status)
	}
	h.setSessionCookie(w, r, tokenString, expiresAt)

	if authState.CallbackURL != "" {
		http.Redirect(w, r, authState.CallbackURL, http.StatusTemporaryRedirect)
		return nil
	}
	h.resolveUserAvatar(ctx, &user)
	return web.Respond(ctx, w, toAppUser(user), http.StatusOK)
}

func buildMicrosoftFullName(identity microsoft.Identity, email string) string {
	if fullName := strings.TrimSpace(identity.FullName); fullName != "" {
		return fullName
	}
	fullName := strings.TrimSpace(strings.TrimSpace(identity.FirstName) + " " + strings.TrimSpace(identity.LastName))
	if fullName != "" {
		return fullName
	}
	if preferred := strings.TrimSpace(identity.PreferredUsername); preferred != "" {
		return preferred
	}
	if localPart := strings.TrimSpace(strings.Split(email, "@")[0]); localPart != "" {
		return localPart
	}
	return "User"
}

func microsoftFailureURL(callbackURL, providerError string) string {
	parsed, err := url.Parse(strings.TrimSpace(callbackURL))
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return ""
	}
	query := parsed.Query()
	if parsed.Path == "/auth-callback" {
		parsed.Path = "/"
	}
	message := "Microsoft sign-in failed. Please try again."
	if providerError == "access_denied" {
		message = "Microsoft sign-in was cancelled."
	}
	query.Set("error", message)
	parsed.RawQuery = query.Encode()
	return parsed.String()
}

func (h *Handlers) CompleteGoogleAuth(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	if h.cache == nil {
		return web.RespondError(ctx, w, errors.New("auth session cache is not configured"), http.StatusServiceUnavailable)
	}

	callback, err := web.ParseOAuthCallbackQuery(r.URL.Query())
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	if callback.ProviderError != "" {
		return web.RespondError(ctx, w, errors.New("google sign-in was not completed"), http.StatusBadRequest)
	}

	var authState googleAuthState
	if err := takeOAuthState(
		ctx,
		h.cache,
		cache.AuthGoogleStateCacheKey(callback.State),
		cache.LegacyAuthGoogleStateCacheKey(callback.State),
		&authState,
	); err != nil {
		return web.RespondError(ctx, w, errors.New("invalid oauth state"), http.StatusUnauthorized)
	}

	identity, err := h.googleService.ExchangeCode(ctx, callback.Code)
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
		status, publicError := publicSignInError(err)
		return web.RespondError(ctx, w, publicError, status)
	}

	if authState.CallbackURL != "" {
		http.Redirect(w, r, authState.CallbackURL, http.StatusTemporaryRedirect)
		return nil
	}

	return web.Respond(ctx, w, toAppUser(user), http.StatusOK)
}

func oauthCallbackURLQuery(r *http.Request) (string, error) {
	value, _, err := web.OptionalTextQueryParameter(
		r.URL.Query(),
		"callbackURL",
		maxCallbackURLLength,
		maxCallbackURLLength,
	)
	return value, err
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

	user, err := h.users.GetUserByEmailAnyStatus(ctx, identity.Email)
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

	user, err = h.reactivateUserForSignIn(ctx, user)
	if err != nil {
		return users.CoreUser{}, err
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
