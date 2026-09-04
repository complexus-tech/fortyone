package googledrive

import (
	"errors"
	"strings"
	"unicode"

	"github.com/complexus-tech/projects-api/internal/modules/googledrive/domain"
)

func validateCreatedFile(file domain.ProviderFile, fileType domain.FileType) error {
	expectedMimeType := googleDocumentMimeType
	if fileType == domain.FileTypeSpreadsheet {
		expectedMimeType = googleSpreadsheetMimeType
	}
	if file.Trashed || file.MimeType != expectedMimeType {
		return errors.New("Google returned an unexpected created file")
	}
	return nil
}

func hasOAuthScope(scopes []string, required string) bool {
	required = strings.TrimSpace(required)
	for _, scope := range scopes {
		if strings.TrimSpace(scope) == required {
			return true
		}
	}
	return false
}

func supportedFile(mimeType string) bool {
	switch mimeType {
	case googleDocumentMimeType, googleSpreadsheetMimeType, googlePresentationMimeType,
		"application/pdf", "text/plain", "text/csv",
		"image/png", "image/jpeg", "image/webp":
		return true
	default:
		return false
	}
}

func readableFile(mimeType string) bool {
	switch mimeType {
	case googleDocumentMimeType, googleSpreadsheetMimeType, googlePresentationMimeType,
		"text/plain", "text/csv":
		return true
	default:
		return false
	}
}

func validProviderFileID(value string) bool {
	if value == "" || len(value) > maxProviderFileIDBytes {
		return false
	}
	for _, character := range value {
		if (character >= 'a' && character <= 'z') || (character >= 'A' && character <= 'Z') ||
			(character >= '0' && character <= '9') || character == '-' || character == '_' {
			continue
		}
		return false
	}
	return true
}

func normalizeResourceKey(value *string) (*string, error) {
	if value == nil {
		return nil, nil
	}
	normalized := strings.TrimSpace(*value)
	if normalized == "" {
		return nil, nil
	}
	if len(normalized) > maxResourceKeyBytes {
		return nil, domain.ErrInvalidInput
	}
	for _, character := range normalized {
		if unicode.IsControl(character) || character == ',' || character == '/' {
			return nil, domain.ErrInvalidInput
		}
	}
	return &normalized, nil
}
