package slack

import (
	"context"

	"github.com/google/uuid"
)

// SlackFileImportIntent contains only identifiers needed to import a confirmed
// Slack-hosted file. The worker resolves the current bot credential and file
// metadata when it runs.
type SlackFileImportIntent struct {
	IdempotencyKey    string
	WorkspaceID       uuid.UUID
	ActorID           uuid.UUID
	InstallationID    uuid.UUID
	InstallGeneration uuid.UUID
	StoryID           uuid.UUID
	SlackTeamID       string
	SlackUserID       string
	ChannelID         string
	ThreadTS          string
	MessageTS         string
	FileID            string
}

type SlackFileImportQueue interface {
	QueueSlackFileImport(context.Context, SlackFileImportIntent) error
}
