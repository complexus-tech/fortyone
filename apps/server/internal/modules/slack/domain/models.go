package slackdomain

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	MaxAgentGuidanceRunes = 4000
	UninstallMaxAttempts  = 8
)

type Workspace struct {
	ID   uuid.UUID
	Slug string
	Name string
}

type Team struct {
	ID    uuid.UUID
	Code  string
	Name  string
	Color string
}

type Status struct {
	ID       uuid.UUID
	Name     string
	Category string
}

type TeamMember struct {
	UserID   uuid.UUID
	Username string
	FullName string
	Email    string
}

type Label struct {
	ID   uuid.UUID
	Name string
}

type Objective struct {
	ID   uuid.UUID
	Name string
}

type WorkspaceMember struct {
	UserID uuid.UUID
	Email  string
}

type UserLinkUpsert struct {
	SlackUserID string
	UserID      uuid.UUID
	LinkedVia   string
}

type UserLink struct {
	SlackUserID string
	UserID      uuid.UUID
	LinkedVia   string
	LinkedAt    time.Time
}

type Installation struct {
	ID              uuid.UUID
	WorkspaceID     uuid.UUID
	SlackTeamID     string
	SlackTeamName   string
	SlackTeamDomain string
	BotUserID       *string
	// BotAccessToken contains a shared-vault envelope. The historic name is
	// retained at this compatibility boundary while plaintext writes are
	// rejected by the service and repository.
	BotAccessToken    string
	CredentialVersion int
	InstallGeneration uuid.UUID
	AuthorizedAt      time.Time
	SlackAppID        *string
	EnterpriseID      *string
	AuthedUserID      *string
	Scope             *string
	IsActive          bool
	InstalledByUserID *uuid.UUID
	RevokedAt         *time.Time
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type Channel struct {
	ID                    uuid.UUID
	WorkspaceID           uuid.UUID
	SlackWorkspaceID      uuid.UUID
	SlackChannelID        string
	Name                  string
	IsPrivate             bool
	IsArchived            bool
	IsMember              bool
	IsActive              bool
	IsAssistantConfigured bool
	LastSyncedAt          *time.Time
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

type OAuthInstallation struct {
	SlackTeamID       string
	SlackTeamName     string
	SlackTeamDomain   string
	BotUserID         *string
	BotAccessToken    string
	CredentialVersion int
	InstallGeneration uuid.UUID
	SlackAppID        *string
	EnterpriseID      *string
	AuthedUserID      *string
	Scope             *string
}

type UninstallKind string

const (
	UninstallDisconnect      UninstallKind = "disconnect"
	UninstallWorkspaceDelete UninstallKind = "workspace_delete"
	UninstallOrphanedOAuth   UninstallKind = "orphaned_oauth"
)

type UninstallStatus string

const (
	UninstallPending            UninstallStatus = "pending"
	UninstallProcessing         UninstallStatus = "processing"
	UninstallCompleted          UninstallStatus = "completed"
	UninstallFailed             UninstallStatus = "failed"
	UninstallRevocationRequired UninstallStatus = "revocation_required"
)

func ParseUninstallKind(value string) (UninstallKind, error) {
	kind := UninstallKind(strings.TrimSpace(value))
	switch kind {
	case UninstallDisconnect, UninstallWorkspaceDelete, UninstallOrphanedOAuth:
		return kind, nil
	default:
		return "", fmt.Errorf("%w: unsupported uninstall kind %q", ErrInvalidInput, value)
	}
}

func ParseUninstallStatus(value string) (UninstallStatus, error) {
	status := UninstallStatus(strings.TrimSpace(value))
	switch status {
	case UninstallPending,
		UninstallProcessing,
		UninstallCompleted,
		UninstallFailed,
		UninstallRevocationRequired:
		return status, nil
	default:
		return "", fmt.Errorf("%w: unsupported uninstall status %q", ErrInvalidInput, value)
	}
}

type Uninstall struct {
	ID                   uuid.UUID
	SlackWorkspaceID     uuid.UUID
	WorkspaceID          uuid.UUID
	InstallGeneration    uuid.UUID
	SlackTeamID          string
	UninstallKind        UninstallKind
	CredentialPayload    string
	CredentialKeyVersion int
	Status               UninstallStatus
	AttemptCount         int
	LastError            *string
	NextAttemptAt        *time.Time
	ProcessingStartedAt  *time.Time
	CompletedAt          *time.Time
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

type EnqueueUninstall struct {
	SlackWorkspaceID     uuid.UUID
	WorkspaceID          uuid.UUID
	InstallGeneration    uuid.UUID
	SlackTeamID          string
	UninstallKind        UninstallKind
	CredentialPayload    string
	CredentialKeyVersion int
}

func (input EnqueueUninstall) Validate() error {
	if input.SlackWorkspaceID == uuid.Nil {
		return errors.Join(ErrInvalidInput, errors.New("slack installation id is required"))
	}
	if input.WorkspaceID == uuid.Nil {
		return errors.Join(ErrInvalidInput, errors.New("workspace id is required"))
	}
	if input.InstallGeneration == uuid.Nil {
		return errors.Join(ErrInvalidInput, errors.New("installation generation is required"))
	}
	if strings.TrimSpace(input.SlackTeamID) == "" {
		return errors.Join(ErrInvalidInput, errors.New("slack team id is required"))
	}
	if _, err := ParseUninstallKind(string(input.UninstallKind)); err != nil {
		return err
	}
	if strings.TrimSpace(input.CredentialPayload) == "" || input.CredentialKeyVersion <= 0 {
		return errors.Join(ErrInvalidInput, errors.New("versioned credential payload is required"))
	}
	return nil
}

type ChannelUpsert struct {
	SlackChannelID string
	Name           string
	IsPrivate      bool
	IsArchived     bool
	IsMember       bool
}

type RequestLogInsert struct {
	RequestType  string
	Endpoint     string
	WorkspaceID  *uuid.UUID
	SlackTeamID  *string
	SlackUserID  *string
	SlackChannel *string
	Command      *string
	TriggerID    *string
	RequestBody  *string
	Headers      []byte
	ResponseCode int
	Outcome      string
	ErrorMessage *string
}

type RequestLog struct {
	ID           uuid.UUID
	RequestType  string
	Endpoint     string
	WorkspaceID  *uuid.UUID
	SlackTeamID  *string
	SlackUserID  *string
	SlackChannel *string
	Command      *string
	TriggerID    *string
	RequestBody  *string
	Headers      []byte
	ResponseCode int
	Outcome      string
	ErrorMessage *string
	CreatedAt    time.Time
}

type AgentSettings struct {
	WorkspaceID uuid.UUID
	Guidance    string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type UpdateAgentSettings struct {
	Guidance string
}

type ChannelTeamAccess struct {
	SlackChannelID string
	TeamID         uuid.UUID
}

type AssistantChannelTeamScope struct {
	AllowedTeamIDs []uuid.UUID
	SharedTeamIDs  []uuid.UUID
}

type LegacyInstallationCredential struct {
	SlackWorkspaceID  uuid.UUID
	WorkspaceID       uuid.UUID
	SlackTeamID       string
	InstallGeneration uuid.UUID
	Credential        string
	CredentialVersion int
}

type LegacyUninstallCredential struct {
	UninstallID       uuid.UUID
	WorkspaceID       uuid.UUID
	SlackTeamID       string
	InstallGeneration uuid.UUID
	Credential        string
	CredentialVersion int
}

type InstallationCredentialForRewrap struct {
	SlackWorkspaceID  uuid.UUID
	WorkspaceID       uuid.UUID
	SlackTeamID       string
	InstallGeneration uuid.UUID
	Credential        string
	CredentialVersion int
}

type UninstallCredentialForRewrap struct {
	UninstallID       uuid.UUID
	WorkspaceID       uuid.UUID
	SlackTeamID       string
	InstallGeneration uuid.UUID
	Credential        string
	CredentialVersion int
}
