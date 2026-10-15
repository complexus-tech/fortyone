import {
  act,
  fireEvent,
  render,
  screen,
  waitFor,
} from "@testing-library/react";
import ky from "ky";
import type { useMayaRealtimeVoice } from "@/modules/maya/hooks/use-maya-realtime-voice";
import { ChatInput } from "./chat-input";

jest.mock("ky", () => ({ __esModule: true, default: { post: jest.fn() } }));
jest.mock("@/hooks", () => ({
  useTerminology: () => ({ getTermDisplay: () => "tasks" }),
}));
jest.mock("next/navigation", () => ({ usePathname: () => "/test/maya" }));
jest.mock("react-dropzone", () => ({
  useDropzone: () => ({ getInputProps: () => ({}), open: jest.fn() }),
}));
jest.mock("@/modules/story/components/story-attachment-preview", () => ({
  StoryAttachmentPreview: () => null,
}));
const mockRecorder = {
  isRecording: false,
  recordingState: "idle",
  recordingDuration: 0,
  formatDuration: () => "0:00",
  MAX_RECORDING_TIME: 60,
  startRecording: jest.fn(),
  stopRecording: jest.fn(),
  resetRecording: jest.fn(),
};
jest.mock("@/hooks/use-voice-recording", () => ({
  useVoiceRecording: () => mockRecorder,
}));

const voice = {
  status: "idle",
  connect: jest.fn(),
  disconnect: jest.fn(),
} as unknown as ReturnType<typeof useMayaRealtimeVoice>;
const onChange = jest.fn();
const props = {
  attachments: [],
  conversationId: "conversation-a",
  googleDriveFiles: [],
  messagesCount: 0,
  onAttachmentsChange: jest.fn(),
  onChange,
  onGoogleDriveFileRemove: jest.fn(),
  onSend: jest.fn(),
  onStop: jest.fn(),
  realtimeVoice: voice,
  status: "ready" as const,
  value: "Original draft",
};

beforeEach(() => {
  jest.clearAllMocks();
  delete window.__FORTYONE_OFFLINE_PAGE__;
  jest.spyOn(navigator, "onLine", "get").mockReturnValue(true);
  Object.assign(mockRecorder, { isRecording: false, recordingState: "idle" });
});
afterEach(() => {
  jest.restoreAllMocks();
  delete window.__FORTYONE_OFFLINE_PAGE__;
});

it("waits for the final audio and uploads its actual format", async () => {
  let finish!: (blob: Blob) => void;
  mockRecorder.stopRecording.mockReturnValue(
    new Promise((resolve) => {
      finish = resolve;
    }),
  );
  Object.assign(mockRecorder, {
    isRecording: true,
    recordingState: "recording",
  });
  jest.mocked(ky.post).mockReturnValue({
    json: () => Promise.resolve({ text: "spoken words" }),
  } as ReturnType<typeof ky.post>);
  render(<ChatInput {...props} />);
  fireEvent.click(screen.getByRole("button", { name: "Send recording" }));
  expect(ky.post).not.toHaveBeenCalled();
  await act(async () => {
    finish(new Blob(["audio"], { type: "audio/mp4" }));
  });
  await waitFor(() => {
    expect(onChange).toHaveBeenCalled();
  });
  const options = jest.mocked(ky.post).mock.calls[0][1];
  const file = (options?.body as FormData).get("audio") as File;
  expect(file.name).toBe("recording.m4a");
  expect(file.type).toBe("audio/mp4");
  expect(onChange).toHaveBeenCalledWith({
    target: { value: "Original draft spoken words" },
  });
});

