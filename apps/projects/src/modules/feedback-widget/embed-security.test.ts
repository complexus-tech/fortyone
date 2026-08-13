/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import {
  getFeedbackWidgetFrameAncestors,
  isAllowedFeedbackWidgetParent,
} from "./embed-security";

const config = {
  allowedOrigins: ["https://app.example.com", "http://localhost:3000"],
  enabled: true,
  widgetKeyId: "widget-key",
};

describe("feedback widget embed security", () => {
  it("allows only an exact configured parent origin", () => {
    expect(
      isAllowedFeedbackWidgetParent(config, "https://app.example.com"),
    ).toBe(true);
    expect(
      isAllowedFeedbackWidgetParent(config, "https://sub.app.example.com"),
    ).toBe(false);
    expect(
      isAllowedFeedbackWidgetParent(config, "https://app.example.com/path"),
    ).toBe(false);
  });

  it("fails closed when embeds are disabled", () => {
    expect(
      isAllowedFeedbackWidgetParent(
        { ...config, enabled: false },
        "https://app.example.com",
      ),
    ).toBe(false);
  });

  it("builds a CSP from validated exact origins only", () => {
    expect(
      getFeedbackWidgetFrameAncestors([
        "https://app.example.com",
        "https://app.example.com/not-an-origin",
      ]),
    ).toBe("frame-ancestors https://app.example.com");
    expect(getFeedbackWidgetFrameAncestors([])).toBe("frame-ancestors 'none'");
  });
});
