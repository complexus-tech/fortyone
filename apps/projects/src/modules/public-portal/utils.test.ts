/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import {
  getPortalAccountPathBySlug,
  getViewerProfileHrefByPortalSlug,
} from "./utils";

describe("public portal account paths", () => {
  it("preserves portal context when opening account settings", () => {
    expect(getPortalAccountPathBySlug("city-roads")).toBe(
      "/portal/city-roads/account?portal=city-roads",
    );
  });

  it("builds the feedback contributor profile destination", () => {
    expect(getViewerProfileHrefByPortalSlug("city-roads")).toBe(
      "/portal/city-roads/people/me",
    );
  });
});
