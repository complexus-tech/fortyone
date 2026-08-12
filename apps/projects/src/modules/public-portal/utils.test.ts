/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import {
  getAuthorPathByPortalSlug,
  getCrossPortalRequestHref,
  getGlobalProfileHref,
} from "./utils";

describe("public portal profile paths", () => {
  it("uses the global profile outside workspace subdomain deployments", () => {
    expect(getGlobalProfileHref()).toBe("/profile");
  });

  it("links global activity back to its originating portal", () => {
    expect(
      getCrossPortalRequestHref("city-roads", "city-roads", "safer-crossing"),
    ).toBe("/portal/city-roads/feedback/safer-crossing");
  });

  it("does not create profile links for anonymous contributors", () => {
    expect(getAuthorPathByPortalSlug("city-roads", null)).toBeNull();
    expect(
      getAuthorPathByPortalSlug(
        "city-roads",
        "00000000-0000-0000-0000-000000000000",
      ),
    ).toBeNull();
  });
});
