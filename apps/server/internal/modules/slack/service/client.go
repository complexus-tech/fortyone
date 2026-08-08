package slack

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const defaultSlackAPIBaseURL = "https://slack.com/api"

type RateLimitError struct {
	Method     string
	RetryAfter time.Duration
}

type SlackAPIError struct {
	Method string
	Code   string
}

func (e *SlackAPIError) Error() string {
	return fmt.Sprintf("slack api %s returned %s", e.Method, e.Code)
}

func SlackAPIErrorCode(err error) (string, bool) {
	var apiErr *SlackAPIError
	if !errors.As(err, &apiErr) {
		return "", false
	}
	return apiErr.Code, true
}

func (e *RateLimitError) Error() string {
	return fmt.Sprintf("slack api %s rate limited; retry after %s", e.Method, e.RetryAfter)
}

func SlackRetryAfter(err error) (time.Duration, bool) {
	var rateLimit *RateLimitError
	if !errors.As(err, &rateLimit) {
		return 0, false
	}
	return rateLimit.RetryAfter, true
}

type slackWebClient struct {
	baseURL string
	client  *http.Client
}

func newSlackWebClient(client *http.Client) *slackWebClient {
	if client == nil {
		client = &http.Client{Timeout: 8 * time.Second}
	}
	return &slackWebClient{baseURL: defaultSlackAPIBaseURL, client: client}
}

func (c *slackWebClient) callJSON(ctx context.Context, botToken, method string, payload, output any) error {
	if c == nil || c.client == nil {
		return errors.New("slack web client is not configured")
	}
	method = strings.TrimSpace(method)
	if method == "" {
		return errors.New("slack api method is required")
	}
	var body io.Reader
	if payload != nil {
		encoded, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("encode slack %s request: %w", method, err)
		}
		body = bytes.NewReader(encoded)
	}
	httpMethod := http.MethodPost
	if payload == nil {
		httpMethod = http.MethodGet
	}
	req, err := http.NewRequestWithContext(ctx, httpMethod, strings.TrimRight(c.baseURL, "/")+"/"+method, body)
	if err != nil {
		return fmt.Errorf("create slack %s request: %w", method, err)
	}
	if payload != nil {
		req.Header.Set("Content-Type", "application/json; charset=utf-8")
	}
	if token := strings.TrimSpace(botToken); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	response, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("call slack %s: %w", method, err)
	}
	defer response.Body.Close()
	if response.StatusCode == http.StatusTooManyRequests {
		return &RateLimitError{Method: method, RetryAfter: parseRetryAfter(response.Header.Get("Retry-After"))}
	}
	data, err := io.ReadAll(io.LimitReader(response.Body, 1<<20))
	if err != nil {
		return fmt.Errorf("read slack %s response: %w", method, err)
	}
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("slack api %s returned http %d", method, response.StatusCode)
	}
	var envelope struct {
		OK    bool   `json:"ok"`
		Error string `json:"error"`
	}
	if err := json.Unmarshal(data, &envelope); err != nil {
		return fmt.Errorf("decode slack %s response: %w", method, err)
	}
	if !envelope.OK {
		providerError := strings.TrimSpace(envelope.Error)
		if providerError == "" {
			providerError = "unknown_error"
		}
		return &SlackAPIError{Method: method, Code: providerError}
	}
	if output != nil {
		if err := json.Unmarshal(data, output); err != nil {
			return fmt.Errorf("decode slack %s output: %w", method, err)
		}
	}
	return nil
}

func (c *slackWebClient) appsUninstall(ctx context.Context, clientID, clientSecret, botToken string) error {
	values := url.Values{}
	values.Set("client_id", strings.TrimSpace(clientID))
	values.Set("client_secret", strings.TrimSpace(clientSecret))
	values.Set("token", strings.TrimSpace(botToken))
	return c.callForm(ctx, "apps.uninstall", values, nil)
}

func (c *slackWebClient) oauthV2Access(ctx context.Context, clientID, clientSecret, redirectURL, code string, output any) error {
	values := url.Values{}
	values.Set("client_id", strings.TrimSpace(clientID))
	values.Set("client_secret", strings.TrimSpace(clientSecret))
	values.Set("redirect_uri", strings.TrimSpace(redirectURL))
	values.Set("code", strings.TrimSpace(code))
	return c.callForm(ctx, "oauth.v2.access", values, output)
}

func (c *slackWebClient) callForm(ctx context.Context, method string, values url.Values, output any) error {
	if c == nil || c.client == nil {
		return errors.New("slack web client is not configured")
	}
	method = strings.TrimSpace(method)
	if method == "" {
		return errors.New("slack api method is required")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(c.baseURL, "/")+"/"+method, strings.NewReader(values.Encode()))
	if err != nil {
		return fmt.Errorf("create slack %s request: %w", method, err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("call slack %s: %w", method, err)
	}
	defer response.Body.Close()
	if response.StatusCode == http.StatusTooManyRequests {
		return &RateLimitError{Method: method, RetryAfter: parseRetryAfter(response.Header.Get("Retry-After"))}
	}
	data, err := io.ReadAll(io.LimitReader(response.Body, 1<<20))
	if err != nil {
		return fmt.Errorf("read slack %s response: %w", method, err)
	}
	var envelope struct {
		OK    bool   `json:"ok"`
		Error string `json:"error"`
	}
	if err := json.Unmarshal(data, &envelope); err != nil {
		return fmt.Errorf("decode slack %s response: %w", method, err)
	}
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices || !envelope.OK {
		providerError := strings.TrimSpace(envelope.Error)
		if providerError == "" {
			providerError = strconv.Itoa(response.StatusCode)
		}
		return &SlackAPIError{Method: method, Code: providerError}
	}
	if output != nil {
		if err := json.Unmarshal(data, output); err != nil {
			return fmt.Errorf("decode slack %s output: %w", method, err)
		}
	}
	return nil
}

func parseRetryAfter(value string) time.Duration {
	seconds, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil || seconds <= 0 {
		seconds = 1
	}
	return time.Duration(seconds) * time.Second
}
