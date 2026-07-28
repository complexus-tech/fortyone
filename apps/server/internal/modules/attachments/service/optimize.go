package attachments

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"strings"

	_ "image/jpeg"
	_ "image/png"

	"golang.org/x/image/draw"
)

const maxImagePixels = 32_000_000

type imageOptimizationPolicy struct {
	MaxDimension int
	JPEGQuality  int
}

type optimizedImage struct {
	Data        []byte
	ContentType string
}

var (
	attachmentImagePolicy = imageOptimizationPolicy{
		MaxDimension: 1920,
		JPEGQuality:  78,
	}
	avatarImagePolicy = imageOptimizationPolicy{
		MaxDimension: 1024,
		JPEGQuality:  76,
	}
)

func optimizeImageBytes(data []byte, contentType string, policy imageOptimizationPolicy) (optimizedImage, bool) {
	if !isOptimizableImageType(contentType) {
		return optimizedImage{}, false
	}

	config, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil || config.Width <= 0 || config.Height <= 0 || config.Width*config.Height > maxImagePixels {
		return optimizedImage{}, false
	}

	source, format, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return optimizedImage{}, false
	}

	bounds := source.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	targetWidth, targetHeight, resized := constrainedDimensions(width, height, policy.MaxDimension)
	if resized {
		target := image.NewNRGBA(image.Rect(0, 0, targetWidth, targetHeight))
		draw.CatmullRom.Scale(target, target.Bounds(), source, bounds, draw.Over, nil)
		source = target
	}

	var output bytes.Buffer
	switch format {
	case "jpeg":
		if err := jpeg.Encode(&output, source, &jpeg.Options{Quality: policy.JPEGQuality}); err != nil {
			return optimizedImage{}, false
		}
		contentType = "image/jpeg"
	case "png":
		encoder := png.Encoder{CompressionLevel: png.BestCompression}
		if err := encoder.Encode(&output, source); err != nil {
			return optimizedImage{}, false
		}
		contentType = "image/png"
	default:
		return optimizedImage{}, false
	}

	if output.Len() >= len(data) {
		return optimizedImage{}, false
	}

	return optimizedImage{
		Data:        output.Bytes(),
		ContentType: contentType,
	}, true
}

// OptimizeStoredAttachment compresses an uploaded JPEG or PNG in place.
func (s *Service) OptimizeStoredAttachment(ctx context.Context, blobName string) error {
	blobName = strings.TrimSpace(blobName)
	if blobName == "" {
		return fmt.Errorf("%w: blob name is required", ErrImageOptimizationSkipped)
	}

	attachment, err := s.repo.GetAttachmentByBlobName(ctx, blobName)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return ErrImageOptimizationSkipped
		}
		return fmt.Errorf("get attachment for optimization: %w", err)
	}

	if !isOptimizableImageType(normalizeContentType(attachment.MimeType)) {
		return ErrImageOptimizationNotApplicable
	}

	data, storageContentType, err := s.storage.DownloadFile(ctx, s.config.AttachmentsBucket, blobName)
	if err != nil {
		return fmt.Errorf("download attachment for optimization: %w", err)
	}
	if len(data) == 0 {
		return ErrImageOptimizationSkipped
	}

	contentType := normalizeContentType(storageContentType)
	if contentType == "" || contentType == "application/octet-stream" {
		contentType = normalizeContentType(attachment.MimeType)
	}

	finalData := data
	finalContentType := contentType
	optimized, optimizedOK := optimizeImageBytes(data, contentType, attachmentImagePolicy)
	if optimizedOK {
		finalData = optimized.Data
		finalContentType = optimized.ContentType
		if _, err := s.storage.UploadFile(
			ctx,
			s.config.AttachmentsBucket,
			blobName,
			bytes.NewReader(finalData),
			finalContentType,
		); err != nil {
			return fmt.Errorf("replace attachment with optimized image: %w", err)
		}
	}

	if err := s.repo.UpdateAttachmentStorageMetadata(
		ctx,
		blobName,
		int64(len(finalData)),
		finalContentType,
	); err != nil {
		if errors.Is(err, ErrNotFound) {
			return ErrImageOptimizationSkipped
		}
		return fmt.Errorf("update optimized attachment metadata: %w", err)
	}

	if !optimizedOK {
		return ErrImageOptimizationSkipped
	}

	return nil
}

func normalizeContentType(contentType string) string {
	return strings.ToLower(strings.TrimSpace(strings.Split(contentType, ";")[0]))
}

func constrainedDimensions(width, height, maxDimension int) (int, int, bool) {
	if maxDimension <= 0 || (width <= maxDimension && height <= maxDimension) {
		return width, height, false
	}

	if width >= height {
		targetWidth := maxDimension
		targetHeight := max(1, height*maxDimension/width)
		return targetWidth, targetHeight, true
	}

	targetHeight := maxDimension
	targetWidth := max(1, width*maxDimension/height)
	return targetWidth, targetHeight, true
}

func isOptimizableImageType(contentType string) bool {
	switch normalizeContentType(contentType) {
	case "image/jpeg", "image/png":
		return true
	default:
		return false
	}
}
