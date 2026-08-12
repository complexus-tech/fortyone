/* global describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import {
  FEEDBACK_WIDGET_CHANNEL,
  getTrustedWidgetOrigin,
  isFeedbackWidgetMessage,
  postFeedbackWidgetMessage,
} from "./protocol";

describe("feedback widget protocol", () => {
  it("accepts exact HTTP origins and rejects paths or unsafe protocols", () => {
    expect(getTrustedWidgetOrigin("https://product.example.com")).toBe(
      "https://product.example.com",
    );
    expect(getTrustedWidgetOrigin("http://localhost:3000")).toBe(
      "http://localhost:3000",
    );
    expect(getTrustedWidgetOrigin("http://product.example.com")).toBeNull();
    expect(
      getTrustedWidgetOrigin("https://product.example.com/path"),
    ).toBeNull();
    expect(getTrustedWidgetOrigin("ftp://product.example.com")).toBeNull();
  });

  it("scopes messages to the channel, version, and widget instance", () => {
    expect(
      isFeedbackWidgetMessage(
        {
          channel: FEEDBACK_WIDGET_CHANNEL,
          event: "close",
          instanceId: "widget-1",
          version: 1,
        },
        "widget-1",
      ),
    ).toBe(true);
    expect(
      isFeedbackWidgetMessage(
        {
          channel: FEEDBACK_WIDGET_CHANNEL,
          event: "close",
          instanceId: "widget-2",
          version: 1,
        },
        "widget-1",
      ),
    ).toBe(false);
  });

  it("posts only to the validated parent origin", () => {
    const postMessage = jest.fn();
    Object.defineProperty(window, "parent", {
      configurable: true,
      value: { postMessage },
    });

    postFeedbackWidgetMessage(
      "ready",
      "widget-1",
      "https://product.example.com",
    );

    expect(postMessage).toHaveBeenCalledWith(
      expect.objectContaining({
        channel: FEEDBACK_WIDGET_CHANNEL,
        event: "ready",
        instanceId: "widget-1",
      }),
      "https://product.example.com",
    );
  });
});
