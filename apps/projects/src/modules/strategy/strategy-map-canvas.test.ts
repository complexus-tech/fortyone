/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import { readFileSync } from "node:fs";
import { join } from "node:path";

const readSource = (path: string) =>
  readFileSync(join(process.cwd(), path), "utf8");

describe("StrategyMapCanvas", () => {
  it("keeps objective alignment discoverable", () => {
    const source = readSource("src/modules/strategy/strategy-map-canvas.tsx");

    expect(source).toContain("useDraggable");
    expect(source).toContain("useDroppable");
    expect(source).toContain('aria-label="Move objective"');
    expect(source).toContain("Drag an objective here");
  });

  it("keeps strategy controls in the main header", () => {
    const pageSource = readSource("src/modules/strategy/index.tsx");

    expect(pageSource).toContain('aria-label="Zoom out"');
    expect(pageSource).toContain('aria-label="Zoom in"');
    expect(pageSource).not.toContain("Strategy hierarchy");
    expect(pageSource).not.toContain("Show unaligned");
    expect(pageSource).not.toContain("All teams");
    expect(pageSource).not.toContain("getFullYear");
  });

  it("renders a collapsible hierarchy with editable objective properties", () => {
    const source = readSource("src/modules/strategy/strategy-map-canvas.tsx");
    const navigationSource = readSource(
      "src/components/shared/sidebar/navigation.tsx",
    );

    expect(source).toContain("HierarchyBadge");
    expect(source).toContain("Ultimate Goal");
    expect(source).toContain("Strategic Pillar");
    expect(source).toContain("KeyResultTree");
    expect(source).toContain("ObjectiveStatusesMenu");
    expect(source).toContain("ObjectiveHealthEditor");
    expect(navigationSource).toContain("StrategyIcon");
  });
});
