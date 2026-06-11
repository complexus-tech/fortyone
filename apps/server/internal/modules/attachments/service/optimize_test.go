package attachments

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"testing"
)

func TestOptimizeImageBytesResizesAndCompressesJPEG(t *testing.T) {
	source := image.NewRGBA(image.Rect(0, 0, 2400, 1400))
	for y := 0; y < source.Bounds().Dy(); y++ {
		for x := 0; x < source.Bounds().Dx(); x++ {
			source.Set(x, y, color.RGBA{
				R: uint8(x % 255),
				G: uint8(y % 255),
				B: uint8((x + y) % 255),
				A: 255,
			})
		}
	}

	var input bytes.Buffer
	if err := jpeg.Encode(&input, source, &jpeg.Options{Quality: 95}); err != nil {
		t.Fatalf("encode source jpeg: %v", err)
	}

	optimized, ok := optimizeImageBytes(input.Bytes(), "image/jpeg", attachmentImagePolicy)
	if !ok {
		t.Fatal("expected jpeg to be optimized")
	}
	if optimized.ContentType != "image/jpeg" {
		t.Fatalf("expected jpeg content type, got %q", optimized.ContentType)
	}
	if len(optimized.Data) >= input.Len() {
		t.Fatalf("expected optimized image to be smaller: optimized=%d original=%d", len(optimized.Data), input.Len())
	}

	config, _, err := image.DecodeConfig(bytes.NewReader(optimized.Data))
	if err != nil {
		t.Fatalf("decode optimized jpeg config: %v", err)
	}
	if config.Width > attachmentImagePolicy.MaxDimension || config.Height > attachmentImagePolicy.MaxDimension {
		t.Fatalf("optimized dimensions exceed max: %dx%d", config.Width, config.Height)
	}
}

func TestOptimizeImageBytesSkipsUnsupportedImageType(t *testing.T) {
	optimized, ok := optimizeImageBytes([]byte("gif data"), "image/gif", attachmentImagePolicy)
	if ok {
		t.Fatalf("expected unsupported image type to be skipped, got %#v", optimized)
	}
}
