package slack

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"golang.org/x/net/html"
)

const (
	slackTaskEntityType           = "slack#/entities/task"
	slackUserFieldType            = "slack#/types/user"
	slackDateFieldType            = "slack#/types/date"
	slackStoryExternalRefType     = "story"
	slackOpenStoryActionID        = "fortyone_open_story"
	slackConfirmMutationActionID  = "fortyone_confirm_story_mutation"
	slackCancelMutationActionID   = "fortyone_cancel_story_mutation"
	slackWorkObjectTitleLimit     = 3000
	slackWorkObjectTextFieldLimit = 3000
	slackWorkObjectSelectLimit    = 100
	slackButtonValueLimit         = 2000
)

var (
	ErrSlackStoryPreviewAccessDenied = errors.New("Slack story preview access was not granted")
	ErrInvalidFortyOneStoryURL       = errors.New("invalid FortyOne story URL")

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
	Attributes SlackWorkObjectAttributes       `json:"attributes"`
	Fields     map[string]SlackWorkObjectField `json:"fields,omitempty"`
	Actions    *SlackWorkObjectActions         `json:"actions,omitempty"`
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
	Type     string               `json:"type,omitempty"`
	Format   string               `json:"format,omitempty"`
	User     *SlackWorkObjectUser `json:"user,omitempty"`
	TagColor string               `json:"tag_color,omitempty"`
	Edit     *SlackWorkObjectEdit `json:"edit,omitempty"`
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
	Channel          string                   `json:"channel"`
	TS               string                   `json:"ts"`
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
}

