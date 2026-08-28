package gitlab

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/complexus-tech/projects-api/internal/platform/codehost"
)

const maxGitLabResponseBytes = 1 << 20

func (adapter *Adapter) do(
	ctx context.Context,
	installation codehost.InstallationRef,
	method, resourcePath string,
	query url.Values,
	input, output any,
) (http.Header, error) {
	if adapter == nil || adapter.baseURL == nil || adapter.client == nil || adapter.tokens == nil {
		return nil, codehost.ErrAuthentication
	}
	if err := validateInstallation(installation); err != nil {
		return nil, err
	}
	token, err := adapter.tokens.AccessToken(ctx, installation)
	if err != nil {
		return nil, errors.Join(codehost.ErrAuthentication, err)
	}
	if err := validateAccessToken(token); err != nil {
		return nil, err
	}

	var body io.Reader
	if input != nil {
		encoded, encodeErr := json.Marshal(input)
		if encodeErr != nil {
			return nil, errors.Join(codehost.ErrInvalidInput, encodeErr)
		}
		defer clear(encoded)
		body = bytes.NewReader(encoded)
	}
	target := *adapter.baseURL
	target.Path = strings.TrimSuffix(target.Path, "/") + "/" + strings.TrimPrefix(resourcePath, "/")
	target.RawQuery = query.Encode()
	request, err := http.NewRequestWithContext(ctx, method, target.String(), body)
	if err != nil {
		return nil, errors.Join(codehost.ErrInvalidInput, err)
	}
	request.Header.Set("Accept", "application/json")
	if input != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	setAuthorization(request.Header, token)
	response, err := adapter.client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("GitLab API request: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, maxGitLabResponseBytes))
		return response.Header.Clone(), classifyResponse(response.StatusCode)
	}
	if output == nil {
		read, readErr := io.Copy(io.Discard, io.LimitReader(response.Body, maxGitLabResponseBytes+1))
		if readErr != nil {
			return nil, fmt.Errorf("read GitLab API response: %w", readErr)
		}
		if read > maxGitLabResponseBytes {
			return nil, errors.New("GitLab API response exceeds one mebibyte")
		}
		return response.Header.Clone(), nil
	}
	limited := io.LimitReader(response.Body, maxGitLabResponseBytes+1)
	encoded, err := io.ReadAll(limited)
	if err != nil {
		return nil, fmt.Errorf("read GitLab API response: %w", err)
	}
	if len(encoded) > maxGitLabResponseBytes {
		return nil, errors.New("GitLab API response exceeds one mebibyte")
	}
	if err := json.Unmarshal(encoded, output); err != nil {
		return nil, fmt.Errorf("decode GitLab API response: %w", err)
	}
	return response.Header.Clone(), nil
}

func validateAccessToken(token AccessToken) error {
	if strings.TrimSpace(token.Value) == "" || strings.ContainsAny(token.Value, "\r\n") {
		return errors.Join(codehost.ErrAuthentication, errors.New("GitLab access token is invalid"))
	}
	switch token.Kind {
	case TokenOAuthBearer, TokenPrivate:
		return nil
	default:
		return errors.Join(codehost.ErrAuthentication, errors.New("GitLab access token kind is unsupported"))
	}
}

func setAuthorization(headers http.Header, token AccessToken) {
	if token.Kind == TokenPrivate {
		headers.Set("PRIVATE-TOKEN", token.Value)
		return
	}
	headers.Set("Authorization", "Bearer "+token.Value)
}

func classifyResponse(status int) error {
	switch status {
	case http.StatusUnauthorized:
		return fmt.Errorf("%w: GitLab rejected the credential", codehost.ErrAuthentication)
	case http.StatusForbidden:
		return fmt.Errorf("%w: GitLab denied the installation grant", codehost.ErrGrantRevoked)
	case http.StatusNotFound:
		return codehost.ErrNotFound
	case http.StatusTooManyRequests:
		return codehost.ErrRateLimited
	default:
		return fmt.Errorf("GitLab API returned status %d", status)
	}
}

func validateInstallation(installation codehost.InstallationRef) error {
	if err := codehost.ValidateInstallation(installation); err != nil || installation.Provider != ProviderKey {
		return codehost.ErrInvalidInput
	}
	return nil
}
