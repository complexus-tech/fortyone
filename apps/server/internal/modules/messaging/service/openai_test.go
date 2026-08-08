package messaging

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestNewOpenAIAssistantUsesDefaultModel(t *testing.T) {
	t.Parallel()

	assistant, err := NewOpenAIAssistant(
		OpenAIConfig{APIKey: "test-key"},
		&assistantToolStub{},
	)
	if err != nil {
		t.Fatalf("construct assistant: %v", err)
	}
	if assistant.model != "gpt-5.4-mini" {
		t.Fatalf("expected default model gpt-5.4-mini, got %q", assistant.model)
	}
}

func TestOpenAIAssistantRespondPreservesCompleteOutputAcrossToolLoop(t *testing.T) {
	t.Parallel()

	workspaceID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	userID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	executor := &assistantToolStub{result: json.RawMessage(`{"teams":[]}`)}

	var requestNumber atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/v1/responses" {
			t.Errorf("unexpected request path %q", request.URL.Path)
		}
		if authorization := request.Header.Get("Authorization"); authorization != "Bearer test-key" {
			t.Errorf("unexpected authorization header %q", authorization)
		}

		requestBody, err := io.ReadAll(request.Body)
		if err != nil {
			t.Errorf("read request: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		var envelope map[string]json.RawMessage
		if err := json.Unmarshal(requestBody, &envelope); err != nil {
			t.Errorf("decode request envelope: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if rawStore, present := envelope["store"]; !present || string(rawStore) != "false" {
			t.Errorf("request must explicitly send store:false, got %s", rawStore)
		}
		var body responsesAPIRequest
		if err := json.Unmarshal(requestBody, &body); err != nil {
			t.Errorf("decode request: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if body.Store {
			t.Error("store must be false")
		}
		if body.Reasoning.Effort != "low" {
			t.Errorf("expected low reasoning, got %q", body.Reasoning.Effort)
		}
		if body.ParallelToolCalls {
			t.Error("parallel tool calls must be disabled")
		}
		if body.MaxOutputTokens != 321 {
			t.Errorf("expected bounded output of 321, got %d", body.MaxOutputTokens)
		}
		if len(body.Include) != 1 || body.Include[0] != "reasoning.encrypted_content" {
			t.Errorf("encrypted reasoning include missing: %#v", body.Include)
		}
		if body.SafetyIdentifier != safetyIdentifier(userID.String()) {
			t.Errorf("unexpected safety identifier %q", body.SafetyIdentifier)
		}
		if strings.Contains(body.SafetyIdentifier, userID.String()) {
			t.Error("safety identifier must not contain the raw user ID")
		}

		switch requestNumber.Add(1) {
		case 1:
			if len(body.Input) != 3 {
				t.Errorf("expected two conversation turns and prompt, got %d inputs", len(body.Input))
			}
			assertMessageInput(t, body.Input[0], "user", "What am I working on?")
			assertMessageInput(t, body.Input[1], "assistant", "Which area should I check?")
			assertMessageInput(t, body.Input[2], "user", "List my teams")
			io.WriteString(w, `{
  "id":"resp_1",
  "status":"completed",
  "output":[
    {"type":"reasoning","id":"rs_1","encrypted_content":"opaque-reasoning","summary":[]},
    {"type":"function_call","id":"fc_1","call_id":"call_1","name":"list_teams","arguments":"{}","status":"completed"},
    {"type":"function_call","id":"fc_2","call_id":"call_2","name":"list_teams","arguments":"{}","status":"completed"}
  ],
  "usage":{"input_tokens":10,"output_tokens":4,"total_tokens":14}
}`)
		case 2:
			if len(body.Input) != 8 {
				t.Errorf("expected original input, all output items, and two tool outputs; got %d", len(body.Input))
			}
			var reasoning map[string]any
			if err := json.Unmarshal(body.Input[3], &reasoning); err != nil {
				t.Errorf("decode preserved reasoning: %v", err)
			}
			if reasoning["encrypted_content"] != "opaque-reasoning" {
				t.Errorf("reasoning item was not completely preserved: %#v", reasoning)
			}
			assertFunctionCall(t, body.Input[4], "call_1")
			assertFunctionCall(t, body.Input[5], "call_2")
			assertFunctionOutput(t, body.Input[6], "call_1", `{"teams":[]}`)
			assertFunctionOutput(t, body.Input[7], "call_2", `{"teams":[]}`)
			io.WriteString(w, `{
  "id":"resp_2",
  "status":"completed",
  "output":[{"type":"message","role":"assistant","status":"completed","content":[
    {"type":"output_text","text":"You belong to ","annotations":[]},
    {"type":"output_text","text":"no teams.","annotations":[]}
  ]}],
  "usage":{"input_tokens":20,"output_tokens":5,"total_tokens":25}
}`)
		default:
			t.Errorf("unexpected extra Responses API call")
			w.WriteHeader(http.StatusInternalServerError)
		}
	}))
	defer server.Close()

	assistant := mustTestAssistant(t, server.URL+"/v1", executor, OpenAIConfig{MaxOutputTokens: 321})
	response, err := assistant.Respond(context.Background(), Request{
		WorkspaceID: workspaceID,
		UserID:      userID,
		Conversation: []ConversationTurn{
			{Role: RoleUser, Text: "What am I working on?"},
			{Role: RoleAssistant, Text: "Which area should I check?"},
		},
		Prompt: "List my teams",
	})
	if err != nil {
		t.Fatalf("respond: %v", err)
	}
	if response.Text != "You belong to no teams." {
		t.Fatalf("unexpected response text %q", response.Text)
	}
	if response.Usage != (Usage{InputTokens: 30, OutputTokens: 9, TotalTokens: 39}) {
		t.Fatalf("unexpected aggregate usage %#v", response.Usage)
	}
	if requestNumber.Load() != 2 {
		t.Fatalf("expected two Responses API calls, got %d", requestNumber.Load())
	}
	if calls := executor.Calls(); len(calls) != 2 || calls[0].Name != toolListTeams || calls[1].Name != toolListTeams {
		t.Fatalf("unexpected tool calls %#v", calls)
	}
	for _, scope := range executor.Scopes() {
		if scope.WorkspaceID != workspaceID || scope.UserID != userID {
			t.Fatalf("tool received incorrect scope %#v", scope)
		}
	}
}

func TestOpenAIAssistantResponseErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		body     string
		expected error
	}{
		{
			name:     "incomplete",
			body:     `{"status":"incomplete","output":[],"incomplete_details":{"reason":"max_output_tokens"}}`,
			expected: ErrResponseIncomplete,
		},
		{
			name:     "failed",
			body:     `{"status":"failed","output":[],"error":{"code":"server_error","message":"try again"}}`,
			expected: ErrResponseFailed,
		},
		{
			name:     "refusal",
			body:     `{"status":"completed","output":[{"type":"message","content":[{"type":"refusal","refusal":"I cannot help with that."}]}]}`,
			expected: ErrResponseRefused,
		},
		{
			name:     "malformed json",
			body:     `{"status":`,
			expected: ErrMalformedResponse,
		},
		{
			name:     "malformed function arguments",
			body:     `{"status":"completed","output":[{"type":"function_call","call_id":"call_1","name":"list_teams","arguments":"not-json"}]}`,
			expected: ErrMalformedResponse,
		},
		{
			name:     "unknown tool",
			body:     `{"status":"completed","output":[{"type":"function_call","call_id":"call_1","name":"delete_story","arguments":"{}"}]}`,
			expected: ErrUnknownTool,
		},
		{
			name:     "empty completed output",
			body:     `{"status":"completed","output":[]}`,
			expected: ErrMalformedResponse,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				io.WriteString(w, test.body)
			}))
			defer server.Close()

			assistant := mustTestAssistant(t, server.URL, &assistantToolStub{result: json.RawMessage(`{}`)}, OpenAIConfig{})
			_, err := assistant.Respond(context.Background(), validTestRequest())
			if !errors.Is(err, test.expected) {
				t.Fatalf("expected %v, got %v", test.expected, err)
			}
		})
	}
}

func TestOpenAIAssistantStopsAfterSixToolSteps(t *testing.T) {
	t.Parallel()

	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requestNumber := requests.Add(1)
		fmt.Fprintf(w, `{"status":"completed","output":[{"type":"function_call","call_id":"call_%d","name":"list_teams","arguments":"{}"}]}`, requestNumber)
	}))
	defer server.Close()

	executor := &assistantToolStub{result: json.RawMessage(`{"teams":[]}`)}
	assistant := mustTestAssistant(t, server.URL, executor, OpenAIConfig{})
	_, err := assistant.Respond(context.Background(), validTestRequest())
	if !errors.Is(err, ErrMaxToolSteps) {
		t.Fatalf("expected max-step error, got %v", err)
	}
	if requests.Load() != maximumToolSteps+1 {
		t.Fatalf("expected %d model calls, got %d", maximumToolSteps+1, requests.Load())
	}
	if calls := executor.Calls(); len(calls) != maximumToolSteps {
		t.Fatalf("expected %d executed tools, got %d", maximumToolSteps, len(calls))
	}
}

func TestOpenAIAssistantReturnsStructuredAPIError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("x-request-id", "req_test")
		w.WriteHeader(http.StatusTooManyRequests)
		io.WriteString(w, `{"error":{"code":"rate_limit_exceeded","message":"slow down","type":"requests"}}`)
	}))
	defer server.Close()

	assistant := mustTestAssistant(t, server.URL, &assistantToolStub{result: json.RawMessage(`{}`)}, OpenAIConfig{})
	_, err := assistant.Respond(context.Background(), validTestRequest())
	var apiError *APIError
	if !errors.As(err, &apiError) {
		t.Fatalf("expected APIError, got %v", err)
	}
	if apiError.StatusCode != http.StatusTooManyRequests || apiError.Code != "rate_limit_exceeded" || apiError.RequestID != "req_test" {
		t.Fatalf("unexpected API error %#v", apiError)
	}
}

