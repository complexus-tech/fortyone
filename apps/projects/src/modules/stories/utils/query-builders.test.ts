/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import { buildGroupedStoriesQuery } from "./query-builders";

describe("buildGroupedStoriesQuery", () => {
  it("serializes the selected order direction", () => {
    expect(
      buildGroupedStoriesQuery({
        groupBy: "status",
        orderBy: "created",
        orderDirection: "asc",
      }),
    ).toContain("orderDirection=asc");
  });
});
