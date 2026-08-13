package feedbackhttp

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"

	feedback "github.com/complexus-tech/projects-api/internal/modules/feedback/service"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

const feedbackIdentityBodyLimit = 16 << 10

func (h *Handlers) RequestContributorVerification(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	var input AppVerificationRequest
	if status, err := decodePublicRequest(w, r, &input, feedbackIdentityBodyLimit); err != nil {
		return web.RespondError(ctx, w, err, status)
	}
	if strings.TrimSpace(input.Source) == "" {
		input.Source = feedback.ContributorSessionSourcePortal
	}
	challenge, err := h.feedback.RequestContributorVerification(
		ctx,
		web.Params(r, "portalSlug"),
		input.Email,
		input.DisplayName,
		input.HideNamePublicly,
		input.Source,
	)
	if errors.Is(err, feedback.ErrVerificationAttempts) {
		return web.Respond(ctx, w, AppVerificationAccepted{Accepted: true}, http.StatusAccepted)
	}
	if err != nil {
		return web.RespondError(ctx, w, err, httpStatus(err))
	}
	return web.Respond(ctx, w, AppVerificationAccepted{Accepted: true, ExpiresAt: &challenge.ExpiresAt}, http.StatusAccepted)
}

func (h *Handlers) ConfirmContributorVerification(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	var input AppVerificationConfirm
	if status, err := decodePublicRequest(w, r, &input, feedbackIdentityBodyLimit); err != nil {
		return web.RespondError(ctx, w, err, status)
	}
	if strings.TrimSpace(input.Source) == "" {
		input.Source = feedback.ContributorSessionSourcePortal
	}
	result, err := h.feedback.ConfirmContributorVerification(
		ctx,
		web.Params(r, "portalSlug"),
		input.Token,
		input.Email,
		input.Code,
		input.Source,
	)
	if err != nil {
		return web.RespondError(ctx, w, err, httpStatus(err))
	}
	unreadCount, err := h.feedback.GetUnreadUpdateCount(ctx, web.Params(r, "portalSlug"), result.Participant)
	if err != nil {
		return web.RespondError(ctx, w, err, httpStatus(err))
	}
	w.Header().Set("Cache-Control", "no-store")
	return web.Respond(ctx, w, toAppSession(result, true, unreadCount), http.StatusOK)
}

func (h *Handlers) GetContributorSession(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	result, err := h.feedback.ResolveContributorSession(ctx, web.Params(r, "portalSlug"), r.Header.Get("Authorization"), "")
	if err != nil {
		return web.RespondError(ctx, w, err, httpStatus(err))
	}
	unreadCount, err := h.feedback.GetUnreadUpdateCount(ctx, web.Params(r, "portalSlug"), result.Participant)
	if err != nil {
		return web.RespondError(ctx, w, err, httpStatus(err))
	}
	w.Header().Set("Cache-Control", "no-store")
	return web.Respond(ctx, w, toAppSession(result, false, unreadCount), http.StatusOK)
}

func (h *Handlers) RevokeContributorSession(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	if err := h.feedback.RevokeContributorSession(ctx, web.Params(r, "portalSlug"), r.Header.Get("Authorization")); err != nil {
		return web.RespondError(ctx, w, err, httpStatus(err))
	}
	w.Header().Set("Cache-Control", "no-store")
	return web.Respond(ctx, w, nil, http.StatusNoContent)
}

func (h *Handlers) GetItemFollow(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	participant, itemID, err := h.resolveContributorForItem(ctx, r)
	if err != nil {
		return web.RespondError(ctx, w, err, httpStatus(err))
	}
	state, err := h.feedback.GetItemFollow(ctx, web.Params(r, "portalSlug"), itemID, participant)
	if err != nil {
		return web.RespondError(ctx, w, err, httpStatus(err))
	}
	return web.Respond(ctx, w, AppFollowState{ItemID: state.ItemID, Following: state.Following}, http.StatusOK)
}

func (h *Handlers) FollowItem(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	return h.setItemFollow(ctx, w, r, true)
}

func (h *Handlers) UnfollowItem(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	return h.setItemFollow(ctx, w, r, false)
}

func (h *Handlers) setItemFollow(ctx context.Context, w http.ResponseWriter, r *http.Request, following bool) error {
	participant, itemID, err := h.resolveContributorForItem(ctx, r)
	if err != nil {
		return web.RespondError(ctx, w, err, httpStatus(err))
	}
	state, err := h.feedback.FollowItem(ctx, web.Params(r, "portalSlug"), itemID, participant, following)
	if err != nil {
		return web.RespondError(ctx, w, err, httpStatus(err))
	}
	return web.Respond(ctx, w, AppFollowState{ItemID: state.ItemID, Following: state.Following}, http.StatusOK)
}

