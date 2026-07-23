/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { headers } from "next/headers";
import { getApiUrl } from "@/lib/api-url";
import {
  getPublicContributor,
  getPublicContributorComments,
  getPublicPortal,
} from "./query";

jest.mock("next/headers", () => ({
  headers: jest.fn(),
}));

jest.mock("@/lib/api-url", () => ({
  getApiUrl: jest.fn(),
}));

const headersMock = jest.mocked(headers);
const getApiUrlMock = jest.mocked(getApiUrl);

const response = (data: unknown) =>
  ({
    json: async () => ({ data }),
    ok: true,
  }) as Response;

describe("public portal query caching", () => {
  beforeEach(() => {
    getApiUrlMock.mockReturnValue("https://api.fortyone.test");
    headersMock.mockResolvedValue(
      new Headers({ host: "art-circles.fortyone.app" }),
    );
    global.fetch = jest.fn(async (input: Parameters<typeof fetch>[0]) => {
      const url = String(input);

      if (url.endsWith("/workspaces/art-circles/portal")) {
        return response({
          avatarUrl: null,
          color: "#111111",
          name: "Art Circles",
          slug: "art-circles",
        });
      }

      return response({
        boards: [],
        id: "portal-1",
        items: [],
        name: "Art Circles",
        slug: "feedback",
      });
    });
  });

  it("keeps public portal requests uncached by default", async () => {
    await getPublicPortal("feedback");

    expect(global.fetch).toHaveBeenNthCalledWith(
      1,
      "https://api.fortyone.test/workspaces/art-circles/portal",
      { cache: "no-store" },
    );
    expect(global.fetch).toHaveBeenNthCalledWith(
      2,
      "https://api.fortyone.test/workspaces/art-circles/portals/feedback/feedback",
      { cache: "no-store" },
    );
  });

  it("requests only active feedback for the default public list", async () => {
    await getPublicPortal("feedback", { status: "active" });

    expect(global.fetch).toHaveBeenNthCalledWith(
      2,
      "https://api.fortyone.test/workspaces/art-circles/portals/feedback/feedback?status=active",
      { cache: "no-store" },
    );
  });

  it("opts the public roadmap bootstrap into timed revalidation", async () => {
    await getPublicPortal("feedback", {}, { revalidateSeconds: 5 * 60 });

    expect(global.fetch).toHaveBeenNthCalledWith(
      1,
      "https://api.fortyone.test/workspaces/art-circles/portal",
      { next: { revalidate: 300 } },
    );
    expect(global.fetch).toHaveBeenNthCalledWith(
      2,
      "https://api.fortyone.test/workspaces/art-circles/portals/feedback/feedback",
      { next: { revalidate: 300 } },
    );
  });

  it("loads contributor summaries from the workspace-scoped public route", async () => {
    global.fetch = jest.fn(async () =>
      response({
        avatarUrl: null,
        id: "author-1",
        joinedAt: "2026-07-20T10:00:00.000Z",
        name: "Joseph Mukorivo",
        stats: { commentCount: 4, feedbackCount: 2, voteScore: 9 },
      }),
    );

    await getPublicContributor("feedback", "author-1");

    expect(global.fetch).toHaveBeenCalledWith(
      "https://api.fortyone.test/workspaces/art-circles/portals/feedback/feedback/contributors/author-1",
      { cache: "no-store" },
    );
  });

  it("loads paginated contributor comments from the global public route", async () => {
    headersMock.mockResolvedValue(new Headers({ host: "localhost:3000" }));
    global.fetch = jest.fn(async () =>
      response({
        comments: [],
        pagination: { hasMore: false, nextPage: 3, page: 2, pageSize: 10 },
      }),
    );

    await getPublicContributorComments("feedback", "author-1", 2, 10);

    expect(global.fetch).toHaveBeenCalledWith(
      "https://api.fortyone.test/portals/feedback/feedback/contributors/author-1/comments?page=2&pageSize=10",
      { cache: "no-store" },
    );
  });
});
