package usersgrp

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/complexus-tech/projects-api/internal/core/users"
	"github.com/complexus-tech/projects-api/internal/web/mid"
	"github.com/complexus-tech/projects-api/pkg/events"
	"github.com/complexus-tech/projects-api/pkg/google"
	"github.com/complexus-tech/projects-api/pkg/publisher"
	"github.com/complexus-tech/projects-api/pkg/validate"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	ErrInvalidWorkspaceID = errors.New("workspace id is not in its proper form")
)

type Handlers struct {
	users         *users.Service
	secretKey     string
	googleService *google.Service
	publisher     *publisher.Publisher
}

// New constructs a new users handlers instance.
func New(users *users.Service, secretKey string, googleService *google.Service, publisher *publisher.Publisher) *Handlers {
	return &Handlers{
		users:         users,
		secretKey:     secretKey,
		googleService: googleService,
		publisher:     publisher,
	}
}

// GetProfile returns the current user's profile
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

	return web.Respond(ctx, w, toAppUser(user), http.StatusOK)
}

// UpdateProfile updates the current user's profile
func (h *Handlers) UpdateProfile(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	var req UpdateProfileRequest
	if err := web.Decode(r, &req); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	updates := users.CoreUpdateUser{
		Username:  req.Username,
		FullName:  req.FullName,
		AvatarURL: req.AvatarURL,
	}

	if err := h.users.UpdateUser(ctx, userID, updates); err != nil {
		if errors.Is(err, users.ErrNotFound) {
			return web.RespondError(ctx, w, err, http.StatusNotFound)
		}
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	// Get updated user profile
	user, err := h.users.GetUser(ctx, userID)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	return web.Respond(ctx, w, toAppUser(user), http.StatusOK)
}

// DeleteProfile soft deletes the current user's account
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

// SwitchWorkspace updates the current user's active workspace
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

	// Get updated user profile
	user, err := h.users.GetUser(ctx, userID)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	return web.Respond(ctx, w, toAppUser(user), http.StatusOK)
}

// List returns a list of users for a workspace.
func (h *Handlers) List(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspaceIdParam := web.Params(r, "workspaceId")
	workspaceId, err := uuid.Parse(workspaceIdParam)
	if err != nil {
		web.RespondError(ctx, w, ErrInvalidWorkspaceID, http.StatusBadRequest)
		return nil
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

	users, err := h.users.List(ctx, workspaceId, teamID)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}
	web.Respond(ctx, w, toAppUsers(users), http.StatusOK)
	return nil
}

// GoogleAuth handles authentication with Google
func (h *Handlers) GoogleAuth(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	var req GoogleAuthRequest
	if err := web.Decode(r, &req); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	// Verify Google token
	payload, err := h.googleService.VerifyToken(ctx, req.Token)
	if err != nil {
		switch {
		case errors.Is(err, google.ErrInvalidToken), errors.Is(err, google.ErrEmailMismatch):
			return web.RespondError(ctx, w, err, http.StatusUnauthorized)
		default:
			return web.RespondError(ctx, w, err, http.StatusInternalServerError)
		}
	}

	// Verify email is verified by Google
	if verified, ok := payload.Claims["email_verified"].(bool); !ok || !verified {
		return web.RespondError(ctx, w, errors.New("email not verified by google"), http.StatusUnauthorized)
	}

	// Try to find existing user
	user, err := h.users.GetUserByEmail(ctx, payload.Claims["email"].(string))
	if err != nil && !errors.Is(err, users.ErrNotFound) {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	if errors.Is(err, users.ErrNotFound) {
		// Create new user
		newUser := users.CoreNewUser{
			Email:     payload.Claims["email"].(string),
			FullName:  payload.Claims["name"].(string),
			AvatarURL: payload.Claims["picture"].(string),
		}
		user, err = h.users.Register(ctx, newUser)
		if err != nil {
			return web.RespondError(ctx, w, err, http.StatusInternalServerError)
		}
	}

	// Generate JWT token
	claims := jwt.RegisteredClaims{
		Subject:   user.ID.String(),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24 * 14)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		NotBefore: jwt.NewNumericDate(time.Now()),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(h.secretKey))
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	user.Token = &tokenString
	return web.Respond(ctx, w, toAppUser(user), http.StatusOK)
}

