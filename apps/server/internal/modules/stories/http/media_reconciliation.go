package storieshttp

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"strings"

	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
	"github.com/google/uuid"
	"golang.org/x/net/html"
)

const reconcileDescriptionMediaField = "reconcileDescriptionMedia"

func storyMediaReconciliationRequest(
	requestData map[string]json.RawMessage,
	workspaceSlug string,
	storyID uuid.UUID,
) (bool, []uuid.UUID, error) {
	rawReconcile, present := requestData[reconcileDescriptionMediaField]
	if !present {
		return false, nil, nil
	}

	var reconcile bool
	if err := json.Unmarshal(rawReconcile, &reconcile); err != nil {
		return false, nil, fmt.Errorf("%s must be a boolean: %w", reconcileDescriptionMediaField, err)
	}
	if !reconcile {
		return false, nil, nil
	}

	rawContent, present := requestData["descriptionHTML"]
	if !present || bytes.Equal(bytes.TrimSpace(rawContent), []byte("null")) {
		return false, nil, errors.New("descriptionHTML is required when reconciling story media")
	}
	var contentHTML string
	if err := json.Unmarshal(rawContent, &contentHTML); err != nil {
		return false, nil, fmt.Errorf("descriptionHTML must be a string: %w", err)
	}

	references, err := parseStoryMediaReferences(contentHTML, workspaceSlug, storyID)
	if err != nil {
		return false, nil, err
	}
	return true, references, nil
}

func parseStoryMediaReferences(contentHTML, workspaceSlug string, storyID uuid.UUID) ([]uuid.UUID, error) {
	if storyID == uuid.Nil || strings.TrimSpace(workspaceSlug) == "" {
		return nil, stories.ErrInvalidStoryMediaReference
	}

	tokenizer := html.NewTokenizer(strings.NewReader(contentHTML))
	references := make([]uuid.UUID, 0)
	seen := make(map[uuid.UUID]struct{})
	for {
		tokenType := tokenizer.Next()
		switch tokenType {
		case html.ErrorToken:
			if err := tokenizer.Err(); err != nil && !errors.Is(err, io.EOF) {
				return nil, fmt.Errorf("parse story description media: %w", err)
			}
			return references, nil
		case html.StartTagToken, html.SelfClosingTagToken:
			token := tokenizer.Token()
			if token.Data != "img" && token.Data != "video" {
				continue
			}

			src, attachmentValue := storyMediaAttributes(token.Attr)
			if attachmentValue == "" {
				continue
			}
			attachmentID, err := uuid.Parse(attachmentValue)
			if err != nil || attachmentID == uuid.Nil {
				return nil, stories.ErrInvalidStoryMediaReference
			}
			if err := validateStableStoryMediaURL(src, workspaceSlug, storyID, attachmentID); err != nil {
				return nil, err
			}
			if _, exists := seen[attachmentID]; exists {
				continue
			}
			seen[attachmentID] = struct{}{}
			references = append(references, attachmentID)
		}
	}
}

func storyMediaAttributes(attributes []html.Attribute) (src, attachmentID string) {
	for _, attribute := range attributes {
		switch attribute.Key {
		case "src":
			src = strings.TrimSpace(attribute.Val)
		case "data-attachment-id":
			attachmentID = strings.TrimSpace(attribute.Val)
		}
	}
	return src, attachmentID
}

func validateStableStoryMediaURL(
	rawURL, workspaceSlug string,
	storyID, attachmentID uuid.UUID,
) error {
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Opaque != "" || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
		return stories.ErrInvalidStoryMediaReference
	}
	if parsed.Scheme != "" && parsed.Scheme != "http" && parsed.Scheme != "https" {
		return stories.ErrInvalidStoryMediaReference
	}
	if parsed.Scheme != "" && parsed.Host == "" {
		return stories.ErrInvalidStoryMediaReference
	}
	if parsed.Scheme == "" && parsed.Host != "" {
		return stories.ErrInvalidStoryMediaReference
	}

	expectedPath := storyMediaURL(workspaceSlug, storyID, attachmentID)
	if parsed.EscapedPath() != expectedPath {
		return stories.ErrInvalidStoryMediaReference
	}
	return nil
}
