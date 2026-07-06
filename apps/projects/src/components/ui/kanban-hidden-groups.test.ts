/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import type { StoriesViewOptions } from "./stories-view-options-button";
import {
  getHiddenKanbanGroupKeys,
  hideKanbanGroup,
  showKanbanGroup,
} from "./kanban-hidden-groups";

const viewOptions: StoriesViewOptions = {
  groupBy: "status",
  orderBy: "created",
  showEmptyGroups: true,
  showSubStories: false,
  displayColumns: ["ID", "Status"],
};

describe("kanban hidden groups", () => {
  it("hides groups for the active grouping only", () => {
    const nextViewOptions = hideKanbanGroup(viewOptions, "status-started");

    expect(getHiddenKanbanGroupKeys(nextViewOptions)).toEqual([
      "status-started",
    ]);
    expect(
      getHiddenKanbanGroupKeys({ ...nextViewOptions, groupBy: "priority" }),
    ).toEqual([]);
  });

  it("does not duplicate hidden groups", () => {
    const nextViewOptions = hideKanbanGroup(
      hideKanbanGroup(viewOptions, "status-started"),
      "status-started",
    );

    expect(getHiddenKanbanGroupKeys(nextViewOptions)).toEqual([
      "status-started",
    ]);
  });

  it("restores groups for the active grouping only", () => {
    const hiddenViewOptions = {
      ...viewOptions,
      hiddenKanbanGroups: {
        status: ["status-started"],
        priority: ["High"],
      },
    };

    expect(
      getHiddenKanbanGroupKeys(
        showKanbanGroup(hiddenViewOptions, "status-started"),
      ),
    ).toEqual([]);
    expect(
      getHiddenKanbanGroupKeys({
        ...showKanbanGroup(hiddenViewOptions, "status-started"),
        groupBy: "priority",
      }),
    ).toEqual(["High"]);
  });
});