func (h *Handlers) ExchangePreferenceToken(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	var input AppPreferenceExchange
	if status, err := decodePublicRequest(w, r, &input, feedbackIdentityBodyLimit); err != nil {
		return web.RespondError(ctx, w, err, status)
	}
	result, err := h.feedback.ExchangeUnsubscribeToken(ctx, web.Params(r, "portalSlug"), input.Token)
	if err != nil {
		return web.RespondError(ctx, w, err, httpStatus(err))
	}
	w.Header().Set("Cache-Control", "no-store")
	return web.Respond(ctx, w, toAppSession(result, true, 0), http.StatusOK)
}

func (h *Handlers) GetContributorPreferences(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	result, err := h.feedback.ResolveContributorAuthorization(ctx, web.Params(r, "portalSlug"), r.Header.Get("Authorization"))
	if err != nil {
		return web.RespondError(ctx, w, err, httpStatus(err))
	}
	preferences, err := h.feedback.GetContributorPreferences(ctx, web.Params(r, "portalSlug"), result.Participant)
	if err != nil {
		return web.RespondError(ctx, w, err, httpStatus(err))
	}
	w.Header().Set("Cache-Control", "no-store")
	return web.Respond(ctx, w, toAppPreferences(preferences), http.StatusOK)
}

func (h *Handlers) UpdateContributorPreferences(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	result, err := h.feedback.ResolveContributorAuthorization(ctx, web.Params(r, "portalSlug"), r.Header.Get("Authorization"))
	if err != nil {
		return web.RespondError(ctx, w, err, httpStatus(err))
	}
	var input AppUpdatePreferences
	if status, decodeErr := decodePublicRequest(w, r, &input, feedbackIdentityBodyLimit); decodeErr != nil {
		return web.RespondError(ctx, w, decodeErr, status)
	}
	for _, item := range input.Items {
		if item.ItemID == uuid.Nil {
			return web.RespondError(ctx, w, feedback.ErrInvalidInput, http.StatusBadRequest)
		}
		if _, err := h.feedback.FollowItem(ctx, web.Params(r, "portalSlug"), item.ItemID, result.Participant, item.Following); err != nil {
			return web.RespondError(ctx, w, err, httpStatus(err))
		}
	}
	if input.PortalEmailsEnabled != nil {
		if _, err := h.feedback.SetPortalEmailPreference(ctx, web.Params(r, "portalSlug"), result.Participant, *input.PortalEmailsEnabled); err != nil {
			return web.RespondError(ctx, w, err, httpStatus(err))
		}
	}
	preferences, err := h.feedback.GetContributorPreferences(ctx, web.Params(r, "portalSlug"), result.Participant)
	if err != nil {
		return web.RespondError(ctx, w, err, httpStatus(err))
	}
	w.Header().Set("Cache-Control", "no-store")
	return web.Respond(ctx, w, toAppPreferences(preferences), http.StatusOK)
}

func (h *Handlers) ListPublicUpdates(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	page, pageSize := updatesPagination(r)
	updates, err := h.feedback.ListPublicUpdates(ctx, web.Params(r, "portalSlug"), page, pageSize)
	if err != nil {
		return web.RespondError(ctx, w, err, httpStatus(err))
	}
	unreadCount := 0
	if strings.TrimSpace(r.Header.Get("Authorization")) != "" {
		session, sessionErr := h.feedback.ResolveContributorAuthorization(ctx, web.Params(r, "portalSlug"), r.Header.Get("Authorization"))
		if sessionErr != nil {
			return web.RespondError(ctx, w, sessionErr, httpStatus(sessionErr))
		}
		unreadCount, err = h.feedback.GetUnreadUpdateCount(ctx, web.Params(r, "portalSlug"), session.Participant)
		if err != nil {
			return web.RespondError(ctx, w, err, httpStatus(err))
		}
	}
	return web.Respond(ctx, w, toAppUpdatesPage(updates, unreadCount), http.StatusOK)
}

func (h *Handlers) GetPublicUpdate(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	update, err := h.feedback.GetPublicUpdate(ctx, web.Params(r, "portalSlug"), web.Params(r, "updateSlug"))
	if err != nil {
		return web.RespondError(ctx, w, err, httpStatus(err))
	}
	return web.Respond(ctx, w, toAppUpdate(update), http.StatusOK)
}

func (h *Handlers) MarkUpdatesSeen(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	session, err := h.feedback.ResolveContributorAuthorization(ctx, web.Params(r, "portalSlug"), r.Header.Get("Authorization"))
	if err != nil {
		return web.RespondError(ctx, w, err, httpStatus(err))
	}
	seenAt, err := h.feedback.MarkUpdatesSeen(ctx, web.Params(r, "portalSlug"), session.Participant)
	if err != nil {
		return web.RespondError(ctx, w, err, httpStatus(err))
	}
	w.Header().Set("Cache-Control", "no-store")
	return web.Respond(ctx, w, AppUpdatesSeen{UnreadUpdateCount: 0, LastSeenAt: seenAt}, http.StatusOK)
}

