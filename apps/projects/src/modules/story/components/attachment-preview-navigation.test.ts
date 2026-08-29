/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import { getAdjacentAttachmentIndex } from "./attachment-preview-navigation";

describe("attachment preview navigation", () => {
  it("wraps from the first attachment to the last", () => {
    expect(getAdjacentAttachmentIndex(0, 4, -1)).toBe(3);
  });

  it("wraps from the last attachment to the first", () => {
    expect(getAdjacentAttachmentIndex(3, 4, 1)).toBe(0);
  });
});
