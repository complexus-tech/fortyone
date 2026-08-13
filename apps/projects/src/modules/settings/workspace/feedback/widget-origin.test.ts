/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import {
  normalizeFeedbackWidgetOrigin,
  normalizeFeedbackWidgetOrigins,
} from "./widget-origin";

describe("feedback widget origin normalization", () => {
  it("normalizes exact HTTPS and localhost origins", () => {
    expect(
      normalizeFeedbackWidgetOrigin(" https://App.Example.com:8443/ "),
    ).toBe("https://app.example.com:8443");
    expect(normalizeFeedbackWidgetOrigin("http://localhost:3000")).toBe(
      "http://localhost:3000",
    );
  });

  it.each([
    "http://app.example.com",
    "https://*.example.com",
    "https://user:pass@app.example.com",
    "https://app.example.com/widget",
    "https://app.example.com?source=widget",
  ])("rejects unsafe or inexact origin %s", (origin) => {
    expect(() => normalizeFeedbackWidgetOrigin(origin)).toThrow();
  });

  it("rejects duplicate normalized origins", () => {
    expect(() =>
      normalizeFeedbackWidgetOrigins([
        "https://app.example.com",
        "https://APP.example.com/",
      ]),
    ).toThrow("only be added once");
  });
});
