/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { get, post, put } from "@/lib/http";
import type { ImportStoriesRequest } from "./api";
import {
  addExistingImportMemberToTeam,
  alignImportObjectiveToPillar,
  buildImportStoryRequests,
  createImportKeyResults,
  createImportLabel,
  createImportObjective,
  createImportSprint,
  createImportStrategicPillar,
  createImportStoryAssociation,
  createImportStoryLink,
  getImportObjectiveKeyResults,
  getImportObjectiveStatuses,
  getImportStoryAssociations,
  getImportStoryCollaboratorIds,
  getImportStoryLinks,
  getImportStrategyMap,
  getImportTeamLabels,
  getImportTeamMembers,
  getImportTeamObjectives,
  getImportTeamSprints,
  getImportWorkspaceLabels,
  getImportWorkspaceMembers,
  updateImportStoryCollaborators,
} from "./api";

jest.mock("@/lib/http", () => ({
  get: jest.fn(),
  post: jest.fn(),
  put: jest.fn(),
}));

const getMock = jest.mocked(get);
const postMock = jest.mocked(post);
const putMock = jest.mocked(put);
const ctx = { workspaceSlug: "acme" };

beforeEach(() => {
  jest.clearAllMocks();
});

const story = (description: string) => ({
  description,
  priority: "No Priority" as const,
  teamId: "team-1",
  title: "Imported task",
});

const requestFor = (items: ImportStoriesRequest["items"]) => ({
  items,
  provider: "file" as const,
  sourceDigest: "a".repeat(64),
});

describe("import API batching", () => {
  it("keeps every request within the server's 50-item contract", () => {
    const requests = buildImportStoryRequests(
      requestFor(
        Array.from({ length: 101 }, (_, index) => ({
          sourceKey: `row-${index + 1}`,
          story: story("Short description"),
        })),
      ),
    );

    expect(requests.map(({ items }) => items.length)).toEqual([50, 50, 1]);
  });

  it("splits heavily escaped descriptions below the request body limit", () => {
    const requests = buildImportStoryRequests(
      requestFor(
        Array.from({ length: 20 }, (_, index) => ({
          sourceKey: `row-${index + 1}`,
          story: story("\u0000".repeat(20_000)),
        })),
      ),
    );

    expect(requests.length).toBeGreaterThan(1);
    for (const request of requests) {
      expect(
        new TextEncoder().encode(JSON.stringify(request)).byteLength,
      ).toBeLessThanOrEqual(850 * 1024);
    }
    expect(requests.flatMap(({ items }) => items)).toHaveLength(20);
  });

  it("preserves a stable source namespace across every batch", () => {
    const requests = buildImportStoryRequests({
      ...requestFor(
        Array.from({ length: 51 }, (_, index) => ({
          sourceKey: `card-${index + 1}`,
          story: story("Trello card"),
        })),
      ),
      sourceNamespace: "trello:board:product-roadmap",
    });

    expect(requests).toHaveLength(2);
    expect(requests).toEqual(
      expect.arrayContaining([
        expect.objectContaining({
          sourceNamespace: "trello:board:product-roadmap",
        }),
      ]),
    );
    expect(
      requests.every(
        ({ sourceNamespace }) =>
          sourceNamespace === "trello:board:product-roadmap",
      ),
    ).toBe(true);
  });

  it("preserves reviewed story effort fields in the import request", () => {
    const requests = buildImportStoryRequests(
      requestFor([
        {
          sourceKey: "task-1",
          story: {
            ...story("Explicit source effort"),
            estimateValue: 5,
            estimatedDurationMinutes: 90,
            minimumFocusBlockMinutes: 30,
          },
        },
      ]),
    );

    expect(requests[0]?.items[0]?.story).toMatchObject({
      estimateValue: 5,
      estimatedDurationMinutes: 90,
      minimumFocusBlockMinutes: 30,
    });
  });
});

