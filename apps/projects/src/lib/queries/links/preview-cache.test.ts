/* global beforeEach, describe, expect, it, jest -- Provided by the projects test runner. */
import {
  PREVIEW_CACHE_TTL_MS,
  readPreviewCache,
  removePreviewCache,
  writePreviewCache,
} from "./preview-cache";

const isTitle = (value: unknown): value is { title: string } =>
  typeof value === "object" &&
  value !== null &&
  "title" in value &&
  typeof value.title === "string";

beforeEach(() => {
  localStorage.clear();
});

describe("durable preview cache", () => {
  it("keeps successful previews for a full day and expires them afterward", () => {
    const now = Date.now();
    writePreviewCache("scoped-key", { title: "Saved preview" }, now);
    const time = jest.spyOn(Date, "now");
    try {
      time.mockReturnValue(now + PREVIEW_CACHE_TTL_MS - 1);
      expect(readPreviewCache("scoped-key", isTitle)?.data.title).toBe(
        "Saved preview",
      );
      time.mockReturnValue(now + PREVIEW_CACHE_TTL_MS);
      expect(readPreviewCache("scoped-key", isTitle)).toBeUndefined();
      expect(localStorage.length).toBe(0);
    } finally {
      time.mockRestore();
    }
  });

  it("evicts old previews instead of letting browser storage grow without bound", () => {
    for (let index = 0; index < 120; index++)
      writePreviewCache(
        String(index),
        { title: `Preview ${index}` },
        Date.now() - 120 + index,
      );
    expect(localStorage.length).toBe(100);
    expect(readPreviewCache("0", isTitle)).toBeUndefined();
    expect(readPreviewCache("119", isTitle)?.data.title).toBe("Preview 119");
  });

  it("ignores malformed entries and leaves unrelated browser data untouched", () => {
    localStorage.setItem("fortyone:link-preview:v1:broken", "not JSON");
    localStorage.setItem("other-setting", "keep");
    expect(readPreviewCache("broken", isTitle)).toBeUndefined();
    writePreviewCache("valid", { title: "Valid" });
    expect(localStorage.getItem("other-setting")).toBe("keep");
    expect(localStorage.getItem("fortyone:link-preview:v1:broken")).toBeNull();
    removePreviewCache("valid");
    expect(readPreviewCache("valid", isTitle)).toBeUndefined();
  });

  it("falls back to normal queries when browser storage is disabled", () => {
    const read = jest
      .spyOn(Storage.prototype, "getItem")
      .mockImplementation(() => {
        throw new Error("Blocked");
      });
    const write = jest
      .spyOn(Storage.prototype, "setItem")
      .mockImplementation(() => {
        throw new Error("Blocked");
      });
    try {
      expect(readPreviewCache("key", isTitle)).toBeUndefined();
      expect(() => {
        writePreviewCache("key", { title: "Saved" });
      }).not.toThrow();
    } finally {
      read.mockRestore();
      write.mockRestore();
    }
  });
});
