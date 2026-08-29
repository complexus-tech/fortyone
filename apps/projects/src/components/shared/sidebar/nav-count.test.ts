/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import { formatNavCount } from "./nav-count";

describe("formatNavCount", () => {
  it.each([
    [0, null],
    [-1, null],
    [1, "1"],
    [9, "9"],
    [10, "10+"],
    [19, "10+"],
    [20, "20+"],
    [49, "20+"],
    [50, "50+"],
    [999, "50+"],
    [1000, "1K+"],
    [2_500, "2K+"],
    [9_999, "9K+"],
    [1_000_000, "9K+"],
  ])("formats %i as %s", (count, expected) => {
    expect(formatNavCount(count)).toBe(expected);
  });
});