func (h *Handlers) ListWorkspaceUpdates(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	page, pageSize := updatesPagination(r)
	updates, err := h.feedback.ListWorkspaceUpdates(ctx, workspace.ID, page, pageSize)
	if err != nil {
		return web.RespondError(ctx, w, err, httpStatus(err))
	}
	return web.Respond(ctx, w, toAppUpdatesPage(updates, 0), http.StatusOK)
}

func (h *Handlers) GetWorkspaceUpdate(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, updateID, err := workspaceAndUpdateID(ctx, r)
	if err != nil {
		return web.RespondError(ctx, w, err, httpStatus(err))
	}
	update, err := h.feedback.GetWorkspaceUpdate(ctx, workspace, updateID)
	if err != nil {
		return web.RespondError(ctx, w, err, httpStatus(err))
	}
	return web.Respond(ctx, w, toAppUpdate(update), http.StatusOK)
}

func (h *Handlers) CreateUpdate(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	actorID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	var input AppUpdateInput
	if err := web.Decode(r, &input); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	update, err := h.feedback.CreateUpdate(ctx, toCoreUpdateInput(workspace.ID, uuid.Nil, actorID, input))
	if err != nil {
		return web.RespondError(ctx, w, err, httpStatus(err))
	}
	return web.Respond(ctx, w, toAppUpdate(update), http.StatusCreated)
}

func (h *Handlers) UpdateUpdate(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspaceID, updateID, err := workspaceAndUpdateID(ctx, r)
	if err != nil {
		return web.RespondError(ctx, w, err, httpStatus(err))
	}
	actorID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	var input AppUpdateInput
	if err := web.Decode(r, &input); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	update, err := h.feedback.UpdateUpdate(ctx, toCoreUpdateInput(workspaceID, updateID, actorID, input))
	if err != nil {
		return web.RespondError(ctx, w, err, httpStatus(err))
	}
	return web.Respond(ctx, w, toAppUpdate(update), http.StatusOK)
}

func (h *Handlers) DeleteUpdate(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspaceID, updateID, err := workspaceAndUpdateID(ctx, r)
	if err != nil {
		return web.RespondError(ctx, w, err, httpStatus(err))
	}
	if err := h.feedback.DeleteUpdate(ctx, workspaceID, updateID); err != nil {
		return web.RespondError(ctx, w, err, httpStatus(err))
	}
	return web.Respond(ctx, w, nil, http.StatusNoContent)
}

func (h *Handlers) PublishUpdate(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspaceID, updateID, err := workspaceAndUpdateID(ctx, r)
	if err != nil {
		return web.RespondError(ctx, w, err, httpStatus(err))
	}
	actorID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	update, err := h.feedback.PublishUpdate(ctx, workspaceID, updateID, actorID)
	if err != nil {
		return web.RespondError(ctx, w, err, httpStatus(err))
	}
	return web.Respond(ctx, w, toAppUpdate(update), http.StatusOK)
}

func (h *Handlers) UnpublishUpdate(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspaceID, updateID, err := workspaceAndUpdateID(ctx, r)
	if err != nil {
		return web.RespondError(ctx, w, err, httpStatus(err))
	}
	update, err := h.feedback.UnpublishUpdate(ctx, workspaceID, updateID)
	if err != nil {
		return web.RespondError(ctx, w, err, httpStatus(err))
	}
	return web.Respond(ctx, w, toAppUpdate(update), http.StatusOK)
}

func (h *Handlers) GetWidgetSettings(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	portalID, err := uuid.Parse(web.Params(r, "portalId"))
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	settings, err := h.feedback.GetWidgetSettings(ctx, workspace.ID, portalID)
	if err != nil {
		return web.RespondError(ctx, w, err, httpStatus(err))
	}
	return web.Respond(ctx, w, toAppWidgetSettings(settings, ""), http.StatusOK)
}

func (h *Handlers) UpdateWidgetSettings(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	portalID, err := uuid.Parse(web.Params(r, "portalId"))
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	var input AppWidgetSettingsInput
	if err := web.Decode(r, &input); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	settings, err := h.feedback.UpdateWidgetSettings(ctx, feedback.CoreWidgetSettingsInput{
		WorkspaceID: workspace.ID, PortalID: portalID, Enabled: input.Enabled, AllowedOrigins: input.AllowedOrigins,
	})
	if err != nil {
		return web.RespondError(ctx, w, err, httpStatus(err))
	}
	return web.Respond(ctx, w, toAppWidgetSettings(settings, ""), http.StatusOK)
}

