/* global describe, expect, it -- Jest provides these test globals. */

import { extractRealtimeClientAction } from "./realtime-client-actions";

describe("extractRealtimeClientAction", () => {
  it("returns a safe navigation action and hides it from the model", () => {
    const result = extractRealtimeClientAction({
      clientAction: { path: "/story/story-id", type: "navigate" },
      message: "Opening story.",
      success: true,
    });

    expect(result.action).toEqual({
      path: "/story/story-id",
      type: "navigate",
    });
    expect(result.modelOutput).toEqual({
      message: "Opening story.",
      success: true,
    });
  });

  it("rejects external and protocol-relative navigation", () => {
    expect(
      extractRealtimeClientAction({
        clientAction: { path: "https://example.com", type: "navigate" },
      }).action,
    ).toBeNull();
    expect(
      extractRealtimeClientAction({
        clientAction: { path: "//example.com", type: "navigate" },
      }).action,
    ).toBeNull();
  });

  it("accepts supported themes and rejects unknown themes", () => {
    expect(
      extractRealtimeClientAction({
        clientAction: { theme: "dark", type: "theme" },
      }).action,
    ).toEqual({ theme: "dark", type: "theme" });
    expect(
      extractRealtimeClientAction({
        clientAction: { theme: "sepia", type: "theme" },
      }).action,
    ).toBeNull();
  });
});
