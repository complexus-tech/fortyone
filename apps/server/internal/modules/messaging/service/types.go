// Package messaging provides provider-neutral conversational assistance for
// messaging integrations such as Slack, Microsoft Teams, and WhatsApp.
package messaging

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
)

// ConversationRole is a provider-neutral role in a visible conversation.
type ConversationRole string

const (
	RoleUser      ConversationRole = "user"
	RoleAssistant ConversationRole = "assistant"
)

var (
	ErrInvalidRequest         = errors.New("invalid assistant request")
	ErrMessageTooLarge        = errors.New("assistant message is too large")
	ErrAssistantNotConfigured = errors.New("assistant is not configured")
	ErrResponseRefused        = errors.New("assistant response refused")
	ErrResponseIncomplete     = errors.New("assistant response incomplete")
	ErrResponseFailed         = errors.New("assistant response failed")
	ErrMalformedResponse      = errors.New("malformed assistant response")
	ErrMaxToolSteps           = errors.New("assistant exceeded maximum tool steps")
	ErrUnknownTool            = errors.New("unknown assistant tool")
	ErrInvalidToolArguments   = errors.New("invalid assistant tool arguments")
	ErrToolExecution          = errors.New("assistant tool execution failed")
	ErrTeamNotAccessible      = errors.New("team is not accessible to the user")
)

// Assistant answers a provider-neutral conversation request. Provider adapters
// are responsible for loading and persisting the visible conversation turns.
type Assistant interface {
	Respond(ctx context.Context, request Request) (Response, error)
}

// Request contains the authoritative FortyOne identity and the visible
// conversation leading up to the current prompt.
type Request struct {
	WorkspaceID  uuid.UUID
	UserID       uuid.UUID
	Conversation []ConversationTurn
	Prompt       string
}

// ConversationTurn is a visible user or assistant message. Provider-specific
// message identifiers intentionally live outside this package.
type ConversationTurn struct {
	Role ConversationRole
	Text string
}

// Response is the assistant's provider-neutral text response and aggregate
// model usage across every Responses API call needed to produce it.
type Response struct {
	Text  string
	Usage Usage
}

// Usage contains aggregate token usage for quota accounting.
type Usage struct {
	InputTokens  int
	OutputTokens int
	TotalTokens  int
}

// ToolScope is the complete authorization scope available to a tool call.
type ToolScope struct {
	WorkspaceID uuid.UUID
	UserID      uuid.UUID
}

// ToolDefinition is an OpenAI Responses function tool definition. Parameters
// must be a strict JSON schema.
type ToolDefinition struct {
	Type        string         `json:"type"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  map[string]any `json:"parameters"`
	Strict      bool           `json:"strict"`
}

// ToolCall is a normalized Responses API function call.
type ToolCall struct {
	ID        string
	Name      string
	Arguments json.RawMessage
}

// ToolExecutor exposes a fixed read-only tool catalog and executes calls within
// an authoritative FortyOne workspace/user scope.
type ToolExecutor interface {
	Definitions() []ToolDefinition
	Execute(ctx context.Context, scope ToolScope, call ToolCall) (json.RawMessage, error)
}
