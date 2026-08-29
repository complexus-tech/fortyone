package emailagent

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOpenAIGeneratorUsesLunaLowStrictStatelessRequest(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		require.Equal(t, "/v1/responses", request.URL.Path)
		require.Equal(t, "Bearer test-key", request.Header.Get("Authorization"))
		require.Equal(t, "fortyone-email-agent/1.0", request.Header.Get("User-Agent"))

		var payload map[string]any
		require.NoError(t, json.NewDecoder(request.Body).Decode(&payload))
		require.Equal(t, "gpt-5.6-luna", payload["model"])
		require.Equal(t, false, payload["store"])
		require.Equal(t, "maya_email_012345", payload["safety_identifier"])
		reasoning := payload["reasoning"].(map[string]any)
		require.Equal(t, "low", reasoning["effort"])
		require.Equal(t, "current_turn", reasoning["context"])
		format := payload["text"].(map[string]any)["format"].(map[string]any)
		require.Equal(t, true, format["strict"])
		require.Equal(t, "json_schema", format["type"])

		input := payload["input"].([]any)
		require.Len(t, input, 2)
		developerText := input[0].(map[string]any)["content"].([]any)[0].(map[string]any)["text"].(string)
		userText := input[1].(map[string]any)["content"].([]any)[0].(map[string]any)["text"].(string)
		require.Contains(t, developerText, "untrusted data")
		require.Contains(t, developerText, "Never return confirm or cancel")
		require.Contains(t, userText, "untrusted conversation and product data")
		require.Contains(t, userText, `"message":"Please update it"`)
		require.NotContains(t, userText, "maya_email_012345")

		writeOpenAIResponse(t, response, answerDecision("A useful answer", "I can help with that.", ""), Usage{
			InputTokens:  120,
			OutputTokens: 30,
			TotalTokens:  150,
		})
	}))
	defer server.Close()

	generator, err := NewOpenAIGenerator(OpenAIConfig{
		APIKey:  "test-key",
		BaseURL: server.URL + "/v1",
	})
	require.NoError(t, err)
	require.True(t, generator.Enabled())

	generation, err := generator.Generate(context.Background(), ModelRequest{
		SafetyIdentifier: "maya_email_012345",
		Subject:          "An update",
		Message:          "Please update it",
		History:          HistoryWindow{Turns: []HistoryTurn{}},
		Facts:            []GroundedFact{},
		Targets:          []ModelTarget{},
		Choices:          []ModelChoice{},
	})

	require.NoError(t, err)
	require.Equal(t, IntentAnswer, generation.Decision.Intent)
	require.Equal(t, 150, generation.Usage.TotalTokens)
}

func TestOpenAIGeneratorReturnsSanitizedAPIError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
		response.Header().Set("x-request-id", "req_123")
		response.WriteHeader(http.StatusTooManyRequests)
		_, _ = response.Write([]byte(`{"error":{"code":"rate_limit_exceeded","message":"secret provider detail"}}`))
	}))
	defer server.Close()
	generator, err := NewOpenAIGenerator(OpenAIConfig{APIKey: "test-key", BaseURL: server.URL})
	require.NoError(t, err)

	_, err = generator.Generate(context.Background(), validDirectModelRequest())

	var apiError *OpenAIAPIError
	require.ErrorAs(t, err, &apiError)
	require.Equal(t, http.StatusTooManyRequests, apiError.StatusCode)
	require.Equal(t, "rate_limit_exceeded", apiError.Code)
	require.Equal(t, "req_123", apiError.RequestID)
	require.NotContains(t, err.Error(), "secret provider detail")
}

func TestOpenAIGeneratorRejectsRefusalAndUnknownStructuredFields(t *testing.T) {
	t.Parallel()

	t.Run("refusal", func(t *testing.T) {
		t.Parallel()
		server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
			response.Header().Set("Content-Type", "application/json")
			_, _ = response.Write([]byte(`{
				"status":"completed",
				"output":[{"type":"message","content":[{"type":"refusal","refusal":"No"}]}],
				"usage":{"input_tokens":1,"output_tokens":1,"total_tokens":2}
			}`))
		}))
		defer server.Close()
		generator, err := NewOpenAIGenerator(OpenAIConfig{APIKey: "test-key", BaseURL: server.URL})
		require.NoError(t, err)

		_, err = generator.Generate(context.Background(), validDirectModelRequest())
		require.ErrorIs(t, err, ErrGeneratorRefused)
	})

	t.Run("unknown structured field", func(t *testing.T) {
		t.Parallel()
		server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
			output := `{"intent":"answer","copy":{"subject":{"text":"Answer","references":[]},"blocks":[{"kind":"paragraph","text":"Useful answer","items":[],"references":[]}]},"proposal":null,"unexpected":true}`
			response.Header().Set("Content-Type", "application/json")
			_, _ = fmt.Fprintf(response, `{"status":"completed","output":[{"type":"message","content":[{"type":"output_text","text":%q}]}],"usage":{"input_tokens":1,"output_tokens":1,"total_tokens":2}}`, output)
		}))
		defer server.Close()
		generator, err := NewOpenAIGenerator(OpenAIConfig{APIKey: "test-key", BaseURL: server.URL})
		require.NoError(t, err)

		_, err = generator.Generate(context.Background(), validDirectModelRequest())
		require.ErrorIs(t, err, ErrMalformedGeneration)
		require.ErrorContains(t, err, "unknown field")
	})
}

