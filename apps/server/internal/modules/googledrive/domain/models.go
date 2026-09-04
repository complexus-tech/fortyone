package domain

import (
	"time"

	"github.com/google/uuid"
)

const (
	TargetStory     TargetType = "story"
	TargetObjective TargetType = "objective"
	TargetDocument  TargetType = "document"
	TargetComment   TargetType = "comment"

	FileTypeDocument    FileType = "document"
	FileTypeSpreadsheet FileType = "spreadsheet"
)

type TargetType string
type FileType string

func (target TargetType) Valid() bool {
	switch target {
	case TargetStory, TargetObjective, TargetDocument, TargetComment:
		return true
	default:
		return false
	}
}

func (fileType FileType) Valid() bool {
	return fileType == FileTypeDocument || fileType == FileTypeSpreadsheet
}

type OAuthToken struct {
	AccessToken  string    `json:"accessToken"`
	RefreshToken string    `json:"refreshToken"`
	TokenType    string    `json:"tokenType"`
	Expiry       time.Time `json:"expiry"`
}

type OAuthState struct {
	StateHash     string
	WorkspaceID   uuid.UUID
	UserID        uuid.UUID
	WorkspaceSlug string
	ReturnURL     *string
	CodeVerifier  string
	ExpiresAt     time.Time
}

type Account struct {
	ID                      uuid.UUID
	UserID                  uuid.UUID
	GoogleSubject           string
	Email                   string
	DisplayName             *string
	CredentialPayload       string
	CredentialVersion       int16
	InstallationGeneration  uuid.UUID
	Scopes                  []string
	ExpiresAt               time.Time
	RequiresReauthorization bool
	CreatedAt               time.Time
	UpdatedAt               time.Time
}

type Connection struct {
	WorkspaceID uuid.UUID
	Account
}

type RevocationCandidate struct {
	ID            uuid.UUID
	UserID        uuid.UUID
	GoogleSubject string
}

type Revocation struct {
	ID                     uuid.UUID
	SourceAccountID        *uuid.UUID
	UserID                 uuid.UUID
	GoogleSubject          string
	InstallationGeneration uuid.UUID
	CredentialPayload      string
	CredentialVersion      int16
	AttemptCount           int
	ClaimToken             uuid.UUID
	LeaseExpiresAt         time.Time
	CreatedAt              time.Time
	UpdatedAt              time.Time
}

func (revocation Revocation) Account() Account {
	return Account{
		ID:                     pointerUUIDValue(revocation.SourceAccountID),
		UserID:                 revocation.UserID,
		GoogleSubject:          revocation.GoogleSubject,
		CredentialPayload:      revocation.CredentialPayload,
		CredentialVersion:      revocation.CredentialVersion,
		InstallationGeneration: revocation.InstallationGeneration,
	}
}

func pointerUUIDValue(value *uuid.UUID) uuid.UUID {
	if value == nil {
		return uuid.Nil
	}
	return *value
}

type ProviderFile struct {
	ID          string
	ResourceKey *string
	Name        string
	MimeType    string
	WebViewLink string
	// ThumbnailLink is short-lived provider metadata and must never be persisted
	// or returned to clients directly. It is fetched again before every preview.
	ThumbnailLink string `json:"-"`
	DriveID       *string
	Version       *string
	SizeBytes     *int64
	ModifiedAt    *time.Time
	Trashed       bool
	Metadata      []byte
}

type FileReference struct {
	ID              uuid.UUID  `json:"id"`
	FileID          string     `json:"-"`
	Name            string     `json:"name"`
	MimeType        string     `json:"mimeType"`
	WebViewLink     string     `json:"webViewLink"`
	ResourceKey     *string    `json:"-"`
	ModifiedTime    *time.Time `json:"modifiedTime,omitempty"`
	ConnectionEmail *string    `json:"connectionEmail,omitempty"`
	Availability    string     `json:"availability"`
	TargetType      TargetType `json:"targetType"`
	TargetID        uuid.UUID  `json:"targetId"`
	CreatedAt       time.Time  `json:"createdAt"`
	UpdatedAt       time.Time  `json:"updatedAt"`

	InternalFileID  uuid.UUID  `json:"-"`
	Version         *string    `json:"-"`
	Account         *Account   `json:"-"`
	GrantGeneration *uuid.UUID `json:"-"`
}

type Content struct {
	ReferenceID uuid.UUID  `json:"referenceId"`
	Name        string     `json:"name"`
	MimeType    string     `json:"mimeType"`
	WebViewLink string     `json:"webViewLink"`
	ModifiedAt  *time.Time `json:"modifiedTime,omitempty"`
	Text        string     `json:"content"`
	ContentType string     `json:"contentType"`
	Truncated   bool       `json:"truncated"`
	BytesRead   int        `json:"bytesRead"`
}

type CreateOperation struct {
	ID             uuid.UUID
	WorkspaceID    uuid.UUID
	UserID         uuid.UUID
	IdempotencyKey string
	RequestHash    string
	TargetType     TargetType
	TargetID       uuid.UUID
	FileType       FileType
	Title          string
	Status         string
	GoogleFileID   *string
	ReferenceID    *uuid.UUID
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
