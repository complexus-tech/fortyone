package chatsessionsdomain

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"unicode"
)

const (
	MutationApprovalUncertainOutputMessage = "Maya could not verify whether this approved change finished. Check the workspace before trying it again; an identical change is blocked until this execution is reconciled."
	MutationApprovalSkippedOutputMessage   = "Maya did not run this approved change because an earlier approved change was unresolved. Review the earlier result, then ask Maya to prepare this change again."
)

var retryableStoryMutationTools = map[string]struct{}{
	"bulkCreateStories": {},
	"bulkDeleteStories": {},
	"createStory":       {},
	"deleteStory":       {},
}

var (
	storyContentInputKeys = []string{
		"title", "description", "descriptionHTML",
	}
	storyPlanningInputKeys = []string{
		"assigneeId", "priority", "estimateValue",
		"estimatedDurationMinutes", "minimumFocusBlockMinutes",
		"autoSchedulingEnabled", "labelIds", "sprintId", "objectiveId",
		"keyResultId", "parentId", "startDate", "endDate",
	}
	createStoryInputKeys = append(
		append(append([]string{}, storyContentInputKeys...), "teamId", "statusId"),
		storyPlanningInputKeys...,
	)
	bulkStoryInputKeys = append(
		append(append([]string{}, storyContentInputKeys...), "teamId", "statusId"),
		storyPlanningInputKeys...,
	)
	bulkStorySharedInputKeys = append(
		[]string{"teamId", "statusId"},
		storyPlanningInputKeys...,
	)
)

type MutationApprovalRetryIntent struct {
	Fingerprint string
	ToolCallID  string
	ToolName    string
}

type MutationApprovalRetryPreparer func(intent MutationApprovalRetryIntent) (bool, error)

// PrepareMutationApprovalRetries reopens only an exact, server-verifiable
// retry of a supported idempotent story mutation. The caller performs the
// durable ledger transition in the supplied callback so transcript and ledger
// changes remain in one transaction.
func PrepareMutationApprovalRetries(
	current []any,
	incoming []any,
	prepare MutationApprovalRetryPreparer,
) ([]any, error) {
	if len(current) == 0 || len(current) != len(incoming) || prepare == nil {
		return current, nil
	}

	reopenedValue, err := cloneJSONValue(current)
	if err != nil {
		return nil, err
	}
	reopened, ok := reopenedValue.([]any)
	if !ok {
		return nil, fmt.Errorf("%w: transcript must be an array", ErrMessageWriteInvalid)
	}

	currentMessage, currentOK := asObject(reopened[len(reopened)-1])
	incomingMessage, incomingOK := asObject(incoming[len(incoming)-1])
	if !currentOK || !incomingOK || messageRole(currentMessage) != "assistant" || !sameMessageIdentity(currentMessage, incomingMessage) {
		return reopened, nil
	}
	currentParts, currentOK := asArray(currentMessage["parts"])
	incomingParts, incomingOK := asArray(incomingMessage["parts"])
	if !currentOK || !incomingOK || len(currentParts) != len(incomingParts) {
		return reopened, nil
	}

	for index := range currentParts {
		currentPart, currentObject := asObject(currentParts[index])
		incomingPart, incomingObject := asObject(incomingParts[index])
		if !currentObject || !incomingObject || !sameToolPartIdentity(currentPart, incomingPart) || !sameToolPartBase(currentPart, incomingPart) {
			continue
		}
		if toolState(incomingPart) != "approval-responded" || !approvalGranted(incomingPart["approval"]) || !validApprovalTransition(currentPart["approval"], incomingPart["approval"]) {
			continue
		}

		toolName, ok := retryableStoryMutationToolName(currentPart)
		if !ok || !approvalGranted(currentPart["approval"]) || !retryableMutationOutput(currentPart) {
			continue
		}
		fingerprint, err := ApprovedMutationFingerprint(toolName, currentPart["input"])
		if err != nil {
			return nil, err
		}
		toolCallID := currentPart["toolCallId"].(string)
		prepared, err := prepare(MutationApprovalRetryIntent{
			Fingerprint: fingerprint,
			ToolCallID:  toolCallID,
			ToolName:    toolName,
		})
		if err != nil {
			return nil, err
		}
		if !prepared {
			continue
		}

		reopenedPart, err := cloneObject(currentPart)
		if err != nil {
			return nil, err
		}
		reopenedPart["state"] = "approval-responded"
		delete(reopenedPart, "output")
		delete(reopenedPart, "errorText")
		currentParts[index] = reopenedPart
	}

	return reopened, nil
}