func TestResponseFormatIsStrictAndExcludesModelControlledConfirmationOrBatchFeedback(t *testing.T) {
	t.Parallel()

	format := ResponseFormat()
	require.Equal(t, true, format["strict"])
	schema := format["schema"].(map[string]any)
	require.NoError(t, assertStrictSchema(schema, "schema"))

	properties := schema["properties"].(map[string]any)
	intentValues := properties["intent"].(map[string]any)["enum"].([]string)
	require.ElementsMatch(t, []string{"answer", "clarify", "propose", "refuse"}, intentValues)
	require.NotContains(t, intentValues, "confirm")
	require.NotContains(t, intentValues, "cancel")

	proposalAnyOf := properties["proposal"].(map[string]any)["anyOf"].([]any)
	proposal := proposalAnyOf[0].(map[string]any)
	feedbackAnyOf := proposal["properties"].(map[string]any)["feedback"].(map[string]any)["anyOf"].([]any)
	feedback := feedbackAnyOf[0].(map[string]any)
	feedbackProperties := feedback["properties"].(map[string]any)
	require.Contains(t, feedbackProperties, "targetReference")
	require.NotContains(t, feedbackProperties, "targetReferences")
}

func validDirectModelRequest() ModelRequest {
	return ModelRequest{
		SafetyIdentifier: "maya_email_valid",
		Message:          "Help me",
		History:          HistoryWindow{Turns: []HistoryTurn{}},
		Facts:            []GroundedFact{},
		Targets:          []ModelTarget{},
		Choices:          []ModelChoice{},
	}
}

func writeOpenAIResponse(t *testing.T, response http.ResponseWriter, decision ModelDecision, usage Usage) {
	t.Helper()
	decisionJSON, err := json.Marshal(decision)
	require.NoError(t, err)
	payload := map[string]any{
		"status": "completed",
		"output": []any{
			map[string]any{"type": "reasoning", "encrypted_content": "opaque"},
			map[string]any{
				"type": "message",
				"content": []any{map[string]any{
					"type": "output_text",
					"text": string(decisionJSON),
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

func assertStrictSchema(schema map[string]any, path string) error {
	if schemaType, _ := schema["type"].(string); schemaType == "object" {
		if schema["additionalProperties"] != false {
			return fmt.Errorf("%s is not strict", path)
		}
		properties, _ := schema["properties"].(map[string]any)
		required, _ := schema["required"].([]string)
		if len(properties) != len(required) {
			return fmt.Errorf("%s does not require every property", path)
		}
		for name, raw := range properties {
			property, ok := raw.(map[string]any)
			if !ok {
				continue
			}
			if err := assertStrictSchema(property, path+"."+name); err != nil {
				return err
			}
		}
	}
	if items, ok := schema["items"].(map[string]any); ok {
		if err := assertStrictSchema(items, path+"[]"); err != nil {
			return err
		}
	}
	if alternatives, ok := schema["anyOf"].([]any); ok {
		for index, raw := range alternatives {
			alternative, ok := raw.(map[string]any)
			if !ok {
				continue
			}
			if err := assertStrictSchema(alternative, fmt.Sprintf("%s.anyOf[%d]", path, index)); err != nil {
				return err
			}
		}
	}
	return nil
}

func TestOpenAIConfigRejectsCredentialBearingBaseURL(t *testing.T) {
	t.Parallel()

	_, err := NewOpenAIGenerator(OpenAIConfig{BaseURL: "https://user:secret@example.com/v1"})
	require.Error(t, err)
	_, err = NewOpenAIGenerator(OpenAIConfig{BaseURL: "javascript:alert(1)"})
	require.Error(t, err)
	_, err = NewOpenAIGenerator(OpenAIConfig{BaseURL: "http://api.openai.com/v1"})
	require.Error(t, err)

	disabled, err := NewOpenAIGenerator(OpenAIConfig{})
	require.NoError(t, err)
	require.False(t, disabled.Enabled())
	_, err = disabled.Generate(context.Background(), validDirectModelRequest())
	require.ErrorIs(t, err, ErrGeneratorNotConfigured)
}

func TestStableSafetyIdentifierDoesNotExposeRawIdentity(t *testing.T) {
	t.Parallel()

	request := baseRequest()
	request.SafetyIdentifier = "sensitive-user@example.com"
	first := stableSafetyIdentifier(request)
	second := stableSafetyIdentifier(request)

	require.Equal(t, first, second)
	require.True(t, strings.HasPrefix(first, "maya_email_"))
	require.NotContains(t, first, request.SafetyIdentifier)
}