func TestOpenAIAssistantAppliesBoundedRequestTimeout(t *testing.T) {
	t.Parallel()

	assistant, err := NewOpenAIAssistant(OpenAIConfig{
		APIKey:     "test-key",
		HTTPClient: contextBlockingDoer{},
		Timeout:    25 * time.Millisecond,
	}, &assistantToolStub{result: json.RawMessage(`{}`)})
	if err != nil {
		t.Fatalf("construct assistant: %v", err)
	}
	started := time.Now()
	_, err = assistant.Respond(context.Background(), validTestRequest())
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected deadline error, got %v", err)
	}
	if elapsed := time.Since(started); elapsed > time.Second {
		t.Fatalf("assistant did not enforce its timeout; elapsed %s", elapsed)
	}
}

func TestNewOpenAIAssistantRejectsNonStrictOrMutatingTools(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		definition ToolDefinition
	}{
		{
			name: "non strict schema",
			definition: ToolDefinition{
				Type:   "function",
				Name:   toolListTeams,
				Strict: true,
				Parameters: map[string]any{
					"type":                 "object",
					"properties":           map[string]any{"query": map[string]any{"type": "string"}},
					"required":             []string{},
					"additionalProperties": false,
				},
			},
		},
		{
			name: "mutating tool",
			definition: ToolDefinition{
				Type:       "function",
				Name:       "create_story",
				Strict:     true,
				Parameters: strictObjectSchema(map[string]any{}, []string{}),
			},
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			executor := &assistantToolStub{definitions: []ToolDefinition{test.definition}}
			if _, err := NewOpenAIAssistant(OpenAIConfig{APIKey: "test"}, executor); err == nil {
				t.Fatal("expected constructor validation error")
			}
		})
	}
}

func mustTestAssistant(t *testing.T, baseURL string, executor ToolExecutor, overrides OpenAIConfig) *OpenAIAssistant {
	t.Helper()
	config := OpenAIConfig{
		APIKey:          "test-key",
		Model:           "test-model",
		BaseURL:         baseURL,
		Timeout:         2 * time.Second,
		MaxOutputTokens: overrides.MaxOutputTokens,
		Instructions:    "Test instructions.",
	}
	assistant, err := NewOpenAIAssistant(config, executor)
	if err != nil {
		t.Fatalf("construct assistant: %v", err)
	}
	return assistant
}

func validTestRequest() Request {
	return Request{
		WorkspaceID: uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"),
		UserID:      uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"),
		Prompt:      "List my teams",
	}
}

func assertMessageInput(t *testing.T, raw json.RawMessage, role, content string) {
	t.Helper()
	var input messageInput
	if err := json.Unmarshal(raw, &input); err != nil {
		t.Fatalf("decode message input: %v", err)
	}
	if input.Role != role || input.Content != content {
		t.Errorf("unexpected message input %#v", input)
	}
}

func assertFunctionCall(t *testing.T, raw json.RawMessage, callID string) {
	t.Helper()
	var call functionCallOutput
	if err := json.Unmarshal(raw, &call); err != nil {
		t.Fatalf("decode function call: %v", err)
	}
	if call.Type != "function_call" || call.CallID != callID || call.Name != toolListTeams || call.Arguments != "{}" {
		t.Errorf("unexpected function call %#v", call)
	}
}

func assertFunctionOutput(t *testing.T, raw json.RawMessage, callID, output string) {
	t.Helper()
	var item functionCallOutputInput
	if err := json.Unmarshal(raw, &item); err != nil {
		t.Fatalf("decode function output: %v", err)
	}
	if item.Type != "function_call_output" || item.CallID != callID || item.Output != output {
		t.Errorf("unexpected function output %#v", item)
	}
}

type assistantToolStub struct {
	mu          sync.Mutex
	definitions []ToolDefinition
	result      json.RawMessage
	err         error
	calls       []ToolCall
	scopes      []ToolScope
}

type contextBlockingDoer struct{}

func (contextBlockingDoer) Do(request *http.Request) (*http.Response, error) {
	<-request.Context().Done()
	return nil, request.Context().Err()
}

func (s *assistantToolStub) Definitions() []ToolDefinition {
	if s.definitions != nil {
		return cloneToolDefinitions(s.definitions)
	}
	return []ToolDefinition{
		{
			Type:        "function",
			Name:        toolListTeams,
			Description: "List joined teams.",
			Strict:      true,
			Parameters:  strictObjectSchema(map[string]any{}, []string{}),
		},
	}
}

func (s *assistantToolStub) Execute(_ context.Context, scope ToolScope, call ToolCall) (json.RawMessage, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.calls = append(s.calls, call)
	s.scopes = append(s.scopes, scope)
	return cloneRawMessage(s.result), s.err
}

func (s *assistantToolStub) Calls() []ToolCall {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]ToolCall(nil), s.calls...)
}

func (s *assistantToolStub) Scopes() []ToolScope {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]ToolScope(nil), s.scopes...)
}
