/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import { toPublicPortal } from "./data";

describe("public portal data", () => {
  it("maps API portal data with workspace branding", () => {
    const portal = toPublicPortal(
      {
        boards: [],
        id: "portal-1",
        items: [],
        itemsHasMore: true,
        name: "Acme City",
        slug: "acme-city",
      },
      {
        avatarUrl: "https://cdn.fortyone.app/workspaces/acme-logo.png",
        color: "#123456",
        name: "Acme City",
        slug: "acme-city",
      },
    );

    expect(portal.workspace).toEqual({
      avatarUrl: "https://cdn.fortyone.app/workspaces/acme-logo.png",
      color: "#123456",
      name: "Acme City",
      slug: "acme-city",
    });
    expect(portal.requestsHasMore).toBe(true);
  });
});