// SlackDeliveryAuthorization freezes the team boundary used to generate a
// channel response. It is persisted with the delivery but intentionally never
// forwarded to Slack. Recovery must prove this boundary is still authorized.
type SlackDeliveryAuthorization struct {
	AllowedTeamIDs []uuid.UUID `json:"allowed_team_ids"`
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

// ParseFortyOneStoryURL accepts only canonical production story routes under a
// single FortyOne workspace subdomain. This rejects look-alike hosts, API/docs
// routes, credentials, ports, and encoded path separators before any lookup.
func ParseFortyOneStoryURL(rawURL string) (FortyOneStoryLink, error) {
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil || !strings.EqualFold(parsed.Scheme, "https") || parsed.User != nil || parsed.Port() != "" {
		return FortyOneStoryLink{}, ErrInvalidFortyOneStoryURL
	}

	host := strings.ToLower(strings.TrimSuffix(parsed.Hostname(), "."))
	const domainSuffix = ".fortyone.app"
	if !strings.HasSuffix(host, domainSuffix) {
		return FortyOneStoryLink{}, ErrInvalidFortyOneStoryURL
	}
	workspaceSlug := strings.TrimSuffix(host, domainSuffix)
	if strings.Contains(workspaceSlug, ".") || !workspaceSlugPattern.MatchString(workspaceSlug) {
		return FortyOneStoryLink{}, ErrInvalidFortyOneStoryURL
	}

	escapedPath := strings.ToLower(parsed.EscapedPath())
	if strings.Contains(escapedPath, "%2f") || strings.Contains(escapedPath, "%5c") {
		return FortyOneStoryLink{}, ErrInvalidFortyOneStoryURL
	}
	segments := strings.Split(strings.Trim(parsed.Path, "/"), "/")
	if len(segments) != 2 || segments[0] != "work" {
		return FortyOneStoryLink{}, ErrInvalidFortyOneStoryURL
	}
	storyReference := strings.ToUpper(strings.TrimSpace(segments[1]))
	if !storyReferencePattern.MatchString(storyReference) {
		return FortyOneStoryLink{}, ErrInvalidFortyOneStoryURL
	}

	postedURL := *parsed
	postedURL.Scheme = "https"
	postedURL.Host = host
	canonicalURL := url.URL{
		Scheme: "https",
		Host:   host,
		Path:   "/work/" + storyReference,
	}
	return FortyOneStoryLink{
		PostedURL:      postedURL.String(),
		CanonicalURL:   canonicalURL.String(),
		WorkspaceSlug:  workspaceSlug,
		StoryReference: storyReference,
	}, nil
}

// BuildSlackStoryUnfurlRequest creates a Work Object response for one
// link_shared event after the caller has proven the actor can read the story.
func BuildSlackStoryUnfurlRequest(channelID, messageTS string, input SlackStoryWorkObjectInput) (SlackChatUnfurlRequest, error) {
	if err := validateSlackUnfurlDestination(channelID, messageTS); err != nil {
		return SlackChatUnfurlRequest{}, err
	}
	entity, _, err := buildSlackStoryWorkObject(input, true)
	if err != nil {
		return SlackChatUnfurlRequest{}, err
	}
	metadata := SlackWorkObjectMetadata{Entities: []SlackWorkObjectEntity{entity}}
	return SlackChatUnfurlRequest{
		Channel:  strings.TrimSpace(channelID),
		TS:       strings.TrimSpace(messageTS),
		Metadata: &metadata,
	}, nil
}

// BuildSlackStoryAuthenticationUnfurlRequest creates Slack's private account
// linking prompt. Use this only for an unlinked user. A linked user who cannot
// access the story must receive no unfurl at all, avoiding an existence leak.
func BuildSlackStoryAuthenticationUnfurlRequest(channelID, messageTS, authURL string) (SlackChatUnfurlRequest, error) {
	if err := validateSlackUnfurlDestination(channelID, messageTS); err != nil {
		return SlackChatUnfurlRequest{}, err
	}
	if !isSafeFortyOneHTTPSURL(authURL) {
		return SlackChatUnfurlRequest{}, errors.New("invalid FortyOne Slack account-link URL")
	}
	return SlackChatUnfurlRequest{
		Channel:          strings.TrimSpace(channelID),
		TS:               strings.TrimSpace(messageTS),
		UserAuthRequired: true,
		UserAuthURL:      strings.TrimSpace(authURL),
		UserAuthMessage:  "Connect your FortyOne account to preview this link.",
	}, nil
}

func BuildSlackStoryEntityDetailsRequest(triggerID string, input SlackStoryWorkObjectInput) (SlackEntityDetailsRequest, error) {
	triggerID = strings.TrimSpace(triggerID)
	if triggerID == "" {
		return SlackEntityDetailsRequest{}, errors.New("Slack entity details trigger is required")
	}
	entity, _, err := buildSlackStoryWorkObject(input, false)
	if err != nil {
		return SlackEntityDetailsRequest{}, err
	}
	return SlackEntityDetailsRequest{TriggerID: triggerID, Metadata: &entity}, nil
}

func BuildSlackStoryAuthenticationEntityDetailsRequest(triggerID, authURL string) (SlackEntityDetailsRequest, error) {
	triggerID = strings.TrimSpace(triggerID)
	if triggerID == "" {
		return SlackEntityDetailsRequest{}, errors.New("Slack entity details trigger is required")
	}
	if !isSafeFortyOneHTTPSURL(authURL) {
		return SlackEntityDetailsRequest{}, errors.New("invalid FortyOne Slack account-link URL")
	}
	return SlackEntityDetailsRequest{
		TriggerID:        triggerID,
		UserAuthRequired: true,
		UserAuthURL:      strings.TrimSpace(authURL),
	}, nil
}

func BuildSlackStoryEntityDetailsErrorRequest(triggerID, message string) (SlackEntityDetailsRequest, error) {
	triggerID = strings.TrimSpace(triggerID)
	if triggerID == "" {
		return SlackEntityDetailsRequest{}, errors.New("Slack entity details trigger is required")
	}
	message = truncateSlackWorkObjectText(message, 500)
	if message == "" {
		message = "FortyOne could not save these changes. Refresh the task and try again."
	}
	return SlackEntityDetailsRequest{
		TriggerID: triggerID,
		Error: &SlackEntityDetailsError{
			Status:        "edit_error",
			CustomMessage: message,
		},
	}, nil
}

// BuildSlackStoryCreationReceipt builds a Work Object notification while
// preserving the intentionally minimal top line: "Joseph created WEB-123".
func BuildSlackStoryCreationReceipt(creatorName string, input SlackStoryWorkObjectInput) (SlackStoryCreationReceipt, error) {
	entity, link, err := buildSlackStoryWorkObject(input, false)
	if err != nil {
		return SlackStoryCreationReceipt{}, err
	}
	creatorLabel := slackMrkdwnTextEscaper.Replace(strings.Join(strings.Fields(creatorName), " "))
	if creatorLabel == "" {
		creatorLabel = "A team member"
	}
	disableUnfurls := false
	metadata := SlackWorkObjectMetadata{Entities: []SlackWorkObjectEntity{entity}}
	return SlackStoryCreationReceipt{
		Text: fmt.Sprintf("%s created <%s|%s>", creatorLabel, link.CanonicalURL, link.StoryReference),
		ProviderPayload: SlackProviderPayload{
			Metadata:    &metadata,
			UnfurlLinks: &disableUnfurls,
			UnfurlMedia: &disableUnfurls,
		},
	}, nil
}

// BuildSlackMutationConfirmationProviderPayload returns a generic Block Kit
// confirmation payload suitable for the same durable provider_payload column
// used by rich story receipts. The opaque token is never rendered as text.
func BuildSlackMutationConfirmationProviderPayload(prompt, confirmationToken, slackUserID string) (SlackProviderPayload, error) {
	prompt = truncateSlackWorkObjectText(prompt, slackWorkObjectTextFieldLimit)
	confirmationToken = strings.TrimSpace(confirmationToken)
	if prompt == "" {
		return SlackProviderPayload{}, errors.New("Slack mutation confirmation prompt is required")
	}
	if confirmationToken == "" {
		return SlackProviderPayload{}, errors.New("Slack mutation confirmation token is invalid")
	}
	actionValue, err := encodeSlackMutationActionValue(slackUserID, confirmationToken)
	if err != nil {
		return SlackProviderPayload{}, err
	}
	return SlackProviderPayload{Blocks: []SlackBlock{
		{
			Type: "section",
			Text: &SlackTextObject{Type: "mrkdwn", Text: prompt},
		},
		{
			Type:    "actions",
			BlockID: "fortyone_story_mutation_confirmation",
			Elements: []SlackBlockElement{
				{
					Type:               "button",
					ActionID:           slackConfirmMutationActionID,
					Text:               &SlackTextObject{Type: "plain_text", Text: "Confirm"},
					Value:              actionValue,
					Style:              "primary",
					AccessibilityLabel: "Confirm story change",
				},
				{
					Type:               "button",
					ActionID:           slackCancelMutationActionID,
					Text:               &SlackTextObject{Type: "plain_text", Text: "Cancel"},
					Value:              actionValue,
					AccessibilityLabel: "Cancel story change",
				},
			},
		},
	}}, nil
}

type slackMutationActionValue struct {
	SlackUserID string `json:"slack_user_id"`
	Token       string `json:"token"`
}

func encodeSlackMutationActionValue(slackUserID, token string) (string, error) {
	value := slackMutationActionValue{
		SlackUserID: strings.TrimSpace(slackUserID),
		Token:       strings.TrimSpace(token),
	}
	if value.SlackUserID == "" || value.Token == "" {
		return "", errors.New("Slack mutation action actor and token are required")
	}
	encoded, err := json.Marshal(value)
	if err != nil {
		return "", fmt.Errorf("encode Slack mutation action: %w", err)
	}
	result := base64.RawURLEncoding.EncodeToString(encoded)
	if utf8.RuneCountInString(result) > slackButtonValueLimit {
		return "", errors.New("Slack mutation confirmation token is invalid")
	}
	return result, nil
}

func decodeSlackMutationActionValue(raw string) (slackMutationActionValue, error) {
	decoded, err := base64.RawURLEncoding.DecodeString(strings.TrimSpace(raw))
	if err != nil {
		return slackMutationActionValue{}, errors.New("invalid Slack mutation action")
	}
	var value slackMutationActionValue
	if err := json.Unmarshal(decoded, &value); err != nil {
		return slackMutationActionValue{}, errors.New("invalid Slack mutation action")
	}
	value.SlackUserID = strings.TrimSpace(value.SlackUserID)
	value.Token = strings.TrimSpace(value.Token)
	if value.SlackUserID == "" || value.Token == "" {
		return slackMutationActionValue{}, errors.New("invalid Slack mutation action")
	}
	return value, nil
}

func buildSlackStoryWorkObject(input SlackStoryWorkObjectInput, includeAppUnfurlURL bool) (SlackWorkObjectEntity, FortyOneStoryLink, error) {
	if !input.AccessGranted {
		return SlackWorkObjectEntity{}, FortyOneStoryLink{}, ErrSlackStoryPreviewAccessDenied
	}
	link, err := ParseFortyOneStoryURL(input.StoryURL)
	if err != nil {
		return SlackWorkObjectEntity{}, FortyOneStoryLink{}, err
	}
	title := truncateSlackWorkObjectText(input.Title, slackWorkObjectTitleLimit)
	if title == "" {
		return SlackWorkObjectEntity{}, FortyOneStoryLink{}, errors.New("Slack story Work Object title is required")
	}

	fields := make(map[string]SlackWorkObjectField, 8)
	description := truncateSlackWorkObjectText(slackWorkObjectDescription(input.Description), slackWorkObjectTextFieldLimit)
	if description != "" || input.Editable {
		descriptionField := SlackWorkObjectField{Value: description, Format: "markdown"}
		if input.Editable {
			descriptionField.Edit = &SlackWorkObjectEdit{
				Enabled:  true,
				Optional: true,
				Text:     &SlackWorkObjectEditText{MaxLength: modalDescriptionMaxRunes},
			}
		}
		fields["description"] = descriptionField
	}
	if createdBy := slackWorkObjectUser(input.CreatorSlackUserID, input.CreatorName); createdBy != nil {
		fields["created_by"] = SlackWorkObjectField{Type: slackUserFieldType, User: createdBy}
	}
	if !input.CreatedAt.IsZero() {
		fields["date_created"] = SlackWorkObjectField{Value: input.CreatedAt.UTC().Unix()}
	}
	if !input.UpdatedAt.IsZero() {
		fields["date_updated"] = SlackWorkObjectField{Value: input.UpdatedAt.UTC().Unix()}
	}
	assignee := slackWorkObjectUser(input.AssigneeSlackUserID, input.AssigneeName)
	if assignee != nil || (input.Editable && len(input.AssigneeOptions) > 0) {
		if assignee == nil {
			assignee = &SlackWorkObjectUser{Text: "No assignee"}
		}
		assigneeField := SlackWorkObjectField{Type: slackUserFieldType, User: assignee}
		if input.Editable && validSlackWorkObjectSelectOptions(input.AssigneeOptions) {
			assigneeField.Edit = &SlackWorkObjectEdit{
				Enabled:  true,
				Optional: true,
				Select: &SlackWorkObjectEditSelect{
					CurrentValue:  strings.TrimSpace(input.AssigneeID),
					StaticOptions: cloneSlackWorkObjectSelectOptions(input.AssigneeOptions),
				},
			}
		}
		fields["assignee"] = assigneeField
	}
	if status := truncateSlackWorkObjectText(input.Status, 255); status != "" {
		statusField := SlackWorkObjectField{Value: status, TagColor: normalizeSlackTagColor(input.StatusColor)}
		if input.Editable && strings.TrimSpace(input.StatusID) != "" && validSlackWorkObjectSelectOptions(input.StatusOptions) {
			statusField.Edit = &SlackWorkObjectEdit{
				Enabled: true,
				Select: &SlackWorkObjectEditSelect{
					CurrentValue:  strings.TrimSpace(input.StatusID),
					StaticOptions: cloneSlackWorkObjectSelectOptions(input.StatusOptions),
				},
			}
		}
		fields["status"] = statusField
	}
	if input.DueDate != nil && !input.DueDate.IsZero() {
		fields["due_date"] = SlackWorkObjectField{
			Value: input.DueDate.UTC().Format(time.DateOnly),
			Type:  slackDateFieldType,
			Edit:  slackOptionalWorkObjectEdit(input.Editable),
		}
	} else if input.Editable {
		fields["due_date"] = SlackWorkObjectField{
			Type: slackDateFieldType,
			Edit: slackOptionalWorkObjectEdit(true),
		}
	}
	if priority := truncateSlackWorkObjectText(input.Priority, 255); priority != "" {
		priorityField := SlackWorkObjectField{Value: priority}
		if input.Editable {
			priorityField.Edit = &SlackWorkObjectEdit{
				Enabled: true,
				Select: &SlackWorkObjectEditSelect{
					CurrentValue:  priority,
					StaticOptions: slackWorkObjectPriorityOptions(),
				},
			}
		}
		fields["priority"] = priorityField
	}

	lastModified := input.UpdatedAt
	if lastModified.IsZero() {
		lastModified = input.CreatedAt
	}
	entity := SlackWorkObjectEntity{
		URL: link.CanonicalURL,
		ExternalRef: SlackWorkObjectExternalRef{
			ID:   slackStoryExternalRefID(link, input.ExternalID),
			Type: slackStoryExternalRefType,
		},
		EntityType: slackTaskEntityType,
		EntityPayload: SlackWorkObjectEntityPayload{
			Attributes: SlackWorkObjectAttributes{
				Title: SlackWorkObjectTitle{
					Text: title,
					Edit: slackStoryTitleEdit(input.Editable),
				},
				DisplayID:            link.StoryReference,
				MetadataLastModified: unixTimestamp(lastModified),
			},
			Fields: fields,
			Actions: &SlackWorkObjectActions{
				PrimaryActions: []SlackWorkObjectAction{{
					Text:               "Open in FortyOne",
					ActionID:           slackOpenStoryActionID,
					Value:              link.StoryReference,
					URL:                link.CanonicalURL,
					AccessibilityLabel: "Open " + link.StoryReference + " in FortyOne",
				}},
			},
		},
	}
	if includeAppUnfurlURL {
		entity.AppUnfurlURL = link.PostedURL
	}
	return entity, link, nil
}

func slackStoryTitleEdit(enabled bool) *SlackWorkObjectEdit {
	if !enabled {
		return nil
	}
	return &SlackWorkObjectEdit{
		Enabled: true,
		Text: &SlackWorkObjectEditText{
			MinLength: 1,
			MaxLength: modalTitleMaxRunes,
		},
	}
}

func slackOptionalWorkObjectEdit(enabled bool) *SlackWorkObjectEdit {
	if !enabled {
		return nil
	}
	return &SlackWorkObjectEdit{Enabled: true, Optional: true}
}

func slackWorkObjectPriorityOptions() []SlackWorkObjectSelectOption {
	priorities := []string{slackPriorityNoPriority, "Low", "Medium", "High", "Urgent"}
	options := make([]SlackWorkObjectSelectOption, 0, len(priorities))
	for _, priority := range priorities {
		options = append(options, newSlackWorkObjectSelectOption(priority, priority))
	}
	return options
}

func newSlackWorkObjectSelectOption(value, label string) SlackWorkObjectSelectOption {
	return SlackWorkObjectSelectOption{
		Value: strings.TrimSpace(value),
		Text: SlackWorkObjectOptionText{
			Type: "plain_text",
			Text: truncateSlackWorkObjectText(label, 75),
		},
	}
}

func validSlackWorkObjectSelectOptions(options []SlackWorkObjectSelectOption) bool {
	if len(options) == 0 || len(options) > slackWorkObjectSelectLimit {
		return false
	}
	seen := make(map[string]struct{}, len(options))
	for _, option := range options {
		value := strings.TrimSpace(option.Value)
		label := strings.TrimSpace(option.Text.Text)
		if value == "" || label == "" || len(value) > 150 {
			return false
		}
		if _, exists := seen[value]; exists {
			return false
		}
		seen[value] = struct{}{}
	}
	return true
}

func cloneSlackWorkObjectSelectOptions(options []SlackWorkObjectSelectOption) []SlackWorkObjectSelectOption {
	return append([]SlackWorkObjectSelectOption(nil), options...)
}

func slackStoryExternalRefID(link FortyOneStoryLink, externalID string) string {
	externalID = strings.TrimSpace(externalID)
	if externalID == "" {
		externalID = link.StoryReference
	}
	return strings.ToLower(link.WorkspaceSlug) + ":" + externalID
}

func slackStoryExternalRefMatches(link FortyOneStoryLink, externalID, actual string) bool {
	actual = strings.TrimSpace(actual)
	return actual == slackStoryExternalRefID(link, externalID) ||
		actual == slackStoryExternalRefID(link, "") ||
		strings.EqualFold(actual, link.StoryReference)
}

func validateSlackUnfurlDestination(channelID, messageTS string) error {
	channelID = strings.TrimSpace(channelID)
	messageTS = strings.TrimSpace(messageTS)
	if channelID == "" || messageTS == "" || strings.ContainsAny(channelID+messageTS, " \t\r\n") {
		return errors.New("Slack unfurl channel and timestamp are required")
	}
	return nil
}

func isSafeFortyOneHTTPSURL(rawURL string) bool {
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil || !strings.EqualFold(parsed.Scheme, "https") || parsed.User != nil || parsed.Port() != "" {
		return false
	}
	host := strings.ToLower(strings.TrimSuffix(parsed.Hostname(), "."))
	return host == "fortyone.app" || strings.HasSuffix(host, ".fortyone.app")
}

func slackWorkObjectUser(slackUserID, displayName string) *SlackWorkObjectUser {
	slackUserID = strings.ToUpper(strings.TrimSpace(slackUserID))
	if slackUserIDPattern.MatchString(slackUserID) {
		return &SlackWorkObjectUser{UserID: slackUserID}
	}
	displayName = truncateSlackWorkObjectText(displayName, 255)
	if displayName == "" {
		return nil
	}
	return &SlackWorkObjectUser{Text: displayName}
}

func normalizeSlackTagColor(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "red", "yellow", "green", "gray", "blue":
		return strings.ToLower(strings.TrimSpace(value))
	default:
		return ""
	}
}

