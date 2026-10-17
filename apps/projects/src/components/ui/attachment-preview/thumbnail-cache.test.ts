/** @jest-environment node */ // eslint-disable-line tsdoc/syntax -- Jest environment directive.
/* global afterEach, beforeEach, describe, expect, it, jest -- Jest globals. */
import {
  createThumbnailCache,
  THUMBNAIL_ACCESS_LEASE_MS,
} from "./thumbnail-cache";

const response = (size = 4) =>
  new Response(new Uint8Array(size), {
    headers: { "Content-Type": "image/jpeg", ETag: '"version-1"' },
  });
const signal = () => new AbortController().signal;
let fetchMock: jest.SpyInstance;
beforeEach(() => {
  fetchMock = jest
    .spyOn(globalThis, "fetch")
    .mockImplementation(async () => response());
});
afterEach(() => {
  jest.restoreAllMocks();
});

describe("private thumbnail cache", () => {
  it("coalesces concurrent consumers and reuses pixels after virtualized remount", async () => {
    const cache = createThumbnailCache();
    const [first, second] = await Promise.all([
      cache.load("/one", signal()),
      cache.load("/one", signal()),
    ]);
    const remount = await cache.load("/one", signal());
    expect(fetchMock).toHaveBeenCalledTimes(1);
    expect(first.blob).toBe(second.blob);
    expect(remount.blob).toBe(first.blob);
  });

  it("rechecks permissions with ETag after the access lease without downloading unchanged pixels", async () => {
    const cache = createThumbnailCache();
    let now = 1_000;
    jest.spyOn(Date, "now").mockImplementation(() => now);
    const original = await cache.load("/one", signal());
    now += THUMBNAIL_ACCESS_LEASE_MS;
    fetchMock.mockResolvedValueOnce(new Response(null, { status: 304 }));
    const revalidated = await cache.load("/one", signal());
    expect(revalidated.blob).toBe(original.blob);
    expect(revalidated.checkedAt).toBe(now);
    expect(fetchMock).toHaveBeenLastCalledWith(
      "/one",
      expect.objectContaining({
        headers: { "If-None-Match": '"version-1"' },
        cache: "no-store",
      }),
    );
  });

  it("drops cached bytes on revoked access and does not serve a stale image", async () => {
    const cache = createThumbnailCache();
    await cache.load("/one", signal());
    fetchMock.mockResolvedValueOnce(new Response(null, { status: 403 }));
    await expect(cache.load("/one", signal(), false, true)).rejects.toThrow(
      "unavailable",
    );
    await cache.load("/one", signal());
    expect(fetchMock).toHaveBeenCalledTimes(3);
    expect(fetchMock.mock.calls[2][1].headers).toBeUndefined();
  });

  it("forces an explicit refresh and notifies active cards only after success", async () => {
    const cache = createThumbnailCache();
    const listener = jest.fn();
    cache.subscribe(listener);
    await cache.load("/one", signal());
    await cache.refresh("/one", signal());
    expect(fetchMock.mock.calls[1][0]).toBe("/one?refresh=1");
    expect(listener).toHaveBeenCalledTimes(1);
  });

  it("isolates scopes and releases data on identity change", async () => {
    const alice = createThumbnailCache();
    const bob = createThumbnailCache();
    await alice.load("/same", signal());
    await bob.load("/same", signal());
    alice.clear();
    await alice.load("/same", signal());
    expect(fetchMock).toHaveBeenCalledTimes(3);
  });

  it("evicts the least recently used entry at the count and byte budgets", async () => {
    const cache = createThumbnailCache();
    await Promise.all(
      Array.from({ length: 65 }, (_, index) =>
        cache.load(`/file-${index}`, signal()),
      ),
    );
    await cache.load("/file-0", signal());
    expect(fetchMock).toHaveBeenCalledTimes(66);
    cache.clear();
    fetchMock.mockClear();
    fetchMock.mockImplementation(async () => response(2 * 1024 * 1024));
    await Promise.all(
      Array.from({ length: 9 }, (_, index) =>
        cache.load(`/large-${index}`, signal()),
      ),
    );
    await cache.load("/large-0", signal());
    expect(fetchMock).toHaveBeenCalledTimes(10);
  });

  it("enforces the streamed byte limit even without content length", async () => {
    const cache = createThumbnailCache();
    fetchMock.mockResolvedValueOnce(response(2 * 1024 * 1024 + 1));
    await expect(cache.load("/oversized", signal())).rejects.toThrow(
      "size limit",
    );
  });

  it("cancels a shared request only after its last consumer unmounts", async () => {
    const cache = createThumbnailCache();
    let upstream: AbortSignal | undefined;
    fetchMock.mockImplementation(
      (_url, options: RequestInit) =>
        new Promise<Response>((_resolve, reject) => {
          upstream = options.signal!;
          upstream.addEventListener("abort", () => {
            reject(new DOMException("Aborted", "AbortError"));
          });
        }),
    );
    const first = new AbortController();
    const second = new AbortController();
    const one = cache
      .load("/shared", first.signal)
      .catch((error: unknown) => error);
    const two = cache
      .load("/shared", second.signal)
      .catch((error: unknown) => error);
    first.abort();
    expect(upstream?.aborted).toBe(false);
    second.abort();
    expect(upstream?.aborted).toBe(true);
    await Promise.all([one, two]);
  });

  it("waits for the worker with cancellable bounded polling", async () => {
    jest.useFakeTimers();
    try {
      const cache = createThumbnailCache();
      fetchMock.mockResolvedValueOnce(new Response(null, { status: 202 }));
      const loading = cache.load("/processing", signal());
      await jest.advanceTimersByTimeAsync(2_000);
      expect((await loading).blob.size).toBe(4);
      expect(fetchMock).toHaveBeenCalledTimes(2);
    } finally {
      jest.useRealTimers();
    }
  });
});
