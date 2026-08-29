package messaging

import (
	"context"
	"fmt"
	"strings"
)

const defaultUnavailableMessage = "conversational assistance is unavailable"

type unavailableAssistant struct {
	message string
}

// NewUnavailableAssistant returns a provider-neutral permanent-failure
// assistant for processes that start without model configuration. The message
// must be an operator-safe reason and should never contain secret values.
func NewUnavailableAssistant(message string) Assistant {
	message = strings.TrimSpace(message)
	if message == "" {
		message = defaultUnavailableMessage
	}
	return unavailableAssistant{message: message}
}

func (a unavailableAssistant) Respond(context.Context, Request) (Response, error) {
	return Response{}, fmt.Errorf("%w: %s", ErrAssistantNotConfigured, a.message)
}
