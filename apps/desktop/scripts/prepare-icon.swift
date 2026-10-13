import CoreGraphics
import Foundation
import ImageIO
import UniformTypeIdentifiers

struct IconPreparationError: LocalizedError {
    let message: String
    var errorDescription: String? { message }
}

func prepareIcon() throws {
    let desktopDirectory = URL(fileURLWithPath: #filePath)
        .deletingLastPathComponent().deletingLastPathComponent()
    let iconDirectory = desktopDirectory.appendingPathComponent("src-tauri/icons")
    let sourceURL = iconDirectory.appendingPathComponent("source.png")
    let outputURL = iconDirectory.appendingPathComponent("source-cropped.png")
    let canvasSize = 256
    let inset = 24

    guard let source = CGImageSourceCreateWithURL(sourceURL as CFURL, nil),
          let original = CGImageSourceCreateImageAtIndex(source, 0, nil),
          original.width == canvasSize, original.height == canvasSize else {
        throw IconPreparationError(message: "Expected the original 256×256 source.png artwork")
    }

    let cropRect = CGRect(
        x: inset, y: inset,
        width: canvasSize - 2 * inset, height: canvasSize - 2 * inset
    )
    guard let cropped = original.cropping(to: cropRect),
          let context = CGContext(
            data: nil, width: canvasSize, height: canvasSize,
            bitsPerComponent: 8, bytesPerRow: canvasSize * 4,
            space: CGColorSpaceCreateDeviceRGB(),
            bitmapInfo: CGImageAlphaInfo.premultipliedLast.rawValue
                | CGBitmapInfo.byteOrder32Big.rawValue
          ) else {
        throw IconPreparationError(message: "Could not create the cropped icon canvas")
    }

    let canvas = CGRect(x: 0, y: 0, width: canvasSize, height: canvasSize)
    context.interpolationQuality = .high
    context.draw(cropped, in: canvas)
    // The crop is opaque black at its edges. Restore the original tile's alpha
    // silhouette without shrinking the enlarged white mark or redrawing it.
    context.setBlendMode(.destinationIn)
    context.draw(original, in: canvas)

    guard let prepared = context.makeImage(),
          let destination = CGImageDestinationCreateWithURL(
            outputURL as CFURL, UTType.png.identifier as CFString, 1, nil
          ) else {
        throw IconPreparationError(message: "Could not write source-cropped.png")
    }
    CGImageDestinationAddImage(destination, prepared, nil)
    guard CGImageDestinationFinalize(destination) else {
        throw IconPreparationError(message: "Could not finish writing source-cropped.png")
    }
    print("Prepared \(outputURL.path): 208×208 crop, original 256×256 alpha silhouette")
}

do {
    try prepareIcon()
} catch {
    fputs("Icon preparation failed: \(error.localizedDescription)\n", stderr)
    exit(EXIT_FAILURE)
}
