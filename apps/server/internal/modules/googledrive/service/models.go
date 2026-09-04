package googledrive

import (
	"context"
	"time"

	"github.com/complexus-tech/projects-api/internal/modules/googledrive/domain"
	"github.com/complexus-tech/projects-api/internal/platform/credentialvault"
	"github.com/google/uuid"
)

type Config struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	PickerAPIKey string
	AppID        string
	WebsiteURL   string
	Credentials  CredentialVault
}

type Integration struct {
	Configured              bool    `json:"configured"`
	Connected               bool    `json:"connected"`
	Email                   *string `json:"email,omitempty"`
	Status                  string  `json:"status"`
	RequiresReauthorization bool    `json:"requiresReauthorization"`
}

type PickerSession struct {
	AccessToken string  `json:"accessToken"`
	APIKey      string  `json:"apiKey"`
	AppID       string  `json:"appId"`
	Origin      *string `json:"origin,omitempty"`
}

type SelectedFile struct {
	ID          string
	Name        *string
	MimeType    *string
	ResourceKey *string
}

type AttachInput struct {
	WorkspaceID uuid.UUID
	UserID      uuid.UUID
	TargetType  domain.TargetType
	TargetID    uuid.UUID
	Files       []SelectedFile
}

type CreateFileInput struct {
	WorkspaceID    uuid.UUID
	UserID         uuid.UUID
	WorkspaceSlug  string
	TargetType     domain.TargetType
	TargetID       uuid.UUID
	FileType       domain.FileType
	Title          string
	IdempotencyKey string
}

type ImportInput struct {
	WorkspaceID    uuid.UUID
	UserID         uuid.UUID
	ReferenceID    uuid.UUID
	Visibility     string
	IdempotencyKey string
}

type ImportResult struct {
	DocumentID        uuid.UUID `json:"documentId"`
	SourceReferenceID uuid.UUID `json:"sourceReferenceId"`
}

type Repository interface {
	SaveOAuthState(context.Context, domain.OAuthState) error
	ConsumeOAuthState(context.Context, string, time.Time) (domain.OAuthState, error)
	WithinProviderUserLifecycle(context.Context, uuid.UUID, func(context.Context) error) error
	WithinProviderSubjectLifecycle(context.Context, string, func(context.Context) error) error
	UpsertConnection(context.Context, uuid.UUID, domain.Account) (domain.Connection, error)
	GetConnection(context.Context, uuid.UUID, uuid.UUID) (domain.Connection, error)
	GetActiveAccountBySubject(context.Context, string) (domain.Account, error)
	CompareAndSwapCredential(context.Context, domain.Account, string, time.Time) (bool, error)
	MarkReauthorizationRequired(context.Context, domain.Account, string) error
	Disconnect(context.Context, uuid.UUID, uuid.UUID) (bool, error)
	EnqueueRevocation(context.Context, domain.Revocation) (domain.RevocationCandidate, error)

	TargetAccessible(context.Context, uuid.UUID, uuid.UUID, domain.TargetType, uuid.UUID) (bool, error)
	TargetMutable(context.Context, uuid.UUID, uuid.UUID, domain.TargetType, uuid.UUID) (bool, error)
	AttachFile(context.Context, uuid.UUID, uuid.UUID, uuid.UUID, domain.TargetType, uuid.UUID, domain.ProviderFile) (domain.FileReference, error)
	AttachFiles(context.Context, uuid.UUID, uuid.UUID, uuid.UUID, domain.TargetType, uuid.UUID, []domain.ProviderFile) ([]domain.FileReference, error)
	ListReferences(context.Context, uuid.UUID, uuid.UUID, domain.TargetType, uuid.UUID) ([]domain.FileReference, error)
	GetReference(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) (domain.FileReference, error)
	RevalidateReference(context.Context, uuid.UUID, uuid.UUID, uuid.UUID, domain.FileReference, domain.ProviderFile) (uuid.UUID, error)
	DeleteReferenceGrant(context.Context, uuid.UUID, uuid.UUID, uuid.UUID, uuid.UUID, uuid.UUID) error
	MarkReferenceUnavailable(context.Context, uuid.UUID, uuid.UUID) error
	DeleteReference(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) error
	CreateOperation(context.Context, domain.CreateOperation) (domain.CreateOperation, bool, error)
	ClaimOperation(context.Context, uuid.UUID, time.Time, time.Time) (domain.CreateOperation, bool, error)
	CompleteOperation(context.Context, uuid.UUID, string, uuid.UUID) error
	FailOperation(context.Context, uuid.UUID, string) error
	CreateImportOperation(context.Context, domain.ImportOperation) (domain.ImportOperation, bool, error)
	ClaimImportOperation(context.Context, uuid.UUID, uuid.UUID, time.Time, time.Time) (domain.ImportOperation, bool, error)
	FailImportOperation(context.Context, uuid.UUID, uuid.UUID, string) error
	FinalizeDocumentImport(context.Context, domain.ImportFinalization) (uuid.UUID, error)
}

type CredentialVault interface {
	Seal(credentialvault.Context, []byte) (string, error)
	Open(credentialvault.Context, string) (credentialvault.Secret, error)
}

type ProviderClient interface {
	AuthorizationURL(state, verifier string) (string, error)
	Exchange(context.Context, string, string) (domain.OAuthToken, []string, error)
	Refresh(context.Context, string) (domain.OAuthToken, error)
	UserInfo(context.Context, string) (ProviderUser, error)
	Revoke(context.Context, string) error
	GetFile(context.Context, string, string, *string) (domain.ProviderFile, error)
	CreateFile(context.Context, string, domain.FileType, string, string) (domain.ProviderFile, error)
	FindCreatedFile(context.Context, string, string) (*domain.ProviderFile, error)
	PopulateFile(context.Context, string, domain.ProviderFile, string) error
	ReadFile(context.Context, string, domain.ProviderFile, int64) (ProviderContent, error)
	ReadThumbnail(context.Context, string, string, int64) (Preview, error)
}

type ProviderUser struct {
	Subject     string
	Email       string
	DisplayName *string
}

type ProviderContent struct {
	Text        string
	ContentType string
	Truncated   bool
	BytesRead   int
}

type Preview struct {
	Bytes       []byte
	ContentType string
}
