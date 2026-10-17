package attachments

import (
	"bytes"
	"context"
	"errors"
	"image"
	"image/color"
	_ "image/gif"
	"image/jpeg"
	"os"
	"os/exec"
	"path/filepath"

	"golang.org/x/image/draw"
	_ "golang.org/x/image/webp"
)

const thumbnailDimension = 840

func renderAttachmentThumbnail(ctx context.Context, data []byte, mimeType string) ([]byte, error) {
	if len(data) == 0 || len(data) > thumbnailMaxSourceBytes {
		return nil, ErrThumbnailUnavailable
	}
	if mimeType == "application/pdf" {
		return renderPDFThumbnail(ctx, data)
	}
	config, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil || config.Width < 1 || config.Height < 1 ||
		int64(config.Width)*int64(config.Height) > maxImagePixels {
		return nil, ErrThumbnailUnavailable
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	source, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, ErrThumbnailUnavailable
	}
	width, height, _ := constrainedDimensions(config.Width, config.Height, thumbnailDimension)
	target := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(target, target.Bounds(), image.NewUniform(color.White), image.Point{}, draw.Src)
	draw.CatmullRom.Scale(target, target.Bounds(), source, source.Bounds(), draw.Over, nil)
	var output bytes.Buffer
	if err := jpeg.Encode(&output, target, &jpeg.Options{Quality: 80}); err != nil {
		return nil, err
	}
	if output.Len() > ThumbnailMaxBytes {
		return nil, ErrThumbnailUnavailable
	}
	return output.Bytes(), ctx.Err()
}

func renderPDFThumbnail(ctx context.Context, data []byte) ([]byte, error) {
	if !bytes.HasPrefix(bytes.TrimSpace(data), []byte("%PDF-")) {
		return nil, ErrThumbnailUnavailable
	}
	directory, err := os.MkdirTemp("", "fortyone-thumbnail-*")
	if err != nil {
		return nil, errors.New("create thumbnail workspace")
	}
	defer os.RemoveAll(directory)
	input := filepath.Join(directory, "source.pdf")
	if err := os.WriteFile(input, data, 0600); err != nil {
		return nil, errors.New("prepare thumbnail source")
	}
	output := filepath.Join(directory, "page")
	// Poppler runs as the unprivileged worker user. Only page one is decoded,
	// raster dimensions are capped, and no shell/user-derived argument is used.
	command := exec.CommandContext(ctx, "pdftoppm", "-f", "1", "-l", "1", "-singlefile", "-scale-to", "840",
		"-jpeg", "-jpegopt", "quality=80", input, output)
	if err := command.Run(); err != nil {
		return nil, errors.New("PDF thumbnail renderer unavailable or document could not be rendered")
	}
	path := output + ".jpg"
	info, err := os.Stat(path)
	if err != nil || info.Size() < 1 || info.Size() > ThumbnailMaxBytes {
		return nil, ErrThumbnailUnavailable
	}
	return os.ReadFile(path)
}
