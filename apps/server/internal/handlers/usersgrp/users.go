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
	SessionDuration       = time.Hour * 24 * 40
)

type Handlers struct {
	users         *users.Service
	attachments   users.AttachmentsService
	secretKey     string
	googleService *google.Service
	publisher     *publisher.Publisher
}

func New(users *users.Service, attachments users.AttachmentsService, secretKey string, googleService *google.Service, publisher *publisher.Publisher) *Handlers {
	return &Handlers{
		users:         users,
		attachments:   attachments,
		secretKey:     secretKey,
		googleService: googleService,
		publisher:     publisher,
	}
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

	return web.Respond(ctx, w, toAppUser(user), http.StatusOK)
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
	web.Respond(ctx, w, toAppUsers(users), http.StatusOK)
	return nil
}

func (h *Handlers) GoogleAuth(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	var req GoogleAuthRequest
	if err := web.Decode(r, &req); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	payload, err := h.googleService.VerifyToken(ctx, req.Token)
	if err != nil {
		switch {
		case errors.Is(err, google.ErrInvalidToken), errors.Is(err, google.ErrEmailMismatch):
			return web.RespondError(ctx, w, err, http.StatusUnauthorized)
		default:
			return web.RespondError(ctx, w, err, http.StatusInternalServerError)
		}
	}

	if verified, ok := payload.Claims["email_verified"].(bool); !ok || !verified {
		return web.RespondError(ctx, w, errors.New("email not verified by google"), http.StatusUnauthorized)
	}

	user, err := h.users.GetUserByEmail(ctx, payload.Claims["email"].(string))
	if err != nil && !errors.Is(err, users.ErrNotFound) {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	if errors.Is(err, users.ErrNotFound) {
		newUser := users.CoreNewUser{
			Email:     payload.Claims["email"].(string),
			FullName:  payload.Claims["name"].(string),
			AvatarURL: payload.Claims["picture"].(string),
			Timezone:  "Antarctica/Troll", // Default timezone for new users
		}
		user, err = h.users.Register(ctx, newUser)
		if err != nil {
			return web.RespondError(ctx, w, err, http.StatusInternalServerError)
		}
	}

	claims := jwt.RegisteredClaims{
		Subject:   user.ID.String(),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(SessionDuration)),
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

	claims := jwt.RegisteredClaims{
		Subject:   user.ID.String(),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(SessionDuration)),
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
