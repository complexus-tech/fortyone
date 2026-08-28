package messaging

import (
	"encoding/json"
	"strings"
)

type responsesAPIRequest struct {
	Model             string             `json:"model"`
	Instructions      string             `json:"instructions"`
	Input             []json.RawMessage  `json:"input"`
	Tools             []ToolDefinition   `json:"tools"`
	Store             bool               `json:"store"`
	Reasoning         responsesReasoning `json:"reasoning"`
	MaxOutputTokens   int                `json:"max_output_tokens"`
	ParallelToolCalls bool               `json:"parallel_tool_calls"`
	SafetyIdentifier  string             `json:"safety_identifier"`
	Include           []string           `json:"include"`
}

type responsesReasoning struct {
	Effort  string `json:"effort"`
	Context string `json:"context,omitempty"`
}

func responsesReasoningForModel(model string) responsesReasoning {
	reasoning := responsesReasoning{Effort: legacyMessagingReasoningEffort}
	if strings.HasPrefix(strings.ToLower(strings.TrimSpace(model)), "gpt-5.6") {
		reasoning.Effort = defaultMessagingReasoningEffort
		reasoning.Context = "current_turn"
	}
	return reasoning
}

type messageInput struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type functionCallOutputInput struct {
	Type   string `json:"type"`
	CallID string `json:"call_id"`
	Output string `json:"output"`
}

type responsesAPIResponse struct {
	ID                string                      `json:"id"`
	Status            string                      `json:"status"`
	Output            []json.RawMessage           `json:"output"`
	Error             *openAIError                `json:"error"`
	IncompleteDetails *responsesIncompleteDetails `json:"incomplete_details"`
	Usage             responsesUsage              `json:"usage"`
}

type responsesIncompleteDetails struct {
	Reason string `json:"reason"`
}

type responsesUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
	TotalTokens  int `json:"total_tokens"`
}

type openAIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Type    string `json:"type"`
}

func (e *openAIError) message() string {
	if e == nil {
		return ""
	}
	if strings.TrimSpace(e.Message) != "" {
		return e.Message
	}
	return e.Type
}

type functionCallOutput struct {
	Type      string `json:"type"`
	CallID    string `json:"call_id"`
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type messageOutput struct {
	Type    string            `json:"type"`
	Content []json.RawMessage `json:"content"`
}

type outputContent struct {
	Type    string `json:"type"`
	Text    string `json:"text"`
	Refusal string `json:"refusal"`
}

type outputAnalysis struct {
	text    string
	refusal string
	calls   []ToolCall
}
