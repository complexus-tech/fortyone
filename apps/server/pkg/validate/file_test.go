package validate

import (
	"archive/zip"
	"bytes"
	"mime/multipart"
	"testing"
)

type memoryMultipartFile struct {
	*bytes.Reader
}

func (*memoryMultipartFile) Close() error { return nil }

func TestAttachmentContentTypeOfficeOpenXML(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		filename        string
		documentPart    string
		wantContentType string
	}{
		{
			name:            "Word document",
			filename:        "brief.docx",
			documentPart:    "word/document.xml",
			wantContentType: "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		},
		{
			name:            "Excel spreadsheet",
			filename:        "forecast.xlsx",
			documentPart:    "xl/workbook.xml",
			wantContentType: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		},
		{
			name:            "PowerPoint presentation",
			filename:        "update.pptx",
			documentPart:    "ppt/presentation.xml",
			wantContentType: "application/vnd.openxmlformats-officedocument.presentationml.presentation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			content := officeOpenXMLFixture(t, tt.documentPart)
			file := &memoryMultipartFile{Reader: bytes.NewReader(content)}
			header := &multipart.FileHeader{Filename: tt.filename, Size: int64(len(content))}

			contentType, err := AttachmentContentType(file, header)
			if err != nil {
				t.Fatalf("AttachmentContentType() error = %v", err)
			}
			if contentType != tt.wantContentType {
				t.Fatalf("AttachmentContentType() = %q, want %q", contentType, tt.wantContentType)
			}
			if err := Attachment(file, header); err != nil {
				t.Fatalf("Attachment() error = %v", err)
			}
		})
	}
}

func TestAttachmentRejectsRenamedZIP(t *testing.T) {
	t.Parallel()

	content := officeOpenXMLFixture(t, "unrelated/file.txt")
	file := &memoryMultipartFile{Reader: bytes.NewReader(content)}
	header := &multipart.FileHeader{Filename: "not-a-document.docx", Size: int64(len(content))}

	if err := Attachment(file, header); err != ErrInvalidFileType {
		t.Fatalf("Attachment() error = %v, want %v", err, ErrInvalidFileType)
	}
}

func TestAttachmentContentTypeCSV(t *testing.T) {
	t.Parallel()

	content := []byte("name,status\nAda,active\n")
	file := &memoryMultipartFile{Reader: bytes.NewReader(content)}
	header := &multipart.FileHeader{Filename: "people.csv", Size: int64(len(content))}

	contentType, err := AttachmentContentType(file, header)
	if err != nil {
		t.Fatalf("AttachmentContentType() error = %v", err)
	}
	if contentType != "text/csv" {
		t.Fatalf("AttachmentContentType() = %q, want text/csv", contentType)
	}
}

func TestAttachmentWithoutSizeLimitPreservesTypeValidation(t *testing.T) {
	content := []byte("story notes\n")
	header := &multipart.FileHeader{Filename: "notes.txt", Size: MaxAttachmentSize + 1}
	file := &memoryMultipartFile{Reader: bytes.NewReader(content)}
	if err := Attachment(file, header); err != ErrFileTooLarge {
		t.Fatalf("direct attachment error = %v, want %v", err, ErrFileTooLarge)
	}
	if err := AttachmentWithoutSizeLimit(file, header); err != nil {
		t.Fatalf("provider attachment error = %v", err)
	}

	invalid := &memoryMultipartFile{Reader: bytes.NewReader([]byte("<html>not an attachment</html>"))}
	if err := AttachmentWithoutSizeLimit(invalid, &multipart.FileHeader{Filename: "fake.txt", Size: 30}); err != ErrInvalidFileType {
		t.Fatalf("invalid provider attachment error = %v, want %v", err, ErrInvalidFileType)
	}
}

func TestAttachmentContentTypeLegacyOffice(t *testing.T) {
	t.Parallel()

	content := append(
		[]byte{0xD0, 0xCF, 0x11, 0xE0, 0xA1, 0xB1, 0x1A, 0xE1},
		make([]byte, 512)...,
	)
	file := &memoryMultipartFile{Reader: bytes.NewReader(content)}
	header := &multipart.FileHeader{Filename: "brief.doc", Size: int64(len(content))}

	contentType, err := AttachmentContentType(file, header)
	if err != nil {
		t.Fatalf("AttachmentContentType() error = %v", err)
	}
	if contentType != "application/msword" {
		t.Fatalf("AttachmentContentType() = %q, want application/msword", contentType)
	}
}

func officeOpenXMLFixture(t *testing.T, documentPart string) []byte {
	t.Helper()

	var content bytes.Buffer
	archive := zip.NewWriter(&content)
	for _, name := range []string{"[Content_Types].xml", documentPart} {
		entry, err := archive.Create(name)
		if err != nil {
			t.Fatalf("create ZIP entry %q: %v", name, err)
		}
		if _, err := entry.Write([]byte("fixture")); err != nil {
			t.Fatalf("write ZIP entry %q: %v", name, err)
		}
	}
	if err := archive.Close(); err != nil {
		t.Fatalf("close ZIP fixture: %v", err)
	}

	return content.Bytes()
}
