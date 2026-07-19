/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import {
  getCanonicalPublicPath,
  getInternalPublicPath,
  isPublicPath,
} from "./public-portal-routes";

describe("workspace public portal routes", () => {
  it("rewrites the short feedback URL to the internal portal route", () => {
    expect(getInternalPublicPath("/feedback", "art-circles")).toBe(
      "/portal/art-circles/feedback",
    );
  });

  it("redirects the legacy portal URL to the canonical short URL", () => {
    expect(
      getCanonicalPublicPath("/portal/art-circles/feedback", "art-circles"),
    ).toBe("/feedback");
  });

  it("recognizes short feedback routes as public", () => {
    expect(isPublicPath("/feedback")).toBe(true);
    expect(isPublicPath("/feedback/improve-mobile-navigation")).toBe(true);
    expect(isPublicPath("/feedback-private")).toBe(false);
  });

  it("maps legacy request paths onto canonical feedback paths", () => {
    expect(
      getCanonicalPublicPath(
        "/portal/art-circles/requests/improve-mobile-navigation",
        "art-circles",
      ),
    ).toBe("/feedback/improve-mobile-navigation");
  });

  it("rewrites roadmap and update paths through the current workspace", () => {
    expect(getInternalPublicPath("/roadmap", "art-circles")).toBe(
      "/portal/art-circles/roadmap",
    );
    expect(getInternalPublicPath("/updates", "art-circles")).toBe(
      "/portal/art-circles/updates",
    );
  });

  it("keeps portal account settings outside workspace application routes", () => {
    expect(getInternalPublicPath("/account", "art-circles")).toBe(
      "/portal/art-circles/account",
    );
    expect(
      getCanonicalPublicPath("/portal/art-circles/account", "art-circles"),
    ).toBe("/account");
  });

  it("keeps feedback detail paths on the workspace subdomain", () => {
    expect(
      getInternalPublicPath(
        "/feedback/improve-mobile-navigation",
        "art-circles",
      ),
    ).toBe("/portal/art-circles/feedback/improve-mobile-navigation");
  });
});
