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
    const propertiesStart = boardSource.indexOf(
      '<Flex align="center" className="mt-1 gap-1.5" wrap>',
    );
    const properties = boardSource.slice(
      propertiesStart,
      boardSource.indexOf("</Flex>", propertiesStart),
    );

    expect(boardSource).toContain("RoadmapKeyResultSummary");
    expect(properties).toContain("RoadmapKeyResultSummary");
    expect(boardSource).toContain("RoadmapObjectiveListItem");
    expect(boardSource).not.toContain("TableHeader");
    expect(keyResultSource).toContain("KeyResultContextMenu");
    expect(keyResultSource).toContain("RoadmapKeyResultRow");
    expect(keyResultSource).toContain("getKeyResultReference");
    expect(keyResultSource).toContain(
      'className="px-5 py-2.5 md:pr-12 md:pl-18"',
    );
    expect(keyResultSource).toContain("formatKeyResultValue");
    expect(keyResultSource).toContain("<CircleProgressBar");
    expect(keyResultSource).not.toContain("<ProgressBar ");
    expect(keyResultSource).toContain("py-[0.925rem]");
    expect(keyResultSource).toContain("h-[1.1rem] w-[1.1rem] shrink-0");
    expect(keyResultSource).not.toContain("bg-surface-muted/45");
    expect(keyResultSource).not.toContain("pr-4 text-[0.9375rem]");
    expect(keyResultSource).toContain('size="sm"');
    expect(keyResultSource).toContain("keyResult.endDate");
    expect(keyResultSource).toContain(
      'keyResult.measurementType === "boolean"',
    );
    expect(keyResultSource).not.toContain("w-[104px] truncate");
    expect(keyResultSource).toContain("keyResultProgress");
    expect(cardSource).toContain("onToggleExpanded");
    expect(cardSource).toContain("childCount > 0");
    expect(cardSource).toContain("<CircleProgressBar");
    expect(cardSource).toContain("strokeWidth={2.8}");
    expect(cardSource).toContain("objectiveReference");
    expect(cardSource).toContain('className="gap-1 pr-2"');
    expect(cardSource).toContain('rounded="md"');
    expect(cardSource).toContain('size="xs"');
    expect(cardSource).toContain('variant="outline"');
    expect(cardSource).toContain('className="hidden @6xl:inline"');

    const objectivePropertiesStart = cardSource.indexOf(
      '<Flex align="center" className="shrink-0 gap-2">',
    );
    const objectivePropertiesEnd = cardSource.indexOf(
      "</Flex>\n    </RowWrapper>",
      objectivePropertiesStart,
    );
    const rowProperties = cardSource.slice(
      objectivePropertiesStart,
      objectivePropertiesEnd,
    );

    expect(rowProperties.indexOf("ObjectiveStatusesMenu")).toBeLessThan(
      rowProperties.indexOf("ObjectiveHealthEditor"),
    );
    expect(rowProperties.indexOf("ObjectiveHealthEditor")).toBeLessThan(
      rowProperties.indexOf("PrioritiesMenu"),
    );
    expect(rowProperties).toContain("border-success/20 bg-success/10");
    expect(rowProperties).toContain("border-warning/20 bg-warning/10");
    expect(rowProperties).toContain("border-danger/20 bg-danger/10");
    expect(rowProperties.indexOf("PrioritiesMenu")).toBeLessThan(
      rowProperties.indexOf("CircleProgressBar"),
    );
    expect(rowProperties.indexOf("CircleProgressBar")).toBeLessThan(
      rowProperties.indexOf("DatePicker"),
    );
    expect(rowProperties.indexOf("DatePicker")).toBeLessThan(
      rowProperties.indexOf("AssigneesMenu"),
    );
  });

  it("matches the list forecast badge to the objective property controls", () => {
    const cardSource = readSource("src/modules/objectives/components/card.tsx");
    const forecastSource = readSource(
      "src/modules/objectives/components/objective-forecast-risk.tsx",
    );

    expect(cardSource).toContain('size="row"');
    expect(forecastSource).toContain("h-[2.1rem] rounded-xl px-2 text-[1rem]");
  });

  it("opens the shared key-result details panel from roadmap surfaces", () => {
    const viewsSource = readSource(
      "src/modules/roadmap/components/objective-views.tsx",
    );
    const detailsSource = readSource(
      "src/modules/key-results/components/key-result-details.tsx",
    );

    expect(viewsSource).toContain("selectKeyResult");
    expect(viewsSource).toContain("KeyResultDetails");
    expect(detailsSource).toContain("Progress history");
    expect(detailsSource).toContain("Update progress");
    expect(detailsSource).toContain("AssigneesMenu");
    expect(detailsSource).toContain("DatePicker");
    expect(detailsSource).toContain('label="Objective"');
    expect(detailsSource).not.toContain('label="Parent objective"');
  });

  it("reuses roadmap views for team-scoped objectives", () => {
    const roadmapSource = readSource("src/modules/roadmap/index.tsx");
    const layoutStateSource = readSource(
      "src/modules/roadmap/use-roadmap-layout.ts",
    );
    const teamObjectivesSource = readSource("src/modules/objectives/index.tsx");
    const teamHeaderSource = readSource(
      "src/modules/objectives/components/team-header.tsx",
    );

    expect(roadmapSource).toContain("getRoadmapLayoutLabel(layout)");
    expect(roadmapSource).toContain("useRoadmapLayout()");
    expect(layoutStateSource).toContain('useQueryState(\n    "view"');
    expect(layoutStateSource).toContain('history: "push"');
    expect(teamObjectivesSource).toContain("useTeamObjectives(teamId)");
    expect(teamObjectivesSource).toContain("<ObjectiveViews");
    expect(teamObjectivesSource).toContain("useRoadmapLayout()");
    expect(teamHeaderSource).toContain("<RoadmapLayoutSwitcher");
    expect(teamHeaderSource).toContain("<ObjectiveViewOptionsButton");
    expect(teamHeaderSource).toContain("getRoadmapLayoutLabel(layout)");
  });

  it("keeps linked work in details rather than the edit dialog", () => {
    const detailsSource = readSource(
      "src/modules/key-results/components/key-result-details.tsx",
    );
    const dialogSource = readSource(
      "src/modules/objectives/stories/overview/update-key-result-dialog.tsx",
    );

    expect(detailsSource).toContain("Linked");
    expect(dialogSource).not.toContain("linkedStories");
    expect(dialogSource).not.toContain("Loading linked work");
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
