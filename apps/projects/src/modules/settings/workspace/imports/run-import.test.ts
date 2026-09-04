/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import type { Team } from "@/modules/teams/types";
import type { ImportDraft, ImportTask } from "./schema";
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
  createImportTeam,
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
  getImportTeamStatuses,
  getImportWorkspaceLabels,
  getImportWorkspaceMembers,
  importStoriesBatch,
  updateImportStoryCollaborators,
} from "./api";
import {
  getImportSourceTeamDestination,
  resolveImportSourceTeam,
  runImport,
} from "./run-import";

jest.mock("./api", () => ({
  addExistingImportMemberToTeam: jest.fn(),
  alignImportObjectiveToPillar: jest.fn(),
  buildImportStoryRequests: jest.fn(),
  createImportKeyResults: jest.fn(),
  createImportLabel: jest.fn(),
  createImportObjective: jest.fn(),
  createImportSprint: jest.fn(),
  createImportStrategicPillar: jest.fn(),
  createImportStoryAssociation: jest.fn(),
  createImportStoryLink: jest.fn(),
  createImportTeam: jest.fn(),
  getImportObjectiveKeyResults: jest.fn(),
  getImportObjectiveStatuses: jest.fn(),
  getImportStoryAssociations: jest.fn(),
  getImportStoryCollaboratorIds: jest.fn(),
  getImportStoryLinks: jest.fn(),
  getImportStrategyMap: jest.fn(),
  getImportTeamLabels: jest.fn(),
  getImportTeamMembers: jest.fn(),
  getImportTeamObjectives: jest.fn(),
  getImportTeamSprints: jest.fn(),
  getImportTeamStatuses: jest.fn(),
  getImportWorkspaceLabels: jest.fn(),
  getImportWorkspaceMembers: jest.fn(),
  importStoriesBatch: jest.fn(),
  updateImportStoryCollaborators: jest.fn(),
}));

const buildRequestsMock = jest.mocked(buildImportStoryRequests);
const addExistingMemberMock = jest.mocked(addExistingImportMemberToTeam);
const alignObjectiveToPillarMock = jest.mocked(alignImportObjectiveToPillar);
const createKeyResultsMock = jest.mocked(createImportKeyResults);
const createLabelMock = jest.mocked(createImportLabel);
const createObjectiveMock = jest.mocked(createImportObjective);
const createSprintMock = jest.mocked(createImportSprint);
const createStrategicPillarMock = jest.mocked(createImportStrategicPillar);
const createStoryAssociationMock = jest.mocked(createImportStoryAssociation);
const createStoryLinkMock = jest.mocked(createImportStoryLink);
const createTeamMock = jest.mocked(createImportTeam);
const getObjectiveKeyResultsMock = jest.mocked(getImportObjectiveKeyResults);
const getObjectiveStatusesMock = jest.mocked(getImportObjectiveStatuses);
const getStoryAssociationsMock = jest.mocked(getImportStoryAssociations);
const getStoryCollaboratorIdsMock = jest.mocked(getImportStoryCollaboratorIds);
const getStoryLinksMock = jest.mocked(getImportStoryLinks);
const getStrategyMapMock = jest.mocked(getImportStrategyMap);
const getTeamLabelsMock = jest.mocked(getImportTeamLabels);
const getTeamMembersMock = jest.mocked(getImportTeamMembers);
const getTeamObjectivesMock = jest.mocked(getImportTeamObjectives);
const getTeamSprintsMock = jest.mocked(getImportTeamSprints);
const getTeamStatusesMock = jest.mocked(getImportTeamStatuses);
const getWorkspaceLabelsMock = jest.mocked(getImportWorkspaceLabels);
const getWorkspaceMembersMock = jest.mocked(getImportWorkspaceMembers);
const importStoriesBatchMock = jest.mocked(importStoriesBatch);
const updateStoryCollaboratorsMock = jest.mocked(
  updateImportStoryCollaborators,
);

const ctx = { workspaceSlug: "workspace" };

const task = (sourceId: string, title: string): ImportTask => ({
  sourceId,
  title,
  description: "",
  status: null,
  statusCategory: null,
  priority: "No Priority",
  estimateValue: null,
  estimatedDurationMinutes: null,
  minimumFocusBlockMinutes: null,
  assigneeEmail: null,
  assigneeName: null,
  assigneePersonSourceId: null,
  associations: [],
  collaboratorPersonSourceIds: [],
  teamSourceId: null,
  parentSourceId: null,
  objectiveSourceId: null,
  keyResultSourceId: null,
  sprintSourceId: null,
  labelSourceIds: [],
  links: [],
  startDate: null,
  endDate: null,
});

const draft = (overrides: Partial<ImportDraft> = {}): ImportDraft => ({
  sourceType: "json",
  sourceNamespace: null,
  summary: "Import test",
  warnings: [],
  mapping: null,
  teams: [],
  people: [],
  strategicPillars: [],
  labels: [],
  objectives: [],
  keyResults: [],
  sprints: [],
  tasks: [],
  columns: [],
  fileHash: "a".repeat(64),
  fileName: "import.json",
  rows: [],
  ...overrides,
});

const importObjective = (
  sourceId: string,
  name: string,
  overrides: Partial<ImportDraft["objectives"][number]> = {},
): ImportDraft["objectives"][number] => ({
  sourceId,
  name,
  description: null,
  shortSummary: null,
  color: null,
  isPrivate: false,
  status: null,
  statusCategory: null,
  priority: "No Priority",
  pillarSourceId: null,
  leadPersonSourceId: null,
  teamSourceId: null,
  startDate: null,
  endDate: null,
  ...overrides,
});

const importKeyResult = (
  sourceId: string,
  objectiveSourceId: string,
  name: string,
): ImportDraft["keyResults"][number] => ({
  sourceId,
  objectiveSourceId,
  name,
  measurementType: "percentage",
  startValue: 0,
  currentValue: 25,
  targetValue: 100,
  leadPersonSourceId: null,
  contributorPersonSourceIds: [],
  startDate: "2026-09-01",
  endDate: "2026-09-30",
});

const run = (
  input: Partial<Parameters<typeof runImport>[0]> &
    Pick<Parameters<typeof runImport>[0], "draft">,
) =>
  runImport({
    actorUserId: "actor-user",
    confirmedMemberIdsByIdentityKey: new Map(),
    ctx,
    existingTeams: [],
    fallbackTeamCode: "DST",
    fallbackTeamCreated: false,
    fallbackTeamId: "destination-team",
    fallbackTeamIsPrivate: false,
    fallbackTeamIsNew: false,
    fallbackTeamName: "Destination",
    forceCreateObjectiveSourceIds: new Set(),
    joinedTeamIds: new Set(["destination-team"]),
    onProgress: jest.fn(),
    selectedObjectiveSourceIds: new Set(),
    selectedStrategicPillarSourceIds: new Set(),
    selectedTaskIndexes: new Set(),
    sourceObjectiveCache: new Map(),
    sourceTeamCache: new Map(),
    structureMode: "single",
    ...input,
  });

beforeEach(() => {
  jest.clearAllMocks();
  buildRequestsMock.mockImplementation((request) => [request]);
  getObjectiveStatusesMock.mockResolvedValue([]);
  getWorkspaceMembersMock.mockResolvedValue([]);
  getTeamStatusesMock.mockResolvedValue([
    {
      id: "todo",
      name: "Todo",
      category: "unstarted",
      isDefault: true,
      orderIndex: 1,
    },
  ] as never);
  getTeamMembersMock.mockResolvedValue([]);
  getTeamObjectivesMock.mockResolvedValue([]);
  getTeamLabelsMock.mockResolvedValue([]);
  getTeamSprintsMock.mockResolvedValue([]);
  getWorkspaceLabelsMock.mockResolvedValue([]);
  getStrategyMapMock.mockResolvedValue({
    ultimateGoal: "",
    description: null,
    pillars: [],
  });
  alignObjectiveToPillarMock.mockResolvedValue(undefined);
  getStoryAssociationsMock.mockResolvedValue([]);
  getStoryCollaboratorIdsMock.mockResolvedValue([]);
  getStoryLinksMock.mockResolvedValue([]);
  createStoryAssociationMock.mockResolvedValue({} as never);
  updateStoryCollaboratorsMock.mockResolvedValue(undefined);
  importStoriesBatchMock.mockImplementation(async (request) => ({
    data: {
      counts: {
        total: request.items.length,
        created: request.items.length,
        replayed: 0,
        failed: 0,
      },
      items: request.items.map(({ sourceKey }, index) => ({
        sourceKey,
        storyId: `story-${index + 1}`,
        created: true,
        error: null,
      })),
    },
  }));
});

