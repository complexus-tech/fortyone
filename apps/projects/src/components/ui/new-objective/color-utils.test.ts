/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import { colors } from "lib";
import { getAvailableObjectiveColor } from "./color-utils";

describe("getAvailableObjectiveColor", () => {
  it("returns the first unused objective color", () => {
    expect(getAvailableObjectiveColor([])).toBe(colors[0]);
    expect(getAvailableObjectiveColor([colors[0]])).toBe(colors[1]);
  });

  it("compares used colors case-insensitively", () => {
    expect(getAvailableObjectiveColor([colors[0].toLowerCase()])).toBe(
      colors[1],
    );
  });

  it("reuses the first palette color after every color has been used", () => {
    expect(getAvailableObjectiveColor(colors)).toBe(colors[0]);
  });
});
