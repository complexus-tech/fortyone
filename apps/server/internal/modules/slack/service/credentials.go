package slack

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/complexus-tech/projects-api/internal/platform/secretbox"
)

type slackCredential struct {
	AccessToken  string     `json:"accessToken"`
	RefreshToken string     `json:"refreshToken,omitempty"`
	ExpiresAt    *time.Time `json:"expiresAt,omitempty"`
}

type credentialCodec struct {
	box *secretbox.Box
}

func newCredentialCodec(secret string) (*credentialCodec, error) {
	box, err := secretbox.New(secret)
	if err != nil {
		return nil, err
	}
	return &credentialCodec{box: box}, nil
}

func (c *credentialCodec) seal(credential slackCredential) (string, int, error) {
	if c == nil || c.box == nil {
		return "", 0, errors.New("slack credential encryption is not configured")
	}
	credential.AccessToken = strings.TrimSpace(credential.AccessToken)
	credential.RefreshToken = strings.TrimSpace(credential.RefreshToken)
	if credential.AccessToken == "" {
		return "", 0, errors.New("slack access token is required")
	}
	payload, err := json.Marshal(credential)
	if err != nil {
		return "", 0, fmt.Errorf("encode slack credential: %w", err)
	}
	value, err := c.box.Seal(payload)
	if err != nil {
		return "", 0, fmt.Errorf("encrypt slack credential: %w", err)
	}
	return value, secretbox.CurrentVersion(), nil
}

func (c *credentialCodec) open(value string) (slackCredential, int, error) {
	if c == nil || c.box == nil {
		return slackCredential{}, 0, errors.New("slack credential encryption is not configured")
	}
	opened, err := c.box.Open(value)
	if err != nil {
		return slackCredential{}, 0, fmt.Errorf("decrypt slack credential: %w", err)
	}
	if opened.Version == 0 {
		token := strings.TrimSpace(string(opened.Plaintext))
		if token == "" {
			return slackCredential{}, 0, errors.New("slack access token is empty")
		}
		return slackCredential{AccessToken: token}, 0, nil
	}
	var credential slackCredential
	if err := json.Unmarshal(opened.Plaintext, &credential); err != nil {
		return slackCredential{}, 0, fmt.Errorf("decode slack credential: %w", err)
	}
	credential.AccessToken = strings.TrimSpace(credential.AccessToken)
	credential.RefreshToken = strings.TrimSpace(credential.RefreshToken)
	if credential.AccessToken == "" {
		return slackCredential{}, 0, errors.New("slack access token is empty")
	}
	return credential, opened.Version, nil
}

func (c *credentialCodec) sealPayload(payload []byte) (string, error) {
	if c == nil || c.box == nil {
		return "", errors.New("slack payload encryption is not configured")
	}
	if len(payload) == 0 {
		return "", errors.New("slack payload is empty")
	}
	value, err := c.box.Seal(payload)
	if err != nil {
		return "", fmt.Errorf("encrypt Slack payload: %w", err)
	}
	return value, nil
}

func (c *credentialCodec) openPayload(value string) ([]byte, error) {
	if c == nil || c.box == nil {
		return nil, errors.New("slack payload encryption is not configured")
	}
	opened, err := c.box.Open(value)
	if err != nil {
		return nil, fmt.Errorf("decrypt Slack payload: %w", err)
	}
	if opened.Version == 0 {
		return nil, errors.New("unencrypted Slack payload is not allowed")
	}
	return opened.Plaintext, nil
}
