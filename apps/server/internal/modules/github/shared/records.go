package githubshared

import (
	"time"

	"github.com/google/uuid"
)

// RepositoryRecord is the minimum repository grant and routing context used by
// GitHub workflows. It deliberately contains no provider SDK objects.
type RepositoryRecord struct {
	ID                   uuid.UUID
	WorkspaceID          uuid.UUID
	WorkspaceSlug        string
	FullName             string
	OwnerLogin           string
	RepositorySlug       string
	DefaultBranch        string
	GitHubInstallationID int64
}

// StoryMatch is the story identity and workflow context resolved from a GitHub
// reference such as TEAM-123.
type StoryMatch struct {
	StoryID    uuid.UUID
	StatusID   uuid.UUID
	TeamID     uuid.UUID
	TeamCode   string
	SequenceID int
	Title      string
}

// FortyOneUser is the provider-neutral identity data required to attribute a
// GitHub comment to a FortyOne user.
type FortyOneUser struct {
	Username  string
	FullName  *string
	AvatarURL *string
}

type IssueSyncLinkRecord struct {
	ID             uuid.UUID
	RepositoryID   uuid.UUID
	RepositoryName string
	TeamID         uuid.UUID
	TeamName       string
	TeamColor      string
	SyncDirection  string
	IsActive       bool
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type BidirectionalIssueSyncLink struct {
	ID                   uuid.UUID
	RepositoryID         uuid.UUID
	TeamID               uuid.UUID
	SyncDirection        string
	RepositoryName       string
	OwnerLogin           string
	RepositorySlug       string
	RepositoryHTMLURL    string
	GitHubInstallationID int64
}

type IssueStoryLink struct {
	ID             uuid.UUID
	StoryID        uuid.UUID
	RepositoryID   uuid.UUID
	GitHubID       int64
	GitHubNumber   int
	URL            string
	Title          *string
	State          *string
	LastSyncedFrom *string
	LastSyncHash   *string
}

type TeamStatus struct {
	ID       uuid.UUID
	Name     string
	Category string
	Color    string
}

// StoryGitHubLink is the public GitHub link projection returned by the API.
type StoryGitHubLink struct {
	ID                 uuid.UUID `json:"id"`
	ExternalType       string    `json:"externalType"`
	GitHubNumber       *int      `json:"githubNumber"`
	URL                string    `json:"url"`
	Title              *string   `json:"title"`
	State              *string   `json:"state"`
	ReviewState        *string   `json:"reviewState"`
	ReviewsApproved    int       `json:"reviewsApproved"`
	ReviewsChangesReq  int       `json:"reviewsChangesRequested"`
	CheckState         *string   `json:"checkState"`
	RepositoryFullName string    `json:"repositoryFullName"`
	RefName            *string   `json:"refName"`
	CreatedAt          time.Time `json:"createdAt"`
}

// StoryIssueWithInstallation contains only the grant data needed to write a
// comment to a linked issue.
type StoryIssueWithInstallation struct {
	RepositoryID         uuid.UUID
	GitHubNumber         int
	OwnerLogin           string
	RepositorySlug       string
	GitHubInstallationID int64
}

// CredentialRecord is an encrypted provider credential plus the generation
// used in its credential-vault binding. Payload is never plaintext.
type CredentialRecord struct {
	UserID          uuid.UUID
	Payload         string
	EnvelopeVersion int
	Generation      uuid.UUID
}

// LegacyCredentialRecord exists only for the one-way vault migration path.
type LegacyCredentialRecord struct {
	UserID    uuid.UUID
	Plaintext string
}

// InstallationPayload and RepositoryPayload are stable provider-data transfer
// objects. SDK-specific types are mapped into these values at the edge.
type InstallationPayload struct {
	ID                  int64                      `json:"id"`
	Account             InstallationAccountPayload `json:"account"`
	RepositorySelection string                     `json:"repository_selection"`
	Permissions         map[string]string          `json:"permissions"`
	Events              []string                   `json:"events"`
	Sender              InstallationSenderPayload  `json:"sender"`
}

type InstallationAccountPayload struct {
	ID        int64   `json:"id"`
	Login     string  `json:"login"`
	Type      string  `json:"type"`
	AvatarURL *string `json:"avatar_url"`
}

type InstallationSenderPayload struct {
	ID int64 `json:"id"`
}

type RepositoryPayload struct {
	ID            int64                  `json:"id"`
	Name          string                 `json:"name"`
	FullName      string                 `json:"full_name"`
	Description   *string                `json:"description"`
	HTMLURL       string                 `json:"html_url"`
	CloneURL      string                 `json:"clone_url"`
	SSHURL        string                 `json:"ssh_url"`
	DefaultBranch string                 `json:"default_branch"`
	Private       bool                   `json:"private"`
	Archived      bool                   `json:"archived"`
	Disabled      bool                   `json:"disabled"`
	Owner         RepositoryOwnerPayload `json:"owner"`
}

type RepositoryOwnerPayload struct {
	ID    int64  `json:"id"`
	Login string `json:"login"`
}
