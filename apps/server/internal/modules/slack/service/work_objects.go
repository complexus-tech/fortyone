package slack

import (
	"errors"
	"regexp"
	"time"

	"github.com/google/uuid"
)

const (
	slackTaskEntityType            = "slack#/entities/task"
	slackUserFieldType             = "slack#/types/user"
	slackDateFieldType             = "slack#/types/date"
	slackStoryExternalRefType      = "story"
	slackRequestExternalRefType    = "request"
	slackObjectiveExternalRefType  = "objective"
	slackSprintExternalRefType     = "sprint"
	slackOpenStoryActionID         = "fortyone_open_story"
	slackOpenRequestActionID       = "fortyone_open_request"
	slackOpenObjectiveActionID     = "fortyone_open_objective"
	slackOpenSprintActionID        = "fortyone_open_sprint"
	slackEditStoryStatusActionID   = "fortyone_edit_story_status"
	slackEditStoryPriorityActionID = "fortyone_edit_story_priority"
	slackConfirmMutationActionID   = "fortyone_confirm_story_mutation"
	slackCancelMutationActionID    = "fortyone_cancel_story_mutation"
	slackWorkObjectTitleLimit      = 3000
	slackWorkObjectTextFieldLimit  = 3000
	slackWorkObjectSelectLimit     = 100
	slackButtonValueLimit          = 2000
)

var (
	ErrSlackStoryPreviewAccessDenied     = errors.New("slack story preview access was not granted")
	ErrInvalidFortyOneStoryURL           = errors.New("invalid FortyOne story URL")
	ErrSlackRequestPreviewAccessDenied   = errors.New("slack request preview access was not granted")
	ErrInvalidFortyOneRequestURL         = errors.New("invalid FortyOne request URL")
	ErrSlackObjectivePreviewAccessDenied = errors.New("slack objective preview access was not granted")
	ErrInvalidFortyOneObjectiveURL       = errors.New("invalid FortyOne objective URL")
	ErrSlackSprintPreviewAccessDenied    = errors.New("slack sprint preview access was not granted")
	ErrInvalidFortyOneSprintURL          = errors.New("invalid FortyOne sprint URL")

	workspaceSlugPattern  = regexp.MustCompile(`^[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?$`)
	storyReferencePattern = regexp.MustCompile(`^[A-Z_]+-[1-9][0-9]*$`)
	slackUserIDPattern    = regexp.MustCompile(`^[UW][A-Z0-9]+$`)
)

// FortyOneStoryLink is the trusted, canonical identity extracted from a
// production FortyOne story URL.
type FortyOneStoryLink struct {
	PostedURL      string
	CanonicalURL   string
	WorkspaceSlug  string
	StoryReference string
}

// FortyOneRequestLink is the trusted canonical identity extracted from a
// production FortyOne integration-request URL.
type FortyOneRequestLink struct {
	PostedURL     string
	CanonicalURL  string
	WorkspaceSlug string
	TeamID        uuid.UUID
	RequestID     uuid.UUID
}

// FortyOneObjectiveLink is the trusted canonical identity extracted from an
// objective URL.
type FortyOneObjectiveLink struct {
	PostedURL     string
	CanonicalURL  string
	WorkspaceSlug string
	TeamID        uuid.UUID
	ObjectiveID   uuid.UUID
}

// FortyOneSprintLink is the trusted canonical identity extracted from a sprint
// URL. Sprint pages use the stories sub-route as their canonical detail URL.
type FortyOneSprintLink struct {
	PostedURL     string
	CanonicalURL  string
	WorkspaceSlug string
	TeamID        uuid.UUID
	SprintID      uuid.UUID
}