func truncateSlackWorkObjectText(value string, limit int) string {
	value = strings.TrimSpace(value)
	if value == "" || limit <= 0 || utf8.RuneCountInString(value) <= limit {
		return value
	}
	runes := []rune(value)
	return strings.TrimSpace(string(runes[:limit-1])) + "…"
}

// slackWorkObjectDescription converts rich editor HTML to readable text before
// it enters a Slack Work Object. Plain-text descriptions are returned exactly
// as written so code examples such as "value <T>" are not mistaken for HTML.
func slackWorkObjectDescription(value string) string {
	value = strings.TrimSpace(value)
	if value == "" || !strings.Contains(value, "<") {
		return value
	}

	tokenizer := html.NewTokenizer(strings.NewReader(value))
	var output strings.Builder
	sawRichTextMarkup := false
	suppressedDepth := 0
	for {
		tokenType := tokenizer.Next()
		switch tokenType {
		case html.ErrorToken:
			if err := tokenizer.Err(); err != nil && !errors.Is(err, io.EOF) {
				return value
			}
			if !sawRichTextMarkup {
				return value
			}
			return normalizeSlackWorkObjectPlainText(output.String())
		case html.TextToken:
			if suppressedDepth == 0 {
				output.Write(tokenizer.Text())
			}
		case html.StartTagToken, html.SelfClosingTagToken, html.EndTagToken:
			token := tokenizer.Token()
			tag := strings.ToLower(token.Data)
			if !isSlackRichTextHTMLTag(tag) {
				continue
			}
			sawRichTextMarkup = true
			if tag == "script" || tag == "style" {
				if tokenType == html.StartTagToken {
					suppressedDepth++
				} else if tokenType == html.EndTagToken && suppressedDepth > 0 {
					suppressedDepth--
				}
				continue
			}
			if suppressedDepth > 0 {
				continue
			}
			if isSlackRichTextBlockTag(tag) {
				appendSlackWorkObjectLineBreak(&output)
			}
			if tag == "li" && tokenType != html.EndTagToken {
				output.WriteString("• ")
			}
		}
	}
}

