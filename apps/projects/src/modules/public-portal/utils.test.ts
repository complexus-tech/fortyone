/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import { getCrossPortalRequestHref, getGlobalProfileHref } from "./utils";

describe("public portal profile paths", () => {
  it("uses the global profile outside workspace subdomain deployments", () => {
    expect(getGlobalProfileHref()).toBe("/profile");
  });

  it("links global activity back to its originating portal", () => {
    expect(
      getCrossPortalRequestHref("city-roads", "city-roads", "safer-crossing"),
    ).toBe("/portal/city-roads/feedback/safer-crossing");
  });
});
