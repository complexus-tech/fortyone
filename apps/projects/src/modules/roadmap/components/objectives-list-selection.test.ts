/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import { readFileSync } from "node:fs";
import { join } from "node:path";

const readSource = (path: string) =>
  readFileSync(join(process.cwd(), path), "utf8");

describe("Roadmap objective list selection", () => {
  it("places the collapse arrow after the group identity", () => {
    const source = readSource(
      "src/modules/roadmap/components/objectives-board.tsx",
    );
    const header = source.slice(
      source.indexOf("const ObjectiveGroupHeader"),
      source.indexOf("const HiddenObjectiveGroups"),
    );

    expect(header.indexOf("<ObjectiveGroupIdentity")).toBeLessThan(
      header.indexOf("<ArrowDownIcon"),
    );
  });

  it("supports row and group selection without the objective row icon", () => {
    const boardSource = readSource(
      "src/modules/roadmap/components/objectives-board.tsx",
    );
    const cardSource = readSource("src/modules/objectives/components/card.tsx");

    expect(boardSource).toContain("selectedObjectives");
    expect(boardSource).toContain("<ObjectivesToolbar");
    expect(cardSource).toContain("onSelectionChange");
    expect(cardSource).toContain("onSelectionChange ? null");
  });

  it("exposes only supported objective bulk operations", () => {
    const toolbarSource = readSource(
      "src/modules/roadmap/components/objectives-toolbar.tsx",
    );

    expect(toolbarSource).toContain("ObjectiveStatusesMenu");
    expect(toolbarSource).toContain("PrioritiesMenu");
    expect(toolbarSource).toContain("AssigneesMenu");
    expect(toolbarSource).toContain("HealthMenu");
    expect(toolbarSource).toContain("useBulkDeleteObjectivesMutation");
    expect(toolbarSource).not.toContain("ArchiveIcon");
  });
});
