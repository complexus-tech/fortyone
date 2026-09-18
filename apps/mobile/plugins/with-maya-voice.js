const { withAppDelegate, withInfoPlist } = require("expo/config-plugins");
const withWebRTC = require("@config-plugins/react-native-webrtc").default;

const AUDIO_CONFIGURATION = `    // Maya: prefer speaker without overriding headphones or Bluetooth.
    // WebRTC owns activation and teardown; configure its defaults only.
    let mayaAudioConfiguration = RTCAudioSessionConfiguration.webRTC()
    mayaAudioConfiguration.categoryOptions.insert(.defaultToSpeaker)
    RTCAudioSessionConfiguration.setWebRTC(mayaAudioConfiguration)
`;

function configureVoiceAudio(contents) {
  if (!/^import WebRTC\s*$/m.test(contents)) {
    contents = `import WebRTC\n${contents}`;
  }
  if (contents.includes(AUDIO_CONFIGURATION)) return contents;
  const launchMethod =
    /(didFinishLaunchingWithOptions[\s\S]*?\)\s*->\s*Bool\s*\{\n)/;
  if (!launchMethod.test(contents)) {
    throw new Error("Maya voice: could not locate AppDelegate launch method.");
  }
  return contents.replace(launchMethod, `$1${AUDIO_CONFIGURATION}\n`);
}

module.exports = function withMayaVoice(config) {
  config = withAppDelegate(config, (result) => {
    if (result.modResults.language !== "swift") {
      throw new Error("Maya voice requires a Swift AppDelegate.");
    }
    result.modResults.contents = configureVoiceAudio(
      result.modResults.contents,
    );
    return result;
  });
  // Maya is audio-only. The upstream WebRTC plugin also configures a camera.
  // Plist mods run in reverse registration order, so remove it after WebRTC.
  config = withInfoPlist(config, (result) => {
    delete result.modResults.NSCameraUsageDescription;
    return result;
  });
  return withWebRTC(config, {
    microphonePermission:
      "Allow FortyOne to use your microphone to dictate messages and talk with Maya.",
  });
};

module.exports.configureVoiceAudio = configureVoiceAudio;
