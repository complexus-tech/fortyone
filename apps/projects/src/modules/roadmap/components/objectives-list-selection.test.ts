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

  it("shows key results on board cards and as collapsible list rows", () => {
    const boardSource = readSource(
      "src/modules/roadmap/components/objectives-board.tsx",
    );
    const keyResultSource = readSource(
      "src/modules/roadmap/components/roadmap-key-results.tsx",
    );
    const cardSource = readSource("src/modules/objectives/components/card.tsx");

    expect(boardSource).toContain("RoadmapKeyResultSummary");
    expect(boardSource).toContain("RoadmapObjectiveListItem");
    expect(keyResultSource).toContain("KeyResultContextMenu");
    expect(keyResultSource).toContain("RoadmapKeyResultRow");
    expect(cardSource).toContain("onToggleExpanded");
    expect(cardSource).toContain("childCount > 0");
  });

  it("opens the shared key-result details panel from roadmap surfaces", () => {
    const pageSource = readSource("src/modules/roadmap/index.tsx");
    const detailsSource = readSource(
      "src/modules/key-results/components/key-result-details.tsx",
    );

    expect(pageSource).toContain("selectKeyResult");
    expect(pageSource).toContain("KeyResultDetails");
    expect(detailsSource).toContain("Progress history");
    expect(detailsSource).toContain("Update progress");
    expect(detailsSource).toContain("AssigneesMenu");
    expect(detailsSource).toContain("DatePicker");
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