// SlackStoryWorkObjectInput contains already-authorized story data. Callers
// must resolve the Slack actor and current FortyOne team membership before
// setting AccessGranted. The builder intentionally cannot produce story
// metadata when that proof is absent.
type SlackStoryWorkObjectInput struct {
	AccessGranted       bool
	Editable            bool
	ExternalID          string
	StoryURL            string
	Title               string
	Description         string
	Status              string
	StatusID            string
	StatusOptions       []SlackWorkObjectSelectOption
	StatusColor         string
	Priority            string
	AssigneeID          string
	AssigneeOptions     []SlackWorkObjectSelectOption
	AssigneeSlackUserID string
	AssigneeName        string
	CreatorSlackUserID  string
	CreatorName         string
	DueDate             *time.Time
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

// SlackRequestWorkObjectInput contains already-authorized request data. Request
// Work Objects are deliberately read-only: triage and conversion remain in
// FortyOne, while Slack presents the current request state and canonical link.
type SlackRequestWorkObjectInput struct {
	AccessGranted       bool
	RequestURL          string
	Title               string
	Description         string
	Status              string
	Priority            string
	AssigneeSlackUserID string
	AssigneeName        string
	CreatorSlackUserID  string
	CreatorName         string
	DueDate             *time.Time
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

// SlackObjectiveWorkObjectInput contains already-authorized objective data.
// Objective and sprint Work Objects are read-only in Slack for now.
type SlackObjectiveWorkObjectInput struct {
	AccessGranted   bool
	ObjectiveURL    string
	ExternalID      string
	Title           string
	Description     string
	Health          string
	Progress        string
	LeadSlackUserID string
	LeadName        string
	StartDate       *time.Time
	EndDate         *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// SlackSprintWorkObjectInput contains already-authorized sprint data.
type SlackSprintWorkObjectInput struct {
	AccessGranted bool
	SprintURL     string
	ExternalID    string
	Title         string
	Goal          string
	Status        string
	Progress      string
	StartDate     *time.Time
	EndDate       *time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type SlackWorkObjectMetadata struct {
	Entities []SlackWorkObjectEntity `json:"entities"`
}

type SlackWorkObjectEntity struct {
	AppUnfurlURL  string                       `json:"app_unfurl_url,omitempty"`
	URL           string                       `json:"url"`
	ExternalRef   SlackWorkObjectExternalRef   `json:"external_ref"`
	EntityType    string                       `json:"entity_type"`
	EntityPayload SlackWorkObjectEntityPayload `json:"entity_payload"`
}

type SlackWorkObjectExternalRef struct {
	ID   string `json:"id"`
	Type string `json:"type,omitempty"`
}

type SlackWorkObjectEntityPayload struct {
	Attributes   SlackWorkObjectAttributes       `json:"attributes"`
	Fields       map[string]SlackWorkObjectField `json:"fields,omitempty"`
	CustomFields []SlackWorkObjectCustomField    `json:"custom_fields,omitempty"`
	DisplayOrder []string                        `json:"display_order,omitempty"`
	Actions      *SlackWorkObjectActions         `json:"actions,omitempty"`
}

type SlackWorkObjectAttributes struct {
	Title                SlackWorkObjectTitle `json:"title"`
	DisplayID            string               `json:"display_id,omitempty"`
	MetadataLastModified int64                `json:"metadata_last_modified,omitempty"`
}

type SlackWorkObjectTitle struct {
	Text string               `json:"text"`
	Edit *SlackWorkObjectEdit `json:"edit,omitempty"`
}

type SlackWorkObjectField struct {
	Value    any                  `json:"value,omitempty"`
	Label    string               `json:"label,omitempty"`
	Type     string               `json:"type,omitempty"`
	Format   string               `json:"format,omitempty"`
	User     *SlackWorkObjectUser `json:"user,omitempty"`
	TagColor string               `json:"tag_color,omitempty"`
	Edit     *SlackWorkObjectEdit `json:"edit,omitempty"`
}

type SlackWorkObjectCustomField struct {
	Key   string `json:"key"`
	Label string `json:"label"`
	Value any    `json:"value"`
	Type  string `json:"type"`
}

type SlackWorkObjectEdit struct {
	Enabled  bool                       `json:"enabled"`
	Optional bool                       `json:"optional,omitempty"`
	Text     *SlackWorkObjectEditText   `json:"text,omitempty"`
	Select   *SlackWorkObjectEditSelect `json:"select,omitempty"`
}

type SlackWorkObjectEditText struct {
	MinLength int `json:"min_length,omitempty"`
	MaxLength int `json:"max_length,omitempty"`
}

type SlackWorkObjectEditSelect struct {
	CurrentValue  string                        `json:"current_value,omitempty"`
	StaticOptions []SlackWorkObjectSelectOption `json:"static_options,omitempty"`
}

type SlackWorkObjectSelectOption struct {
	Value string                    `json:"value"`
	Text  SlackWorkObjectOptionText `json:"text"`
}

type SlackWorkObjectOptionText struct {
	Type string `json:"type,omitempty"`
	Text string `json:"text"`
}

type SlackWorkObjectUser struct {
	UserID string `json:"user_id,omitempty"`
	Text   string `json:"text,omitempty"`
}

type SlackWorkObjectActions struct {
	PrimaryActions  []SlackWorkObjectAction `json:"primary_actions,omitempty"`
	OverflowActions []SlackWorkObjectAction `json:"overflow_actions,omitempty"`
}

type SlackWorkObjectAction struct {
	Text               string `json:"text"`
	ActionID           string `json:"action_id"`
	Value              string `json:"value,omitempty"`
	URL                string `json:"url,omitempty"`
	Style              string `json:"style,omitempty"`
	AccessibilityLabel string `json:"accessibility_label,omitempty"`
}

// SlackChatUnfurlRequest is the JSON contract accepted by chat.unfurl. An
// authorization request deliberately carries no entity metadata so it cannot
// disclose a story before the Slack user is linked.
type SlackChatUnfurlRequest struct {
	Channel          string                   `json:"channel,omitempty"`
	TS               string                   `json:"ts,omitempty"`
	UnfurlID         string                   `json:"unfurl_id,omitempty"`
	Source           string                   `json:"source,omitempty"`
	Metadata         *SlackWorkObjectMetadata `json:"metadata,omitempty"`
	UserAuthRequired bool                     `json:"user_auth_required,omitempty"`
	UserAuthURL      string                   `json:"user_auth_url,omitempty"`
	UserAuthMessage  string                   `json:"user_auth_message,omitempty"`
}

// SlackEntityDetailsRequest is the JSON contract accepted by
// entity.presentDetails. Unlike chat.unfurl, Slack expects exactly one entity
// and the app_unfurl_url field must be omitted from its metadata.
type SlackEntityDetailsRequest struct {
	TriggerID        string                   `json:"trigger_id"`
	Metadata         *SlackWorkObjectEntity   `json:"metadata,omitempty"`
	UserAuthRequired bool                     `json:"user_auth_required,omitempty"`
	UserAuthURL      string                   `json:"user_auth_url,omitempty"`
	Error            *SlackEntityDetailsError `json:"error,omitempty"`
}

type SlackEntityDetailsError struct {
	Status        string `json:"status"`
	CustomMessage string `json:"custom_message,omitempty"`
}

// SlackProviderPayload is the durable Slack-specific portion of an outbound
// message. It is stored separately from the accessible top-level text so a
// retried delivery can recreate Work Objects or interactive confirmations
// without coupling the generic outbox to Slack's Block Kit schema.
type SlackProviderPayload struct {
	Blocks               []SlackBlock                `json:"blocks,omitempty"`
	Metadata             *SlackWorkObjectMetadata    `json:"metadata,omitempty"`
	UnfurlLinks          *bool                       `json:"unfurl_links,omitempty"`
	UnfurlMedia          *bool                       `json:"unfurl_media,omitempty"`
	Authorization        *SlackDeliveryAuthorization `json:"authorization,omitempty"`
	RequestThreadBinding *SlackRequestThreadBinding  `json:"request_thread_binding,omitempty"`
	AuthorSlackUserID    string                      `json:"author_slack_user_id,omitempty"`
}

// SlackDeliveryAuthorization freezes the team boundary used to generate a
// channel response. It is persisted with the delivery but intentionally never
// forwarded to Slack. Recovery must prove this boundary is still authorized.
type SlackDeliveryAuthorization struct {
	AllowedTeamIDs []uuid.UUID `json:"allowed_team_ids"`
	SharedTeamIDs  []uuid.UUID `json:"shared_team_ids,omitempty"`
	ActorUserID    *uuid.UUID  `json:"actor_user_id,omitempty"`
	Scope          string      `json:"scope,omitempty"`
}

const slackDeliveryAuthorizationScopeActorMembership = "actor_membership"

// SlackRequestThreadBinding is a durable continuation executed only after the
// acknowledgement has a provider message ID. It is never forwarded to Slack.
type SlackRequestThreadBinding struct {
	IntegrationRequestID    uuid.UUID `json:"integration_request_id"`
	ExternalSourceMessageID string    `json:"external_source_message_id,omitempty"`
	SourceURL               string    `json:"source_url,omitempty"`
}

type SlackBlock struct {
	Type     string              `json:"type"`
	BlockID  string              `json:"block_id,omitempty"`
	Text     *SlackTextObject    `json:"text,omitempty"`
	Elements []SlackBlockElement `json:"elements,omitempty"`
}

type SlackTextObject struct {
	Type  string `json:"type"`
	Text  string `json:"text"`
	Emoji bool   `json:"emoji,omitempty"`
}

type SlackBlockElement struct {
	Type               string           `json:"type"`
	ActionID           string           `json:"action_id,omitempty"`
	Text               *SlackTextObject `json:"text,omitempty"`
	Value              string           `json:"value,omitempty"`
	URL                string           `json:"url,omitempty"`
	Style              string           `json:"style,omitempty"`
	AccessibilityLabel string           `json:"accessibility_label,omitempty"`
}

// SlackStoryCreationReceipt is the rich, accessible content passed to
// chat.postMessage. Text remains the notification and screen-reader fallback;
// Slack renders the Work Object beneath it from ProviderPayload.Metadata.
type SlackStoryCreationReceipt struct {
	Text            string
	ProviderPayload SlackProviderPayload
}

// SlackRequestCreationReceipt shares the durable Work Object delivery envelope
// with story receipts while preserving a request-specific builder and copy.
type SlackRequestCreationReceipt = SlackStoryCreationReceipt