describe("multi-entity import runner", () => {
  it("finishes parent collaborator reconciliation before importing its child", async () => {
    let releaseCollaborators = () => {};
    const collaboratorsFinished = new Promise<void>((resolve) => {
      releaseCollaborators = resolve;
    });
    let notifyCollaboratorsStarted = () => {};
    const collaboratorsStarted = new Promise<void>((resolve) => {
      notifyCollaboratorsStarted = resolve;
    });
    updateStoryCollaboratorsMock.mockImplementationOnce(async () => {
      notifyCollaboratorsStarted();
      await collaboratorsFinished;
    });
    getWorkspaceMembersMock.mockResolvedValue([
      {
        id: "collaborator-id",
        email: "collaborator@example.com",
        fullName: "Collaborator",
        username: "collaborator",
        isActive: true,
        isSystem: false,
      },
    ] as never);
    const onProgress = jest.fn();
    const importRun = run({
      draft: draft({
        people: [
          {
            sourceId: "collaborator",
            name: "Collaborator",
            email: "collaborator@example.com",
            teamSourceIds: [],
          },
        ],
        tasks: [
          { ...task("child", "Child"), parentSourceId: "parent" },
          {
            ...task("parent", "Parent"),
            collaboratorPersonSourceIds: ["collaborator"],
          },
        ],
      }),
      selectedTaskIndexes: new Set([0, 1]),
      onProgress,
    });

    await collaboratorsStarted;
    try {
      expect(importStoriesBatchMock).toHaveBeenCalledTimes(1);
      expect(importStoriesBatchMock.mock.calls[0][0].items).toEqual([
        expect.objectContaining({ sourceKey: "parent" }),
      ]);
      expect(updateStoryCollaboratorsMock).toHaveBeenCalledWith(
        "story-1",
        ["collaborator-id"],
        ctx,
      );
    } finally {
      releaseCollaborators();
    }

    const result = await importRun;
    expect(importStoriesBatchMock).toHaveBeenCalledTimes(2);
    expect(importStoriesBatchMock.mock.calls[1][0].items).toEqual([
      expect.objectContaining({
        sourceKey: "child",
        story: expect.objectContaining({ parentId: "story-1" }),
      }),
    ]);
    expect(result).toMatchObject({
      created: 2,
      failed: 0,
      appliedCollaborators: 1,
      addedMemberships: 1,
    });
    expect(onProgress).toHaveBeenLastCalledWith(100);
  });

  it("blocks a failed parent's descendants while preserving independent results", async () => {
    importStoriesBatchMock.mockImplementationOnce(async (request) => ({
      data: {
        counts: { total: 2, created: 1, replayed: 0, failed: 1 },
        items: request.items.map(({ sourceKey }) => ({
          sourceKey,
          storyId: sourceKey === "parent" ? null : "independent-story",
          created: sourceKey !== "parent",
          error:
            sourceKey === "parent"
              ? { code: "invalid_story", message: "Parent rejected" }
              : null,
        })),
      },
    }));

    const result = await run({
      draft: draft({
        tasks: [
          { ...task("child", "Child"), parentSourceId: "parent" },
          task("parent", "Parent"),
          task("independent", "Independent"),
        ],
      }),
      selectedTaskIndexes: new Set([0, 1, 2]),
    });

    expect(importStoriesBatchMock).toHaveBeenCalledTimes(1);
    expect(result).toMatchObject({ created: 1, failed: 2, replayed: 0 });
    expect(result.items).toContainEqual({
      sourceKey: "child",
      storyId: null,
      created: false,
      error: {
        code: "parent_import_failed",
        message: "The parent work item could not be imported.",
      },
    });
  });

  it("adds an exact workspace-email match to the imported team", async () => {
    getWorkspaceMembersMock.mockResolvedValue([
      {
        id: "member-1",
        email: "owner@example.com",
        fullName: "Existing Owner",
        username: "owner",
        isActive: true,
        isSystem: false,
      },
    ] as never);

    const result = await run({
      draft: draft({
        people: [
          {
            sourceId: "person-1",
            name: "Source Owner",
            email: "OWNER@example.com",
            teamSourceIds: ["source-team"],
          },
        ],
      }),
    });

    expect(addExistingMemberMock).toHaveBeenCalledWith(
      "destination-team",
      "member-1",
      ctx,
    );
    expect(result.addedMemberships).toBe(1);
  });

  it("keeps an explicit member selection authoritative during execution", async () => {
    const selectedMember = {
      id: "selected-member",
      email: "selected@example.com",
      fullName: "Selected Member",
      username: "selected",
      isActive: true,
      isSystem: false,
    };
    const emailMember = {
      id: "email-member",
      email: "source@example.com",
      fullName: "Email Member",
      username: "email-member",
      isActive: true,
      isSystem: false,
    };
    getWorkspaceMembersMock.mockResolvedValue([
      selectedMember,
      emailMember,
    ] as never);
    getTeamMembersMock.mockResolvedValue([emailMember] as never);
    const assignedTask = {
      ...task("assigned-task", "Explicit member assignment"),
      assigneePersonSourceId: "person-1",
      assigneeEmail: emailMember.email,
      assigneeName: emailMember.fullName,
    };

    const result = await run({
      confirmedMemberIdsByIdentityKey: new Map([
        ["source:person-1", selectedMember.id],
      ]),
      draft: draft({ tasks: [assignedTask] }),
      selectedTaskIndexes: new Set([0]),
    });

    expect(result.unresolvedPeople).toBe(0);
    expect(buildRequestsMock.mock.calls[0]?.[0].items[0]?.story).toMatchObject({
      assigneeId: selectedMember.id,
    });
  });

  it("keeps an explicit leave-unmapped choice authoritative", async () => {
    const emailMember = {
      id: "email-member",
      email: "source@example.com",
      fullName: "Email Member",
      username: "email-member",
      isActive: true,
      isSystem: false,
    };
    getWorkspaceMembersMock.mockResolvedValue([emailMember] as never);
    getTeamMembersMock.mockResolvedValue([emailMember] as never);
    const assignedTask = {
      ...task("assigned-task", "Unmapped assignment"),
      assigneePersonSourceId: "person-1",
      assigneeEmail: emailMember.email,
      assigneeName: emailMember.fullName,
    };

    const result = await run({
      confirmedMemberIdsByIdentityKey: new Map([["source:person-1", null]]),
      draft: draft({ tasks: [assignedTask] }),
      selectedTaskIndexes: new Set([0]),
    });

    expect(result.unresolvedPeople).toBe(1);
    expect(
      buildRequestsMock.mock.calls[0]?.[0].items[0]?.story,
    ).not.toHaveProperty("assigneeId");
  });

  it("counts an unresolved source-team member once without story assignments", async () => {
    const unresolvedPerson = {
      sourceId: "missing-member",
      name: "Missing Member",
      email: null,
      teamSourceIds: ["source-team", "source-team"],
    };

    const result = await run({
      draft: draft({ people: [unresolvedPerson] }),
    });

    expect(result.unresolvedPeople).toBe(1);
  });

  it("requires explicit confirmation before using a unique name-only match", async () => {
    const workspaceMember = {
      id: "member-1",
      email: "ada@example.com",
      fullName: "Ada Lovelace",
      username: "ada",
      isActive: true,
      isSystem: false,
    };
    getWorkspaceMembersMock.mockResolvedValue([workspaceMember] as never);
    const person = {
      sourceId: "person-1",
      name: "Ada Lovelace",
      email: null,
      teamSourceIds: [],
    };
    const assignedTask = {
      ...task("assigned-task", "Name-only assignment"),
      assigneePersonSourceId: person.sourceId,
      assigneeName: person.name,
    };

    const unconfirmed = await run({
      draft: draft({ people: [person], tasks: [assignedTask] }),
      selectedTaskIndexes: new Set([0]),
    });

    expect(addExistingMemberMock).not.toHaveBeenCalled();
    expect(unconfirmed.unresolvedPeople).toBe(1);
    expect(
      buildRequestsMock.mock.calls[0]?.[0].items[0]?.story,
    ).not.toHaveProperty("assigneeId");
  });

  it("adds and assigns an explicitly selected name-only match", async () => {
    const workspaceMember = {
      id: "member-1",
      email: "ada@example.com",
      fullName: "Ada Lovelace",
      username: "ada",
      isActive: true,
      isSystem: false,
    };
    getWorkspaceMembersMock.mockResolvedValue([workspaceMember] as never);
    const person = {
      sourceId: "person-1",
      name: "Ada Lovelace",
      email: null,
      teamSourceIds: [],
    };
    const assignedTask = {
      ...task("assigned-task", "Confirmed name-only assignment"),
      assigneePersonSourceId: person.sourceId,
      assigneeName: person.name,
    };

    const confirmed = await run({
      confirmedMemberIdsByIdentityKey: new Map([
        ["source:person-1", workspaceMember.id],
      ]),
      draft: draft({ people: [person], tasks: [assignedTask] }),
      selectedTaskIndexes: new Set([0]),
    });

    expect(addExistingMemberMock).toHaveBeenCalledWith(
      "destination-team",
      workspaceMember.id,
      ctx,
    );
    expect(confirmed).toMatchObject({
      addedMemberships: 1,
      unresolvedPeople: 0,
    });
    expect(buildRequestsMock.mock.calls[0]?.[0].items[0]?.story).toMatchObject({
      assigneeId: workspaceMember.id,
    });
  });

  it("lets each duplicate source name be explicitly mapped", async () => {
    getWorkspaceMembersMock.mockResolvedValue([
      {
        id: "member-1",
        email: "alex@example.com",
        fullName: "Alex Kim",
        username: "alex",
        isActive: true,
        isSystem: false,
      },
    ] as never);
    const people = ["person-1", "person-2"].map((sourceId) => ({
      sourceId,
      name: "Alex Kim",
      email: null,
      teamSourceIds: ["source-team"],
    }));

    const result = await run({
      confirmedMemberIdsByIdentityKey: new Map([
        ["source:person-1", "member-1"],
        ["source:person-2", "member-1"],
      ]),
      draft: draft({ people }),
    });

    expect(addExistingMemberMock).toHaveBeenCalledTimes(1);
    expect(addExistingMemberMock).toHaveBeenCalledWith(
      "destination-team",
      "member-1",
      ctx,
    );
    expect(result.unresolvedPeople).toBe(0);
  });

  it("uses an explicit workspace-member selection for a non-exact name", async () => {
    const workspaceMember = {
      id: "member-1",
      email: "joseph@example.com",
      fullName: "Joseph Smith",
      username: "joseph",
      isActive: true,
      isSystem: false,
    };
    getWorkspaceMembersMock.mockResolvedValue([workspaceMember] as never);
    const person = {
      sourceId: "person-1",
      name: "Joe Smith",
      email: null,
      teamSourceIds: [],
    };
    const assignedTask = {
      ...task("assigned-task", "Fuzzy member selection"),
      assigneePersonSourceId: person.sourceId,
      assigneeName: person.name,
    };

    const result = await run({
      confirmedMemberIdsByIdentityKey: new Map([
        ["source:person-1", workspaceMember.id],
      ]),
      draft: draft({ people: [person], tasks: [assignedTask] }),
      selectedTaskIndexes: new Set([0]),
    });

    expect(result.unresolvedPeople).toBe(0);
    expect(buildRequestsMock.mock.calls[0]?.[0].items[0]?.story).toMatchObject({
      assigneeId: workspaceMember.id,
    });
  });

  it("uses one explicit member selection to resolve conflicting source claims", async () => {
    const workspaceMember = {
      id: "member-1",
      email: "chosen@example.com",
      fullName: "Chosen Member",
      username: "chosen",
      isActive: true,
      isSystem: false,
    };
    getWorkspaceMembersMock.mockResolvedValue([workspaceMember] as never);
    const firstTask = {
      ...task("task-1", "First claim"),
      assigneePersonSourceId: "person-1",
      assigneeEmail: "one@example.com",
      assigneeName: "First Claim",
    };
    const secondTask = {
      ...task("task-2", "Second claim"),
      assigneePersonSourceId: "person-1",
      assigneeEmail: "two@example.com",
      assigneeName: "Second Claim",
    };

    const result = await run({
      confirmedMemberIdsByIdentityKey: new Map([
        ["source:person-1", workspaceMember.id],
      ]),
      draft: draft({ tasks: [firstTask, secondTask] }),
      selectedTaskIndexes: new Set([0, 1]),
    });

    expect(result.unresolvedPeople).toBe(0);
    expect(
      buildRequestsMock.mock.calls[0]?.[0].items.map(
        (item) => item.story.assigneeId,
      ),
    ).toEqual([workspaceMember.id, workspaceMember.id]);
  });

  it("derives stable source keys from selected tasks only", async () => {
    const result = await run({
      draft: draft({
        tasks: [task("duplicate", "Selected"), task("duplicate", "Excluded")],
      }),
      selectedTaskIndexes: new Set([0]),
    });

    expect(importStoriesBatchMock).toHaveBeenCalledWith(
      expect.objectContaining({
        items: [
          expect.objectContaining({
            sourceKey: "duplicate",
            story: expect.objectContaining({ title: "Selected" }),
          }),
        ],
      }),
      ctx,
    );
    expect(result).toMatchObject({ created: 1, failed: 0, replayed: 0 });
  });

  it("sends explicit story effort through the import endpoint payload", async () => {
    await run({
      draft: draft({
        tasks: [
          {
            ...task("task-with-effort", "Estimated source task"),
            estimateValue: 5,
            estimatedDurationMinutes: 90,
            minimumFocusBlockMinutes: 30,
          },
        ],
      }),
      selectedTaskIndexes: new Set([0]),
    });

    expect(importStoriesBatchMock).toHaveBeenCalledWith(
      expect.objectContaining({
        items: [
          expect.objectContaining({
            story: expect.objectContaining({
              estimateValue: 5,
              estimatedDurationMinutes: 90,
              minimumFocusBlockMinutes: 30,
            }),
          }),
        ],
      }),
      ctx,
    );
  });

  it("partitions mixed Jira and file identities into provider-safe requests", async () => {
    const tasks = [
      task("pro-42", "Jira-shaped task"),
      task("trello-card-id", "Non-Jira task"),
    ];

    await run({
      draft: draft({ sourceType: "jira_csv", tasks }),
      selectedTaskIndexes: new Set([0, 1]),
    });

    expect(buildRequestsMock.mock.calls.map(([request]) => request)).toEqual([
      expect.objectContaining({
        items: [expect.objectContaining({ sourceKey: "PRO-42" })],
        provider: "jira_csv",
      }),
      expect.objectContaining({
        items: [expect.objectContaining({ sourceKey: "trello-card-id" })],
        provider: "file",
      }),
    ]);
  });

  it("reuses the seeded fallback for the first detected source team", async () => {
    const sourceTeam = {
      sourceId: "source-team",
      name: "A source team name that is longer than twenty four characters",
      code: null,
      color: null,
      description: null,
      isPrivate: true,
    };
    const destination = getImportSourceTeamDestination(sourceTeam);

    const result = await run({
      draft: draft({ teams: [sourceTeam] }),
      fallbackTeamCode: destination.code,
      fallbackTeamCreated: true,
      fallbackTeamId: "seeded-team",
      fallbackTeamIsPrivate: true,
      fallbackTeamIsNew: true,
      fallbackTeamName: destination.name,
      structureMode: "preserve",
    });

    expect(destination.name).toHaveLength(24);
    expect(createTeamMock).not.toHaveBeenCalled();
    expect(result.createdTeams).toBe(1);
    expect(getTeamStatusesMock).toHaveBeenCalledTimes(1);
  });

  it("does not auto-match a source team to a newly created fallback with different privacy", async () => {
    const sourceTeam = {
      sourceId: "private-source-team",
      name: "Private platform",
      code: "PVT",
      color: null,
      description: null,
      isPrivate: true,
    };
    createTeamMock.mockResolvedValue({
      data: { id: "private-destination-team" },
      error: null,
    } as never);

    const result = await run({
      draft: draft({ teams: [sourceTeam] }),
      fallbackTeamCreated: true,
      fallbackTeamIsNew: true,
      fallbackTeamIsPrivate: false,
      fallbackTeamName: sourceTeam.name,
      structureMode: "preserve",
    });

    expect(createTeamMock).toHaveBeenCalledWith(
      expect.objectContaining({ isPrivate: true }),
      ctx,
    );
    expect(getTeamStatusesMock).toHaveBeenCalledWith(
      "private-destination-team",
      ctx,
    );
    expect(result.createdTeams).toBe(2);
  });

  it("blocks private source work from being combined into a public team", async () => {
    const privateSourceTeam = {
      sourceId: "private-source-team",
      name: "Private platform",
      code: "PVT",
      color: null,
      description: null,
      isPrivate: true,
    };
    const privateTask = {
      ...task("private-task", "Confidential migration"),
      teamSourceId: privateSourceTeam.sourceId,
    };

    await expect(
      run({
        draft: draft({ teams: [privateSourceTeam], tasks: [privateTask] }),
        fallbackTeamIsPrivate: false,
        selectedTaskIndexes: new Set([0]),
        structureMode: "single",
      }),
    ).rejects.toThrow(
      "Private source work cannot be combined into a public destination team",
    );
    expect(importStoriesBatchMock).not.toHaveBeenCalled();
  });

  it("reuses a long source team on retry using the normalized created name", async () => {
    const sourceTeam = {
      sourceId: "source-team",
      name: "A source team name that is longer than twenty four characters",
      code: null,
      color: null,
      description: null,
      isPrivate: false,
    };
    const destination = getImportSourceTeamDestination(sourceTeam);
    const existingTeam = {
      id: "existing-source-team",
      ...destination,
    } as Team;

    await run({
      draft: draft({ teams: [sourceTeam] }),
      existingTeams: [existingTeam],
      structureMode: "preserve",
    });

    expect(createTeamMock).not.toHaveBeenCalled();
    expect(getTeamStatusesMock).toHaveBeenCalledWith(
      "existing-source-team",
      ctx,
    );
  });

  it("uses the same privacy-aware source-team resolution policy for name and code matches", () => {
    const sourceTeam = {
      sourceId: "source-team",
      name: "Platform",
      code: "PLT",
      color: null,
      description: null,
      isPrivate: true,
    };
    const privateTeam = {
      id: "private-team",
      name: "Different name",
      code: "PLT",
      isPrivate: true,
    } as Team;
    const publicTeam = {
      id: "public-team",
      name: "Platform",
      code: "PUB",
      isPrivate: false,
    } as Team;

    expect(
      resolveImportSourceTeam(sourceTeam, [publicTeam, privateTeam]),
    ).toEqual({ kind: "ambiguous" });
    expect(resolveImportSourceTeam(sourceTeam, [publicTeam])).toEqual({
      kind: "none",
    });
    expect(
      resolveImportSourceTeam(
        { ...sourceTeam, code: "MOB", isPrivate: false },
        [
          {
            id: "platform-team",
            name: "Platform",
            code: "WEB",
            isPrivate: false,
          } as Team,
          {
            id: "mobile-team",
            name: "Mobile",
            code: "MOB",
            isPrivate: false,
          } as Team,
        ],
      ),
    ).toEqual({ kind: "ambiguous" });
    expect(
      resolveImportSourceTeam({ ...sourceTeam, code: null }, [
        { ...privateTeam, id: "private-1", name: "Platform" },
        { ...privateTeam, id: "private-2", name: "Platform" },
      ]),
    ).toEqual({ kind: "ambiguous" });
    expect(
      resolveImportSourceTeam(sourceTeam, [
        { ...privateTeam, id: "private-1", name: "Platform", code: "WEB" },
        { ...privateTeam, id: "private-2", name: "Platform", code: "PLT" },
      ]),
    ).toEqual({
      kind: "unique",
      team: {
        ...privateTeam,
        id: "private-2",
        name: "Platform",
        code: "PLT",
      },
    });
  });

  it("adds the importing admin to a uniquely reused source team before reads", async () => {
    const sourceTeam = {
      sourceId: "source-team",
      name: "Platform",
      code: "PLT",
      color: null,
      description: null,
      isPrivate: false,
    };
    const existingTeam = {
      id: "existing-source-team",
      name: "Platform",
      code: "PLT",
      isPrivate: false,
    } as Team;

    const result = await run({
      draft: draft({ teams: [sourceTeam] }),
      existingTeams: [existingTeam],
      structureMode: "preserve",
    });

    expect(addExistingMemberMock).toHaveBeenCalledWith(
      "existing-source-team",
      "actor-user",
      ctx,
    );
    expect(getTeamStatusesMock).toHaveBeenCalledWith(
      "existing-source-team",
      ctx,
    );
    expect(result.addedMemberships).toBe(1);
  });

  it("does not route work under an ambiguous source team to the fallback", async () => {
    const sourceTeam = {
      sourceId: "source-team",
      name: "Platform",
      code: null,
      color: null,
      description: null,
      isPrivate: false,
    };
    const matchingTeams = ["one", "two"].map(
      (id) =>
        ({
          id,
          name: "Platform",
          code: id.toUpperCase(),
          isPrivate: false,
        }) as Team,
    );

    const result = await run({
      draft: draft({
        teams: [sourceTeam],
        tasks: [
          {
            ...task("card-1", "Keep this scoped"),
            teamSourceId: "source-team",
          },
        ],
      }),
      existingTeams: matchingTeams,
      joinedTeamIds: new Set(["destination-team", "one", "two"]),
      selectedTaskIndexes: new Set([0]),
      structureMode: "preserve",
    });

    expect(createTeamMock).not.toHaveBeenCalled();
    expect(importStoriesBatchMock).not.toHaveBeenCalled();
    expect(result).toMatchObject({ destinationConflicts: 1, failed: 1 });
    expect(result.items[0]?.error?.code).toBe("destination_team_conflict");
  });

  it("blocks multiple source teams from collapsing into one destination team", async () => {
    const sourceTeams = [
      {
        sourceId: "source-one",
        name: "Platform",
        code: "PLT",
        color: null,
        description: null,
        isPrivate: false,
      },
      {
        sourceId: "source-two",
        name: " platform ",
        code: "PLT",
        color: null,
        description: null,
        isPrivate: false,
      },
    ];
    const existingTeam = {
      id: "platform-team",
      name: "Platform",
      code: "PLT",
      isPrivate: false,
    } as Team;

    const result = await run({
      draft: draft({
        teams: sourceTeams,
        tasks: sourceTeams.map((sourceTeam, index) => ({
          ...task(`task-${index + 1}`, `Task ${index + 1}`),
          teamSourceId: sourceTeam.sourceId,
        })),
      }),
      existingTeams: [existingTeam],
      joinedTeamIds: new Set(["destination-team", existingTeam.id]),
      selectedTaskIndexes: new Set([0, 1]),
      structureMode: "preserve",
    });

    expect(createTeamMock).not.toHaveBeenCalled();
    expect(importStoriesBatchMock).not.toHaveBeenCalled();
    expect(result).toMatchObject({
      destinationConflicts: 2,
      failed: 2,
    });
  });

  it("preserves objective fidelity when creating an unmatched objective", async () => {
    const objective = importObjective("objective-source", "Private launch", {
      description: "Restricted plan",
      shortSummary: "Launch privately",
      color: "#aabbcc",
      isPrivate: true,
      priority: "High",
    });
    getObjectiveStatusesMock.mockResolvedValue([
      {
        id: "planned",
        name: "Planned",
        category: "unstarted",
        isDefault: true,
        orderIndex: 1,
      },
    ] as never);
    createObjectiveMock.mockResolvedValue({
      objective: {
        id: "private-objective",
        name: objective.name,
        teamId: "destination-team",
        isPrivate: true,
      },
      keyResults: [],
    } as never);

    const result = await run({
      draft: draft({ objectives: [objective] }),
      selectedObjectiveSourceIds: new Set([objective.sourceId]),
    });

    expect(createObjectiveMock).toHaveBeenCalledWith(
      expect.objectContaining({
        color: "#AABBCC",
        description: "Restricted plan",
        isPrivate: true,
        priority: "High",
        shortSummary: "Launch privately",
      }),
      ctx,
    );
    expect(result).toMatchObject({
      createdObjectives: 1,
      destinationConflicts: 0,
    });
  });

  it("creates a strategic pillar and aligns its selected objective", async () => {
    const pillar = {
      sourceId: "pillar-source",
      name: "Customer growth",
      description: "Grow the active customer base",
      orderIndex: 0,
    };
    const objective = importObjective(
      "objective-source",
      "Improve activation",
      { pillarSourceId: pillar.sourceId },
    );
    getTeamObjectivesMock.mockResolvedValue([
      {
        id: "objective-id",
        name: objective.name,
        teamId: "destination-team",
        isPrivate: false,
      },
    ] as never);
    createStrategicPillarMock.mockResolvedValue({
      id: "pillar-id",
      name: pillar.name,
      description: pillar.description,
      orderIndex: pillar.orderIndex,
      objectiveIds: [],
    });

    const result = await run({
      draft: draft({ strategicPillars: [pillar], objectives: [objective] }),
      selectedObjectiveSourceIds: new Set([objective.sourceId]),
      selectedStrategicPillarSourceIds: new Set([pillar.sourceId]),
    });

    expect(createStrategicPillarMock).toHaveBeenCalledWith(
      {
        name: pillar.name,
        description: pillar.description,
        orderIndex: pillar.orderIndex,
      },
      ctx,
    );
    expect(alignObjectiveToPillarMock).toHaveBeenCalledWith(
      "objective-id",
      "pillar-id",
      ctx,
    );
    expect(result).toMatchObject({
      alignedObjectives: 1,
      createdStrategicPillars: 1,
      destinationConflicts: 0,
    });
  });

  it("imports a selected strategic pillar without creating or loading a fallback team", async () => {
    const pillar = {
      sourceId: "pillar-source",
      name: "Customer growth",
      description: "Grow the active customer base",
      orderIndex: 0,
    };
    createStrategicPillarMock.mockResolvedValue({
      id: "pillar-id",
      name: pillar.name,
      description: pillar.description,
      orderIndex: pillar.orderIndex,
      objectiveIds: [],
    });

    const result = await run({
      draft: draft({ strategicPillars: [pillar] }),
      fallbackTeamId: null,
      selectedStrategicPillarSourceIds: new Set([pillar.sourceId]),
    });

    expect(createStrategicPillarMock).toHaveBeenCalledTimes(1);
    expect(getObjectiveStatusesMock).not.toHaveBeenCalled();
    expect(getWorkspaceMembersMock).not.toHaveBeenCalled();
    expect(getTeamStatusesMock).not.toHaveBeenCalled();
    expect(getTeamMembersMock).not.toHaveBeenCalled();
    expect(result).toMatchObject({
      createdStrategicPillars: 1,
      teamId: null,
    });
  });

  it("requires a fallback destination for team-scoped work", async () => {
    await expect(
      run({
        draft: draft({ tasks: [task("task-1", "Team-scoped work")] }),
        fallbackTeamId: null,
        selectedTaskIndexes: new Set([0]),
      }),
    ).rejects.toThrow("A destination team is required for team-scoped work");

    expect(getObjectiveStatusesMock).not.toHaveBeenCalled();
    expect(importStoriesBatchMock).not.toHaveBeenCalled();
  });

  it("does not import or align an unselected strategic pillar", async () => {
    const pillar = {
      sourceId: "pillar-source",
      name: "Customer growth",
      description: null,
      orderIndex: 0,
    };
    const objective = importObjective(
      "objective-source",
      "Improve activation",
      { pillarSourceId: pillar.sourceId },
    );
    getTeamObjectivesMock.mockResolvedValue([
      {
        id: "objective-id",
        name: objective.name,
        teamId: "destination-team",
        isPrivate: false,
      },
    ] as never);

    const result = await run({
      draft: draft({ strategicPillars: [pillar], objectives: [objective] }),
      selectedObjectiveSourceIds: new Set([objective.sourceId]),
    });

    expect(getStrategyMapMock).not.toHaveBeenCalled();
    expect(createStrategicPillarMock).not.toHaveBeenCalled();
    expect(alignObjectiveToPillarMock).not.toHaveBeenCalled();
    expect(result).toMatchObject({
      alignedObjectives: 0,
      createdStrategicPillars: 0,
      destinationConflicts: 0,
    });
  });

  it("reuses an exact pillar alignment without creating or writing on retry", async () => {
    const pillar = {
      sourceId: "pillar-source",
      name: " Customer   Growth ",
      description: null,
      orderIndex: 3,
    };
    const objective = importObjective(
      "objective-source",
      "Improve activation",
      { pillarSourceId: pillar.sourceId },
    );
    getStrategyMapMock.mockResolvedValue({
      ultimateGoal: "Grow sustainably",
      description: null,
      pillars: [
        {
          id: "pillar-id",
          name: "customer growth",
          description: "Existing description",
          orderIndex: 0,
          objectiveIds: ["objective-id"],
        },
      ],
    });
    getTeamObjectivesMock.mockResolvedValue([
      {
        id: "objective-id",
        name: objective.name,
        teamId: "destination-team",
        isPrivate: false,
      },
    ] as never);

    const result = await run({
      draft: draft({ strategicPillars: [pillar], objectives: [objective] }),
      selectedObjectiveSourceIds: new Set([objective.sourceId]),
      selectedStrategicPillarSourceIds: new Set([pillar.sourceId]),
    });

    expect(createStrategicPillarMock).not.toHaveBeenCalled();
    expect(alignObjectiveToPillarMock).not.toHaveBeenCalled();
    expect(result).toMatchObject({
      alignedObjectives: 0,
      createdStrategicPillars: 0,
      destinationConflicts: 0,
    });
  });

  it("skips duplicate source pillar names and ambiguous destination matches", async () => {
    const duplicateSourcePillars = [
      {
        sourceId: "pillar-source-1",
        name: "Customer growth",
        description: null,
        orderIndex: 0,
      },
      {
        sourceId: "pillar-source-2",
        name: " customer   GROWTH ",
        description: null,
        orderIndex: 1,
      },
    ];

    const duplicateResult = await run({
      draft: draft({ strategicPillars: duplicateSourcePillars }),
      selectedStrategicPillarSourceIds: new Set(
        duplicateSourcePillars.map((pillar) => pillar.sourceId),
      ),
    });

    expect(createStrategicPillarMock).not.toHaveBeenCalled();
    expect(duplicateResult.destinationConflicts).toBe(2);

    getStrategyMapMock.mockResolvedValue({
      ultimateGoal: "",
      description: null,
      pillars: [
        {
          id: "pillar-1",
          name: "Reliability",
          description: null,
          orderIndex: 0,
          objectiveIds: [],
        },
        {
          id: "pillar-2",
          name: " reliability ",
          description: null,
          orderIndex: 1,
          objectiveIds: [],
        },
      ],
    });
    const ambiguousResult = await run({
      draft: draft({
        strategicPillars: [
          {
            sourceId: "pillar-source",
            name: "Reliability",
            description: null,
            orderIndex: 0,
          },
        ],
      }),
      selectedStrategicPillarSourceIds: new Set(["pillar-source"]),
    });

    expect(createStrategicPillarMock).not.toHaveBeenCalled();
    expect(ambiguousResult.destinationConflicts).toBe(1);
  });

  it("does not overwrite a different existing objective pillar alignment", async () => {
    const pillar = {
      sourceId: "pillar-source",
      name: "Customer growth",
      description: null,
      orderIndex: 0,
    };
    const objective = importObjective(
      "objective-source",
      "Improve activation",
      { pillarSourceId: pillar.sourceId },
    );
    getStrategyMapMock.mockResolvedValue({
      ultimateGoal: "",
      description: null,
      pillars: [
        {
          id: "desired-pillar",
          name: pillar.name,
          description: null,
          orderIndex: 0,
          objectiveIds: [],
        },
        {
          id: "current-pillar",
          name: "Operational excellence",
          description: null,
          orderIndex: 1,
          objectiveIds: ["objective-id"],
        },
      ],
    });
    getTeamObjectivesMock.mockResolvedValue([
      {
        id: "objective-id",
        name: objective.name,
        teamId: "destination-team",
        isPrivate: false,
      },
    ] as never);

    const result = await run({
      draft: draft({ strategicPillars: [pillar], objectives: [objective] }),
      selectedObjectiveSourceIds: new Set([objective.sourceId]),
      selectedStrategicPillarSourceIds: new Set([pillar.sourceId]),
    });

    expect(alignObjectiveToPillarMock).not.toHaveBeenCalled();
    expect(result).toMatchObject({
      alignedObjectives: 0,
      destinationConflicts: 1,
    });
  });

  it("skips an objective when an exact destination name exists with incompatible privacy", async () => {
    const objective = importObjective("objective-source", "Private launch", {
      isPrivate: true,
    });
    getTeamObjectivesMock.mockResolvedValue([
      {
        id: "public-objective",
        name: objective.name,
        teamId: "destination-team",
        isPrivate: false,
      },
    ] as never);

    const result = await run({
      draft: draft({ objectives: [objective] }),
      selectedObjectiveSourceIds: new Set([objective.sourceId]),
    });

    expect(createObjectiveMock).not.toHaveBeenCalled();
    expect(result).toMatchObject({
      createdObjectives: 0,
      destinationConflicts: 1,
    });
  });

  it("skips an ambiguous objective even when legacy force-create input is present", async () => {
    const objective = importObjective("objective-source", "Shared roadmap");
    getObjectiveStatusesMock.mockResolvedValue([
      {
        id: "planned",
        name: "Planned",
        category: "unstarted",
        isDefault: true,
        orderIndex: 1,
      },
    ] as never);
    getTeamObjectivesMock.mockResolvedValue([
      {
        id: "objective-1",
        name: objective.name,
        teamId: "destination-team",
        isPrivate: false,
      },
      {
        id: "objective-2",
        name: objective.name,
        teamId: "destination-team",
        isPrivate: false,
      },
    ] as never);

    const result = await run({
      draft: draft({ objectives: [objective] }),
      forceCreateObjectiveSourceIds: new Set([objective.sourceId]),
      selectedObjectiveSourceIds: new Set([objective.sourceId]),
    });
    expect(createObjectiveMock).not.toHaveBeenCalled();
    expect(result.destinationConflicts).toBe(1);
  });

  it("skips every selected source objective with the same normalized name and target team", async () => {
    const objectives = [
      importObjective("objective-1", "Migration roadmap"),
      importObjective("objective-2", "  migration   ROADMAP "),
    ];

    const result = await run({
      draft: draft({ objectives }),
      selectedObjectiveSourceIds: new Set(
        objectives.map((objective) => objective.sourceId),
      ),
    });

    expect(createObjectiveMock).not.toHaveBeenCalled();
    expect(result).toMatchObject({
      createdObjectives: 0,
      destinationConflicts: 2,
    });
  });

  it("reuses a cached objective mapping before destination-name resolution", async () => {
    const objective = importObjective("objective-source", "Cached roadmap");
    const sourceObjectiveCache = new Map([
      [
        objective.sourceId,
        { id: "cached-objective", teamId: "destination-team" },
      ],
    ]);

    const result = await run({
      draft: draft({ objectives: [objective] }),
      forceCreateObjectiveSourceIds: new Set([objective.sourceId]),
      selectedObjectiveSourceIds: new Set([objective.sourceId]),
      sourceObjectiveCache,
    });

    expect(createObjectiveMock).not.toHaveBeenCalled();
    expect(result.createdObjectives).toBe(0);
  });

  it("skips a key result when its normalized destination name is ambiguous", async () => {
    const objective = {
      sourceId: "objective-source",
      name: "Migration objective",
      description: null,
      shortSummary: null,
      color: null,
      isPrivate: false,
      status: null,
      statusCategory: null,
      priority: "No Priority" as const,
      pillarSourceId: null,
      leadPersonSourceId: null,
      teamSourceId: null,
      startDate: null,
      endDate: null,
    };
    const keyResult = {
      sourceId: "key-result-source",
      name: "Move every card",
      objectiveSourceId: objective.sourceId,
      measurementType: "percentage" as const,
      startValue: 0,
      currentValue: 25,
      targetValue: 100,
      leadPersonSourceId: null,
      contributorPersonSourceIds: [],
      startDate: "2026-09-01",
      endDate: "2026-09-30",
    };
    getTeamObjectivesMock.mockResolvedValue([
      {
        id: "objective-id",
        name: objective.name,
        teamId: "destination-team",
        isPrivate: false,
      },
    ] as never);
    getObjectiveKeyResultsMock.mockResolvedValue([
      { id: "ambiguous-1", name: keyResult.name },
      { id: "ambiguous-2", name: keyResult.name },
    ] as never);
    const result = await run({
      draft: draft({ objectives: [objective], keyResults: [keyResult] }),
      selectedObjectiveSourceIds: new Set([objective.sourceId]),
    });

    expect(createKeyResultsMock).not.toHaveBeenCalled();
    expect(result).toMatchObject({
      createdKeyResults: 0,
      destinationConflicts: 1,
    });
  });

  it("skips every duplicate normalized source key-result name", async () => {
    const objective = importObjective(
      "objective-source",
      "Migration objective",
    );
    const keyResults = [
      importKeyResult("key-result-1", objective.sourceId, "Move every card"),
      importKeyResult("key-result-2", objective.sourceId, " move  every card "),
    ];
    getTeamObjectivesMock.mockResolvedValue([
      {
        id: "objective-id",
        name: objective.name,
        teamId: "destination-team",
        isPrivate: false,
      },
    ] as never);
    getObjectiveKeyResultsMock.mockResolvedValue([]);

    const result = await run({
      draft: draft({ objectives: [objective], keyResults }),
      selectedObjectiveSourceIds: new Set([objective.sourceId]),
    });

    expect(createKeyResultsMock).not.toHaveBeenCalled();
    expect(result).toMatchObject({
      createdKeyResults: 0,
      destinationConflicts: 2,
    });
  });

  it("maps created key results by response position and makes their objective authoritative", async () => {
    const objectiveA = importObjective("objective-a", "Objective A");
    const objectiveB = importObjective("objective-b", "Objective B");
    const keyResults = [
      importKeyResult("key-result-1", objectiveB.sourceId, "First result"),
      importKeyResult("key-result-2", objectiveB.sourceId, "Second result"),
    ];
    getTeamObjectivesMock.mockResolvedValue([
      {
        id: "objective-a-id",
        name: objectiveA.name,
        teamId: "destination-team",
        isPrivate: false,
      },
      {
        id: "objective-b-id",
        name: objectiveB.name,
        teamId: "destination-team",
        isPrivate: false,
      },
    ] as never);
    getObjectiveKeyResultsMock.mockResolvedValue([]);
    createKeyResultsMock.mockResolvedValue([
      { id: "position-1", name: "Server response one" },
      { id: "position-2", name: "Server response two" },
    ] as never);
    const firstTask = {
      ...task("task-1", "First linked task"),
      objectiveSourceId: objectiveA.sourceId,
      keyResultSourceId: keyResults[0].sourceId,
    };
    const secondTask = {
      ...task("task-2", "Second linked task"),
      keyResultSourceId: keyResults[1].sourceId,
    };

    await run({
      draft: draft({
        objectives: [objectiveA, objectiveB],
        keyResults,
        tasks: [firstTask, secondTask],
      }),
      selectedObjectiveSourceIds: new Set([
        objectiveA.sourceId,
        objectiveB.sourceId,
      ]),
      selectedTaskIndexes: new Set([0, 1]),
    });

    const request = buildRequestsMock.mock.calls[0][0];
    expect(request.items).toEqual([
      expect.objectContaining({
        story: expect.objectContaining({
          keyResultId: "position-1",
          objectiveId: "objective-b-id",
        }),
      }),
      expect.objectContaining({
        story: expect.objectContaining({
          keyResultId: "position-2",
          objectiveId: "objective-b-id",
        }),
      }),
    ]);
  });

  it("keeps workspace-scoped objective and key-result links on a cross-team story", async () => {
    const sourceTeams = [
      {
        sourceId: "source-a",
        name: "Alpha",
        code: "ALP",
        color: null,
        description: null,
        isPrivate: false,
      },
      {
        sourceId: "source-b",
        name: "Beta",
        code: "BET",
        color: null,
        description: null,
        isPrivate: false,
      },
    ];
    const destinationTeams = sourceTeams.map(
      (team, index) =>
        ({
          id: `team-${index + 1}`,
          name: team.name,
          code: team.code,
          isPrivate: false,
        }) as Team,
    );
    const objective = importObjective(
      "objective-source",
      "Workspace objective",
      { teamSourceId: "source-a" },
    );
    const keyResult = importKeyResult(
      "key-result-source",
      objective.sourceId,
      "Reach the target",
    );
    const linkedTask = {
      ...task("task-source", "Cross-team delivery"),
      teamSourceId: "source-b",
      objectiveSourceId: objective.sourceId,
      keyResultSourceId: keyResult.sourceId,
    };
    getTeamObjectivesMock.mockResolvedValue([
      {
        id: "objective-id",
        name: objective.name,
        teamId: "team-1",
        isPrivate: false,
      },
    ] as never);
    getObjectiveKeyResultsMock.mockResolvedValue([
      { id: "key-result-id", name: keyResult.name },
    ] as never);

    await run({
      draft: draft({
        teams: sourceTeams,
        objectives: [objective],
        keyResults: [keyResult],
        tasks: [linkedTask],
      }),
      existingTeams: destinationTeams,
      joinedTeamIds: new Set(["destination-team", "team-1", "team-2"]),
      selectedObjectiveSourceIds: new Set([objective.sourceId]),
      selectedTaskIndexes: new Set([0]),
      structureMode: "preserve",
    });

    const importedStory = buildRequestsMock.mock.calls
      .flatMap(([request]) => request.items)
      .find(({ sourceKey }) => sourceKey === linkedTask.sourceId)?.story;
    expect(importedStory).toMatchObject({
      teamId: "team-2",
      objectiveId: "objective-id",
      keyResultId: "key-result-id",
    });
  });

  it("counts and omits each same-team-only sprint relationship mismatch", async () => {
    const sourceTeams = [
      {
        sourceId: "source-a",
        name: "Alpha",
        code: "ALP",
        color: null,
        description: null,
        isPrivate: false,
      },
      {
        sourceId: "source-b",
        name: "Beta",
        code: "BET",
        color: null,
        description: null,
        isPrivate: false,
      },
    ];
    const destinationTeams = sourceTeams.map(
      (team, index) =>
        ({
          id: `team-${index + 1}`,
          name: team.name,
          code: team.code,
          isPrivate: false,
        }) as Team,
    );
    const objective = importObjective(
      "objective-source",
      "Workspace objective",
      { teamSourceId: "source-a" },
    );
    const sprint = {
      sourceId: "sprint-source",
      name: "Cross-team sprint",
      goal: null,
      teamSourceId: "source-b",
      objectiveSourceId: objective.sourceId,
      startDate: "2026-09-01",
      endDate: "2026-09-14",
    };
    const linkedTask = {
      ...task("task-source", "Cross-team sprint delivery"),
      teamSourceId: "source-a",
      sprintSourceId: sprint.sourceId,
    };
    getTeamObjectivesMock.mockResolvedValue([
      {
        id: "objective-id",
        name: objective.name,
        teamId: "team-1",
        isPrivate: false,
      },
    ] as never);
    createSprintMock.mockResolvedValue({
      id: "sprint-id",
      name: sprint.name,
      teamId: "team-2",
      startDate: sprint.startDate,
      endDate: sprint.endDate,
    } as never);

    const result = await run({
      draft: draft({
        teams: sourceTeams,
        objectives: [objective],
        sprints: [sprint],
        tasks: [linkedTask],
      }),
      existingTeams: destinationTeams,
      joinedTeamIds: new Set(["destination-team", "team-1", "team-2"]),
      selectedObjectiveSourceIds: new Set([objective.sourceId]),
      selectedTaskIndexes: new Set([0]),
      structureMode: "preserve",
    });

    expect(createSprintMock).not.toHaveBeenCalled();
    const importedStory = buildRequestsMock.mock.calls
      .flatMap(([request]) => request.items)
      .find(({ sourceKey }) => sourceKey === linkedTask.sourceId)?.story;
    expect(importedStory).not.toHaveProperty("sprintId");
    expect(result.destinationConflicts).toBe(2);
  });

  it("omits a cross-team parent link and counts it once regardless of source order", async () => {
    const sourceTeams = [
      {
        sourceId: "source-a",
        name: "Alpha",
        code: "ALP",
        color: null,
        description: null,
        isPrivate: false,
      },
      {
        sourceId: "source-b",
        name: "Beta",
        code: "BET",
        color: null,
        description: null,
        isPrivate: false,
      },
    ];
    const destinationTeams = sourceTeams.map(
      (team, index) =>
        ({
          id: `team-${index + 1}`,
          name: team.name,
          code: team.code,
          isPrivate: false,
        }) as Team,
    );
    const child = {
      ...task("child", "Cross-team child"),
      teamSourceId: "source-b",
      parentSourceId: "parent",
    };
    const parent = {
      ...task("parent", "Source parent"),
      teamSourceId: "source-a",
    };

    const result = await run({
      draft: draft({ teams: sourceTeams, tasks: [child, parent] }),
      existingTeams: destinationTeams,
      joinedTeamIds: new Set(["destination-team", "team-1", "team-2"]),
      selectedTaskIndexes: new Set([0, 1]),
      structureMode: "preserve",
    });

    const importedChild = buildRequestsMock.mock.calls
      .flatMap(([request]) => request.items)
      .find(({ sourceKey }) => sourceKey === child.sourceId)?.story;
    expect(importedChild).toMatchObject({ teamId: "team-2" });
    expect(importedChild).not.toHaveProperty("parentId");
    expect(result.destinationConflicts).toBe(1);
  });

  it("skips ambiguous labels and sprints instead of choosing the first", async () => {
    const label = {
      sourceId: "label-source",
      name: "Migration",
      color: null,
      teamSourceId: null,
    };
    const sprint = {
      sourceId: "sprint-source",
      name: "Migration sprint",
      goal: null,
      teamSourceId: null,
      objectiveSourceId: null,
      startDate: "2026-09-01",
      endDate: "2026-09-14",
    };
    getWorkspaceLabelsMock.mockResolvedValue([
      { id: "label-1", name: label.name, teamId: null },
      { id: "label-2", name: label.name, teamId: null },
    ] as never);
    getTeamSprintsMock.mockResolvedValue([
      {
        id: "sprint-1",
        name: sprint.name,
        teamId: "destination-team",
        startDate: sprint.startDate,
        endDate: sprint.endDate,
      },
      {
        id: "sprint-2",
        name: sprint.name,
        teamId: "destination-team",
        startDate: sprint.startDate,
        endDate: sprint.endDate,
      },
    ] as never);

    const result = await run({
      draft: draft({ labels: [label], sprints: [sprint] }),
    });

    expect(createLabelMock).not.toHaveBeenCalled();
    expect(createSprintMock).not.toHaveBeenCalled();
    expect(result.destinationConflicts).toBe(2);
  });

  it("skips every source label that collapses to the same destination label", async () => {
    const labels = [
      {
        sourceId: "label-source-a",
        name: "Migration",
        color: "#123456",
        teamSourceId: null,
      },
      {
        sourceId: "label-source-b",
        name: " migration ",
        color: "#654321",
        teamSourceId: null,
      },
    ];
    const labeledTask = {
      ...task("task-source", "Task with duplicate source labels"),
      labelSourceIds: labels.map((label) => label.sourceId),
    };

    const result = await run({
      draft: draft({ labels, tasks: [labeledTask] }),
      selectedTaskIndexes: new Set([0]),
    });

    expect(createLabelMock).not.toHaveBeenCalled();
    expect(
      buildRequestsMock.mock.calls[0]?.[0].items[0]?.story,
    ).not.toHaveProperty("labelIds");
    expect(result).toMatchObject({
      createdLabels: 0,
      destinationConflicts: 2,
    });
  });

  it("skips every source sprint that collapses to the same destination sprint", async () => {
    const sprints = [
      {
        sourceId: "sprint-source-a",
        name: "Migration sprint",
        goal: "Move the first board",
        teamSourceId: null,
        objectiveSourceId: null,
        startDate: "2026-09-01",
        endDate: "2026-09-14",
      },
      {
        sourceId: "sprint-source-b",
        name: " migration sprint ",
        goal: "Move the second board",
        teamSourceId: null,
        objectiveSourceId: null,
        startDate: "2026-09-01",
        endDate: "2026-09-14",
      },
    ];

    const result = await run({ draft: draft({ sprints }) });

    expect(createSprintMock).not.toHaveBeenCalled();
    expect(result).toMatchObject({
      createdSprints: 0,
      destinationConflicts: 2,
    });
  });

  it("keeps equal label names independent across different destination teams", async () => {
    const sourceTeams = [
      {
        sourceId: "source-team-a",
        name: "Alpha",
        code: "ALP",
        color: null,
        description: null,
        isPrivate: false,
      },
      {
        sourceId: "source-team-b",
        name: "Beta",
        code: "BET",
        color: null,
        description: null,
        isPrivate: false,
      },
    ];
    const destinationTeams = sourceTeams.map(
      (team, index) =>
        ({
          id: `team-${index + 1}`,
          name: team.name,
          code: team.code,
          isPrivate: false,
        }) as Team,
    );
    const labels = sourceTeams.map((team, index) => ({
      sourceId: `label-source-${index + 1}`,
      name: "Migration",
      color: null,
      teamSourceId: team.sourceId,
    }));
    createLabelMock.mockImplementation(
      async (input) =>
        ({
          id: `label-${input.teamId}`,
          name: input.name,
          color: input.color,
          teamId: input.teamId ?? null,
        }) as never,
    );

    const result = await run({
      draft: draft({ labels, teams: sourceTeams }),
      existingTeams: destinationTeams,
      joinedTeamIds: new Set([
        "destination-team",
        ...destinationTeams.map((team) => team.id),
      ]),
      structureMode: "preserve",
    });

    expect(createLabelMock).toHaveBeenCalledTimes(2);
    expect(result).toMatchObject({
      createdLabels: 2,
      destinationConflicts: 0,
    });
  });

  it("does not preflight-conflict equal sprint schedules mapped to different objectives", async () => {
    const objectives = [
      importObjective("objective-source-a", "First outcome"),
      importObjective("objective-source-b", "Second outcome"),
    ];
    const destinationObjectives = objectives.map((objective, index) => ({
      id: `objective-${index + 1}`,
      name: objective.name,
      teamId: "destination-team",
      isPrivate: objective.isPrivate,
    }));
    const sprints = objectives.map((objective, index) => ({
      sourceId: `sprint-source-${index + 1}`,
      name: "Migration sprint",
      goal: null,
      teamSourceId: null,
      objectiveSourceId: objective.sourceId,
      startDate: "2026-09-01",
      endDate: "2026-09-14",
    }));
    getTeamObjectivesMock.mockResolvedValue(destinationObjectives as never);
    createSprintMock.mockImplementation(
      async (input) =>
        ({
          id: "created-sprint",
          name: input.name,
          objectiveId: input.objectiveId,
          teamId: input.teamId,
          startDate: input.startDate,
          endDate: input.endDate,
        }) as never,
    );

    const result = await run({
      draft: draft({ objectives, sprints }),
      selectedObjectiveSourceIds: new Set(
        objectives.map((objective) => objective.sourceId),
      ),
    });

    expect(createSprintMock).toHaveBeenCalledTimes(1);
    expect(result).toMatchObject({
      createdSprints: 1,
      destinationConflicts: 1,
    });
  });

  it("preserves an unscoped source label as one global label", async () => {
    const label = {
      sourceId: "global-label",
      name: "Customer request",
      color: "#123456",
      teamSourceId: null,
    };
    const firstTask = {
      ...task("task-a", "First labeled task"),
      labelSourceIds: [label.sourceId],
    };
    const secondTask = {
      ...task("task-b", "Second labeled task"),
      labelSourceIds: [label.sourceId],
    };
    createLabelMock.mockResolvedValue({
      id: "global-label-id",
      name: label.name,
      color: label.color,
      teamId: null,
    } as never);

    await run({
      draft: draft({ labels: [label], tasks: [firstTask, secondTask] }),
      selectedTaskIndexes: new Set([0, 1]),
    });

    expect(createLabelMock).toHaveBeenCalledTimes(1);
    expect(createLabelMock).toHaveBeenCalledWith(
      { color: label.color, name: label.name },
      ctx,
    );
    expect(
      buildRequestsMock.mock.calls.flatMap(([request]) =>
        request.items.map(({ story }) => story.labelIds),
      ),
    ).toEqual([["global-label-id"], ["global-label-id"]]);
  });

  it("imports a global label without requiring or creating a team", async () => {
    const label = {
      sourceId: "global-label",
      name: "Workspace-wide",
      color: null,
      teamSourceId: null,
    };
    createLabelMock.mockResolvedValue({
      id: "global-label-id",
      name: label.name,
      color: "#697386",
      teamId: null,
    } as never);

    const result = await run({
      draft: draft({ labels: [label] }),
      fallbackTeamId: null,
      joinedTeamIds: new Set(),
    });

    expect(getWorkspaceLabelsMock).toHaveBeenCalledWith(ctx);
    expect(getTeamStatusesMock).not.toHaveBeenCalled();
    expect(createTeamMock).not.toHaveBeenCalled();
    expect(createLabelMock).toHaveBeenCalledWith(
      { color: "#697386", name: label.name },
      ctx,
    );
    expect(result).toMatchObject({
      createdLabels: 1,
      createdTeams: 0,
      teamId: null,
    });
  });

  it("does not reuse a sprint aligned to a different objective", async () => {
    const objective = importObjective("objective-source", "Growth outcome");
    const sprint = {
      sourceId: "sprint-source",
      name: "Growth sprint",
      goal: null,
      teamSourceId: null,
      objectiveSourceId: objective.sourceId,
      startDate: "2026-09-01",
      endDate: "2026-09-14",
    };
    getTeamObjectivesMock.mockResolvedValue([
      {
        id: "objective-id",
        name: objective.name,
        teamId: "destination-team",
        isPrivate: false,
      },
    ] as never);
    getTeamSprintsMock.mockResolvedValue([
      {
        id: "wrong-sprint-id",
        name: sprint.name,
        objectiveId: "different-objective-id",
        teamId: "destination-team",
        startDate: sprint.startDate,
        endDate: sprint.endDate,
      },
    ] as never);

    const result = await run({
      draft: draft({ objectives: [objective], sprints: [sprint] }),
      selectedObjectiveSourceIds: new Set([objective.sourceId]),
    });

    expect(createSprintMock).not.toHaveBeenCalled();
    expect(result.createdSprints).toBe(0);
    expect(result.destinationConflicts).toBe(1);
  });

  it("creates only missing story links by normalized exact URL", async () => {
    const linkedTask = {
      ...task("linked-task", "Task with source links"),
      links: [
        { title: "Trello card", url: "https://trello.com/c/source-card" },
        { title: "Migration brief", url: "https://example.com/brief" },
      ],
    };
    getStoryLinksMock.mockResolvedValue([
      {
        id: "existing-link",
        storyId: "story-1",
        title: "Existing source",
        url: "https://trello.com/c/source-card",
      },
    ] as never);

    const result = await run({
      draft: draft({ tasks: [linkedTask] }),
      selectedTaskIndexes: new Set([0]),
    });

    expect(getStoryLinksMock).toHaveBeenCalledWith("story-1", ctx);
    expect(createStoryLinkMock).toHaveBeenCalledTimes(1);
    expect(createStoryLinkMock).toHaveBeenCalledWith(
      {
        storyId: "story-1",
        title: "Migration brief",
        url: "https://example.com/brief",
      },
      ctx,
    );
    expect(result).toMatchObject({ createdLinks: 1, unresolvedLinks: 0 });
  });

  it("reports a failed link write without failing the story import", async () => {
    const linkedTask = {
      ...task("linked-task", "Task with unavailable link"),
      links: [{ title: null, url: "https://example.com/unavailable" }],
    };
    createStoryLinkMock.mockRejectedValue(
      new Error("Link service unavailable"),
    );

    const result = await run({
      draft: draft({ tasks: [linkedTask] }),
      selectedTaskIndexes: new Set([0]),
    });

    expect(result).toMatchObject({
      created: 1,
      createdLinks: 0,
      failed: 0,
      unresolvedLinks: 1,
    });
  });

  it("creates canonical story associations once and reverses blocked-by edges", async () => {
    const firstTask = {
      ...task("task-a", "First task"),
      associations: [
        { type: "blocks" as const, targetSourceId: "task-b" },
        { type: "related" as const, targetSourceId: "task-b" },
        { type: "duplicate" as const, targetSourceId: "task-b" },
      ],
    };
    const secondTask = {
      ...task("task-b", "Second task"),
      associations: [
        { type: "blocked_by" as const, targetSourceId: "task-a" },
        { type: "related" as const, targetSourceId: "task-a" },
        { type: "duplicate" as const, targetSourceId: "task-a" },
      ],
    };

    const result = await run({
      draft: draft({ tasks: [firstTask, secondTask] }),
      selectedTaskIndexes: new Set([0, 1]),
    });

    expect(createStoryAssociationMock.mock.calls).toEqual([
      ["story-1", { toStoryId: "story-2", type: "blocking" }, ctx],
      ["story-1", { toStoryId: "story-2", type: "related" }, ctx],
      ["story-1", { toStoryId: "story-2", type: "duplicate" }, ctx],
    ]);
    expect(result).toMatchObject({
      createdAssociations: 3,
      unresolvedAssociations: 0,
    });
  });

  it("reuses an exact existing association on retry", async () => {
    const firstTask = {
      ...task("task-a", "First task"),
      associations: [{ type: "blocks" as const, targetSourceId: "task-b" }],
    };
    getStoryAssociationsMock.mockResolvedValue([
      {
        id: "association-1",
        fromStoryId: "story-1",
        toStoryId: "story-2",
        type: "blocking",
      },
    ]);

    const result = await run({
      draft: draft({
        tasks: [firstTask, task("task-b", "Second task")],
      }),
      selectedTaskIndexes: new Set([0, 1]),
    });

    expect(createStoryAssociationMock).not.toHaveBeenCalled();
    expect(result.createdAssociations).toBe(0);
  });

  it("never creates an association to an unselected target", async () => {
    const firstTask = {
      ...task("task-a", "First task"),
      associations: [{ type: "related" as const, targetSourceId: "task-b" }],
    };

    const result = await run({
      draft: draft({ tasks: [firstTask, task("task-b", "Second task")] }),
      selectedTaskIndexes: new Set([0]),
    });

    expect(createStoryAssociationMock).not.toHaveBeenCalled();
    expect(result.unresolvedAssociations).toBe(1);
  });

  it("surfaces cross-team associations as deterministic conflicts", async () => {
    const sourceTeams = [
      {
        sourceId: "source-a",
        name: "Alpha",
        code: "ALP",
        color: null,
        description: null,
        isPrivate: false,
      },
      {
        sourceId: "source-b",
        name: "Beta",
        code: "BET",
        color: null,
        description: null,
        isPrivate: false,
      },
    ];
    const destinationTeams = sourceTeams.map(
      (team, index) =>
        ({
          id: `team-${index + 1}`,
          name: team.name,
          code: team.code,
          isPrivate: false,
        }) as Team,
    );
    const firstTask = {
      ...task("task-a", "First task"),
      teamSourceId: "source-a",
      associations: [{ type: "blocks" as const, targetSourceId: "task-b" }],
    };
    const secondTask = {
      ...task("task-b", "Second task"),
      teamSourceId: "source-b",
    };

    const result = await run({
      draft: draft({
        teams: sourceTeams,
        tasks: [firstTask, secondTask],
      }),
      existingTeams: destinationTeams,
      joinedTeamIds: new Set(["destination-team", "team-1", "team-2"]),
      selectedTaskIndexes: new Set([0, 1]),
      structureMode: "preserve",
    });

    expect(createStoryAssociationMock).not.toHaveBeenCalled();
    expect(result).toMatchObject({
      destinationConflicts: 1,
      unresolvedAssociations: 1,
    });
  });

  it("merges collaborators non-destructively on replay and deduplicates unresolved people", async () => {
    const people = [
      {
        sourceId: "assignee",
        name: "Assigned Person",
        email: "assignee@example.com",
        teamSourceIds: [],
      },
      {
        sourceId: "collaborator",
        name: "Collaborating Person",
        email: "collaborator@example.com",
        teamSourceIds: [],
      },
      {
        sourceId: "unknown",
        name: "Unknown Person",
        email: null,
        teamSourceIds: [],
      },
    ];
    getWorkspaceMembersMock.mockResolvedValue([
      {
        id: "assignee-id",
        email: "assignee@example.com",
        fullName: "Assigned Person",
        username: "assignee",
        isActive: true,
        isSystem: false,
      },
      {
        id: "collaborator-id",
        email: "collaborator@example.com",
        fullName: "Collaborating Person",
        username: "collaborator",
        isActive: true,
        isSystem: false,
      },
    ] as never);
    importStoriesBatchMock.mockImplementation(async (request) => ({
      data: {
        counts: {
          total: request.items.length,
          created: 0,
          replayed: request.items.length,
          failed: 0,
        },
        items: request.items.map(({ sourceKey }, index) => ({
          sourceKey,
          storyId: `replayed-story-${index + 1}`,
          created: false,
          error: null,
        })),
      },
    }));
    getStoryCollaboratorIdsMock.mockResolvedValue(["manual-member"]);
    const assignedTask = {
      ...task("task-1", "Assigned work"),
      assigneePersonSourceId: "assignee",
      collaboratorPersonSourceIds: [
        "assignee",
        "collaborator",
        "collaborator",
        "unknown",
      ],
    };
    const unresolvedTask = {
      ...task("task-2", "Unresolved work"),
      assigneePersonSourceId: "unknown",
    };
    const importDraft = draft({
      sourceNamespace: "trello:board:product",
      people,
      tasks: [assignedTask, unresolvedTask],
    });

    const result = await run({
      draft: importDraft,
      selectedTaskIndexes: new Set([0, 1]),
    });

    expect(buildRequestsMock).toHaveBeenCalledWith(
      expect.objectContaining({
        sourceNamespace: "trello:board:product",
      }),
    );
    expect(updateStoryCollaboratorsMock).toHaveBeenCalledWith(
      "replayed-story-1",
      ["manual-member", "collaborator-id"],
      ctx,
    );
    expect(result).toMatchObject({
      appliedCollaborators: 1,
      replayed: 2,
      unresolvedPeople: 1,
    });

    updateStoryCollaboratorsMock.mockClear();
    getStoryCollaboratorIdsMock.mockResolvedValue([
      "manual-member",
      "collaborator-id",
    ]);
    const retried = await run({
      draft: importDraft,
      selectedTaskIndexes: new Set([0, 1]),
    });
    expect(updateStoryCollaboratorsMock).not.toHaveBeenCalled();
    expect(retried.appliedCollaborators).toBe(0);
  });
});
