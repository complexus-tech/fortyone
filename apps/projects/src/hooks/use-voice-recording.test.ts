/* global afterEach, beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { act, cleanup, renderHook } from "@testing-library/react";
import { toast } from "sonner";
import { useVoiceRecording } from "./use-voice-recording";

jest.mock("sonner", () => ({ toast: { error: jest.fn() } }));

class MockMediaRecorder {
  static instances: MockMediaRecorder[] = [];
  static supportedTypes = new Set(["audio/webm;codecs=opus"]);
  static defaultMimeType = "audio/mp4";
  static startError: Error | null = null;
  static constructorError: Error | null = null;
  static isTypeSupported = (type: string) => this.supportedTypes.has(type);

  state: RecordingState = "inactive";
  mimeType: string;
  options?: MediaRecorderOptions;
  ondataavailable: ((event: { data: Blob }) => void) | null = null;
  onstop: (() => void) | null = null;
  onerror: ((event: { error: Error }) => void) | null = null;

  constructor(_stream: MediaStream, options?: MediaRecorderOptions) {
    if (MockMediaRecorder.constructorError) {
      throw MockMediaRecorder.constructorError;
    }
    this.options = options;
    this.mimeType = options?.mimeType ?? MockMediaRecorder.defaultMimeType;
    MockMediaRecorder.instances.push(this);
  }

  start = jest.fn(() => {
    if (MockMediaRecorder.startError) throw MockMediaRecorder.startError;
    this.state = "recording";
  });

  stop = jest.fn(() => {
    this.state = "inactive";
  });

  finish(...chunks: string[]) {
    for (const chunk of chunks) {
      this.ondataavailable?.({
        data: new Blob([chunk], { type: this.mimeType }),
      });
    }
    this.state = "inactive";
    this.onstop?.();
  }
}

const mediaRecorderDescriptor = Object.getOwnPropertyDescriptor(
  globalThis,
  "MediaRecorder",
);
const mediaDevicesDescriptor = Object.getOwnPropertyDescriptor(
  navigator,
  "mediaDevices",
);
const getUserMedia = jest.fn();
const stopTrack = jest.fn();
const stream = {
  getTracks: () => [{ stop: stopTrack }],
} as unknown as MediaStream;

beforeEach(() => {
  jest.useFakeTimers({ doNotFake: ["queueMicrotask", "nextTick"] });
  jest.clearAllMocks();
  MockMediaRecorder.instances = [];
  MockMediaRecorder.supportedTypes = new Set(["audio/webm;codecs=opus"]);
  MockMediaRecorder.defaultMimeType = "audio/mp4";
  MockMediaRecorder.startError = null;
  MockMediaRecorder.constructorError = null;
  Object.defineProperty(globalThis, "MediaRecorder", {
    configurable: true,
    value: MockMediaRecorder,
  });
  Object.defineProperty(navigator, "mediaDevices", {
    configurable: true,
    value: { getUserMedia },
  });
  getUserMedia.mockResolvedValue(stream);
});

afterEach(() => {
  cleanup();
  jest.useRealTimers();
  if (mediaRecorderDescriptor) {
    Object.defineProperty(globalThis, "MediaRecorder", mediaRecorderDescriptor);
  } else {
    Reflect.deleteProperty(globalThis, "MediaRecorder");
  }
  if (mediaDevicesDescriptor) {
    Object.defineProperty(navigator, "mediaDevices", mediaDevicesDescriptor);
  } else {
    Reflect.deleteProperty(navigator, "mediaDevices");
  }
});

describe("useVoiceRecording", () => {
  it("waits for final audio data rather than a fixed delay and stops only once", async () => {
    const { result } = renderHook(() => useVoiceRecording());
    await act(async () => result.current.startRecording());
    const recorder = MockMediaRecorder.instances[0];
    expect(recorder.options?.mimeType).toBe("audio/webm;codecs=opus");

    let stopped!: Promise<Blob | null>;
    const onFinished = jest.fn();
    act(() => {
      stopped = result.current.stopRecording();
      void stopped.then(onFinished);
      expect(result.current.stopRecording()).toBe(stopped);
      jest.advanceTimersByTime(1000);
    });
    expect(onFinished).not.toHaveBeenCalled();
    expect(recorder.stop).toHaveBeenCalledTimes(1);
    expect(result.current.recordingState).toBe("processing");

    await act(async () => {
      recorder.finish("first chunk", "final chunk");
      await stopped;
    });
    const blob = await stopped;
    expect(blob?.size).toBe("first chunkfinal chunk".length);
    expect(blob?.type).toBe("audio/webm;codecs=opus");
    expect(stopTrack).toHaveBeenCalledTimes(1);
    expect(jest.getTimerCount()).toBe(0);
  });

  it.each([
    { supported: ["audio/mp4"], expectedOption: "audio/mp4" },
    { supported: [], expectedOption: undefined },
  ])(
    "uses a supported format or the recorder default: $supported",
    async ({ supported, expectedOption }) => {
      MockMediaRecorder.supportedTypes = new Set(supported);
      const { result } = renderHook(() => useVoiceRecording());
      await act(async () => result.current.startRecording());
      const recorder = MockMediaRecorder.instances[0];
      expect(recorder.options?.mimeType).toBe(expectedOption);
      recorder.mimeType = "audio/mp4;codecs=mp4a.40.2";
      let stopped!: Promise<Blob | null>;
      act(() => {
        stopped = result.current.stopRecording();
        recorder.finish("audio");
      });
      expect((await stopped)?.type).toBe("audio/mp4;codecs=mp4a.40.2");
    },
  );

  it("guards duplicate starts while permission is pending and releases a cancelled stream", async () => {
    let resolvePermission!: (value: MediaStream) => void;
    getUserMedia.mockReturnValue(
      new Promise<MediaStream>((resolve) => {
        resolvePermission = resolve;
      }),
    );
    const { result } = renderHook(() => useVoiceRecording());
    let started!: Promise<void>;
    act(() => {
      started = result.current.startRecording();
      void result.current.startRecording();
    });
    expect(getUserMedia).toHaveBeenCalledTimes(1);
    expect(result.current.recordingState).toBe("requesting");

    await act(async () => {
      expect(await result.current.stopRecording()).toBeNull();
      resolvePermission(stream);
      await started;
    });
    expect(result.current.recordingState).toBe("idle");
    expect(stopTrack).toHaveBeenCalledTimes(1);
    expect(MockMediaRecorder.instances).toHaveLength(0);
    expect(toast.error).not.toHaveBeenCalled();
  });

  it("releases a stream that arrives after unmount", async () => {
    let resolvePermission!: (value: MediaStream) => void;
    getUserMedia.mockReturnValue(
      new Promise<MediaStream>((resolve) => {
        resolvePermission = resolve;
      }),
    );
    const { result, unmount } = renderHook(() => useVoiceRecording());
    let started!: Promise<void>;
    act(() => {
      started = result.current.startRecording();
    });
    unmount();
    await act(async () => {
      resolvePermission(stream);
      await started;
    });
    expect(stopTrack).toHaveBeenCalledTimes(1);
    expect(MockMediaRecorder.instances).toHaveLength(0);
  });

  it("keeps a new recording active when a cancelled permission request resolves late", async () => {
    let resolveOldPermission!: (value: MediaStream) => void;
    getUserMedia.mockReturnValueOnce(
      new Promise<MediaStream>((resolve) => {
        resolveOldPermission = resolve;
      }),
    );
    const stopOldTrack = jest.fn();
    const oldStream = {
      getTracks: () => [{ stop: stopOldTrack }],
    } as unknown as MediaStream;
    const { result } = renderHook(() => useVoiceRecording());
    let oldStart!: Promise<void>;
    act(() => {
      oldStart = result.current.startRecording();
      result.current.resetRecording();
    });
    await act(async () => result.current.startRecording());
    await act(async () => {
      resolveOldPermission(oldStream);
      await oldStart;
    });
    expect(result.current.isRecording).toBe(true);
    expect(MockMediaRecorder.instances).toHaveLength(1);
    expect(stopOldTrack).toHaveBeenCalledTimes(1);
    expect(stopTrack).not.toHaveBeenCalled();
  });

  it("stops active capture and clears its timer when unmounted", async () => {
    const { result, unmount } = renderHook(() => useVoiceRecording());
    await act(async () => result.current.startRecording());
    const recorder = MockMediaRecorder.instances[0];
    unmount();
    expect(recorder.stop).toHaveBeenCalledTimes(1);
    expect(stopTrack).toHaveBeenCalledTimes(1);
    expect(jest.getTimerCount()).toBe(0);
  });

  it("cancels capture and resolves any pending stop on unmount", async () => {
    const autoStop = jest.fn();
    const { result, unmount } = renderHook(() => useVoiceRecording(autoStop));
    await act(async () => result.current.startRecording());
    const recorder = MockMediaRecorder.instances[0];
    let stopped!: Promise<Blob | null>;
    act(() => {
      stopped = result.current.stopRecording();
    });
    unmount();
    expect(await stopped).toBeNull();
    expect(stopTrack).toHaveBeenCalledTimes(1);
    expect(recorder.onstop).toBeNull();
    expect(jest.getTimerCount()).toBe(0);
    expect(autoStop).not.toHaveBeenCalled();
  });

  it("automatically stops once and passes the completed Blob to the current callback", async () => {
    const oldCallback = jest.fn();
    const callback = jest.fn();
    const { result, rerender } = renderHook(
      ({ onAutoStop }) => useVoiceRecording(onAutoStop),
      { initialProps: { onAutoStop: oldCallback } },
    );
    await act(async () => result.current.startRecording());
    rerender({ onAutoStop: callback });
    const recorder = MockMediaRecorder.instances[0];
    act(() => {
      jest.advanceTimersByTime(60_000);
    });
    expect(recorder.stop).toHaveBeenCalledTimes(1);
    expect(callback).not.toHaveBeenCalled();
    expect(result.current.recordingDuration).toBe(60);
    act(() => {
      recorder.finish("complete recording");
      jest.advanceTimersByTime(10_000);
    });
    expect(oldCallback).not.toHaveBeenCalled();
    expect(callback).toHaveBeenCalledTimes(1);
    expect(callback.mock.calls[0][0].size).toBe("complete recording".length);
    expect(stopTrack).toHaveBeenCalledTimes(1);
  });

  it.each(["constructor", "start"])(
    "releases the microphone when %s fails",
    async (failure) => {
      const error = new DOMException("Unsupported", "NotSupportedError");
      if (failure === "constructor") MockMediaRecorder.constructorError = error;
      else MockMediaRecorder.startError = error;
      const { result } = renderHook(() => useVoiceRecording());
      await act(async () => result.current.startRecording());
      expect(stopTrack).toHaveBeenCalledTimes(1);
      expect(result.current.recordingState).toBe("idle");
      expect(jest.getTimerCount()).toBe(0);
      expect(toast.error).toHaveBeenCalledWith("Failed to start recording", {
        description: expect.stringContaining("not supported"),
      });
    },
  );

  it.each([
    ["NotAllowedError", "Allow microphone access"],
    ["NotFoundError", "No available microphone"],
    ["NotReadableError", "another app"],
  ])(
    "explains %s without calling it a generic permission failure",
    async (name, description) => {
      getUserMedia.mockRejectedValue(new DOMException("Unavailable", name));
      const { result } = renderHook(() => useVoiceRecording());
      await act(async () => result.current.startRecording());
      expect(result.current.recordingState).toBe("idle");
      expect(toast.error).toHaveBeenCalledWith("Failed to start recording", {
        description: expect.stringContaining(description),
      });
    },
  );

  it("returns to idle without transcribing when no audio is produced", async () => {
    const callback = jest.fn();
    const { result } = renderHook(() => useVoiceRecording(callback));
    await act(async () => result.current.startRecording());
    act(() => {
      jest.advanceTimersByTime(60_000);
      MockMediaRecorder.instances[0].finish();
    });
    expect(callback).not.toHaveBeenCalled();
    expect(result.current.recordingState).toBe("idle");
    expect(stopTrack).toHaveBeenCalledTimes(1);
    expect(toast.error).toHaveBeenCalledWith("No audio was recorded", {
      description: expect.any(String),
    });
  });

  it("releases capture and discards partial audio after a recorder error", async () => {
    const callback = jest.fn();
    const { result } = renderHook(() => useVoiceRecording(callback));
    await act(async () => result.current.startRecording());
    const recorder = MockMediaRecorder.instances[0];
    act(() => {
      recorder.ondataavailable?.({ data: new Blob(["partial"]) });
      recorder.onerror?.({
        error: new DOMException("Unavailable", "NotReadableError"),
      });
    });
    expect(await result.current.stopRecording()).toBeNull();
    expect(result.current.recordingState).toBe("idle");
    expect(stopTrack).toHaveBeenCalledTimes(1);
    expect(callback).not.toHaveBeenCalled();
    expect(jest.getTimerCount()).toBe(0);
    expect(toast.error).toHaveBeenCalledWith("Recording was interrupted", {
      description: expect.stringContaining("another app"),
    });
  });
});
