package slackhttp

import (
	"context"
	"errors"
	"net/http"
	"strings"

	slack "github.com/complexus-tech/projects-api/internal/modules/slack/service"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

type AppSlackChannelAudience struct {
	Channel      AppSlackChannel `json:"channel"`
	IsConfigured bool            `json:"isConfigured"`
	TeamIDs      []string        `json:"teamIds"`
}

type AppUpdateSlackChannelAudience struct {
	// IsConfigured is optional for compatibility with clients deployed before
	// explicit channel selection. An omitted value retains the previous PUT
	// behavior and configures the channel.
	IsConfigured *bool    `json:"isConfigured"`
	TeamIDs      []string `json:"teamIds"`
}

func (h *Handlers) ListChannelAudiences(ctx context.Context, w http.ResponseWriter, _ *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	audiences, err := h.service.ListChannelAudiences(ctx, workspace.ID)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}
	return web.Respond(ctx, w, toAppChannelAudiences(audiences), http.StatusOK)
}

func (h *Handlers) UpdateChannelAudience(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	actorID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	channelID := strings.TrimSpace(web.Params(r, "channelId"))
	if channelID == "" {
		return web.RespondError(ctx, w, errors.New("Slack channel is required"), http.StatusBadRequest)
	}
	var input AppUpdateSlackChannelAudience
	if err := web.Decode(r, &input); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	teamIDs := make([]uuid.UUID, 0, len(input.TeamIDs))
	for _, rawTeamID := range input.TeamIDs {
		teamID, parseErr := uuid.Parse(strings.TrimSpace(rawTeamID))
		if parseErr != nil {
			return web.RespondError(ctx, w, parseErr, http.StatusBadRequest)
		}
		teamIDs = append(teamIDs, teamID)
	}
	if err := h.service.UpdateChannelAudience(
		ctx,
		workspace.ID,
		actorID,
		channelID,
		assistantConfiguredOrDefault(input.IsConfigured),
		teamIDs,
	); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	return web.Respond(ctx, w, nil, http.StatusNoContent)
}

func assistantConfiguredOrDefault(value *bool) bool {
	if value == nil {
		return true
	}
	return *value
}

func toAppChannelAudiences(input []slack.CoreSlackChannelAudience) []AppSlackChannelAudience {
	output := make([]AppSlackChannelAudience, 0, len(input))
	for _, audience := range input {
		teamIDs := make([]string, 0, len(audience.TeamIDs))
		for _, teamID := range audience.TeamIDs {
			teamIDs = append(teamIDs, teamID.String())
		}
		output = append(output, AppSlackChannelAudience{
			Channel:      toAppChannel(audience.Channel),
			IsConfigured: audience.IsConfigured,
			TeamIDs:      teamIDs,
		})
	}
	return output
}
