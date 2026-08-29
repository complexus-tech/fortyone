package slack

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/url"
	"time"
)

func (p *EventProcessor) accountLinkMessage(ctx context.Context, workspace workspaceRecord, event normalizedSlackEvent) (string, error) {
	link, err := p.accountLinkURL(ctx, workspace, event)
	if err != nil {
		return "", err
	}
	return "Connect your FortyOne account before using Maya in Slack: " + link, nil
}

func (p *EventProcessor) accountLinkURL(ctx context.Context, workspace workspaceRecord, event normalizedSlackEvent) (string, error) {
	nonce := make([]byte, 32)
	if _, err := io.ReadFull(p.random, nonce); err != nil {
		return "", fmt.Errorf("generate Slack account-link nonce: %w", err)
	}
	digest := sha256.Sum256(nonce)
	if err := p.store.CreateNonce(ctx, nonceInput{
		Provider:            "slack",
		Purpose:             "account_link",
		NonceHash:           digest[:],
		WorkspaceID:         workspace.ID,
		ExternalWorkspaceID: event.TeamID,
		ExternalUserID:      event.UserID,
		ExpiresAt:           p.clock.Now().UTC().Add(15 * time.Minute),
	}); err != nil {
		return "", err
	}
	link := buildWorkspaceURL(p.website, workspace.Slug, "settings", "integrations", "slack")
	if link == "" {
		return "", errors.New("build Slack account-link URL")
	}
	link += "?slack_link_token=" + url.QueryEscape(base64.RawURLEncoding.EncodeToString(nonce))
	return link, nil
}
