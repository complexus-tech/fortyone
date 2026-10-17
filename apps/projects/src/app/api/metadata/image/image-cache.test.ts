/* @jest-environment node */
/* eslint-disable no-await-in-loop -- Drain consecutive promise continuations before asserting shared work. */
/* global describe, expect, it, jest -- Jest globals are provided by the projects test runner. */
import type { PreviewImage, PreviewImageStore } from "./object-store";
import { createImageCache } from "./image-cache";

jest.mock("./object-store", () => ({ previewImageStore: {} }));

const image = { body: Buffer.from("raster-image"), contentType: "image/png" };
const storageFixture = (): PreviewImageStore => {
  const entries = new Map<string, PreviewImage>();
  return {
    read: async (key) => entries.get(key),
    write: async (key, value) => {
      entries.set(key, value);
    },
  };
};

describe("metadata image cache", () => {
  it("reuses validated image bytes across instances and fetches a new image revision only once", async () => {
    const storage = storageFixture();
    const load = jest.fn().mockResolvedValue(image);
    const first = createImageCache({ storage, load });
    const second = createImageCache({ storage, load });
    await first.get("https://example.com/og.png", "1");
    await second.get("https://example.com/og.png", "1");
    expect(load).toHaveBeenCalledTimes(1);
    await Promise.all([
      first.get("https://example.com/og.png", "2"),
      first.get("https://example.com/og.png", "2"),
    ]);
    expect(load).toHaveBeenCalledTimes(2);
  });

  it("negative-caches unavailable images and retries after the short failure TTL", async () => {
    let clock = 0;
    const load = jest
      .fn()
      .mockRejectedValueOnce(new Error("Unavailable"))
      .mockResolvedValue(image);
    const cache = createImageCache({ load, now: () => clock });
    await expect(
      cache.get("https://example.com/og.png", "1"),
    ).rejects.toThrow();
    await expect(
      cache.get("https://example.com/og.png", "1"),
    ).rejects.toThrow();
    expect(load).toHaveBeenCalledTimes(1);
    clock = 30_001;
    expect(await cache.get("https://example.com/og.png", "1")).toEqual(image);
  });

  it("cancels one client without aborting another subscriber", async () => {
    let finish!: (value: PreviewImage) => void;
    const load = jest.fn(
      () =>
        new Promise<PreviewImage>((resolve) => {
          finish = resolve;
        }),
    );
    const cache = createImageCache({ load });
    const controller = new AbortController();
    const first = cache.get(
      "https://example.com/og.png",
      "1",
      controller.signal,
    );
    const second = cache.get("https://example.com/og.png", "1");
    const aborted = expect(first).rejects.toMatchObject({ name: "AbortError" });
    controller.abort();
    await aborted;
    for (let i = 0; i < 4; i++) await Promise.resolve();
    finish(image);
    expect(await second).toEqual(image);
    expect(load).toHaveBeenCalledTimes(1);
  });

  it("validates targets before cache or upstream access", async () => {
    const load = jest.fn();
    const cache = createImageCache({ load });
    await expect(cache.get("http://localhost/private", "1")).rejects.toThrow();
    expect(load).not.toHaveBeenCalled();
  });
});