func approvalGranted(raw any) bool {
	approval, ok := asObject(raw)
	if !ok {
		return false
	}
	approved, ok := approval["approved"].(bool)
	return ok && approved
}

func retryableStoryMutationToolName(part map[string]any) (string, bool) {
	partType, ok := part["type"].(string)
	if !ok || !strings.HasPrefix(partType, "tool-") {
		return "", false
	}
	toolName := strings.TrimPrefix(partType, "tool-")
	if _, ok := retryableStoryMutationTools[toolName]; !ok {
		return "", false
	}
	if toolName == "bulkDeleteStories" {
		input, ok := asObject(part["input"])
		if !ok {
			return "", false
		}
		if hardDelete, _ := input["hardDelete"].(bool); hardDelete {
			return "", false
		}
	}
	return toolName, true
}

func retryableMutationOutput(part map[string]any) bool {
	switch toolState(part) {
	case "approval-responded":
		_, hasOutput := part["output"]
		return !hasOutput
	case "output-available":
		output, ok := asObject(part["output"])
		if !ok || len(output) != 2 {
			return false
		}
		errorMessage, errorOK := output["error"].(string)
		success, successOK := output["success"].(bool)
		return errorOK && successOK && !success && errorMessage == MutationApprovalUncertainOutputMessage
	default:
		return false
	}
}

func ApprovedMutationFingerprint(toolName string, input any) (string, error) {
	approvedInput, err := approvedMutationInput(toolName, input)
	if err != nil {
		return "", err
	}
	payload := map[string]any{
		"approved": true,
		"input":    approvedInput,
		"toolName": toolName,
	}
	var canonical bytes.Buffer
	if err := writeCanonicalJavaScriptJSON(&canonical, payload); err != nil {
		return "", fmt.Errorf("canonicalize approved mutation: %w", err)
	}
	digest := sha256.Sum256(canonical.Bytes())
	return hex.EncodeToString(digest[:]), nil
}

func approvedMutationInput(toolName string, input any) (any, error) {
	cloned, err := cloneJSONValue(input)
	if err != nil {
		return nil, err
	}
	object, ok := cloned.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("%w: approved mutation input must be an object", ErrMessageWriteInvalid)
	}

	switch toolName {
	case "createStory":
		return normalizeCreateStoryFingerprintInput(object)
	case "bulkCreateStories":
		return normalizeBulkCreateStoriesFingerprintInput(object)
	case "deleteStory":
		return normalizeDeleteStoryFingerprintInput(object)
	case "bulkDeleteStories":
		return normalizeBulkDeleteStoriesFingerprintInput(object)
	default:
		return nil, fmt.Errorf("%w: unsupported retryable mutation %q", ErrMessageWriteInvalid, toolName)
	}
}

func normalizeCreateStoryFingerprintInput(input map[string]any) (map[string]any, error) {
	result := projectObject(input, createStoryInputKeys)
	if _, ok := result["title"].(string); !ok {
		return nil, fmt.Errorf("%w: createStory title is invalid", ErrMessageWriteInvalid)
	}
	if _, ok := result["teamId"].(string); !ok {
		return nil, fmt.Errorf("%w: createStory teamId is invalid", ErrMessageWriteInvalid)
	}
	if priority, exists := result["priority"]; !exists || priority == nil {
		result["priority"] = "No Priority"
	}
	deleteNullProperty(result, "autoSchedulingEnabled")
	return result, nil
}

