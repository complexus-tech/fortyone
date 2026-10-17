/* @jest-environment node */
/* global afterEach, beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */
import { Readable } from "node:stream";

const mockSend = jest.fn();
jest.mock("@aws-sdk/client-s3", () => ({
  ...jest.requireActual("@aws-sdk/client-s3"),
  S3Client: jest.fn().mockImplementation(() => ({ send: mockSend })),
}));

const previousBucket = process.env.PREVIEW_S3_BUCKET;
const image = { body: Buffer.from("raster-image"), contentType: "image/png" };

beforeEach(() => {
  jest.resetModules();
  mockSend.mockReset();
  process.env.PREVIEW_S3_BUCKET = "preview-cache-test";
  jest.spyOn(console, "warn").mockImplementation(() => undefined);
});

afterEach(() => {
  jest.restoreAllMocks();
  if (previousBucket === undefined) delete process.env.PREVIEW_S3_BUCKET;
  else process.env.PREVIEW_S3_BUCKET = previousBucket;
});

describe("preview image object store", () => {
  it("publishes an upstream image even when a cold cache read exceeds its shorter deadline", async () => {
    const timeout = jest.spyOn(AbortSignal, "timeout");
    mockSend
      .mockRejectedValueOnce(
        new DOMException("Slow connection", "TimeoutError"),
      )
      .mockResolvedValueOnce({});
    const { previewImageStore } = await import("./object-store");
    const { createImageCache } = await import("./image-cache");
    const load = jest.fn().mockResolvedValue(image);
    const cache = createImageCache({ storage: previewImageStore, load });

    await expect(cache.get("https://example.com/og.png", "1")).resolves.toEqual(
      image,
    );

    expect(mockSend).toHaveBeenCalledTimes(2);
    expect(mockSend.mock.calls[1][0].input).toMatchObject({
      Bucket: "preview-cache-test",
      Key: expect.stringMatching(/^previews\/external\/v1\/[a-f0-9]{64}$/),
      Body: image.body,
      ContentType: "image/png",
    });
    expect(timeout).toHaveBeenCalledWith(1_500);
    expect(timeout).toHaveBeenCalledWith(5_000);
    expect(previewImageStore.available?.()).toBe(false);

    // Cooling down reads must not cause a repeat upstream load in this process.
    await cache.get("https://example.com/og.png", "1");
    expect(load).toHaveBeenCalledTimes(1);
    expect(mockSend).toHaveBeenCalledTimes(2);
  });

  it("keeps reads available during a write failure and retries writes after their cooldown", async () => {
    let clock = Date.now();
    jest.spyOn(Date, "now").mockImplementation(() => clock);
    mockSend
      .mockRejectedValueOnce(new Error("Write unavailable"))
      .mockResolvedValueOnce({
        Body: Readable.from([image.body]),
        ContentType: image.contentType,
        ContentLength: image.body.byteLength,
        Metadata: { createdat: String(clock) },
      })
      .mockResolvedValueOnce({});
    const { previewImageStore } = await import("./object-store");

    await previewImageStore.write("new-image", image);
    expect(previewImageStore.available?.()).toBe(true);
    await expect(previewImageStore.read("existing-image")).resolves.toEqual(
      image,
    );
    await previewImageStore.write("new-image", image);
    expect(mockSend).toHaveBeenCalledTimes(2);

    clock += 30_001;
    await previewImageStore.write("new-image", image);
    expect(mockSend).toHaveBeenCalledTimes(3);
  });
});
