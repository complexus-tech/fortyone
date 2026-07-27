/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import { getStoryPath, getStoryReference, isStoryUuid } from "./story-url";

describe("story URL helpers", () => {
  it("builds the canonical path from the team code and sequence ID", () => {
    const story = {
      id: "d0c8baaf-d40e-4d2f-8f37-b702da402085",
      sequenceId: 571,
      teamCode: " prd ",
    };

    expect(getStoryReference(story)).toBe("PRD-571");
    expect(getStoryPath(story)).toBe("/work/PRD-571");
  });

  it("falls back to the UUID when reference data is unavailable", () => {
    expect(
      getStoryPath({
        id: "d0c8baaf-d40e-4d2f-8f37-b702da402085",
      }),
    ).toBe("/work/d0c8baaf-d40e-4d2f-8f37-b702da402085");
  });

  it("recognizes supported UUID story identifiers", () => {
    expect(isStoryUuid("d0c8baaf-d40e-4d2f-8f37-b702da402085")).toBe(true);
    expect(isStoryUuid("PRD-571")).toBe(false);
  });
});
