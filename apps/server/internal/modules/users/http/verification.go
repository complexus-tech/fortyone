package usershttp

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	users "github.com/complexus-tech/projects-api/internal/modules/users/service"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/events"
	"github.com/complexus-tech/projects-api/pkg/validate"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

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
	callbackURL, err := sanitizeCallbackURL(req.CallbackURL, h.cookieDomain, h.websiteURL)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	allowed, err := h.enforceVerificationRateLimits(
		ctx,
		w,
		"send",
		verificationRateLimitIdentity{kind: "email", value: req.Email, limit: verificationSendEmailLimit},
		verificationRateLimitIdentity{kind: "network", value: verificationNetworkIdentity(r), limit: verificationSendNetworkLimit},
	)
	if err != nil || !allowed {
		return err
	}

	// The confirm endpoint determines whether this email belongs to an existing
	// user after the token is validated. Avoid blocking link delivery on a
	// request-time user lookup.
	tokenType := users.TokenTypeRegistration
	token, err := h.users.CreateVerificationToken(ctx, req.Email, tokenType, time.Now().Add(10*time.Minute))
	if err != nil {
		return respondVerificationTokenCreationError(ctx, w, err)
	}

	event := events.Event{
		Type: events.EmailVerification,
		Payload: events.EmailVerificationPayload{
			Email:       req.Email,
			IsMobile:    req.IsMobile,
			Token:       token.Token,
			TokenType:   tokenType,
			CallbackURL: callbackURL,
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

	allowed, err := h.enforceVerificationRateLimits(
		ctx,
		w,
		"confirm",
		verificationRateLimitIdentity{kind: "email", value: req.Email, limit: verificationConfirmEmailLimit},
		verificationRateLimitIdentity{kind: "token", value: req.Token, limit: verificationConfirmTokenLimit},
		verificationRateLimitIdentity{kind: "network", value: verificationNetworkIdentity(r), limit: verificationConfirmNetworkLimit},
	)
	if err != nil || !allowed {
		return err
	}

	_, err = h.users.ConsumeVerificationToken(
		ctx,
		req.Email,
		req.Token,
		users.TokenTypeRegistration,
		users.TokenTypeLogin,
	)
	if err != nil {
		switch {
		case errors.Is(err, users.ErrTokenExpired),
			errors.Is(err, users.ErrTokenUsed),
			errors.Is(err, users.ErrInvalidToken):
			return web.RespondError(ctx, w, users.ErrInvalidToken, http.StatusBadRequest)
		default:
			return web.RespondError(ctx, w, err, http.StatusInternalServerError)
		}
	}

	user, err := h.users.GetUserByEmailAnyStatus(ctx, req.Email)
	if err != nil && !errors.Is(err, users.ErrNotFound) {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}
	if errors.Is(err, users.ErrNotFound) {
		user, err = h.users.Register(ctx, users.CoreNewUser{
			Email:    req.Email,
			Timezone: "Antarctica/Troll",
		})
		if err != nil {
			return web.RespondError(ctx, w, err, http.StatusInternalServerError)
		}
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
	h.resolveUserAvatar(ctx, &user)
	return web.Respond(ctx, w, toAppUser(user), http.StatusOK)
}

func (h *Handlers) GenerateSessionCode(ctx context.Context, w http.ResponseWriter, _ *http.Request) error {
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	user, err := h.users.GetUser(ctx, userID)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	token, err := h.users.CreateVerificationToken(ctx, user.Email, users.TokenTypeLogin, time.Now().Add(5*time.Minute))
	if err != nil {
		return respondVerificationTokenCreationError(ctx, w, err)
	}

	return web.Respond(ctx, w, GenerateSessionCodeResponse{
		Code:  token.Token,
		Email: user.Email,
	}, http.StatusOK)
}

func respondVerificationTokenCreationError(ctx context.Context, w http.ResponseWriter, err error) error {
	if errors.Is(err, users.ErrTooManyAttempts) {
		return web.RespondError(ctx, w, users.ErrTooManyAttempts, http.StatusTooManyRequests)
	}
	return web.RespondError(ctx, w, err, http.StatusInternalServerError)
}
