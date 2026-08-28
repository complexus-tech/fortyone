package usershttp

import (
	"context"
	"errors"
	"math"
	"net/http"
	"net/url"

	users "github.com/complexus-tech/projects-api/internal/modules/users/service"
	"github.com/complexus-tech/projects-api/internal/platform/actors"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/internal/platform/pagination"
	"github.com/complexus-tech/projects-api/internal/platform/workschedule"
	"github.com/complexus-tech/projects-api/pkg/web"
)

type userListQuery struct {
	Filter users.CoreListUsersFilter
	Page   *pagination.OffsetParams
}

func parseUserListQuery(values url.Values) (userListQuery, error) {
	teamID, err := web.OptionalUUIDQueryParameter(values, "teamId")
	if err != nil {
		return userListQuery{}, err
	}
	search, _, err := web.OptionalTextQueryParameter(
		values, "search", web.DefaultMaxQueryParameterBytes, web.DefaultMaxQueryParameterBytes,
	)
	if err != nil {
		return userListQuery{}, err
	}
	query := userListQuery{Filter: users.CoreListUsersFilter{
		TeamID: teamID,
		Search: search,
	}}
	if !pagination.OffsetRequested(values) {
		return query, nil
	}
	params, err := pagination.ParseOffsetQuery(values, pagination.OffsetQueryConfig{
		DefaultPageSize: pagination.DefaultMenuPageSize,
		MaximumPageSize: pagination.MaximumPageSize,
		MaximumOffset:   math.MaxInt32,
	})
	if err != nil {
		return userListQuery{}, err
	}
	query.Page = &params
	return query, nil
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
	if req.WorkSchedule != nil {
		updates.WorkSchedule = &users.CoreWorkScheduleOverride{
			WorkingDays:        append([]int(nil), req.WorkSchedule.WorkingDays...),
			WorkingStartMinute: req.WorkSchedule.WorkingStartMinute,
			WorkingEndMinute:   req.WorkSchedule.WorkingEndMinute,
		}
	}

	if err := h.users.UpdateUser(ctx, userID, updates); err != nil {
		if errors.Is(err, users.ErrNotFound) {
			return web.RespondError(ctx, w, err, http.StatusNotFound)
		}
		if errors.Is(err, workschedule.ErrInvalidWorkingDays) || errors.Is(err, workschedule.ErrInvalidWorkingHours) {
			return web.RespondError(ctx, w, err, http.StatusBadRequest)
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

	query, err := parseUserListQuery(r.URL.Query())
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	filter := query.Filter
	if query.Page != nil {
		params := query.Page
		page, pageSize := params.Page, params.PageSize
		filter.Limit = pageSize + 1
		filter.Offset = params.Offset()

		users, err := h.users.List(ctx, workspace.ID, filter)
		if err != nil {
			return web.RespondError(ctx, w, err, http.StatusBadRequest)
		}

		hasMore := len(users) > pageSize
		if hasMore {
			users = users[:pageSize]
		}

		h.resolveUserAvatars(ctx, users)
		web.Respond(ctx, w, toAppMembersResponse(users, page, pageSize, hasMore), http.StatusOK)
		return nil
	}

	users, err := h.users.List(ctx, workspace.ID, filter)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	h.resolveUserAvatars(ctx, users)
	web.Respond(ctx, w, toAppUsers(users), http.StatusOK)
	return nil
}

func (h *Handlers) GetMayaAssignee(ctx context.Context, w http.ResponseWriter, _ *http.Request) error {
	email, ok := actors.EmailForKey(actors.KeySystem)
	if !ok {
		return web.RespondError(ctx, w, errors.New("maya actor is not configured"), http.StatusInternalServerError)
	}

	user, err := h.users.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, users.ErrNotFound) {
			return web.RespondError(ctx, w, err, http.StatusNotFound)
		}
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}
	if !user.IsSystem {
		return web.RespondError(ctx, w, errors.New("maya actor is not a system user"), http.StatusInternalServerError)
	}

	h.resolveUserAvatar(ctx, &user)
	return web.Respond(ctx, w, toAppUser(user), http.StatusOK)
}
