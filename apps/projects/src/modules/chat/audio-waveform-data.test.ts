/* global afterEach, beforeEach, describe, expect, it, jest -- Provided by the projects test runner. */
import { waitFor } from "@testing-library/react";
import { createAudioWaveformStore } from "./audio-waveform-data";

const source = { id: "voice-note", url: "/api/files/voice-note" };
const signal = () => new AbortController().signal;
const deferred = <T>() => {
  let resolve: (value: T) => void = () => undefined;
  const promise = new Promise<T>((finish) => {
    resolve = finish;
  });
  return { promise, resolve };
};
const audioBuffer = (
  channels = [[0.1, -0.5, 0.25, 1]],
  duration = 4,
): AudioBuffer =>
  ({
    duration,
    length: channels[0].length,
    numberOfChannels: channels.length,
    getChannelData: (channel: number) => Float32Array.from(channels[channel]),
  }) as AudioBuffer;

const response = (
  chunks = [new Uint8Array([1, 2, 3])],
  contentLength = "3",
) => {
  let index = 0;
  const reader = {
    read: jest.fn(async () => {
      if (index >= chunks.length) return { done: true };
      const value = chunks[index++];
      return { done: false, value };
    }),
    cancel: jest.fn(async () => undefined),
    releaseLock: jest.fn(),
  };
  const body = {
    getReader: () => reader,
    cancel: jest.fn(async () => undefined),
  };
  return {
    value: {
      ok: true,
      headers: { get: () => contentLength },
      body,
    } as unknown as Response,
    reader,
    body,
  };
};

const originalFetch = globalThis.fetch;
const originalContext = globalThis.OfflineAudioContext;
const fetchMock = jest.fn<Promise<Response>, Parameters<typeof fetch>>();
const decode = jest.fn<Promise<AudioBuffer>, [ArrayBuffer]>();
const context = jest.fn(() => ({ decodeAudioData: decode }));

beforeEach(() => {
  fetchMock.mockReset().mockImplementation(async () => response().value);
  decode.mockReset().mockResolvedValue(audioBuffer());
  context.mockClear();
  Object.defineProperty(globalThis, "fetch", {
    configurable: true,
    value: fetchMock,
  });
  Object.defineProperty(globalThis, "OfflineAudioContext", {
    configurable: true,
    value: context,
  });
});

afterEach(() => {
  Object.defineProperty(globalThis, "fetch", {
    configurable: true,
    value: originalFetch,
  });
  Object.defineProperty(globalThis, "OfflineAudioContext", {
    configurable: true,
    value: originalContext,
  });
  jest.useRealTimers();
  jest.restoreAllMocks();
});

