/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { headers } from "next/headers";
import { getApiUrl } from "@/lib/api-url";
import { getPublicPortal } from "./query";

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

  it("opts the public roadmap bootstrap into timed revalidation", async () => {
    await getPublicPortal(
      "feedback",
      {},
      { revalidateSeconds: 5 * 60 },
    );

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
});
