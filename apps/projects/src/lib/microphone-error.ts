export const getMicrophoneErrorDescription = (error: unknown): string => {
  const name = error instanceof Error ? error.name : "";

  switch (name) {
    case "NotAllowedError":
    case "SecurityError":
      return "Allow microphone access in your browser or system settings, then try again.";
    case "NotFoundError":
    case "OverconstrainedError":
      return "No available microphone was found. Connect a microphone and try again.";
    case "NotReadableError":
    case "AbortError":
      return "Your microphone could not be opened. Check whether another app is using it, then try again.";
    case "NotSupportedError":
      return "Audio recording is not supported in this app or browser.";
    default:
      return "Check that your microphone is connected and available, then try again.";
  }
};
