/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import { publicPortalFixture } from "./fixtures";
import { applyPublicPortalWorkspace } from "./data";

describe("public portal data", () => {
  it("applies public workspace branding over fixture branding", () => {
    const portal = applyPublicPortalWorkspace(publicPortalFixture, {
      avatarUrl: "https://cdn.fortyone.app/workspaces/acme-logo.png",
      color: "#123456",
      name: "Acme City",
      slug: "acme-city",
    });

    expect(portal.workspace).toEqual({
      avatarUrl: "https://cdn.fortyone.app/workspaces/acme-logo.png",
      color: "#123456",
      name: "Acme City",
      slug: "acme-city",
    });
  });
});
