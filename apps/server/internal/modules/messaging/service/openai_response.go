package messaging

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Respond executes a bounded Responses API function-call loop. Visible
// conversation state is supplied by the caller and OpenAI storage is disabled.
func (a *OpenAIAssistant) Respond(ctx context.Context, request Request) (Response, error) {
	normalized, err := NormalizeRequest(request)
	if err != nil {
		return Response{}, err
	}
	request = normalized
	ctx, cancel := context.WithTimeout(ctx, a.timeout)
	defer cancel()

	input, err := responseInput(request)
	if err != nil {
		return Response{}, err
	}
	scope := ToolScope{
		WorkspaceID:    request.WorkspaceID,
		UserID:         request.UserID,
		AllowedTeamIDs: cloneOptionalUUIDs(request.AllowedTeamIDs),
		SharedTeamIDs:  cloneOptionalUUIDs(request.SharedTeamIDs),
		AllowMutations: request.AllowMutations,
		WebsiteURL:     request.WebsiteURL,
		SourceURL:      request.SourceURL,
		WorkspaceSlug:  runtimeWorkspaceSlug(request.RuntimeContext),
		Timezone:       runtimeTimezone(request.RuntimeContext),
	}
	safetyIdentifier := safetyIdentifier(request.UserID.String())
	instructions, err := instructionsForRequest(a.instructions, request)
	if err != nil {
		return Response{}, fmt.Errorf("%w: render runtime context: %v", ErrInvalidRequest, err)
	}
	definitions := toolDefinitionsForRequest(a.definitions, request.AllowMutations)
	usage := Usage{}
	toolSteps := 0

	for {
		apiResponse, err := a.createResponse(ctx, input, safetyIdentifier, instructions, definitions)
		if err != nil {
			return Response{Usage: usage}, err
		}
		addUsage(&usage, apiResponse.Usage)
		if err := validateResponseStatus(apiResponse); err != nil {
			return Response{Usage: usage}, err
		}

		analysis, err := analyzeResponseOutput(apiResponse.Output)
		if err != nil {
			return Response{Usage: usage}, err
		}
		if analysis.refusal != "" {
			return Response{Usage: usage}, fmt.Errorf("%w: %s", ErrResponseRefused, analysis.refusal)
		}
		if len(analysis.calls) == 0 {
			text := strings.TrimSpace(analysis.text)
			if text == "" {
				return Response{Usage: usage}, fmt.Errorf("%w: completed response contained no text", ErrMalformedResponse)
			}
			return Response{Text: text, Usage: usage}, nil
		}
		if toolSteps >= maximumToolSteps {
			return Response{Usage: usage}, ErrMaxToolSteps
		}
		if containsStoryMutationCall(analysis.calls) && len(analysis.calls) != 1 {
			return Response{Usage: usage}, fmt.Errorf("%w: a story mutation proposal must be the only tool call in a response", ErrMalformedResponse)
		}

		// Responses output items are valid subsequent input items. Preserve every
		// heterogeneous item, including encrypted reasoning, before tool outputs.
		input = append(input, cloneRawMessages(apiResponse.Output)...)
		for _, call := range analysis.calls {
			if _, ok := a.toolNames[call.Name]; !ok {
				return Response{Usage: usage}, fmt.Errorf("%w: %s", ErrUnknownTool, call.Name)
			}
			if isStoryMutationTool(call.Name) && !request.AllowMutations {
				return Response{Usage: usage}, fmt.Errorf("%w: %s is disabled for this request", ErrUnknownTool, call.Name)
			}
			result, err := a.tools.Execute(ctx, scope, call)
			if err != nil {
				return Response{Usage: usage}, fmt.Errorf("%w: %s: %w", ErrToolExecution, call.Name, err)
			}
			if len(result) == 0 || !json.Valid(result) {
				return Response{Usage: usage}, fmt.Errorf("%w: tool %s returned invalid JSON", ErrMalformedResponse, call.Name)
			}
			if len(result) > maximumToolOutputBytes {
				return Response{Usage: usage}, fmt.Errorf("%w: tool %s output exceeds %d bytes", ErrMalformedResponse, call.Name, maximumToolOutputBytes)
			}
			if isStoryMutationTool(call.Name) {
				confirmation, proposed, err := mutationConfirmationFromToolResult(result)
				if err != nil {
					return Response{Usage: usage}, err
				}
				if !proposed {
					return Response{Usage: usage}, fmt.Errorf("%w: mutation tool %s did not return a confirmation proposal", ErrMalformedResponse, call.Name)
				}
				return Response{
					Text:         confirmation.Prompt,
					Usage:        usage,
					Confirmation: confirmation,
				}, nil
			}
			output, err := json.Marshal(functionCallOutputInput{
				Type:   "function_call_output",
				CallID: call.ID,
				Output: string(result),
			})
			if err != nil {
				return Response{Usage: usage}, fmt.Errorf("encode function output: %w", err)
			}
			input = append(input, output)
		}
		toolSteps++
	}
}

func (a *OpenAIAssistant) createResponse(
	ctx context.Context,
	input []json.RawMessage,
	safetyID string,
	instructions string,
	definitions []ToolDefinition,
) (responsesAPIResponse, error) {
	payload := responsesAPIRequest{
		Model:             a.model,
		Instructions:      instructions,
		Input:             input,
		Tools:             definitions,
		Store:             false,
		Reasoning:         responsesReasoningForModel(a.model),
		MaxOutputTokens:   a.maxOutputTokens,
		ParallelToolCalls: false,
		SafetyIdentifier:  safetyID,
		Include:           []string{"reasoning.encrypted_content"},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return responsesAPIResponse{}, fmt.Errorf("encode OpenAI request: %w", err)
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, a.endpoint, bytes.NewReader(body))
	if err != nil {
		return responsesAPIResponse{}, fmt.Errorf("create OpenAI request: %w", err)
	}
	request.Header.Set("Authorization", "Bearer "+a.apiKey)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")
	request.Header.Set("User-Agent", "fortyone-messaging-assistant/1.0")

	response, err := a.httpClient.Do(request)
	if err != nil {
		return responsesAPIResponse{}, fmt.Errorf("call OpenAI Responses API: %w", err)
	}
	if response == nil || response.Body == nil {
		return responsesAPIResponse{}, fmt.Errorf("%w: OpenAI returned an empty HTTP response", ErrMalformedResponse)
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(io.LimitReader(response.Body, maximumResponseBodyBytes+1))
	if err != nil {
		return responsesAPIResponse{}, fmt.Errorf("read OpenAI response: %w", err)
	}
	if len(responseBody) > maximumResponseBodyBytes {
		return responsesAPIResponse{}, fmt.Errorf("%w: OpenAI response exceeds %d bytes", ErrMalformedResponse, maximumResponseBodyBytes)
	}
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return responsesAPIResponse{}, decodeAPIError(response.StatusCode, response.Header.Get("x-request-id"), responseBody)
	}

	var decoded responsesAPIResponse
	if err := json.Unmarshal(responseBody, &decoded); err != nil {
		return responsesAPIResponse{}, fmt.Errorf("%w: decode OpenAI response: %v", ErrMalformedResponse, err)
	}
	return decoded, nil
}