func normalizeBulkCreateStoriesFingerprintInput(input map[string]any) (map[string]any, error) {
	result := projectObject(input, []string{"sharedValues", "storiesData"})
	if shared, exists := result["sharedValues"]; exists {
		if shared == nil {
			delete(result, "sharedValues")
		} else {
			sharedObject, ok := shared.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("%w: bulkCreateStories sharedValues is invalid", ErrMessageWriteInvalid)
			}
			normalizedShared := projectObject(sharedObject, bulkStorySharedInputKeys)
			deleteNullProperty(normalizedShared, "teamId")
			deleteNullProperty(normalizedShared, "priority")
			deleteNullProperty(normalizedShared, "autoSchedulingEnabled")
			result["sharedValues"] = normalizedShared
		}
	}

	stories, ok := result["storiesData"].([]any)
	if !ok || len(stories) == 0 {
		return nil, fmt.Errorf("%w: bulkCreateStories storiesData is invalid", ErrMessageWriteInvalid)
	}
	for index, rawStory := range stories {
		story, ok := rawStory.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("%w: bulkCreateStories story %d is invalid", ErrMessageWriteInvalid, index)
		}
		normalizedStory := projectObject(story, bulkStoryInputKeys)
		if _, ok := normalizedStory["title"].(string); !ok {
			return nil, fmt.Errorf("%w: bulkCreateStories story %d title is invalid", ErrMessageWriteInvalid, index)
		}
		deleteNullProperty(normalizedStory, "priority")
		deleteNullProperty(normalizedStory, "autoSchedulingEnabled")
		stories[index] = normalizedStory
	}
	result["storiesData"] = stories
	return result, nil
}

func normalizeDeleteStoryFingerprintInput(input map[string]any) (map[string]any, error) {
	result := projectObject(input, []string{"storyId", "storyTitle"})
	if _, ok := result["storyId"].(string); !ok {
		return nil, fmt.Errorf("%w: deleteStory storyId is invalid", ErrMessageWriteInvalid)
	}
	title, ok := result["storyTitle"].(string)
	if !ok {
		return nil, fmt.Errorf("%w: deleteStory storyTitle is invalid", ErrMessageWriteInvalid)
	}
	result["storyTitle"] = trimJavaScriptSpace(title)
	result["confirmed"] = true
	return result, nil
}

func normalizeBulkDeleteStoriesFingerprintInput(input map[string]any) (map[string]any, error) {
	result := projectObject(input, []string{"storyIds", "storyTitles"})
	storyIDs, idsOK := result["storyIds"].([]any)
	titles, titlesOK := result["storyTitles"].([]any)
	if !idsOK || !titlesOK || len(storyIDs) == 0 || len(storyIDs) != len(titles) {
		return nil, fmt.Errorf("%w: bulkDeleteStories targets are invalid", ErrMessageWriteInvalid)
	}
	for index := range storyIDs {
		if _, ok := storyIDs[index].(string); !ok {
			return nil, fmt.Errorf("%w: bulkDeleteStories story ID %d is invalid", ErrMessageWriteInvalid, index)
		}
		title, ok := titles[index].(string)
		if !ok {
			return nil, fmt.Errorf("%w: bulkDeleteStories story title %d is invalid", ErrMessageWriteInvalid, index)
		}
		titles[index] = trimJavaScriptSpace(title)
	}
	result["storyIds"] = storyIDs
	result["storyTitles"] = titles
	result["confirmed"] = true
	return result, nil
}

func projectObject(input map[string]any, allowedKeys []string) map[string]any {
	result := make(map[string]any, len(allowedKeys))
	for _, key := range allowedKeys {
		if value, exists := input[key]; exists {
			result[key] = value
		}
	}
	return result
}

