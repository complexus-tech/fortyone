import AppKit
import Darwin
import Foundation
import WebKit

guard CommandLine.arguments.count == 2 else {
  fputs("Expected a generated fixture directory. Use run.mjs --run.\n", stderr)
  exit(1)
}
let directory = CommandLine.arguments[1]
var timebase = mach_timebase_info_data_t()
guard mach_timebase_info(&timebase) == KERN_SUCCESS else {
  fputs("Cannot read the Mach clock timebase.\n", stderr)
  exit(1)
}
let nativeMetadata: [String: Any] = [
  "timebase": ["numer": timebase.numer, "denom": timebase.denom]
]
let metadataData = try JSONSerialization.data(
  withJSONObject: nativeMetadata, options: [.sortedKeys])
try metadataData.write(to: URL(fileURLWithPath: "\(directory)/native-metadata.json"))
FileManager.default.createFile(atPath: "\(directory)/run.log", contents: nil)
let logFile = try FileHandle(forWritingTo: URL(fileURLWithPath: "\(directory)/run.log"))
func log(_ value: String) {
  print(value)
  fflush(stdout)
  try? logFile.write(contentsOf: Data((value + "\n").utf8))
}
func read(_ name: String) throws -> String {
  try String(contentsOfFile: "\(directory)/\(name)", encoding: .utf8)
}
func webProcesses() -> Set<Int32> {
  let task = Process()
  task.executableURL = URL(fileURLWithPath: "/bin/ps")
  task.arguments = ["-axo", "pid=,comm="]
  let pipe = Pipe()
  task.standardOutput = pipe
  do {
    try task.run()
  } catch {
    log("Cannot identify WebContent process; CPU samples unavailable: \(error)")
    return []
  }
  let data = pipe.fileHandleForReading.readDataToEndOfFile()
  task.waitUntilExit()
  let text = String(data: data, encoding: .utf8) ?? ""
  return Set(
    text.split(separator: "\n").filter { $0.contains("/com.apple.WebKit.WebContent") }.compactMap {
      Int32($0.trimmingCharacters(in: .whitespaces).split(separator: " ").first ?? "")
    })
}
func cpuTime(_ pid: Int32) -> Double? {
  var usage = rusage_info_v2()
  let status = withUnsafeMutablePointer(to: &usage) { ptr in
    ptr.withMemoryRebound(to: rusage_info_t?.self, capacity: 1) {
      proc_pid_rusage(pid, RUSAGE_INFO_V2, $0)
    }
  }
  // proc_pid_rusage reports Mach ticks, not nanoseconds on Apple silicon.
  return status == 0
    ? Double(usage.ri_user_time + usage.ri_system_time) * Double(timebase.numer)
      / Double(timebase.denom) / 1_000_000 : nil
}
let baselinePids = webProcesses()
let app = NSApplication.shared
app.setActivationPolicy(.regular)
app.finishLaunching()
final class Handler: NSObject, WKScriptMessageHandler, NSWindowDelegate {
  var onComplete: (() -> Void)?
  var complete = false
  var result: [[String: Any]] = []
  var processID: Int32?
  var starts: [String: Double] = [:]
  func persist() {
    do {
      let data = try JSONSerialization.data(
        withJSONObject: result, options: [.prettyPrinted, .sortedKeys])
      try data.write(to: URL(fileURLWithPath: "\(directory)/events.json"), options: .atomic)
    } catch {
      log("Cannot save measurement events: \(error)")
      exit(1)
    }
  }
  func windowWillClose(_ notification: Notification) {
    persist()
    app.terminate(nil)
  }
  func userContentController(
    _ userContentController: WKUserContentController, didReceive message: WKScriptMessage
  ) {
    guard var payload = message.body as? [String: Any], let type = payload["type"] as? String else {
      return
    }
    if processID == nil {
      let new = webProcesses().subtracting(baselinePids)
      if new.count == 1 { processID = new.first }
    }
    if type == "phase-start", let name = payload["name"] as? String, let pid = processID {
      starts[name] = cpuTime(pid)
    }
    if type == "phase-end", var value = payload["result"] as? [String: Any],
      let name = value["name"] as? String, let pid = processID, let before = starts[name],
      let after = cpuTime(pid)
    {
      value["webContentCpuMs"] = after - before
      payload["result"] = value
    }
    payload["webContentPid"] = processID
    result.append(payload)
    if type == "ready" || type == "fatal" { log("EVENT \(type) \(payload)") }
    if type == "trial-end" || type == "fatal" {
      complete = true
      onComplete?()
    }
  }
}
let handler = Handler()
let config = WKWebViewConfiguration()
config.websiteDataStore = .nonPersistent()
config.userContentController.addUserScript(
  WKUserScript(
    source:
      "window.addEventListener('error',event=>window.webkit.messageHandlers.result.postMessage({type:'fatal',error:event.message}));",
    injectionTime: .atDocumentStart, forMainFrameOnly: true))
