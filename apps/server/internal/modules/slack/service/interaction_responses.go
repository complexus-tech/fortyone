package slack

import (
	"encoding/json"
	"net/http"
	"strings"
)

func interactionValidationErrors(errorsByBlock map[string]string) (InteractionResponse, error) {
	body, err := json.Marshal(map[string]any{
		"response_action": "errors",
		"errors":          errorsByBlock,
	})
	if err != nil {
		return InteractionResponse{}, err
	}
	return InteractionResponse{StatusCode: http.StatusOK, ContentType: "application/json", Body: body}, nil
}

func interactionClearResponse() (InteractionResponse, error) {
	body, err := json.Marshal(map[string]string{"response_action": "clear"})
	if err != nil {
		return InteractionResponse{}, err
	}
	return InteractionResponse{StatusCode: http.StatusOK, ContentType: "application/json", Body: body}, nil
}

func interactionOptionsResponse(options []map[string]any) (InteractionResponse, error) {
	if options == nil {
		options = make([]map[string]any, 0)
	}
	options = limitedSlackOptions(options)
	body, err := json.Marshal(map[string]any{"options": options})
	if err != nil {
		return InteractionResponse{}, err
	}
	return InteractionResponse{StatusCode: http.StatusOK, ContentType: "application/json", Body: body}, nil
}

func interactionErrorMessage(err error) string {
	if err == nil {
		return "Unable to create story. Please try again."
	}
	message := strings.TrimSpace(err.Error())
	if message == "" {
		return "Unable to create story. Please try again."
	}
	const maxLength = 180
	if len(message) > maxLength {
		return strings.TrimSpace(message[:maxLength-3]) + "..."
	}
	return message
}
