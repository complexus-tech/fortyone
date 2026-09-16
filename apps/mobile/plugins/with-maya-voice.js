const { withInfoPlist } = require("expo/config-plugins");
const withWebRTC = require("@config-plugins/react-native-webrtc").default;

module.exports = function withMayaVoice(config) {
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
