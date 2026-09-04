package usershttp

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	users "github.com/complexus-tech/projects-api/internal/modules/users/service"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/web"
)

type onboardingTourProgressQuery struct {
	TourKey     string
	TourVersion string
}

func parseOnboardingTourProgressQuery(values url.Values) (onboardingTourProgressQuery, error) {
	tourKey, err := web.RequiredTextQueryParameter(
		values,
		"tourKey",
		users.MaximumOnboardingTourKeyRunes*4,
		users.MaximumOnboardingTourKeyRunes,
	)
	if err != nil {
		return onboardingTourProgressQuery{}, err
	}
	tourVersion, err := web.RequiredTextQueryParameter(
		values,
		"tourVersion",
		users.MaximumOnboardingTourVersionRunes*4,
		users.MaximumOnboardingTourVersionRunes,
	)
	if err != nil {
		return onboardingTourProgressQuery{}, err
	}
	return onboardingTourProgressQuery{TourKey: tourKey, TourVersion: tourVersion}, nil
}

func (h *Handlers) GetOnboardingTourProgress(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "handlers.users.GetOnboardingTourProgress")
	defer span.End()

	query, err := parseOnboardingTourProgressQuery(r.URL.Query())
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	progress, err := h.users.GetOnboardingTourProgress(ctx, userID, users.CoreOnboardingTourScope{
		TourKey:     query.TourKey,
		TourVersion: query.TourVersion,
	})
	if err != nil {
		if errors.Is(err, users.ErrNotFound) {
			return web.RespondError(ctx, w, err, http.StatusNotFound)
		}
		if errors.Is(err, users.ErrInvalidOnboardingTour) {
			return web.RespondError(ctx, w, err, http.StatusBadRequest)
		}
		return web.RespondError(ctx, w, fmt.Errorf("get onboarding tour progress: %w", err), http.StatusInternalServerError)
	}

	span.AddEvent("onboarding tour progress retrieved")
	return web.Respond(ctx, w, toAppOnboardingTourProgress(progress), http.StatusOK)
}

func (h *Handlers) UpdateOnboardingTourProgress(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "handlers.users.UpdateOnboardingTourProgress")
	defer span.End()

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	var req UpdateOnboardingTourProgressRequest
	if err := web.Decode(r, &req); err != nil {
		return web.RespondError(ctx, w, fmt.Errorf("invalid request: %w", err), http.StatusBadRequest)
	}

	progress, err := h.users.UpdateOnboardingTourProgress(
		ctx,
		userID,
		toCoreUpdateOnboardingTourProgress(req),
	)
	if err != nil {
		if errors.Is(err, users.ErrNotFound) {
			return web.RespondError(ctx, w, err, http.StatusNotFound)
		}
		if errors.Is(err, users.ErrInvalidOnboardingTour) {
			return web.RespondError(ctx, w, err, http.StatusBadRequest)
		}
		return web.RespondError(ctx, w, fmt.Errorf("update onboarding tour progress: %w", err), http.StatusInternalServerError)
	}

	span.AddEvent("onboarding tour progress updated")
	return web.Respond(ctx, w, toAppOnboardingTourProgress(progress), http.StatusOK)
}
