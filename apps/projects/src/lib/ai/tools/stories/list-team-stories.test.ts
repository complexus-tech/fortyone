/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import { buildGroupedStoriesQuery } from "@/modules/stories/utils/query-builders";
import { listTeamStoriesInputSchema } from "./list-team-stories";

jest.mock("ai", () => ({
  tool: (definition: unknown) => definition,
}));

describe("listTeamStoriesInputSchema", () => {
  it("produces a supported query for a current-user story request", () => {
    const input = listTeamStoriesInputSchema.parse({
      filters: {
        assignedToMe: true,
        assigneeIds: [],
        epicId: "",
        objectiveId: "",
        parentId: "  ",
        statusIds: [],
        teamIds: [],
        titleContains: "",
      },
    });

    expect(input).toEqual({
      filters: {
        assignedToMe: true,
        assigneeIds: [],
        objectiveId: "",
        parentId: "  ",
        statusIds: [],
        storiesPerGroup: 20,
        teamIds: [],
        titleContains: "",
      },
      groupBy: "status",
    });
    expect(
      buildGroupedStoriesQuery({
        groupBy: input.groupBy,
        ...input.filters,
      }),
    ).toBe("?groupBy=status&assignedToMe=true&storiesPerGroup=20");
  });
});
