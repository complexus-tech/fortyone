package codehost

import (
	"fmt"
	"net/url"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
)

func ValidateInstallation(installation InstallationRef) error {
	if installation.Provider == "" || installation.WorkspaceID == uuid.Nil ||
		installation.InstallationID == uuid.Nil || installation.Generation == uuid.Nil ||
		!validText(installation.ExternalInstallationID, 255) {
		return ErrInvalidInput
	}
	return nil
}

func ValidateCursor(cursor Cursor) error {
	if cursor.Limit < 1 || cursor.Limit > 100 || len(cursor.Value) > 1024 || !utf8.ValidString(cursor.Value) {
		return ErrInvalidInput
	}
	return nil
}

func ValidateRepository(repository RepositoryRef) error {
	if !validText(repository.ExternalID, 255) || !validText(repository.Owner, 512) ||
		!validText(repository.Name, 255) || !validText(repository.FullName, 512) ||
		!validHTTPSURL(repository.WebURL) {
		return ErrInvalidInput
	}
	return nil
}

func validHTTPSURL(value string) bool {
	if !validText(value, 2048) {
		return false
	}
	parsed, err := url.Parse(strings.TrimSpace(value))
	return err == nil && strings.EqualFold(parsed.Scheme, "https") && parsed.Host != "" && parsed.User == nil
}

func ValidateCreateWorkItem(command CreateWorkItem) error {
	if err := ValidateRepository(command.Repository); err != nil {
		return err
	}
	if !validText(command.Title, 512) || len(command.Body) > 1<<20 || !utf8.ValidString(command.Body) {
		return ErrInvalidInput
	}
	return nil
}

func ValidateAddComment(command AddComment) error {
	if err := ValidateRepository(command.WorkItem.Repository); err != nil {
		return err
	}
	if command.WorkItem.Number < 1 || !validText(command.Body, 1<<20) {
		return ErrInvalidInput
	}
	return nil
}

func ParseWorkItemState(value string) (WorkItemState, error) {
	switch normalized := WorkItemState(strings.ToLower(strings.TrimSpace(value))); normalized {
	case WorkItemStateOpen, WorkItemStateClosed:
		return normalized, nil
	default:
		return "", fmt.Errorf("%w: unsupported work-item state", ErrInvalidInput)
	}
}

func validText(value string, maximum int) bool {
	value = strings.TrimSpace(value)
	if value == "" || len(value) > maximum || !utf8.ValidString(value) {
		return false
	}
	for _, character := range value {
		if character == 0 || character == 0x7f {
			return false
		}
	}
	return true
}