func isSlackRichTextHTMLTag(tag string) bool {
	switch tag {
	case "a", "b", "blockquote", "br", "code", "del", "div", "em", "h1", "h2", "h3", "h4", "h5", "h6", "hr", "i", "li", "ol", "p", "pre", "s", "script", "span", "strong", "style", "table", "tbody", "td", "th", "thead", "tr", "u", "ul":
		return true
	default:
		return false
	}
}

func isSlackRichTextBlockTag(tag string) bool {
	switch tag {
	case "blockquote", "br", "div", "h1", "h2", "h3", "h4", "h5", "h6", "hr", "li", "ol", "p", "pre", "table", "tbody", "td", "th", "thead", "tr", "ul":
		return true
	default:
		return false
	}
}

func appendSlackWorkObjectLineBreak(output *strings.Builder) {
	current := output.String()
	if current != "" && !strings.HasSuffix(current, "\n") {
		output.WriteByte('\n')
	}
}

func normalizeSlackWorkObjectPlainText(value string) string {
	lines := strings.Split(strings.ReplaceAll(value, "\u00a0", " "), "\n")
	normalized := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.Join(strings.Fields(line), " ")
		if line == "" {
			continue
		}
		normalized = append(normalized, line)
	}
	return strings.Join(normalized, "\n")
}

func unixTimestamp(value time.Time) int64 {
	if value.IsZero() {
		return 0
	}
	return value.UTC().Unix()
}
