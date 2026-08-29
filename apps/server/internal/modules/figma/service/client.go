package figma

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

const figmaAPIBase = "https://api.figma.com"

type apiClient struct {
	http   *http.Client
	config Config
}

type APIError struct {
	StatusCode int
	RetryAfter time.Duration
	Message    string
}

func parseRetryAfter(value string) time.Duration {
	seconds, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil || seconds <= 0 {
		return 0
	}
	return time.Duration(seconds) * time.Second
}

func (e *APIError) Error() string {
	if e.StatusCode == http.StatusTooManyRequests {
		if e.RetryAfter > 0 {
			return fmt.Sprintf("Figma rate limit reached; try again in %s", e.RetryAfter)
		}
		return "Figma rate limit reached; try again later"
	}
	return fmt.Sprintf("Figma API returned %d: %s", e.StatusCode, e.Message)
}

type oauthTokenResponse struct {
	UserIDString string `json:"user_id_string"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
}

type figmaUser struct {
	ID     string `json:"id"`
	Email  string `json:"email"`
	Handle string `json:"handle"`
}

type fileResponse struct {
	Name         string                     `json:"name"`
	LastModified time.Time                  `json:"lastModified"`
	ThumbnailURL string                     `json:"thumbnailUrl"`
	Version      string                     `json:"version"`
	Nodes        map[string]json.RawMessage `json:"nodes"`
	Document     json.RawMessage            `json:"document"`
}

type nodeEnvelope struct {
	Document figmaNode `json:"document"`
}

type figmaNode struct {
	ID         string      `json:"id"`
	Name       string      `json:"name"`
	Type       string      `json:"type"`
	Characters string      `json:"characters"`
	Children   []figmaNode `json:"children"`
}

const (
	maxImportedTextItems      = 24
	maxImportedTextItemRunes  = 500
	maxImportedTextTotalRunes = 4_000
)

func collectTextContent(node figmaNode) []string {
	items := make([]string, 0, maxImportedTextItems)
	seen := make(map[string]struct{}, maxImportedTextItems)
	totalRunes := 0
	var visit func(figmaNode)
	visit = func(current figmaNode) {
		if len(items) >= maxImportedTextItems {
			return
		}
		if current.Type == "TEXT" {
			value := strings.TrimSpace(current.Characters)
			runes := []rune(value)
			if len(runes) > maxImportedTextItemRunes {
				runes = runes[:maxImportedTextItemRunes]
				value = string(runes)
			}
			if value != "" {
				if _, exists := seen[value]; !exists {
					remaining := maxImportedTextTotalRunes - totalRunes
					if remaining <= 0 {
						return
					}
					if len(runes) > remaining {
						runes = runes[:remaining]
						value = string(runes)
					}
					seen[value] = struct{}{}
					items = append(items, value)
					totalRunes += len(runes)
				}
			}
		}
		for _, child := range current.Children {
			visit(child)
			if len(items) >= maxImportedTextItems {
				return
			}
		}
	}
	visit(node)
	return items
}

func extractTextContent(data json.RawMessage) []string {
	var node figmaNode
	if err := json.Unmarshal(data, &node); err != nil {
		return nil
	}
	return collectTextContent(node)
}

func (c apiClient) exchange(ctx context.Context, code, redirectURL, verifier string) (oauthTokenResponse, error) {
	values := url.Values{"redirect_uri": {redirectURL}, "code": {code}, "grant_type": {"authorization_code"}, "code_verifier": {verifier}}
	return c.tokenRequest(ctx, "/v1/oauth/token", values)
}

func (c apiClient) refresh(ctx context.Context, refreshToken string) (oauthTokenResponse, error) {
	return c.tokenRequest(ctx, "/v1/oauth/refresh", url.Values{"refresh_token": {refreshToken}})
}

func (c apiClient) tokenRequest(ctx context.Context, path string, values url.Values) (oauthTokenResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, figmaAPIBase+path, strings.NewReader(values.Encode()))
	if err != nil {
		return oauthTokenResponse{}, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(c.config.ClientID, c.config.ClientSecret)
	var response oauthTokenResponse
	if err := c.do(req, &response); err != nil {
		return oauthTokenResponse{}, err
	}
	if strings.TrimSpace(response.AccessToken) == "" {
		return oauthTokenResponse{}, errors.New("figma returned an empty access token")
	}
	return response, nil
}

func (c apiClient) currentUser(ctx context.Context, token string) (figmaUser, error) {
	var response figmaUser
	err := c.get(ctx, token, "/v1/me", &response)
	return response, err
}

func (c apiClient) resolve(ctx context.Context, token string, artifact Artifact) (Artifact, error) {
	path := "/v1/files/" + url.PathEscape(artifact.FileKey)
	if artifact.NodeID != nil {
		path += "/nodes?ids=" + url.QueryEscape(*artifact.NodeID)
	}
	var response fileResponse
	if err := c.get(ctx, token, path, &response); err != nil {
		return Artifact{}, err
	}
	artifact.FileName = response.Name
	artifact.Version = optional(response.Version)
	artifact.LastModified = &response.LastModified
	artifact.ThumbnailURL = optional(response.ThumbnailURL)
	artifact.Metadata = response.Document
	artifact.TextContent = extractTextContent(response.Document)
	if artifact.NodeID != nil {
		raw, ok := response.Nodes[*artifact.NodeID]
		if !ok {
			return Artifact{}, errors.New("the linked Figma frame was not found")
		}
		var node nodeEnvelope
		if err := json.Unmarshal(raw, &node); err != nil {
			return Artifact{}, err
		}
		artifact.NodeName = optional(node.Document.Name)
		artifact.NodeType = optional(node.Document.Type)
		artifact.Metadata = raw
		artifact.TextContent = collectTextContent(node.Document)
		var images struct {
			Images map[string]string `json:"images"`
		}
		if err := c.get(ctx, token, "/v1/images/"+url.PathEscape(artifact.FileKey)+"?ids="+url.QueryEscape(*artifact.NodeID)+"&format=png&scale=1", &images); err == nil {
			artifact.ThumbnailURL = optional(images.Images[*artifact.NodeID])
		}
	}
	return artifact, nil
}

func (c apiClient) createDevResource(ctx context.Context, token string, link StoryLink, storyURL string) (*string, error) {
	if link.Artifact.NodeID == nil {
		return nil, nil
	}
	body := map[string]any{"dev_resources": []map[string]string{{"file_key": link.Artifact.FileKey, "node_id": *link.Artifact.NodeID, "name": "FortyOne story", "url": storyURL}}}
	var response struct {
		Links []struct {
			ID string `json:"id"`
		} `json:"links_created"`
		Errors []map[string]any `json:"errors"`
	}
	if err := c.jsonRequest(ctx, token, http.MethodPost, "/v1/dev_resources", body, &response); err != nil {
		return nil, err
	}
	if len(response.Links) == 0 {
		if len(response.Errors) > 0 {
			return nil, fmt.Errorf("figma could not create the story backlink")
		}
		return nil, nil
	}
	return &response.Links[0].ID, nil
}

func (c apiClient) deleteDevResource(ctx context.Context, token, fileKey, resourceID string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, figmaAPIBase+"/v1/files/"+url.PathEscape(fileKey)+"/dev_resources/"+url.PathEscape(resourceID), nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	return c.do(req, nil)
}

func (c apiClient) createWebhook(ctx context.Context, token, eventType, fileKey, endpoint, passcode string) (int64, error) {
	body := map[string]any{"event_type": eventType, "context": "file", "context_id": fileKey, "endpoint": endpoint, "passcode": passcode, "description": "FortyOne linked design updates"}
	var response struct {
		ID FlexibleInt64 `json:"id"`
	}
	if err := c.jsonRequest(ctx, token, http.MethodPost, "/v2/webhooks", body, &response); err != nil {
		return 0, err
	}
	if response.ID <= 0 {
		return 0, errors.New("figma returned an invalid webhook id")
	}
	return int64(response.ID), nil
}

func (c apiClient) deleteWebhook(ctx context.Context, token string, webhookID int64) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, fmt.Sprintf("%s/v2/webhooks/%d", figmaAPIBase, webhookID), nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	return c.do(req, nil)
}

func (c apiClient) get(ctx context.Context, token, path string, target any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, figmaAPIBase+path, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	return c.do(req, target)
}

func (c apiClient) jsonRequest(ctx context.Context, token, method, path string, body, target any) error {
	payload, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, method, figmaAPIBase+path, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	return c.do(req, target)
}

func (c apiClient) do(req *http.Request, target any) error {
	response, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(response.Body, 4096))
		return &APIError{
			StatusCode: response.StatusCode,
			RetryAfter: parseRetryAfter(response.Header.Get("Retry-After")),
			Message:    strings.TrimSpace(string(body)),
		}
	}
	if target == nil || response.StatusCode == http.StatusNoContent {
		return nil
	}
	return json.NewDecoder(response.Body).Decode(target)
}

func optional(value string) *string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return &value
}
