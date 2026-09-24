package slackdomain

import "github.com/google/uuid"

// FileImport contains identifiers only. Credentials and Slack download URLs
// are resolved from the current installation when the job is claimed.
type FileImport struct {
	ID                uuid.UUID
	WorkspaceID       uuid.UUID
	SlackWorkspaceID  uuid.UUID
	InstallGeneration uuid.UUID
	SlackTeamID       string
	SlackUserID       string
	ChannelID         string
	ThreadTS          string
	MessageTS         string
	FileID            string
	StoryID           uuid.UUID
	ActorID           uuid.UUID
	AttemptCount      int32
}

type RegisterFileImport struct {
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