func deleteNullProperty(object map[string]any, key string) {
	if value, exists := object[key]; exists && value == nil {
		delete(object, key)
	}
}

func trimJavaScriptSpace(value string) string {
	return strings.TrimFunc(value, func(character rune) bool {
		return unicode.IsSpace(character) || character == '\uFEFF'
	})
}

func writeCanonicalJavaScriptJSON(buffer *bytes.Buffer, value any) error {
	switch typed := value.(type) {
	case nil:
		buffer.WriteString("null")
	case bool:
		buffer.WriteString(strconv.FormatBool(typed))
	case string:
		return writeJavaScriptJSONString(buffer, typed)
	case float64:
		formatted, err := formatJavaScriptNumber(typed)
		if err != nil {
			return err
		}
		buffer.WriteString(formatted)
	case json.Number:
		number, err := typed.Float64()
		if err != nil {
			return fmt.Errorf("invalid JSON number %q: %w", typed, err)
		}
		formatted, err := formatJavaScriptNumber(number)
		if err != nil {
			return err
		}
		buffer.WriteString(formatted)
	case []any:
		buffer.WriteByte('[')
		for index, item := range typed {
			if index > 0 {
				buffer.WriteByte(',')
			}
			if err := writeCanonicalJavaScriptJSON(buffer, item); err != nil {
				return err
			}
		}
		buffer.WriteByte(']')
	case map[string]any:
		keys := make([]string, 0, len(typed))
		for key := range typed {
			keys = append(keys, key)
		}
		sort.Slice(keys, func(left, right int) bool {
			leftFolded := strings.ToLower(keys[left])
			rightFolded := strings.ToLower(keys[right])
			if leftFolded == rightFolded {
				return keys[left] < keys[right]
			}
			return leftFolded < rightFolded
		})
		buffer.WriteByte('{')
		for index, key := range keys {
			if index > 0 {
				buffer.WriteByte(',')
			}
			if err := writeJavaScriptJSONString(buffer, key); err != nil {
				return err
			}
			buffer.WriteByte(':')
			if err := writeCanonicalJavaScriptJSON(buffer, typed[key]); err != nil {
				return err
			}
		}
		buffer.WriteByte('}')
	default:
		return fmt.Errorf("unsupported JSON value %T", value)
	}
	return nil
}

func writeJavaScriptJSONString(buffer *bytes.Buffer, value string) error {
	var encoded bytes.Buffer
	encoder := json.NewEncoder(&encoded)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(value); err != nil {
		return err
	}
	quoted := bytes.TrimSuffix(encoded.Bytes(), []byte{'\n'})
	quoted = bytes.ReplaceAll(quoted, []byte(`\u2028`), []byte("\u2028"))
	quoted = bytes.ReplaceAll(quoted, []byte(`\u2029`), []byte("\u2029"))
	buffer.Write(quoted)
	return nil
}

func formatJavaScriptNumber(value float64) (string, error) {
	if math.IsInf(value, 0) || math.IsNaN(value) {
		return "", fmt.Errorf("invalid JSON number %v", value)
	}
	if value == 0 {
		return "0", nil
	}
	absolute := math.Abs(value)
	if absolute >= 1e-6 && absolute < 1e21 {
		return strconv.FormatFloat(value, 'f', -1, 64), nil
	}
	mantissaAndExponent := strings.Split(strconv.FormatFloat(value, 'e', -1, 64), "e")
	if len(mantissaAndExponent) != 2 {
		return "", fmt.Errorf("invalid JSON number %v", value)
	}
	exponent, err := strconv.Atoi(mantissaAndExponent[1])
	if err != nil {
		return "", fmt.Errorf("invalid JSON number %v", value)
	}
	return fmt.Sprintf("%se%+d", mantissaAndExponent[0], exponent), nil
}
