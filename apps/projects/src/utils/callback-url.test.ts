/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import {
  getAuthCallbackPath,
  getLoginUrl,
  getSafeCallbackUrl,
  withCallbackUrl,
} from "./callback-url";

describe("callback URL helpers", () => {
  it("allows relative application paths", () => {
    expect(getSafeCallbackUrl("/portal/city-roads/feedback")).toBe(
      "/portal/city-roads/feedback",
    );
  });

  it("allows FortyOne workspace subdomains", () => {
    expect(getSafeCallbackUrl("https://city-roads.fortyone.app/feedback")).toBe(
      "https://city-roads.fortyone.app/feedback",
    );
  });

  it("rejects local absolute URLs in production", () => {
    const originalDomain = process.env.NEXT_PUBLIC_DOMAIN;
    process.env.NEXT_PUBLIC_DOMAIN = "fortyone.app";

    try {
      expect(
        getSafeCallbackUrl("http://localhost:3000/portal/city-roads/feedback"),
      ).toBeUndefined();
    } finally {
      if (originalDomain === undefined) {
        delete process.env.NEXT_PUBLIC_DOMAIN;
      } else {
        process.env.NEXT_PUBLIC_DOMAIN = originalDomain;
      }
    }
  });

  it("rejects protocol-relative and external URLs", () => {
    expect(getSafeCallbackUrl("//malicious.example.com")).toBeUndefined();
    expect(
      getSafeCallbackUrl("https://malicious.example.com/feedback"),
    ).toBeUndefined();
    expect(
      getSafeCallbackUrl("http://city-roads.fortyone.app/feedback"),
    ).toBeUndefined();
    expect(
      getSafeCallbackUrl("https://fortyone.app@malicious.example.com"),
    ).toBeUndefined();
    expect(getSafeCallbackUrl("/\\malicious.example.com")).toBeUndefined();
    expect(getSafeCallbackUrl(`/feedback\u0000malicious`)).toBeUndefined();
    expect(
      getSafeCallbackUrl(`/feedback?next=${"a".repeat(2048)}`),
    ).toBeUndefined();
  });

  it("preserves a callback URL through the authentication callback", () => {
    expect(getAuthCallbackPath("/portal/city-roads/feedback")).toBe(
      "/auth-callback?callbackUrl=%2Fportal%2Fcity-roads%2Ffeedback",
    );
  });

  it("preserves a callback URL when returning to login", () => {
    expect(getLoginUrl("/portal/city-roads/feedback")).toContain(
      "callbackUrl=%2Fportal%2Fcity-roads%2Ffeedback",
    );
  });

  it("appends callback URLs to paths that already have query parameters", () => {
    expect(withCallbackUrl("/signup?source=portal", "/feedback")).toBe(
      "/signup?source=portal&callbackUrl=%2Ffeedback",
    );
  });
});
