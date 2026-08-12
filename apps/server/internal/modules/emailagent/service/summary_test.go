package emailagent

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestOpenAISummarizerIncrementallyPreservesDurableConversationState(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		var payload map[string]any
		require.NoError(t, json.NewDecoder(request.Body).Decode(&payload))
		require.Equal(t, "gpt-5.6-luna", payload["model"])
		require.Equal(t, false, payload["store"])
		reasoning := payload["reasoning"].(map[string]any)
		require.Equal(t, "low", reasoning["effort"])
		require.Equal(t, "current_turn", reasoning["context"])
		safetyIdentifier := payload["safety_identifier"].(string)
		require.True(t, strings.HasPrefix(safetyIdentifier, "maya_email_"))
		require.NotContains(t, safetyIdentifier, "joseph@example.com")
		require.Equal(t, safetyIdentifierForIdentity("joseph@example.com"), safetyIdentifier)
		format := payload["text"].(map[string]any)["format"].(map[string]any)
		require.Equal(t, true, format["strict"])

		input := payload["input"].([]any)
		developerText := input[0].(map[string]any)["content"].([]any)[0].(map[string]any)["text"].(string)
		userText := input[1].(map[string]any)["content"].([]any)[0].(map[string]any)["text"].(string)
		require.Contains(t, developerText, "user corrections")
		require.Contains(t, developerText, "open proposal")
		require.Contains(t, developerText, "untrusted data")
		require.Contains(t, userText, `"previousSummary":"The objective was At Risk."`)
		require.Contains(t, userText, "It is now On Track")

		writeSummaryResponse(t, response, "The user corrected the objective health from At Risk to On Track. An update to On Track is proposed and awaits CONFIRM or CANCEL.", Usage{
			InputTokens:  80,
			OutputTokens: 20,
			TotalTokens:  100,
		})
	}))
	defer server.Close()

	summarizer, err := NewOpenAISummarizer(OpenAIConfig{
		APIKey:  "test-key",
		BaseURL: server.URL,
	})
	require.NoError(t, err)
	require.True(t, summarizer.Enabled())

	generation, err := summarizer.Summarize(context.Background(), SummaryRequest{
		SafetyIdentifier: "joseph@example.com",
		PreviousSummary:  "The objective was At Risk.",
		OmittedTurns: []HistoryTurn{
			{Role: RoleUser, Text: "It is now On Track, not At Risk.", SentAt: time.Now()},
			{Role: RoleAssistant, Text: "I propose updating it to On Track. Reply CONFIRM or CANCEL.", SentAt: time.Now()},
		},
	})

	require.NoError(t, err)
	require.Contains(t, generation.Summary, "corrected")
	require.Contains(t, generation.Summary, "awaits CONFIRM or CANCEL")
	require.Equal(t, 100, generation.Usage.TotalTokens)
}

func TestOpenAISummarizerValidatesInputBeforeProviderCall(t *testing.T) {
	t.Parallel()

	called := false
	summarizer, err := NewOpenAISummarizer(OpenAIConfig{
		APIKey: "test-key",
		HTTPClient: httpDoerFunc(func(*http.Request) (*http.Response, error) {
			called = true
			return nil, nil
		}),
	})
	require.NoError(t, err)

	_, err = summarizer.Summarize(context.Background(), SummaryRequest{
		SafetyIdentifier: "user",
		OmittedTurns:     nil,
	})

	require.ErrorIs(t, err, ErrInvalidRequest)
	require.False(t, called)
}

func TestOpenAISummarizerRejectsUnsafeMalformedOrOverlongOutput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		output string
	}{
		{name: "empty", output: ""},
		{name: "markdown heading", output: "# Conversation summary"},
		{name: "markdown list", output: "- User corrected the date."},
		{name: "HTML", output: "<b>User corrected the date.</b>"},
		{name: "markdown link", output: "The task is at [FortyOne](https://example.com)."},
		{name: "overlong", output: strings.Repeat("x", maxGeneratedSummaryRunes+1)},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
				writeSummaryResponse(t, response, test.output, Usage{InputTokens: 1, OutputTokens: 1, TotalTokens: 2})
			}))
			defer server.Close()
			summarizer, err := NewOpenAISummarizer(OpenAIConfig{APIKey: "key", BaseURL: server.URL})
			require.NoError(t, err)

			_, err = summarizer.Summarize(context.Background(), validSummaryRequest())

			require.ErrorIs(t, err, ErrMalformedGeneration)
		})
	}
}

