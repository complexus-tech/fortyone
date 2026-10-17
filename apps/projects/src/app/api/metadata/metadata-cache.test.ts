/* global describe, expect, it, jest -- Jest globals are provided by the projects test runner. */
import type { LinkMetadata } from "@/lib/queries/links/metadata";
import {
  METADATA_FAILURE_TTL_MS,
  METADATA_SUCCESS_TTL_MS,
} from "@/lib/queries/links/metadata";
import { createMetadataCache, MetadataCapacityError } from "./metadata-cache";
import { MetadataFetchError } from "./target-policy";

const deferredMetadata = () => {
  let resolve!: (value: LinkMetadata) => void;
  let reject!: (reason: Error) => void;
  const promise = new Promise<LinkMetadata>((accept, fail) => {
    resolve = accept;
    reject = fail;
  });
  return { promise, resolve, reject };
};

describe("shared public metadata cache", () => {
  it("coalesces fragment and host variants while preserving meaningful query strings", async () => {
    const request = deferredMetadata();
    const load = jest.fn().mockReturnValue(request.promise);
    const cache = createMetadataCache(load);
    const first = cache.get("https://EXAMPLE.com:443/article?view=one#intro");
    const second = cache.get("https://example.com/article?view=one#details");
    await Promise.resolve();
    expect(load).toHaveBeenCalledTimes(1);
    expect(load.mock.calls[0][0].href).toBe(
      "https://example.com/article?view=one",
    );
    request.resolve({ title: "Shared article" });
    await expect(Promise.all([first, second])).resolves.toEqual([
      { title: "Shared article" },
      { title: "Shared article" },
    ]);
    await cache.get("https://example.com/article?view=one#another");
    expect(load).toHaveBeenCalledTimes(1);
    await cache.get("https://example.com/article?view=two");
    expect(load).toHaveBeenCalledTimes(2);
  });

  it("expires successful metadata after one hour", async () => {
    let clock = 0;
    const load = jest.fn().mockResolvedValue({ title: "Article" });
    const cache = createMetadataCache(load, { now: () => clock });
    await cache.get("https://example.com/article");
    clock = METADATA_SUCCESS_TTL_MS - 1;
    await cache.get("https://example.com/article");
    expect(load).toHaveBeenCalledTimes(1);
    clock++;
    await cache.get("https://example.com/article");
    expect(load).toHaveBeenCalledTimes(2);
  });

  it("uses a short negative TTL for failures and empty pages", async () => {
    let clock = 0;
    const load = jest
      .fn()
      .mockRejectedValueOnce(new MetadataFetchError("timeout", "Timed out"))
      .mockResolvedValueOnce({})
      .mockResolvedValue({ title: "Recovered" });
    const cache = createMetadataCache(load, { now: () => clock });
    await expect(cache.get("https://example.com")).rejects.toMatchObject({
      code: "timeout",
    });
    await expect(cache.get("https://example.com")).rejects.toMatchObject({
      code: "timeout",
    });
    expect(load).toHaveBeenCalledTimes(1);
    clock += METADATA_FAILURE_TTL_MS;
    await expect(cache.get("https://example.com")).resolves.toEqual({});
    clock += METADATA_FAILURE_TTL_MS - 1;
    await cache.get("https://example.com");
    expect(load).toHaveBeenCalledTimes(2);
    clock++;
    await expect(cache.get("https://example.com")).resolves.toEqual({
      title: "Recovered",
    });
  });

  it("evicts the least recently used entry when the cache is full", async () => {
    const load = jest.fn().mockResolvedValue({ title: "Article" });
    const cache = createMetadataCache(load, { maxEntries: 2 });
    await cache.get("https://example.com/one");
    await cache.get("https://example.com/two");
    await cache.get("https://example.com/one");
    await cache.get("https://example.com/three");
    await cache.get("https://example.com/one");
    expect(load).toHaveBeenCalledTimes(3);
    await cache.get("https://example.com/two");
    expect(load).toHaveBeenCalledTimes(4);
  });

  it("keeps a shared fetch alive when one caller cancels", async () => {
    const request = deferredMetadata();
    const load = jest.fn().mockReturnValue(request.promise);
    const cache = createMetadataCache(load);
    const firstController = new AbortController();
    const first = cache.get("https://example.com", firstController.signal);
    const canceled = expect(first).rejects.toMatchObject({
      name: "AbortError",
    });
    const second = cache.get("https://example.com");
    await Promise.resolve();
    firstController.abort();
    await canceled;
    expect(load.mock.calls[0][1].aborted).toBe(false);
    request.resolve({ title: "Still needed" });
    await expect(second).resolves.toEqual({ title: "Still needed" });
    await cache.get("https://example.com");
    expect(load).toHaveBeenCalledTimes(1);
  });

  it("aborts abandoned work without caching cancellation or allowing an old result to replace a new one", async () => {
    const abandoned = deferredMetadata();
    const load = jest
      .fn()
      .mockReturnValueOnce(abandoned.promise)
      .mockResolvedValue({ title: "New result" });
    const cache = createMetadataCache(load);
    const controller = new AbortController();
    const first = cache.get("https://example.com", controller.signal);
    const canceled = expect(first).rejects.toMatchObject({
      name: "AbortError",
    });
    await Promise.resolve();
    controller.abort();
    await canceled;
    expect(load.mock.calls[0][1].aborted).toBe(true);
    await expect(cache.get("https://example.com")).resolves.toEqual({
      title: "New result",
    });
    abandoned.resolve({ title: "Stale result" });
    await Promise.resolve();
    await expect(cache.get("https://example.com")).resolves.toEqual({
      title: "New result",
    });
    expect(load).toHaveBeenCalledTimes(2);
  });

  it("bounds distinct in-flight fetches, admits duplicates, and never caches admission failures", async () => {
    const request = deferredMetadata();
    const load = jest
      .fn()
      .mockReturnValueOnce(request.promise)
      .mockResolvedValue({ title: "Second" });
    const cache = createMetadataCache(load, { maxInFlight: 1 });
    const first = cache.get("https://example.com/first");
    const duplicate = cache.get("https://example.com/first#anchor");
    await expect(
      cache.get("https://example.com/second"),
    ).rejects.toBeInstanceOf(MetadataCapacityError);
    request.resolve({ title: "First" });
    await Promise.all([first, duplicate]);
    await expect(cache.get("https://example.com/second")).resolves.toEqual({
      title: "Second",
    });
    expect(load).toHaveBeenCalledTimes(2);
  });

  it.each([
    "ftp://example.com",
    "https://user:secret@example.com",
    "http://localhost",
    "https://example.com:8443",
  ])("validates targets before any lookup: %s", async (url) => {
    const load = jest.fn();
    const cache = createMetadataCache(load);
    await expect(cache.get(url)).rejects.toBeInstanceOf(MetadataFetchError);
    expect(load).not.toHaveBeenCalled();
  });

  it("does not start an already cancelled caller", async () => {
    const load = jest.fn();
    const cache = createMetadataCache(load);
    const controller = new AbortController();
    controller.abort();
    await expect(
      cache.get("https://example.com", controller.signal),
    ).rejects.toMatchObject({ name: "AbortError" });
    expect(load).not.toHaveBeenCalled();
  });
});
