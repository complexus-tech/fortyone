import { act, renderHook } from "@testing-library/react";
import {
  CHAT_VOICE_MAX_BYTES,
  CHAT_VOICE_MAX_SECONDS,
  useChatVoiceRecording,
} from "./use-chat-voice-recording";

class Recorder {
  static instances: Recorder[] = [];
  static isTypeSupported = jest.fn(
    (type: string) => type === "audio/webm;codecs=opus",
  );
  state = "inactive";
  mimeType: string;
  ondataavailable: ((event: BlobEvent) => void) | null = null;
  onstop: (() => void) | null = null;
  onerror: ((event: { error: DOMException }) => void) | null = null;
  start = jest.fn(() => {
    this.state = "recording";
  });
  stop = jest.fn(() => {
    this.state = "inactive";
  });
  constructor(_stream: MediaStream, options: MediaRecorderOptions) {
    this.mimeType = options.mimeType ?? "audio/webm";
    Recorder.instances.push(this);
  }
  data(blob: Blob) {
    this.ondataavailable?.({ data: blob } as BlobEvent);
  }
  finish() {
    this.state = "inactive";
    this.onstop?.();
  }
}
const originalRecorder = globalThis.MediaRecorder;
const originalDevices = Object.getOwnPropertyDescriptor(
  navigator,
  "mediaDevices",
);
const originalCreate = Object.getOwnPropertyDescriptor(URL, "createObjectURL");
const originalRevoke = Object.getOwnPropertyDescriptor(URL, "revokeObjectURL");
const createObjectURL = jest.fn(() => "blob:voice-preview");
const revokeObjectURL = jest.fn();
const release = jest.fn();
const stream = {
  getTracks: () => [{ stop: release }],
} as unknown as MediaStream;
const getUserMedia = jest.fn<Promise<MediaStream>, [MediaStreamConstraints]>();

beforeEach(() => {
  jest.useFakeTimers();
  jest.setSystemTime(new Date("2026-10-17T09:00:00Z"));
  jest.clearAllMocks();
  Recorder.instances = [];
  Recorder.isTypeSupported.mockImplementation(
    (type) => type === "audio/webm;codecs=opus",
  );
  globalThis.MediaRecorder = Recorder as unknown as typeof MediaRecorder;
  Object.defineProperty(navigator, "mediaDevices", {
    configurable: true,
    value: { getUserMedia },
  });
  getUserMedia.mockResolvedValue(stream);
  URL.createObjectURL = createObjectURL;
  URL.revokeObjectURL = revokeObjectURL;
});
afterEach(() => {
  jest.useRealTimers();
});
afterAll(() => {
  globalThis.MediaRecorder = originalRecorder;
  if (originalDevices)
    Object.defineProperty(navigator, "mediaDevices", originalDevices);
  else Reflect.deleteProperty(navigator, "mediaDevices");
  if (originalCreate)
    Object.defineProperty(URL, "createObjectURL", originalCreate);
  else Reflect.deleteProperty(URL, "createObjectURL");
  if (originalRevoke)
    Object.defineProperty(URL, "revokeObjectURL", originalRevoke);
  else Reflect.deleteProperty(URL, "revokeObjectURL");
});