config.userContentController.add(handler, name: "result")
let view = WKWebView(frame: NSRect(x: 0, y: 0, width: 1440, height: 850), configuration: config)
let window = NSWindow(
  contentRect: view.frame, styleMask: [.titled, .closable, .miniaturizable], backing: .buffered,
  defer: false)
window.title = "FortyOne isolated squircle performance benchmark"
window.delegate = handler
window.contentView = view
window.level = .floating
window.center()
window.makeKeyAndOrderFront(nil)
app.activate(ignoringOtherApps: true)
struct Trial: Codable {
  let id: String
  let mode: String
  let cards: Int
  let warmup: Bool
  let pair: Int?
}
let plan = try JSONDecoder().decode([Trial].self, from: Data(try read("plan.json").utf8))
let fixtureCSS = try read("fixture.css")
let fallbackCSS = try read("fallback.css")
let library = try read("hyperellipse-browser.js")
let fixture = try read("fixture.js")
let header =
  "<!doctype html><html class=dark><head><meta charset=utf-8><style>\(fixtureCSS)</style><style>\(fallbackCSS)</style></head><body><header class=toolbar><strong>FortyOne performance fixture</strong><span>Work</span><span>Search</span><span>Profile</span></header><div class=layout><aside>Workspace<br>My work<br>Stories<br>Roadmap<br>Settings</aside><main id=board></main></div>"

var nextTrial = 0
var currentID = ""
func finishTrial() {
  log("END \(currentID) \(Date().timeIntervalSince1970)")
  handler.persist()
  if handler.result.last?["type"] as? String == "fatal" {
    app.terminate(nil)
    return
  }
  DispatchQueue.main.asyncAfter(deadline: .now() + 0.4) { startTrial() }
}
func startTrial() {
  guard nextTrial < plan.count else {
    log("BENCHMARK_COMPLETE")
    app.terminate(nil)
    return
  }
  let trial = plan[nextTrial]
  nextTrial += 1
  handler.complete = false
  handler.starts = [:]
  currentID = trial.id
  let json: String
  do {
    json = String(decoding: try JSONEncoder().encode(trial), as: UTF8.self)
  } catch {
    log("Cannot encode trial configuration: \(error)")
    handler.persist()
    app.terminate(nil)
    return
  }
  let html =
    header
    + "<script>const CONFIG=\(json);</script><script>\(library)</script><script>\(fixture)</script></body></html>"
  log(
    "START \(currentID) \(Date().timeIntervalSince1970) visible=\(window.isVisible) active=\(app.isActive) occluded=\(window.occlusionState.rawValue)"
  )
  view.loadHTMLString(html, baseURL: URL(fileURLWithPath: directory, isDirectory: true))
  let id = currentID
  DispatchQueue.main.asyncAfter(deadline: .now() + 45) {
    if currentID == id && !handler.complete {
      log("TIMEOUT \(id)")
      handler.persist()
      app.terminate(nil)
    }
  }
}
handler.onComplete = finishTrial
DispatchQueue.main.asyncAfter(deadline: .now() + 0.5) { startTrial() }
// A normal application event loop is required for native WKWebView frame scheduling.
app.run()
