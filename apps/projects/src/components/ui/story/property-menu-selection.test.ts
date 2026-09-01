import { readFileSync } from "node:fs";
import { join } from "node:path";
import {
  getNumberedMenuItem,
  isPropertySelectionActive,
  shouldApplyPropertySelection,
} from "./property-menu-selection";

describe("story property menu selection", () => {
  it("dispatches repeated and empty selections in bulk mode", () => {
    expect(
      shouldApplyPropertySelection("bulk", "status-backlog", "status-backlog"),
    ).toBe(true);
    expect(shouldApplyPropertySelection("bulk", null, null)).toBe(true);
    expect(
      shouldApplyPropertySelection("bulk", "No Priority", "No Priority"),
    ).toBe(true);
  });

  it("preserves single-story no-op and active-state behavior", () => {
    expect(
      shouldApplyPropertySelection(
        "single",
        "status-backlog",
        "status-backlog",
      ),
    ).toBe(false);
    expect(isPropertySelectionActive("single", null, null)).toBe(true);
    expect(isPropertySelectionActive("bulk", null, null)).toBe(false);
  });

  it("resolves numbered shortcuts from the supplied scoped menu items", () => {
    const productStatuses = [
      { id: "product-backlog", teamId: "product" },
      { id: "product-started", teamId: "product" },
    ];

    expect(getNumberedMenuItem(productStatuses, "1")?.id).toBe(
      "product-started",
    );
    expect(getNumberedMenuItem(productStatuses, "2")).toBeUndefined();
    expect(getNumberedMenuItem(productStatuses, "status")).toBeUndefined();

    const statusesMenuSource = readFileSync(
      join(process.cwd(), "src/components/ui/story/statuses-menu.tsx"),
      "utf8",
    );
    expect(statusesMenuSource).toContain(
      "getNumberedMenuItem(filteredStatuses, value)",
    );
  });
});
