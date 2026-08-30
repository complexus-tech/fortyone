/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import { getAcceptedStoryPath } from "./accepted-story-path";

describe("getAcceptedStoryPath", () => {
  it("matches the canonical work route for an accepted story id", () => {
    expect(getAcceptedStoryPath("story id/with slash")).toBe(
      "/work/story%20id%2Fwith%20slash",
    );
  });
});
