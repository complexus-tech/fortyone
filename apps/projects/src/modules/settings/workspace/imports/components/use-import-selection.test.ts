/* global expect, it -- Jest globals are provided by the projects test runner. */

import { act, renderHook } from "@testing-library/react";
import type { ImportDraft, ImportTask } from "../schema";
import { useImportSelection } from "./use-import-selection";

const task = (index: number): ImportTask => ({
  sourceId: `card-${index}`,
  title: `Card ${index}`,
  description: "",
  priority: "No Priority",
  status: null,
  statusCategory: null,
  estimateValue: null,
  estimatedDurationMinutes: null,
  minimumFocusBlockMinutes: null,
  assigneeEmail: null,
  assigneeName: null,
  assigneePersonSourceId: null,
  teamSourceId: null,
  parentSourceId: null,
  objectiveSourceId: null,
  keyResultSourceId: null,
  sprintSourceId: null,
  startDate: null,
  endDate: null,
  associations: [],
  collaboratorPersonSourceIds: [],
  labelSourceIds: [],
  links: [],
});

it("preserves manual card exclusions when archived cards are shown and hidden", () => {
  const draft: ImportDraft = {
    sourceType: "json",
    sourceNamespace: "trello:board:product",
    sourceMetadata: {
      archivedTaskSourceIds: ["card-0", "card-54"],
      nestedChecklistItemCount: 0,
      platform: "trello",
    },
    summary: "Imported board",
    warnings: [],
    mapping: null,
    teams: [],
    people: [],
    labels: [],
    strategicPillars: [],
    objectives: [],
    keyResults: [],
    sprints: [],
    tasks: Array.from({ length: 55 }, (_, index) => task(index)),
    columns: [],
    rows: [],
    fileHash: "board-digest",
    fileName: "board.json",
  };
  const { result } = renderHook(() =>
    useImportSelection({
      draft,
      destination: { kind: "existing", teamId: "team" },
      structureMode: "single",
      knownWorkspaceTeams: [],
      fallbackTargetPlan: {
        teamConflict: false,
        teamId: "team",
        teamKey: "team",
        teamLabel: "Team",
      },
      sourceTeamTargetPlans: new Map(),
    }),
  );

  act(() => {
    result.current.initializeArchivedRows(draft);
  });
  expect(result.current.selectedTasks).toHaveLength(53);
  expect(result.current.reviewPageCount).toBe(2);

  act(() => {
    result.current.toggleTask(10, false);
    result.current.setReviewPage(1);
  });
  act(() => {
    result.current.toggleArchivedTrelloCards(true);
  });
  expect(result.current.excludedRows).toEqual(new Set([10]));
  expect(result.current.selectedTasks).toHaveLength(54);
  expect(result.current.reviewPage).toBe(0);

  act(() => {
    result.current.toggleArchivedTrelloCards(false);
  });
  expect(result.current.excludedRows).toEqual(new Set([0, 10, 54]));
  expect(result.current.selectedTasks).toHaveLength(52);
  expect(
    result.current.reviewTasks.map(({ taskIndex }) => taskIndex),
  ).not.toContain(54);
  expect(
    result.current.reviewTasks.map(({ taskIndex }) => taskIndex),
  ).toContain(10);
});
