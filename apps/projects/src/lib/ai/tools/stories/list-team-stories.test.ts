/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import { buildGroupedStoriesQuery } from "@/modules/stories/utils/query-builders";
import type { Story } from "@/modules/stories/types";
import {
  filterStoriesByAssignedBy,
  listTeamStoriesInputSchema,
} from "./list-team-stories";

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

describe("filterStoriesByAssignedBy", () => {
  const story = (id: string, assignedBy?: Story["assignedBy"]): Story =>
    ({ id, assignedBy }) as Story;

  it("resolves a unique first-name match and excludes unknown attribution", () => {
    const daria = {
      id: "daria-1",
      username: "daria",
      fullName: "Daria Jones",
      avatarUrl: null,
      isActive: true,
      isSystem: false,
    };
    const result = filterStoriesByAssignedBy(
      [
        story("one", daria),
        story("legacy"),
        story("two", {
          ...daria,
          id: "sam-1",
          username: "sam",
          fullName: "Sam Lee",
        }),
      ],
      "Daria",
    );

    expect(result.resolved?.id).toBe("daria-1");
    expect(result.stories.map(({ id }) => id)).toEqual(["one"]);
  });

  it("returns candidates instead of guessing between matching people", () => {
    const result = filterStoriesByAssignedBy(
      [
        story("one", {
          id: "daria-1",
          username: "daria.j",
          fullName: "Daria Jones",
          avatarUrl: null,
          isActive: true,
          isSystem: false,
        }),
        story("two", {
          id: "daria-2",
          username: "daria.m",
          fullName: "Daria Morgan",
          avatarUrl: null,
          isActive: true,
          isSystem: false,
        }),
      ],
      "Daria",
    );

    expect(result.resolved).toBeUndefined();
    expect(result.stories).toEqual([]);
    expect(result.candidates).toHaveLength(2);
  });
});
