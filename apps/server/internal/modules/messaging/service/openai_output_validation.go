package messaging

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

func validateResponseStatus(response responsesAPIResponse) error {
	switch response.Status {
	case "completed":
		if response.Error != nil {
			return fmt.Errorf("%w: completed response included error %s", ErrMalformedResponse, response.Error.message())
		}
		return nil
	case "incomplete":
		reason := "unknown reason"
		if response.IncompleteDetails != nil && strings.TrimSpace(response.IncompleteDetails.Reason) != "" {
			reason = response.IncompleteDetails.Reason
		}
		return fmt.Errorf("%w: %s", ErrResponseIncomplete, reason)
	case "failed", "cancelled":
		message := response.Status
		if response.Error != nil && response.Error.message() != "" {
			message = response.Error.message()
		}
		return fmt.Errorf("%w: %s", ErrResponseFailed, message)
	case "queued", "in_progress":
		return fmt.Errorf("%w: unexpected synchronous status %q", ErrResponseFailed, response.Status)
	case "":
		return fmt.Errorf("%w: response status is missing", ErrMalformedResponse)
	default:
		return fmt.Errorf("%w: unknown response status %q", ErrMalformedResponse, response.Status)
	}
}

func analyzeResponseOutput(output []json.RawMessage) (outputAnalysis, error) {
	if len(output) == 0 {
		return outputAnalysis{}, fmt.Errorf("%w: response output is empty", ErrMalformedResponse)
	}
	var analysis outputAnalysis
	for index, raw := range output {
		var header struct {
			Type string `json:"type"`
		}
		if err := json.Unmarshal(raw, &header); err != nil || header.Type == "" {
			return outputAnalysis{}, fmt.Errorf("%w: output item %d has no valid type", ErrMalformedResponse, index)
		}
		switch header.Type {
		case "function_call":
			var item functionCallOutput
			if err := json.Unmarshal(raw, &item); err != nil {
				return outputAnalysis{}, fmt.Errorf("%w: decode function call: %v", ErrMalformedResponse, err)
			}
			if strings.TrimSpace(item.CallID) == "" || strings.TrimSpace(item.Name) == "" {
				return outputAnalysis{}, fmt.Errorf("%w: function call is missing call_id or name", ErrMalformedResponse)
			}
			arguments := json.RawMessage(item.Arguments)
			var object map[string]json.RawMessage
			if !json.Valid(arguments) || json.Unmarshal(arguments, &object) != nil || object == nil {
				return outputAnalysis{}, fmt.Errorf("%w: function %s arguments are not a JSON object", ErrMalformedResponse, item.Name)
			}
			analysis.calls = append(analysis.calls, ToolCall{
				ID:        item.CallID,
				Name:      item.Name,
				Arguments: cloneRawMessage(arguments),
			})
		case "message":
			var item messageOutput
			if err := json.Unmarshal(raw, &item); err != nil {
				return outputAnalysis{}, fmt.Errorf("%w: decode message output: %v", ErrMalformedResponse, err)
			}
			for contentIndex, contentRaw := range item.Content {
				var content outputContent
				if err := json.Unmarshal(contentRaw, &content); err != nil || content.Type == "" {
					return outputAnalysis{}, fmt.Errorf("%w: message content %d has no valid type", ErrMalformedResponse, contentIndex)
				}
				switch content.Type {
				case "output_text":
					analysis.text += content.Text
				case "refusal":
					if analysis.refusal != "" && content.Refusal != "" {
						analysis.refusal += " "
					}
					analysis.refusal += strings.TrimSpace(content.Refusal)
				}
			}
		}
	}
	return analysis, nil
}

func validateToolDefinitions(definitions []ToolDefinition) error {
	if len(definitions) == 0 || len(definitions) > len(allowedToolNames) {
		return fmt.Errorf("assistant must expose between 1 and %d approved tools", len(allowedToolNames))
	}
	seen := make(map[string]struct{}, len(definitions))
	for _, definition := range definitions {
		if definition.Type != "function" || !definition.Strict {
			return fmt.Errorf("assistant tool %q must be a strict function", definition.Name)
		}
		if _, allowed := allowedToolNames[definition.Name]; !allowed {
			return fmt.Errorf("assistant tool %q is not in the approved catalog", definition.Name)
		}
		if _, duplicate := seen[definition.Name]; duplicate {
			return fmt.Errorf("assistant tool %q is duplicated", definition.Name)
		}
		seen[definition.Name] = struct{}{}
		if err := validateStrictObjectSchema(definition.Parameters, definition.Name); err != nil {
			return err
		}
	}
	return nil
}

func containsStoryMutationCall(calls []ToolCall) bool {
	for _, call := range calls {
		if isStoryMutationTool(call.Name) {
			return true
		}
	}
	return false
}

func isStoryMutationTool(name string) bool {
	return name == toolCreateStory || name == toolCreateStories || name == toolUpdateStory || name == toolAddComment || name == toolAddRelationship
}

func validateStrictObjectSchema(schema map[string]any, path string) error {
	if schema["type"] != "object" {
		return fmt.Errorf("assistant tool schema %s must have object type", path)
	}
	additionalProperties, ok := schema["additionalProperties"].(bool)
	if !ok || additionalProperties {
		return fmt.Errorf("assistant tool schema %s must set additionalProperties to false", path)
	}
	properties, ok := schema["properties"].(map[string]any)
	if !ok {
		return fmt.Errorf("assistant tool schema %s must define properties", path)
	}
	required, err := requiredSet(schema["required"])
	if err != nil {
		return fmt.Errorf("assistant tool schema %s: %w", path, err)
	}
	if len(required) != len(properties) {
		return fmt.Errorf("assistant tool schema %s must require every property", path)
	}
	for name, value := range properties {
		if _, ok := required[name]; !ok {
			return fmt.Errorf("assistant tool schema %s must require property %s", path, name)
		}
		property, ok := value.(map[string]any)
		if !ok {
			return fmt.Errorf("assistant tool schema %s.%s must be an object", path, name)
		}
		if property["type"] == "object" {
			if err := validateStrictObjectSchema(property, path+"."+name); err != nil {
				return err
			}
		}
		if property["type"] == "array" {
			if items, ok := property["items"].(map[string]any); ok && items["type"] == "object" {
				if err := validateStrictObjectSchema(items, path+"."+name+"[]"); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func requiredSet(value any) (map[string]struct{}, error) {
	result := map[string]struct{}{}
	switch required := value.(type) {
	case []string:
		if required == nil {
			return nil, errors.New("required must be a non-null array")
		}
		for _, name := range required {
			result[name] = struct{}{}
		}
	case []any:
		if required == nil {
			return nil, errors.New("required must be a non-null array")
		}
		for _, value := range required {
			name, ok := value.(string)
			if !ok {
				return nil, errors.New("required entries must be strings")
			}
			result[name] = struct{}{}
		}
	default:
		return nil, errors.New("required must be an array")
	}
	return result, nil
}
