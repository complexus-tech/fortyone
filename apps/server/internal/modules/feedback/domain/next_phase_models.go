package feedbackdomain

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/google/uuid"
)

// SummaryText returns the optional short publication summary or the title when
// no explicit summary was supplied.
func (update CoreFeedbackUpdate) SummaryText() string {
	if update.Summary != nil && strings.TrimSpace(*update.Summary) != "" {
		return strings.TrimSpace(*update.Summary)
	}
	return update.Title
}

const (
	ParticipationModeVerifiedGuest = "verified_guest"

	ParticipationIntentVerifiedGuest = "verified_guest"
	ParticipationIntentExternal      = "external"

	ContributorKindVerifiedGuest = "verified_guest"
	ContributorKindExternal      = "external"

	GuestIdentityPolicyShowIdentity       = "show_identity"
	GuestIdentityPolicyAllowPublicMasking = "allow_public_masking"
	GuestIdentityPolicyAlwaysMaskGuests   = "always_mask_guests"

	ContributorSessionSourcePortal      = "portal"
	ContributorSessionSourceWidget      = "widget"
	ContributorSessionSourcePreferences = "preferences"

	FeedbackUpdateStatusDraft     = "draft"
	FeedbackUpdateStatusPublished = "published"
)

type CoreParticipant struct {
	ID              uuid.UUID
	PortalID        uuid.UUID
	UserID          uuid.UUID
	Kind            string
	Email           string
	EmailVerifiedAt *time.Time
	DisplayName     string
	AvatarURL       *string
	ExternalID      string
	PublicMasked    bool
	BlockedAt       *time.Time
	BlockedReason   string
	LastSeenAt      *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type CoreParticipantSession struct {
	ID            uuid.UUID
	PortalID      uuid.UUID
	ContributorID uuid.UUID
	Source        string
	ExpiresAt     time.Time
	RevokedAt     *time.Time
	LastUsedAt    *time.Time
	CreatedAt     time.Time
}

type CoreVerificationRequest struct {
	PortalID     uuid.UUID
	Email        string
	DisplayName  string
	PublicMasked bool
	TokenHash    []byte
	CodeHash     []byte
	Source       string
	ExpiresAt    time.Time
}

type CoreVerificationChallenge struct {
	ID        uuid.UUID
	ExpiresAt time.Time
}

type CoreVerificationConfirmation struct {
	PortalID         uuid.UUID
	TokenHash        []byte
	Email            string
	CodeHash         []byte
	Source           string
	SessionTokenHash []byte
	SessionExpiresAt time.Time
}

type CoreContributorSessionResult struct {
	Participant CoreParticipant
	Session     CoreParticipantSession
	Token       string
}

type CoreUnsubscribeTokenInput struct {
	PortalID      uuid.UUID
	ContributorID uuid.UUID
	ItemID        *uuid.UUID
	Purpose       string
	TokenHash     []byte
	ExpiresAt     time.Time
}

type CoreResolvedParticipant struct {
	Participant CoreParticipant
	AccountID   uuid.UUID
	SessionID   uuid.UUID
}

type CoreContributorItemInput struct {
	Item        CoreItemInput
	Participant CoreParticipant
}

type CoreContributorCommentInput struct {
	WorkspaceID uuid.UUID
	PortalID    uuid.UUID
	ItemID      uuid.UUID
	Participant CoreParticipant
	ParentID    *uuid.UUID
	Body        string
}

type CoreContributorVoteInput struct {
	WorkspaceID uuid.UUID
	ItemID      uuid.UUID
	Participant CoreParticipant
	Vote        int
}

type CoreFollowState struct {
	ItemID        uuid.UUID
	ItemSlug      string
	Title         string
	ContributorID uuid.UUID
	Following     bool
	CreatedAt     *time.Time
}

type CoreContributorPreferences struct {
	PortalID            uuid.UUID
	ContributorID       uuid.UUID
	PortalEmailsEnabled bool
	ItemFollows         []CoreFollowState
	UpdatedAt           time.Time
}

type CoreDeliveryRecipient struct {
	ContributorID uuid.UUID
	Email         string
	DisplayName   string
	Kind          string
}

type CoreAccountUpdateRecipient struct {
	UserID uuid.UUID
	ItemID uuid.UUID
}

type CoreMergeItemInput struct {
	WorkspaceID  uuid.UUID
	SourceItemID uuid.UUID
	TargetItemID uuid.UUID
	ActorID      uuid.UUID
	Access       CoreAccessScope
}

type CoreMergeOutboxEvent struct {
	EventID      uuid.UUID
	WorkspaceID  uuid.UUID
	PortalID     uuid.UUID
	SourceItemID uuid.UUID
	TargetItemID uuid.UUID
	ActorID      uuid.UUID
	MergedAt     time.Time
	ClaimToken   uuid.UUID
	AttemptCount int
	Payload      json.RawMessage
}

type CoreMergeRecipient struct {
	ContributorID uuid.UUID
	UserID        uuid.UUID
	Kind          string
}

type CoreMergeCandidate struct {
	ID           uuid.UUID
	Slug         string
	Title        string
	Status       string
	VoteCount    int
	CommentCount int
}

type CoreMergeCandidatesPage struct {
	Candidates []CoreMergeCandidate
	HasMore    bool
}

type CoreCreateDeliveryInput struct {
	DeliveryID     uuid.UUID
	PortalID       uuid.UUID
	ContributorID  uuid.UUID
	ItemID         *uuid.UUID
	UpdateID       *uuid.UUID
	EventType      string
	DedupeKey      string
	Subject        string
	Message        string
	DestinationURL string
	TokenHash      []byte
}

type CoreDelivery struct {
	ID                 uuid.UUID
	PortalID           uuid.UUID
	ContributorID      uuid.UUID
	Email              string
	DisplayName        string
	PortalName         string
	PortalSlug         string
	ItemID             *uuid.UUID
	UpdateID           *uuid.UUID
	EventType          string
	DedupeKey          string
	Subject            string
	Message            string
	DestinationURL     string
	Status             string
	AttemptCount       int
	FinalFailureReason string
	CreatedAt          time.Time
}

type CoreFeedbackUpdate struct {
	ID                uuid.UUID
	WorkspaceID       uuid.UUID
	PortalID          uuid.UUID
	Slug              string
	Title             string
	Summary           *string
	Body              string
	CoverImageURL     *string
	Status            string
	PublishedAt       *time.Time
	PublishedByUserID *uuid.UUID
	LinkedItems       []CoreUpdateItem
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type CoreUpdateItem struct {
	ID     uuid.UUID
	Slug   string
	Title  string
	Status string
}

type CoreUpdateInput struct {
	WorkspaceID   uuid.UUID
	PortalID      uuid.UUID
	UpdateID      uuid.UUID
	ActorID       uuid.UUID
	Title         string
	Slug          string
	Summary       *string
	Body          string
	CoverImageURL *string
	ItemIDs       []uuid.UUID
	Access        CoreAccessScope
}

type CoreUpdatesPage struct {
	Updates  []CoreFeedbackUpdate
	Page     int
	PageSize int
	HasMore  bool
}

type CoreWidgetSettings struct {
	PortalID                 uuid.UUID
	Enabled                  bool
	WidgetKeyID              uuid.UUID
	AllowedOrigins           []string
	SigningSecretEncrypted   string
	SigningSecretVersion     int
	PreviousVersionExpiresAt *time.Time
	CreatedAt                time.Time
	UpdatedAt                time.Time
}

type CoreWidgetSettingsInput struct {
	WorkspaceID    uuid.UUID
	PortalID       uuid.UUID
	Enabled        bool
	AllowedOrigins []string
	Access         CoreAccessScope
}

type CoreWidgetSecretResult struct {
	Settings      CoreWidgetSettings
	SigningSecret string
}

type CoreWidgetIdentityAssertion struct {
	Version     int    `json:"version"`
	KeyID       string `json:"keyId"`
	ExternalID  string `json:"externalId"`
	Email       string `json:"email"`
	DisplayName string `json:"displayName"`
	AvatarURL   string `json:"avatarUrl,omitempty"`
	IssuedAt    int64  `json:"iat"`
	ExpiresAt   int64  `json:"exp"`
	Nonce       string `json:"nonce"`
	Origin      string `json:"origin"`
}

type CoreWidgetSessionInput struct {
	PortalSlug   string
	Assertion    string
	ParentOrigin string
}
