package validate

import (
	"archive/zip"
	"bytes"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

const (
	// Maximum file sizes
	MaxProfileImageSize  = 6 * 1024 * 1024  // 6MB
	MaxWorkspaceLogoSize = 6 * 1024 * 1024  // 6MB
	MaxAttachmentSize    = 25 * 1024 * 1024 // 25MB
)

var (
	ErrFileTooLarge    = errors.New("file size exceeds maximum allowed size")
	ErrInvalidFileType = errors.New("invalid file type")
	ErrFileNameTooLong = errors.New("filename is too long")
	ErrEmptyFile       = errors.New("file is empty")

	// Allowed image MIME types
	allowedImageTypes = map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
		"image/gif":  true,
		"image/webp": true,
	}

	// Allowed document MIME types
	allowedDocumentTypes = map[string]bool{
		"application/pdf":    true,
		"application/msword": true,
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document": true,
		"application/vnd.ms-excel": true,
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":         true,
		"application/vnd.ms-powerpoint":                                             true,
		"application/vnd.openxmlformats-officedocument.presentationml.presentation": true,
		"text/plain": true,
		"text/csv":   true,
		"video/mp4":  true,
	}
)

type officeOpenXMLType struct {
	mimeType       string
	requiredPrefix string
}

var officeOpenXMLTypes = map[string]officeOpenXMLType{
	".docx": {
		mimeType:       "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		requiredPrefix: "word/",
	},
	".xlsx": {
		mimeType:       "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		requiredPrefix: "xl/",
	},
	".pptx": {
		mimeType:       "application/vnd.openxmlformats-officedocument.presentationml.presentation",
		requiredPrefix: "ppt/",
	},
}

var legacyOfficeTypes = map[string]string{
	".doc": "application/msword",
	".xls": "application/vnd.ms-excel",
	".ppt": "application/vnd.ms-powerpoint",
}

// ProfileImage validates a file for use as a profile image
func ProfileImage(file multipart.File, fileHeader *multipart.FileHeader) error {
	// Check file size
	if fileHeader.Size > MaxProfileImageSize {
		return ErrFileTooLarge
	}

	// Check if file is empty
	if fileHeader.Size == 0 {
		return ErrEmptyFile
	}

	// Check filename length
	if len(fileHeader.Filename) > 255 {
		return ErrFileNameTooLong
	}

	// Get MIME type
	buffer := make([]byte, 512)
	_, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return err
	}

	// Reset file position
	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}

	// Check MIME type
	mimeType := http.DetectContentType(buffer)
	if !allowedImageTypes[mimeType] {
		return ErrInvalidFileType
	}

	return nil
}

// WorkspaceLogo validates a file for use as a workspace logo
func WorkspaceLogo(file multipart.File, fileHeader *multipart.FileHeader) error {
	// Use same validation as profile image
	return ProfileImage(file, fileHeader)
}

// Attachment validates a file for use as a story attachment
func Attachment(file multipart.File, fileHeader *multipart.FileHeader) error {
	// Check file size
	if fileHeader.Size > MaxAttachmentSize {
		return ErrFileTooLarge
	}

	// Check if file is empty
	if fileHeader.Size == 0 {
		return ErrEmptyFile
	}

	// Check filename length
	if len(fileHeader.Filename) > 255 {
		return ErrFileNameTooLong
	}

	mimeType, err := AttachmentContentType(file, fileHeader)
	if err != nil {
		return err
	}

	// Allow both images and documents
	if !allowedImageTypes[mimeType] && !allowedDocumentTypes[mimeType] {
		return ErrInvalidFileType
	}

	return nil
}

// AttachmentContentType returns a canonical MIME type based on the file's
// contents. Office Open XML files are ZIP containers, so inspecting only the
// first 512 bytes incorrectly classifies valid DOCX, XLSX, and PPTX uploads.
func AttachmentContentType(file multipart.File, fileHeader *multipart.FileHeader) (string, error) {
	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return "", err
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return "", err
	}

	content := buffer[:n]
	extension := strings.ToLower(filepath.Ext(fileHeader.Filename))

	if officeType, ok := officeOpenXMLTypes[extension]; ok && hasZIPSignature(content) {
		valid, err := hasOfficeOpenXMLStructure(file, fileHeader.Size, officeType.requiredPrefix)
		if err != nil {
			return "", err
		}
		if !valid {
			return "", ErrInvalidFileType
		}
		return officeType.mimeType, nil
	}

	if mimeType, ok := legacyOfficeTypes[extension]; ok && hasLegacyOfficeSignature(content) {
		return mimeType, nil
	}

	mimeType := normalizeContentType(http.DetectContentType(content))
	if extension == ".csv" && mimeType == "text/plain" {
		return "text/csv", nil
	}

	return mimeType, nil
}

func hasOfficeOpenXMLStructure(file multipart.File, size int64, requiredPrefix string) (bool, error) {
	archive, err := zip.NewReader(file, size)
	if err != nil {
		return false, nil
	}

	hasContentTypes := false
	hasDocumentPart := false
	for _, entry := range archive.File {
		switch {
		case entry.Name == "[Content_Types].xml":
			hasContentTypes = true
		case strings.HasPrefix(entry.Name, requiredPrefix):
			hasDocumentPart = true
		}
		if hasContentTypes && hasDocumentPart {
			return true, nil
		}
	}

	return false, nil
}

func hasZIPSignature(content []byte) bool {
	return bytes.HasPrefix(content, []byte{'P', 'K', 0x03, 0x04}) ||
		bytes.HasPrefix(content, []byte{'P', 'K', 0x05, 0x06}) ||
		bytes.HasPrefix(content, []byte{'P', 'K', 0x07, 0x08})
}

func hasLegacyOfficeSignature(content []byte) bool {
	return bytes.HasPrefix(content, []byte{0xD0, 0xCF, 0x11, 0xE0, 0xA1, 0xB1, 0x1A, 0xE1})
}

func normalizeContentType(contentType string) string {
	return strings.ToLower(strings.TrimSpace(strings.Split(contentType, ";")[0]))
}

// GenerateFileName generates a unique filename while preserving the original extension
func GenerateFileName(originalFilename string) string {
	ext := filepath.Ext(originalFilename)
	return strings.ToLower(uuid.New().String() + ext)
}
