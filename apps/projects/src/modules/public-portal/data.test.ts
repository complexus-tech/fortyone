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
  });

  it("maps anonymous participation when the portal enables it", () => {
    const portal = toPublicPortal({
      boards: [],
      id: "portal-1",
      items: [
        {
          authorId: null,
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
        authorName: "Anonymous",
      }),
    );
  });
});
