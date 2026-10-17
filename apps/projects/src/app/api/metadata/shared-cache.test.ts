/* eslint-disable no-await-in-loop -- Drain consecutive promise continuations before asserting shared work. */
/* global describe, expect, it, jest -- Jest globals are provided by the projects test runner. */
import { METADATA_SUCCESS_TTL_MS } from "@/lib/queries/links/metadata";
import type { SharedPreviewCache } from "./shared-cache";
import { createMetadataCache } from "./metadata-cache";

const sharedFixture = () => {
  const values = new Map<string, string>();
  const locks = new Map<string, string>();
  const shared: SharedPreviewCache = {
    read: async (key) => values.get(key) ?? null,
    write: async (key, value) => {
      values.set(key, value);
    },
    claim: async (key, token) => {
      if (locks.has(key)) return false;
      locks.set(key, token);
      return true;
    },
    release: async (key, token) => {
      if (locks.get(key) === token) locks.delete(key);
    },
  };
  return shared;
};

describe("replicated metadata cache", () => {
  it("shares successful data across independent instances and observes manual refresh revisions", async () => {
    const shared = sharedFixture();
    const firstLoad = jest
      .fn()
      .mockResolvedValue({ title: "Original", version: "1" });
    const secondLoad = jest
      .fn()
      .mockResolvedValue({ title: "Updated", version: "2" });
    const first = createMetadataCache(firstLoad, { shared });
    const second = createMetadataCache(secondLoad, { shared });
    await first.get("https://example.com");
    expect(await second.get("https://example.com")).toMatchObject({
      title: "Original",
    });
    expect(secondLoad).not.toHaveBeenCalled();
    await second.get("https://example.com", undefined, { force: true });
    expect(await first.get("https://example.com")).toMatchObject({
      title: "Updated",
    });
    expect(firstLoad).toHaveBeenCalledTimes(1);
  });

  it("returns stale data immediately, schedules one refresh, and keeps last-good data on failure", async () => {
    let clock = 0;
    const work: (() => Promise<void>)[] = [];
    const load = jest
      .fn()
      .mockResolvedValueOnce({ title: "Original", version: "1" })
      .mockRejectedValue(new Error("Offline"));
    const cache = createMetadataCache(load, {
      now: () => clock,
      schedule: (job) => work.push(job),
    });
    await cache.get("https://example.com");
    clock = METADATA_SUCCESS_TTL_MS + 1;
    expect(await cache.get("https://example.com")).toMatchObject({
      title: "Original",
      stale: true,
    });
    expect(load).toHaveBeenCalledTimes(1);
    await work[0]();
    expect(await cache.get("https://example.com")).toMatchObject({
      title: "Original",
    });
    await expect(
      cache.get("https://example.com", undefined, { force: true }),
    ).rejects.toThrow();
    expect(await cache.get("https://example.com")).toMatchObject({
      title: "Original",
    });
  });

  it("coalesces cold misses across instances with a distributed lease", async () => {
    const shared = sharedFixture();
    let finish!: (value: { title: string; version: string }) => void;
    const load = jest.fn(
      () =>
        new Promise<{ title: string; version: string }>((resolve) => {
          finish = resolve;
        }),
    );
    const first = createMetadataCache(load, { shared });
    const secondLoad = jest.fn();
    const second = createMetadataCache(secondLoad, { shared });
    const firstRequest = first.get("https://example.com");
    for (let i = 0; i < 5; i++) await Promise.resolve();
    const secondRequest = second.get("https://example.com");
    for (let i = 0; i < 5; i++) await Promise.resolve();
    finish({ title: "Shared", version: "1" });
    expect(await firstRequest).toMatchObject({ title: "Shared" });
    expect(await secondRequest).toMatchObject({ title: "Shared" });
    expect(secondLoad).not.toHaveBeenCalled();
  });

  it("falls back to bounded local coalescing when shared storage is unavailable", async () => {
    const load = jest.fn().mockResolvedValue({ title: "Available" });
    const shared = sharedFixture();
    shared.read = async () => null;
    shared.claim = async () => null;
    const cache = createMetadataCache(load, { shared });
    await Promise.all([
      cache.get("https://example.com"),
      cache.get("https://example.com"),
    ]);
    await cache.get("https://example.com");
    expect(load).toHaveBeenCalledTimes(1);
  });
});
