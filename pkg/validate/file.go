package validate

import (
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
	MaxAttachmentSize    = 10 * 1024 * 1024 // 10MB
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
	}
)

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

	// Allow both images and documents
	if !allowedImageTypes[mimeType] && !allowedDocumentTypes[mimeType] {
		return ErrInvalidFileType
	}

	return nil
}

// GenerateFileName generates a unique filename while preserving the original extension
func GenerateFileName(originalFilename string) string {
	ext := filepath.Ext(originalFilename)
	return strings.ToLower(uuid.New().String() + ext)
}
