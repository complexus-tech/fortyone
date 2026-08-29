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
    expect(portal.participationMode).toBe("account_required");
    expect(portal.guestIdentityPolicy).toBe("show_identity");
    expect(portal.hasPublishedUpdates).toBe(false);
  });

  it("maps anonymous participation when the portal enables it", () => {
    const portal = toPublicPortal({
      boards: [],
      id: "portal-1",
      items: [
        {
          authorId: null,
          authorMasked: true,
          authorName: "Anonymous",
          boardId: "board-1",
          commentCount: 0,
          createdAt: "2026-08-12T10:00:00.000Z",
          description: "The curb ramp is blocked.",
          id: "feedback-1",
          slug: "blocked-curb-ramp",
          status: "pending",
          title: "Blocked curb ramp",
          voteCount: 0,
        },
      ],
      name: "Acme City",
      participationMode: "anonymous_allowed",
      slug: "acme-city",
    });

    expect(portal.participationMode).toBe("anonymous_allowed");
    expect(portal.requests[0]).toEqual(
      expect.objectContaining({
        authorId: null,
        authorMasked: true,
        authorName: "Anonymous",
      }),
    );
  });

  it("maps verified guest policy and hides update navigation until publication exists", () => {
    const portal = toPublicPortal({
      boards: [],
      guestIdentityPolicy: "allow_public_masking",
      hasPublishedUpdates: true,
      id: "portal-1",
      items: [],
      name: "Acme City",
      participationMode: "verified_guest",
      slug: "acme-city",
    });

    expect(portal).toEqual(
      expect.objectContaining({
        guestIdentityPolicy: "allow_public_masking",
        hasPublishedUpdates: true,
        participationMode: "verified_guest",
      }),
    );
  });
});
