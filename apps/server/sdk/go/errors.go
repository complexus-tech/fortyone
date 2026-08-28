package fortyone

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

// APIError is the safe, transport-neutral subset of a FortyOne error response.
// It deliberately does not retain the response body or authentication data.
type APIError struct {
	StatusCode        int
	Code              string
	Message           string
	RequestID         string
	Fields            []ErrorField
	RetryAfterSeconds int
}

func (err *APIError) Error() string {
	if err.RequestID == "" {
		return fmt.Sprintf("FortyOne API request failed: %s (HTTP %d)", err.Message, err.StatusCode)
	}
	return fmt.Sprintf("FortyOne API request failed: %s (HTTP %d, request %s)", err.Message, err.StatusCode, err.RequestID)
}

// NewAPIError converts an unsuccessful response to a safe structured error.
// Callers should branch on Code or StatusCode, never Message.
func NewAPIError(status int, header http.Header, body []byte) *APIError {
	result := &APIError{
		StatusCode: status,
		Code:       "unexpected_response",
		Message:    fmt.Sprintf("request failed with HTTP %d", status),
	}
	var envelope ComponentsCommonErrorResponse
	if json.Unmarshal(body, &envelope) == nil && envelope.Error.Code != "" && envelope.Error.Message != "" {
		result.Code = envelope.Error.Code
		result.Message = envelope.Error.Message
		result.RequestID = envelope.Error.RequestId
		if envelope.Error.Fields != nil {
			result.Fields = append([]ErrorField(nil), (*envelope.Error.Fields)...)
		}
	}
	if result.RequestID == "" {
		result.RequestID = strings.TrimSpace(header.Get("X-Request-ID"))
	}
	if seconds, err := strconv.Atoi(strings.TrimSpace(header.Get("Retry-After"))); err == nil && seconds > 0 {
		result.RetryAfterSeconds = seconds
	}
	return result
}