it("appends a transcript to the latest draft rather than overwriting edits", async () => {
  let finish!: (result: { text: string }) => void;
  jest.mocked(ky.post).mockReturnValue({
    json: () =>
      new Promise<{ text: string }>((resolve) => {
        finish = resolve;
      }),
  } as ReturnType<typeof ky.post>);
  mockRecorder.stopRecording.mockResolvedValue(
    new Blob(["audio"], { type: "audio/webm" }),
  );
  Object.assign(mockRecorder, {
    isRecording: true,
    recordingState: "recording",
  });
  const view = render(<ChatInput {...props} />);
  fireEvent.click(screen.getByRole("button", { name: "Send recording" }));
  await waitFor(() => {
    expect(ky.post).toHaveBeenCalledTimes(1);
  });
  view.rerender(<ChatInput {...props} value="Edited draft" />);
  await act(async () => {
    finish({ text: "spoken words" });
  });
  expect(onChange).toHaveBeenCalledWith({
    target: { value: "Edited draft spoken words" },
  });
});

it("cancels transcription when switching conversations and ignores a late response", async () => {
  let finish!: (result: { text: string }) => void;
  jest.mocked(ky.post).mockReturnValue({
    json: () =>
      new Promise<{ text: string }>((resolve) => {
        finish = resolve;
      }),
  } as ReturnType<typeof ky.post>);
  mockRecorder.stopRecording.mockResolvedValue(
    new Blob(["audio"], { type: "audio/webm" }),
  );
  Object.assign(mockRecorder, {
    isRecording: true,
    recordingState: "recording",
  });
  const view = render(<ChatInput {...props} />);
  fireEvent.click(screen.getByRole("button", { name: "Send recording" }));
  await waitFor(() => {
    expect(ky.post).toHaveBeenCalledTimes(1);
  });
  const signal = jest.mocked(ky.post).mock.calls[0][1]?.signal;
  view.rerender(<ChatInput {...props} conversationId="conversation-b" />);
  expect(signal?.aborted).toBe(true);
  await act(async () => {
    finish({ text: "old transcript" });
  });
  expect(onChange).not.toHaveBeenCalled();
});

it.each(["requesting", "recording", "processing"])(
  "prevents Live voice while recording is %s",
  (recordingState) => {
    mockRecorder.recordingState = recordingState;
    render(<ChatInput {...props} />);
    expect(screen.getByRole("button", { name: "Live voice" })).toBeDisabled();
  },
);

it("aborts transcription on connection loss and ignores the late response", async () => {
  let finish!: (result: { text: string }) => void;
  jest.mocked(ky.post).mockReturnValue({
    json: () =>
      new Promise<{ text: string }>((resolve) => {
        finish = resolve;
      }),
  } as ReturnType<typeof ky.post>);
  mockRecorder.stopRecording.mockResolvedValue(
    new Blob(["audio"], { type: "audio/webm" }),
  );
  Object.assign(mockRecorder, {
    isRecording: true,
    recordingState: "recording",
  });
  render(<ChatInput {...props} />);
  fireEvent.click(screen.getByRole("button", { name: "Send recording" }));
  await waitFor(() => {
    expect(ky.post).toHaveBeenCalledTimes(1);
  });
  const signal = jest.mocked(ky.post).mock.calls[0][1]?.signal;
  act(() => {
    jest.spyOn(navigator, "onLine", "get").mockReturnValue(false);
    window.dispatchEvent(new Event("offline"));
  });
  expect(signal?.aborted).toBe(true);
  expect(mockRecorder.resetRecording).toHaveBeenCalled();
  await act(async () => {
    finish({ text: "late transcript" });
  });
  expect(onChange).not.toHaveBeenCalled();
});

it("disables both voice starters on a cached offline page", () => {
  window.__FORTYONE_OFFLINE_PAGE__ = true;
  render(<ChatInput {...props} />);
  expect(
    screen.getByRole("button", { name: "Record a voice message" }),
  ).toBeDisabled();
  expect(screen.getByRole("button", { name: "Live voice" })).toBeDisabled();
});

it("allows cancelling an in-flight live connection", () => {
  render(
    <ChatInput {...props} realtimeVoice={{ ...voice, status: "connecting" }} />,
  );
  fireEvent.click(screen.getByRole("button", { name: "Connecting" }));
  expect(voice.disconnect).toHaveBeenCalledTimes(1);
  expect(
    screen.getByRole("button", { name: "Record a voice message" }),
  ).toBeDisabled();
});
