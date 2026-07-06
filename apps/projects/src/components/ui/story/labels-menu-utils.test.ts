/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import type { Label } from "@/types";
import { canCreateLabelFromQuery } from "./labels-menu-utils";

const label = (name: string): Label => ({
  id: name,
  name,
  color: "#000000",
  teamId: "team-1",
  workspaceId: "workspace-1",
  createdAt: "2026-06-15T00:00:00Z",
  updatedAt: "2026-06-15T00:00:00Z",
});

describe("labels menu utils", () => {
  it("allows creating a searched label when no exact label exists", () => {
    expect(canCreateLabelFromQuery("Frontend", [label("Backend")])).toBe(true);
  });

  it("does not allow creating an exact label duplicate", () => {
    expect(canCreateLabelFromQuery(" frontend ", [label("Frontend")])).toBe(
      false,
    );
  });

  it("does not allow creating an empty label", () => {
    expect(canCreateLabelFromQuery("   ", [])).toBe(false);
  });
});
