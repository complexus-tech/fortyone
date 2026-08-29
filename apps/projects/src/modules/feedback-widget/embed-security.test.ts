/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import {
  getFeedbackWidgetFrameAncestors,
  isValidFeedbackWidgetParent,
} from "./embed-security";

describe("feedback widget embed security", () => {
  it("allows any valid exact parent origin", () => {
    expect(isValidFeedbackWidgetParent("https://app.example.com")).toBe(true);
    expect(isValidFeedbackWidgetParent("https://docs.example.org:8443")).toBe(
      true,
    );
    expect(isValidFeedbackWidgetParent("http://localhost:3000")).toBe(true);
  });

  it("rejects malformed, insecure, and non-origin parent values", () => {
    expect(isValidFeedbackWidgetParent(null)).toBe(false);
    expect(isValidFeedbackWidgetParent("http://app.example.com")).toBe(false);
    expect(isValidFeedbackWidgetParent("https://app.example.com/path")).toBe(
      false,
    );
    expect(isValidFeedbackWidgetParent("https://*.example.com")).toBe(false);
  });

  it("binds each iframe response to the requesting parent only", () => {
    expect(getFeedbackWidgetFrameAncestors("https://app.example.com")).toBe(
      "frame-ancestors https://app.example.com",
    );
    expect(getFeedbackWidgetFrameAncestors(null)).toBe(
      "frame-ancestors 'none'",
    );
  });
});
