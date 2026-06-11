package attachments

import (
	"bytes"
	"image"
	"image/jpeg"
	"image/png"

	_ "image/jpeg"
	_ "image/png"

	"golang.org/x/image/draw"
)

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
	switch contentType {
	case "image/jpeg", "image/png":
		return true
	default:
		return false
	}
}