// SendEmailVerification sends a verification email to the user
func (h *Handlers) SendEmailVerification(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	var req EmailVerificationRequest
	if err := web.Decode(r, &req); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	// Validate and normalize email
	normalizedEmail, err := validate.Email(req.Email)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	req.Email = normalizedEmail

	// Check if too many attempts
	count, err := h.users.GetValidTokenCount(ctx, req.Email, 10*time.Minute)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	if count >= 3 {
		return web.RespondError(ctx, w, users.ErrTooManyAttempts, http.StatusTooManyRequests)
	}

	// Check if user exists to determine token type
	_, err = h.users.GetUserByEmail(ctx, req.Email)
	tokenType := users.TokenTypeRegistration
	if err == nil {
		tokenType = users.TokenTypeLogin
	} else if !errors.Is(err, users.ErrNotFound) {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	// Create new verification token
	token, err := h.users.CreateVerificationToken(ctx, req.Email, tokenType)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	// Publish email verification event
	event := events.Event{
		Type: events.EmailVerification,
		Payload: events.EmailVerificationPayload{
			Email:     req.Email,
			Token:     token.Token,
			TokenType: string(tokenType),
		},
		Timestamp: time.Now(),
		ActorID:   uuid.Nil, // System generated event
	}

	if err := h.publisher.Publish(ctx, event); err != nil {
		return web.RespondError(ctx, w, fmt.Errorf("failed to publish email verification event: %w", err), http.StatusInternalServerError)
	}

	return web.Respond(ctx, w, nil, http.StatusNoContent)
}

// VerifyEmail verifies an email verification token
func (h *Handlers) VerifyEmail(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	var req VerifyEmailRequest
	if err := web.Decode(r, &req); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	// Validate and normalize email
	normalizedEmail, err := validate.Email(req.Email)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	req.Email = normalizedEmail

	// Get and validate token
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

	// Verify email matches
	if verificationToken.Email != req.Email {
		return web.RespondError(ctx, w, errors.New("email mismatch"), http.StatusBadRequest)
	}

	// Mark token as used
	if err := h.users.MarkTokenUsed(ctx, verificationToken.ID); err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	// Try to find existing user
	user, err := h.users.GetUserByEmail(ctx, req.Email)
	if err != nil && !errors.Is(err, users.ErrNotFound) {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	if errors.Is(err, users.ErrNotFound) {
		// Create new user with verified email
		newUser := users.CoreNewUser{
			Email: req.Email,
		}
		user, err = h.users.Register(ctx, newUser)
		if err != nil {
			return web.RespondError(ctx, w, err, http.StatusInternalServerError)
		}
	}

	// Generate JWT token
	claims := jwt.RegisteredClaims{
		Subject:   user.ID.String(),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24 * 14)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		NotBefore: jwt.NewNumericDate(time.Now()),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(h.secretKey))
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	user.Token = &tokenString
	return web.Respond(ctx, w, toAppUser(user), http.StatusOK)
}

// GetAutomationPreferences retrieves the user's automation preferences for a workspace
func (h *Handlers) GetAutomationPreferences(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "handlers.users.GetAutomationPreferences")
	defer span.End()

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	workspaceID, err := uuid.Parse(web.Params(r, "workspaceId"))
	if err != nil {
		return web.RespondError(ctx, w, fmt.Errorf("invalid workspace ID: %w", err), http.StatusBadRequest)
	}

	preferences, err := h.users.GetAutomationPreferences(ctx, userID, workspaceID)
	if err != nil {
		return web.RespondError(ctx, w, fmt.Errorf("failed to get automation preferences: %w", err), http.StatusInternalServerError)
	}

	span.AddEvent("automation preferences retrieved")
	return web.Respond(ctx, w, toAppAutomationPreferences(preferences), http.StatusOK)
}

// UpdateAutomationPreferences updates the user's automation preferences for a workspace
func (h *Handlers) UpdateAutomationPreferences(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "handlers.users.UpdateAutomationPreferences")
	defer span.End()

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	workspaceID, err := uuid.Parse(web.Params(r, "workspaceId"))
	if err != nil {
		return web.RespondError(ctx, w, fmt.Errorf("invalid workspace ID: %w", err), http.StatusBadRequest)
	}

	var req UpdateAutomationPreferencesRequest
	if err := web.Decode(r, &req); err != nil {
		return web.RespondError(ctx, w, fmt.Errorf("invalid request: %w", err), http.StatusBadRequest)
	}

	updates := toCoreUpdateAutomationPreferences(req)
	if err := h.users.UpdateAutomationPreferences(ctx, userID, workspaceID, updates); err != nil {
		return web.RespondError(ctx, w, fmt.Errorf("failed to update automation preferences: %w", err), http.StatusInternalServerError)
	}

	// Get the updated preferences to return to the client
	preferences, err := h.users.GetAutomationPreferences(ctx, userID, workspaceID)
	if err != nil {
		return web.RespondError(ctx, w, fmt.Errorf("failed to get updated automation preferences: %w", err), http.StatusInternalServerError)
	}

	span.AddEvent("automation preferences updated")
	return web.Respond(ctx, w, toAppAutomationPreferences(preferences), http.StatusOK)
}
