package slackhttp

import (
	slack "github.com/complexus-tech/projects-api/internal/modules/slack/service"
)

type AppWorkspaceSettings struct {
	DefaultCreateMode string `json:"defaultCreateMode"`
	CreatedAt         string `json:"createdAt"`
	UpdatedAt         string `json:"updatedAt"`
}

type AppSlackWorkspace struct {
	ID                string  `json:"id"`
	SlackTeamID       string  `json:"slackTeamId"`
	SlackTeamName     string  `json:"slackTeamName"`
	SlackTeamDomain   string  `json:"slackTeamDomain"`
	BotUserID         *string `json:"botUserId,omitempty"`
	Scope             *string `json:"scope,omitempty"`
	IsActive          bool    `json:"isActive"`
	InstalledByUserID *string `json:"installedByUserId,omitempty"`
	CreatedAt         string  `json:"createdAt"`
	UpdatedAt         string  `json:"updatedAt"`
}

type AppSlackChannel struct {
	ID             string  `json:"id"`
	SlackChannelID string  `json:"slackChannelId"`
	Name           string  `json:"name"`
	IsPrivate      bool    `json:"isPrivate"`
	IsArchived     bool    `json:"isArchived"`
	IsMember       bool    `json:"isMember"`
	IsActive       bool    `json:"isActive"`
	LastSyncedAt   *string `json:"lastSyncedAt,omitempty"`
	CreatedAt      string  `json:"createdAt"`
	UpdatedAt      string  `json:"updatedAt"`
}

type AppSlackChannelLink struct {
	ID             string `json:"id"`
	SlackChannelID string `json:"slackChannelId"`
	TeamID         string `json:"teamId"`
	TeamCode       string `json:"teamCode"`
	TeamName       string `json:"teamName"`
	TeamColor      string `json:"teamColor"`
	IsActive       bool   `json:"isActive"`
	CreatedAt      string `json:"createdAt"`
	UpdatedAt      string `json:"updatedAt"`
}

type AppIntegration struct {
	Settings       AppWorkspaceSettings  `json:"settings"`
	SlackWorkspace *AppSlackWorkspace    `json:"slackWorkspace,omitempty"`
	Channels       []AppSlackChannel     `json:"channels"`
	ChannelLinks   []AppSlackChannelLink `json:"channelLinks"`
}

type AppCreateInstallSession struct {
	InstallURL string `json:"installUrl"`
}

type AppUpdateWorkspaceSettingsRequest struct {
	DefaultCreateMode *string `json:"defaultCreateMode"`
}

type AppCreateChannelLinkRequest struct {
	SlackChannelID string `json:"slackChannelId"`
	TeamID         string `json:"teamId"`
}

func toAppIntegration(input slack.CoreIntegration) AppIntegration {
	out := AppIntegration{
		Settings:     toAppWorkspaceSettings(input.Settings),
		Channels:     make([]AppSlackChannel, 0, len(input.Channels)),
		ChannelLinks: make([]AppSlackChannelLink, 0, len(input.ChannelLinks)),
	}
	if input.SlackWorkspace != nil {
		workspace := toAppSlackWorkspace(*input.SlackWorkspace)
		out.SlackWorkspace = &workspace
	}
	for _, channel := range input.Channels {
		out.Channels = append(out.Channels, toAppChannel(channel))
	}
	for _, link := range input.ChannelLinks {
		out.ChannelLinks = append(out.ChannelLinks, toAppChannelLink(link))
	}
	return out
}

func toAppWorkspaceSettings(input slack.CoreWorkspaceSettings) AppWorkspaceSettings {
	return AppWorkspaceSettings{
		DefaultCreateMode: input.DefaultCreateMode,
		CreatedAt:         input.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:         input.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func toAppSlackWorkspace(input slack.CoreSlackWorkspace) AppSlackWorkspace {
	var installedBy *string
	if input.InstalledByUserID != nil {
		value := input.InstalledByUserID.String()
		installedBy = &value
	}
	return AppSlackWorkspace{
		ID:                input.ID.String(),
		SlackTeamID:       input.SlackTeamID,
		SlackTeamName:     input.SlackTeamName,
		SlackTeamDomain:   input.SlackTeamDomain,
		BotUserID:         input.BotUserID,
		Scope:             input.Scope,
		IsActive:          input.IsActive,
		InstalledByUserID: installedBy,
		CreatedAt:         input.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:         input.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func toAppChannel(input slack.CoreSlackChannel) AppSlackChannel {
	var lastSyncedAt *string
	if input.LastSyncedAt != nil {
		value := input.LastSyncedAt.Format("2006-01-02T15:04:05Z07:00")
		lastSyncedAt = &value
	}
	return AppSlackChannel{
		ID:             input.ID.String(),
		SlackChannelID: input.SlackChannelID,
		Name:           input.Name,
		IsPrivate:      input.IsPrivate,
		IsArchived:     input.IsArchived,
		IsMember:       input.IsMember,
		IsActive:       input.IsActive,
		LastSyncedAt:   lastSyncedAt,
		CreatedAt:      input.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:      input.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func toAppChannelLink(input slack.CoreSlackChannelLink) AppSlackChannelLink {
	return AppSlackChannelLink{
		ID:             input.ID.String(),
		SlackChannelID: input.SlackChannelID,
		TeamID:         input.TeamID.String(),
		TeamCode:       input.TeamCode,
		TeamName:       input.TeamName,
		TeamColor:      input.TeamColor,
		IsActive:       input.IsActive,
		CreatedAt:      input.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:      input.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}
