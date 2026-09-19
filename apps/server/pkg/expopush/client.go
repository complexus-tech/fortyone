package expopush

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const endpoint = "https://exp.host/--/api/v2/push/send"

type Message struct {
	To       string         `json:"to"`
	Title    string         `json:"title"`
	Body     string         `json:"body"`
	Sound    string         `json:"sound,omitempty"`
	Priority string         `json:"priority,omitempty"`
	Data     map[string]any `json:"data"`
}

type Result struct {
	InvalidTokens []string
}

type Sender interface {
	Send(context.Context, []Message) (Result, error)
}

type Client struct {
	httpClient *http.Client
}

func New(httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 15 * time.Second}
	}
	return &Client{httpClient: httpClient}
}

type responseEnvelope struct {
	Data []struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Details struct {
			Error string `json:"error"`
		} `json:"details"`
	} `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

func (client *Client) Send(ctx context.Context, messages []Message) (Result, error) {
	if len(messages) == 0 {
		return Result{}, nil
	}
	if len(messages) > 100 {
		return Result{}, fmt.Errorf("Expo push batch exceeds 100 messages")
	}
	payload, err := json.Marshal(messages)
	if err != nil {
		return Result{}, fmt.Errorf("marshal Expo push messages: %w", err)
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return Result{}, fmt.Errorf("build Expo push request: %w", err)
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Content-Type", "application/json")
	response, err := client.httpClient.Do(request)
	if err != nil {
		return Result{}, fmt.Errorf("send Expo push request: %w", err)
	}
	defer response.Body.Close()
	limited := io.LimitReader(response.Body, 1<<20)
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		body, _ := io.ReadAll(limited)
		return Result{}, fmt.Errorf("Expo push request failed with status %d: %s", response.StatusCode, string(body))
	}
	var envelope responseEnvelope
	if err := json.NewDecoder(limited).Decode(&envelope); err != nil {
		return Result{}, fmt.Errorf("decode Expo push response: %w", err)
	}
	if len(envelope.Errors) > 0 {
		return Result{}, fmt.Errorf("Expo push request rejected: %s", envelope.Errors[0].Message)
	}
	if len(envelope.Data) != len(messages) {
		return Result{}, fmt.Errorf("Expo push response returned %d tickets for %d messages", len(envelope.Data), len(messages))
	}
	result := Result{InvalidTokens: make([]string, 0)}
	for index, ticket := range envelope.Data {
		if ticket.Status == "ok" {
			continue
		}
		if ticket.Details.Error == "DeviceNotRegistered" {
			result.InvalidTokens = append(result.InvalidTokens, messages[index].To)
			continue
		}
		return result, fmt.Errorf("Expo push ticket rejected: %s (%s)", ticket.Message, ticket.Details.Error)
	}
	return result, nil
}