describe("conversation waveform data", () => {
  it("extracts bounded normalized peaks from every audio channel and caches only the result", async () => {
    const firstChannel = Array.from({ length: 96 }, () => 0);
    const secondChannel = [...firstChannel];
    firstChannel[1] = -0.5;
    secondChannel[95] = 1;
    decode.mockResolvedValue(audioBuffer([firstChannel, secondChannel], 12));
    const store = createAudioWaveformStore();
    const data = await store.load(source, signal());
    expect(data?.peaks).toHaveLength(48);
    expect(data?.peaks[0]).toBe(0.5);
    expect(data?.peaks[47]).toBe(1);
    expect(data?.duration).toBe(12);
    expect(context).toHaveBeenCalledWith(1, 1, 8_000);
    expect(fetchMock).toHaveBeenCalledWith(source.url, {
      signal: expect.any(AbortSignal),
      credentials: "same-origin",
    });
    expect(await store.load(source, signal())).toBe(data);
    expect(fetchMock).toHaveBeenCalledTimes(1);
    expect(Object.keys(data ?? {})).toEqual(["peaks", "duration"]);
    store.destroy();
  });

  it("represents silent audio honestly without inventing a waveform", async () => {
    decode.mockResolvedValue(audioBuffer([[0, 0, 0, 0]]));
    const store = createAudioWaveformStore();
    expect((await store.load(source, signal()))?.peaks).toEqual([0, 0, 0, 0]);
    store.destroy();
  });

  it("coalesces duplicate consumers and serializes different fetch/decode jobs", async () => {
    const pendingDecode = deferred<AudioBuffer>();
    decode.mockReturnValueOnce(pendingDecode.promise);
    const store = createAudioWaveformStore();
    const first = store.load(source, signal());
    const duplicate = store.load(source, signal());
    const other = store.load(
      { ...source, url: "/api/files/replaced" },
      signal(),
    );
    await waitFor(() => {
      expect(decode).toHaveBeenCalledTimes(1);
    });
    expect(fetchMock).toHaveBeenCalledTimes(1);
    pendingDecode.resolve(audioBuffer());
    expect(await duplicate).toBe(await first);
    await other;
    expect(fetchMock).toHaveBeenCalledTimes(2);
    store.destroy();
  });

  it("keeps a shared request alive until its last consumer aborts", async () => {
    const pendingDecode = deferred<AudioBuffer>();
    decode.mockReturnValueOnce(pendingDecode.promise);
    const firstController = new AbortController();
    const secondController = new AbortController();
    const store = createAudioWaveformStore();
    const first = store.load(source, firstController.signal);
    const second = store.load(source, secondController.signal);
    await waitFor(() => {
      expect(decode).toHaveBeenCalledTimes(1);
    });
    firstController.abort();
    expect(await first).toBeNull();
    const requestSignal = fetchMock.mock.calls[0][1]?.signal;
    expect(requestSignal?.aborted).toBe(false);
    secondController.abort();
    expect(await second).toBeNull();
    expect(requestSignal?.aborted).toBe(true);
    const retry = store.load(source, signal());
    pendingDecode.resolve(audioBuffer());
    expect(await retry).not.toBeNull();
    expect(fetchMock).toHaveBeenCalledTimes(2);
    store.destroy();
  });

  it("does not download a queued recording after its consumer scrolls away", async () => {
    const pendingDecode = deferred<AudioBuffer>();
    decode.mockReturnValueOnce(pendingDecode.promise);
    const store = createAudioWaveformStore();
    const first = store.load(source, signal());
    const secondController = new AbortController();
    const queued = store.load(
      { id: "other", url: "/other" },
      secondController.signal,
    );
    secondController.abort();
    expect(await queued).toBeNull();
    pendingDecode.resolve(audioBuffer());
    await first;
    expect(fetchMock).toHaveBeenCalledTimes(1);
    store.destroy();
  });

  it("resets on destroy, ignores stale results, and supports a Strict Mode remount", async () => {
    const pendingDecode = deferred<AudioBuffer>();
    decode.mockReturnValueOnce(pendingDecode.promise);
    const store = createAudioWaveformStore();
    const stale = store.load(source, signal());
    await waitFor(() => {
      expect(decode).toHaveBeenCalledTimes(1);
    });
    store.destroy();
    expect(await stale).toBeNull();
    const fresh = store.load(source, signal());
    pendingDecode.resolve(audioBuffer([[1, 1]], 99));
    expect((await fresh)?.duration).toBe(4);
    expect(fetchMock).toHaveBeenCalledTimes(2);
    store.destroy();
  });

  it("limits the cache to 64 entries while retaining recently used recordings", async () => {
    const store = createAudioWaveformStore();
    await Promise.all(
      Array.from({ length: 64 }, (_, index) =>
        store.load({ id: String(index), url: `/${String(index)}` }, signal()),
      ),
    );
    await store.load({ id: "0", url: "/0" }, signal());
    await store.load({ id: "64", url: "/64" }, signal());
    await store.load({ id: "0", url: "/0" }, signal());
    expect(fetchMock).toHaveBeenCalledTimes(65);
    await store.load({ id: "1", url: "/1" }, signal());
    expect(fetchMock).toHaveBeenCalledTimes(66);
    store.destroy();
  });

  it("rejects known long clips before download and unknown long clips after decode", async () => {
    const store = createAudioWaveformStore();
    expect(
      await store.load({ ...source, durationSeconds: 301 }, signal()),
    ).toBeNull();
    expect(fetchMock).not.toHaveBeenCalled();
    decode.mockResolvedValue(audioBuffer([[0.5]], 301));
    expect(
      await store.load({ id: "unknown", url: "/unknown" }, signal()),
    ).toBeNull();
    expect(fetchMock).toHaveBeenCalledTimes(1);
    store.destroy();
  });

  it("rejects advertised oversized downloads without reading and actual oversized streams without decoding", async () => {
    const advertised = response([], String(20 * 1024 * 1024 + 1));
    const oversized = response(
      [new Uint8Array(20 * 1024 * 1024), new Uint8Array(1)],
      "",
    );
    fetchMock
      .mockResolvedValueOnce(advertised.value)
      .mockResolvedValueOnce(oversized.value);
    const store = createAudioWaveformStore();
    expect(await store.load(source, signal())).toBeNull();
    expect(advertised.body.cancel).toHaveBeenCalled();
    expect(advertised.reader.read).not.toHaveBeenCalled();
    expect(
      await store.load({ id: "oversized", url: "/oversized" }, signal()),
    ).toBeNull();
    expect(oversized.reader.cancel).toHaveBeenCalled();
    expect(oversized.reader.releaseLock).toHaveBeenCalled();
    expect(decode).not.toHaveBeenCalled();
    store.destroy();
  });

  it("briefly caches decode failures and retries after the negative cache expires", async () => {
    const now = jest.spyOn(Date, "now").mockReturnValue(1_000);
    decode.mockRejectedValueOnce(new Error("Unsupported media"));
    const store = createAudioWaveformStore();
    expect(await store.load(source, signal())).toBeNull();
    expect(await store.load(source, signal())).toBeNull();
    expect(fetchMock).toHaveBeenCalledTimes(1);
    now.mockReturnValue(61_001);
    expect(await store.load(source, signal())).not.toBeNull();
    expect(fetchMock).toHaveBeenCalledTimes(2);
    store.destroy();
  });

  it("times out a stalled download so the next visible recording can load", async () => {
    jest.useFakeTimers();
    fetchMock.mockImplementationOnce(
      (_url, options) =>
        new Promise((_resolve, reject) => {
          options?.signal?.addEventListener("abort", () => {
            reject(new Error("Aborted"));
          });
        }),
    );
    const store = createAudioWaveformStore();
    const stalled = store.load(source, signal());
    const queued = store.load({ id: "next", url: "/next" }, signal());
    await jest.advanceTimersByTimeAsync(15_000);
    expect(await stalled).toBeNull();
    expect(await queued).not.toBeNull();
    expect(fetchMock.mock.calls[0][1]?.signal?.aborted).toBe(true);
    store.destroy();
  });

  it("skips unsupported browsers and already aborted consumers", async () => {
    Object.defineProperty(globalThis, "OfflineAudioContext", {
      configurable: true,
      value: undefined,
    });
    const store = createAudioWaveformStore();
    expect(await store.load(source, signal())).toBeNull();
    const controller = new AbortController();
    controller.abort();
    expect(
      await store.load({ id: "aborted", url: "/aborted" }, controller.signal),
    ).toBeNull();
    expect(fetchMock).not.toHaveBeenCalled();
    store.destroy();
  });
});
