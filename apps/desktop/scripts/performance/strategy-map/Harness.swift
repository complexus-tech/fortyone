import AppKit
import Foundation
import WebKit

let directory = CommandLine.arguments[1]
let app = NSApplication.shared
app.setActivationPolicy(.regular)
final class Handler: NSObject, WKScriptMessageHandler, NSWindowDelegate {
  var events: [[String: Any]] = []
  func save() {
    if let data = try? JSONSerialization.data(
      withJSONObject: events, options: [.prettyPrinted, .sortedKeys])
    {
      try? data.write(to: URL(fileURLWithPath: directory + "/results.json"))
    }
  }
  func userContentController(
    _ userContentController: WKUserContentController, didReceive message: WKScriptMessage
  ) {
    guard let event = message.body as? [String: Any] else { return }
    events.append(event)
    let kind = event["type"] as? String ?? "unknown"
    print(kind)
    fflush(stdout)
    save()
    if kind == "complete" || kind == "fatal" {
      DispatchQueue.main.asyncAfter(deadline: .now() + 0.2) { app.terminate(nil) }
    }
  }
  func windowWillClose(_ notification: Notification) {
    save()
    app.terminate(nil)
  }
}
let handler = Handler()
let config = WKWebViewConfiguration()
config.websiteDataStore = .nonPersistent()
config.userContentController.add(handler, name: "result")
config.userContentController.addUserScript(
  WKUserScript(
    source:
      "window.addEventListener('error',e=>window.webkit.messageHandlers.result.postMessage({type:'fatal',error:e.message}));",
    injectionTime: .atDocumentStart, forMainFrameOnly: true))
let view = WKWebView(frame: NSRect(x: 0, y: 0, width: 1440, height: 850), configuration: config)
let window = NSWindow(
  contentRect: view.frame, styleMask: [.titled, .closable], backing: .buffered, defer: false)
window.title = "FortyOne synthetic Strategy Map motion check"
window.contentView = view
window.delegate = handler
window.center()
window.makeKeyAndOrderFront(nil)
app.activate(ignoringOtherApps: true)
view.loadFileURL(
  URL(fileURLWithPath: directory + "/index.html"),
  allowingReadAccessTo: URL(fileURLWithPath: directory))
DispatchQueue.main.asyncAfter(deadline: .now() + 90) {
  handler.events.append(["type": "timeout"])
  handler.save()
  app.terminate(nil)
}
app.run()
