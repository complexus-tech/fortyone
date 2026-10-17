import { act, fireEvent, render, screen } from "@testing-library/react";
import { VoiceRecorder } from "./voice-recorder";

const preview = {
  file: new File(["audio"], "note.webm", { type: "audio/webm" }),
  url: "blob:local-preview",
  durationSeconds: 2.5,
};
const mockStart = jest.fn();
const mockDiscard = jest.fn();
const mockPlayback = {
  pause: jest.fn(),
  registerExternalPause: jest.fn(() => jest.fn()),
};
jest.mock("./use-chat-voice-recording", () => ({
  useChatVoiceRecording: () => ({
    phase: "ready",
    elapsedSeconds: 2,
    preview,
    error: null,
    start: mockStart,
    stop: jest.fn(),
    discard: mockDiscard,
  }),
}));
jest.mock("./audio-playback-context", () => ({
  useAudioPlaybackController: () => mockPlayback,
}));

beforeEach(() => {
  jest.clearAllMocks();
  jest.spyOn(HTMLMediaElement.prototype, "pause").mockImplementation(() => {});
});
afterEach(() => {
  jest.restoreAllMocks();
});

describe("voice note review", () => {
  it("starts only on the microphone button and keeps a failed attachment available for retry", async () => {
    const onAttach = jest.fn().mockResolvedValue(false);
    const onActiveChange = jest.fn();
    render(
      <VoiceRecorder
        disabled={false}
        onActiveChange={onActiveChange}
        onAttach={onAttach}
      />,
    );
    expect(mockStart).not.toHaveBeenCalled();
    fireEvent.click(
      screen.getByRole("button", { name: "Record a voice note" }),
    );
    expect(mockStart).toHaveBeenCalledTimes(1);
    expect(mockPlayback.pause).toHaveBeenCalledTimes(1);
    expect(onAttach).not.toHaveBeenCalled();
    expect(onActiveChange).toHaveBeenLastCalledWith(true);
    const audio = screen.getByLabelText("Voice note preview");
    fireEvent.play(audio);
    expect(mockPlayback.pause).toHaveBeenCalledTimes(2);
    await act(async () => {
      fireEvent.click(
        screen.getByRole("button", { name: "Attach voice note" }),
      );
    });
    expect(onAttach).toHaveBeenCalledWith(preview.file, 2.5);
    expect(screen.getByLabelText("Voice note preview")).toBeInTheDocument();
    expect(mockDiscard).not.toHaveBeenCalled();
    onAttach.mockResolvedValue(true);
    await act(async () => {
      fireEvent.click(
        screen.getByRole("button", { name: "Attach voice note" }),
      );
    });
    expect(
      screen.queryByLabelText("Voice note preview"),
    ).not.toBeInTheDocument();
    expect(mockDiscard).toHaveBeenCalledTimes(1);
    expect(onActiveChange).toHaveBeenLastCalledWith(false);
  });
  it("discards the local preview without attaching or sending it", () => {
    const onAttach = jest.fn();
    render(
      <VoiceRecorder
        disabled={false}
        onActiveChange={jest.fn()}
        onAttach={onAttach}
      />,
    );
    fireEvent.click(
      screen.getByRole("button", { name: "Record a voice note" }),
    );
    fireEvent.click(screen.getByRole("button", { name: "Discard" }));
    expect(mockDiscard).toHaveBeenCalledTimes(1);
    expect(onAttach).not.toHaveBeenCalled();
  });
});