func (h *Handlers) CreateWidgetSigningSecret(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	return h.issueWidgetSecret(ctx, w, r, false)
}

func (h *Handlers) RotateWidgetSigningSecret(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	return h.issueWidgetSecret(ctx, w, r, true)
}

func (h *Handlers) issueWidgetSecret(ctx context.Context, w http.ResponseWriter, r *http.Request, rotate bool) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	portalID, err := uuid.Parse(web.Params(r, "portalId"))
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	var result feedback.CoreWidgetSecretResult
	if rotate {
		result, err = h.feedback.RotateWidgetSigningSecret(ctx, workspace.ID, portalID)
	} else {
		result, err = h.feedback.CreateWidgetSigningSecret(ctx, workspace.ID, portalID)
	}
	if err != nil {
		return web.RespondError(ctx, w, err, httpStatus(err))
	}
	w.Header().Set("Cache-Control", "no-store")
	return web.Respond(ctx, w, toAppWidgetSettings(result.Settings, result.SigningSecret), http.StatusOK)
}

func (h *Handlers) GetPublicWidgetConfig(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	settings, err := h.feedback.GetPublicWidgetSettings(ctx, web.Params(r, "portalSlug"))
	if err != nil {
		return web.RespondError(ctx, w, err, httpStatus(err))
	}
	return web.Respond(ctx, w, AppPublicWidgetConfig{Enabled: settings.Enabled, WidgetKeyID: settings.WidgetKeyID, AllowedOrigins: settings.AllowedOrigins}, http.StatusOK)
}

func (h *Handlers) CreateWidgetContributorSession(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	var input AppWidgetSessionRequest
	if status, err := decodePublicRequest(w, r, &input, feedbackIdentityBodyLimit); err != nil {
		return web.RespondError(ctx, w, err, status)
	}
	result, err := h.feedback.CreateWidgetContributorSession(ctx, feedback.CoreWidgetSessionInput{
		PortalSlug: web.Params(r, "portalSlug"), Assertion: input.Assertion, ParentOrigin: input.ParentOrigin,
	})
	if err != nil {
		return web.RespondError(ctx, w, err, httpStatus(err))
	}
	unreadCount, err := h.feedback.GetUnreadUpdateCount(ctx, web.Params(r, "portalSlug"), result.Participant)
	if err != nil {
		return web.RespondError(ctx, w, err, httpStatus(err))
	}
	w.Header().Set("Cache-Control", "no-store")
	return web.Respond(ctx, w, toAppSession(result, true, unreadCount), http.StatusOK)
}

func (h *Handlers) resolveContributorForItem(ctx context.Context, r *http.Request) (feedback.CoreParticipant, uuid.UUID, error) {
	itemID, err := uuid.Parse(web.Params(r, "itemId"))
	if err != nil {
		return feedback.CoreParticipant{}, uuid.Nil, feedback.ErrInvalidInput
	}
	resolved, err := h.resolvePublicParticipant(ctx, r, "")
	if err != nil {
		return feedback.CoreParticipant{}, uuid.Nil, err
	}
	return resolved.Participant, itemID, nil
}

func (h *Handlers) resolvePublicParticipant(ctx context.Context, r *http.Request, expectedKind string) (feedback.CoreResolvedParticipant, error) {
	accountID, _ := mid.GetUserID(ctx)
	return h.feedback.ResolvePublicParticipant(ctx, web.Params(r, "portalSlug"), accountID, r.Header.Get("Authorization"), expectedKind)
}

func updatesPagination(r *http.Request) (int, int) {
	page, pageSize := 1, 20
	if value, err := strconv.Atoi(r.URL.Query().Get("page")); err == nil && value > 0 {
		page = value
	}
	if value, err := strconv.Atoi(r.URL.Query().Get("pageSize")); err == nil && value > 0 {
		pageSize = value
	}
	return page, pageSize
}

func workspaceAndUpdateID(ctx context.Context, r *http.Request) (uuid.UUID, uuid.UUID, error) {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}
	updateID, err := uuid.Parse(web.Params(r, "updateId"))
	if err != nil {
		return uuid.Nil, uuid.Nil, feedback.ErrInvalidInput
	}
	return workspace.ID, updateID, nil
}

func toCoreUpdateInput(workspaceID, updateID, actorID uuid.UUID, input AppUpdateInput) feedback.CoreUpdateInput {
	return feedback.CoreUpdateInput{
		WorkspaceID: workspaceID, PortalID: input.PortalID, UpdateID: updateID, ActorID: actorID,
		Title: input.Title, Slug: input.Slug, Summary: input.Summary, Body: input.Body,
		CoverImageURL: input.CoverImageURL, ItemIDs: input.ItemIDs,
	}
}
