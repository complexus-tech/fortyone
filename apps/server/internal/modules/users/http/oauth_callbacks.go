package usershttp

import (
	"context"
	"errors"
	"net/http"
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

type googleAuthenticationService interface {
	VerifyToken(context.Context, string) (google.Identity, error)
	AuthCodeURL(string) (string, error)
	ExchangeCode(context.Context, string) (google.Identity, error)
}

type microsoftAuthenticationService interface {
	AuthCodeURL(string, string, string) (string, error)
	ExchangeCode(context.Context, string, string, string) (microsoft.Identity, error)
}

const (
	oauthFailureGeneric     = "oauth_failed"
	oauthFailureCancelled   = "oauth_cancelled"
	oauthFailureExpired     = "oauth_expired"
	oauthFailureUnavailable = "account_unavailable"
)

func (h *Handlers) GoogleAuth(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	var req GoogleAuthRequest
	if err := web.Decode(r, &req); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	if h.googleService == nil {
		return web.RespondError(ctx, w, google.ErrNotConfigured, http.StatusServiceUnavailable)
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
	callbackURL, err := oauthCallbackURLQuery(r)
	if err == nil {
		callbackURL, err = sanitizeCallbackURL(callbackURL, h.cookieDomain, h.websiteURL)
	}
	if err != nil {
		return h.redirectOAuthFailure(ctx, w, r, "", oauthFailureGeneric)
	}
	if h.cache == nil || h.googleService == nil {
		return h.redirectOAuthFailure(ctx, w, r, callbackURL, oauthFailureGeneric)
	}

	state, err := h.createSessionToken()
	if err != nil {
		return h.redirectOAuthFailure(ctx, w, r, callbackURL, oauthFailureGeneric)
	}

	if err := h.cache.Set(ctx, cache.AuthGoogleStateCacheKey(state), googleAuthState{CallbackURL: callbackURL}, oauthStateTTL); err != nil {
		return h.redirectOAuthFailure(ctx, w, r, callbackURL, oauthFailureGeneric)
	}

	authURL, err := h.googleService.AuthCodeURL(state)
	if err != nil {
		_ = h.cache.Delete(ctx, cache.AuthGoogleStateCacheKey(state))
		return h.redirectOAuthFailure(ctx, w, r, callbackURL, oauthFailureGeneric)
	}

	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
	return nil
}

func (h *Handlers) StartMicrosoftAuth(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	callbackURL, err := oauthCallbackURLQuery(r)
	if err == nil {
		callbackURL, err = sanitizeCallbackURL(callbackURL, h.cookieDomain, h.websiteURL)
	}
	if err != nil {
		return h.redirectOAuthFailure(ctx, w, r, "", oauthFailureGeneric)
	}
	if h.cache == nil || h.microsoftService == nil {
		return h.redirectOAuthFailure(ctx, w, r, callbackURL, oauthFailureGeneric)
	}
	state, err := h.createSessionToken()
	if err != nil {
		return h.redirectOAuthFailure(ctx, w, r, callbackURL, oauthFailureGeneric)
	}
	nonce, err := h.createSessionToken()
	if err != nil {
		return h.redirectOAuthFailure(ctx, w, r, callbackURL, oauthFailureGeneric)
	}
	verifier := oauth2.GenerateVerifier()
	if err := h.cache.Set(ctx, cache.AuthMicrosoftStateCacheKey(state), microsoftAuthState{
		CallbackURL: callbackURL,
		Verifier:    verifier,
		Nonce:       nonce,
	}, oauthStateTTL); err != nil {
		return h.redirectOAuthFailure(ctx, w, r, callbackURL, oauthFailureGeneric)
	}

	authURL, err := h.microsoftService.AuthCodeURL(state, nonce, verifier)
	if err != nil {
		_ = h.cache.Delete(ctx, cache.AuthMicrosoftStateCacheKey(state))
		return h.redirectOAuthFailure(ctx, w, r, callbackURL, oauthFailureGeneric)
	}
	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
	return nil
}

func (h *Handlers) CompleteMicrosoftAuth(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	if h.cache == nil {
		return h.redirectOAuthFailure(ctx, w, r, "", oauthFailureGeneric)
	}
	if h.microsoftService == nil {
		return h.redirectOAuthFailure(ctx, w, r, "", oauthFailureGeneric)
	}

	callback, err := web.ParseOAuthCallbackQuery(r.URL.Query())
	if err != nil {
		return h.redirectOAuthFailure(ctx, w, r, "", oauthFailureGeneric)
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
		return h.redirectOAuthFailure(ctx, w, r, "", oauthStateFailureCode(err))
	}

	if providerError := callback.ProviderError; providerError != "" {
		return h.redirectOAuthFailure(ctx, w, r, authState.CallbackURL, oauthProviderFailureCode(providerError))
	}

	identity, err := h.microsoftService.ExchangeCode(ctx, callback.Code, authState.Verifier, authState.Nonce)
	if err != nil {
		return h.redirectOAuthFailure(ctx, w, r, authState.CallbackURL, oauthFailureGeneric)
	}

	email, err := validate.Email(identity.Email)
	if err != nil {
		return h.redirectOAuthFailure(ctx, w, r, authState.CallbackURL, oauthFailureGeneric)
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
		return h.redirectOAuthFailure(ctx, w, r, authState.CallbackURL, oauthSignInFailureCode(err))
	}
	user, err = h.reactivateUserForSignIn(ctx, user)
	if err != nil {
		return h.redirectOAuthFailure(ctx, w, r, authState.CallbackURL, oauthSignInFailureCode(err))
	}

	tokenString, err := h.createSessionToken()
	if err != nil {
		return h.redirectOAuthFailure(ctx, w, r, authState.CallbackURL, oauthFailureGeneric)
	}
	expiresAt := time.Now().Add(SessionDuration)
	if err := h.persistSession(ctx, user.ID, tokenString, expiresAt); err != nil {
		return h.redirectOAuthFailure(ctx, w, r, authState.CallbackURL, oauthSignInFailureCode(err))
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

func (h *Handlers) CompleteGoogleAuth(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	if h.cache == nil {
		return h.redirectOAuthFailure(ctx, w, r, "", oauthFailureGeneric)
	}
	if h.googleService == nil {
		return h.redirectOAuthFailure(ctx, w, r, "", oauthFailureGeneric)
	}

	callback, err := web.ParseOAuthCallbackQuery(r.URL.Query())
	if err != nil {
		return h.redirectOAuthFailure(ctx, w, r, "", oauthFailureGeneric)
	}

	var authState googleAuthState
	if err := takeOAuthState(
		ctx,
		h.cache,
		cache.AuthGoogleStateCacheKey(callback.State),
		cache.LegacyAuthGoogleStateCacheKey(callback.State),
		&authState,
	); err != nil {
		return h.redirectOAuthFailure(ctx, w, r, "", oauthStateFailureCode(err))
	}
	if callback.ProviderError != "" {
		return h.redirectOAuthFailure(ctx, w, r, authState.CallbackURL, oauthProviderFailureCode(callback.ProviderError))
	}

	identity, err := h.googleService.ExchangeCode(ctx, callback.Code)
	if err != nil {
		return h.redirectOAuthFailure(ctx, w, r, authState.CallbackURL, oauthFailureGeneric)
	}

	if !identity.EmailVerified {
		return h.redirectOAuthFailure(ctx, w, r, authState.CallbackURL, oauthFailureGeneric)
	}

	user, err := h.authenticateWithGoogleIdentity(ctx, w, r, identity)
	if err != nil {
		return h.redirectOAuthFailure(ctx, w, r, authState.CallbackURL, oauthSignInFailureCode(err))
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
