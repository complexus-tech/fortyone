/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import { getStoryPath, getStoryReference, isStoryUuid } from "./story";

describe("story routing", () => {
  it("uses a team reference when the story has one", () => {
    expect(
      getStoryReference({
        id: "story-id",
        sequenceId: 42,
        teamCode: " product ",
      }),
    ).toBe("PRODUCT-42");
  });

  it("falls back to the identifier when the reference is incomplete", () => {
    expect(getStoryReference({ id: "story-id", teamCode: "PRODUCT" })).toBe(
      "story-id",
    );
  });

  it("encodes the generated route reference", () => {
    expect(getStoryPath({ id: "story/id" })).toBe("/work/story%2Fid");
  });

  it("recognizes canonical UUIDs", () => {
    expect(isStoryUuid("9c412737-4c72-4b3e-a14a-584694d8d580")).toBe(true);
    expect(isStoryUuid("not-a-uuid")).toBe(false);
  });
});
