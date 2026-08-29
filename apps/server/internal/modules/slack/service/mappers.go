package slack

import (
	"encoding/json"
)

func toCoreSlackWorkspace(record slackWorkspaceRecord) CoreSlackWorkspace {
	return CoreSlackWorkspace{
		ID:                record.ID,
		SlackTeamID:       record.SlackTeamID,
		SlackTeamName:     record.SlackTeamName,
		SlackTeamDomain:   record.SlackTeamDomain,
		BotUserID:         record.BotUserID,
		Scope:             record.Scope,
		IsActive:          record.IsActive,
		InstalledByUserID: record.InstalledByUserID,
		CreatedAt:         record.CreatedAt,
		UpdatedAt:         record.UpdatedAt,
	}
}

func toCoreChannels(records []slackChannelRecord) []CoreSlackChannel {
	channels := make([]CoreSlackChannel, 0, len(records))
	for _, record := range records {
		channels = append(channels, CoreSlackChannel{
			ID:             record.ID,
			SlackChannelID: record.SlackChannelID,
			Name:           record.Name,
			IsPrivate:      record.IsPrivate,
			IsArchived:     record.IsArchived,
			IsMember:       record.IsMember,
			IsActive:       record.IsActive,
			LastSyncedAt:   record.LastSyncedAt,
			CreatedAt:      record.CreatedAt,
			UpdatedAt:      record.UpdatedAt,
		})
	}
	return channels
}

func toCoreRequestLog(record slackRequestLogRecord) CoreRequestLog {
	headers := map[string]string{}
	if len(record.Headers) > 0 {
		_ = json.Unmarshal(record.Headers, &headers)
	}
	return CoreRequestLog{
		ID:           record.ID,
		RequestType:  record.RequestType,
		Endpoint:     record.Endpoint,
		WorkspaceID:  record.WorkspaceID,
		SlackTeamID:  record.SlackTeamID,
		SlackUserID:  record.SlackUserID,
		SlackChannel: record.SlackChannel,
		Command:      record.Command,
		TriggerID:    record.TriggerID,
		RequestBody:  record.RequestBody,
		Headers:      headers,
		ResponseCode: record.ResponseCode,
		Outcome:      record.Outcome,
		ErrorMessage: record.ErrorMessage,
		CreatedAt:    record.CreatedAt,
	}
}
