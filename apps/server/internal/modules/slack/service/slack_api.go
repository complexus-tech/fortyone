package slack

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
)

func (s *Service) callSlackAPI(ctx context.Context, botToken, endpoint string, payload any, out any) error {
	method, err := slackAPIMethod(endpoint)
	if err != nil {
		return err
	}
	return s.slackClient().callJSON(ctx, botToken, method, payload, out)
}

func (s *Service) slackClient() *slackWebClient {
	if s.webClient == nil {
		s.webClient = newSlackWebClient(s.client)
		return s.webClient
	}
	if s.client != nil && s.webClient.client != s.client {
		baseURL := s.webClient.baseURL
		s.webClient = newSlackWebClient(s.client)
		if strings.TrimSpace(baseURL) != "" && baseURL != defaultSlackAPIBaseURL {
			s.webClient.baseURL = baseURL
		}
	}
	return s.webClient
}

func slackAPIMethod(endpoint string) (string, error) {
	endpoint = strings.TrimSpace(endpoint)
	if endpoint == "" {
		return "", errors.New("slack api endpoint is required")
	}
	parsed, err := url.Parse(endpoint)
	if err != nil {
		return "", fmt.Errorf("parse slack api endpoint: %w", err)
	}
	method := strings.TrimPrefix(parsed.Path, "/api/")
	if method == parsed.Path {
		method = strings.TrimPrefix(parsed.Path, "/")
	}
	if method == "" {
		return "", errors.New("slack api method is required")
	}
	if parsed.RawQuery != "" {
		method += "?" + parsed.RawQuery
	}
	return method, nil
}