func TestOpenAISummarizerRejectsUnknownJSONAndRefusals(t *testing.T) {
	t.Parallel()

	t.Run("unknown field", func(t *testing.T) {
		t.Parallel()
		server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
			output := `{"summary":"Useful summary.","extra":true}`
			_, _ = fmt.Fprintf(response, `{"status":"completed","output":[{"type":"message","content":[{"type":"output_text","text":%q}]}],"usage":{"input_tokens":1,"output_tokens":1,"total_tokens":2}}`, output)
		}))
		defer server.Close()
		summarizer, err := NewOpenAISummarizer(OpenAIConfig{APIKey: "key", BaseURL: server.URL})
		require.NoError(t, err)

		_, err = summarizer.Summarize(context.Background(), validSummaryRequest())
		require.ErrorIs(t, err, ErrMalformedGeneration)
		require.ErrorContains(t, err, "unknown field")
	})

	t.Run("refusal", func(t *testing.T) {
		t.Parallel()
		server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
			_, _ = response.Write([]byte(`{"status":"completed","output":[{"type":"message","content":[{"type":"refusal","refusal":"No"}]}],"usage":{"input_tokens":1,"output_tokens":1,"total_tokens":2}}`))
		}))
		defer server.Close()
		summarizer, err := NewOpenAISummarizer(OpenAIConfig{APIKey: "key", BaseURL: server.URL})
		require.NoError(t, err)

		_, err = summarizer.Summarize(context.Background(), validSummaryRequest())
		require.ErrorIs(t, err, ErrGeneratorRefused)
	})
}

func TestSummaryResponseFormatIsStrictAndBounded(t *testing.T) {
	t.Parallel()

	format := SummaryResponseFormat()
	require.Equal(t, true, format["strict"])
	schema := format["schema"].(map[string]any)
	require.NoError(t, assertStrictSchema(schema, "schema"))
	summary := schema["properties"].(map[string]any)["summary"].(map[string]any)
	require.Equal(t, maxGeneratedSummaryRunes, summary["maxLength"])
}

func TestGeneratedSummaryAllowsAPlainURLWhenItIsDurableData(t *testing.T) {
	t.Parallel()

	err := validateGeneratedSummary("The unresolved request refers to https://example.com/ticket/42.")

	require.NoError(t, err)
}

type httpDoerFunc func(*http.Request) (*http.Response, error)

func (fn httpDoerFunc) Do(request *http.Request) (*http.Response, error) {
	return fn(request)
}

func validSummaryRequest() SummaryRequest {
	return SummaryRequest{
		SafetyIdentifier: "user-123",
		PreviousSummary:  "The objective is At Risk.",
		OmittedTurns: []HistoryTurn{
			{Role: RoleUser, Text: "It is now On Track."},
		},
	}
}

func writeSummaryResponse(t *testing.T, response http.ResponseWriter, summary string, usage Usage) {
	t.Helper()
	outputJSON, err := json.Marshal(summaryOutput{Summary: summary})
	require.NoError(t, err)
	payload := map[string]any{
		"status": "completed",
		"output": []any{
			map[string]any{"type": "reasoning", "encrypted_content": "opaque"},
			map[string]any{
				"type": "message",
				"content": []any{map[string]any{
					"type": "output_text",
					"text": string(outputJSON),
				}},
			},
		},
		"usage": map[string]any{
			"input_tokens":  usage.InputTokens,
			"output_tokens": usage.OutputTokens,
			"total_tokens":  usage.TotalTokens,
		},
	}
	response.Header().Set("Content-Type", "application/json")
	require.NoError(t, json.NewEncoder(response).Encode(payload))
}
