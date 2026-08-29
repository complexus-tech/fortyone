package slackdomain

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

// WorkspaceActorQuery identifies a human read at the persistence boundary.
// ActorID is never inferred from transport context inside the repository.
type WorkspaceActorQuery struct {
	WorkspaceID uuid.UUID
	ActorID     uuid.UUID
}

func (query WorkspaceActorQuery) Validate() error {
	if query.WorkspaceID == uuid.Nil || query.ActorID == uuid.Nil {
		return errors.Join(ErrInvalidInput, errors.New("workspace and actor are required"))
	}
	return nil
}

type UpsertInstallationCommand struct {
	WorkspaceID  uuid.UUID
	ActorID      uuid.UUID
	Installation OAuthInstallation
	Now          time.Time
}

func (command UpsertInstallationCommand) Validate() error {
	if err := (WorkspaceActorQuery{WorkspaceID: command.WorkspaceID, ActorID: command.ActorID}).Validate(); err != nil {
		return err
	}
	if strings.TrimSpace(command.Installation.SlackTeamID) == "" ||
		command.Installation.InstallGeneration == uuid.Nil || command.Now.IsZero() {
		return errors.Join(ErrInvalidInput, errors.New("slack team, installation generation, and time are required"))
	}
	return nil
}

type DisconnectInstallationCommand struct {
	WorkspaceID uuid.UUID
	ActorID     uuid.UUID
	Now         time.Time
}

func (command DisconnectInstallationCommand) Validate() error {
	if err := (WorkspaceActorQuery{WorkspaceID: command.WorkspaceID, ActorID: command.ActorID}).Validate(); err != nil {
		return err
	}
	if command.Now.IsZero() {
		return errors.Join(ErrInvalidInput, errors.New("disconnect time is required"))
	}
	return nil
}

type SyncChannelsCommand struct {
	WorkspaceID            uuid.UUID
	ActorID                uuid.UUID
	InstallationID         uuid.UUID
	InstallationGeneration uuid.UUID
	Channels               []ChannelUpsert
	Now                    time.Time
}

func (command SyncChannelsCommand) Validate() error {
	if err := (WorkspaceActorQuery{WorkspaceID: command.WorkspaceID, ActorID: command.ActorID}).Validate(); err != nil {
		return err
	}
	if command.InstallationID == uuid.Nil || command.InstallationGeneration == uuid.Nil || command.Now.IsZero() {
		return errors.Join(ErrInvalidInput, errors.New("installation, generation, and sync time are required"))
	}
	for _, channel := range command.Channels {
		if strings.TrimSpace(channel.SlackChannelID) == "" || strings.TrimSpace(channel.Name) == "" {
			return errors.Join(ErrInvalidInput, errors.New("slack channel id and name are required"))
		}
	}
	return nil
}

type ListRequestLogsQuery struct {
	WorkspaceID uuid.UUID
	ActorID     uuid.UUID
	Limit       int32
}

func (query ListRequestLogsQuery) Validate() error {
	if err := (WorkspaceActorQuery{WorkspaceID: query.WorkspaceID, ActorID: query.ActorID}).Validate(); err != nil {
		return err
	}
	if query.Limit <= 0 || query.Limit > 200 {
		return errors.Join(ErrInvalidInput, errors.New("request log limit must be between 1 and 200"))
	}
	return nil
}

type UpdateAgentSettingsCommand struct {
	WorkspaceID uuid.UUID
	ActorID     uuid.UUID
	Guidance    string
	Now         time.Time
}

func (command UpdateAgentSettingsCommand) Validate() error {
	if err := (WorkspaceActorQuery{WorkspaceID: command.WorkspaceID, ActorID: command.ActorID}).Validate(); err != nil {
		return err
	}
	if command.Now.IsZero() {
		return errors.Join(ErrInvalidInput, errors.New("settings update time is required"))
	}
	if len([]rune(strings.TrimSpace(command.Guidance))) > MaxAgentGuidanceRunes {
		return errors.Join(ErrInvalidInput, errors.New("slack agent guidance is too long"))
	}
	return nil
}

type ReplaceChannelAudienceCommand struct {
	WorkspaceID            uuid.UUID
	ActorID                uuid.UUID
	InstallationID         uuid.UUID
	InstallationGeneration uuid.UUID
	SlackChannelID         string
	Configured             bool
	TeamIDs                []uuid.UUID
	Now                    time.Time
}

func (command ReplaceChannelAudienceCommand) Validate() error {
	if err := (WorkspaceActorQuery{WorkspaceID: command.WorkspaceID, ActorID: command.ActorID}).Validate(); err != nil {
		return err
	}
	if command.InstallationID == uuid.Nil || command.InstallationGeneration == uuid.Nil ||
		strings.TrimSpace(command.SlackChannelID) == "" || command.Now.IsZero() {
		return errors.Join(ErrInvalidInput, errors.New("installation, generation, Slack channel, and time are required"))
	}
	for _, teamID := range command.TeamIDs {
		if teamID == uuid.Nil {
			return errors.Join(ErrInvalidInput, errors.New("channel audience contains an empty team id"))
		}
	}
	return nil
}