describe("multi-entity import API helpers", () => {
  it("lists members and destination entities with workspace-scoped routes", async () => {
    getMock.mockResolvedValue({ data: [] });

    await expect(getImportWorkspaceMembers(ctx)).resolves.toEqual([]);
    await expect(getImportTeamMembers("team-1", ctx)).resolves.toEqual([]);
    await expect(getImportObjectiveStatuses(ctx)).resolves.toEqual([]);
    await expect(getImportTeamObjectives("team-1", ctx)).resolves.toEqual([]);
    await expect(
      getImportObjectiveKeyResults("objective-1", ctx),
    ).resolves.toEqual([]);
    await expect(getImportTeamLabels("team-1", ctx)).resolves.toEqual([]);
    await expect(getImportTeamSprints("team-1", ctx)).resolves.toEqual([]);
    await expect(getImportWorkspaceLabels(ctx)).resolves.toEqual([]);

    expect(getMock.mock.calls).toEqual([
      ["members", ctx],
      ["members?teamId=team-1", ctx],
      ["objective-statuses", ctx],
      ["objectives?teamId=team-1", ctx],
      ["objectives/objective-1/key-results", ctx],
      ["labels?teamId=team-1", ctx],
      ["sprints?teamId=team-1", ctx],
      ["labels", ctx],
    ]);
  });

  it("only adds an already-resolved workspace member to a team", async () => {
    postMock.mockResolvedValue({ data: { teamId: "team-1" } });

    await expect(
      addExistingImportMemberToTeam("team-1", "member-1", ctx),
    ).resolves.toEqual({ teamId: "team-1" });
    expect(postMock).toHaveBeenCalledWith(
      "teams/team-1/members",
      { userId: "member-1" },
      ctx,
    );
  });

  it("replaces story collaborators through the idempotent relationship endpoint", async () => {
    getMock.mockResolvedValue({
      data: { collaboratorIds: ["existing-member"] },
    });
    putMock.mockResolvedValue({ data: null });

    await expect(
      getImportStoryCollaboratorIds("story-1", ctx),
    ).resolves.toEqual(["existing-member"]);
    await expect(
      updateImportStoryCollaborators("story-1", ["member-1", "member-2"], ctx),
    ).resolves.toBeUndefined();
    expect(getMock).toHaveBeenCalledWith("stories/story-1", ctx);
    expect(putMock).toHaveBeenCalledWith(
      "stories/story-1/collaborators",
      { collaboratorIds: ["member-1", "member-2"] },
      ctx,
    );
  });

  it("lists and creates story associations through existing story routes", async () => {
    const association = {
      id: "association-1",
      fromStoryId: "story-1",
      toStoryId: "story-2",
      type: "blocking" as const,
    };
    getMock.mockResolvedValue({ data: { associations: [association] } });
    postMock.mockResolvedValue({ data: association });

    await expect(getImportStoryAssociations("story-1", ctx)).resolves.toEqual([
      association,
    ]);
    await expect(
      createImportStoryAssociation(
        "story-1",
        { toStoryId: "story-2", type: "blocking" },
        ctx,
      ),
    ).resolves.toEqual(association);
    expect(getMock).toHaveBeenCalledWith("stories/story-1", ctx);
    expect(postMock).toHaveBeenCalledWith(
      "stories/story-1/associations",
      { toStoryId: "story-2", type: "blocking" },
      ctx,
    );
  });

  it("lists and creates external story links through existing link routes", async () => {
    const link = {
      id: "link-1",
      storyId: "story-1",
      title: "Original Trello card",
      url: "https://trello.com/c/example",
      createdAt: "2026-09-01T00:00:00Z",
      updatedAt: "2026-09-01T00:00:00Z",
    };
    getMock.mockResolvedValue({ data: [link] });
    postMock.mockResolvedValue({ data: link });

    await expect(getImportStoryLinks("story-1", ctx)).resolves.toEqual([link]);
    await expect(
      createImportStoryLink(
        {
          storyId: "story-1",
          title: "Original Trello card",
          url: "https://trello.com/c/example",
        },
        ctx,
      ),
    ).resolves.toEqual(link);
    expect(getMock).toHaveBeenCalledWith("stories/story-1/links", ctx);
    expect(postMock).toHaveBeenCalledWith(
      "links",
      {
        storyId: "story-1",
        title: "Original Trello card",
        url: "https://trello.com/c/example",
      },
      ctx,
    );
  });

  it("uses existing create contracts for objectives, key results, labels, and sprints", async () => {
    const objectiveInput = {
      name: "Improve reliability",
      statusId: "objective-status-1",
      teamId: "team-1",
    };
    const keyResults = [
      {
        name: "Reach 99.9% availability",
        measurementType: "percentage" as const,
        startValue: 99,
        currentValue: 99,
        targetValue: 99.9,
        startDate: "2026-01-01",
        endDate: "2026-03-31",
      },
    ];
    const labelInput = {
      color: "#4A90E2",
      name: "Migration",
      teamId: "team-1",
    };
    const sprintInput = {
      name: "Migration sprint",
      teamId: "team-1",
      startDate: "2026-01-01",
      endDate: "2026-01-14",
    };
    postMock
      .mockResolvedValueOnce({
        data: { objective: { id: "objective-1" }, keyResults: [] },
      })
      .mockResolvedValueOnce({ data: [] })
      .mockResolvedValueOnce({ data: { id: "label-1" } })
      .mockResolvedValueOnce({ data: { id: "sprint-1" } });

    await createImportObjective(objectiveInput, ctx);
    await createImportKeyResults("objective-1", keyResults, ctx);
    await createImportLabel(labelInput, ctx);
    await createImportSprint(sprintInput, ctx);

    expect(postMock.mock.calls).toEqual([
      ["objectives", objectiveInput, ctx],
      ["objectives/objective-1/key-results", { keyResults }, ctx],
      ["labels", labelInput, ctx],
      ["sprints", sprintInput, ctx],
    ]);
  });

  it("loads, creates, and aligns strategic pillars through the strategy map", async () => {
    getMock.mockResolvedValue({
      data: {
        ultimateGoal: "Grow sustainably",
        description: null,
        pillars: [
          {
            id: "pillar-1",
            name: "Customer growth",
            description: null,
            orderIndex: 0,
            objectiveIds: null,
          },
        ],
      },
    });
    postMock.mockResolvedValue({
      data: {
        id: "pillar-2",
        name: "Reliability",
        description: "Protect customer trust",
        orderIndex: 1,
        objectiveIds: [],
      },
    });
    putMock.mockResolvedValue({ data: null });

    await expect(getImportStrategyMap(ctx)).resolves.toMatchObject({
      pillars: [{ id: "pillar-1", objectiveIds: [] }],
    });
    await expect(
      createImportStrategicPillar(
        {
          name: "Reliability",
          description: "Protect customer trust",
          orderIndex: 1,
        },
        ctx,
      ),
    ).resolves.toMatchObject({ id: "pillar-2" });
    await expect(
      alignImportObjectiveToPillar("objective-1", "pillar-2", ctx),
    ).resolves.toBeUndefined();

    expect(getMock).toHaveBeenCalledWith("strategy-map", ctx);
    expect(postMock).toHaveBeenCalledWith(
      "strategy-map/pillars",
      {
        name: "Reliability",
        description: "Protect customer trust",
        orderIndex: 1,
      },
      ctx,
    );
    expect(putMock).toHaveBeenCalledWith(
      "strategy-map/objectives/objective-1",
      { pillarId: "pillar-2" },
      ctx,
    );
  });

  it("surfaces API error messages and missing response data", async () => {
    getMock
      .mockResolvedValueOnce({ error: { message: "Membership denied" } })
      .mockResolvedValueOnce({});

    await expect(getImportWorkspaceMembers(ctx)).rejects.toThrow(
      "Membership denied",
    );
    await expect(getImportObjectiveStatuses(ctx)).rejects.toThrow(
      "Unable to load objective statuses",
    );

    putMock.mockResolvedValue({ error: { message: "Collaborator denied" } });
    await expect(
      updateImportStoryCollaborators("story-1", ["member-1"], ctx),
    ).rejects.toThrow("Collaborator denied");
  });
});
