package storieshttp

import (
	"encoding/hex"
	"errors"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/google/uuid"
)

var jiraCSVSourceKeyPattern = regexp.MustCompile(`^[A-Z][A-Z0-9]+-[1-9][0-9]*$`)

const (
	maximumStoryImportItems     = 50
	maximumImportSourceKeyBytes = 256
	sha256DigestHexLength       = 64
	storyImportProviderJiraCSV  = "jira_csv"
	storyImportProviderFile     = "file"
)

type AppStoryImportRequest struct {
	Provider     string               `json:"provider" validate:"required,oneof=jira_csv file"`
	SourceDigest string               `json:"sourceDigest" validate:"required,len=64"`
	Items        []AppStoryImportItem `json:"items" validate:"required,min=1,max=50,dive"`
}

type AppStoryImportItem struct {
	SourceKey string      `json:"sourceKey" validate:"required"`
	Story     AppNewStory `json:"story" validate:"required"`
}

func (request AppStoryImportRequest) Validate() error {
	if len(request.Items) < 1 || len(request.Items) > maximumStoryImportItems {
		return errors.New("story import must contain between 1 and 50 reviewed items")
	}
	if request.Provider != storyImportProviderJiraCSV && request.Provider != storyImportProviderFile {
		return errors.New("story import provider is invalid")
	}
	if len(request.SourceDigest) != sha256DigestHexLength {
		return errors.New("sourceDigest must be a SHA-256 hexadecimal digest")
	}
	if _, err := hex.DecodeString(request.SourceDigest); err != nil {
		return errors.New("sourceDigest must be a SHA-256 hexadecimal digest")
	}

	seenSourceKeys := make(map[string]struct{}, len(request.Items))
	for _, item := range request.Items {
		if err := validateImportSourceKey(item.SourceKey); err != nil {
			return err
		}
		if request.Provider == storyImportProviderJiraCSV && !jiraCSVSourceKeyPattern.MatchString(item.SourceKey) {
			return errors.New("jira_csv source keys must be Jira issue keys")
		}
		if _, duplicate := seenSourceKeys[item.SourceKey]; duplicate {
			return errors.New("story import contains duplicate source keys")
		}
		seenSourceKeys[item.SourceKey] = struct{}{}
		if item.Story.IdempotencyKey != nil {
			return errors.New("story import items cannot provide idempotencyKey")
		}
	}
	return nil
}

func validateImportSourceKey(sourceKey string) error {
	if sourceKey == "" || strings.TrimSpace(sourceKey) != sourceKey || !utf8.ValidString(sourceKey) || len(sourceKey) > maximumImportSourceKeyBytes {
		return errors.New("story import contains an invalid source key")
	}
	for _, character := range sourceKey {
		if unicode.IsControl(character) {
			return errors.New("story import contains an invalid source key")
		}
	}
	return nil
}

type AppStoryImportResponse struct {
	Counts AppStoryImportCounts       `json:"counts"`
	Items  []AppStoryImportItemResult `json:"items"`
}

type AppStoryImportCounts struct {
	Total    int `json:"total"`
	Created  int `json:"created"`
	Replayed int `json:"replayed"`
	Failed   int `json:"failed"`
}

type AppStoryImportItemResult struct {
	SourceKey string                   `json:"sourceKey"`
	StoryID   *uuid.UUID               `json:"storyId"`
	Created   bool                     `json:"created"`
	Error     *AppStoryImportItemError `json:"error"`
}

type AppStoryImportItemError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
