package slackhttp

import (
	slack "github.com/complexus-tech/projects-api/internal/modules/slack/service"
)

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

type AppIntegration struct {
	SlackWorkspace *AppSlackWorkspace `json:"slackWorkspace,omitempty"`
	Channels       []AppSlackChannel  `json:"channels"`
}

type AppCreateInstallSession struct {
	InstallURL string `json:"installUrl"`
}

type AppLinkSlackAccountRequest struct {
	Token string `json:"token" validate:"required"`
}

type AppRequestLog struct {
	ID           string            `json:"id"`
	RequestType  string            `json:"requestType"`
	Endpoint     string            `json:"endpoint"`
	WorkspaceID  *string           `json:"workspaceId,omitempty"`
	SlackTeamID  *string           `json:"slackTeamId,omitempty"`
	SlackUserID  *string           `json:"slackUserId,omitempty"`
	SlackChannel *string           `json:"slackChannelId,omitempty"`
	Command      *string           `json:"command,omitempty"`
	TriggerID    *string           `json:"triggerId,omitempty"`
	RequestBody  *string           `json:"requestBody,omitempty"`
	Headers      map[string]string `json:"headers"`
	ResponseCode int               `json:"responseCode"`
	Outcome      string            `json:"outcome"`
	ErrorMessage *string           `json:"errorMessage,omitempty"`
	CreatedAt    string            `json:"createdAt"`
}

type AppRuntimeActor struct {
	SlackTeamID    string `json:"teamId"`
	SlackUserID    string `json:"userId"`
	SlackUserName  string `json:"userName"`
	SlackChannelID string `json:"channelId"`
	SlackChannel   string `json:"channelName"`
	SlackMessageTS string `json:"messageTs"`
	SlackThreadTS  string `json:"threadTs"`
}

type AppRuntimeOptionsRequest struct {
	Actor  AppRuntimeActor `json:"actor"`
	Query  string          `json:"query"`
	TeamID string          `json:"teamId"`
}

type AppRuntimeCreateStoryRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	TeamID      string `json:"teamId"`
	StatusID    string `json:"statusId"`
	Priority    string `json:"priority"`
	AssigneeID  string `json:"assigneeId"`
	ObjectiveID string `json:"objectiveId"`
	Source      struct {
		SlackTeamID    string `json:"teamId"`
		SlackUserID    string `json:"userId"`
		SlackUserName  string `json:"userName"`
		SlackChannelID string `json:"channelId"`
		SlackChannel   string `json:"channelName"`
		SlackMessageTS string `json:"messageTs"`
		SlackThreadTS  string `json:"threadTs"`
		SlackText      string `json:"messageText"`
	} `json:"source"`
}

type AppRuntimeThreadCommentRequest struct {
	Actor       AppRuntimeActor `json:"actor"`
	MessageText string          `json:"messageText"`
	StoryID     string          `json:"storyId"`
}

type AppRuntimeStoryUnfurlRequest struct {
	Actor AppRuntimeActor `json:"actor"`
	URL   string          `json:"url"`
}

func toCoreRuntimeActor(input AppRuntimeActor) slack.CoreRuntimeActor {
	return slack.CoreRuntimeActor{
		SlackTeamID:    input.SlackTeamID,
		SlackUserID:    input.SlackUserID,
		SlackUserName:  input.SlackUserName,
		SlackChannelID: input.SlackChannelID,
		SlackChannel:   input.SlackChannel,
		SlackMessageTS: input.SlackMessageTS,
		SlackThreadTS:  input.SlackThreadTS,
	}
}

func toCoreRuntimeCreateStoryInput(input AppRuntimeCreateStoryRequest) slack.CoreRuntimeCreateStoryInput {
	return slack.CoreRuntimeCreateStoryInput{
		Title:       input.Title,
		Description: input.Description,
		TeamID:      input.TeamID,
		StatusID:    input.StatusID,
		Priority:    input.Priority,
		AssigneeID:  input.AssigneeID,
		ObjectiveID: input.ObjectiveID,
		Source:      input.Source,
	}
}

func toAppIntegration(input slack.CoreIntegration) AppIntegration {
	out := AppIntegration{
		Channels: make([]AppSlackChannel, 0, len(input.Channels)),
	}
	if input.SlackWorkspace != nil {
		workspace := toAppSlackWorkspace(*input.SlackWorkspace)
		out.SlackWorkspace = &workspace
	}
	for _, channel := range input.Channels {
		out.Channels = append(out.Channels, toAppChannel(channel))
	}
	return out
}

func toAppRequestLogs(input []slack.CoreRequestLog) []AppRequestLog {
	out := make([]AppRequestLog, 0, len(input))
	for _, item := range input {
		out = append(out, toAppRequestLog(item))
	}
	return out
}

func toAppRequestLog(input slack.CoreRequestLog) AppRequestLog {
	var workspaceID *string
	if input.WorkspaceID != nil {
		value := input.WorkspaceID.String()
		workspaceID = &value
	}
	return AppRequestLog{
		ID:           input.ID.String(),
		RequestType:  input.RequestType,
		Endpoint:     input.Endpoint,
		WorkspaceID:  workspaceID,
		SlackTeamID:  input.SlackTeamID,
		SlackUserID:  input.SlackUserID,
		SlackChannel: input.SlackChannel,
		Command:      input.Command,
		TriggerID:    input.TriggerID,
		RequestBody:  input.RequestBody,
		Headers:      input.Headers,
		ResponseCode: input.ResponseCode,
		Outcome:      input.Outcome,
		ErrorMessage: input.ErrorMessage,
		CreatedAt:    input.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
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
