import assert from "node:assert/strict";
import { createRequire } from "node:module";
import test from "node:test";

const require = createRequire(import.meta.url);
const { configureVoiceAudio } = require("./with-maya-voice.js") as {
  configureVoiceAudio: (contents: string) => string;
};

const appDelegate = `import Expo
import React
class AppDelegate: ExpoAppDelegate {
  public override func application(
    _ application: UIApplication,
    didFinishLaunchingWithOptions launchOptions: [UIApplication.LaunchOptionsKey: Any]? = nil
  ) -> Bool {
    let delegate = ReactNativeDelegate()
    return super.application(application, didFinishLaunchingWithOptions: launchOptions)
  }
}`;

test("voice defaults are configured before React starts without activating audio", () => {
  const result = configureVoiceAudio(appDelegate);
  assert.match(result, /^import WebRTC\n/);
  assert.match(result, /categoryOptions\.insert\(\.defaultToSpeaker\)/);
  assert.ok(
    result.indexOf("RTCAudioSessionConfiguration.setWebRTC") <
      result.indexOf("let delegate"),
  );
  // Preserve WebRTC's other options and let it own the active session lifecycle.
  assert.doesNotMatch(
    result,
    /categoryOptions\s*=|setActive|overrideOutputAudioPort/,
  );
});

test("repeated native prebuilds do not duplicate the voice configuration", () => {
  const once = configureVoiceAudio(appDelegate);
  assert.equal(configureVoiceAudio(once), once);
});

test("unsupported AppDelegate templates fail instead of silently omitting routing", () => {
  assert.throws(() => configureVoiceAudio("import Expo"), /launch method/);
});
