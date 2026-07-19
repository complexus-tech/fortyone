/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import {
  parsePublicPortalFilters,
  toPublicPortalSearchParams,
} from "./query-params";

describe("public portal query params", () => {
  it("parses supported feedback filters and ignores unsupported values", () => {
    expect(
      parsePublicPortalFilters(
        new URLSearchParams({
          boardId: "road-repairs",
          search: "  traffic signal  ",
          sort: "newest",
          status: "reviewing",
        }),
      ),
    ).toEqual({
      boardId: "road-repairs",
      search: "traffic signal",
      sort: "newest",
      status: "reviewing",
    });

    expect(
      parsePublicPortalFilters(
        new URLSearchParams({ sort: "popular", status: "draft" }),
      ),
    ).toEqual({
      boardId: undefined,
      search: "",
      sort: "top",
      status: undefined,
    });
  });

  it("serializes shareable feedback filter state", () => {
    const params = toPublicPortalSearchParams({
      boardId: "road-repairs",
      search: "traffic signal",
      sort: "oldest",
      status: "planned",
    });

    expect(params.get("boardId")).toBe("road-repairs");
    expect(params.get("search")).toBe("traffic signal");
    expect(params.get("sort")).toBe("oldest");
    expect(params.get("status")).toBe("planned");
  });
});
