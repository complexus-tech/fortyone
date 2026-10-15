import AppKit
import Foundation

let iconDirectory = URL(fileURLWithPath: #filePath)
    .deletingLastPathComponent().deletingLastPathComponent()
    .appendingPathComponent("src-tauri/icons")
let source = iconDirectory.appendingPathComponent("tray-template.svg")
let destination = iconDirectory.appendingPathComponent("tray-template.png")

// Rasterize the existing vector mark at 2× for the 18-point macOS status item.
// Its transparent canvas is separate from the background used by the Dock icon.
guard let image = NSImage(contentsOf: source),
      let bitmap = NSBitmapImageRep(
        bitmapDataPlanes: nil, pixelsWide: 36, pixelsHigh: 36,
        bitsPerSample: 8, samplesPerPixel: 4, hasAlpha: true,
        isPlanar: false, colorSpaceName: .deviceRGB, bytesPerRow: 0, bitsPerPixel: 0
      ),
      let context = NSGraphicsContext(bitmapImageRep: bitmap) else {
    fatalError("Could not render the menu-bar logo")
}
NSGraphicsContext.saveGraphicsState()
NSGraphicsContext.current = context
context.imageInterpolation = .high
image.draw(in: NSRect(x: 0, y: 0, width: 36, height: 36))
NSGraphicsContext.restoreGraphicsState()
guard let png = bitmap.representation(using: .png, properties: [:]) else {
    fatalError("Could not encode the menu-bar logo")
}
try png.write(to: destination, options: .atomic)
print("Prepared transparent 36×36 menu-bar logo")
