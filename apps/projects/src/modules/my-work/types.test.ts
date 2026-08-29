/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import { normalizeMyWorkLayout } from "./types";

describe("normalizeMyWorkLayout", () => {
  it("preserves supported layouts", () => {
    expect(normalizeMyWorkLayout("list", "kanban")).toBe("list");
    expect(normalizeMyWorkLayout("kanban", "list")).toBe("kanban");
  });

  it("migrates the legacy calendar layout to the supplied fallback", () => {
    expect(normalizeMyWorkLayout("calendar", "kanban")).toBe("kanban");
    expect(normalizeMyWorkLayout("calendar", "list")).toBe("list");
  });
});
