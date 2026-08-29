package usershttp

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	users "github.com/complexus-tech/projects-api/internal/modules/users/service"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/cache"
	"github.com/complexus-tech/projects-api/pkg/validate"
	"github.com/complexus-tech/projects-api/pkg/web"
)

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

	const multipartOverheadAllowance int64 = 1 << 20
	if err := web.ParseMultipartForm(w, r, validate.MaxProfileImageSize+multipartOverheadAllowance); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	defer func() {
		if err := web.RemoveMultipartForm(r); err != nil && h.log != nil {
			h.log.Warn(ctx, "failed to remove profile image upload temporary files", "error", err)
		}
	}()

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
			_ = h.cache.Delete(ctx, cache.LegacyAuthSessionCacheKey(cookie.Value))
		}
	}
	h.clearSessionCookie(w, r)
	return web.Respond(ctx, w, nil, http.StatusNoContent)
}

// AddUserMemory adds a new memory item for the user.
