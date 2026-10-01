package scimdomain

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrUnauthorized = errors.New("SCIM authentication is required")
	ErrForbidden    = errors.New("SCIM operation is not allowed")
	ErrNotFound     = errors.New("SCIM resource not found")
	ErrInvalidInput = errors.New("invalid SCIM input")
	ErrConflict     = errors.New("SCIM resource already exists")
	ErrChanged      = errors.New("SCIM resource changed")
)

const (
	UserSchema       = "urn:ietf:params:scim:schemas:core:2.0:User"
	EnterpriseSchema = "urn:ietf:params:scim:schemas:extension:enterprise:2.0:User"
	ListSchema       = "urn:ietf:params:scim:api:messages:2.0:ListResponse"
	PatchSchema      = "urn:ietf:params:scim:api:messages:2.0:PatchOp"
	ErrorSchema      = "urn:ietf:params:scim:api:messages:2.0:Error"
)

type Scope struct{ WorkspaceID, CredentialID, IssuerID uuid.UUID }
type Name struct {
	Formatted       string `json:"formatted,omitempty"`
	GivenName       string `json:"givenName,omitempty"`
	FamilyName      string `json:"familyName,omitempty"`
	MiddleName      string `json:"middleName,omitempty"`
	HonorificPrefix string `json:"honorificPrefix,omitempty"`
	HonorificSuffix string `json:"honorificSuffix,omitempty"`
}
type Email struct {
	Value   string `json:"value"`
	Type    string `json:"type,omitempty"`
	Primary bool   `json:"primary,omitempty"`
}
type Manager struct {
	Value       string `json:"value,omitempty"`
	Ref         string `json:"$ref,omitempty"`
	DisplayName string `json:"displayName,omitempty"`
}
type Enterprise struct {
	EmployeeNumber string   `json:"employeeNumber,omitempty"`
	CostCenter     string   `json:"costCenter,omitempty"`
	Organization   string   `json:"organization,omitempty"`
	Division       string   `json:"division,omitempty"`
	Department     string   `json:"department,omitempty"`
	Manager        *Manager `json:"manager,omitempty"`
}
type Profile struct {
	DisplayName string      `json:"displayName,omitempty"`
	Name        *Name       `json:"name,omitempty"`
	Emails      []Email     `json:"emails,omitempty"`
	Enterprise  *Enterprise `json:"enterprise,omitempty"`
}
type Input struct {
	ID          string          `json:"id,omitempty"`
	Meta        json.RawMessage `json:"meta,omitempty"`
	Schemas     []string        `json:"schemas"`
	UserName    string          `json:"userName"`
	ExternalID  *string         `json:"externalId,omitempty"`
	Active      *bool           `json:"active,omitempty"`
	DisplayName string          `json:"displayName,omitempty"`
	Name        *Name           `json:"name,omitempty"`
	Emails      []Email         `json:"emails,omitempty"`
	Enterprise  *Enterprise     `json:"urn:ietf:params:scim:schemas:extension:enterprise:2.0:User,omitempty"`
}
type User struct {
	ID, WorkspaceID, UserID uuid.UUID
	UserName                string
	ExternalID              *string
	Active                  bool
	Profile                 json.RawMessage
	PreviousRole            string
	Version                 int64
	CreatedAt, UpdatedAt    time.Time
}
type Filter struct {
	Field  string
	Value  string
	Active *bool
}
type Page struct {
	StartIndex, Count int
	Filter            Filter
}
type Credential struct {
	ID        uuid.UUID  `json:"id"`
	Name      string     `json:"name"`
	Prefix    string     `json:"prefix"`
	CreatedAt time.Time  `json:"createdAt"`
	ExpiresAt time.Time  `json:"expiresAt"`
	RevokedAt *time.Time `json:"revokedAt"`
}
type MintInput struct {
	Name         string `json:"name" validate:"required,max=100"`
	LifetimeDays int    `json:"lifetimeDays" validate:"required,min=1,max=365"`
}
type Minted struct {
	Credential Credential `json:"credential"`
	Token      string     `json:"token"`
}
type Status struct {
	Credentials     []Credential `json:"credentials"`
	ManagedUsers    int64        `json:"managedUsers"`
	PendingSeatSync bool         `json:"pendingSeatSync"`
	SeatSyncError   string       `json:"seatSyncError"`
}

type AdminScope struct{ WorkspaceID, ActorID uuid.UUID }
type Mutation struct {
	UserName        string
	ExternalID      *string
	Active          bool
	Profile         []byte
	Email, FullName string
}
type Patch struct {
	Schemas    []string    `json:"schemas"`
	Operations []Operation `json:"Operations"`
}
type Operation struct {
	Op    string          `json:"op"`
	Path  string          `json:"path,omitempty"`
	Value json.RawMessage `json:"value,omitempty"`
}

// Invalid carries protocol error classification without exposing provider input.
type Invalid struct{ Type, Detail string }

func (e *Invalid) Error() string { return e.Detail }
func (e *Invalid) Unwrap() error { return ErrInvalidInput }