describe("chat voice recording", () => {
  it("requests the microphone only after start and waits for the final chunk before review", async () => {
    const { result, unmount } = renderHook(useChatVoiceRecording);
    expect(getUserMedia).not.toHaveBeenCalled();
    await act(async () => {
      await result.current.start();
    });
    expect(result.current.phase).toBe("recording");
    expect(getUserMedia).toHaveBeenCalledWith({ audio: true });
    const recorder = Recorder.instances[0];
    expect(recorder.start).toHaveBeenCalledWith(1000);
    act(() => {
      jest.advanceTimersByTime(2500);
      result.current.stop();
    });
    expect(result.current.phase).toBe("stopping");
    expect(result.current.preview).toBeNull();
    act(() => {
      recorder.data(new Blob(["last audio chunk"]));
      recorder.finish();
    });
    expect(result.current.phase).toBe("ready");
    expect(result.current.preview?.file.type).toBe("audio/webm;codecs=opus");
    expect(result.current.preview?.file.name).toMatch(/\.webm$/);
    expect(result.current.preview?.durationSeconds).toBe(2.5);
    expect(release).toHaveBeenCalledTimes(1);
    unmount();
    expect(revokeObjectURL).toHaveBeenCalledWith("blob:voice-preview");
  });
  it("uses MP4 when WebM is unavailable", async () => {
    Recorder.isTypeSupported.mockImplementation((type) => type === "audio/mp4");
    const { result } = renderHook(useChatVoiceRecording);
    await act(async () => {
      await result.current.start();
    });
    act(() => {
      result.current.stop();
      Recorder.instances[0].data(new Blob(["aac"]));
      Recorder.instances[0].finish();
    });
    expect(result.current.preview?.file.type).toBe("audio/mp4");
    expect(result.current.preview?.file.name).toMatch(/\.m4a$/);
  });
  it("releases a microphone granted after cancellation without starting capture", async () => {
    let resolve!: (value: MediaStream) => void;
    getUserMedia.mockReturnValue(
      new Promise((done) => {
        resolve = done;
      }),
    );
    const { result } = renderHook(useChatVoiceRecording);
    let starting!: Promise<void>;
    act(() => {
      starting = result.current.start();
    });
    expect(result.current.phase).toBe("requesting");
    act(() => {
      result.current.discard();
    });
    await act(async () => {
      resolve(stream);
      await starting;
    });
    expect(result.current.phase).toBe("idle");
    expect(Recorder.instances).toHaveLength(0);
    expect(release).toHaveBeenCalledTimes(1);
  });
  it("handles permission denial and duplicate start clicks", async () => {
    getUserMedia.mockRejectedValue(
      new DOMException("Denied", "NotAllowedError"),
    );
    const { result } = renderHook(useChatVoiceRecording);
    await act(async () => {
      await Promise.all([result.current.start(), result.current.start()]);
    });
    expect(getUserMedia).toHaveBeenCalledTimes(1);
    expect(result.current.error).toContain("Allow microphone access");
    expect(result.current.phase).toBe("idle");
  });
  it("stops at five minutes without uploading or sending", async () => {
    const { result } = renderHook(useChatVoiceRecording);
    await act(async () => {
      await result.current.start();
    });
    act(() => {
      jest.advanceTimersByTime(CHAT_VOICE_MAX_SECONDS * 1000);
    });
    expect(Recorder.instances[0].stop).toHaveBeenCalledTimes(1);
    expect(result.current.phase).toBe("stopping");
    act(() => {
      Recorder.instances[0].data(new Blob(["audio"]));
      Recorder.instances[0].finish();
    });
    expect(result.current.preview?.durationSeconds).toBe(300);
    expect(result.current.phase).toBe("ready");
  });
  it("discards oversized recordings and releases the stream", async () => {
    const { result } = renderHook(useChatVoiceRecording);
    await act(async () => {
      await result.current.start();
    });
    const chunk = new Blob(["audio"]);
    Object.defineProperty(chunk, "size", { value: CHAT_VOICE_MAX_BYTES + 1 });
    act(() => {
      Recorder.instances[0].data(chunk);
    });
    expect(result.current.preview).toBeNull();
    expect(result.current.error).toContain("20 MB");
    expect(release).toHaveBeenCalledTimes(1);
    expect(createObjectURL).not.toHaveBeenCalled();
  });
  it("stops capture on unmount and times out an unfinished recorder", async () => {
    const first = renderHook(useChatVoiceRecording);
    await act(async () => {
      await first.result.current.start();
    });
    first.unmount();
    expect(Recorder.instances[0].stop).toHaveBeenCalledTimes(1);
    expect(release).toHaveBeenCalledTimes(1);
    const second = renderHook(useChatVoiceRecording);
    await act(async () => {
      await second.result.current.start();
    });
    act(() => {
      second.result.current.stop();
      jest.advanceTimersByTime(5000);
    });
    expect(second.result.current.phase).toBe("idle");
    expect(second.result.current.error).toContain("could not be finished");
    expect(release).toHaveBeenCalledTimes(2);
  });
});
